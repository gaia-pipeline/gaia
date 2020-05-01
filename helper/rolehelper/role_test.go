package rolehelper

import (
	"sort"
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
	sort.Slice(value, func(i, j int) bool {
		return value[i] < value[j]
	})
	sort.Slice(expect, func(i, j int) bool {
		return expect[i] < expect[j]
	})
	for i := range expect {
		if expect[i] != value[i] {
			t.Fatalf("value %s should exist: %s", expect[i], value[i])
		}
	}
}
