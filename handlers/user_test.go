package handlers

import (
	"testing"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"io/ioutil"

	"crypto/rand"
	"crypto/rsa"

	jwt "github.com/dgrijalva/jwt-go"
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
	}

	e := echo.New()
	InitHandlers(e)

	body := map[string]string{
		"username": "admin",
		"password": "admin",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/login", bytes.NewBuffer(bodyBytes))
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
	req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/login", bytes.NewBuffer(bodyBytes))
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
