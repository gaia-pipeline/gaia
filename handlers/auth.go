package handlers

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia"
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
func AuthMiddleware() echo.MiddlewareFunc {
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
			if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				return next(c)
			}

			return c.String(http.StatusUnauthorized, errNotAuthorized.Error())
		}
	}
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
