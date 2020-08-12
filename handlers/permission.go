package handlers

import (
	"net/http"

	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/labstack/echo/v4"
)

// PermissionGetAll simply returns a list of all the roles available.
// @Summary Returns a list of default roles.
// @Description Returns a list of all the roles available.
// @Tags rbac
// @Success 200 {array} gaia.UserRoleCategory
// @Router /permission [get]
func PermissionGetAll(c echo.Context) error {
	return c.JSON(http.StatusOK, rolehelper.DefaultUserRoles)
}
