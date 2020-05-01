package handlers

import (
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

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

	// Endpoints for Gaia primary instance
	if gaia.Cfg.Mode == gaia.ModeServer {
		// Users
		{
			e.POST(p+"login", UserLogin)
			e.PUT(p+"user/:username/reset-trigger-token", UserResetTriggerToken)
			rbac := NewRBACMiddleware(rolehelper.UserCategory)
			e.GET(p+"users", UserGetAll, rbac.Do(rolehelper.ListRole))
			e.POST(p+"user/password", UserChangePassword, rbac.Do(rolehelper.ChangePasswordRole))
			e.DELETE(p+"user/:username", UserDelete, rbac.Do(rolehelper.DeleteRole))
			e.POST(p+"user", UserAdd, rbac.Do(rolehelper.CreateRole))
		}

		// User Permissions
		{
			perms := e.Group(p + "permission")
			perms.GET("", PermissionGetAll)
			rbac := NewRBACMiddleware(rolehelper.UserPermissionCategory)
			e.GET(p+"user/:username/permissions", UserGetPermissions, rbac.Do(rolehelper.GetRole))
			e.PUT(p+"user/:username/permissions", UserPutPermissions, rbac.Do(rolehelper.UpdateRole))
		}

		// Pipelines
		// Create pipeline provider
		pipelineProvider := pipelines.NewPipelineProvider(pipelines.Dependencies{
			Scheduler:       s.deps.Scheduler,
			PipelineService: s.deps.PipelineService,
		})
		{
			e.POST(p+"pipeline/:pipelineid/:pipelinetoken/trigger", pipelineProvider.PipelineTrigger)
			e.PUT(p+"pipeline/:pipelineid/reset-trigger-token", pipelineProvider.PipelineResetToken)
			e.POST(p+"pipeline/periodicschedules", pipelineProvider.PipelineCheckPeriodicSchedules)
			rbac := NewRBACMiddleware(rolehelper.PipelineCategory)
			e.POST(p+"pipeline", pipelineProvider.CreatePipeline, rbac.Do(rolehelper.CreateRole))
			e.POST(p+"pipeline/gitlsremote", pipelineProvider.PipelineGitLSRemote, rbac.Do(rolehelper.CreateRole))
			e.GET(p+"pipeline/name", pipelineProvider.PipelineNameAvailable, rbac.Do(rolehelper.CreateRole))
			e.POST(p+"pipeline/githook", GitWebHook, rbac.Do(rolehelper.CreateRole))
			e.GET(p+"pipeline/created", pipelineProvider.CreatePipelineGetAll, rbac.Do(rolehelper.ListRole))
			e.GET(p+"pipeline", pipelineProvider.PipelineGetAll, rbac.Do(rolehelper.ListRole))
			e.GET(p+"pipeline/latest", pipelineProvider.PipelineGetAllWithLatestRun, rbac.Do(rolehelper.ListRole))
			e.GET(p+"pipeline/:pipelineid", pipelineProvider.PipelineGet, rbac.Do(rolehelper.GetRole))
			e.PUT(p+"pipeline/:pipelineid", pipelineProvider.PipelineUpdate, rbac.Do(rolehelper.UpdateRole))
			e.DELETE(p+"pipeline/:pipelineid", pipelineProvider.PipelineDelete, rbac.Do(rolehelper.DeleteRole))
			e.POST(p+"pipeline/:pipelineid/start", pipelineProvider.PipelineStart, rbac.Do(rolehelper.StartRole))
		}

		// Pipeline Run
		{
			rbac := NewRBACMiddleware(rolehelper.PipelineRunCategory)
			e.POST(p+"pipelinerun/:pipelineid/:runid/stop", pipelineProvider.PipelineStop, rbac.Do(rolehelper.StopRole))
			e.GET(p+"pipelinerun/:pipelineid/:runid", pipelineProvider.PipelineRunGet, rbac.Do(rolehelper.GetRole))
			e.GET(p+"pipelinerun/:pipelineid/latest", pipelineProvider.PipelineGetLatestRun, rbac.Do(rolehelper.GetRole))
			e.GET(p+"pipelinerun/:pipelineid", pipelineProvider.PipelineGetAllRuns, rbac.Do(rolehelper.ListRole))
			e.GET(p+"pipelinerun/:pipelineid/:runid/log", pipelineProvider.GetJobLogs, rbac.Do(rolehelper.LogsRole))
		}

		// Settings
		{
			e.POST(p+"settings/poll/on", SettingsPollOn)
			e.POST(p+"settings/poll/off", SettingsPollOff)
			e.GET(p+"settings/poll", SettingsPollGet)
		}

		// Secrets
		{
			rbac := NewRBACMiddleware(rolehelper.SecretCategory)
			e.GET(p+"secrets", ListSecrets, rbac.Do(rolehelper.ListRole))
			e.DELETE(p+"secret/:key", RemoveSecret, rbac.Do(rolehelper.DeleteRole))
			e.POST(p+"secret", SetSecret, rbac.Do(rolehelper.CreateRole))
			e.PUT(p+"secret/update", SetSecret, rbac.Do(rolehelper.UpdateRole))
		}
	}

	// Worker
	// initialize the worker provider
	workerProvider := workers.NewWorkerProvider(workers.Dependencies{
		Scheduler: s.deps.Scheduler,
	})
	{
		rbac := NewRBACMiddleware(rolehelper.WorkerCategory)
		e.GET(p+"worker/secret", workerProvider.GetWorkerRegisterSecret, rbac.Do(rolehelper.GetRegistrationSecretRole))
		e.POST(p+"worker/register", workerProvider.RegisterWorker)
		e.GET(p+"worker/status", workerProvider.GetWorkerStatusOverview, rbac.Do(rolehelper.GetOverviewRole))
		e.GET(p+"worker", workerProvider.GetWorker, rbac.Do(rolehelper.GetWorkerRole))
		e.DELETE(p+"worker/:workerid", workerProvider.DeregisterWorker, rbac.Do(rolehelper.DeregisterWorkerRole))
		e.POST(p+"worker/secret", workerProvider.ResetWorkerRegisterSecret, rbac.Do(rolehelper.ResetWorkerRegisterSecretRole))
	}

	// Middleware
	e.Use(middleware.Recover())
	// e.Use(middleware.Logger())
	e.Use(middleware.BodyLimit("32M"))
	e.Use(AuthMiddleware())

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
