package scheduler

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/store"
	uuid "github.com/satori/go.uuid"
)

const (
	// Maximum buffer limit for scheduler
	schedulerBufferLimit = 50

	// schedulerIntervalSeconds defines the interval the scheduler will look
	// for new work to schedule. Definition in seconds.
	schedulerIntervalSeconds = 3

	// errCircularDep is thrown when a circular dependency has been detected.
	errCircularDep = "circular dependency detected between %s and %s"

	// argTypeVault is the argument type vault.
	argTypeVault = "vault"

	// logFlushInterval defines the interval where logs will be flushed to disk.
	logFlushInterval = 1
)

var (
	// errCreateCMDForPipeline is thrown when we couldnt create a command to start
	// a plugin.
	errCreateCMDForPipeline = errors.New("could not create execute command for plugin")

	// Java executeable name
	javaExecName = "java"

	// Python executeable name
	pythonExecName = "python"
)

// Plugin represents the plugin implementation which is used
// during scheduling and execution.
type Plugin interface {
	// NewPlugin creates a new instance of plugin
	NewPlugin(ca security.CAAPI) Plugin

	// Init initializes the go-plugin client and generates a
	// new certificate pair for gaia and the plugin/pipeline.
	Init(command *exec.Cmd, logPath *string) error

	// Validate validates the plugin interface.
	Validate() error

	// Execute executes one job of a pipeline.
	Execute(j *gaia.Job) error

	// GetJobs returns all real jobs from the pipeline.
	GetJobs() ([]gaia.Job, error)

	// FlushLogs flushes the logs.
	FlushLogs() error

	// Close closes the connection and cleansup open file writes.
	Close()
}

// GaiaScheduler is a job scheduler for gaia pipeline runs.
type GaiaScheduler interface {
	Init() error
	SchedulePipeline(p *gaia.Pipeline, args []gaia.Argument) (*gaia.PipelineRun, error)
	SetPipelineJobs(p *gaia.Pipeline) error
	StopPipelineRun(p *gaia.Pipeline, runID int) error
}

var _ GaiaScheduler = (*Scheduler)(nil)

// Scheduler represents the schuler object
type Scheduler struct {
	// buffered channel which is used as queue
	scheduledRuns chan gaia.PipelineRun

	// storeService is an instance of store.
	// Use this to talk to the store.
	storeService store.GaiaStore

	// pluginSystem is the used plugin system.
	pluginSystem Plugin

	// ca is the instance of the CA used to handle certs.
	ca security.CAAPI

	// vault is the instance of the vault.
	vault security.VaultAPI
}

