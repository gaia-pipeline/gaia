package pipeline

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/workers/scheduler"
	hclog "github.com/hashicorp/go-hclog"
)

type mockScheduleService struct {
	scheduler.GaiaScheduler
	err error
}

func (ms *mockScheduleService) SetPipelineJobs(p *gaia.Pipeline) error {
	return ms.err
}

func TestCheckActivePipelines(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCheckActivePipelines")
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
	ap := NewActivePipelines()
	GlobalActivePipelines = ap
	// Mock scheduler service
	ms := new(mockScheduleService)
	services.MockSchedulerService(ms)

	pipeline1 := gaia.Pipeline{
		ID:      1,
		Name:    "testpipe",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	// Create fake binary
	src := GetExecPath(pipeline1)
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(src)

	// Manually run check
	checkActivePipelines()

	// Check if pipeline was added to store
	_, err := dataStore.PipelineGet(pipeline1.ID)
	if err != nil {
		t.Error("cannot find pipeline in store")
	}
}

func TestTurningThePollerOn(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestTurningThePollerOn")
	dataDir := tmp

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         true,
	}

	defer StopPoller()
	err := StartPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
}

func TestTurningThePollerOnWhilePollingIsDisabled(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestTurningThePollerOnWhilePollingIsDisabled")
	dataDir := tmp

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         false,
	}

	err := StartPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
	if isPollerRunning != false {
		t.Fatal("expected isPollerRunning to be false. was: ", isPollerRunning)
	}
}

func TestTurningThePollerOnWhilePollingIsEnabled(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestTurningThePollerOnWhilePollingIsEnabled")
	dataDir := tmp

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         true,
	}
	defer StopPoller()
	err := StartPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
	if isPollerRunning != true {
		t.Fatal("expected isPollerRunning to be true. was: ", isPollerRunning)
	}
}

func TestTurningThePollerOff(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestTurningThePollerOff")
	dataDir := tmp

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         true,
	}

	err := StartPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
	if isPollerRunning != true {
		t.Fatal("expected isPollerRunning to be true. was: ", isPollerRunning)
	}

	err = StopPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
	if isPollerRunning != false {
		t.Fatal("expected isPollerRunning to be false. was: ", isPollerRunning)
	}
}

func TestTogglePoller(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestTogglePoller")
	dataDir := tmp

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         true,
	}

	err := StartPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
	err = StartPoller()
	if err == nil {
		t.Fatal("starting the poller again should have failed")
	}
	err = StopPoller()
	if err != nil {
		t.Fatal("stopping the poller while it's running should not have failed. got: ", err)
	}
	err = StopPoller()
	if err == nil {
		t.Fatal("stopping the poller again while it's stopped should have failed.")
	}
}
