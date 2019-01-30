package handlers

import (
	"net/http"

	"github.com/gaia-pipeline/gaia/auth"
	"github.com/labstack/echo"
)

// Simply retrieves a list of all user role categories
func PermissionGetAll(c echo.Context) error {
	return c.JSON(http.StatusOK, auth.DefaultUserRoles)
}
