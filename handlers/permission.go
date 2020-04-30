package handlers

import (
	"fmt"
	"net/http"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/labstack/echo"
)

// PermissionGetAll simply returns a list of all the roles available.
func PermissionGetAll(c echo.Context) error {
	return c.JSON(http.StatusOK, rolehelper.DefaultUserRoles)
}

// UserGetPermissions returns the permissions for a user.
func UserGetPermissions(c echo.Context) error {
	u := c.Param("username")
	storeService, _ := services.StorageService()
	perms, err := storeService.UserPermissionsGet(u)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, perms)
}

// UserPutPermissions adds or updates permissions for a user.
func UserPutPermissions(c echo.Context) error {
	var perms *gaia.UserPermission
	if err := c.Bind(&perms); err != nil {
		return c.String(http.StatusBadRequest, "Invalid parameters given for request")
	}
	storeService, _ := services.StorageService()
	err := storeService.UserPermissionsPut(perms)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Permissions have been updated")
}

// UserPermissionGroupAdd adds a new permission group. Errors if already exists.
func UserPermissionGroupAdd(c echo.Context) error {
	var group *gaia.UserPermissionGroup
	if err := c.Bind(&group); err != nil {
		return c.String(http.StatusBadRequest, "Invalid parameters given for request")
	}

	storeService, _ := services.StorageService()

	// Check for an existing group and fail if it exists.
	existing, err := storeService.UserPermissionGroupGet(group.Name)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if existing != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Permission group %s already exists", group.Name))
	}

	err = storeService.UserPermissionGroupPut(group)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, fmt.Sprintf("Created permission group %s", group.Name))
}

// UserPermissionGroupUpdate updates a permission group.
func UserPermissionGroupUpdate(c echo.Context) error {
	var group *gaia.UserPermissionGroup
	if err := c.Bind(&group); err != nil {
		return c.String(http.StatusBadRequest, "Invalid parameters given for request")
	}

	storeService, _ := services.StorageService()

	err := storeService.UserPermissionGroupPut(group)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, fmt.Sprintf("Updated permission group %s", group.Name))
}

// UserPermissionGroupGet gets a permission group with the specified name.
func UserPermissionGroupGet(c echo.Context) error {
	name := c.Param("name")

	storeService, _ := services.StorageService()

	group, err := storeService.UserPermissionGroupGet(name)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, group)
}

// UserPermissionGroupGetAll returns all permission groups.
func UserPermissionGroupGetAll(c echo.Context) error {
	storeService, _ := services.StorageService()

	groups, err := storeService.UserPermissionGroupGetAll()
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, groups)
}

// UserPermissionGroupDelete deletes a permission group with the specified name.
func UserPermissionGroupDelete(c echo.Context) error {
	name := c.Param("name")

	storeService, _ := services.StorageService()

	err := storeService.UserPermissionGroupDelete(name)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, fmt.Sprintf("Deleted permission group %s", name))
}
