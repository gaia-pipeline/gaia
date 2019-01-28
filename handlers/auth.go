package handlers

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
	"net/http"
	"strings"
)

var (
	// errNotAuthorized is thrown when user wants to access resource which is protected
	errNotAuthorized = errors.New("no or invalid jwt token provided. You are not authorized")
)

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
				// TODO: roles should never be null at this point!
				if okUsername && okPerms && roles != nil {
					// Look through the perms until we find that the user has this permission
					err := roleAuth.CheckRole(roles, c.Request().Method, c.Path())
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

type AuthConfig struct {
	RoleCategories []*gaia.UserRoleCategory
}

// Check if the given roles are valid
func (ra *AuthConfig) CheckRole(roles interface{}, method, path string) error {
	perm := ra.getRoleApiEndpoint(method, path)
	if perm == "" {
		return nil
	}
	for _, role := range roles.([]interface{}) {
		if role.(string) == perm {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Required permission %s", perm))
}

func (ra *AuthConfig) getRoleApiEndpoint(method, path string) string {
	for _, pcs := range ra.RoleCategories {
		for _, p := range pcs.Roles {
			for _, apie := range p.ApiEndpoint {
				if method == apie.Method && path == apie.Path {
					return p.FlatName(pcs.Name)
				}
			}
		}
	}
	return ""
}

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
