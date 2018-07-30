package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/GeertJohan/go.rice"

	"crypto/rsa"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	// apiVersion represents the current API version
	apiVersion = "v1"
)

var (
	// errNotAuthorized is thrown when user wants to access resource which is protected
	errNotAuthorized = errors.New("no or invalid jwt token provided. You are not authorized")

	// errPathLength is a validation error during pipeline name input
	errPathLength = errors.New("name of pipeline is empty or one of the path elements length exceeds 50 characters")

	// errPipelineNotFound is thrown when a pipeline was not found with the given id
	errPipelineNotFound = errors.New("pipeline not found with the given id")

	// errInvalidPipelineID is thrown when the given pipeline id is not valid
	errInvalidPipelineID = errors.New("the given pipeline id is not valid")

	// errPipelineRunNotFound is thrown when a pipeline run was not found with the given id
	errPipelineRunNotFound = errors.New("pipeline run not found with the given id")

	// errLogNotFound is thrown when a job log file was not found
	errLogNotFound = errors.New("job log file not found")

	// errPipelineDelete is thrown when a pipeline binary could not be deleted
	errPipelineDelete = errors.New("pipeline could not be deleted. Perhaps you don't have the right permissions")

	// errPipelineRename is thrown when a pipeline binary could not be renamed
	errPipelineRename = errors.New("pipeline could not be renamed")
)

// InitHandlers initializes(registers) all handlers
func InitHandlers(e *echo.Echo) error {
	// Define prefix
	p := "/api/" + apiVersion + "/"

	// --- Register handlers at echo instance ---

	// Users
	e.POST(p+"login", UserLogin)
	e.GET(p+"users", UserGetAll)
	e.POST(p+"user/password", UserChangePassword)
	e.DELETE(p+"user/:username", UserDelete)
	e.POST(p+"user", UserAdd)

	// Pipelines
	e.POST(p+"pipeline", CreatePipeline)
	e.POST(p+"pipeline/gitlsremote", PipelineGitLSRemote)
	e.GET(p+"pipeline/created", CreatePipelineGetAll)
	e.GET(p+"pipeline/name", PipelineNameAvailable)
	e.GET(p+"pipeline", PipelineGetAll)
	e.GET(p+"pipeline/:pipelineid", PipelineGet)
	e.PUT(p+"pipeline/:pipelineid", PipelineUpdate)
	e.DELETE(p+"pipeline/:pipelineid", PipelineDelete)
	e.POST(p+"pipeline/:pipelineid/start", PipelineStart)
	e.GET(p+"pipeline/latest", PipelineGetAllWithLatestRun)

	// PipelineRun
	e.GET(p+"pipelinerun/:pipelineid/:runid", PipelineRunGet)
	e.GET(p+"pipelinerun/:pipelineid", PipelineGetAllRuns)
	e.GET(p+"pipelinerun/:pipelineid/latest", PipelineGetLatestRun)
	e.GET(p+"pipelinerun/:pipelineid/:runid/log", GetJobLogs)

	// Secrets
	e.GET(p+"secrets", ListSecrets)
	e.DELETE(p+"secret/:key", RemoveSecret)
	e.POST(p+"secret", SetSecret)
	e.PUT(p+"secret/update", SetSecret)

	// Middleware
	e.Use(middleware.Recover())
	//e.Use(middleware.Logger())
	e.Use(middleware.BodyLimit("32M"))
	e.Use(authBarrier)

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
		e.GET("/assets/css/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/assets/js/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/assets/fonts/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/assets/img/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
	}

	return nil
}

// authBarrier is the middleware which prevents user exploits.
// It makes sure that the request contains a valid jwt token.
// TODO: Role based access
func authBarrier(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Login and static resources are open
		if strings.Contains(c.Path(), "/login") || c.Path() == "/" || strings.Contains(c.Path(), "/assets/") || c.Path() == "/favicon.ico" {
			return next(c)
		}

		// Get JWT token
		jwtRaw := c.Request().Header.Get("Authorization")
		split := strings.Split(jwtRaw, " ")
		if len(split) != 2 {
			return c.String(http.StatusForbidden, errNotAuthorized.Error())
		}
		jwtString := split[1]

		// Parse token
		token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
			signingMethodError := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			switch token.Method.(type) {
			case *jwt.SigningMethodHMAC:
				if _, ok := gaia.Cfg.JWTKey.([]byte); !ok {
					return nil, signingMethodError
				}
				return gaia.Cfg.JWTKey, nil
			case *jwt.SigningMethodRSA:
				if _, ok := gaia.Cfg.JWTKey.(*rsa.PrivateKey); !ok {
					return nil, signingMethodError
				}
				return gaia.Cfg.JWTKey.(*rsa.PrivateKey).Public(), nil
			default:
				return nil, signingMethodError
			}
		})
		if err != nil {
			return c.String(http.StatusForbidden, err.Error())
		}

		// Validate token
		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// All ok, continue
			return next(c)
		}
		return c.String(http.StatusForbidden, errNotAuthorized.Error())
	}
}
