package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
)

type MockVaultStorer struct {
	Error error
}

var storeBts []byte

func (mvs *MockVaultStorer) Init() error {
	storeBts = make([]byte, 0)
	return mvs.Error
}

func (mvs *MockVaultStorer) Read() ([]byte, error) {
	return storeBts, mvs.Error
}

func (mvs *MockVaultStorer) Write(data []byte) error {
	storeBts = data
	return mvs.Error
}

func TestHookReceive(t *testing.T) {
	dataDir, err := ioutil.TempDir("", "TestHookReceive")
	if err != nil {
		t.Fatalf("error creating data dir %v", err.Error())
	}
	pipelineService := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})
	defer func() {
		gaia.Cfg = nil
		_ = os.RemoveAll(dataDir)
	}()
	gaia.Cfg = &gaia.Config{
		Logger:    hclog.NewNullLogger(),
		DataPath:  dataDir,
		CAPath:    dataDir,
		VaultPath: dataDir,
		HomePath:  dataDir,
	}

	m := new(MockVaultStorer)
	v, _ := services.VaultService(m)
	v.Add("GITHUB_WEBHOOK_SECRET", []byte("superawesomesecretgithubpassword"))
	defer func() {
		services.MockVaultService(nil)
	}()
	e := echo.New()

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	p := gaia.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
		Repo: &gaia.GitRepo{
			URL: "https://github.com/Codertocat/Hello-World",
		},
	}

	ap.Append(p)

	_ = handlerService.InitHandlers(e)

	t.Run("successfully extracting path information from payload", func(t *testing.T) {
		payload, _ := ioutil.ReadFile(filepath.Join("fixtures", "hook_basic_push_payload.json"))
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/githook", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		// Use https://www.freeformatter.com/hmac-generator.html#ad-output for example
		// to calculate a new sha if the fixture would change.
		req.Header.Set("x-hub-signature", "sha1=940e53f44518a6cf9ba002c29c8ace7799af2b13")
		req.Header.Set("x-github-event", "push")
		req.Header.Set("X-github-delivery", "1234asdf")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = GitWebHook(c)

		// Expected failure because repository does not exist
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("want response code %v got %v", http.StatusInternalServerError, rec.Code)
		}

		// Checking body to make sure it's the failure we want
		body, _ := ioutil.ReadAll(rec.Body)
		want := "failed to build pipeline:  repository does not exist\n"
		if string(body) != want {
			t.Fatalf("want body: %s, got: %s", want, string(body))
		}
	})

	t.Run("only push events are accepted", func(t *testing.T) {
		payload, _ := ioutil.ReadFile(filepath.Join("fixtures", "hook_basic_push_payload.json"))
		req := httptest.NewRequest(echo.POST, "/api/"+gaia.APIVersion+"/pipeline/githook", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		// Use https://www.freeformatter.com/hmac-generator.html#ad-output for example
		// to calculate a new sha if the fixture would change.
		req.Header.Set("x-hub-signature", "sha1=940e53f44518a6cf9ba002c29c8ace7799af2b13")
		req.Header.Set("x-github-event", "pull")
		req.Header.Set("X-github-delivery", "1234asdf")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = GitWebHook(c)

		// Expected failure because repository does not exist
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("want response code %v got %v", http.StatusBadRequest, rec.Code)
		}

		// Checking body to make sure it's the failure we want
		body, _ := ioutil.ReadAll(rec.Body)
		want := "invalid event"
		if string(body) != want {
			t.Fatalf("want body: %s, got: %s", want, string(body))
		}
	})
}
