package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	gStore "github.com/gaia-pipeline/gaia/store"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"

	"github.com/labstack/echo"
)

type mockPermissionsStoreService struct {
	gStore.GaiaStore
	upp  func(perms *gaia.UserPermission) error
	upg  func(username string) (*gaia.UserPermission, error)
	upd  func(username string) error
	upgp func(group *gaia.UserPermissionGroup) error
	upgg func(name string) (*gaia.UserPermissionGroup, error)
	upga func() ([]*gaia.UserPermissionGroup, error)
	upgd func(name string) error
}

func (m mockPermissionsStoreService) UserPermissionsPut(perms *gaia.UserPermission) error {
	return m.upp(perms)
}

func (m mockPermissionsStoreService) UserPermissionsGet(username string) (*gaia.UserPermission, error) {
	return m.upg(username)
}

func (m mockPermissionsStoreService) UserPermissionsDelete(username string) error {
	return m.upd(username)
}

func (m mockPermissionsStoreService) UserPermissionGroupPut(group *gaia.UserPermissionGroup) error {
	return m.upgp(group)
}

func (m mockPermissionsStoreService) UserPermissionGroupGet(name string) (*gaia.UserPermissionGroup, error) {
	return m.upgg(name)
}

func (m mockPermissionsStoreService) UserPermissionGroupGetAll() ([]*gaia.UserPermissionGroup, error) {
	return m.upga()
}

func (m mockPermissionsStoreService) UserPermissionGroupDelete(name string) error {
	return m.upgd(name)
}

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
	storeService := mockPermissionsStoreService{
		upp: func(perms *gaia.UserPermission) error {
			return nil
		},
	}

	services.MockStorageService(&storeService)
	defer services.MockStorageService(nil)

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
	storeService := mockPermissionsStoreService{
		upp: func(perms *gaia.UserPermission) error {
			return errors.New("error")
		},
	}

	services.MockStorageService(&storeService)
	defer services.MockStorageService(nil)

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
	t.Run("get test-user success", func(t *testing.T) {
		storeService := mockPermissionsStoreService{
			upg: func(username string) (*gaia.UserPermission, error) {
				return &gaia.UserPermission{
					Username: "test-user",
				}, nil
			},
		}

		services.MockStorageService(&storeService)
		defer services.MockStorageService(nil)

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
	})

	t.Run("get test-user error", func(t *testing.T) {
		storeService := mockPermissionsStoreService{
			upg: func(username string) (*gaia.UserPermission, error) {
				return nil, errors.New("error")
			},
		}

		services.MockStorageService(&storeService)
		defer services.MockStorageService(nil)

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
	})
}

func TestUserGetPermissionsErrors(t *testing.T) {
	storeService := mockPermissionsStoreService{
		upg: func(username string) (*gaia.UserPermission, error) {
			return nil, errors.New("")
		},
	}

	services.MockStorageService(&storeService)
	defer services.MockStorageService(nil)

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
