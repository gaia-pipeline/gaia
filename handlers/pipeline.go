package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/pipeline"
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
)

const (
	// Split char to separate path from pipeline and name
	pipelinePathSplitChar = "/"
)

// PipelineGitLSRemote checks for available git remote branches.
// This is the perfect way to check if we have access to a given repo.
func PipelineGitLSRemote(c echo.Context) error {
	repo := &gaia.GitRepo{}
	if err := c.Bind(repo); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Check for remote branches
	err := pipeline.GitLSRemote(repo)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}

	// Return branches
	return c.JSON(http.StatusOK, repo.Branches)
}

// CreatePipeline accepts all data needed to create a pipeline.
// It then starts the create pipeline execution process async.
func CreatePipeline(c echo.Context) error {
	p := &gaia.CreatePipeline{}
	if err := c.Bind(p); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Set initial value
	p.Created = time.Now()
	p.StatusType = gaia.CreatePipelineRunning
	p.ID = uuid.Must(uuid.NewV4(), nil).String()

	// Save this pipeline to our store
	err := storeService.CreatePipelinePut(p)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot put pipeline into store", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Cloning the repo and compiling the pipeline will be done async
	go pipeline.CreatePipeline(p)

	return nil
}

// CreatePipelineGetAll returns a json array of
// all pipelines which are about to get compiled and
// all pipelines which have been compiled.
func CreatePipelineGetAll(c echo.Context) error {
	// Get all create pipelines
	pipelineList, err := storeService.CreatePipelineGet()
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot get create pipelines from store", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Return all create pipelines
	return c.JSON(http.StatusOK, pipelineList)
}

// PipelineNameAvailable looks up if the given pipeline name is
// available and valid.
func PipelineNameAvailable(c echo.Context) error {
	pName := c.QueryParam("name")

	// The name could contain a path. Split it up
	path := strings.Split(pName, pipelinePathSplitChar)

	// Iterate all objects
	for _, s := range path {
		// Length should be correct
		if len(s) < 1 || len(s) > 50 {
			return c.String(http.StatusBadRequest, errPathLength.Error())
		}

		// TODO check if pipeline name is already in use
	}

	return nil
}

// PipelineGetAll returns all registered pipelines.
func PipelineGetAll(c echo.Context) error {
	var pipelines []gaia.Pipeline

	// Get all active pipelines
	for pipeline := range pipeline.GlobalActivePipelines.Iter() {
		pipelines = append(pipelines, pipeline)
	}

	// Return as json
	return c.JSON(http.StatusOK, pipelines)
}

// PipelineGet accepts a pipeline id and returns the pipeline object.
func PipelineGet(c echo.Context) error {
	pipelineIDStr := c.Param("pipelineid")

	// Convert string to int because id is int
	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Look up pipeline for the given id
	var foundPipeline gaia.Pipeline
	for pipeline := range pipeline.GlobalActivePipelines.Iter() {
		if pipeline.ID == pipelineID {
			foundPipeline = pipeline
		}
	}

	if foundPipeline.Name != "" {
		return c.JSON(http.StatusOK, foundPipeline)
	}

	// Pipeline not found
	return c.String(http.StatusNotFound, errPipelineNotFound.Error())
}

// PipelineUpdate updates the given pipeline.
func PipelineUpdate(c echo.Context) error {
	p := gaia.Pipeline{}
	if err := c.Bind(&p); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Look up pipeline for the given id
	var foundPipeline gaia.Pipeline
	for pipeline := range pipeline.GlobalActivePipelines.Iter() {
		if pipeline.ID == p.ID {
			foundPipeline = pipeline
		}
	}

	if foundPipeline.Name == "" {
		return c.String(http.StatusNotFound, errPipelineNotFound.Error())
	}

	// We're only handling pipeline name updates for now.
	if foundPipeline.Name != p.Name {
		// Pipeline name has been changed

		currentName := foundPipeline.Name

		// Rename binary
		err := pipeline.RenameBinary(foundPipeline, p.Name)
		if err != nil {
			return c.String(http.StatusInternalServerError, errPipelineRename.Error())
		}

		// Update name and exec path
		foundPipeline.Name = p.Name
		foundPipeline.ExecPath = pipeline.GetExecPath(p)

		// Update pipeline in store
		err = storeService.PipelinePut(&foundPipeline)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Update active pipelines
		pipeline.GlobalActivePipelines.ReplaceByName(currentName, foundPipeline)

	}

	return c.String(http.StatusOK, "Pipeline has been updated")
}

// PipelineDelete accepts a pipeline id and deletes it from the
// store. It also removes the binary inside the pipeline folder.
func PipelineDelete(c echo.Context) error {
	pipelineIDStr := c.Param("pipelineid")

	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Look up pipeline for the given id
	var foundPipeline gaia.Pipeline
	var index int
	var deletedPipelineIndex int
	for pipeline := range pipeline.GlobalActivePipelines.Iter() {
		if pipeline.ID == pipelineID {
			foundPipeline = pipeline
			deletedPipelineIndex = index
		}
		index++
	}

	if foundPipeline.Name == "" {
		return c.String(http.StatusNotFound, errPipelineNotFound.Error())
	}

	// Delete pipeline binary
	err = pipeline.DeleteBinary(foundPipeline)
	if err != nil {
		return c.String(http.StatusInternalServerError, errPipelineDelete.Error())
	}

	// Delete pipeline from store
	err = storeService.PipelineDelete(pipelineID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Remove from active pipelines
	pipeline.GlobalActivePipelines.Remove(deletedPipelineIndex)

	return c.String(http.StatusOK, "Pipeline has been deleted")
}

// PipelineStart starts a pipeline by the given id.
// Afterwards it returns the created/scheduled pipeline run.
func PipelineStart(c echo.Context) error {
	pipelineIDStr := c.Param("pipelineid")

	// Convert string to int because id is int
	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Look up pipeline for the given id
	var foundPipeline gaia.Pipeline
	for pipeline := range pipeline.GlobalActivePipelines.Iter() {
		if pipeline.ID == pipelineID {
			foundPipeline = pipeline
		}
	}

	if foundPipeline.Name != "" {
		pipelineRun, err := schedulerService.SchedulePipeline(&foundPipeline)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		} else if pipelineRun != nil {
			return c.JSON(http.StatusCreated, pipelineRun)
		}
	}

	// Pipeline not found
	return c.String(http.StatusNotFound, errPipelineNotFound.Error())
}

type getAllWithLatestRun struct {
	Pipeline    gaia.Pipeline    `json:"p"`
	PipelineRun gaia.PipelineRun `json:"r"`
}

// PipelineGetAllWithLatestRun returns the latest of all registered pipelines
// included with the latest run.
func PipelineGetAllWithLatestRun(c echo.Context) error {
	// Get all active pipelines
	var pipelines []gaia.Pipeline
	for pipeline := range pipeline.GlobalActivePipelines.Iter() {
		pipelines = append(pipelines, pipeline)
	}

	// Iterate all pipelines
	var pipelinesWithLatestRun []getAllWithLatestRun
	for _, pipeline := range pipelines {
		// Get the latest run by the given pipeline id
		run, err := storeService.PipelineGetLatestRun(pipeline.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Append run if one exists
		g := getAllWithLatestRun{}
		g.Pipeline = pipeline
		if run != nil {
			g.PipelineRun = *run
		}

		// Append
		pipelinesWithLatestRun = append(pipelinesWithLatestRun, g)
	}

	return c.JSON(http.StatusOK, pipelinesWithLatestRun)
}
