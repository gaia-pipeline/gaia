package handlers

import (
	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
	"net/http"
)

func PermissionGetAll(c echo.Context) error {
	return c.JSON(http.StatusOK, gaia.PermissionsCategories)
}
