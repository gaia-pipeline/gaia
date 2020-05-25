package handlers

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/casbin/casbin/v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/labstack/echo"
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

// AuthMiddleware is middleware used for each request. Includes functionality that validates the JWT tokens and user
// permissions.
func AuthMiddleware(roleAuth *AuthConfig) echo.MiddlewareFunc {
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
				if okUsername && okPerms && roles != nil {
					// Look through the perms until we find that the user has this permission
					err := roleAuth.checkRole(roles, c.Request().Method, c.Path())
					if err != nil {
						return c.String(http.StatusForbidden, fmt.Sprintf("Permission denied for user %s. %s", username, err.Error()))
					}

					sub := username.(string)
					valid, err := roleAuth.enforceRBAC(c, sub)
					if err != nil {
						return c.String(http.StatusInternalServerError, fmt.Sprintf("Unknown error has occured."))
					}
					if !valid {
						return c.String(http.StatusForbidden, fmt.Sprintf("Permission denied for user."))
					}

				}
				return next(c)
			}
			return c.String(http.StatusUnauthorized, errNotAuthorized.Error())
		}
	}
}

// AuthConfig is a simple config struct to be passed into AuthMiddleware. Currently allows the ability to specify
// the permission roles required for each echo endpoint.
type AuthConfig struct {
	RoleCategories []*gaia.UserRoleCategory
	enforcer       casbin.IEnforcer
	apiGroup       gaia.RBACAPIGroup
}

// Finds the required role for the metho & path specified. If it exists we validate that the provided user roles have
// the permission role. If not, error specifying the required role.
func (ra *AuthConfig) checkRole(userRoles interface{}, method, path string) error {
	perm := ra.getRequiredRole(method, path)
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
func (ra *AuthConfig) getRequiredRole(method, path string) string {
	for _, category := range ra.RoleCategories {
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

func (ra *AuthConfig) enforceRBAC(c echo.Context, username string) (bool, error) {
	group := ra.apiGroup

	endpoint, ok := group.Endpoints[c.Path()]
	if !ok {
		gaia.Cfg.Logger.Warn("path not mapped to api group", "path", c.Path())
		return true, nil
	}

	perm, ok := endpoint.Methods[c.Request().Method]
	if !ok {
		gaia.Cfg.Logger.Warn("method not mapped to api group path", "path", c.Path(), "method", c.Request().Method)
		return true, nil
	}

	splitAction := strings.Split(perm, "/")
	namespace := splitAction[0]
	action := splitAction[1]

	fullResource := "*"
	if endpoint.Param != "" {
		param := c.Param(endpoint.Param)
		if param == "" {
			return false, fmt.Errorf("param %s missing", endpoint.Param)
		}
		fullResource = fmt.Sprintf("%s/%s/%s", namespace, endpoint.Param, param)
	}

	valid, err := ra.enforcer.Enforce(username, namespace, fullResource, action)
	if err != nil {
		return false, err
	}

	gaia.Cfg.Logger.Warn("permission denied for user", "username", username, "namespace", namespace, "resource", fullResource, "action", action)
	return valid, nil
}

func loadAPIGroup() (gaia.RBACAPIGroup, error) {
	file, err := ioutil.ReadFile("apigroup-core.yml")
	if err != nil {
		return gaia.RBACAPIGroup{}, err
	}

	var apiGroup gaia.RBACAPIGroup
	if err := yaml.Unmarshal(file, &apiGroup); err != nil {
		return gaia.RBACAPIGroup{}, err
	}

	return apiGroup, nil
}
