package providers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	"github.com/labstack/echo"
)

var (
	// errPipelineRunNotFound is thrown when a pipeline run was not found with the given id
	errPipelineRunNotFound = errors.New("pipeline run not found with the given id")
)

// jobLogs represents the json format which is returned
// by GetJobLogs.
type jobLogs struct {
	Log      string `json:"log"`
	Finished bool   `json:"finished"`
}

// PipelineRunGet returns details about a specific pipeline run.
// Required parameters are pipelineid and runid.
func (pp *pipelineProvider) PipelineRunGet(c echo.Context) error {
	// Convert string to int because id is int
	storeService, _ := services.StorageService()
	pipelineID, err := strconv.Atoi(c.Param("pipelineid"))
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Convert string to int because id is int
	runID, err := strconv.Atoi(c.Param("runid"))
	if err != nil {
		return c.String(http.StatusBadRequest, errPipelineRunNotFound.Error())
	}

	// Find pipeline run in store
	pipelineRun, err := storeService.PipelineGetRunByPipelineIDAndID(pipelineID, runID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	} else if pipelineRun == nil {
		return c.String(http.StatusNotFound, errPipelineRunNotFound.Error())
	}

	// Return pipeline run
	return c.JSON(http.StatusOK, pipelineRun)
}

// PipelineStop stops a running pipeline.
func (pp *pipelineProvider) PipelineStop(c echo.Context) error {
	// Get parameters and validate
	pipelineID := c.Param("pipelineid")
	pipelineRunID := c.Param("runid")

	// Transform pipelineid to int
	p, err := strconv.Atoi(pipelineID)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid pipeline id given")
	}

	// Transform pipelinerunid to int
	r, err := strconv.Atoi(pipelineRunID)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid pipeline run id given")
	}

	// Look up pipeline for the given id
	var foundPipeline gaia.Pipeline
	for _, pipe := range pipeline.GlobalActivePipelines.GetAll() {
		if pipe.ID == p {
			foundPipeline = pipe
			break
		}
	}

	if foundPipeline.Name != "" {
		err = pp.deps.Scheduler.StopPipelineRun(&foundPipeline, r)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.String(http.StatusOK, "pipeline successfully stopped")
	}

	// Pipeline not found
	return c.String(http.StatusNotFound, errPipelineNotFound.Error())
}

// PipelineGetAllRuns returns all runs about the given pipeline.
func (pp *pipelineProvider) PipelineGetAllRuns(c echo.Context) error {
	// Convert string to int because id is int
	storeService, _ := services.StorageService()
	pipelineID, err := strconv.Atoi(c.Param("pipelineid"))
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Get all runs by the given pipeline id
	runs, err := storeService.PipelineGetAllRunsByPipelineID(pipelineID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, runs)
}

// PipelineGetLatestRun returns the latest run of a pipeline, given by id.
func (pp *pipelineProvider) PipelineGetLatestRun(c echo.Context) error {
	// Convert string to int because id is int
	storeService, _ := services.StorageService()
	pipelineID, err := strconv.Atoi(c.Param("pipelineid"))
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Get the latest run by the given pipeline id
	run, err := storeService.PipelineGetLatestRun(pipelineID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, run)
}

// GetJobLogs returns logs from a pipeline run.
//
// Required parameters:
// pipelineid - Related pipeline id
// pipelinerunid - Related pipeline run id
func (pp *pipelineProvider) GetJobLogs(c echo.Context) error {
	// Get parameters and validate
	storeService, _ := services.StorageService()
	pipelineID := c.Param("pipelineid")
	pipelineRunID := c.Param("runid")

	// Transform pipelineid to int
	p, err := strconv.Atoi(pipelineID)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid pipeline id given")
	}

	// Transform pipelinerunid to int
	r, err := strconv.Atoi(pipelineRunID)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid pipeline run id given")
	}

	run, err := storeService.PipelineGetRunByPipelineIDAndID(p, r)
	if err != nil {
		return c.String(http.StatusBadRequest, "cannot find pipeline run with given pipeline id and pipeline run id")
	}

	// Create return object
	jL := jobLogs{}

	// Determine if job has been finished
	if run.Status == gaia.RunFailed || run.Status == gaia.RunSuccess || run.Status == gaia.RunCancelled {
		jL.Finished = true
	}

	// Check if log file exists
	logFilePath := filepath.Join(gaia.Cfg.WorkspacePath, pipelineID, pipelineRunID, gaia.LogsFolderName, gaia.LogsFileName)
	if _, err := os.Stat(logFilePath); err == nil {
		content, err := ioutil.ReadFile(logFilePath)
		if err != nil {
			return c.String(http.StatusInternalServerError, "cannot read pipeline run log file")
		}

		// Convert logs
		jL.Log = string(content)
	}

	// Return logs
	return c.JSON(http.StatusOK, jL)
}
