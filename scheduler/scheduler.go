package scheduler

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
		// Take one scheduled run, block if there are no scheduled pipelines
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
			r.Status = gaia.RunFailed
		} else if pipeline == nil {
			gaia.Cfg.Logger.Debug("wanted to execute job for pipeline which does not exist", "run", r)
			r.Status = gaia.RunFailed
		}

		if r.Status == gaia.RunFailed {
			// Update entry in store
			err = s.storeService.PipelinePutRun(&r)
			if err != nil {
				gaia.Cfg.Logger.Debug("could not put pipeline run into store during executing work", "error", err.Error())
			}
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
	for id := range scheduled {
		// push scheduled run into our channel
		s.scheduledRuns <- (*scheduled[id])

		// Mark them as scheduled
		scheduled[id].Status = gaia.RunScheduled

		// Update entry in store
		err = s.storeService.PipelinePutRun(scheduled[id])
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
		PipelineID:   p.ID,
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

	// Check if this pipeline has jobs declared
	if len(r.Jobs) == 0 {
		return
	}

	// Schedule jobs and execute them.
	// Also update the run in the store.
	s.scheduleJobsByPriority(r, p)
}

// executeJob executes a single job.
// This method is blocking.
func executeJob(job *gaia.Job, p *gaia.Pipeline, wg *sync.WaitGroup, triggerSave chan bool) {
	defer wg.Done()
	defer func() {
		triggerSave <- true
	}()

	// In testmode we do not test this.
	// TODO: Bad testing. Fix this asap!
	if gaia.Cfg.TestMode {
		job.Status = gaia.JobSuccess
		return
	}

	// Lets be pessimistic
	job.Status = gaia.JobFailed

	// Create the start command for the pipeline
	c := createPipelineCmd(p)
	if c == nil {
		gaia.Cfg.Logger.Debug("cannot execute pipeline job", "error", errCreateCMDForPipeline.Error(), "job", job)
		return
	}

	// Create new plugin instance
	pC := plugin.NewPlugin(c)

	// Connect to plugin(pipeline)
	if err := pC.Connect(); err != nil {
		gaia.Cfg.Logger.Debug("cannot connect to pipeline", "error", err.Error(), "pipeline", p)
		return
	}
	defer pC.Close()

	// Execute job
	if err := pC.Execute(job); err != nil {
		// TODO: Show it to user
		gaia.Cfg.Logger.Debug("error during job execution", "error", err.Error(), "job", job)
	}

	// If we are here, the job execution was ok
	job.Status = gaia.JobSuccess
}

// scheduleJobsByPriority schedules the given jobs by their respective
// priority. This method is designed to be recursive and blocking.
// If jobs have the same priority, they will be executed in parallel.
func (s *Scheduler) scheduleJobsByPriority(r *gaia.PipelineRun, p *gaia.Pipeline) {
	// Do a prescheduling and set it to the first waiting job
	var lowestPrio int64
	for _, job := range r.Jobs {
		if job.Status == gaia.JobWaitingExec {
			lowestPrio = job.Priority
			break
		}
	}

	// Find the job with the lowest priority
	for _, job := range r.Jobs {
		if job.Priority < lowestPrio && job.Status == gaia.JobWaitingExec {
			lowestPrio = job.Priority
		}
	}

	// We might have multiple jobs with the same priority.
	// It means these jobs should be started in parallel.
	var wg sync.WaitGroup
	triggerSave := make(chan bool)
	for id, job := range r.Jobs {
		if job.Priority == lowestPrio {
			// Increase wait group by one
			wg.Add(1)

			// Execute this job in a separate goroutine
			go executeJob(&r.Jobs[id], p, &wg, triggerSave)
		}
	}

	// Create channel for storing job run results and spawn results routine
	go s.getJobResultsAndStore(triggerSave, r)

	// Wait until all jobs have been finished and close results channel
	wg.Wait()
	close(triggerSave)

	// Check if a job has been failed. If so, stop execution.
	// We also check if all jobs has been executed.
	var notExecJob bool
	for _, job := range r.Jobs {
		switch job.Status {
		case gaia.JobFailed:
			return
		case gaia.JobWaitingExec:
			notExecJob = true
		}
	}

	// All jobs have been executed
	if !notExecJob {
		return
	}

	// Run scheduleJobsByPriority again until all jobs have been executed
	s.scheduleJobsByPriority(r, p)
}

// getJobResultsAndStore
func (s *Scheduler) getJobResultsAndStore(triggerSave chan bool, r *gaia.PipelineRun) {
	for _ = range triggerSave {
		// TODO: Bad testing. Fix this asap!
		if gaia.Cfg.TestMode {
			continue
		}

		// Store update
		s.storeService.PipelinePutRun(r)
	}
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
