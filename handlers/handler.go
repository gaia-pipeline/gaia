package handlers

import (
	"net/http"
	"time"

	"github.com/gaia-pipeline/gaia/security/rbac"

	"github.com/gaia-pipeline/gaia/helper/cachehelper"

	"github.com/gaia-pipeline/gaia/helper/resourcehelper"
	"github.com/gaia-pipeline/gaia/services"

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
	storeService, _ := services.StorageService()

	service := rbac.NewService(storeService, cachehelper.NewCache(time.Minute*10))
	enforcer := rbac.NewPolicyEnforcer(service)
	policyEnforcer := newPolicyEnforcerMiddleware(enforcer)

	// Endpoints for Gaia primary instance
	if gaia.Cfg.Mode == gaia.ModeServer {
		// Users
		e.POST(p+"login", UserLogin)
		e.PUT(p+"user/:username/reset-trigger-token", UserResetTriggerToken)

		e.GET(p+"users", UserGetAll, policyEnforcer.do(rbac.UserNamespace, rbac.GetAction))
		e.POST(p+"user/password", UserChangePassword, policyEnforcer.do(rbac.UserNamespace, rbac.ChangePasswordAction))
		e.DELETE(p+"user/:username", UserDelete, policyEnforcer.do(rbac.UserNamespace, rbac.DeleteAction))
		e.POST(p+"user", UserAdd, policyEnforcer.do(rbac.UserNamespace, rbac.CreateAction))

		e.GET(p+"user/:username/permissions", UserGetPermissions, policyEnforcer.do(rbac.UserPermissionNamespace, rbac.GetAction))
		e.PUT(p+"user/:username/permissions", UserPutPermissions, policyEnforcer.do(rbac.UserPermissionNamespace, rbac.UpdateAction))

		perms := e.Group(p + "permission")
		perms.GET("", PermissionGetAll)

		// RBAC
		rbacHandler := newRBACHandler(storeService, resourcehelper.NewMarshaller())
		e.GET(p+"rbac/policy/:name", rbacHandler.AuthPolicyResourceGet)
		e.POST(p+"rbac/policy", rbacHandler.AuthPolicyResourcePut)
		e.PUT(p+"rbac/policy/:name/assign/:username", rbacHandler.AuthPolicyAssignmentPut)

		// Pipelines
		// Create pipeline provider
		pipelineProvider := pipelines.NewPipelineProvider(pipelines.Dependencies{
			Scheduler:       s.deps.Scheduler,
			PipelineService: s.deps.PipelineService,
		})

		e.POST(p+"pipeline", pipelineProvider.CreatePipeline, policyEnforcer.do(rbac.PipelineNamespace, rbac.CreateAction))
		e.POST(p+"pipeline/gitlsremote", pipelineProvider.PipelineGitLSRemote, policyEnforcer.do(rbac.PipelineNamespace, rbac.CreateAction))
		e.GET(p+"pipeline/name", pipelineProvider.PipelineNameAvailable, policyEnforcer.do(rbac.PipelineNamespace, rbac.CreateAction))
		e.POST(p+"pipeline/githook", GitWebHook, policyEnforcer.do(rbac.PipelineNamespace, rbac.CreateAction))
		e.GET(p+"pipeline/created", pipelineProvider.CreatePipelineGetAll, policyEnforcer.do(rbac.PipelineNamespace, rbac.ListAction))
		e.GET(p+"pipeline", pipelineProvider.PipelineGetAll, policyEnforcer.do(rbac.PipelineNamespace, rbac.ListAction))
		e.GET(p+"pipeline/latest", pipelineProvider.PipelineGetAllWithLatestRun, policyEnforcer.do(rbac.PipelineNamespace, rbac.ListAction))
		e.GET(p+"pipeline/:pipelineid", pipelineProvider.PipelineGet, policyEnforcer.do(rbac.PipelineNamespace, rbac.GetAction))
		e.PUT(p+"pipeline/:pipelineid", pipelineProvider.PipelineUpdate, policyEnforcer.do(rbac.PipelineNamespace, rbac.UpdateAction))
		e.DELETE(p+"pipeline/:pipelineid", pipelineProvider.PipelineDelete, policyEnforcer.do(rbac.PipelineNamespace, rbac.DeleteAction))
		e.POST(p+"pipeline/:pipelineid/start", pipelineProvider.PipelineStart, policyEnforcer.do(rbac.PipelineNamespace, rbac.StartAction))

		e.POST(p+"pipeline/:pipelineid/:pipelinetoken/trigger", pipelineProvider.PipelineTrigger)
		e.PUT(p+"pipeline/:pipelineid/reset-trigger-token", pipelineProvider.PipelineResetToken)
		e.POST(p+"pipeline/periodicschedules", pipelineProvider.PipelineCheckPeriodicSchedules)

		// Settings
		e.POST(p+"settings/poll/on", SettingsPollOn)
		e.POST(p+"settings/poll/off", SettingsPollOff)
		e.GET(p+"settings/poll", SettingsPollGet)

		// PipelineRun
		e.POST(p+"pipelinerun/:pipelineid/:runid/stop", pipelineProvider.PipelineStop, policyEnforcer.do(rbac.PipelineRunNamespace, rbac.StopAction))
		e.GET(p+"pipelinerun/:pipelineid/:runid", pipelineProvider.PipelineRunGet, policyEnforcer.do(rbac.PipelineRunNamespace, rbac.GetAction))
		e.GET(p+"pipelinerun/:pipelineid/latest", pipelineProvider.PipelineGetLatestRun, policyEnforcer.do(rbac.PipelineRunNamespace, rbac.GetAction))
		e.GET(p+"pipelinerun/:pipelineid", pipelineProvider.PipelineGetAllRuns, policyEnforcer.do(rbac.PipelineRunNamespace, rbac.ListAction))
		e.GET(p+"pipelinerun/:pipelineid/:runid/log", pipelineProvider.GetJobLogs, policyEnforcer.do(rbac.PipelineRunNamespace, rbac.LogsAction))

		// Secrets
		e.GET(p+"secrets", ListSecrets, policyEnforcer.do(rbac.SecretNamespace, rbac.ListAction))
		e.DELETE(p+"secret/:key", RemoveSecret, policyEnforcer.do(rbac.SecretNamespace, rbac.DeleteAction))
		e.POST(p+"secret", SetSecret, policyEnforcer.do(rbac.SecretNamespace, rbac.CreateAction))
		e.PUT(p+"secret/update", SetSecret, policyEnforcer.do(rbac.SecretNamespace, rbac.UpdateAction))
	}

	// Worker
	// initialize the worker provider
	workerProvider := workers.NewWorkerProvider(workers.Dependencies{
		Scheduler:   s.deps.Scheduler,
		Certificate: s.deps.Certificate,
	})

	e.POST(p+"worker/register", workerProvider.RegisterWorker)
	e.GET(p+"worker/secret", workerProvider.GetWorkerRegisterSecret, policyEnforcer.do(rbac.WorkerNamespace, rbac.GetRegistrationSecretAction))
	e.POST(p+"worker/secret", workerProvider.ResetWorkerRegisterSecret, policyEnforcer.do(rbac.WorkerNamespace, rbac.ResetWorkerRegisterSecretAction))
	e.GET(p+"worker/status", workerProvider.GetWorkerStatusOverview, policyEnforcer.do(rbac.WorkerNamespace, rbac.GetOverviewAction))
	e.GET(p+"worker", workerProvider.GetWorker, policyEnforcer.do(rbac.WorkerNamespace, rbac.GetWorkerAction))
	e.DELETE(p+"worker/:workerid", workerProvider.DeregisterWorker, policyEnforcer.do(rbac.WorkerNamespace, rbac.DeregisterWorkerAction))

	// Middleware
	e.Use(middleware.Recover())
	// e.Use(middleware.Logger())
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
