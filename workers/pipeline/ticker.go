package pipeline

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/filehelper"
	"github.com/gaia-pipeline/gaia/helper/pipelinehelper"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/robfig/cron"
)

const (
	// tickerIntervalSeconds defines how often the ticker will tick.
	// Definition in seconds.
	tickerIntervalSeconds = 5
)

var pollerDone = make(chan struct{}, 1)
var isPollerRunning bool

// StopPoller sends a done signal to the polling timer if it's running.
func StopPoller() error {
	if isPollerRunning {
		isPollerRunning = false
		select {
		case pollerDone <- struct{}{}:
		default:
		}
		return nil
	}
	return errors.New("poller is not running")
}

// StartPoller starts the poller if it's not already running.
func StartPoller() error {
	if isPollerRunning {
		return errors.New("poller is already running")
	}
	if gaia.Cfg.Poll {
		if gaia.Cfg.PVal < 1 || gaia.Cfg.PVal > 99 {
			errorMessage := fmt.Sprintf("Invalid value defined for poll interval. Will be using default of 1. Value was: %d, should be between 1-99.", gaia.Cfg.PVal)
			gaia.Cfg.Logger.Info(errorMessage)
			gaia.Cfg.PVal = 1
		}
		pollTicker := time.NewTicker(time.Duration(gaia.Cfg.PVal) * time.Minute)
		go func() {
			defer pollTicker.Stop()
			for {
				select {
				case <-pollTicker.C:
					updateAllCurrentPipelines()
				case <-pollerDone:
					pollTicker.Stop()
					return
				}
			}
		}()
		isPollerRunning = true
	}
	return nil
}

// InitTicker initiates the pipeline ticker.
// This periodic job will check for new pipelines.
func (s *gaiaPipelineService) InitTicker() {
	// Init global active pipelines slice
	GlobalActivePipelines = NewActivePipelines()

	// Check immediately to make sure we fill the list as fast as possible.
	s.CheckActivePipelines()

	// Create ticker
	ticker := time.NewTicker(tickerIntervalSeconds * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.CheckActivePipelines()
				updateWorker()
			}
		}
	}()

	_ = StartPoller()
}

// checkActivePipelines looks up all files in the pipeline folder.
// Every file will be handled as an active pipeline and therefore
// saved in the global active pipelines slice.
func (s *gaiaPipelineService) CheckActivePipelines() {
	storeService, _ := services.StorageService()
	var existingPipelineNames []string
	files, err := ioutil.ReadDir(gaia.Cfg.PipelinePath)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot read pipelines folder", "error", err.Error(), "path", gaia.Cfg.PipelinePath)
	} else {
		// Iterate all found pipelines
		for _, file := range files {
			n := strings.TrimSpace(file.Name())

			// Get pipeline type
			pType, err := getPipelineType(n)
			if err != nil {
				gaia.Cfg.Logger.Debug("at least one pipeline in pipeline folder is missing the type definition")
				gaia.Cfg.Logger.Debug("Info", "name", n)
				gaia.Cfg.Logger.Error("error thrown", "error", err.Error())
				continue
			}

			// Get real pipeline name and check if the global active pipelines slice
			// already contains it.
			pName := pipelinehelper.GetRealPipelineName(n, pType)
			// Add the real pipeline name to the slice of existing pipeline names.
			existingPipelineNames = append(existingPipelineNames, pName)
			if GlobalActivePipelines.Contains(pName) {
				// If SHA256Sum is set, we should check if pipeline has been changed.
				p := GlobalActivePipelines.GetByName(pName)
				if p != nil && p.SHA256Sum != nil {
					// Get SHA256 Checksum
					checksum, err := filehelper.GetSHA256Sum(filepath.Join(gaia.Cfg.PipelinePath, file.Name()))
					if err != nil {
						gaia.Cfg.Logger.Debug("cannot calculate SHA256 checksum for pipeline", "error", err.Error(), "pipeline", p)
						continue
					}

					// Pipeline has been changed?
					if !bytes.Equal(p.SHA256Sum, checksum) {
						// update pipeline if needed
						if err = updatePipeline(p); err != nil {
							_ = storeService.PipelinePut(p)
							gaia.Cfg.Logger.Debug("cannot update pipeline", "error", err.Error(), "pipeline", p)
							continue
						}

						// Let us try again to start the plugin and receive all implemented jobs
						if err = s.deps.Scheduler.SetPipelineJobs(p); err != nil {
							// Mark that this pipeline is broken.
							p.IsNotValid = true
						}

						// Replace pipeline
						if ok := GlobalActivePipelines.Replace(*p); !ok {
							gaia.Cfg.Logger.Debug("cannot replace pipeline in global pipeline list", "pipeline", p)
						}
					}
				}

				// Its already in the list
				continue
			}

			// Get pipeline from store.
			pipeline, err := storeService.PipelineGetByName(pName)
			if err != nil {
				// If we have an error here we are in trouble.
				gaia.Cfg.Logger.Error("cannot access pipelines bucket. Data corrupted?", "error", err.Error())
				continue
			}

			// Pipeline is a drop-in build. Set up a template for it.
			shouldStore := false
			if pipeline == nil {
				pipeline = &gaia.Pipeline{
					Name:     pName,
					Type:     pType,
					ExecPath: filepath.Join(gaia.Cfg.PipelinePath, file.Name()),
					Created:  time.Now(),
					Tags:     []string{pType.String()},
				}
				shouldStore = true
			}

			// We calculate a SHA256 Checksum and store it.
			// We use this to estimate if a pipeline has been changed.
			pipelineCheckSum, err := filehelper.GetSHA256Sum(pipeline.ExecPath)
			if err != nil {
				gaia.Cfg.Logger.Debug("cannot calculate sha256 checksum for pipeline", "error", err.Error(), "pipeline", pipeline)
				continue
			}

			// update pipeline if needed
			if !bytes.Equal(pipeline.SHA256Sum, pipelineCheckSum) {
				if err = updatePipeline(pipeline); err != nil {
					_ = storeService.PipelinePut(pipeline)
					gaia.Cfg.Logger.Error("cannot update pipeline", "error", err.Error(), "pipeline", pipeline)
					continue
				}
				_ = storeService.PipelinePut(pipeline)
			}

			// Let us try to start the plugin and receive all implemented jobs
			if err = s.deps.Scheduler.SetPipelineJobs(pipeline); err != nil {
				// Mark that this pipeline is broken.
				pipeline.IsNotValid = true
				gaia.Cfg.Logger.Error("cannot get pipeline jobs", "error", err.Error(), "pipeline", pipeline)
			}

			// Set up periodic schedules of this pipeline.
			if !pipeline.IsNotValid && len(pipeline.PeriodicSchedules) > 0 {
				// We prevent side effects here and make sure
				// that no scheduling is already running.
				if pipeline.CronInst != nil {
					pipeline.CronInst.Stop()
				}
				pipeline.CronInst = cron.New()

				// Iterate over all cron schedules.
				for _, schedule := range pipeline.PeriodicSchedules {
					err := pipeline.CronInst.AddFunc(schedule, func() {
						_, err := s.deps.Scheduler.SchedulePipeline(pipeline, gaia.StartReasonScheduled, []*gaia.Argument{})
						if err != nil {
							gaia.Cfg.Logger.Error("cannot schedule pipeline from periodic schedule", "error", err, "pipeline", pipeline)
							// stopping the pipeline scheduler if there was an error in any of the pipeline scheduling
							// example: The pipeline was deleted
							pipeline.CronInst.Stop()
							return
						}

						// Log scheduling information
						gaia.Cfg.Logger.Info("pipeline has been automatically scheduled by periodic scheduling:", "name", pipeline.Name)
					})
					if err != nil {
						gaia.Cfg.Logger.Error("failed to schedule periodic schedule", "error", err)
						// do not schedule if there was an error
						pipeline.CronInst.Stop()
						continue
					}
				}

				// Start schedule process.
				pipeline.CronInst.Start()
			}

			// We encountered a drop-in pipeline previously. Now is the time to save it.
			if shouldStore {
				_ = storeService.PipelinePut(pipeline)
			}

			// We do not update the pipeline in store if it already exists there.
			// We only updated the SHA256 Checksum and the jobs but this is not importent
			// to store and should not have any side effects.

			// Append new pipeline
			GlobalActivePipelines.Append(*pipeline)
		}
	}
	GlobalActivePipelines.RemoveDeletedPipelines(existingPipelineNames)
}

