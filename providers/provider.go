package providers

import (
	"github.com/labstack/echo/v4"
)

// RBACProvider provides all the handler endpoints for RBAC actions.
type RBACProvider interface {
	AddRole(c echo.Context) error
	DeleteRole(c echo.Context) error
	GetAllRoles(c echo.Context) error
	GetUserAttachedRoles(c echo.Context) error
	GetRolesAttachedUsers(c echo.Context) error
	AttachRole(c echo.Context) error
	DetachRole(c echo.Context) error
}

// UserProvider provides all the handler endpoints for User actions.
type UserProvider interface {
	UserLogin(c echo.Context) error
	UserGetAll(c echo.Context) error
	UserChangePassword(c echo.Context) error
	UserResetTriggerToken(c echo.Context) error
	UserDelete(c echo.Context) error
	UserAdd(c echo.Context) error
	UserGetPermissions(c echo.Context) error
	UserPutPermissions(c echo.Context) error
}
