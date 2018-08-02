package handlers

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/labstack/echo"
)

const (
	maxMaxBufferLen = 1024
)

// jobLogs represents the json format which is returned
// by GetJobLogs.
type jobLogs struct {
	Log      string `json:"log"`
	Finished bool   `json:"finished"`
}

// PipelineRunGet returns details about a specific pipeline run.
// Required parameters are pipelineid and runid.
func PipelineRunGet(c echo.Context) error {
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

// PipelineGetAllRuns returns all runs about the given pipeline.
func PipelineGetAllRuns(c echo.Context) error {
	// Convert string to int because id is int
	storeService, _ := services.StorageService()
	pipelineID, err := strconv.Atoi(c.Param("pipelineid"))
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Get all runs by the given pipeline id
	runs, err := storeService.PipelineGetAllRuns(pipelineID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, runs)
}

// PipelineGetLatestRun returns the latest run of a pipeline, given by id.
func PipelineGetLatestRun(c echo.Context) error {
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
func GetJobLogs(c echo.Context) error {
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
	if run.Status == gaia.RunFailed || run.Status == gaia.RunSuccess {
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
