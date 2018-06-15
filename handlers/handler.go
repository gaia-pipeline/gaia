package handlers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	scheduler "github.com/gaia-pipeline/gaia/scheduler"
	"github.com/gaia-pipeline/gaia/store"
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
)

// storeService is an instance of store.
// Use this to talk to the store.
var storeService *store.Store

var schedulerService *scheduler.Scheduler

// jwtKey is a random generated key for jwt signing
var jwtKey []byte

// InitHandlers initializes(registers) all handlers
func InitHandlers(e *echo.Echo, store *store.Store, scheduler *scheduler.Scheduler) error {
	// Set instances
	storeService = store
	schedulerService = scheduler

	// Generate signing key for jwt
	jwtKey = make([]byte, 64)
	_, err := rand.Read(jwtKey)
	if err != nil {
		return err
	}

	// Define prefix
	p := "/api/" + apiVersion + "/"

	// --- Register handlers at iris instance ---

	// Users
	e.POST(p+"login", UserLogin)
	e.GET(p+"users", UserGetAll)

	// Pipelines
	e.POST(p+"pipeline", CreatePipeline)
	e.POST(p+"pipeline/gitlsremote", PipelineGitLSRemote)
	e.GET(p+"pipeline/created", CreatePipelineGetAll)
	e.GET(p+"pipeline/name", PipelineNameAvailable)
	e.GET(p+"pipeline", PipelineGetAll)
	e.GET(p+"pipeline/:pipelineid", PipelineGet)
	e.POST(p+"pipeline/:pipelineid/start", PipelineStart)
	e.GET(p+"pipeline/latest", PipelineGetAllWithLatestRun)

	// PipelineRun
	e.GET(p+"pipelinerun/:pipelineid/:runid", PipelineRunGet)
	e.GET(p+"pipelinerun/:pipelineid", PipelineGetAllRuns)
	e.GET(p+"pipelinerun/:pipelineid/latest", PipelineGetLatestRun)
	e.GET(p+"pipelinerun/:pipelineid/:runid/log", GetJobLogs)

	// Middleware
	e.Use(middleware.Recover())
	//e.Use(middleware.Logger())
	e.Use(middleware.BodyLimit("32M"))
	e.Use(authBarrier)

	// Extra options
	e.HideBanner = true

	return nil
}

// authBarrier is the middleware which prevents user exploits.
// It makes sure that the request contains a valid jwt token.
// TODO: Role based access
func authBarrier(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Login resource is open
		if strings.Contains(c.Path(), "login") {
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
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// return secret
			return jwtKey, nil
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
