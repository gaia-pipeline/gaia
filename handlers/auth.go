package handlers

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/auth"
	"github.com/labstack/echo"
)

var (
	// errNotAuthorized is thrown when user wants to access resource which is protected
	errNotAuthorized = errors.New("no or invalid jwt token provided. You are not authorized")
)

// Authentication middleware used for each request. Includes functionality that validates tokens and user permissions.
func AuthMiddleware(roleAuth *AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Login, WebHook callback and static resources are open
			// The webhook callback has it's own authentication method
			if strings.Contains(c.Path(), "/login") ||
				c.Path() == "/" ||
				strings.Contains(c.Path(), "/assets/") ||
				c.Path() == "/favicon.ico" ||
				strings.Contains(c.Path(), "pipeline/githook") {
				return next(c)
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
				}
				return next(c)
			}
			return c.String(http.StatusUnauthorized, errNotAuthorized.Error())
		}
	}
}

// Simple auth config struct to be passed into the AuthMiddleware. Currently allow the ability to specify the
// permission roles applied for each echo request.
type AuthConfig struct {
	RoleCategories []*gaia.UserRoleCategory
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
	return errors.New(fmt.Sprintf("Required permission role %s", perm))
}

// Iterate over each category to find a permission (if existing) for this API endpoint.
func (ra *AuthConfig) getRequiredRole(method, path string) string {
	for _, category := range ra.RoleCategories {
		for _, role := range category.Roles {
			for _, endpoint := range role.ApiEndpoint {
				// If the http method & path match then return the role required for this endpoint
				if method == endpoint.Method && path == endpoint.Path {
					return auth.FullUserRoleName(category, role)
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
