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
	uuid "github.com/satori/go.uuid"
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
				p := GlobalActivePipelines.Get(pName)
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
						setPipelineJobsTicker(p)

						// Replace pipeline
						if ok := GlobalActivePipelines.Replace(*p); !ok {
							gaia.Cfg.Logger.Debug("cannot replace pipeline in global pipeline list", "pipeline", p)
						}
					}
				}

				// Its already in the list
				continue
			}

			// Create pipeline object and fill it with information
			p := gaia.Pipeline{
				ID:       uuid.Must(uuid.NewV4()).String(),
				Name:     pName,
				Type:     pType,
				ExecPath: gaia.Cfg.PipelinePath + string(os.PathSeparator) + file.Name(),
				Created:  time.Now(),
			}

			// Let us try to start the plugin and receive all implemented jobs
			setPipelineJobsTicker(&p)

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

func setPipelineJobsTicker(p *gaia.Pipeline) {
	err := setPipelineJobs(p)
	if err != nil {
		// We were not able to get jobs from the pipeline.
		// We set the Md5Checksum for later to try it again.
		p.Md5Checksum, err = getMd5Checksum(p.ExecPath)
		if err != nil {
			gaia.Cfg.Logger.Debug("cannot calculate md5 checksum for pipeline", "error", err.Error(), "pipeline", p)
		}
	} else {
		// Reset md5 checksum in case we already set it
		p.Md5Checksum = nil
	}
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
