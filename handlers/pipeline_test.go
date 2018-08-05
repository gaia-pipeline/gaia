package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia/scheduler"

	"github.com/gaia-pipeline/gaia/services"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/pipeline"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
)

type mockScheduleService struct {
	scheduler.GaiaScheduler
	pipelineRun *gaia.PipelineRun
	err         error
}

func (ms *mockScheduleService) SchedulePipeline(p *gaia.Pipeline) (*gaia.PipelineRun, error) {
	return ms.pipelineRun, ms.err
}

func TestPipelineGitLSRemote(t *testing.T) {
	dataDir, _ := ioutil.TempDir("", "TestPipelineGitLSRemote")

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		Logger:   hclog.NewNullLogger(),
		DataPath: dataDir,
	}

	e := echo.New()
	InitHandlers(e)

	t.Run("fails with invalid data", func(t *testing.T) {
		req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/pipeline/gitlsremote", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		PipelineGitLSRemote(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})

	t.Run("fails with invalid access", func(t *testing.T) {
		repoURL := "https://example.com"
		body := map[string]string{
			"url":      repoURL,
			"username": "admin",
			"password": "admin",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/pipeline/gitlsremote", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		PipelineGitLSRemote(c)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})

	t.Run("otherwise succeed", func(t *testing.T) {
		repoURL := "https://github.com/gaia-pipeline/pipeline-test"
		body := map[string]string{
			"url":      repoURL,
			"username": "admin",
			"password": "admin",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/pipeline/gitlsremote", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		PipelineGitLSRemote(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})
}

func TestPipelineUpdate(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineUpdate")
	dataDir := tmp

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
	}

	// Initialize store
	dataStore, _ := services.StorageService()
	dataStore.Init()
	defer func() { services.MockStorageService(nil) }()
	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	// Initialize echo
	e := echo.New()
	InitHandlers(e)

	pipeline1 := gaia.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	pipeline2 := gaia.Pipeline{
		ID:      2,
		Name:    "Pipeline B",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	// Add to store
	err := dataStore.PipelinePut(&pipeline1)
	if err != nil {
		t.Fatal(err)
	}
	// Add to active pipelines
	ap.Append(pipeline1)
	// Create binary
	src := pipeline.GetExecPath(pipeline1)
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(src)

	t.Run("fails for non-existent pipeline", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(pipeline2)
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + apiVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("2")

		PipelineUpdate(c)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected response code %v got %v", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("works for existing pipeline", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(pipeline1)
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + apiVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		PipelineUpdate(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusNotFound, rec.Code)
		}
	})

}

func TestPipelineDelete(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineDelete")
	dataDir := tmp

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		HomePath:     dataDir,
		DataPath:     dataDir,
		PipelinePath: dataDir,
	}

	// Initialize store
	dataStore, _ := services.StorageService()
	dataStore.Init()
	defer func() { services.MockStorageService(nil) }()

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	// Initialize echo
	e := echo.New()
	InitHandlers(e)

	p := gaia.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	// Add to store
	err := dataStore.PipelinePut(&p)
	if err != nil {
		t.Fatal(err)
	}
	// Add to active pipelines
	ap.Append(p)
	// Create binary
	src := pipeline.GetExecPath(p)
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(src)

	ioutil.WriteFile(src, []byte("testcontent"), 0666)

	t.Run("fails for non-existent pipeline", func(t *testing.T) {
		req := httptest.NewRequest(echo.DELETE, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + apiVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("2")

		PipelineDelete(c)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected response code %v got %v", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("works for existing pipeline", func(t *testing.T) {
		req := httptest.NewRequest(echo.DELETE, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + apiVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		PipelineDelete(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusNotFound, rec.Code)
		}
	})

}

func TestPipelineStart(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineStart")
	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		HomePath:     tmp,
		DataPath:     tmp,
		PipelinePath: tmp,
	}

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	// Initialize echo
	e := echo.New()
	InitHandlers(e)

	p := gaia.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	// Add to active pipelines
	ap.Append(p)

	t.Run("can start a pipeline", func(t *testing.T) {
		req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/pipeline/:pipelineid/start", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		ms := new(mockScheduleService)
		pRun := new(gaia.PipelineRun)
		pRun.ID = 999
		ms.pipelineRun = pRun
		services.MockSchedulerService(ms)

		PipelineStart(c)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected response code %v got %v", http.StatusCreated, rec.Code)
		}

		expectedBody := `{"uniqueid":"","id":999,"pipelineid":0,"startdate":"0001-01-01T00:00:00Z","finishdate":"0001-01-01T00:00:00Z","scheduledate":"0001-01-01T00:00:00Z"}`
		body, _ := ioutil.ReadAll(rec.Body)
		if string(body) != expectedBody {
			t.Fatalf("body did not equal expected content. expected: %s, got: %s", expectedBody, string(body))
		}
	})

	t.Run("fails when scheduler throws error", func(t *testing.T) {
		req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/pipeline/:pipelineid/start", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		ms := new(mockScheduleService)
		pRun := new(gaia.PipelineRun)
		pRun.ID = 999
		ms.pipelineRun = pRun
		ms.err = errors.New("failed to run pipeline")
		services.MockSchedulerService(ms)

		PipelineStart(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("fails when scheduler doesn't find the pipeline but does not return error", func(t *testing.T) {
		req := httptest.NewRequest(echo.POST, "/api/"+apiVersion+"/pipeline/:pipelineid/start", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		ms := new(mockScheduleService)
		services.MockSchedulerService(ms)

		PipelineStart(c)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected response code %v got %v", http.StatusNotFound, rec.Code)
		}
	})
}
