package rolehelper

import (
	"testing"

	"github.com/gaia-pipeline/gaia"
)

var mockData = []*gaia.UserRoleCategory{
	{
		Name: "CategoryA",
		Roles: []*gaia.UserRole{
			{
				Name: "RoleA",
			},
			{
				Name: "RoleB",
			},
		},
	},
	{
		Name: "CategoryB",
		Roles: []*gaia.UserRole{
			{
				Name: "RoleA",
			},
			{
				Name: "RoleB",
			},
		},
	},
}

func TestFlatRoleName(t *testing.T) {
	value := FullUserRoleName(mockData[0], mockData[0].Roles[0])
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
