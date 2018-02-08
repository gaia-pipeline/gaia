package pipeline

import (
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/plugin"
)

// setPipelineJobs uses the plugin system to get all
// jobs from the given pipeline.
// This function is blocking and might take some time.
func setPipelineJobs(p *gaia.Pipeline) error {
	// Create the start command for the pipeline
	c := createPipelineCmd(p)
	if c == nil {
		gaia.Cfg.Logger.Debug("cannot set pipeline jobs", "error", errMissingType.Error(), "pipeline", p)
		return errMissingType
	}

	// Create new plugin instance
	pC := plugin.NewPlugin(c)

	// Connect to plugin(pipeline)
	if err := pC.Connect(); err != nil {
		gaia.Cfg.Logger.Debug("cannot connect to pipeline", "error", err.Error(), "pipeline", p)
		return err
	}
	defer pC.Close()

	// Get jobs
	jobs, err := pC.GetJobs()
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot get jobs from pipeline", "error", err.Error(), "pipeline", p)
		return err
	}
	p.Jobs = jobs

	return nil
}
