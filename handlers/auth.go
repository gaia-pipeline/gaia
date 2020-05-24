package handlers

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/gaia-pipeline/gaia/security/rbac"
)

var (
	// errNotAuthorized is thrown when user wants to access resource which is protected
	errNotAuthorized = errors.New("no or invalid jwt token provided. You are not authorized")

	// Non-protected URL paths which are prefix checked
	nonProtectedPathsPrefix = []string{
		"/login",
		"/pipeline/githook",
		"/trigger",
		"/worker/register",
		"/js/",
		"/img/",
		"/fonts/",
		"/css/",
	}

	// Non-protected URL paths which are explicitly checked
	nonProtectedPaths = []string{
		"/",
		"/favicon.ico",
	}
)

// Do is middleware used for each request. Includes functionality that validates the JWT tokens and user
// permissions.
func (a *AuthMiddleware) Do() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if it matches an explicit paths
			for _, paths := range nonProtectedPaths {
				if paths == c.Path() {
					return next(c)
				}
			}

			// Check if it matches an prefix-based paths
			p := "/api/" + gaia.APIVersion
			for _, prefix := range nonProtectedPathsPrefix {
				switch {
				case strings.HasPrefix(c.Path(), p+prefix):
					return next(c)
				case strings.HasPrefix(c.Path(), prefix):
					return next(c)
				}
			}

			token, err := getToken(c)
			if err != nil {
				return c.String(http.StatusUnauthorized, err.Error())
			}

			// Validate token
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// All ok, continue
				username, okUsername := claims["username"]
				roles, okPerms := claims["roles"]
				policies, okPolicies := claims["policies"]
				if okUsername && okPerms && okPolicies && roles != nil {
					// Look through the perms until we find that the user has this permission
					if err := a.checkRole(roles, c.Request().Method, c.Path()); err != nil {
						return c.String(http.StatusForbidden, fmt.Sprintf("Permission denied for user %s: %s", username, err.Error()))
					}

					if err := a.enforceRBAC(c, username.(string), policies); err != nil {
						return c.String(http.StatusForbidden, fmt.Sprintf("Permission denied for user %s: %s", username, err.Error()))
					}

					return next(c)
				}
			}
			return c.String(http.StatusUnauthorized, errNotAuthorized.Error())
		}
	}
}

func (a *AuthMiddleware) enforceRBAC(c echo.Context, username string, policiesFromClaims interface{}) error {
	policies, err := getPoliciesFromClaims(policiesFromClaims)
	if err != nil {
		gaia.Cfg.Logger.Warn("rbac enforcement failed", "error", err.Error(), "username", username)
		return errors.New("failed to get policies")
	}

	group := a.rbacEnforcer.GetDefaultAPIGroup()

	endpoint, ok := group.Endpoints[c.Path()]
	if !ok {
		gaia.Cfg.Logger.Warn("path not mapped to api group", "path", c.Path())
		return nil
	}

	perm, ok := endpoint.Methods[c.Request().Method]
	if !ok {
		gaia.Cfg.Logger.Warn("method not mapped to api group path", "path", c.Path(), "method", c.Request().Method)
		return nil
	}

	ns, act := rbac.ParseStatementAction(perm)

	fullResource := ""
	if endpoint.Param != "" {
		param := c.Param(endpoint.Param)
		if param == "" {
			return fmt.Errorf("param %s missing", endpoint.Param)
		}
		fullResource = fmt.Sprintf("%s/%s/%s", ns, endpoint.Param, param)
	}

	enfCfg := rbac.EnforcerConfig{
		User: rbac.User{
			Username: username,
			Policies: policies,
		},
		Namespace: ns,
		Action:    act,
		Resource:  gaia.RBACPolicyResource(fullResource),
	}

	if err := a.rbacEnforcer.Enforce(enfCfg); err != nil {
		gaia.Cfg.Logger.Warn("rbac enforcement failed", "error", err.Error(), "username", username, "namespace", ns, "action", act)
		return fmt.Errorf("missing required permission %s/%s", ns, act)
	}

	return nil
}

// AuthMiddleware is a simple config struct to be passed into AuthMiddleware. Currently allows the ability to specify
// the permission roles required for each echo endpoint.
type AuthMiddleware struct {
	RoleCategories []*gaia.UserRoleCategory
	rbacEnforcer   rbac.PolicyEnforcer
}

func getPoliciesFromClaims(policyClaims interface{}) (map[string]interface{}, error) {
	if policyClaims == nil || reflect.ValueOf(policyClaims).IsNil() {
		return nil, errors.New("policyClaims is nil")
	}
	pc, ok := policyClaims.(map[string]interface{})
	if !ok {
		return nil, errors.New("policyClaims is not correct type")
	}
	return pc, nil
}

// Finds the required role for the metho & path specified. If it exists we validate that the provided user roles have
// the permission role. If not, error specifying the required role.
func (a *AuthMiddleware) checkRole(userRoles interface{}, method, path string) error {
	perm := a.getRequiredRole(method, path)
	if perm == "" {
		return nil
	}
	for _, role := range userRoles.([]interface{}) {
		if role.(string) == perm {
			return nil
		}
	}
	return fmt.Errorf("required permission role %s", perm)
}

// Iterate over each category to find a permission (if existing) for this API endpoint.
func (a *AuthMiddleware) getRequiredRole(method, path string) string {
	for _, category := range a.RoleCategories {
		for _, role := range category.Roles {
			for _, endpoint := range role.APIEndpoint {
				// If the http method & path match then return the role required for this endpoint
				if method == endpoint.Method && path == endpoint.Path {
					return rolehelper.FullUserRoleName(category, role)
				}
			}
		}
	}
	return ""
}

// Get the JWT token from the echo context
func getToken(c echo.Context) (*jwt.Token, error) {
	// Get the token
	jwtRaw := c.Request().Header.Get("Authorization")
	split := strings.Split(jwtRaw, " ")
	if len(split) != 2 {
		return nil, errNotAuthorized
	}
	jwtString := split[1]

	// Parse token
	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		signingMethodError := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		switch token.Method.(type) {
		case *jwt.SigningMethodHMAC:
			if _, ok := gaia.Cfg.JWTKey.([]byte); !ok {
				return nil, signingMethodError
			}
			return gaia.Cfg.JWTKey, nil
		case *jwt.SigningMethodRSA:
			if _, ok := gaia.Cfg.JWTKey.(*rsa.PrivateKey); !ok {
				return nil, signingMethodError
			}
			return gaia.Cfg.JWTKey.(*rsa.PrivateKey).Public(), nil
		default:
			return nil, signingMethodError
		}
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}
