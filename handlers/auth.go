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

// AuthContext is the wrapped context to pass through the echo handlers and middleware. This allows us to bind the user
// RBAC policy names into the server request context.
type AuthContext struct {
	echo.Context
	username string
	policies map[string]interface{}
}

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
				policies, okPolicies := claims["policies"]
				if okUsername && okPerms && okPolicies && roles != nil {
					// Look through the perms until we find that the user has this permission
					err := roleAuth.checkRole(roles, c.Request().Method, c.Path())
					if err != nil {
						return c.String(http.StatusForbidden, fmt.Sprintf("Permission denied for user %s. %s", username, err.Error()))
					}

					policiesFromClaims, err := getPoliciesFromClaims(policies)
					if err != nil {
						return c.String(http.StatusForbidden, fmt.Sprintf("Permission denied for user %s. %s", username, err.Error()))
					}

					ctx := AuthContext{
						Context:  c,
						username: username.(string),
						policies: policiesFromClaims,
					}
					return next(ctx)
				}
			}
			return c.String(http.StatusUnauthorized, errNotAuthorized.Error())
		}
	}
}

// AuthConfig is a simple config struct to be passed into AuthMiddleware. Currently allows the ability to specify
// the permission roles required for each echo endpoint.
type AuthConfig struct {
	RoleCategories []*gaia.UserRoleCategory
}

func getPoliciesFromClaims(policyClaims interface{}) (map[string]interface{}, error) {
	if policyClaims == nil || reflect.ValueOf(policyClaims).IsNil() {
		return nil, errors.New("nil policyClaims")
	}
	if _, ok := policyClaims.(map[string]interface{}); !ok {
		return nil, errors.New("policyClaims is not correct type")
	}
	return policyClaims.(map[string]interface{}), nil
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
