package handlers

import (
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/labstack/echo"
	"net/http"
)

func PermissionGetAll(c echo.Context) error {
	return c.JSON(http.StatusOK, gaia.PermissionsCategories)
}

func PermissionGroupGetAll(c echo.Context) error {
	store, _ := services.StorageService()
	pgs, err := store.PermissionGroupGetAll()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, pgs)
}

func PermissionGroupCreate(c echo.Context) error {
	var pg *gaia.PermissionGroup

	if err := c.Bind(&pg); err != nil {
		return c.String(http.StatusBadRequest, "Invalid permission group provided")
	}

	store, _ := services.PermissionManager()
	err := store.CreateGroup(pg, false)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "Permission Group added")
}
