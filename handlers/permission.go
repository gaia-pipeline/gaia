package handlers

import (
	"net/http"

	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/labstack/echo"
)

// PermissionGetAll simply returns a list of all the roles available.
func PermissionGetAll(c echo.Context) error {
	return c.JSON(http.StatusOK, rolehelper.DefaultUserRoles)
}
