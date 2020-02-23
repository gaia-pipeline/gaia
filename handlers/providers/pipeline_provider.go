package providers

import (
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	"github.com/gaia-pipeline/gaia/workers/scheduler/service"
)

type Dependencies struct {
	Scheduler       service.GaiaScheduler
	PipelineService pipeline.PipelineService
}

type pipelineProvider struct {
	deps Dependencies
}

func NewPipelineProvider(deps Dependencies) *pipelineProvider {
	return &pipelineProvider{deps: deps}
}
