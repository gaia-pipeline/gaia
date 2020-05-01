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

	userRBAC := NewRBACMiddleware(rolehelper.UserCategory)
	userPermRBAC := NewRBACMiddleware(rolehelper.UserPermissionCategory)
	pipelineRBAC := NewRBACMiddleware(rolehelper.PipelineCategory)
	pipelineRunRBAC := NewRBACMiddleware(rolehelper.PipelineRunCategory)
	secretsRBAC := NewRBACMiddleware(rolehelper.SecretCategory)
	workerRBAC := NewRBACMiddleware(rolehelper.WorkerCategory)

	// Endpoints for Gaia primary instance
	if gaia.Cfg.Mode == gaia.ModeServer {
		// Users
		e.POST(p+"login", UserLogin)
		e.PUT(p+"user/:username/reset-trigger-token", UserResetTriggerToken)
		e.GET(p+"users", UserGetAll, userRBAC.Do(rolehelper.ListRole))
		e.POST(p+"user/password", UserChangePassword, userRBAC.Do(rolehelper.ChangePasswordRole))
		e.DELETE(p+"user/:username", UserDelete, userRBAC.Do(rolehelper.DeleteRole))
		e.POST(p+"user", UserAdd, userRBAC.Do(rolehelper.CreateRole))
		e.GET(p+"user/:username/permissions", UserGetPermissions, userPermRBAC.Do(rolehelper.GetRole))
		e.PUT(p+"user/:username/permissions", UserPutPermissions, userPermRBAC.Do(rolehelper.UpdateRole))

		// Permissions
		ph := permissionHandler{
			defaultRoles: rolehelper.DefaultUserRoles,
		}
		e.GET(p+"permission", ph.PermissionGetAll)

		// Pipelines
		// Create pipeline provider
		pipelineProvider := pipelines.NewPipelineProvider(pipelines.Dependencies{
			Scheduler:       s.deps.Scheduler,
			PipelineService: s.deps.PipelineService,
		})
		e.POST(p+"pipeline/:pipelineid/:pipelinetoken/trigger", pipelineProvider.PipelineTrigger)
		e.PUT(p+"pipeline/:pipelineid/reset-trigger-token", pipelineProvider.PipelineResetToken)
		e.POST(p+"pipeline/periodicschedules", pipelineProvider.PipelineCheckPeriodicSchedules)
		e.POST(p+"pipeline", pipelineProvider.CreatePipeline, pipelineRBAC.Do(rolehelper.CreateRole))
		e.POST(p+"pipeline/gitlsremote", pipelineProvider.PipelineGitLSRemote, pipelineRBAC.Do(rolehelper.CreateRole))
		e.GET(p+"pipeline/name", pipelineProvider.PipelineNameAvailable, pipelineRBAC.Do(rolehelper.CreateRole))
		e.POST(p+"pipeline/githook", GitWebHook, pipelineRBAC.Do(rolehelper.CreateRole))
		e.GET(p+"pipeline/created", pipelineProvider.CreatePipelineGetAll, pipelineRBAC.Do(rolehelper.ListRole))
		e.GET(p+"pipeline", pipelineProvider.PipelineGetAll, pipelineRBAC.Do(rolehelper.ListRole))
		e.GET(p+"pipeline/latest", pipelineProvider.PipelineGetAllWithLatestRun, pipelineRBAC.Do(rolehelper.ListRole))
		e.GET(p+"pipeline/:pipelineid", pipelineProvider.PipelineGet, pipelineRBAC.Do(rolehelper.GetRole))
		e.PUT(p+"pipeline/:pipelineid", pipelineProvider.PipelineUpdate, pipelineRBAC.Do(rolehelper.UpdateRole))
		e.DELETE(p+"pipeline/:pipelineid", pipelineProvider.PipelineDelete, pipelineRBAC.Do(rolehelper.DeleteRole))
		e.POST(p+"pipeline/:pipelineid/start", pipelineProvider.PipelineStart, pipelineRBAC.Do(rolehelper.StartRole))

		// Pipeline Run
		e.POST(p+"pipelinerun/:pipelineid/:runid/stop", pipelineProvider.PipelineStop, pipelineRunRBAC.Do(rolehelper.StopRole))
		e.GET(p+"pipelinerun/:pipelineid/:runid", pipelineProvider.PipelineRunGet, pipelineRunRBAC.Do(rolehelper.GetRole))
		e.GET(p+"pipelinerun/:pipelineid/latest", pipelineProvider.PipelineGetLatestRun, pipelineRunRBAC.Do(rolehelper.GetRole))
		e.GET(p+"pipelinerun/:pipelineid", pipelineProvider.PipelineGetAllRuns, pipelineRunRBAC.Do(rolehelper.ListRole))
		e.GET(p+"pipelinerun/:pipelineid/:runid/log", pipelineProvider.GetJobLogs, pipelineRunRBAC.Do(rolehelper.LogsRole))

		// Settings
		e.POST(p+"settings/poll/on", SettingsPollOn)
		e.POST(p+"settings/poll/off", SettingsPollOff)
		e.GET(p+"settings/poll", SettingsPollGet)

		// Secrets
		e.GET(p+"secrets", ListSecrets, secretsRBAC.Do(rolehelper.ListRole))
		e.DELETE(p+"secret/:key", RemoveSecret, secretsRBAC.Do(rolehelper.DeleteRole))
		e.POST(p+"secret", SetSecret, secretsRBAC.Do(rolehelper.CreateRole))
		e.PUT(p+"secret/update", SetSecret, secretsRBAC.Do(rolehelper.UpdateRole))
	}

	// Worker
	// initialize the worker provider
	workerProvider := workers.NewWorkerProvider(workers.Dependencies{
		Scheduler: s.deps.Scheduler,
	})
	e.POST(p+"worker/register", workerProvider.RegisterWorker)
	e.GET(p+"worker/secret", workerProvider.GetWorkerRegisterSecret, workerRBAC.Do(rolehelper.GetRegistrationSecretRole))
	e.GET(p+"worker/status", workerProvider.GetWorkerStatusOverview, workerRBAC.Do(rolehelper.GetOverviewRole))
	e.GET(p+"worker", workerProvider.GetWorker, workerRBAC.Do(rolehelper.GetWorkerRole))
	e.DELETE(p+"worker/:workerid", workerProvider.DeregisterWorker, workerRBAC.Do(rolehelper.DeregisterWorkerRole))
	e.POST(p+"worker/secret", workerProvider.ResetWorkerRegisterSecret, workerRBAC.Do(rolehelper.ResetWorkerRegisterSecretRole))

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
