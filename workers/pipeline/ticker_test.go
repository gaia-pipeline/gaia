package pipeline

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia/store/memdb"
	"github.com/gaia-pipeline/gaia/workers/scheduler/gaiascheduler"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/hashicorp/go-hclog"
)

type mockScheduleService struct {
	gaiascheduler.GaiaScheduler
	err error
}

func (ms *mockScheduleService) SetPipelineJobs(p *gaia.Pipeline) error {
	return ms.err
}

type mockMemDBService struct {
	worker    *gaia.Worker
	setWorker *gaia.Worker
	memdb.GaiaMemDB
}

func (mm *mockMemDBService) GetAllWorker() []*gaia.Worker {
	return []*gaia.Worker{mm.setWorker}
}
func (mm *mockMemDBService) UpsertWorker(w *gaia.Worker, persist bool) error {
	mm.worker = w
	return nil
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
	gaia.Cfg.Bolt.Mode = 0600

	// Initialize store
	dataStore, err := services.StorageService()
	if err != nil {
		t.Fatal(err)
	}
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
	_, err = dataStore.PipelineGet(pipeline1.ID)
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

func TestUpdateWorker(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestUpdateWorker")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	gaia.Cfg = &gaia.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     tmp,
		HomePath:     tmp,
		PipelinePath: tmp,
	}

	db := &mockMemDBService{}
	services.MockMemDBService(db)
	defer func() { services.MockMemDBService(nil) }()
	db.setWorker = &gaia.Worker{
		Status:      gaia.WorkerActive,
		LastContact: time.Now().Add(-6 * time.Minute),
	}

	// Run update worker
	updateWorker()

	// Validate
	if db.worker == nil {
		t.Fatal("worker should not be nil")
	}
	if db.worker.Status != gaia.WorkerInactive {
		t.Fatalf("expected '%s' but got '%s'", string(gaia.WorkerInactive), string(db.worker.Status))
	}
	db.worker = nil

	// Set new test data
	db.setWorker = &gaia.Worker{
		Status:      gaia.WorkerInactive,
		LastContact: time.Now(),
	}

	// Run update worker
	updateWorker()

	// Validate
	if db.worker == nil {
		t.Fatal("worker should not be nil")
	}
	if db.worker.Status != gaia.WorkerActive {
		t.Fatalf("expected '%s' but got '%s'", string(gaia.WorkerActive), string(db.worker.Status))
	}
}
