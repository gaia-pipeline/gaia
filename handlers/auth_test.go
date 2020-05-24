package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/go-hclog"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/gaia-pipeline/gaia/security/rbac"
	"github.com/labstack/echo"
)

var mockRoleCategories = []*gaia.UserRoleCategory{
	{
		Name: "CatOne",
		Roles: []*gaia.UserRole{
			{
				Name: "GetSingle",
				APIEndpoint: []*gaia.UserRoleEndpoint{
					rolehelper.NewUserRoleEndpoint("GET", "/catone/:id"),
					rolehelper.NewUserRoleEndpoint("GET", "/catone/latest"),
				},
			},
			{
				Name: "PostSingle",
				APIEndpoint: []*gaia.UserRoleEndpoint{
					rolehelper.NewUserRoleEndpoint("POST", "/catone"),
				},
			},
		},
	},
	{
		Name: "CatTwo",
		Roles: []*gaia.UserRole{
			{
				Name: "GetSingle",
				APIEndpoint: []*gaia.UserRoleEndpoint{
					rolehelper.NewUserRoleEndpoint("GET", "/cattwo/:first/:second"),
				},
			},
			{
				Name: "PostSingle",
				APIEndpoint: []*gaia.UserRoleEndpoint{
					rolehelper.NewUserRoleEndpoint("POST", "/cattwo/:first/:second/start"),
				},
			},
		},
	},
}

type mockEnforcer struct {
}

func (m mockEnforcer) Enforce(cfg rbac.EnforcerConfig) error {
	if cfg.Resource == "test/key/123" {
		return errors.New("key enforce test err")
	}
	return nil
}

func (m mockEnforcer) Evaluate(user rbac.User) (gaia.RBACEvaluatedPermissions, error) {
	panic("implement me")
}

func (m mockEnforcer) GetDefaultAPIGroup() gaia.RBACAPIGroup {
	return gaia.RBACAPIGroup{
		Endpoints: map[string]gaia.RBACAPIGroupEndpoint{
			"/test/withResource/:key": {
				Param: "key",
				Methods: map[string]string{
					"GET": "test/get",
				},
			},
			"/test/withoutResource": {
				Methods: map[string]string{
					"POST": "test/create",
				},
			},
		},
	}
}

func makeAuthBarrierRouter() *echo.Echo {
	e := echo.New()
	authMw := AuthMiddleware{
		RoleCategories: mockRoleCategories,
		rbacEnforcer:   &mockEnforcer{},
	}
	e.Use(authMw.Do())

	success := func(c echo.Context) error {
		return c.NoContent(200)
	}

	e.GET("/auth", success)
	e.GET("/catone/:test", success)
	e.GET("/catone/latest", success)
	e.POST("/catone", success)

	// RBAC
	e.GET("/test/withResource/:key", success)
	e.POST("/test/withoutResource", success)

	return e
}

func TestAuthBarrierNoToken(t *testing.T) {
	e := makeAuthBarrierRouter()

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected response code %v got %v", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthBarrierBadHeader(t *testing.T) {
	e := makeAuthBarrierRouter()

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	req.Header.Set("Authorization", "my-token")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected response code %v got %v", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthBarrierHMACTokenWithHMACKey(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: []byte("hmac-jwt-key"),
		Logger: hclog.NewNullLogger(),
	}

	claims := jwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
		Policies: map[string]interface{}{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(gaia.Cfg.JWTKey)

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}
}

func TestAuthBarrierRSATokenWithRSAKey(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	gaia.Cfg = &gaia.Config{
		JWTKey: key,
		Logger: hclog.NewNullLogger(),
	}

	claims := jwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
		Policies: map[string]interface{}{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	tokenstring, _ := token.SignedString(gaia.Cfg.JWTKey)

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}
}

func TestAuthBarrierHMACTokenWithRSAKey(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: &rsa.PrivateKey{},
		Logger: hclog.NewNullLogger(),
	}

	claims := jwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString([]byte("hmac-jwt-key"))

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected response code %v got %v", http.StatusUnauthorized, rec.Code)
	}

	bodyBytes, _ := ioutil.ReadAll(rec.Body)
	body := string(bodyBytes)

	signingMethodError := fmt.Sprintf("unexpected signing method: %v", token.Header["alg"])
	if body != signingMethodError {
		t.Fatalf("expected body '%v' got '%v'", signingMethodError, body)
	}
}

