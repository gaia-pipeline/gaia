package handlers

import (
	"github.com/gaia-pipeline/gaia"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestPermissionGetAll(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/permission")

	ph := permissionHandler{
		defaultRoles: map[gaia.UserRoleCategory]*gaia.UserRoleCategoryDetails{
			"TestCategory": {
				Description: "Description...",
				Roles: map[gaia.UserRole]*gaia.UserRoleDetails{
					"TestCategoryA": {
						Description: "Description...",
					},
					"TestCategoryB": {
						Description: "Description...",
					},
				},
			},
		},
	}
	err := ph.PermissionGetAll(c)

	expected := "[{\"name\":\"TestCategory\",\"description\":\"Description...\",\"roles\":[{\"name\":\"TestCategoryA\",\"description\":\"Description...\"},{\"name\":\"TestCategoryB\",\"description\":\"Description...\"}]}]\n"

	if err != nil {
		t.Fatal("should not error")
	}
	if rec.Code != http.StatusOK {
		t.Fatal("code should be 200")
	}

	actual := rec.Body.String()
	if actual != expected {
		t.Fatalf("expected actual %s to equal %s", actual, expected)
	}
}
