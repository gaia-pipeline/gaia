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

// GetJobLogs returns logs for the given job with paging option.
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
	startPos, err := ctx.Params().GetInt64("start")
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
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		ctx.StatusCode(iris.StatusInternalServerError)
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
	_, err = file.ReadAt(buf, startPos)
	if err != io.EOF && err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return
	}

	// Return logs
	ctx.WriteString(string(buf[:]))
}
