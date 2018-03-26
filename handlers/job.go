package handlers

import (
	"io"
	"os"
	"path/filepath"

	"github.com/gaia-pipeline/gaia"
	"github.com/kataras/iris"
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

// GetJobLogs returns logs and new start position for the given job.
//
// Required parameters:
// pipelineid - Related pipeline id
// pipelinerunid - Related pipeline run id
// jobid - Job id
// start - Start position to read from. If zero starts from the beginning
// maxbufferlen - Maximal returned characters
func GetJobLogs(ctx iris.Context) {
	// Get parameters and validate
	pipelineID := ctx.Params().Get("pipelineid")
	pipelineRunID := ctx.Params().Get("pipelinerunid")
	jobID := ctx.Params().Get("jobid")
	startPos, err := ctx.Params().GetInt("start")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("invalid start position given")
		return
	}
	maxBufferLen, err := ctx.Params().GetInt("maxbufferlen")
	if err != nil || maxBufferLen > maxMaxBufferLen || maxBufferLen < 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.Writef("invalid maxbufferlen provided. Max number is %d", maxMaxBufferLen)
		return
	}

	// Lookup log file
	logFilePath := filepath.Join(gaia.Cfg.WorkspacePath, pipelineID, pipelineRunID, gaia.LogsFolderName, jobID)
	gaia.Cfg.Logger.Debug("logfilepath", "Path", logFilePath)
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.WriteString(errLogNotFound.Error())
		return
	}

	// Open file
	file, err := os.Open(logFilePath)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return
	}
	defer file.Close()

	// Read file
	buf := make([]byte, maxBufferLen)
	bytesRead, err := file.ReadAt(buf, int64(startPos))
	if err != io.EOF && err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return
	}

	// Create return struct
	j := jobLogs{
		Log:      string(buf[:]),
		StartPos: startPos + bytesRead,
	}

	// Return logs
	ctx.JSON(j)
}
