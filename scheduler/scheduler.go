package pipeline

import (
	"errors"
	"os/exec"
	"sync"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/plugin"
	"github.com/gaia-pipeline/gaia/store"
	uuid "github.com/satori/go.uuid"
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
)

// Scheduler represents the schuler object
type Scheduler struct {
	// buffered channel which is used as queue
	scheduledRuns chan gaia.PipelineRun

	// storeService is an instance of store.
	// Use this to talk to the store.
	storeService *store.Store
}

// NewScheduler creates a new instance of Scheduler.
func NewScheduler(store *store.Store) *Scheduler {
	// Create new scheduler
	s := &Scheduler{
		scheduledRuns: make(chan gaia.PipelineRun, schedulerBufferLimit),
		storeService:  store,
	}

	return s
}

// Init initializes the scheduler.
func (s *Scheduler) Init() {
	// Setup workers
	for i := 0; i < gaia.Cfg.Workers; i++ {
		go s.work()
	}

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

// work takes work from the scheduled run buffer channel
// and executes the pipeline. Then repeats.
func (s *Scheduler) work() {
	// This worker never stops working.
	for {
		// Take one scheduled run
		r := <-s.scheduledRuns

		// Mark the scheduled run as running
		r.Status = gaia.RunRunning

		// Update entry in store
		err := s.storeService.PipelinePutRun(&r)
		if err != nil {
			gaia.Cfg.Logger.Debug("could not put pipeline run into store during executing work", "error", err.Error())
			continue
		}

		// Get related pipeline from pipeline run
		pipeline, err := s.storeService.PipelineGet(r.PipelineID)
		if err != nil {
			gaia.Cfg.Logger.Debug("cannot access pipeline during execution", "error", err.Error())
			continue
		} else if pipeline == nil {
			gaia.Cfg.Logger.Debug("wanted to execute pipeline which does not exist", "run", r)
			continue
		}

		// Start pipeline run process
		s.executePipeline(pipeline, &r)
	}
}

// schedule looks in the store for new work to do and schedules it.
func (s *Scheduler) schedule() {
	// Do we have space left in our buffer?
	if len(s.scheduledRuns) >= schedulerBufferLimit {
		// No space left. Exit.
		return
	}

	// Get scheduled pipelines but limit the returning number of elements.
	scheduled, err := s.storeService.PipelineGetScheduled(schedulerBufferLimit)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot get scheduled pipelines", "error", err.Error())
		return
	}

	// Iterate scheduled runs
	for _, run := range scheduled {
		// push scheduled run into our channel
		s.scheduledRuns <- (*run)

		// Mark them as scheduled
		run.Status = gaia.RunScheduled

		// Update entry in store
		err = s.storeService.PipelinePutRun(run)
		if err != nil {
			gaia.Cfg.Logger.Debug("could not put pipeline run into store", "error", err.Error())
		}
	}
}

// SchedulePipeline schedules a pipeline. We create a new schedule object
// and save it in our store. The scheduler will later pick up this schedule object
// and will continue the work.
func (s *Scheduler) SchedulePipeline(p *gaia.Pipeline) error {
	// Get highest public id used for this pipeline
	highestID, err := s.storeService.PipelineGetRunHighestID(p)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot find highest pipeline run id", "error", err.Error())
		return err
	}

	// increment by one
	highestID++

	// Create new not scheduled pipeline run
	run := gaia.PipelineRun{
		UniqueID:     uuid.Must(uuid.NewV4()).String(),
		ID:           highestID,
		ScheduleDate: time.Now(),
		Status:       gaia.RunNotScheduled,
	}

	// Put run into store
	return s.storeService.PipelinePutRun(&run)
}

// executePipeline executes the given pipeline and updates it status periodically.
func (s *Scheduler) executePipeline(p *gaia.Pipeline, r *gaia.PipelineRun) {
	// Set pessimistic values
	r.Status = gaia.RunFailed
	r.RunDate = time.Now()

	// Get all jobs
	var err error
	r.Jobs, err = s.getPipelineJobs(p)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get pipeline jobs before execution", "error", err.Error())

		// Update store
		s.storeService.PipelinePutRun(r)
		return
	}

}

func executeJob(job *gaia.Job, wg *sync.WaitGroup) {
	// TODO
	wg.Done()
}

func executeJobs(jobs []*gaia.Job) {
	// We finished all jobs, exit recursive execution.
	if len(jobs) == 0 {
		return
	}

	// Find the job with the lowest priority
	var lowestPrio int32
	for id, job := range jobs {
		if job.Priority < lowestPrio || id == 0 {
			lowestPrio = job.Priority
		}
	}

	// We allocate a new slice for jobs with higher priority.
	// And also a slice for jobs which we execute now.
	var nextJobs []*gaia.Job
	var execJobs []*gaia.Job

	// We might have multiple jobs with the same priority.
	// It means these jobs should be started in parallel.
	var wg sync.WaitGroup
	for _, job := range jobs {
		if job.Priority == lowestPrio {
			// Increase wait group by one
			wg.Add(1)
			execJobs = append(execJobs, job)

			// Execute this job in a separate goroutine
			go executeJob(job, &wg)
		} else {
			// We add this job to the next list
			nextJobs = append(nextJobs, job)
		}
	}

	// Wait until all jobs has been finished
	wg.Wait()

	// Check if a job has been failed. If so, stop execution.
	for _, job := range execJobs {
		if !job.Success {
			return
		}
	}

	// Run executeJobs again until all jobs have been executed
	executeJobs(nextJobs)
}

// getPipelineJobs uses the plugin system to get all jobs from the given pipeline.
func (s *Scheduler) getPipelineJobs(p *gaia.Pipeline) ([]gaia.Job, error) {
	// Create the start command for the pipeline
	c := createPipelineCmd(p)
	if c == nil {
		gaia.Cfg.Logger.Debug("cannot set pipeline jobs", "error", errCreateCMDForPipeline.Error(), "pipeline", p)
		return nil, errCreateCMDForPipeline
	}

	// Create new plugin instance
	pC := plugin.NewPlugin(c)

	// Connect to plugin(pipeline)
	if err := pC.Connect(); err != nil {
		gaia.Cfg.Logger.Debug("cannot connect to pipeline", "error", err.Error(), "pipeline", p)
		return nil, err
	}
	defer pC.Close()

	return pC.GetJobs()
}

// SetPipelineJobs uses the plugin system to get all jobs from the given pipeline.
// This function is blocking and might take some time.
func (s *Scheduler) SetPipelineJobs(p *gaia.Pipeline) error {
	// Get jobs
	jobs, err := s.getPipelineJobs(p)
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
