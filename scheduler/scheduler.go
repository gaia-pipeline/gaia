package pipeline

import (
	"errors"
	"os/exec"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/plugin"
	"github.com/gaia-pipeline/gaia/store"
)

const (
	// Maximum buffer limit for scheduler
	schedulerBufferLimit = 50

	// schedulerIntervalSeconds defines the interval the scheduler will look
	// for new work to schedule. Definition in seconds.
	schedulerIntervalSeconds = 3
)

var (
	// errCreateCMDForPipeline is thrown when we couldnt create a command to start
	// a plugin.
	errCreateCMDForPipeline = errors.New("could not create execute command for plugin")

	// storeService is an instance of store.
	// Use this to talk to the store.
	storeService *store.Store
)

// Scheduler represents the schuler object
type Scheduler struct {
	// buffered channel which is used as queue
	pipelines chan gaia.Pipeline
}

// NewScheduler creates a new instance of Scheduler.
func NewScheduler(store *store.Store) *Scheduler {
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
				// Do the scheduling
				s.schedule()
			}
		}
	}()
}

// schedule looks in the store for new work to do and schedules it.
func (s *Scheduler) schedule() {
	// Do we have space left in our buffer?
	if len(s.pipelines) >= schedulerBufferLimit {
		// No space left. Exit.
		gaia.Cfg.Logger.Debug("scheduler buffer overflow. Cannot schedule new pipelines...")
		return
	}

	// TODO: Implement schedule
}

// SchedulePipeline schedules a pipeline. That means we create a new schedule object
// and save it in our store. The scheduler will later pick up this schedule object
// and will continue the work.
func (s *Scheduler) SchedulePipeline(p *gaia.Pipeline) error {
	// Load the run history of the pipeline
	history, err := storeService.PipelineGetRunHistory(p)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot access pipeline run history bucket", "error", err.Error())
		return err
	}

	// Check if history is empty
	if history == nil {
		// Create new history object
		history = &gaia.PipelineRunHistory{
			ID:      p.ID,
			History: []gaia.PipelineRun{},
		}
	}

	// Find the highest id and increment by one
	var highestID int
	for _, run := range history.History {
		if run.ID > highestID {
			highestID = run.ID
		}
	}
	highestID++

	// Create new scheduled pipeline run
	run := gaia.PipelineRun{
		ID:           highestID,
		ScheduleDate: time.Now(),
	}

	// Add scheduled pipeline run to history
	history.History = append(history.History, run)

	// Put history into store
	return storeService.PipelinePutRunHistory(history)
}

// SetPipelineJobs uses the plugin system to get all jobs from the given pipeline.
// This function is blocking and might take some time.
func (s *Scheduler) SetPipelineJobs(p *gaia.Pipeline) error {
	// Create the start command for the pipeline
	c := createPipelineCmd(p)
	if c == nil {
		gaia.Cfg.Logger.Debug("cannot set pipeline jobs", "error", errCreateCMDForPipeline.Error(), "pipeline", p)
		return errCreateCMDForPipeline
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

// createPipelineCmd creates the execute command for the plugin system
// dependent on the plugin type.
func createPipelineCmd(p *gaia.Pipeline) *exec.Cmd {
	c := &exec.Cmd{}

	// Dependent on the pipeline type
	switch p.Type {
	case gaia.GOLANG:
		c.Path = p.ExecPath
	default:
		c = nil
	}

	return c
}
