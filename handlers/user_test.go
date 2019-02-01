package handlers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gaia-pipeline/gaia/services"
	gStore "github.com/gaia-pipeline/gaia/store"
	"github.com/pkg/errors"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
)

func TestUserLoginHMACKey(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestUserLoginHMACKey")
	dataDir := tmp

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: []byte("hmac-jwt-key"),
		Logger: hclog.New(&hclog.LoggerOptions{
			Level:  hclog.Trace,
			Output: hclog.DefaultOutput,
			Name:   "Gaia",
		}),
		DataPath: dataDir,
	}

	e := echo.New()
	InitHandlers(e)

	body := map[string]string{
		"username": "admin",
		"password": "admin",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}

	data, err := ioutil.ReadAll(rec.Body)
	user := &gaia.User{}
	err = json.Unmarshal(data, user)
	if err != nil {
		t.Fatalf("error unmarshaling responce %v", err.Error())
	}
	token, _, err := new(jwt.Parser).ParseUnverified(user.Tokenstring, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("error parsing the token %v", err.Error())
	}
	alg := "HS256"
	if token.Header["alg"] != alg {
		t.Fatalf("expected token alg %v got %v", alg, token.Header["alg"])
	}

}

func TestUserLoginRSAKey(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestUserLoginRSAKey")
	dataDir := tmp

	defer func() {
		gaia.Cfg = nil
	}()

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	gaia.Cfg = &gaia.Config{
		JWTKey: key,
		Logger: hclog.New(&hclog.LoggerOptions{
			Level:  hclog.Trace,
			Output: hclog.DefaultOutput,
			Name:   "Gaia",
		}),
		DataPath: dataDir,
	}

	e := echo.New()
	InitHandlers(e)

	body := map[string]string{
		"username": "admin",
		"password": "admin",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}

	data, err := ioutil.ReadAll(rec.Body)
	user := &gaia.User{}
	err = json.Unmarshal(data, user)
	if err != nil {
		t.Fatalf("error unmarshaling responce %v", err.Error())
	}
	token, _, err := new(jwt.Parser).ParseUnverified(user.Tokenstring, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("error parsing the token %v", err.Error())
	}
	alg := "RS512"
	if token.Header["alg"] != alg {
		t.Fatalf("expected token alg %v got %v", alg, token.Header["alg"])
	}
}

type mockStore struct {
	gStore.GaiaStore
	userPermissionsGetFunc func(username string) (*gaia.UserPermission, error)
	userPermissionsPutFunc func(perms *gaia.UserPermission) error
}

func (s *mockStore) UserPermissionsGet(username string) (*gaia.UserPermission, error) {
	return s.userPermissionsGetFunc(username)
}

func (s *mockStore) UserPermissionsPut(perms *gaia.UserPermission) error {
	return s.userPermissionsPutFunc(perms)
}

func TestUserPutPermissions(t *testing.T) {
	ms := &mockStore{
		userPermissionsPutFunc: func(perms *gaia.UserPermission) error {
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
	ms := &mockStore{
		userPermissionsPutFunc: func(perms *gaia.UserPermission) error {
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
	ms := &mockStore{
		userPermissionsGetFunc: func(username string) (*gaia.UserPermission, error) {
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
	ms := &mockStore{
		userPermissionsGetFunc: func(username string) (*gaia.UserPermission, error) {
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
