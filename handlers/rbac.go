package handlers

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/labstack/echo"
	"net/http"
)

// RBACMiddleware is the role-based access control middleware struct.
type RBACMiddleware struct {
	category gaia.UserRoleCategory
}

// NewRBACMiddleware creates a new RBACMiddleware.
func NewRBACMiddleware(category gaia.UserRoleCategory) *RBACMiddleware {
	if _, ok := rolehelper.DefaultUserRoles[category]; !ok {
		gaia.Cfg.Logger.Warn("invalid rbac category: %s", category)
	}
	return &RBACMiddleware{category: category}
}

// Do returns echo.MiddlewareFunc and should be used to enforce RBAC on specific echo handlers.
func (rbac RBACMiddleware) Do(role gaia.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, ok := rolehelper.DefaultUserRoles[rbac.category].Roles[role]; !ok {
				gaia.Cfg.Logger.Warn("invalid rbac category role: %s", role)
			}

			token, err := getToken(c)
			if err != nil {
				return c.String(http.StatusUnauthorized, err.Error())
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				username, okUsername := claims["username"]
				roles, okPerms := claims["roles"]
				if okUsername && okPerms && roles != nil {
					fullRequiredPerm := rolehelper.FullUserRoleName(rbac.category, role)
					err := checkRBAC(roles, fullRequiredPerm)
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

func checkRBAC(userRoles interface{}, fullRequiredRole string) error {
	for _, role := range userRoles.([]interface{}) {
		if role.(string) == fullRequiredRole {
			return nil
		}
	}
	return fmt.Errorf("required permission role %s", fullRequiredRole)
}
