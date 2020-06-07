package handlers

import (
	"log"
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/casbin/casbin/v2"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/handlers/providers/pipelines"
	"github.com/gaia-pipeline/gaia/handlers/providers/workers"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/gaia-pipeline/gaia/security/rbac"
	"github.com/gaia-pipeline/gaia/services"
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

	store, err := services.StorageService()
	if err != nil {
		log.Fatal(err)
	}

	enforcer, err := casbin.NewEnforcer("security/rbac/rbac-model.conf", store.CasbinStore())
	if err != nil {
		log.Fatal(err)
	}
	enforcer.EnableLog(true)

	svc, err := rbac.NewEnforcerSvc(enforcer, "security/rbac/rbac-api-mappings.yml")
	if err != nil {
		log.Fatal(err)
	}

	// Standard API router group.
	apiGrp := e.Group(p)

	// Auth API router group.
	apiAuthGrp := e.Group(p, AuthMiddleware(&AuthConfig{
		RoleCategories: rolehelper.DefaultUserRoles,
		rbacEnforcer:   svc,
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
		apiAuthGrp.GET("pipeline/latest", pipelineProvider.PipelineGetAllWithLatestRun)
		apiAuthGrp.POST("pipeline/periodicschedules", pipelineProvider.PipelineCheckPeriodicSchedules)
		apiGrp.POST("pipeline/githook", GitWebHook)
		apiGrp.POST("pipeline/:pipelineid/:pipelinetoken/trigger", pipelineProvider.PipelineTrigger)

		// Settings
		apiAuthGrp.POST("settings/poll/on", SettingsPollOn)
		apiAuthGrp.POST("settings/poll/off", SettingsPollOff)
		apiAuthGrp.GET("settings/poll", SettingsPollGet)

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
