package handlers

import (
	"errors"
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/rolehelper"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	// errPipelineNotFound is thrown when a pipeline was not found with the given id
	errPipelineNotFound = errors.New("pipeline not found with the given id")

	// errInvalidPipelineID is thrown when the given pipeline id is not valid
	errInvalidPipelineID = errors.New("the given pipeline id is not valid")

	// errPipelineRunNotFound is thrown when a pipeline run was not found with the given id
	errPipelineRunNotFound = errors.New("pipeline run not found with the given id")

	// errPipelineDelete is thrown when a pipeline binary could not be deleted
	errPipelineDelete = errors.New("pipeline could not be deleted. Perhaps you don't have the right permissions")

	// errPipelineRename is thrown when a pipeline binary could not be renamed
	errPipelineRename = errors.New("pipeline could not be renamed")

	// List of secret keys which cannot be modified via the normal Vault API.
	ignoredVaultKeys []string
)

// InitHandlers initializes(registers) all handlers.
func InitHandlers(e *echo.Echo) error {
	// Define prefix
	p := "/api/" + gaia.APIVersion + "/"

	// --- Register handlers at echo instance ---

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
	e.POST(p+"pipeline", CreatePipeline)
	e.POST(p+"pipeline/gitlsremote", PipelineGitLSRemote)
	e.GET(p+"pipeline/name", PipelineNameAvailable)
	e.POST(p+"pipeline/githook", GitWebHook)
	e.GET(p+"pipeline/created", CreatePipelineGetAll)
	e.GET(p+"pipeline", PipelineGetAll)
	e.GET(p+"pipeline/:pipelineid", PipelineGet)
	e.PUT(p+"pipeline/:pipelineid", PipelineUpdate)
	e.DELETE(p+"pipeline/:pipelineid", PipelineDelete)
	e.POST(p+"pipeline/:pipelineid/start", PipelineStart)
	e.POST(p+"pipeline/:pipelineid/:pipelinetoken/trigger", PipelineTrigger)
	e.PUT(p+"pipeline/:pipelineid/reset-trigger-token", PipelineResetToken)
	e.GET(p+"pipeline/latest", PipelineGetAllWithLatestRun)
	e.POST(p+"pipeline/periodicschedules", PipelineCheckPeriodicSchedules)

	// Settings
	e.POST(p+"settings/poll/on", SettingsPollOn)
	e.POST(p+"settings/poll/off", SettingsPollOff)
	e.GET(p+"settings/poll", SettingsPollGet)

	// PipelineRun
	e.POST(p+"pipelinerun/:pipelineid/:runid/stop", PipelineStop)
	e.GET(p+"pipelinerun/:pipelineid/:runid", PipelineRunGet)
	e.GET(p+"pipelinerun/:pipelineid", PipelineGetAllRuns)
	e.GET(p+"pipelinerun/:pipelineid/latest", PipelineGetLatestRun)
	e.GET(p+"pipelinerun/:pipelineid/:runid/log", GetJobLogs)

	// Secrets
	e.GET(p+"secrets", ListSecrets)
	e.DELETE(p+"secret/:key", RemoveSecret)
	e.POST(p+"secret", SetSecret)
	e.PUT(p+"secret/update", SetSecret)

	// Worker
	e.GET(p+"worker/secret", GetWorkerRegisterSecret)
	e.POST(p+"worker/register", RegisterWorker)
	e.GET(p+"worker/status", GetWorkerStatusOverview)
	e.GET(p+"worker", GetWorker)
	e.DELETE(p+"worker/:workerid", DeregisterWorker)
	e.POST(p+"worker/secret", ResetWorkerRegisterSecret)
	e.GET(p+"worker/pipeline-repo/:name", GetPipelineRepositoryInformation)

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
