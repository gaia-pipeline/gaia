package pipeline

import (
	"bytes"
	"crypto/md5"
	"io"
	"io/ioutil"
	"os"
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
				// If Md5Checksum is set, we should check if pipeline has been changed.
				p := GlobalActivePipelines.GetByName(pName)
				if p != nil && p.Md5Checksum != nil {
					// Get MD5 Checksum
					checksum, err := getMd5Checksum(gaia.Cfg.PipelinePath + string(os.PathSeparator) + file.Name())
					if err != nil {
						gaia.Cfg.Logger.Debug("cannot calculate md5 checksum for pipeline", "error", err.Error(), "pipeline", p)
						continue
					}

					// Pipeline has been changed?
					if bytes.Compare(p.Md5Checksum, checksum) != 0 {
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
					ExecPath: gaia.Cfg.PipelinePath + string(os.PathSeparator) + file.Name(),
					Created:  time.Now(),
				}

				// We should store it
				shouldStore = true
			}

			// We calculate a MD5 Checksum and store it.
			// We use this to estimate if a pipeline has been changed.
			pipeline.Md5Checksum, err = getMd5Checksum(pipeline.ExecPath)
			if err != nil {
				gaia.Cfg.Logger.Debug("cannot calculate md5 checksum for pipeline", "error", err.Error(), "pipeline", pipeline)
				continue
			}

			// Let us try to start the plugin and receive all implemented jobs
			schedulerService.SetPipelineJobs(pipeline)

			// Put pipeline into store only when it was new created.
			if shouldStore {
				storeService.PipelinePut(pipeline)
			}

			// We do not update the pipeline in store if it already exists there.
			// We only updated the Md5 Checksum and the jobs but this is not importent
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
		return gaia.UNKNOWN, errMissingType
	}

	// Get last element and look for type
	t := s[len(s)-1]
	switch t {
	case gaia.GOLANG.String():
		return gaia.GOLANG, nil
	}

	return gaia.UNKNOWN, errMissingType
}

// getRealPipelineName removes the suffix from the pipeline name.
func getRealPipelineName(n string, pType gaia.PipelineType) string {
	return strings.TrimSuffix(n, typeDelimiter+pType.String())
}

func getMd5Checksum(file string) ([]byte, error) {
	// Open file
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Create md5 obj and insert bytes
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	// return md5 checksum
	return h.Sum(nil), nil
}
