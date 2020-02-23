package providers

import (
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	"github.com/gaia-pipeline/gaia/workers/scheduler/service"
	"github.com/labstack/echo"
)

// Dependencies define providers and services which this service needs.
type Dependencies struct {
	Scheduler       service.GaiaScheduler
	PipelineService pipeline.PipelineService
}

type pipelineProvider struct {
	deps Dependencies
}

// PipelineProvider defines functionality which this provider provides.
// These are used by the handler service.
type PipelineProvider interface {
	PipelineGitLSRemote(c echo.Context) error
	CreatePipeline(c echo.Context) error
	CreatePipelineGetAll(c echo.Context) error
	PipelineNameAvailable(c echo.Context) error
	PipelineGet(c echo.Context) error
	PipelineGetAll(c echo.Context) error
	PipelineUpdate(c echo.Context) error
	PipelineDelete(c echo.Context) error
	PipelineTrigger(c echo.Context) error
	PipelineResetToken(c echo.Context) error
	PipelineTriggerAuth(c echo.Context) error
	PipelineStart(c echo.Context) error
	PipelineGetAllWithLatestRun(c echo.Context) error
	PipelineCheckPeriodicSchedules(c echo.Context) error
	PipelineStop(c echo.Context) error
	PipelineRunGet(c echo.Context) error
	PipelineGetAllRuns(c echo.Context) error
	PipelineGetLatestRun(c echo.Context) error
	GetJobLogs(c echo.Context) error
}

// NewPipelineProvider creates a new provider with the needed dependencies.
func NewPipelineProvider(deps Dependencies) PipelineProvider {
	return &pipelineProvider{deps: deps}
}