// updateWorker checks the latest worker information and determines the status
// of the worker.
func updateWorker() {
	// Get memdb service
	db, err := services.DefaultMemDBService()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get memdb service via updateWorker", "error", err)
		return
	}

	// Get all worker
	workers := db.GetAllWorker()

	// The maximum last contact time is time.Now() - 5 minutes.
	// Workers with older last contact time will be marked inactive.
	lastContactTime := time.Now().Add(-5 * time.Minute)

	// Iterate all worker
	for _, worker := range workers {
		if worker.LastContact.Before(lastContactTime) {
			if worker.Status == gaia.WorkerActive {
				// Last contact was more than 5 minutes ago.
				// Worker is now marked as inactive.
				worker.Status = gaia.WorkerInactive

				if err := db.UpsertWorker(worker, true); err != nil {
					gaia.Cfg.Logger.Error("failed to store update to worker via updateWorker", "error", err)
				}
			}
		} else if worker.Status == gaia.WorkerInactive {
			// Worker is marked inactive but we got contact.
			// Mark it as healthy.
			worker.Status = gaia.WorkerActive

			if err := db.UpsertWorker(worker, true); err != nil {
				gaia.Cfg.Logger.Error("failed to store update to worker via updateWorker", "error", err)
			}
		}
	}
}

// getPipelineType looks up for specific suffix on the given file name.
// If found, returns the pipeline type.
func getPipelineType(n string) (gaia.PipelineType, error) {
	s := strings.Split(n, typeDelimiter)

	// Length must be higher than one
	if len(s) < 2 {
		return gaia.PTypeUnknown, errMissingType
	}

	// Get last element and look for type
	t := s[len(s)-1]
	switch t {
	case gaia.PTypeGolang.String():
		return gaia.PTypeGolang, nil
	case gaia.PTypeJava.String():
		return gaia.PTypeJava, nil
	case gaia.PTypePython.String():
		return gaia.PTypePython, nil
	case gaia.PTypeCpp.String():
		return gaia.PTypeCpp, nil
	case gaia.PTypeRuby.String():
		return gaia.PTypeRuby, nil
	case gaia.PTypeNodeJS.String():
		return gaia.PTypeNodeJS, nil
	}

	return gaia.PTypeUnknown, errMissingType
}
