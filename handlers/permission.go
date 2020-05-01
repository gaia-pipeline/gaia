package handlers

import (
	"github.com/gaia-pipeline/gaia"
	"net/http"
	"sort"

	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/labstack/echo"
)

// PermissionGetAll simply returns a list of all the roles available.
func PermissionGetAll(c echo.Context) error {
	return c.JSON(http.StatusOK, mapToPermissionCategories(rolehelper.DefaultUserRoles))
}

func mapToPermissionCategories(perms map[gaia.UserRoleCategory]*gaia.UserRoleCategoryDetails) []permissionCategory {
	var permCategories []permissionCategory
	for categoryName, categoryDetails := range perms {
		pc := permissionCategory{
			Name:        string(categoryName),
			Description: categoryDetails.Description,
		}
		for roleName, roleDetails := range categoryDetails.Roles {
			pc.Roles = append(pc.Roles, permissionRole{
				Name:        string(roleName),
				Description: roleDetails.Description,
			})
		}
		sort.Slice(pc.Roles, func(i, j int) bool {
			return pc.Roles[i].Name < pc.Roles[j].Name
		})
		permCategories = append(permCategories, pc)
	}
	sort.Slice(permCategories, func(i, j int) bool {
		return permCategories[i].Name < permCategories[j].Name
	})
	return permCategories
}

type permissionCategory struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Roles       []permissionRole `json:"roles"`
}

type permissionRole struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
