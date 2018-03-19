package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/pipeline"
	"github.com/kataras/iris"
	uuid "github.com/satori/go.uuid"
)

const (
	// Split char to separate path from pipeline and name
	pipelinePathSplitChar = "/"
)

// PipelineGitLSRemote checks for available git remote branches.
// This is the perfect way to check if we have access to a given repo.
func PipelineGitLSRemote(ctx iris.Context) {
	repo := &gaia.GitRepo{}
	if err := ctx.ReadJSON(repo); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// Check for remote branches
	err := pipeline.GitLSRemote(repo)
	if err != nil {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.WriteString(err.Error())
		return
	}

	// Return branches
	ctx.JSON(repo.Branches)
}

// CreatePipeline accepts all data needed to create a pipeline.
// It then starts the create pipeline execution process async.
func CreatePipeline(ctx iris.Context) {
	p := &gaia.CreatePipeline{}
	if err := ctx.ReadJSON(p); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// Set the creation date and unique id
	p.Created = time.Now()
	p.ID = uuid.Must(uuid.NewV4()).String()

	// Save this pipeline to our store
	err := storeService.CreatePipelinePut(p)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		gaia.Cfg.Logger.Debug("cannot put pipeline into store", "error", err.Error())
		return
	}

	// Cloning the repo and compiling the pipeline will be done async
	go pipeline.CreatePipeline(p)
}

// CreatePipelineGetAll returns a json array of
// all pipelines which are about to get compiled and
// all pipelines which have been compiled.
func CreatePipelineGetAll(ctx iris.Context) {
	// Get all create pipelines
	pipelineList, err := storeService.CreatePipelineGet()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		gaia.Cfg.Logger.Debug("cannot get create pipelines from store", "error", err.Error())
		return
	}

	// Return all create pipelines
	ctx.JSON(pipelineList)
}

// PipelineNameAvailable looks up if the given pipeline name is
// available and valid.
func PipelineNameAvailable(ctx iris.Context) {
	p := &gaia.CreatePipeline{}
	if err := ctx.ReadJSON(p); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// The name could contain a path. Split it up
	path := strings.Split(p.Pipeline.Name, pipelinePathSplitChar)

	// Iterate all objects
	for _, s := range path {
		// Length should be correct
		if len(s) < 1 || len(s) > 50 {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.WriteString(errPathLength.Error())
			return
		}

		// TODO check if pipeline name is already in use
	}
}

// PipelineGetAll returns all registered pipelines.
func PipelineGetAll(ctx iris.Context) {
	var pipelines []gaia.Pipeline

	// Get all active pipelines
	for pipeline := range pipeline.GlobalActivePipelines.Iter() {
		pipelines = append(pipelines, pipeline)
	}

	// Return as json
	ctx.JSON(pipelines)
}

// PipelineGet accepts a pipeline id and returns the pipeline object.
func PipelineGet(ctx iris.Context) {
	pipelineIDStr := ctx.Params().Get("id")

	// Convert string to int because id is int
	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(errInvalidPipelineID.Error())
		return
	}

	// Look up pipeline for the given id
	for pipeline := range pipeline.GlobalActivePipelines.Iter() {
		if pipeline.ID == pipelineID {
			ctx.JSON(pipeline)
			return
		}
	}

	// Pipeline not found
	ctx.StatusCode(iris.StatusNotFound)
	ctx.WriteString(errPipelineNotFound.Error())
}

// PipelineStart starts a pipeline by the given id.
// Afterwards it returns the created/scheduled pipeline run.
func PipelineStart(ctx iris.Context) {
	pipelineIDStr := ctx.Params().Get("id")

	// Convert string to int because id is int
	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(errInvalidPipelineID.Error())
		return
	}

	// Look up pipeline for the given id
	for pipeline := range pipeline.GlobalActivePipelines.Iter() {
		if pipeline.ID == pipelineID {
			pipelineRun, err := schedulerService.SchedulePipeline(&pipeline)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				ctx.WriteString(err.Error())
				return
			} else if pipelineRun != nil {
				ctx.StatusCode(iris.StatusCreated)
				ctx.JSON(pipelineRun)
				return
			}
		}
	}

	// Pipeline not found
	ctx.StatusCode(iris.StatusNotFound)
	ctx.WriteString(errPipelineNotFound.Error())
}

// PipelineRunGet returns details about a specific pipeline run.
// Required parameters are pipelineid and runid.
func PipelineRunGet(ctx iris.Context) {
	// Convert string to int because id is int
	pipelineID, err := strconv.Atoi(ctx.Params().Get("pipelineid"))
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(errInvalidPipelineID.Error())
		return
	}

	// Convert string to int because id is int
	runID, err := strconv.Atoi(ctx.Params().Get("runid"))
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(errPipelineRunNotFound.Error())
		return
	}

	// Find pipeline run in store
	pipelineRun, err := storeService.PipelineGetRunByPipelineIDAndID(pipelineID, runID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return
	} else if pipelineRun == nil {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.WriteString(errPipelineRunNotFound.Error())
		return
	}

	// Return pipeline run
	ctx.JSON(pipelineRun)
}

// PipelineGetAllRuns returns all runs about the given pipeline.
func PipelineGetAllRuns(ctx iris.Context) {
	// Convert string to int because id is int
	pipelineID, err := strconv.Atoi(ctx.Params().Get("pipelineid"))
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(errInvalidPipelineID.Error())
		return
	}

	// Get all runs by the given pipeline id
	runs, err := storeService.PipelineGetAllRuns(pipelineID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return
	}

	ctx.JSON(runs)
}
