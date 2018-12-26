package pipeline

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/robfig/cron"
)

const (
	// tickerIntervalSeconds defines how often the ticker will tick.
	// Definition in seconds.
	tickerIntervalSeconds = 5
)

// InitTicker inititates the pipeline ticker.
// This periodic job will check for new pipelines.
func InitTicker() {
	// Init global active pipelines slice
	GlobalActivePipelines = NewActivePipelines()

	// Check immediately to make sure we fill the list as fast as possible.
	checkActivePipelines()

	// Create ticker
	ticker := time.NewTicker(tickerIntervalSeconds * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				checkActivePipelines()
			}
		}
	}()

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
				}
			}
		}()
	}
}

// checkActivePipelines looks up all files in the pipeline folder.
// Every file will be handled as an active pipeline and therefore
// saved in the global active pipelines slice.
func checkActivePipelines() {
	schedulerService, _ := services.SchedulerService()
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
			pName := getRealPipelineName(n, pType)
			// Add the real pipeline name to the slice of existing pipeline names.
			existingPipelineNames = append(existingPipelineNames, pName)
			if GlobalActivePipelines.Contains(pName) {
				// If SHA256Sum is set, we should check if pipeline has been changed.
				p := GlobalActivePipelines.GetByName(pName)
				if p != nil && p.SHA256Sum != nil {
					// Get SHA256 Checksum
					checksum, err := getSHA256Sum(filepath.Join(gaia.Cfg.PipelinePath, file.Name()))
					if err != nil {
						gaia.Cfg.Logger.Debug("cannot calculate SHA256 checksum for pipeline", "error", err.Error(), "pipeline", p)
						continue
					}

					// Pipeline has been changed?
					if bytes.Compare(p.SHA256Sum, checksum) != 0 {
						// update pipeline if needed
						if err = updatePipeline(p); err != nil {
							storeService.PipelinePut(p)
							gaia.Cfg.Logger.Debug("cannot update pipeline", "error", err.Error(), "pipeline", p)
							continue
						}

						// Let us try again to start the plugin and receive all implemented jobs
						if err = schedulerService.SetPipelineJobs(p); err != nil {
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
				}
				shouldStore = true
			}

			// We calculate a SHA256 Checksum and store it.
			// We use this to estimate if a pipeline has been changed.
			pipelineCheckSum, err := getSHA256Sum(pipeline.ExecPath)
			if err != nil {
				gaia.Cfg.Logger.Debug("cannot calculate sha256 checksum for pipeline", "error", err.Error(), "pipeline", pipeline)
				continue
			}

			// update pipeline if needed
			if bytes.Compare(pipeline.SHA256Sum, pipelineCheckSum) != 0 {
				if err = updatePipeline(pipeline); err != nil {
					storeService.PipelinePut(pipeline)
					gaia.Cfg.Logger.Error("cannot update pipeline", "error", err.Error(), "pipeline", pipeline)
					continue
				}
				storeService.PipelinePut(pipeline)
			}

			// Let us try to start the plugin and receive all implemented jobs
			if err = schedulerService.SetPipelineJobs(pipeline); err != nil {
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
				for _, cron := range pipeline.PeriodicSchedules {
					pipeline.CronInst.AddFunc(cron, func() {
						_, err := schedulerService.SchedulePipeline(pipeline, []gaia.Argument{})
						if err != nil {
							gaia.Cfg.Logger.Error("cannot schedule pipeline from periodic schedule", "error", err, "pipeline", pipeline)
							return
						}

						// Log scheduling information
						gaia.Cfg.Logger.Info("pipeline has been automatically scheduled by periodic scheduling:", "name", pipeline.Name)
					})
				}

				// Start schedule process.
				pipeline.CronInst.Start()
			}

			// We encountered a drop-in pipeline previously. Now is the time to save it.
			if shouldStore {
				storeService.PipelinePut(pipeline)
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
	}

	return gaia.PTypeUnknown, errMissingType
}

// getRealPipelineName removes the suffix from the pipeline name.
func getRealPipelineName(n string, pType gaia.PipelineType) string {
	return strings.TrimSuffix(n, typeDelimiter+pType.String())
}

// getSHA256Sum accepts a path to a file.
// It load's the file and calculates a SHA256 Checksum and returns it.
func getSHA256Sum(path string) ([]byte, error) {
	// Open file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Create sha256 obj and insert bytes
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	// return sha256 checksum
	return h.Sum(nil), nil
}
