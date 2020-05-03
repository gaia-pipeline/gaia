package handlers

import (
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/handlers/providers/pipelines"
	"github.com/gaia-pipeline/gaia/handlers/providers/workers"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
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

	// Endpoints for Gaia primary instance
	if gaia.Cfg.Mode == gaia.ModeServer {
		// Users
		e.POST(p+"login", UserLogin)
		e.GET(p+"users", UserGetAll)
		e.POST(p+"user/password", UserChangePassword)
		e.DELETE(p+"user/:username", UserDelete)
		e.GET(p+"user/:username/permissions", UserGetPermissions)
		e.PUT(p+"user/:username/permissions", UserPutPermissions)
		e.POST(p+"user", UserAdd)
		e.PUT(p+"user/:username/reset-trigger-token", UserResetTriggerToken)

		perms := e.Group(p + "permission")
		perms.GET("", PermissionGetAll)

		// Pipelines
		// Create pipeline provider
		pipelineProvider := pipelines.NewPipelineProvider(pipelines.Dependencies{
			Scheduler:       s.deps.Scheduler,
			PipelineService: s.deps.PipelineService,
		})
		e.POST(p+"pipeline", pipelineProvider.CreatePipeline)
		e.POST(p+"pipeline/gitlsremote", pipelineProvider.PipelineGitLSRemote)
		e.GET(p+"pipeline/name", pipelineProvider.PipelineNameAvailable)
		e.POST(p+"pipeline/githook", GitWebHook)
		e.GET(p+"pipeline/created", pipelineProvider.CreatePipelineGetAll)
		e.GET(p+"pipeline", pipelineProvider.PipelineGetAll)
		e.GET(p+"pipeline/:pipelineid", pipelineProvider.PipelineGet)
		e.PUT(p+"pipeline/:pipelineid", pipelineProvider.PipelineUpdate)
		e.DELETE(p+"pipeline/:pipelineid", pipelineProvider.PipelineDelete)
		e.POST(p+"pipeline/:pipelineid/start", pipelineProvider.PipelineStart)
		e.POST(p+"pipeline/:pipelineid/:pipelinetoken/trigger", pipelineProvider.PipelineTrigger)
		e.PUT(p+"pipeline/:pipelineid/reset-trigger-token", pipelineProvider.PipelineResetToken)
		e.GET(p+"pipeline/latest", pipelineProvider.PipelineGetAllWithLatestRun)
		e.POST(p+"pipeline/periodicschedules", pipelineProvider.PipelineCheckPeriodicSchedules)

		// Settings
		e.POST(p+"settings/poll/on", SettingsPollOn)
		e.POST(p+"settings/poll/off", SettingsPollOff)
		e.GET(p+"settings/poll", SettingsPollGet)

		// PipelineRun
		e.POST(p+"pipelinerun/:pipelineid/:runid/stop", pipelineProvider.PipelineStop)
		e.GET(p+"pipelinerun/:pipelineid/:runid", pipelineProvider.PipelineRunGet)
		e.GET(p+"pipelinerun/:pipelineid", pipelineProvider.PipelineGetAllRuns)
		e.GET(p+"pipelinerun/:pipelineid/latest", pipelineProvider.PipelineGetLatestRun)
		e.GET(p+"pipelinerun/:pipelineid/:runid/log", pipelineProvider.GetJobLogs)

		// Secrets
		e.GET(p+"secrets", ListSecrets)
		e.DELETE(p+"secret/:key", RemoveSecret)
		e.POST(p+"secret", SetSecret)
		e.PUT(p+"secret/update", SetSecret)
	}

	// Worker
	// initialize the worker provider
	workerProvider := workers.NewWorkerProvider(workers.Dependencies{
		Scheduler:   s.deps.Scheduler,
		Certificate: s.deps.Certificate,
	})
	e.GET(p+"worker/secret", workerProvider.GetWorkerRegisterSecret)
	e.POST(p+"worker/register", workerProvider.RegisterWorker)
	e.GET(p+"worker/status", workerProvider.GetWorkerStatusOverview)
	e.GET(p+"worker", workerProvider.GetWorker)
	e.DELETE(p+"worker/:workerid", workerProvider.DeregisterWorker)
	e.POST(p+"worker/secret", workerProvider.ResetWorkerRegisterSecret)

	// Middleware
	e.Use(middleware.Recover())
	//e.Use(middleware.Logger())
	e.Use(middleware.BodyLimit("32M"))
	e.Use(AuthMiddleware(&AuthConfig{
		RoleCategories: rolehelper.DefaultUserRoles,
	}))

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
