package rolehelper

import (
	"testing"

	"github.com/gaia-pipeline/gaia"
)

var mockData = map[gaia.UserRoleCategory]*gaia.UserRoleCategoryDetails{
	"CategoryA": {
		Roles: map[gaia.UserRole]*gaia.UserRoleDetails{
			"RoleA": {},
			"RoleB": {},
		},
	},
	"CategoryB": {
		Roles: map[gaia.UserRole]*gaia.UserRoleDetails{
			"RoleA": {},
			"RoleB": {},
		},
	},
}

func TestFlatRoleName(t *testing.T) {
	value := FullUserRoleName("CategoryA", "RoleA")
	expect := "CategoryARoleA"
	if value != expect {
		t.Fatalf("value %s should equal: %s", value, expect)
	}
}

func TestFlattenUserCategoryRoles(t *testing.T) {
	value := FlattenUserCategoryRoles(mockData)
	expect := []string{"CategoryARoleA", "CategoryARoleB", "CategoryBRoleA", "CategoryBRoleB"}

	for i := range expect {
		if expect[i] != value[i] {
			t.Fatalf("value %s should exist: %s", expect[i], value[i])
		}
	}
}
