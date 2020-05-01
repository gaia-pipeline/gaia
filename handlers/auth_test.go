package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
)

func makeAuthBarrierRouter() *echo.Echo {
	e := echo.New()
	e.Use(AuthMiddleware())

	success := func(c echo.Context) error {
		return c.NoContent(200)
	}

	e.GET("/no-rbac", success)
	rbac := NewRBACMiddleware(rolehelper.PipelineCategory)
	e.GET("/rbac", success, rbac.Do(rolehelper.GetRole))

	return e
}

func TestAuthBarrierNoToken(t *testing.T) {
	e := makeAuthBarrierRouter()

	req := httptest.NewRequest(echo.GET, "/no-rbac", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected response code %v got %v", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthBarrierBadHeader(t *testing.T) {
	e := makeAuthBarrierRouter()

	req := httptest.NewRequest(echo.GET, "/no-rbac", nil)
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

	req := httptest.NewRequest(echo.GET, "/no-rbac", nil)
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
	tokenstring, _ := token.SignedString(gaia.Cfg.JWTKey)

	req := httptest.NewRequest(echo.GET, "/no-rbac", nil)
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

	req := httptest.NewRequest(echo.GET, "/no-rbac", nil)
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

	req := httptest.NewRequest(echo.GET, "/no-rbac", nil)
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

func TestAuthBarrierNoPerms(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: []byte("hmac-jwt-key"),
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

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.GET, "/rbac", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)
	e.ServeHTTP(rec, req)
	testPermFailed(t, "PipelineGet", rec.Code, rec.Body.String())
}

func TestAuthBarrierAllPerms(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		JWTKey: []byte("hmac-jwt-key"),
	}

	claims := jwtCustomClaims{
		Username: "test-user",
		Roles:    []string{"PipelineGet"},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Gaia Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(gaia.Cfg.JWTKey)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.GET, "/rbac", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)
	e.ServeHTTP(rec, req)
	testPermSuccess(t, rec.Code, rec.Body.String())
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
