package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
)

const (
	maxMaxBufferLen = 1024
)

// jobLogs represents the json format which is returned
// by GetJobLogs.
type jobLogs struct {
	Log      string `json:"log"`
	StartPos int    `json:"start"`
}

// PipelineRunGet returns details about a specific pipeline run.
// Required parameters are pipelineid and runid.
func PipelineRunGet(c echo.Context) error {
	// Convert string to int because id is int
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

// GetJobLogs returns logs and new start position for the given job.
//
// Required parameters:
// pipelineid - Related pipeline id
// pipelinerunid - Related pipeline run id
// jobid - Job id
// start - Start position to read from. If zero starts from the beginning
// maxbufferlen - Maximal returned characters
func GetJobLogs(c echo.Context) error {
	// Get parameters and validate
	pipelineID := c.Param("pipelineid")
	pipelineRunID := c.Param("runid")
	jobID := c.QueryParam("jobid")
	startPosStr := c.QueryParam("start")
	maxBufferLenStr := c.QueryParam("maxbufferlen")

	// Transform start pos to int
	startPos, err := strconv.Atoi(startPosStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid start position given")
	}

	// Transform max buffer len
	maxBufferLen, err := strconv.Atoi(maxBufferLenStr)
	if err != nil || maxBufferLen > maxMaxBufferLen || maxBufferLen < 0 {
		return c.String(http.StatusBadRequest, fmt.Sprintf("invalid maxbufferlen provided. Max number is %d", maxMaxBufferLen))
	}

	// Lookup log file
	logFilePath := filepath.Join(gaia.Cfg.WorkspacePath, pipelineID, pipelineRunID, gaia.LogsFolderName, jobID)
	gaia.Cfg.Logger.Debug("logfilepath", "Path", logFilePath)
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		return c.String(http.StatusNotFound, errLogNotFound.Error())
	}

	// Open file
	file, err := os.Open(logFilePath)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer file.Close()

	// Read file
	buf := make([]byte, maxBufferLen)
	bytesRead, err := file.ReadAt(buf, int64(startPos))
	if err != io.EOF && err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Create return struct
	j := jobLogs{
		Log:      string(buf[:]),
		StartPos: startPos + bytesRead,
	}

	// Return logs
	return c.JSON(http.StatusOK, j)
}
