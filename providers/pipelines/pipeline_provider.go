package pipelines

import (
	"github.com/labstack/echo/v4"

	"github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	"github.com/gaia-pipeline/gaia/workers/scheduler/service"
)

// Dependencies define providers and services which this service needs.
type Dependencies struct {
	Scheduler       service.GaiaScheduler
	PipelineService pipeline.Servicer
	SettingsStore   store.SettingsStore
}

// PipelineProvider is a provider for all pipeline related operations.
type PipelineProvider struct {
	deps Dependencies
}

// PipelineProviderer defines functionality which this provider provides.
// These are used by the handler service.
type PipelineProviderer interface {
	PipelineGitLSRemote(c echo.Context) error
	CreatePipeline(c echo.Context) error
	CreatePipelineGetAll(c echo.Context) error
	PipelineNameAvailable(c echo.Context) error
	PipelineGet(c echo.Context) error
	PipelineGetAll(c echo.Context) error
	PipelineUpdate(c echo.Context) error
	PipelinePull(c echo.Context) error
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
	GitWebHook(c echo.Context) error
	SettingsPollOn(c echo.Context) error
	SettingsPollOff(c echo.Context) error
	SettingsPollGet(c echo.Context) error
	PipelinePause(c echo.Context) error
	PipelineUnPause(c echo.Context) error
}

// NewPipelineProvider creates a new provider with the needed dependencies.
func NewPipelineProvider(deps Dependencies) *PipelineProvider {
	return &PipelineProvider{deps: deps}
}
