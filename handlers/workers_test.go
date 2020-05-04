package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/handlers/providers/workers"
	"github.com/gaia-pipeline/gaia/plugin"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/services"
	gStore "github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/workers/scheduler/gaiascheduler"
)

type mockStorageService struct {
	worker gaia.Worker
	gStore.GaiaStore
}

type registerWorker struct {
	Secret string   `json:"secret"`
	Name   string   `json:"name"`
	Tags   []string `json:"tags"`
}

type registerResponse struct {
	UniqueID string `json:"uniqueid"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	CACert   string `json:"cacert"`
}

func (m *mockStorageService) WorkerPut(worker *gaia.Worker) error {
	m.worker = *worker
	return nil
}
func (m *mockStorageService) WorkerGet(id string) (*gaia.Worker, error) {
	return &m.worker, nil
}
func (m *mockStorageService) WorkerDelete(id string) error {
	return nil
}

func TestRegisterWorker(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestRegisterWorker")
	if err != nil {
		t.Fatal(err)
	}

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
	defer func() { services.MockStorageService(nil) }()

	// Initialize certificate store
	ca, err := security.InitCA()
	if err != nil {
		t.Fatalf("cannot initialize certificate service: %v", err)
	}
	wp := workers.NewWorkerProvider(workers.Dependencies{Scheduler: nil, Certificate: ca})
	// Initialize vault
	v, err := services.DefaultVaultService()
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

	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       nil,
		PipelineService: nil,
		Certificate:     ca,
	})
	// Initialize echo
	e := echo.New()
	if err := handlerService.InitHandlers(e); err != nil {
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

		if err := wp.RegisterWorker(c); err != nil {
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

		if err := wp.RegisterWorker(c); err != nil {
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

func TestDeregisterWorker(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestDeregisterWorker")
	if err != nil {
		t.Fatal(err)
	}

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
	defer func() { services.MockStorageService(nil) }()

	// Initialize vault
	v, err := services.DefaultVaultService()
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

	ca, _ := security.InitCA()
	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       nil,
		PipelineService: nil,
		Certificate:     ca,
	})
	// Initialize echo
	e := echo.New()
	if err := handlerService.InitHandlers(e); err != nil {
		t.Fatal(err)
	}
	wp := workers.NewWorkerProvider(workers.Dependencies{Scheduler: nil, Certificate: ca})

	// Test with non-existing worker
	t.Run("non-existing worker", func(t *testing.T) {
		req := httptest.NewRequest(echo.DELETE, "/api/"+gaia.APIVersion+"/worker/:workerid", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("workerid")
		c.SetParamValues("non-existing-id")

		if err := wp.DeregisterWorker(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
		bodyBytes, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		if string(bodyBytes[:]) != "worker is not registered" {
			t.Fatalf("return message is not correct: %s", string(bodyBytes[:]))
		}
	})

	// Deregister worker success
	t.Run("deregister worker success", func(t *testing.T) {
		body := registerWorker{
			Name:   "my-worker",
			Secret: string(secret[:]),
			Tags:   []string{"first-tag", "second-tag", "third-tag"},
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/worker/register", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := wp.RegisterWorker(c); err != nil {
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

		// Setup deregister call
		req = httptest.NewRequest(echo.DELETE, "/api/"+gaia.APIVersion+"/worker/:workerid", nil)
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetParamNames("workerid")
		c.SetParamValues(resp.UniqueID)

		// Deregister worker
		if err := wp.DeregisterWorker(c); err != nil {
			t.Fatal(err)
		}

		// Check if memdb service still holds the data
		worker, err := db.GetWorker(resp.UniqueID)
		if err != nil {
			t.Fatal(err)
		}
		if worker != nil {
			t.Fatal("worker has been deregistered but is still in cache/store")
		}
	})
}

func TestGetWorkerRegisterSecret(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestGetWorkerRegisterSecret")
	if err != nil {
		t.Fatal(err)
	}

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     tmp,
		HomePath:     tmp,
		PipelinePath: tmp,
		DevMode:      true,
	}

	// Initialize vault
	v, err := services.DefaultVaultService()
	if err != nil {
		t.Fatalf("cannot initialize vault service: %v", err)
	}

	// Generate global worker secret
	secret := []byte(security.GenerateRandomUUIDV5())
	v.Add(gaia.WorkerRegisterKey, secret)
	if err := v.SaveSecrets(); err != nil {
		t.Fatal(err)
	}
	ca, _ := security.InitCA()
	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       nil,
		PipelineService: nil,
		Certificate:     ca,
	})
	// Initialize echo
	e := echo.New()
	if err := handlerService.InitHandlers(e); err != nil {
		t.Fatal(err)
	}
	wp := workers.NewWorkerProvider(workers.Dependencies{Scheduler: nil, Certificate: ca})
	// Test get global worker secret
	t.Run("global secret success", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/api/"+gaia.APIVersion+"/worker/secret", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := wp.GetWorkerRegisterSecret(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
		bodyBytes, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		if !bytes.Equal(bodyBytes, secret) {
			t.Fatalf("returned global worker secret is incorrect. Got %s want %s", string(bodyBytes[:]), string(secret[:]))
		}
	})
}

type workerStatusOverviewRespoonse struct {
	ActiveWorker    int   `json:"activeworker"`
	SuspendedWorker int   `json:"suspendedworker"`
	InactiveWorker  int   `json:"inactiveworker"`
	FinishedRuns    int64 `json:"finishedruns"`
	QueueSize       int   `json:"queuesize"`
}

func TestGetWorkerStatusOverview(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestGetWorkerStatusOverview")
	if err != nil {
		t.Fatal(err)
	}

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
	defer func() { services.MockStorageService(nil) }()

	// Initialize certificate store
	ca, err := security.InitCA()
	if err != nil {
		t.Fatalf("cannot initialize certificate service: %v", err)
	}

	// Initialize vault
	v, err := services.DefaultVaultService()
	if err != nil {
		t.Fatalf("cannot initialize vault service: %v", err)
	}

	// Initialize memdb service
	db, err := services.MemDBService(dataStore)
	if err != nil {
		t.Fatal(err)
	}

	scheduler, _ := gaiascheduler.NewScheduler(gaiascheduler.Dependencies{
		Store: m,
		DB:    db,
		PS:    &plugin.GoPlugin{},
		CA:    ca,
		Vault: v,
	})

	// Generate global worker secret
	secret := []byte(security.GenerateRandomUUIDV5())
	v.Add(gaia.WorkerRegisterKey, secret)
	if err := v.SaveSecrets(); err != nil {
		t.Fatal(err)
	}

	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       scheduler,
		PipelineService: nil,
		Certificate:     ca,
	})
	// Initialize echo
	e := echo.New()
	if err := handlerService.InitHandlers(e); err != nil {
		t.Fatal(err)
	}

	// Test empty worker status overview
	{
		wp := workers.NewWorkerProvider(workers.Dependencies{
			Scheduler:   scheduler,
			Certificate: ca,
		})
		req := httptest.NewRequest(echo.GET, "/api/"+gaia.APIVersion+"/worker/status", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := wp.GetWorkerStatusOverview(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
		bodyBytes, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		resp := &workerStatusOverviewRespoonse{}
		if err := json.Unmarshal(bodyBytes, resp); err != nil {
			t.Fatalf("failed to unmarshal response: %#v", bodyBytes)
		}

		if resp.FinishedRuns != 0 {
			t.Fatalf("finishedruns should be 0 but is %d", resp.FinishedRuns)
		}
		if resp.QueueSize != 0 {
			t.Fatalf("queuesize should be 0 but is %d", resp.QueueSize)
		}
		if resp.SuspendedWorker != 0 {
			t.Fatalf("suspendedworker should be 0 but is %d", resp.SuspendedWorker)
		}
		if resp.InactiveWorker != 0 {
			t.Fatalf("inactiveworker should be 0 but is %d", resp.InactiveWorker)
		}
	}

	// Test with registered worker
	{
		wp := workers.NewWorkerProvider(workers.Dependencies{
			Scheduler:   scheduler,
			Certificate: ca,
		})
		body := registerWorker{
			Name:   "my-worker",
			Secret: string(secret[:]),
			Tags:   []string{"first-tag", "second-tag", "third-tag"},
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/worker/register", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := wp.RegisterWorker(c); err != nil {
			t.Fatal(err)
		}

		_, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v; body: %s", http.StatusOK, rec.Code, string(bodyBytes[:]))
		}

		req = httptest.NewRequest(echo.GET, "/api/"+gaia.APIVersion+"/worker/status", nil)
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		if err := wp.GetWorkerStatusOverview(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
		bodyBytes, err = ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		resp := &workerStatusOverviewRespoonse{}
		if err := json.Unmarshal(bodyBytes, resp); err != nil {
			t.Fatalf("failed to unmarshal response: %#v", bodyBytes)
		}

		if resp.FinishedRuns != 0 {
			t.Fatalf("finishedruns should be 0 but is %d", resp.FinishedRuns)
		}
		if resp.QueueSize != 0 {
			t.Fatalf("queuesize should be 0 but is %d", resp.QueueSize)
		}
		if resp.SuspendedWorker != 0 {
			t.Fatalf("suspendedworker should be 0 but is %d", resp.SuspendedWorker)
		}
		if resp.InactiveWorker != 0 {
			t.Fatalf("inactiveworker should be 0 but is %d", resp.InactiveWorker)
		}
		if resp.ActiveWorker == 0 {
			t.Fatalf("activeworker should be 1 but is %d", resp.ActiveWorker)
		}
	}
}

func TestGetWorker(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestGetWorker")
	if err != nil {
		t.Fatal(err)
	}

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
	defer func() { services.MockStorageService(nil) }()

	// Initialize vault
	v, err := services.DefaultVaultService()
	if err != nil {
		t.Fatalf("cannot initialize vault service: %v", err)
	}

	// Initialize memdb service
	_, err = services.MemDBService(dataStore)
	if err != nil {
		t.Fatal(err)
	}

	// Generate global worker secret
	secret := []byte(security.GenerateRandomUUIDV5())
	v.Add(gaia.WorkerRegisterKey, secret)
	if err := v.SaveSecrets(); err != nil {
		t.Fatal(err)
	}
	ca, _ := security.InitCA()
	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       nil,
		PipelineService: nil,
		Certificate:     ca,
	})
	// Initialize echo
	e := echo.New()
	if err := handlerService.InitHandlers(e); err != nil {
		t.Fatal(err)
	}

	t.Run("get worker success", func(t *testing.T) {
		wp := workers.NewWorkerProvider(workers.Dependencies{Scheduler: nil, Certificate: ca})
		workerName := "my-worker"
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

		if err := wp.RegisterWorker(c); err != nil {
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

		req = httptest.NewRequest(echo.GET, "/api/"+gaia.APIVersion+"/worker", nil)
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		if err := wp.GetWorker(c); err != nil {
			t.Fatal(err)
		}

		bodyBytes, err = ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v; body: %s", http.StatusOK, rec.Code, string(bodyBytes[:]))
		}
		respWorkers := make([]gaia.Worker, 0)
		if err := json.Unmarshal(bodyBytes, &respWorkers); err != nil {
			t.Fatal(err)
		}
		if len(respWorkers) == 0 {
			t.Fatal("No workers returned but expected at least one")
		}
	})
}

func TestResetWorkerRegisterSecret(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestResetWorkerRegisterSecret")
	if err != nil {
		t.Fatal(err)
	}

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     tmp,
		HomePath:     tmp,
		PipelinePath: tmp,
		DevMode:      true,
	}

	// Initialize vault
	v, err := services.DefaultVaultService()
	if err != nil {
		t.Fatalf("cannot initialize vault service: %v", err)
	}

	// Generate global worker secret
	secret := []byte(security.GenerateRandomUUIDV5())
	v.Add(gaia.WorkerRegisterKey, secret)
	if err := v.SaveSecrets(); err != nil {
		t.Fatal(err)
	}
	ca, _ := security.InitCA()
	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       nil,
		PipelineService: nil,
		Certificate:     ca,
	})
	// Initialize echo
	e := echo.New()
	if err := handlerService.InitHandlers(e); err != nil {
		t.Fatal(err)
	}
	wp := workers.NewWorkerProvider(workers.Dependencies{Scheduler: nil, Certificate: ca})
	// Test reset global worker secret
	t.Run("global secret reset success", func(t *testing.T) {
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/worker/secret", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := wp.ResetWorkerRegisterSecret(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
		bodyBytes, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		if string(bodyBytes[:]) != "global worker registration secret has been successfully reset" {
			t.Fatalf("returned string is not correct: %s", string(bodyBytes[:]))
		}

		// Verify the secret has been changed
		req = httptest.NewRequest(echo.GET, "/api/"+gaia.APIVersion+"/worker/secret", nil)
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		if err := wp.GetWorkerRegisterSecret(c); err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
		bodyBytes, err = ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		if bytes.Equal(bodyBytes, secret) {
			t.Fatalf("returned global worker secret is identical. Got %s and %s", string(bodyBytes[:]), string(secret[:]))
		}
	})
}