func TestAuthBarrierRSATokenWithHMACKey(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: []byte("hmac-jwt-key"),
		Logger: hclog.NewNullLogger(),
	}

	claims := jwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tokenstring, _ := token.SignedString(key)

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected response code %v got %v", http.StatusUnauthorized, rec.Code)
	}

	bodyBytes, _ := ioutil.ReadAll(rec.Body)
	body := string(bodyBytes)

	signingMethodError := fmt.Sprintf("unexpected signing method: %v", token.Header["alg"])
	if body != signingMethodError {
		t.Fatalf("expected body '%v' got '%v'", signingMethodError, body)
	}
}

var roleTests = []struct {
	perm   string
	method string
	path   string
}{
	{"CatOneGetSingle", "GET", "/catone/:id"},
	{"CatOneGetSingle", "GET", "/catone/latest"},
	{"CatOnePostSingle", "POST", "/catone"},
	{"CatTwoGetSingle", "POST", "/cattwo/:first/:second"},
	{"CatTwoPostSingle", "POST", "/cattwo/:first/:second/start"},
}

func TestAuthBarrierNoPerms(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: []byte("hmac-jwt-key"),
		Logger: hclog.NewNullLogger(),
	}

	claims := jwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(gaia.Cfg.JWTKey)

	for _, tt := range roleTests {
		t.Run(tt.perm, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(echo.POST, "/catone", nil)
			req.Header.Set("Authorization", "Bearer "+tokenstring)
			e.ServeHTTP(rec, req)
			testPermFailed(t, tt.perm, rec.Code, rec.Body.String())
		})
	}
}

func TestAuthBarrierAllPerms(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: []byte("hmac-jwt-key"),
		Logger: hclog.NewNullLogger(),
	}

	claims := jwtCustomClaims{
		Username: "test-user",
		Roles:    []string{"CatOneGetSingle", "CatOnePostSingle", "CatTwoGetSingle", "CatTwoPostSingle"},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
		Policies: map[string]interface{}{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(gaia.Cfg.JWTKey)

	for _, tt := range roleTests {
		t.Run(tt.perm, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(echo.POST, "/catone", nil)
			req.Header.Set("Authorization", "Bearer "+tokenstring)
			e.ServeHTTP(rec, req)
			testPermSuccess(t, rec.Code, rec.Body.String())
		})
	}
}

func testPermFailed(t *testing.T, perm string, statusCode int, body string) {
	if body == "" {
		t.Fatalf("expected response body %v got %v", "Permission denied for user "+perm+". Required permission "+perm, body)
	}
	if statusCode != http.StatusForbidden {
		t.Fatalf("expected response code %v got %v", http.StatusForbidden, statusCode)
	}
}

func testPermSuccess(t *testing.T, statusCode int, body string) {
	if body != "" {
		t.Fatalf("expected response body %v got %v", "", body)
	}
	if statusCode != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, statusCode)
	}
}

func TestAuthMiddleware_Do_RBACEnforcer_WithResource_EnforcerError(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: []byte("hmac-jwt-key"),
	}
	gaia.Cfg.Logger = hclog.NewNullLogger()

	claims := jwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		Policies: map[string]interface{}{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(gaia.Cfg.JWTKey)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.GET, "/test/withResource/123", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)
	e.ServeHTTP(rec, req)

	assert.Check(t, cmp.Equal(rec.Code, 403))
	assert.Check(t, cmp.Equal(rec.Body.String(), "Permission denied for user test-user: missing required permission test/get"))
}

func TestAuthMiddleware_Do_RBACEnforcer(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: []byte("hmac-jwt-key"),
	}
	gaia.Cfg.Logger = hclog.NewNullLogger()

	claims := jwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		Policies: map[string]interface{}{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(gaia.Cfg.JWTKey)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.POST, "/test/withoutResource", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)
	e.ServeHTTP(rec, req)

	assert.Check(t, cmp.Equal(rec.Code, 200))
	assert.Check(t, cmp.Equal(rec.Body.String(), ""))
}
