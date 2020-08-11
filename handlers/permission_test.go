package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
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
