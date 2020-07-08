package pipeline

import (
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/workers/scheduler/service"
)

// Dependencies defines dependencies which this service needs to operate.
type Dependencies struct {
	Scheduler service.GaiaScheduler
}

// GaiaPipelineService defines a pipeline service provider providing pipeline related functions.
type GaiaPipelineService struct {
	deps Dependencies
}

// Servicer defines a scheduler service.
type Servicer interface {
	CreatePipeline(p *gaia.CreatePipeline)
	InitTicker()
	CheckActivePipelines()
	UpdateRepository(p *gaia.Pipeline) error
	UpdateAllCurrentPipelines()
	StartPoller() error
	StopPoller() error
}

// NewGaiaPipelineService creates a pipeline service with its required dependencies already wired up
func NewGaiaPipelineService(deps Dependencies) *GaiaPipelineService {
	return &GaiaPipelineService{deps: deps}
}
