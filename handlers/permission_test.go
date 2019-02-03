package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gaia-pipeline/gaia/handlers/mocks"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"

	"github.com/labstack/echo"
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

func TestUserPutPermissions(t *testing.T) {
	ms := &mocks.Store{
		UserPermissionsPutFunc: func(perms *gaia.UserPermission) error {
			return nil
		},
	}

	services.MockStorageService(ms)

	bts, _ := json.Marshal(&gaia.UserPermission{
		Username: "test-user",
		Roles:    []string{"TestRole"},
		Groups:   []string{},
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(bts))
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/user/:username/permissions")
	c.SetParamNames("username")
	c.SetParamValues("test-user")
	_ = UserPutPermissions(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("code is %d. expected %d", rec.Code, http.StatusOK)
	}
}

func TestUserPutPermissionsError(t *testing.T) {
	ms := &mocks.Store{
		UserPermissionsPutFunc: func(perms *gaia.UserPermission) error {
			return errors.New("test error")
		},
	}

	services.MockStorageService(ms)

	bts, _ := json.Marshal(&gaia.UserPermission{
		Username: "test-user",
		Roles:    []string{"TestRole"},
		Groups:   []string{},
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(bts))
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/user/:username/permissions")
	c.SetParamNames("username")
	c.SetParamValues("test-user")
	_ = UserPutPermissions(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code is %d. expected %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUserGetPermissions(t *testing.T) {
	ms := &mocks.Store{
		UserPermissionsGetFunc: func(username string) (*gaia.UserPermission, error) {
			return &gaia.UserPermission{
				Username: "test-user",
				Roles:    []string{"TestRole"},
				Groups:   []string{},
			}, nil
		},
	}

	services.MockStorageService(ms)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/user/:username/permissions")
	c.SetParamNames("username")
	c.SetParamValues("test-user")
	_ = UserGetPermissions(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("code is %d. expected %d", rec.Code, http.StatusOK)
	}
}

func TestUserGetPermissionsErrors(t *testing.T) {
	ms := &mocks.Store{
		UserPermissionsGetFunc: func(username string) (*gaia.UserPermission, error) {
			return nil, errors.New("test error")
		},
	}

	services.MockStorageService(ms)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/user/:username/permissions")
	c.SetParamNames("username")
	c.SetParamValues("test-user")
	_ = UserGetPermissions(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code is %d. expected %d", rec.Code, http.StatusBadRequest)
	}
}
