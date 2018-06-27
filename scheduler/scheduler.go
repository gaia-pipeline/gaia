package scheduler

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
func (s *Scheduler) Init() error {
	// Get number of worker
	w, err := strconv.Atoi(gaia.Cfg.Worker)
	if err != nil {
		return err
	}

	// Setup worker
	for i := 0; i < w; i++ {
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

	return nil
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
		r.StartDate = time.Now()

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

		// Get all jobs
		r.Jobs, err = s.getPipelineJobs(pipeline)
		if err != nil {
			gaia.Cfg.Logger.Error("cannot get pipeline jobs before execution", "error", err.Error())

			// Update store
			r.Status = gaia.RunFailed
			s.storeService.PipelinePutRun(&r)
			continue
		}

		// Check if this pipeline has jobs declared
		if len(r.Jobs) == 0 {
			// Finish pipeline run
			s.finishPipelineRun(&r, gaia.RunSuccess)
			continue
		}

		// Create logs folder for this run
		path := filepath.Join(gaia.Cfg.WorkspacePath, strconv.Itoa(r.PipelineID), strconv.Itoa(r.ID), gaia.LogsFolderName)
		err = os.MkdirAll(path, 0700)
		if err != nil {
			gaia.Cfg.Logger.Error("cannot create pipeline run folder", "error", err.Error(), "path", path)
		}

		// Schedule jobs and execute them.
		// Also update the run in the store.
		s.scheduleJobsByPriority(&r, pipeline)
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
func (s *Scheduler) SchedulePipeline(p *gaia.Pipeline) (*gaia.PipelineRun, error) {
	// Get highest public id used for this pipeline
	highestID, err := s.storeService.PipelineGetRunHighestID(p)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot find highest pipeline run id", "error", err.Error())
		return nil, err
	}

	// increment by one
	highestID++

	// Get jobs
	jobs, err := s.getPipelineJobs(p)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get pipeline jobs during schedule", "error", err.Error(), "pipeline", p)
		return nil, err
	}

	// Create new not scheduled pipeline run
	run := gaia.PipelineRun{
		UniqueID:     uuid.Must(uuid.NewV4(), nil).String(),
		ID:           highestID,
		PipelineID:   p.ID,
		ScheduleDate: time.Now(),
		Jobs:         jobs,
		Status:       gaia.RunNotScheduled,
	}

	// Put run into store
	return &run, s.storeService.PipelinePutRun(&run)
}

// executeJob executes a single job.
// This method is blocking.
func executeJob(job *gaia.Job, p *gaia.Pipeline, logPath string, wg *sync.WaitGroup, triggerSave chan bool) {
	defer wg.Done()
	defer func() {
		triggerSave <- true
	}()

	// Set Job to running
	job.Status = gaia.JobRunning

	// Create the start command for the pipeline
	c := createPipelineCmd(p)
	if c == nil {
		gaia.Cfg.Logger.Debug("cannot execute pipeline job", "error", errCreateCMDForPipeline.Error(), "job", job)
		job.Status = gaia.JobFailed
		return
	}

	// Create new plugin instance
	pC, err := plugin.NewPlugin(c, &logPath)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initiate plugin before job execution", "error", err.Error())
		return
	}

	// Connect to plugin(pipeline)
	if err := pC.Connect(); err != nil {
		gaia.Cfg.Logger.Debug("cannot connect to pipeline", "error", err.Error(), "pipeline", p)
		job.Status = gaia.JobFailed
		return
	}
	defer pC.Close()

	// Execute job
	if err := pC.Execute(job); err != nil {
		// TODO: Show it to user
		gaia.Cfg.Logger.Debug("error during job execution", "error", err.Error(), "job", job)
		job.Status = gaia.JobFailed
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
		if job.Priority == lowestPrio && job.Status == gaia.JobWaitingExec {
			// Increase wait group by one
			wg.Add(1)

			// Execute this job in a separate goroutine
			path := filepath.Join(gaia.Cfg.WorkspacePath, strconv.Itoa(r.PipelineID), strconv.Itoa(r.ID), gaia.LogsFolderName)
			path = filepath.Join(path, strconv.FormatUint(uint64(job.ID), 10))
			go executeJob(&r.Jobs[id], p, path, &wg, triggerSave)
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
			s.finishPipelineRun(r, gaia.RunFailed)
			return
		case gaia.JobWaitingExec:
			notExecJob = true
		}
	}

	// All jobs have been executed
	if !notExecJob {
		s.finishPipelineRun(r, gaia.RunSuccess)
		return
	}

	// Run scheduleJobsByPriority again until all jobs have been executed
	s.scheduleJobsByPriority(r, p)
}

// getJobResultsAndStore
func (s *Scheduler) getJobResultsAndStore(triggerSave chan bool, r *gaia.PipelineRun) {
	for range triggerSave {
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
	pC, err := plugin.NewPlugin(c, nil)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initiate plugin", "error", err.Error())
		return nil, err
	}

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
	case gaia.PTypeGolang:
		c.Path = p.ExecPath
	default:
		c = nil
	}

	return c
}

// finishPipelineRun finishes the pipeline run and stores the results.
func (s *Scheduler) finishPipelineRun(r *gaia.PipelineRun, status gaia.PipelineRunStatus) {
	// Mark pipeline run as success
	r.Status = status

	// Finish date
	r.FinishDate = time.Now()

	// Store it
	err := s.storeService.PipelinePutRun(r)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot store finished pipeline", "error", err.Error())
	}
}
