package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/services"
	gStore "github.com/gaia-pipeline/gaia/store"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
)

type mockStorageService struct {
	worker gaia.Worker
	gStore.GaiaStore
}

func (m *mockStorageService) WorkerPut(worker *gaia.Worker) error {
	m.worker = *worker
	return nil
}
func (m *mockStorageService) WorkerGet(id string) (*gaia.Worker, error) {
	return &m.worker, nil
}

func TestRegisterWorker(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestRegisterWorker")

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     tmp,
		HomePath:     tmp,
		PipelinePath: tmp,
		DevMode:      true,
	}

	// Initialize store
	m := &mockStorageService{}
	services.MockStorageService(m)
	dataStore, _ := services.StorageService()

	// Initialize certificate store
	_, err := services.CertificateService()
	if err != nil {
		t.Fatalf("cannot initialize certificate service: %v", err)
	}

	// Initialize vault
	v, err := services.VaultService(nil)
	if err != nil {
		t.Fatalf("cannot initialize vault service: %v", err)
	}

	// Initialize memdb service
	db, err := services.MemDBService(dataStore)
	if err != nil {
		t.Fatal(err)
	}

	// Generate global worker secret
	secret := []byte(security.GenerateRandomUUIDV5())
	v.Add(gaia.WorkerRegisterKey, secret)
	if err := v.SaveSecrets(); err != nil {
		t.Fatal(err)
	}

	// Initialize echo
	e := echo.New()
	if err := InitHandlers(e); err != nil {
		t.Fatal(err)
	}

	// Test with wrong global secret
	t.Run("wrong global secret", func(t *testing.T) {
		body := registerWorker{
			Secret: "random-wrong-secret",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/worker/register", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := RegisterWorker(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected response code %v got %v", http.StatusForbidden, rec.Code)
		}
		bodyBytes, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		if string(bodyBytes[:]) != "wrong global worker secret provided" {
			t.Fatal("return message is not correct")
		}
	})

	workerName := "my-worker"
	t.Run("register worker success", func(t *testing.T) {
		body := registerWorker{
			Name:   workerName,
			Secret: string(secret[:]),
			Tags:   []string{"first-tag", "second-tag", "third-tag"},
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/worker/register", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := RegisterWorker(c); err != nil {
			t.Fatal(err)
		}

		bodyBytes, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v; body: %s", http.StatusOK, rec.Code, string(bodyBytes[:]))
		}
		resp := &registerResponse{}
		if err := json.Unmarshal(bodyBytes, resp); err != nil {
			t.Fatalf("failed to unmarshal response: %#v", bodyBytes)
		}

		if resp.UniqueID == "" {
			t.Fatal("unique id should be set but got empty string")
		}
		if resp.CACert == "" {
			t.Fatal("ca cert should be set but got empty string")
		}
		if resp.Key == "" {
			t.Fatal("key cert should be set but got empty string")
		}
		if resp.Cert == "" {
			t.Fatal("cert should be set but got empty string")
		}

		// Check if store holds the new registered worker
		worker, err := dataStore.WorkerGet(resp.UniqueID)
		if err != nil {
			t.Fatal(err)
		}
		if worker == nil {
			t.Fatal("failed to get worker from store. It was nil.")
		}

		// Check if memdb service holds the data
		worker, err = db.GetWorker(resp.UniqueID)
		if err != nil {
			t.Fatal(err)
		}
		if worker == nil {
			t.Fatal("failed to get worker from memdb cache. It was nil.")
		}
	})
}
