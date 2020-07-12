package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/gaia-pipeline/gaia/security/rbac"
)

type mockEchoEnforcer struct{}

func (m *mockEchoEnforcer) Enforce(username, method, path string, params map[string]string) error {
	if username == "enforcer-perms-err" {
		return rbac.NewErrPermissionDenied("namespace", "action", "thing")
	}
	if username == "enforcer-err" {
		return errors.New("error")
	}
	return nil
}

var mockRoleAuth = &AuthConfig{
	RoleCategories: []*gaia.UserRoleCategory{
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
	},
	rbacEnforcer: &mockEchoEnforcer{},
}

func makeAuthBarrierRouter() *echo.Echo {
	e := echo.New()
	e.Use(authMiddleware(mockRoleAuth))

	success := func(c echo.Context) error {
		return c.NoContent(200)
	}

	e.GET("/auth", success)
	e.GET("/catone/:test", success)
	e.GET("/catone/latest", success)
	e.POST("/catone", success)
	e.POST("/enforcer/test", success)

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
	}

	claims := gaia.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + gaia.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
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
	}

	claims := gaia.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + gaia.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
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
	}

	claims := gaia.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + gaia.JwtExpiry,
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
	}

	claims := gaia.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + gaia.JwtExpiry,
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
	}

	claims := gaia.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + gaia.JwtExpiry,
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
	}

	claims := gaia.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{"CatOneGetSingle", "CatOnePostSingle", "CatTwoGetSingle", "CatTwoPostSingle"},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + gaia.JwtExpiry,
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

func Test_AuthMiddleware_Enforcer_PermissionDenied(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: []byte("hmac-jwt-key"),
	}

	claims := gaia.JwtCustomClaims{
		Username: "enforcer-perms-err",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + gaia.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(gaia.Cfg.JWTKey)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.POST, "/enforcer/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)
	e.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, http.StatusForbidden)
	assert.Equal(t, rec.Body.String(), "Permission denied. Must have namespace/action thing")
}

func Test_AuthMiddleware_Enforcer_UnknownError(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: []byte("hmac-jwt-key"),
	}
	gaia.Cfg.Logger = hclog.NewNullLogger()

	claims := gaia.JwtCustomClaims{
		Username: "enforcer-err",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + gaia.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(gaia.Cfg.JWTKey)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.POST, "/enforcer/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)
	e.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, http.StatusInternalServerError)
	assert.Equal(t, rec.Body.String(), "Unknown error has occurred while validating permissions.")
}
