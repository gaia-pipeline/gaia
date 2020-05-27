package handlers

import (
	rice "github.com/GeertJohan/go.rice"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/handlers/providers/pipelines"
	"github.com/gaia-pipeline/gaia/handlers/providers/workers"
)

var (
	// List of secret keys which cannot be modified via the normal Vault API.
	ignoredVaultKeys []string
)

// InitHandlers initializes(registers) all handlers.
func (s *GaiaHandler) InitHandlers(e *echo.Echo) error {
	// Define prefix
	p := "/api/" + gaia.APIVersion + "/"

	// --- Register handlers at echo instance ---

	authGrp := e.Group("", AuthMiddleware(&AuthConfig{
		RoleCategories: rolehelper.DefaultUserRoles,
	}))

	// Endpoints for Gaia primary instance
	if gaia.Cfg.Mode == gaia.ModeServer {
		// Users
		e.POST(p+"login", UserLogin)
		authGrp.GET(p+"users", UserGetAll)
		authGrp.POST(p+"user/password", UserChangePassword)
		authGrp.DELETE(p+"user/:username", UserDelete)
		authGrp.GET(p+"user/:username/permissions", UserGetPermissions)
		authGrp.PUT(p+"user/:username/permissions", UserPutPermissions)
		authGrp.POST(p+"user", UserAdd)
		authGrp.PUT(p+"user/:username/reset-trigger-token", UserResetTriggerToken)

		perms := e.Group(p + "permission")
		perms.GET("", PermissionGetAll)

		// Pipelines
		// Create pipeline provider
		pipelineProvider := pipelines.NewPipelineProvider(pipelines.Dependencies{
			Scheduler:       s.deps.Scheduler,
			PipelineService: s.deps.PipelineService,
		})
		authGrp.POST(p+"pipeline", pipelineProvider.CreatePipeline)
		authGrp.POST(p+"pipeline/gitlsremote", pipelineProvider.PipelineGitLSRemote)
		authGrp.GET(p+"pipeline/name", pipelineProvider.PipelineNameAvailable)
		e.POST(p+"pipeline/githook", GitWebHook)
		authGrp.GET(p+"pipeline/created", pipelineProvider.CreatePipelineGetAll)
		authGrp.GET(p+"pipeline", pipelineProvider.PipelineGetAll)
		authGrp.GET(p+"pipeline/:pipelineid", pipelineProvider.PipelineGet)
		authGrp.PUT(p+"pipeline/:pipelineid", pipelineProvider.PipelineUpdate)
		authGrp.DELETE(p+"pipeline/:pipelineid", pipelineProvider.PipelineDelete)
		authGrp.POST(p+"pipeline/:pipelineid/start", pipelineProvider.PipelineStart)
		authGrp.POST(p+"pipeline/:pipelineid/:pipelinetoken/trigger", pipelineProvider.PipelineTrigger)
		authGrp.PUT(p+"pipeline/:pipelineid/reset-trigger-token", pipelineProvider.PipelineResetToken)
		authGrp.GET(p+"pipeline/latest", pipelineProvider.PipelineGetAllWithLatestRun)
		authGrp.POST(p+"pipeline/periodicschedules", pipelineProvider.PipelineCheckPeriodicSchedules)

		// Settings
		authGrp.POST(p+"settings/poll/on", SettingsPollOn)
		authGrp.POST(p+"settings/poll/off", SettingsPollOff)
		authGrp.GET(p+"settings/poll", SettingsPollGet)

		// PipelineRun
		authGrp.POST(p+"pipelinerun/:pipelineid/:runid/stop", pipelineProvider.PipelineStop)
		authGrp.GET(p+"pipelinerun/:pipelineid/:runid", pipelineProvider.PipelineRunGet)
		authGrp.GET(p+"pipelinerun/:pipelineid", pipelineProvider.PipelineGetAllRuns)
		authGrp.GET(p+"pipelinerun/:pipelineid/latest", pipelineProvider.PipelineGetLatestRun)
		authGrp.GET(p+"pipelinerun/:pipelineid/:runid/log", pipelineProvider.GetJobLogs)

		// Secrets
		authGrp.GET(p+"secrets", ListSecrets)
		authGrp.DELETE(p+"secret/:key", RemoveSecret)
		authGrp.POST(p+"secret", SetSecret)
		authGrp.PUT(p+"secret/update", SetSecret)
	}

	// Worker
	// initialize the worker provider
	workerProvider := workers.NewWorkerProvider(workers.Dependencies{
		Scheduler:   s.deps.Scheduler,
		Certificate: s.deps.Certificate,
	})
	authGrp.GET(p+"worker/secret", workerProvider.GetWorkerRegisterSecret)
	e.POST(p+"worker/register", workerProvider.RegisterWorker)
	authGrp.GET(p+"worker/status", workerProvider.GetWorkerStatusOverview)
	authGrp.GET(p+"worker", workerProvider.GetWorker)
	authGrp.DELETE(p+"worker/:workerid", workerProvider.DeregisterWorker)
	authGrp.POST(p+"worker/secret", workerProvider.ResetWorkerRegisterSecret)

	// Middleware
	e.Use(middleware.Recover())
	// e.Use(middleware.Logger())
	e.Use(middleware.BodyLimit("32M"))

	// Extra options
	e.HideBanner = true

	// Are we in production mode?
	if !gaia.Cfg.DevMode {
		staticAssets, err := rice.FindBox("../frontend/dist")
		if err != nil {
			gaia.Cfg.Logger.Error("Cannot find assets in production mode.")
			return err
		}

		// Register handler for static assets
		assetHandler := http.FileServer(staticAssets.HTTPBox())
		e.GET("/", echo.WrapHandler(assetHandler))
		e.GET("/favicon.ico", echo.WrapHandler(assetHandler))
		e.GET("/css/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/js/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/fonts/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/img/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
	}

	// Setup ignored vault keys which cannot be modified directly via the Vault API
	ignoredVaultKeys = make([]string, 0, 1)
	ignoredVaultKeys = append(ignoredVaultKeys, gaia.WorkerRegisterKey)

	return nil
}
