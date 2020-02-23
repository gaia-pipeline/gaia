package pipeline

import (
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/workers/scheduler/service"
)

// Dependencies defines dependencies which this service needs to operate.
type Dependencies struct {
	Scheduler service.GaiaScheduler
}

type gaiaPipelineService struct {
	deps Dependencies
}

// PipelineService defines a scheduler service.
type PipelineService interface {
	CreatePipeline(p *gaia.CreatePipeline)
	InitTicker()
}

// NewGaiaPipelineService creates a pipeline service with its required dependencies already wired up
func NewGaiaPipelineService(deps Dependencies) PipelineService {
	return &gaiaPipelineService{deps: deps}
}
