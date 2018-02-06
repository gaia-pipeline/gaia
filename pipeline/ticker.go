package pipeline

import (
	"io/ioutil"
	"time"

	"github.com/gaia-pipeline/gaia"
)

const (
	tickerIntervalSeconds = 5
)

// InitTicker inititates the pipeline ticker.
// This periodic job will check for new pipelines.
func InitTicker() {
	// Create ticker
	ticker := time.NewTicker(tickerIntervalSeconds * time.Second)
	quit := make(chan struct{})

	// Actual ticker implementation
	go func() {
		for {
			select {
			case <-ticker.C:
				files, err := ioutil.ReadDir(gaia.Cfg.PipelinePath)
				if err != nil {
					gaia.Cfg.Logger.Error("cannot read pipelines folder", "error", err.Error(), "path", gaia.Cfg.PipelinePath)
				} else {
					// Iterate all found pipeline
					for _, file := range files {
						// TODO: Create for every file a pipeline object
						// and store it in a global pipeline array
						gaia.Cfg.Logger.Debug("pipeline found", "name", file.Name())
					}
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
