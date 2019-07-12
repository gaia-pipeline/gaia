package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	gStore "github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	"github.com/gaia-pipeline/gaia/workers/scheduler"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
)

type mockScheduleService struct {
	scheduler.GaiaScheduler
	pipelineRun *gaia.PipelineRun
	err         error
}

func (ms *mockScheduleService) SchedulePipeline(p *gaia.Pipeline, args []*gaia.Argument) (*gaia.PipelineRun, error) {
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
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/gitlsremote", nil)
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
			"user":     "admin",
			"password": "admin",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/gitlsremote", bytes.NewBuffer(bodyBytes))
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
			"url": repoURL,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/gitlsremote", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		PipelineGitLSRemote(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})
}

func TestPipelineGitLSRemoteWithKeys(t *testing.T) {
	samplePrivateKey := `
-----BEGIN RSA PRIVATE KEY-----
MD8CAQACCQDB9DczYvFuZQIDAQABAgkAtqAKvH9QoQECBQDjAl9BAgUA2rkqJQIE
Xbs5AQIEIzWnmQIFAOEml+E=
-----END RSA PRIVATE KEY-----
`
	dataDir, _ := ioutil.TempDir("", "TestPipelineGitLSRemoteWithKeys")

	defer func() {
		gaia.Cfg = nil
	}()

	gaia.Cfg = &gaia.Config{
		Logger:   hclog.NewNullLogger(),
		DataPath: dataDir,
	}

	e := echo.New()
	InitHandlers(e)

	t.Run("invalid hostconfig for github in known_hosts file", func(t *testing.T) {
		buf := new(bytes.Buffer)
		gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
			Level:  hclog.Trace,
			Output: buf,
			Name:   "Gaia",
		})
		hostConfig := "github.comom,1.2.3.4 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="
		knownHostsLocation := filepath.Join(dataDir, ".known_hosts")
		err := ioutil.WriteFile(knownHostsLocation, []byte(hostConfig), gaia.ExecutablePermission)
		if err != nil {
			t.Fatal(err)
		}
		os.Setenv("SSH_KNOWN_HOSTS", knownHostsLocation)
		repoURL := "github.com:gaia-pipeline/pipeline-test"
		gr := gaia.GitRepo{
			URL: repoURL,
			PrivateKey: gaia.PrivateKey{
				Key:      samplePrivateKey,
				Username: "git",
				Password: "",
			},
		}
		bodyBytes, _ := json.Marshal(gr)
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/gitlsremote", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		PipelineGitLSRemote(c)

		// This will fail because the above SSH key is invalid. But that is fine,
		// because the initial host file will fail earlier than that.
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected response code %v got %v", http.StatusForbidden, rec.Code)
		}

		// This is the important bit that needs to be tested.
		want := "knownhosts: key is unknown"
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("wanted buf to contain: '%s', got: '%s'", want, buf.String())
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
	gaia.Cfg.Bolt.Mode = 0600

	// Initialize store
	dataStore, err := services.StorageService()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { services.MockStorageService(nil) }()
	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	// Initialize echo
	e := echo.New()
	InitHandlers(e)

	pipeline1 := gaia.Pipeline{
		ID:                1,
		Name:              "Pipeline A",
		Type:              gaia.PTypeGolang,
		Created:           time.Now(),
		PeriodicSchedules: []string{"0 30 * * * *"},
	}

	pipeline2 := gaia.Pipeline{
		ID:      2,
		Name:    "Pipeline B",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	// Add to store
	err = dataStore.PipelinePut(&pipeline1)
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
		c.SetPath("/api/" + gaia.APIVersion + "/pipeline/:pipelineid")
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
		c.SetPath("/api/" + gaia.APIVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		PipelineUpdate(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})

	t.Run("update periodic schedules success", func(t *testing.T) {
		p := gaia.Pipeline{
			ID:                1,
			Name:              "newname",
			PeriodicSchedules: []string{"0 */1 * * * *"},
		}
		bodyBytes, _ := json.Marshal(p)
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + gaia.APIVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		ms := new(mockScheduleService)
		pRun := new(gaia.PipelineRun)
		pRun.ID = 999
		ms.pipelineRun = pRun
		services.MockSchedulerService(ms)

		PipelineUpdate(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})

	t.Run("update periodic schedules failed", func(t *testing.T) {
		p := gaia.Pipeline{
			ID:                1,
			Name:              "newname",
			PeriodicSchedules: []string{"0 */1 * * * * *"},
		}
		bodyBytes, _ := json.Marshal(p)
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + gaia.APIVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		ms := new(mockScheduleService)
		pRun := new(gaia.PipelineRun)
		pRun.ID = 999
		ms.pipelineRun = pRun
		services.MockSchedulerService(ms)

		PipelineUpdate(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
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
	dataStore, err := services.StorageService()
	if err != nil {
		t.Fatal(err)
	}
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
	err = dataStore.PipelinePut(&p)
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
		c.SetPath("/api/" + gaia.APIVersion + "/pipeline/:pipelineid")
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
		c.SetPath("/api/" + gaia.APIVersion + "/pipeline/:pipelineid")
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
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/:pipelineid/start", nil)
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

		expectedBody := `{"uniqueid":"","id":999,"pipelineid":0,"startdate":"0001-01-01T00:00:00Z","finishdate":"0001-01-01T00:00:00Z","scheduledate":"0001-01-01T00:00:00Z"}
`
		body, _ := ioutil.ReadAll(rec.Body)
		if string(body) != expectedBody {
			t.Fatalf("body did not equal expected content. expected: %s, got: %s", expectedBody, string(body))
		}
	})

	t.Run("fails when scheduler throws error", func(t *testing.T) {
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/:pipelineid/start", nil)
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
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/:pipelineid/start", nil)
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

type mockUserStoreService struct {
	gStore.GaiaStore
	user *gaia.User
	err  error
}

func (m mockUserStoreService) UserGet(username string) (*gaia.User, error) {
	return m.user, m.err
}

func TestPipelineRemoteTrigger(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineRemoteTrigger")
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
		ID:           1,
		Name:         "Pipeline A",
		Type:         gaia.PTypeGolang,
		Created:      time.Now(),
		TriggerToken: "triggerToken",
	}

	// Add to active pipelines
	ap.Append(p)

	t.Run("can trigger a pipeline with auto user", func(t *testing.T) {
		user := gaia.User{}
		user.Username = "auto"
		user.TriggerToken = "triggerToken"
		m := mockUserStoreService{user: &user, err: nil}
		services.MockStorageService(&m)
		defer func() {
			services.MockStorageService(nil)
			services.MockSchedulerService(nil)
		}()

		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/1/triggerToken/trigger", nil)
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("auto", "triggerToken")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid", "pipelinetoken")
		c.SetParamValues("1", "triggerToken")
		ms := new(mockScheduleService)
		pRun := new(gaia.PipelineRun)
		pRun.ID = 999
		ms.pipelineRun = pRun
		services.MockSchedulerService(ms)

		PipelineTrigger(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})
	t.Run("can't trigger a pipeline with invalid auto user", func(t *testing.T) {
		user := gaia.User{}
		user.Username = "auto"
		user.TriggerToken = "triggerToken"
		m := mockUserStoreService{user: &user, err: nil}
		services.MockStorageService(&m)
		defer func() {
			services.MockStorageService(nil)
			services.MockSchedulerService(nil)
		}()

		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/1/triggerToken/trigger", nil)
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("auto", "invalid")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid", "pipelinetoken")
		c.SetParamValues("1", "triggerToken")
		ms := new(mockScheduleService)
		pRun := new(gaia.PipelineRun)
		pRun.ID = 999
		ms.pipelineRun = pRun
		services.MockSchedulerService(ms)

		PipelineTrigger(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("can't trigger a pipeline with invalid token", func(t *testing.T) {
		user := gaia.User{}
		user.Username = "auto"
		user.TriggerToken = "triggerToken"
		m := mockUserStoreService{user: &user, err: nil}
		services.MockStorageService(&m)
		defer func() {
			services.MockStorageService(nil)
			services.MockSchedulerService(nil)
		}()

		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/1/invalid/trigger", nil)
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("auto", "triggerToken")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid", "pipelinetoken")
		c.SetParamValues("1", "invalid")
		ms := new(mockScheduleService)
		pRun := new(gaia.PipelineRun)
		pRun.ID = 999
		ms.pipelineRun = pRun
		services.MockSchedulerService(ms)

		PipelineTrigger(c)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected response code %v got %v", http.StatusForbidden, rec.Code)
		}
	})
	t.Run("can't trigger a pipeline without authentication information", func(t *testing.T) {
		user := gaia.User{}
		user.Username = "auto"
		user.TriggerToken = "triggerToken"
		m := mockUserStoreService{user: &user, err: nil}
		services.MockStorageService(&m)
		defer func() {
			services.MockStorageService(nil)
			services.MockSchedulerService(nil)
		}()

		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/1/invalid/trigger", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid", "pipelinetoken")
		c.SetParamValues("1", "invalid")
		ms := new(mockScheduleService)
		pRun := new(gaia.PipelineRun)
		pRun.ID = 999
		ms.pipelineRun = pRun
		services.MockSchedulerService(ms)

		PipelineTrigger(c)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected response code %v got %v", http.StatusForbidden, rec.Code)
		}
	})
}

type mockPipelineResetStorageService struct {
	gStore.GaiaStore
}

var pipelineNewToken string

func (m mockPipelineResetStorageService) PipelinePut(pipeline *gaia.Pipeline) error {
	pipelineNewToken = pipeline.TriggerToken
	return nil
}

func TestPipelineResetToken(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineResetToken")
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
		ID:           1,
		Name:         "Pipeline A",
		Type:         gaia.PTypeGolang,
		Created:      time.Now(),
		TriggerToken: "triggerToken",
	}

	// Add to active pipelines
	ap.Append(p)

	req := httptest.NewRequest(echo.PUT, "/api/"+gaia.APIVersion+"/pipeline/1/reset-trigger-token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("pipelineid")
	c.SetParamValues("1")
	ms := new(mockScheduleService)
	pRun := new(gaia.PipelineRun)
	pRun.ID = 999
	ms.pipelineRun = pRun
	services.MockSchedulerService(ms)

	m := mockPipelineResetStorageService{}
	services.MockStorageService(&m)

	defer func() {
		services.MockStorageService(nil)
		services.MockSchedulerService(nil)
	}()

	PipelineResetToken(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}

	if pipelineNewToken == "triggerToken" {
		t.Fatal("expected token to be reset. was not reset.")
	}
}

func TestPipelineCheckPeriodicSchedules(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineCheckPeriodicSchedules")
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

	t.Run("invalid cron added", func(t *testing.T) {
		body := []string{
			"* * * * * * *",
			"*/1 * * 200 *",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/periodicschedules", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := PipelineCheckPeriodicSchedules(c)
		if err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("valid cron added", func(t *testing.T) {
		body := []string{
			"0 30 * * * *",
			"0 */5 * * * *",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/periodicschedules", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		PipelineCheckPeriodicSchedules(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})
}

func TestPipelineNameAvailable(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineNameAvailable")
	dataDir := tmp

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		HomePath:     dataDir,
		DataPath:     dataDir,
		PipelinePath: dataDir,
	}

	// Initialize store
	dataStore, err := services.StorageService()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { services.MockStorageService(nil) }()

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	// Initialize echo
	e := echo.New()
	err = InitHandlers(e)
	if err != nil {
		t.Fatal(err)
	}

	p := gaia.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	// Add to store
	err = dataStore.PipelinePut(&p)
	if err != nil {
		t.Fatal(err)
	}
	// Add to active pipelines
	ap.Append(p)

	t.Run("fails for pipeline name already in use", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + gaia.APIVersion + "/pipeline/name")
		q := req.URL.Query()
		q.Add("name", "pipeline a")
		req.URL.RawQuery = q.Encode()

		err := PipelineNameAvailable(c)
		if err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
		bodyBytes, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		nameAlreadyInUseMessage := "pipeline name is already in use"
		if string(bodyBytes[:]) != nameAlreadyInUseMessage {
			t.Fatalf("error message should be '%s' but was '%s'", nameAlreadyInUseMessage, string(bodyBytes[:]))
		}
	})

	t.Run("fails for pipeline name is too long", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + gaia.APIVersion + "/pipeline/name")
		q := req.URL.Query()
		q.Add("name", "pipeline a pipeline a pipeline a pipeline a pipeline a pipeline a pipeline a pipeline a pipeline a pipeline a")
		req.URL.RawQuery = q.Encode()

		PipelineNameAvailable(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
		bodyBytes, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		nameTooLongMessage := "name of pipeline is empty or one of the path elements length exceeds 50 characters"
		if string(bodyBytes[:]) != nameTooLongMessage {
			t.Fatalf("error message should be '%s' but was '%s'", nameTooLongMessage, string(bodyBytes[:]))
		}
	})

	t.Run("works for pipeline with different name", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + gaia.APIVersion + "/pipeline/name")
		q := req.URL.Query()
		q.Add("name", "pipeline b")
		req.URL.RawQuery = q.Encode()

		PipelineNameAvailable(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})
}
