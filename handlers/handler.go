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

	// API router group.
	apiGrp := e.Group(p)

	// API router group with auth middleware.
	apiAuthGrp := e.Group(p, authMiddleware(&AuthConfig{
		RoleCategories: rolehelper.DefaultUserRoles,
		rbacEnforcer:   s.deps.RBACService,
	}))

	// Endpoints for Gaia primary instance
	if gaia.Cfg.Mode == gaia.ModeServer {
		// Users
		apiGrp.POST("login", UserLogin)

		apiAuthGrp.GET("users", UserGetAll)
		apiAuthGrp.POST("user/password", UserChangePassword)
		apiAuthGrp.DELETE("user/:username", UserDelete)
		apiAuthGrp.GET("user/:username/permissions", UserGetPermissions)
		apiAuthGrp.PUT("user/:username/permissions", UserPutPermissions)
		apiAuthGrp.POST("user", UserAdd)
		apiAuthGrp.PUT("user/:username/reset-trigger-token", UserResetTriggerToken)

		apiAuthGrp.GET("permission", PermissionGetAll)

		// Pipelines
		// Create pipeline provider
		pipelineProvider := pipelines.NewPipelineProvider(pipelines.Dependencies{
			Scheduler:       s.deps.Scheduler,
			PipelineService: s.deps.PipelineService,
			SettingsStore:   s.deps.Store,
		})
		apiAuthGrp.POST("pipeline", pipelineProvider.CreatePipeline)
		apiAuthGrp.POST("pipeline/gitlsremote", pipelineProvider.PipelineGitLSRemote)
		apiAuthGrp.GET("pipeline/name", pipelineProvider.PipelineNameAvailable)
		apiAuthGrp.GET("pipeline/created", pipelineProvider.CreatePipelineGetAll)
		apiAuthGrp.GET("pipeline", pipelineProvider.PipelineGetAll)
		apiAuthGrp.GET("pipeline/:pipelineid", pipelineProvider.PipelineGet)
		apiAuthGrp.PUT("pipeline/:pipelineid", pipelineProvider.PipelineUpdate)
		apiAuthGrp.DELETE("pipeline/:pipelineid", pipelineProvider.PipelineDelete)
		apiAuthGrp.POST("pipeline/:pipelineid/start", pipelineProvider.PipelineStart)
		apiAuthGrp.PUT("pipeline/:pipelineid/reset-trigger-token", pipelineProvider.PipelineResetToken)
		apiAuthGrp.POST("pipeline/:pipelineid/pull", pipelineProvider.PipelinePull)
		apiAuthGrp.GET("pipeline/latest", pipelineProvider.PipelineGetAllWithLatestRun)
		apiAuthGrp.POST("pipeline/periodicschedules", pipelineProvider.PipelineCheckPeriodicSchedules)
		apiGrp.POST("pipeline/githook", pipelineProvider.GitWebHook)
		apiGrp.POST("pipeline/:pipelineid/:pipelinetoken/trigger", pipelineProvider.PipelineTrigger)

		// Settings
		settingsHandler := newSettingsHandler(s.deps.Store)
		apiAuthGrp.POST("settings/poll/on", pipelineProvider.SettingsPollOn)
		apiAuthGrp.POST("settings/poll/off", pipelineProvider.SettingsPollOff)
		apiAuthGrp.GET("settings/poll", pipelineProvider.SettingsPollGet)
		apiAuthGrp.GET("settings/rbac", settingsHandler.rbacGet)
		apiAuthGrp.PUT("settings/rbac", settingsHandler.rbacToggle)

		// PipelineRun
		apiAuthGrp.POST("pipelinerun/:pipelineid/:runid/stop", pipelineProvider.PipelineStop)
		apiAuthGrp.GET("pipelinerun/:pipelineid/:runid", pipelineProvider.PipelineRunGet)
		apiAuthGrp.GET("pipelinerun/:pipelineid", pipelineProvider.PipelineGetAllRuns)
		apiAuthGrp.GET("pipelinerun/:pipelineid/latest", pipelineProvider.PipelineGetLatestRun)
		apiAuthGrp.GET("pipelinerun/:pipelineid/:runid/log", pipelineProvider.GetJobLogs)

		// Secrets
		apiAuthGrp.GET("secrets", ListSecrets)
		apiAuthGrp.DELETE("secret/:key", RemoveSecret)
		apiAuthGrp.POST("secret", SetSecret)
		apiAuthGrp.PUT("secret/update", SetSecret)

		// RBAC
		rbacHandler := rbacHandler{
			svc: s.deps.RBACService,
		}
		// RBAC - Management
		apiAuthGrp.GET("rbac/roles", rbacHandler.getAllRoles)
		apiAuthGrp.PUT("rbac/roles/:role", rbacHandler.addRole)
		apiAuthGrp.DELETE("rbac/roles/:role", rbacHandler.deleteRole)
		apiAuthGrp.PUT("rbac/roles/:role/attach/:username", rbacHandler.attachRole)
		apiAuthGrp.DELETE("rbac/roles/:role/attach/:username", rbacHandler.detachRole)
		apiAuthGrp.GET("rbac/roles/:role/attached", rbacHandler.getRolesAttachedUsers)
		// RBAC - Users
		apiAuthGrp.GET("users/:username/rbac/roles", rbacHandler.getUserAttachedRoles)
	}

	// Worker
	// initialize the worker provider
	workerProvider := workers.NewWorkerProvider(workers.Dependencies{
		Scheduler:   s.deps.Scheduler,
		Certificate: s.deps.Certificate,
	})
	apiAuthGrp.GET("worker/secret", workerProvider.GetWorkerRegisterSecret)
	apiAuthGrp.GET("worker/status", workerProvider.GetWorkerStatusOverview)
	apiAuthGrp.GET("worker", workerProvider.GetWorker)
	apiAuthGrp.DELETE("worker/:workerid", workerProvider.DeregisterWorker)
	apiAuthGrp.POST("worker/secret", workerProvider.ResetWorkerRegisterSecret)
	apiGrp.POST("worker/register", workerProvider.RegisterWorker)

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
