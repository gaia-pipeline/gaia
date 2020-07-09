package handlers

import (
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/gaia-pipeline/gaia"
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
		userHandler := NewUserHandler(s.deps.Store, s.deps.RBACService)

		apiGrp.POST("login", userHandler.UserLogin)

		apiAuthGrp.GET("users", userHandler.UserGetAll)
		apiAuthGrp.POST("user/password", userHandler.UserChangePassword)
		apiAuthGrp.DELETE("user/:username", userHandler.UserDelete)
		apiAuthGrp.GET("user/:username/permissions", userHandler.UserGetPermissions)
		apiAuthGrp.PUT("user/:username/permissions", userHandler.UserPutPermissions)
		apiAuthGrp.POST("user", userHandler.UserAdd)
		apiAuthGrp.PUT("user/:username/reset-trigger-token", userHandler.UserResetTriggerToken)

		apiAuthGrp.GET("permission", PermissionGetAll)

		// Pipelines
		// Create pipeline provider
		apiAuthGrp.POST("pipeline", s.deps.PipelineProvider.CreatePipeline)
		apiAuthGrp.POST("pipeline/gitlsremote", s.deps.PipelineProvider.PipelineGitLSRemote)
		apiAuthGrp.GET("pipeline/name", s.deps.PipelineProvider.PipelineNameAvailable)
		apiAuthGrp.GET("pipeline/created", s.deps.PipelineProvider.CreatePipelineGetAll)
		apiAuthGrp.GET("pipeline", s.deps.PipelineProvider.PipelineGetAll)
		apiAuthGrp.GET("pipeline/:pipelineid", s.deps.PipelineProvider.PipelineGet)
		apiAuthGrp.PUT("pipeline/:pipelineid", s.deps.PipelineProvider.PipelineUpdate)
		apiAuthGrp.DELETE("pipeline/:pipelineid", s.deps.PipelineProvider.PipelineDelete)
		apiAuthGrp.POST("pipeline/:pipelineid/start", s.deps.PipelineProvider.PipelineStart)
		apiAuthGrp.PUT("pipeline/:pipelineid/reset-trigger-token", s.deps.PipelineProvider.PipelineResetToken)
		apiAuthGrp.POST("pipeline/:pipelineid/pull", s.deps.PipelineProvider.PipelinePull)
		apiAuthGrp.GET("pipeline/latest", s.deps.PipelineProvider.PipelineGetAllWithLatestRun)
		apiAuthGrp.POST("pipeline/periodicschedules", s.deps.PipelineProvider.PipelineCheckPeriodicSchedules)
		apiGrp.POST("pipeline/githook", s.deps.PipelineProvider.GitWebHook)
		apiGrp.POST("pipeline/:pipelineid/:pipelinetoken/trigger", s.deps.PipelineProvider.PipelineTrigger)

		// Settings
		settingsHandler := newSettingsHandler(s.deps.Store)
		apiAuthGrp.POST("settings/poll/on", s.deps.PipelineProvider.SettingsPollOn)
		apiAuthGrp.POST("settings/poll/off", s.deps.PipelineProvider.SettingsPollOff)
		apiAuthGrp.GET("settings/poll", s.deps.PipelineProvider.SettingsPollGet)
		apiAuthGrp.GET("settings/rbac", settingsHandler.rbacGet)
		apiAuthGrp.PUT("settings/rbac", settingsHandler.rbacPut)

		// PipelineRun
		apiAuthGrp.POST("pipelinerun/:pipelineid/:runid/stop", s.deps.PipelineProvider.PipelineStop)
		apiAuthGrp.GET("pipelinerun/:pipelineid/:runid", s.deps.PipelineProvider.PipelineRunGet)
		apiAuthGrp.GET("pipelinerun/:pipelineid", s.deps.PipelineProvider.PipelineGetAllRuns)
		apiAuthGrp.GET("pipelinerun/:pipelineid/latest", s.deps.PipelineProvider.PipelineGetLatestRun)
		apiAuthGrp.GET("pipelinerun/:pipelineid/:runid/log", s.deps.PipelineProvider.GetJobLogs)

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
	apiAuthGrp.GET("worker/secret", s.deps.WorkerProvider.GetWorkerRegisterSecret)
	apiAuthGrp.GET("worker/status", s.deps.WorkerProvider.GetWorkerStatusOverview)
	apiAuthGrp.GET("worker", s.deps.WorkerProvider.GetWorker)
	apiAuthGrp.DELETE("worker/:workerid", s.deps.WorkerProvider.DeregisterWorker)
	apiAuthGrp.POST("worker/secret", s.deps.WorkerProvider.ResetWorkerRegisterSecret)
	apiGrp.POST("worker/register", s.deps.WorkerProvider.RegisterWorker)

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
