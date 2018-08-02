package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
)

func TestVaultWorkflowAddListDelete(t *testing.T) {
	dataDir, err := ioutil.TempDir("", "temp")
	if err != nil {
		t.Fatalf("error creating data dir %v", err.Error())
	}

	defer func() {
		gaia.Cfg = nil
		os.RemoveAll(dataDir)
	}()

	gaia.Cfg = &gaia.Config{
		Logger:    hclog.NewNullLogger(),
		DataPath:  dataDir,
		CAPath:    dataDir,
		VaultPath: dataDir,
	}

	dataStore, _ := services.StorageService()
	err = dataStore.Init()
	if err != nil {
		t.Fatalf("cannot initialize store: %v", err.Error())
	}

	_, err = services.CertificateService()
	if err != nil {
		t.Fatalf("cannot initialize certificate service: %v", err.Error())
	}

	e := echo.New()
	InitHandlers(e)
	t.Run("can add secret", func(t *testing.T) {
		body := map[string]string{
			"Key":   "Key",
			"Value": "Value",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/secret", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		SetSecret(c)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected response code %v got %v", http.StatusCreated, rec.Code)
		}
	})

	t.Run("can update secret", func(t *testing.T) {
		body := map[string]string{
			"Key":   "Key",
			"Value": "Value",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.PUT, "/api/"+apiVersion+"/secret", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		SetSecret(c)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected response code %v got %v", http.StatusCreated, rec.Code)
		}
	})

	t.Run("can list secrets", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/api/"+apiVersion+"/secrets", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		ListSecrets(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusCreated, rec.Code)
		}
		body, _ := ioutil.ReadAll(rec.Body)
		expectedBody := "[{\"key\":\"Key\",\"value\":\"**********\"}]"
		if string(body) != expectedBody {
			t.Fatalf("response body did not equal expected body: expected: %s, actual: %s", expectedBody, string(body))
		}
	})

	t.Run("can delete secrets", func(t *testing.T) {
		req := httptest.NewRequest(echo.DELETE, "/api/"+apiVersion+"/secret/:key", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("key")
		c.SetParamValues("Key")

		RemoveSecret(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusCreated, rec.Code)
		}
	})

	t.Run("can delete fails if no secret is provided", func(t *testing.T) {
		req := httptest.NewRequest(echo.DELETE, "/api/"+apiVersion+"/secret/:key", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		RemoveSecret(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusCreated, rec.Code)
		}
	})
}