// NewScheduler creates a new instance of Scheduler.
func NewScheduler(store store.GaiaStore, pS Plugin, ca security.CAAPI, vault security.VaultAPI) *Scheduler {
	// Create new scheduler
	s := &Scheduler{
		scheduledRuns: make(chan gaia.PipelineRun, schedulerBufferLimit),
		storeService:  store,
		pluginSystem:  pS,
		ca:            ca,
		vault:         vault,
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

// prepareAndExec does the preparation and starts the execution.
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

	// Check if this pipeline has jobs declared
	if len(r.Jobs) == 0 {
		// Finish pipeline run
		s.finishPipelineRun(&r, gaia.RunSuccess)
		return
	}

	// Check if circular dependency exists
	for _, job := range r.Jobs {
		if _, err := s.checkCircularDep(job, []gaia.Job{}, []gaia.Job{}); err != nil {
			gaia.Cfg.Logger.Info("circular dependency detected", "pipeline", pipeline)
			gaia.Cfg.Logger.Info("information", "info", err.Error())

			// Update store
			s.finishPipelineRun(&r, gaia.RunFailed)
			return
		}
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
	pS := s.pluginSystem.NewPlugin(s.ca)

	// Init the plugin
	path = filepath.Join(path, gaia.LogsFileName)
	if err := pS.Init(c, &path); err != nil {
		gaia.Cfg.Logger.Debug("cannot initialize the plugin", "error", err.Error(), "pipeline", pipeline)
		s.finishPipelineRun(&r, gaia.RunFailed)
		return
	}

	// Validate the plugin(pipeline)
	if err := pS.Validate(); err != nil {
		gaia.Cfg.Logger.Debug("cannot validate pipeline", "error", err.Error(), "pipeline", pipeline)
		s.finishPipelineRun(&r, gaia.RunFailed)
		return
	}
	defer pS.Close()

	// Schedule jobs and execute them.
	// Also update the run in the store.
	s.executeScheduledJobs(r, pS)
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

// StopPipelineRun will prematurely cancel a pipeline run by killing all of its
// jobs and running processes immediately.
func (s *Scheduler) StopPipelineRun(p *gaia.Pipeline, runID int) error {

	// 1. Get all running Jobs
	// 2. Set state to failed and send a finish signal
	// 3. Store the result

	pr, err := s.storeService.PipelineGetRunByPipelineIDAndID(p.ID, runID)
	if err != nil {
		return err
	}
	if pr.Status != gaia.RunRunning {
		return errors.New("pipeline is not in running state")
	}
	for _, job := range pr.Jobs {
		if job.Status == gaia.JobRunning || job.Status == gaia.JobWaitingExec {
			job.Status = gaia.JobFailed
			job.FailPipeline = true
		}
	}
	return s.storeService.PipelinePutRun(pr)
}

var schedulerLock = sync.RWMutex{}

// SchedulePipeline schedules a pipeline. We create a new schedule object
// and save it in our store. The scheduler will later pick this up and will continue the work.
func (s *Scheduler) SchedulePipeline(p *gaia.Pipeline, args []gaia.Argument) (*gaia.PipelineRun, error) {

	// Introduce a semaphore locking here because this function can be called
	// in parallel if multiple users happen to trigger a pipeline run at the same time.
	// (or someone is just simply eager and presses (Start Pipeline) in quick successions).
	// This means that one of the calls will take slightly longer (a couple of nanoseconds)
	// while the other finishes to save the pipelinerun.
	// This is to ensure that the highest ID for the next pipeline is calculated properly.
	schedulerLock.Lock()
	defer schedulerLock.Unlock()

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

	// Load secret from vault and set it
	err = s.vault.LoadSecrets()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot load secrets from vault during schedule pipeline", "error", err.Error())
		return nil, err
	}
	// We have to go through all jobs to find the related arguments.
	// We will only pass related arguments to the specific job.
	for jobID, job := range jobs {
		if job.Args != nil {
			for argID, arg := range job.Args {
				// check if it's of type vault
				if arg.Type == argTypeVault {
					// Get & Set argument
					s, err := s.vault.Get(arg.Key)
					if err != nil {
						gaia.Cfg.Logger.Error("cannot find secret with given key in vault", "key", arg.Key, "pipeline", p)
						return nil, err
					}
					jobs[jobID].Args[argID].Value = string(s)
				} else {
					// Find related argument in given arguments
					for _, givenArg := range args {
						if arg.Key == givenArg.Key {
							jobs[jobID].Args[argID] = givenArg
						}
					}
				}
			}
		}
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

// executeJob executes a job and informs via triggerSave that the job can be saved to the store.
// This method is blocking.
func executeJob(j gaia.Job, pS Plugin, triggerSave chan gaia.Job) {
	defer func() {
		triggerSave <- j
	}()

	// Set Job to running and trigger save
	j.Status = gaia.JobRunning
	triggerSave <- j

	// Execute job
	if err := pS.Execute(&j); err != nil {
		gaia.Cfg.Logger.Debug("error during job execution", "error", err.Error(), "job", j)
	}
}

// checkCircularDep checks for circular dependencies.
// An error is thrown when one is found.
func (s *Scheduler) checkCircularDep(j gaia.Job, resolved []gaia.Job, unresolved []gaia.Job) ([]gaia.Job, error) {
	unresolved = append(unresolved, j)

DEPENDSON_LOOP:
	for _, job := range j.DependsOn {
		// Check if job is already in resolved list
		for _, resolvedJob := range resolved {
			if resolvedJob.ID == job.ID {
				continue DEPENDSON_LOOP
			}
		}

		// Check if job is alreay in unresolved list
		for _, unresolvedJob := range unresolved {
			if unresolvedJob.ID == job.ID {
				// Circular dependency detected
				// Return the conflicting dependencies
				return nil, fmt.Errorf(errCircularDep, unresolvedJob.Title, j.Title)
			}
		}

		// Resolve job
		var err error
		resolved, err = s.checkCircularDep(*job, resolved, unresolved)
		if err != nil {
			return nil, err
		}
	}

	return append(resolved, j), nil
}

// resolveDependencies resolves the dependencies of the given workload
// and sends all resolved workloads to our executeScheduler queue.
// After a workload has been send to the executeScheduler, the method will
// block and wait until the workload is done.
// This method is designed to be called recursive and blocking.
func (s *Scheduler) resolveDependencies(j gaia.Job, mw *managedWorkloads, executeScheduler chan gaia.Job, done chan bool) {
	for _, depJob := range j.DependsOn {
		// Check if this workload is already resolved
		var resolved bool
		for workload := range mw.Iter() {
			if workload.done && workload.job.ID == depJob.ID {
				resolved = true
			}
		}

		// Job has been resolved
		if resolved {
			continue
		}

		// Resolve job
		s.resolveDependencies(*depJob, mw, executeScheduler, done)
	}

	// Queue used to signal that the work should be finished soon.
	// We do not block here because this is just a pre-validation step.
	select {
	case _, ok := <-done:
		if !ok {
			return
		}
	default:
	}

	// If we are here, then the job is resolved.
	// We have to check if the job still needs to be run
	// or if another goroutine has already started the execution.
	relatedWL := mw.GetByID(j.ID)
	if !relatedWL.started {
		// Job has not been executed yet.
		// Send workload to execute scheduler.
		executeScheduler <- j

		// Wait until finished
		<-relatedWL.finishedSig
	} else if !relatedWL.done {
		// Job has been started but not finished.
		// Let us wait till finished.
		<-relatedWL.finishedSig
	}
}

// executeScheduledJobs is a small wrapper around executeScheduler which
// is responsible for finalizing the pipeline run.
func (s *Scheduler) executeScheduledJobs(r gaia.PipelineRun, pS Plugin) {
	// Start the main execute process and wait until finished.
	s.executeScheduler(&r, pS)

	// Run finished. Set pipeline status.
	var runFail bool
	for _, job := range r.Jobs {
		if job.Status != gaia.JobSuccess && job.FailPipeline == true {
			runFail = true
		}
	}

	if runFail {
		s.finishPipelineRun(&r, gaia.RunFailed)
	} else {
		s.finishPipelineRun(&r, gaia.RunSuccess)
	}
}

// executeScheduler is our main function which coordinates the
// whole execution process and dependency resolve algorithm.
func (s *Scheduler) executeScheduler(r *gaia.PipelineRun, pS Plugin) {
	// Create a queue which is used to execute the resolved workloads.
	executeScheduler := make(chan gaia.Job)

	// Done queue to signal go routines to exit.
	// This is usually used when a job failed and the whole pipeline
	// should be cancelled.
	done := make(chan bool)

	// Iterate all jobs from this run
	mw := newManagedWorkloads()
	for _, job := range r.Jobs {
		// Create new workload object
		mw.Append(workload{
			job:         job,
			finishedSig: make(chan bool),
		})

		// Start resolving go routine for this job
		go s.resolveDependencies(job, mw, executeScheduler, done)
	}

	// Create a new ticker (scheduled go routine) which periodically
	// flushes the logs buffer.
	ticker := time.NewTicker(logFlushInterval * time.Second)
	pipelineFinished := make(chan bool)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				pS.FlushLogs()
			case _, ok := <-pipelineFinished:
				if !ok {
					return
				}
			}
		}
	}()

	// Separate channel to save updates about the status of job executions.
	triggerSave := make(chan gaia.Job)

	// Let's loop until we are done
	var finalize bool
	finished := make(chan bool, 1)
	for {
		select {
		case <-finished:
			close(pipelineFinished)
			return
		case j, ok := <-triggerSave:
			if !ok {
				break
			}

			// Filter out the job
			for id, job := range r.Jobs {
				if job.ID == j.ID {
					r.Jobs[id].Status = j.Status
					r.Jobs[id].FailPipeline = j.FailPipeline
					break
				}
			}

			// Store status update
			s.storeService.PipelinePutRun(r)

			// Send signal to resolver that this job is finished.
			if j.Status == gaia.JobSuccess || j.Status == gaia.JobFailed {
				// Job is done
				wl := mw.GetByID(j.ID)
				wl.done = true
				mw.Replace(*wl)

				// Let's check if we are done and if all jobs ran successful.
				var allWLDone = true
				for wl := range mw.Iter() {
					if !wl.done {
						allWLDone = false
					}
				}

				if allWLDone && !finalize {
					close(done)
					close(executeScheduler)
					close(triggerSave)
					finished <- true
					finalize = true
				}

				// Close go-routine which was waiting for this job.
				close(wl.finishedSig)
			}

			// Dependent of the status output, decide what should happen next.
			if !finalize && j.Status == gaia.JobFailed {
				// we are entering the finalize phase
				finalize = true

				// Send done signal to all resolvers
				close(done)

				// read all jobs which are waiting to be executed to free the channel
				var channelClean = false
				for !channelClean {
					select {
					case <-executeScheduler:
						// just read from the channel
					default:
						channelClean = true
					}
				}

				// Close executeScheduler. No new jobs should be scheduled.
				close(executeScheduler)

				// A job failed. Finish all currently running jobs.
				go func() {
					// We might have still running jobs. Wait until all jobs are finished.
					for {
						var notFinishedWorkloadChan chan bool
						for singleWL := range mw.Iter() {
							if singleWL.started && !singleWL.done {
								notFinishedWorkloadChan = singleWL.finishedSig
							}
						}

						if notFinishedWorkloadChan == nil {
							break
						}

						// wait until finished
						<-notFinishedWorkloadChan
					}

					finished <- true
					close(triggerSave)
				}()
			}
		case j, ok := <-executeScheduler:
			if !ok {
				break
			}

			// Get related workload
			wl := mw.GetByID(j.ID)

			// Check if this workload has been already started by another routine.
			if !wl.started {
				// Update
				wl.started = true
				mw.Replace(*wl)

				// Start execution
				go executeJob(j, pS, triggerSave)
			}
		}
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

	// Create new Plugin instance
	pS := s.pluginSystem.NewPlugin(s.ca)

	// Init the go-plugin
	if err := pS.Init(c, nil); err != nil {
		gaia.Cfg.Logger.Debug("cannot initialize the pipeline", "error", err.Error(), "pipeline", p)
		return nil, err
	}

	// Validate the plugin(pipeline)
	if err := pS.Validate(); err != nil {
		gaia.Cfg.Logger.Debug("cannot validate pipeline", "error", err.Error(), "pipeline", p)
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
	case gaia.PTypeJava:
		// Look for java executeable
		path, err := exec.LookPath(javaExecName)
		if err != nil {
			gaia.Cfg.Logger.Debug("cannot find java executeable", "error", err.Error())
			return nil
		}

		// Build start command
		c.Path = path
		c.Args = []string{
			path,
			"-jar",
			p.ExecPath,
		}
	case gaia.PTypePython:
		// Build start command
		c.Path = "/bin/sh"
		c.Args = []string{
			"/bin/sh",
			"-c",
			". bin/activate; exec " + pythonExecName + " -c \"import pipeline; pipeline.main()\"",
		}
		c.Dir = filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpPythonFolder, p.Name)
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
