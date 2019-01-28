package handlers

import (
	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
	"net/http"
)

// Simply retrieves a list of all user role categories
func PermissionGetAll(c echo.Context) error {
	return c.JSON(http.StatusOK, gaia.UserRoleCategories)
}
