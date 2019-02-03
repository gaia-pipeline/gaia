package handlers

import (
	"net/http"

	"github.com/gaia-pipeline/gaia/auth"
	"github.com/labstack/echo"
)

// PermissionGetAll simply returns a list of all the roles available.
func PermissionGetAll(c echo.Context) error {
	return c.JSON(http.StatusOK, auth.DefaultUserRoles)
}
