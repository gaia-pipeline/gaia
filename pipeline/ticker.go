package pipeline

import (
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia"
)

const (
	// tickerIntervalSeconds defines how often the ticker will tick.
	// Definition in seconds.
	tickerIntervalSeconds = 5
)

var (
	// errMissingType is the error thrown when a pipeline is missing the type
	// in the file name.
	errMissingType = errors.New("couldnt find pipeline type definition")
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
				continue
			}

			// Create pipeline object and fill it with information
			p := gaia.Pipeline{
				Name:    pName,
				Type:    pType,
				Created: time.Now(),
			}

			// Append new pipeline
			GlobalActivePipelines.Append(p)
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
