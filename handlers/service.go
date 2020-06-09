package handlers

import (
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/security/rbac"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	"github.com/gaia-pipeline/gaia/workers/scheduler/service"
)

// Dependencies define dependencies for this service.
type Dependencies struct {
	Scheduler       service.GaiaScheduler
	PipelineService pipeline.Service
	Certificate     security.CAAPI
	RBACService     rbac.Service
}

// GaiaHandler defines handler functions throughout Gaia.
type GaiaHandler struct {
	deps Dependencies
}

// NewGaiaHandler creates a new handler service with the required dependencies.
func NewGaiaHandler(deps Dependencies) *GaiaHandler {
	return &GaiaHandler{deps: deps}
}
