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

	jwt "github.com/dgrijalva/jwt-go"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/gaia-pipeline/gaia"
	gStore "github.com/gaia-pipeline/gaia/store"
)

type mockUserStorageService struct {
	gStore.GaiaStore
	user *gaia.User
	err  error
}

func (m *mockUserStorageService) UserAuth(u *gaia.User, updateLastLogin bool) (*gaia.User, error) {
	return m.user, m.err
}

func (m *mockUserStorageService) UserGet(username string) (*gaia.User, error) {
	return m.user, m.err
}

func (m *mockUserStorageService) UserPut(u *gaia.User, encryptPassword bool) error {
	return nil
}

func (m *mockUserStorageService) UserPermissionsGet(username string) (*gaia.UserPermission, error) {
	return &gaia.UserPermission{}, nil
}

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
		Mode:     gaia.ModeServer,
	}

	e := echo.New()

	body := map[string]string{
		"username": "admin",
		"password": "admin",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ms := &mockUserStorageService{user: &gaia.User{
		Username: "username",
		Password: "password",
	}, err: nil}

	handlers := NewUserHandler(ms, nil)
	if err := handlers.UserLogin(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}

	data, err := ioutil.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	user := &gaia.User{}
	err = json.Unmarshal(data, user)
	if err != nil {
		t.Fatalf("error unmarshaling response %v", err.Error())
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

func TestDeleteUserNotAllowedForAutoUser(t *testing.T) {
	dataDir, _ := ioutil.TempDir("", "TestDeleteUserNotAllowedForAutoUser")

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		Logger:    hclog.NewNullLogger(),
		DataPath:  dataDir,
		CAPath:    dataDir,
		VaultPath: dataDir,
	}

	e := echo.New()
	req := httptest.NewRequest(echo.DELETE, "/api/"+gaia.APIVersion+"/user/auto", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlers := NewUserHandler(nil, nil)
	_ = handlers.UserDelete(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
	}
}

func TestResetAutoUserTriggerToken(t *testing.T) {
	dataDir, _ := ioutil.TempDir("", "TestResetAutoUserTriggerToken")

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		Logger:    hclog.NewNullLogger(),
		DataPath:  dataDir,
		CAPath:    dataDir,
		VaultPath: dataDir,
	}

	t.Run("reset auto user token", func(t *testing.T) {
		user := gaia.User{}
		user.Username = "auto"
		user.TriggerToken = "triggerToken"
		ms := &mockUserStorageService{user: &user, err: nil}
		e := echo.New()
		req := httptest.NewRequest(echo.PUT, "/api/"+gaia.APIVersion+"/user/auto/reset-trigger-token", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("username")
		c.SetParamValues("auto")

		handlers := NewUserHandler(ms, nil)
		_ = handlers.UserResetTriggerToken(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v; error: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		if user.TriggerToken == "triggerToken" {
			t.Fatal("user's trigger token should have been reset")
		}
	})
	t.Run("only auto user can reset trigger token", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(echo.PUT, "/api/"+gaia.APIVersion+"/user/auto2/reset-trigger-token", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("username")
		c.SetParamValues("auto2")

		handlers := NewUserHandler(nil, nil)
		_ = handlers.UserResetTriggerToken(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v; error: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})
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
		Mode:     gaia.ModeServer,
	}
	ms := &mockUserStorageService{user: &gaia.User{
		Username: "username",
		Password: "password",
	}, err: nil}

	e := echo.New()

	body := map[string]string{
		"username": "admin",
		"password": "admin",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlers := NewUserHandler(ms, nil)
	if err := handlers.UserLogin(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}

	data, err := ioutil.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	user := &gaia.User{}
	err = json.Unmarshal(data, user)
	if err != nil {
		t.Fatalf("error unmarshaling response %v", err.Error())
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

	handlers := NewUserHandler(ms, nil)
	_ = handlers.UserPutPermissions(c)

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

	handlers := NewUserHandler(ms, nil)
	_ = handlers.UserPutPermissions(c)

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

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/user/:username/permissions")
	c.SetParamNames("username")
	c.SetParamValues("test-user")

	handlers := NewUserHandler(ms, nil)
	_ = handlers.UserGetPermissions(c)

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

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/user/:username/permissions")
	c.SetParamNames("username")
	c.SetParamValues("test-user")

	handlers := NewUserHandler(ms, nil)
	_ = handlers.UserGetPermissions(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code is %d. expected %d", rec.Code, http.StatusBadRequest)
	}
}
