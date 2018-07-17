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

// Plugin represents the plugin implementation which is used
// during scheduling and execution.
type Plugin interface {
	// NewPlugin creates a new instance of plugin
	NewPlugin() Plugin

	// Connect initializes the connection with the execution command
	// and the log path wbere the logs should be stored.
	Connect(command *exec.Cmd, logPath *string) error

	// Execute executes one job of a pipeline.
	Execute(j *gaia.Job) error

	// GetJobs returns all real jobs from the pipeline.
	GetJobs() ([]gaia.Job, error)

	// Close closes the connection and cleansup open file writes.
	Close()
}

// Scheduler represents the schuler object
type Scheduler struct {
	// buffered channel which is used as queue
	scheduledRuns chan gaia.PipelineRun

	// storeService is an instance of store.
	// Use this to talk to the store.
	storeService *store.Store

	// pluginSystem is the used plugin system.
	pluginSystem Plugin
}

// NewScheduler creates a new instance of Scheduler.
func NewScheduler(store *store.Store, pS Plugin) *Scheduler {
	// Create new scheduler
	s := &Scheduler{
		scheduledRuns: make(chan gaia.PipelineRun, schedulerBufferLimit),
		storeService:  store,
		pluginSystem:  pS,
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

// work takes work from the scheduled run buffer channel and starts
// the preparation and execution of the pipeline. Then repeats.
func (s *Scheduler) work() {
	// This worker never stops working.
	for {
		// Take one scheduled run, block if there are no scheduled pipelines
		r := <-s.scheduledRuns

		// Prepare execution and start it
		s.prepareAndExec(r)
	}
}

// prepareAndExec does the real preparation and start the execution.
func (s *Scheduler) prepareAndExec(r gaia.PipelineRun) {
	// Mark the scheduled run as running
	r.Status = gaia.RunRunning
	r.StartDate = time.Now()

	// Update entry in store
	err := s.storeService.PipelinePutRun(&r)
	if err != nil {
		gaia.Cfg.Logger.Debug("could not put pipeline run into store during executing work", "error", err.Error())
		return
	}

	// Get related pipeline from pipeline run
	pipeline, _ := s.storeService.PipelineGet(r.PipelineID)

	// Get all jobs
	r.Jobs, err = s.getPipelineJobs(pipeline)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get pipeline jobs before execution", "error", err.Error())

		// Update store
		r.Status = gaia.RunFailed
		s.storeService.PipelinePutRun(&r)
		return
	}

	// Check if this pipeline has jobs declared
	if len(r.Jobs) == 0 {
		// Finish pipeline run
		s.finishPipelineRun(&r, gaia.RunSuccess)
		return
	}

	// Create logs folder for this run
	path := filepath.Join(gaia.Cfg.WorkspacePath, strconv.Itoa(r.PipelineID), strconv.Itoa(r.ID), gaia.LogsFolderName)
	err = os.MkdirAll(path, 0700)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create pipeline run folder", "error", err.Error(), "path", path)
	}

	// Create the start command for the pipeline
	c := createPipelineCmd(pipeline)
	if c == nil {
		gaia.Cfg.Logger.Debug("cannot create pipeline start command", "error", errCreateCMDForPipeline.Error())
		s.finishPipelineRun(&r, gaia.RunFailed)
		return
	}

	// Create new plugin instance
	pS := s.pluginSystem.NewPlugin()

	// Connect to plugin(pipeline)
	path = filepath.Join(path, gaia.LogsFileName)
	if err := pS.Connect(c, &path); err != nil {
		gaia.Cfg.Logger.Debug("cannot connect to pipeline", "error", err.Error(), "pipeline", pipeline)
		s.finishPipelineRun(&r, gaia.RunFailed)
		return
	}
	defer pS.Close()

	// Schedule jobs and execute them.
	// Also update the run in the store.
	s.scheduleJobsByPriority(r, pS)
}

// schedule looks in the store for new work and schedules it.
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
func executeJob(job gaia.Job, pS Plugin, wg *sync.WaitGroup, triggerSave chan gaia.Job) {
	defer wg.Done()
	defer func() {
		triggerSave <- job
	}()

	// Set Job to running and trigger save
	job.Status = gaia.JobRunning
	triggerSave <- job

	// Execute job
	if err := pS.Execute(&job); err != nil {
		// TODO: Show it to user
		gaia.Cfg.Logger.Debug("error during job execution", "error", err.Error(), "job", job)
		job.Status = gaia.JobFailed
		return
	}

	// If we are here, the job execution was ok
	job.Status = gaia.JobSuccess
}

// scheduleJobsByPriority schedules the given jobs by their respective
// priority. This method is designed to be recursive and blocking.
// If jobs have the same priority, they will be executed in parallel.
func (s *Scheduler) scheduleJobsByPriority(r gaia.PipelineRun, pS Plugin) {
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
	triggerSave := make(chan gaia.Job)
	done := make(chan bool)
	for id, job := range r.Jobs {
		if job.Priority == lowestPrio && job.Status == gaia.JobWaitingExec {
			// Increase wait group by one
			wg.Add(1)

			// Execute this job in a separate goroutine
			go executeJob(r.Jobs[id], pS, &wg, triggerSave)
		}
	}

	// Create channel for storing job run results and spawn results routine
	go func() {
		for {
			j, open := <-triggerSave

			// Channel has been closed
			if !open {
				done <- true
				return
			}

			// Filter out the job
			for id, job := range r.Jobs {
				if job.ID == j.ID {
					r.Jobs[id].Status = j.Status
					break
				}
			}

			// Store update
			s.storeService.PipelinePutRun(&r)
		}
	}()

	// Wait until all jobs have been finished and close results channel
	wg.Wait()
	close(triggerSave)
	<-done

	// Check if a job has been failed. If so, stop execution.
	// We also check if all jobs has been executed.
	var notExecJob bool
	for _, job := range r.Jobs {
		switch job.Status {
		case gaia.JobFailed:
			s.finishPipelineRun(&r, gaia.RunFailed)
			return
		case gaia.JobWaitingExec:
			notExecJob = true
		}
	}

	// All jobs have been executed
	if !notExecJob {
		s.finishPipelineRun(&r, gaia.RunSuccess)
		return
	}

	// Run scheduleJobsByPriority again until all jobs have been executed
	s.scheduleJobsByPriority(r, pS)
}

// getPipelineJobs uses the plugin system to get all jobs from the given pipeline.
func (s *Scheduler) getPipelineJobs(p *gaia.Pipeline) ([]gaia.Job, error) {
	// Create the start command for the pipeline
	c := createPipelineCmd(p)
	if c == nil {
		gaia.Cfg.Logger.Debug("cannot set pipeline jobs", "error", errCreateCMDForPipeline.Error(), "pipeline", p)
		return nil, errCreateCMDForPipeline
	}

	// Create new Plugin instance
	pS := s.pluginSystem.NewPlugin()

	// Connect to plugin(pipeline)
	if err := pS.Connect(c, nil); err != nil {
		gaia.Cfg.Logger.Debug("cannot connect to pipeline", "error", err.Error(), "pipeline", p)
		return nil, err
	}
	defer pS.Close()

	return pS.GetJobs()
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
