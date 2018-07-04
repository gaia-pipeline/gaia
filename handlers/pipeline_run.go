package handlers

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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
	Finished bool   `json:"finished"`
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

// PipelineGetLatestRun returns the latest run of a pipeline, given by id.
func PipelineGetLatestRun(c echo.Context) error {
	// Convert string to int because id is int
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

// GetJobLogs returns jobs for a given job.
// If no jobID is given, a collection of all jobs logs will be returned.
//
// Required parameters:
// pipelineid - Related pipeline id
// pipelinerunid - Related pipeline run id
//
// Optional parameters:
// jobid - Job id
func GetJobLogs(c echo.Context) error {
	// Get parameters and validate
	pipelineID := c.Param("pipelineid")
	pipelineRunID := c.Param("runid")
	jobID := c.QueryParam("jobid")

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

	// Get pipeline run from store
	run, err := storeService.PipelineGetRunByPipelineIDAndID(p, r)
	if err != nil {
		return c.String(http.StatusBadRequest, "cannot find pipeline run with given pipeline id and pipeline run id")
	}

	// jobID is not empty, just return the logs from this job
	if jobID != "" {
		for _, job := range run.Jobs {
			if strconv.FormatUint(uint64(job.ID), 10) == jobID {
				// Get logs
				jL, err := getLogs(pipelineID, pipelineRunID, jobID, false)
				if err != nil {
					return c.String(http.StatusBadRequest, err.Error())
				}

				// Check if job is finished
				if job.Status == gaia.JobSuccess || job.Status == gaia.JobFailed {
					jL.Finished = true
				}

				// We always return an array.
				// It makes a bit easier in the frontend.
				jobLogsList := []jobLogs{}
				jobLogsList = append(jobLogsList, *jL)
				return c.JSON(http.StatusOK, jobLogsList)
			}
		}

		// Logs for given job id not found
		return c.String(http.StatusBadRequest, "cannot find job with given job id")
	}

	// Sort the slice. This is important for the order of the returned logs.
	sort.Slice(run.Jobs, func(i, j int) bool {
		return run.Jobs[i].Priority < run.Jobs[j].Priority
	})

	// Return a collection of all logs
	jobs := []jobLogs{}
	for _, job := range run.Jobs {
		// Get logs
		jL, err := getLogs(pipelineID, pipelineRunID, strconv.FormatUint(uint64(job.ID), 10), true)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		// No error but also no job logs. Job must be in the queue.
		// We skip it so no error will break things.
		if jL == nil {
			continue
		}

		// Check if job is finished
		if job.Status == gaia.JobSuccess || job.Status == gaia.JobFailed {
			jL.Finished = true
		}

		jobs = append(jobs, *jL)
	}

	// Return logs
	return c.JSON(http.StatusOK, jobs)
}

func getLogs(pipelineID, pipelineRunID, jobID string, getAllJobLogs bool) (*jobLogs, error) {
	// Lookup log file
	logFilePath := filepath.Join(gaia.Cfg.WorkspacePath, pipelineID, pipelineRunID, gaia.LogsFolderName, jobID)

	// We only check if logs exist when a specific job log was requested.
	// If we don't do this, get all job logs will fail during a pipeline run.
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		if !getAllJobLogs {
			return nil, err
		}
		return nil, nil
	}

	// Read file
	content, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		return nil, err
	}

	// Create return struct
	return &jobLogs{
		Log: string(content),
	}, nil
}
