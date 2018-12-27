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
