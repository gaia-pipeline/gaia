package pipeline

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/gaia-pipeline/gaia/services"

	"github.com/gaia-pipeline/gaia/store"

	"github.com/gaia-pipeline/gaia"
	"github.com/hashicorp/go-hclog"
)

type mockCreatePipelineStore struct {
	store.GaiaStore
	Error error
}

func (mcp *mockCreatePipelineStore) CreatePipelinePut(p *gaia.CreatePipeline) error {
	return mcp.Error
}

// PipelinePut is a Mock implementation for pipelines
func (mcp *mockCreatePipelineStore) PipelinePut(p *gaia.Pipeline) error {
	return mcp.Error
}

func TestCreatePipelineUnknownType(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreatePipelineUnknownType")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	mcp := new(mockCreatePipelineStore)
	services.MockStorageService(mcp)
	defer func() { services.MockStorageService(nil) }()
	cp := new(gaia.CreatePipeline)
	cp.Pipeline.Type = gaia.PTypeUnknown
	CreatePipeline(cp)
	if cp.Output != "create pipeline failed. Pipeline type is not supported unknown is not supported" {
		t.Fatal("error output was not the expected output. was: ", cp.Output)
	}
	if cp.StatusType != gaia.CreatePipelineFailed {
		t.Fatal("pipeline status is not expected status. was:", cp.StatusType)
	}
}

func TestCreatePipelineMissingGitURL(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreatePipelineMissingGitURL")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	mcp := new(mockCreatePipelineStore)
	services.MockStorageService(mcp)
	defer func() { services.MockStorageService(nil) }()
	cp := new(gaia.CreatePipeline)
	cp.Pipeline.Type = gaia.PTypeGolang
	CreatePipeline(cp)
	if cp.Output != "cannot prepare build: URL field is required" {
		t.Fatal("output was not what was expected. was: ", cp.Output)
	}
}

func TestCreatePipelineFailedToUpdatePipeline(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreatePipelineFailedToUpdatePipeline")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	mcp := new(mockCreatePipelineStore)
	mcp.Error = errors.New("failed")
	services.MockStorageService(mcp)
	defer func() { services.MockStorageService(nil) }()
	cp := new(gaia.CreatePipeline)
	cp.Pipeline.Type = gaia.PTypeGolang
	cp.Pipeline.Repo.URL = "https://github.com/gaia-pipeline/pipeline-test"
	CreatePipeline(cp)
	body, _ := ioutil.ReadAll(buf)
	if !bytes.Contains(body, []byte("cannot put create pipeline into store: error=failed")) {
		t.Fatal("expected log message was not there. was: ", string(body))
	}
}

func TestCreatePipeline(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreatePipeline")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	mcp := new(mockCreatePipelineStore)
	services.MockStorageService(mcp)
	defer func() { services.MockStorageService(nil) }()
	cp := new(gaia.CreatePipeline)
	cp.Pipeline.Type = gaia.PTypeGolang
	cp.Pipeline.Repo.URL = "https://github.com/gaia-pipeline/pipeline-test"
	CreatePipeline(cp)
	if cp.StatusType != gaia.CreatePipelineSuccess {
		t.Fatal("pipeline status was not success. was: ", cp.StatusType)
	}
}
