package pipeline

import (
	"bytes"
	"crypto/sha256"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia"
	scheduler "github.com/gaia-pipeline/gaia/scheduler"
	"github.com/gaia-pipeline/gaia/store"
)

const (
	// tickerIntervalSeconds defines how often the ticker will tick.
	// Definition in seconds.
	tickerIntervalSeconds = 5
)

// storeService is an instance of store.
// Use this to talk to the store.
var storeService *store.Store

// schedulerService is an instance of scheduler.
var schedulerService *scheduler.Scheduler

// InitTicker inititates the pipeline ticker.
// This periodic job will check for new pipelines.
func InitTicker(store *store.Store, scheduler *scheduler.Scheduler) {
	// Init global active pipelines slice
	GlobalActivePipelines = NewActivePipelines()

	// Save instances
	storeService = store
	schedulerService = scheduler

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
}

// checkActivePipelines looks up all files in the pipeline folder.
// Every file will be handled as an active pipeline and therefore
// saved in the global active pipelines slice.
func checkActivePipelines() {
	files, err := ioutil.ReadDir(gaia.Cfg.PipelinePath)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot read pipelines folder", "error", err.Error(), "path", gaia.Cfg.PipelinePath)
	} else {
		// Iterate all found pipelines
		for _, file := range files {
			n := strings.TrimSpace(strings.ToLower(file.Name()))

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
						// Let us try again to start the plugin and receive all implemented jobs
						schedulerService.SetPipelineJobs(p)

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

			// We couldn't finde the pipeline. Create a new one.
			var shouldStore = false
			if pipeline == nil {
				// Create pipeline object and fill it with information
				pipeline = &gaia.Pipeline{
					Name:     pName,
					Type:     pType,
					ExecPath: filepath.Join(gaia.Cfg.PipelinePath, file.Name()),
					Created:  time.Now(),
				}

				// We should store it
				shouldStore = true
			}

			// We calculate a SHA256 Checksum and store it.
			// We use this to estimate if a pipeline has been changed.
			pipeline.SHA256Sum, err = getSHA256Sum(pipeline.ExecPath)
			if err != nil {
				gaia.Cfg.Logger.Debug("cannot calculate sha256 checksum for pipeline", "error", err.Error(), "pipeline", pipeline)
				continue
			}

			// Let us try to start the plugin and receive all implemented jobs
			schedulerService.SetPipelineJobs(pipeline)

			// Put pipeline into store only when it was new created.
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
