package handlers

import (
	"testing"

	"net/http/httptest"

	"net/http"

	"crypto/rsa"
	"time"

	"fmt"
	"io/ioutil"

	"crypto/rand"

	"github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
)

func makeAuthBarrierRouter() *echo.Echo {
	e := echo.New()
	r := e.Router()
	e.Use(authBarrier)

	r.Add(echo.GET, "/auth", func(c echo.Context) error {
		return c.NoContent(200)
	})

	return e
}

func TestAuthBarrierNoToken(t *testing.T) {
	e := makeAuthBarrierRouter()

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected response code %v got %v", http.StatusForbidden, rec.Code)
	}
}

func TestAuthBarrierBadHeader(t *testing.T) {
	e := makeAuthBarrierRouter()

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	req.Header.Set("Authorization", "my-token")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected response code %v got %v", http.StatusForbidden, rec.Code)
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
		"test-user",
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
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

	claims := jwtCustomClaims{
		"test-user",
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + jwtExpiry,
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

	claims := jwtCustomClaims{
		"test-user",
		jwt.StandardClaims{
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

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected response code %v got %v", http.StatusForbidden, rec.Code)
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
		"test-user",
		jwt.StandardClaims{
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

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected response code %v got %v", http.StatusForbidden, rec.Code)
	}

	bodyBytes, _ := ioutil.ReadAll(rec.Body)
	body := string(bodyBytes)

	signingMethodError := fmt.Sprintf("unexpected signing method: %v", token.Header["alg"])
	if body != signingMethodError {
		t.Fatalf("expected body '%v' got '%v'", signingMethodError, body)
	}
}
