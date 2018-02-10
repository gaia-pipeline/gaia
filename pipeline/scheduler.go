package pipeline

import (
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/plugin"
)

const (
	// Maximum buffer limit for scheduler
	schedulerBufferLimit = 50

	// schedulerIntervalSeconds defines the interval the scheduler will look
	// for new work to schedule. Definition in seconds.
	schedulerIntervalSeconds = 3
)

// Scheduler represents the schuler object
type Scheduler struct {
	// buffered channel which is used as queue
	pipelines chan gaia.Pipeline
}

// NewScheduler creates a new instance of Scheduler.
func NewScheduler() *Scheduler {
	// Create new scheduler
	s := &Scheduler{
		pipelines: make(chan gaia.Pipeline, schedulerBufferLimit),
	}

	return s
}

// Init initializes the scheduler.
func (s *Scheduler) Init() {
	// Create a periodic job that fills the scheduler with new pipelines.
	schedulerJob := time.NewTicker(schedulerIntervalSeconds * time.Second)
	go func() {
		for {
			select {
			case <-schedulerJob.C:
				checkActivePipelines()
			}
		}
	}()
}

// Schedule looks in the store for new work to do and schedules it.
func (s *Scheduler) Schedule() {
	// Do we have space left in our buffer?
	if len(s.pipelines) >= schedulerBufferLimit {
		// No space left. Exit.
		gaia.Cfg.Logger.Debug("scheduler buffer overflow. Cannot schedule new pipelines...")
		return
	}

	// TODO: Implement schedule

}

// setPipelineJobs uses the plugin system to get all
// jobs from the given pipeline.
// This function is blocking and might take some time.
func setPipelineJobs(p *gaia.Pipeline) error {
	// Create the start command for the pipeline
	c := createPipelineCmd(p)
	if c == nil {
		gaia.Cfg.Logger.Debug("cannot set pipeline jobs", "error", errMissingType.Error(), "pipeline", p)
		return errMissingType
	}

	// Create new plugin instance
	pC := plugin.NewPlugin(c)

	// Connect to plugin(pipeline)
	if err := pC.Connect(); err != nil {
		gaia.Cfg.Logger.Debug("cannot connect to pipeline", "error", err.Error(), "pipeline", p)
		return err
	}
	defer pC.Close()

	// Get jobs
	jobs, err := pC.GetJobs()
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot get jobs from pipeline", "error", err.Error(), "pipeline", p)
		return err
	}
	p.Jobs = jobs

	return nil
}
