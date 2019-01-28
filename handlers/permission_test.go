package handlers

import (
	"github.com/labstack/echo"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPermissionGetAll(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/permission")
	err := PermissionGetAll(c)

	if err != nil {
		t.Fatal("should not error")
	}
	if rec.Code != http.StatusOK {
		t.Fatal("code should be 200")
	}
}
