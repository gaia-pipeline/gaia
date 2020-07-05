package pipeline

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	hclog "github.com/hashicorp/go-hclog"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
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

type mockScheduler struct {
	Error error
}

func (ms *mockScheduler) Init() {}
func (ms *mockScheduler) SchedulePipeline(p *gaia.Pipeline, startedBy string, args []*gaia.Argument) (*gaia.PipelineRun, error) {
	return nil, nil
}
func (ms *mockScheduler) SetPipelineJobs(p *gaia.Pipeline) error            { return ms.Error }
func (ms *mockScheduler) StopPipelineRun(p *gaia.Pipeline, runid int) error { return ms.Error }
func (ms *mockScheduler) GetFreeWorkers() int32                             { return int32(0) }
func (ms *mockScheduler) CountScheduledRuns() int                           { return 0 }

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
	cp := new(gaia.CreatePipeline)
	cp.Pipeline.Type = gaia.PTypeUnknown
	pipelineService := NewGaiaPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
		Store:     mcp,
	})
	pipelineService.CreatePipeline(cp)
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
	cp := new(gaia.CreatePipeline)
	cp.Pipeline.Type = gaia.PTypeGolang
	pipelineService := NewGaiaPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
		Store:     mcp,
	})
	pipelineService.CreatePipeline(cp)
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
	cp := new(gaia.CreatePipeline)
	cp.Pipeline.Type = gaia.PTypeGolang
	cp.Pipeline.Repo = &gaia.GitRepo{URL: "https://github.com/gaia-pipeline/pipeline-test"}
	pipelineService := NewGaiaPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
		Store:     mcp,
	})
	pipelineService.CreatePipeline(cp)
	body, _ := ioutil.ReadAll(buf)
	if !bytes.Contains(body, []byte("cannot put create pipeline into store: error=failed")) {
		t.Fatal("expected log message was not there. was: ", string(body))
	}
}

func TestCreatePipeline(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreatePipeline")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	gaia.Cfg.PipelinePath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	mcp := new(mockCreatePipelineStore)
	cp := new(gaia.CreatePipeline)
	cp.Pipeline.Name = "test"
	cp.Pipeline.Type = gaia.PTypeGolang
	cp.Pipeline.Repo = &gaia.GitRepo{URL: "https://github.com/gaia-pipeline/pipeline-test"}
	pipelineService := NewGaiaPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
		Store:     mcp,
	})
	pipelineService.CreatePipeline(cp)
	if cp.StatusType != gaia.CreatePipelineSuccess {
		t.Fatal("pipeline status was not success. was: ", cp.StatusType)
	}
}

func TestCreatePipelineSetPipelineJobsFail(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreatePipelineSetPipelineJobsFail")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	gaia.Cfg.PipelinePath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	mcp := new(mockCreatePipelineStore)
	cp := new(gaia.CreatePipeline)
	cp.Pipeline.Name = "test"
	cp.Pipeline.Type = gaia.PTypeGolang
	cp.Pipeline.Repo = &gaia.GitRepo{URL: "https://github.com/gaia-pipeline/pipeline-test"}
	pipelineService := NewGaiaPipelineService(Dependencies{
		Scheduler: &mockScheduleService{
			err: errors.New("error"),
		},
		Store: mcp,
	})
	pipelineService.CreatePipeline(cp)
	if !strings.Contains(cp.Output, "cannot validate pipeline") {
		t.Fatalf("error thrown should contain 'cannot validate pipeline' but its %s", cp.Output)
	}
}
