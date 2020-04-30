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

	"github.com/gaia-pipeline/gaia/workers/pipeline"

	"github.com/gaia-pipeline/gaia/services"
	gStore "github.com/gaia-pipeline/gaia/store"

	"github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia"
	"github.com/hashicorp/go-hclog"
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
		Mode:     gaia.ModeServer,
	}

	pipelineService := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	e := echo.New()
	_ = handlerService.InitHandlers(e)

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

	_, err := services.CertificateService()
	if err != nil {
		t.Fatalf("cannot initialize certificate service: %v", err.Error())
	}

	pipelineService := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	e := echo.New()
	_ = handlerService.InitHandlers(e)
	req := httptest.NewRequest(echo.DELETE, "/api/"+gaia.APIVersion+"/user/auto", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = UserDelete(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
	}
}

type mockUserStorageService struct {
	gStore.GaiaStore
	user *gaia.User
	err  error
}

func (m mockUserStorageService) UserGet(username string) (*gaia.User, error) {
	return m.user, m.err
}

func (m mockUserStorageService) UserPut(u *gaia.User, encryptPassword bool) error {
	return nil
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

	pipelineService := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	_, err := services.CertificateService()
	if err != nil {
		t.Fatalf("cannot initialize certificate service: %v", err.Error())
	}
	t.Run("reset auto user token", func(t *testing.T) {
		user := gaia.User{}
		user.Username = "auto"
		user.TriggerToken = "triggerToken"
		m := mockUserStorageService{user: &user, err: nil}
		services.MockStorageService(&m)
		defer services.MockStorageService(nil)
		e := echo.New()
		_ = handlerService.InitHandlers(e)
		req := httptest.NewRequest(echo.PUT, "/api/"+gaia.APIVersion+"/user/auto/reset-trigger-token", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("username")
		c.SetParamValues("auto")

		_ = UserResetTriggerToken(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v; error: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		if user.TriggerToken == "triggerToken" {
			t.Fatal("user's trigger token should have been reset")
		}
	})
	t.Run("only auto user can reset trigger token", func(t *testing.T) {
		e := echo.New()
		_ = handlerService.InitHandlers(e)
		req := httptest.NewRequest(echo.PUT, "/api/"+gaia.APIVersion+"/user/auto2/reset-trigger-token", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("username")
		c.SetParamValues("auto2")

		_ = UserResetTriggerToken(c)

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

	pipelineService := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	e := echo.New()
	_ = handlerService.InitHandlers(e)

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

type mockPermissionsUserStoreService struct {
	gStore.GaiaStore
	ua   func(u *gaia.User, updateLastLogin bool) (*gaia.User, error)
	upg  func(username string) (*gaia.UserPermission, error)
	upgg func(name string) (*gaia.UserPermissionGroup, error)
}

func (m mockPermissionsUserStoreService) UserPermissionsGet(username string) (*gaia.UserPermission, error) {
	return m.upg(username)
}

func (m mockPermissionsUserStoreService) UserPermissionGroupGet(name string) (*gaia.UserPermissionGroup, error) {
	return m.upgg(name)
}

func (m mockPermissionsUserStoreService) UserAuth(u *gaia.User, updateLastLogin bool) (*gaia.User, error) {
	return m.ua(u, updateLastLogin)
}

func TestUserLoginPerms(t *testing.T) {
	// Create a mock service which returns new and conflicting roles on purpose. All should merge together so there is
	// no duplicates.
	storeService := mockPermissionsUserStoreService{
		upgg: func(name string) (*gaia.UserPermissionGroup, error) {
			return &gaia.UserPermissionGroup{
				Name:  "TestGroup",
				Roles: []string{"TestRoleA", "TestRoleB", "TestRoleC", "TestRoleE"},
			}, nil
		},
		upg: func(username string) (*gaia.UserPermission, error) {
			return &gaia.UserPermission{
				Username: "admin",
				Roles:    []string{"TestRoleA", "TestRoleB", "TestRoleD", "TestRoleF"},
				Groups:   []string{"TestGroup"},
			}, nil
		},
		ua: func(u *gaia.User, updateLastLogin bool) (*gaia.User, error) {
			return &gaia.User{
				Username: "admin",
			}, nil
		},
	}

	tmp, _ := ioutil.TempDir("", "TestUserLoginHMACKey")
	dataDir := tmp

	services.MockStorageService(&storeService)
	defer func() {
		services.MockStorageService(nil)
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

	pipelineService := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	e := echo.New()
	_ = handlerService.InitHandlers(e)

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
		t.Fatalf("error unmarshaling response %v", err.Error())
	}
	token, _, err := new(jwt.Parser).ParseUnverified(user.Tokenstring, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("error parsing the token %v", err.Error())
	}

	roles := token.Claims.(jwt.MapClaims)["roles"].([]interface{})

	expected := []string{"TestRoleA", "TestRoleB", "TestRoleD", "TestRoleF", "TestRoleC", "TestRoleE"}
	for i := range roles {
		role := roles[i].(string)
		if expected[i] != role {
			t.Fatalf("value %s should exist: %s", expected[i], role)
		}
	}
}
