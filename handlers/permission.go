package handlers

import (
	"net/http"

	"github.com/gaia-pipeline/gaia/auth"
	"github.com/labstack/echo/v4"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
)

// PermissionGetAll simply returns a list of all the roles available.
func PermissionGetAll(c echo.Context) error {
	return c.JSON(http.StatusOK, rolehelper.DefaultUserRoles)
}
