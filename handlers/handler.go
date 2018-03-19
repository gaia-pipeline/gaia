package handlers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	scheduler "github.com/gaia-pipeline/gaia/scheduler"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/kataras/iris"
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
func InitHandlers(i *iris.Application, store *store.Store, scheduler *scheduler.Scheduler) error {
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
	i.Post(p+"users/login", UserLogin)

	// Pipelines
	i.Post(p+"pipelines/gitlsremote", PipelineGitLSRemote)
	i.Post(p+"pipelines/create", CreatePipeline)
	i.Get(p+"pipelines/create", CreatePipelineGetAll)
	i.Post(p+"pipelines/name", PipelineNameAvailable)
	i.Get(p+"pipelines", PipelineGetAll)
	i.Get(p+"pipelines/detail/{id:string}", PipelineGet)
	i.Get(p+"pipelines/detail/{pipelineid:string}/{runid:string}", PipelineRunGet)
	i.Get(p+"pipelines/start/{id:string}", PipelineStart)
	i.Get(p+"pipelines/runs/{pipelineid:string}", PipelineGetAllRuns)

	// Jobs
	i.Get(p+"jobs/log{pipelineid:int}{pipelinerunid:int}{jobid:int}{start:int}{maxbufferlen:int}", GetJobLogs)

	// Authentication Barrier
	i.UseGlobal(authBarrier)

	return nil
}

// authBarrier is the middleware which prevents user exploits.
// It makes sure that the request contains a valid jwt token.
// TODO: Role based access
func authBarrier(ctx iris.Context) {
	// Login resource is open
	if strings.Contains(ctx.Path(), "users/login") {
		ctx.Next()
		return
	}

	// Get JWT token
	jwtRaw := ctx.GetHeader("Authorization")
	split := strings.Split(jwtRaw, " ")
	if len(split) != 2 {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.WriteString(errNotAuthorized.Error())
		return
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
		ctx.StatusCode(iris.StatusForbidden)
		ctx.WriteString(err.Error())
		return
	}

	// Validate token
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// All ok, continue
		ctx.Next()
	} else {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.WriteString(errNotAuthorized.Error())
	}
}
