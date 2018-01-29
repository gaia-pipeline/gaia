package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/pipeline"
	"github.com/kataras/iris"
)

var (
	errPathLength = errors.New("name of pipeline is empty or one of the path elements length exceeds 50 characters")
)

const (
	// Split char to separate path from pipeline and name
	pipelinePathSplitChar = "/"

	// Percent of pipeline creation progress after git clone
	pipelineCloneStatus = 25

	// Percent of pipeline creation progress after compile process done
	pipelineCompileStatus = 75

	// Completed percent progress
	pipelineCompleteStatus = 100
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
	ctx.JSON(repo)
}

// PipelineBuildFromSource clones a given git repo and
// compiles the included source file to a pipeline.
func PipelineBuildFromSource(ctx iris.Context) {
	p := &gaia.Pipeline{}
	if err := ctx.ReadJSON(p); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// Set the creation date
	p.CreationDate = time.Now()

	// Save this pipeline to our store
	err := storeService.CreatePipelinePut(p)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		gaia.Cfg.Logger.Debug("cannot put pipeline into store", "error", err.Error())
		return
	}

	// Cloning the repo and compiling the pipeline will be done async
	go createPipelineExecute(p)
}

// createPipelineExecute clones the given git repo and compiles
// the pipeline. After every step, the status will be store.
// This method is designed to be called async.
func createPipelineExecute(p *gaia.Pipeline) {
	// Define build process for the given type
	bP := pipeline.NewBuildPipeline(p.Type)
	if bP == nil {
		// Pipeline type is not supported
		gaia.Cfg.Logger.Debug("create pipeline failed. Pipeline type is not supported", "type", p.Type)
		return
	}

	// Setup environment before cloning repo and command
	cmd, err := bP.PrepareBuild(p)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot prepare build", "error", err.Error())
		return
	}

	// Clone git repo
	err = pipeline.GitCloneRepo(&p.Repo)
	if err != nil {
		// Add error message and store
		p.ErrMsg = err.Error()
		storeService.CreatePipelinePut(p)
		gaia.Cfg.Logger.Debug("cannot clone repo", "error", err.Error())
		return
	}

	// Update status of our pipeline build
	p.Status = pipelineCloneStatus
	err = storeService.CreatePipelinePut(p)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot put create pipeline into store", "error", err.Error())
		return
	}

	// Run compile process
	err = bP.ExecuteBuild(cmd)
	if err != nil {
		// Add error message and store
		p.ErrMsg = err.Error()
		storeService.CreatePipelinePut(p)
		gaia.Cfg.Logger.Debug("cannot compile pipeline", "error", err.Error())
		return
	}

	// Update status of our pipeline build
	p.Status = pipelineCompileStatus
	err = storeService.CreatePipelinePut(p)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot put create pipeline into store", "error", err.Error())
		return
	}

	// Copy compiled binary to plugins folder
	err = bP.CopyBinary(p)
	if err != nil {
		// Add error message and store
		p.ErrMsg = err.Error()
		storeService.CreatePipelinePut(p)
		gaia.Cfg.Logger.Debug("cannot copy compiled binary", "error", err.Error())
		return
	}

	// Set create pipeline status to complete
	p.Status = pipelineCompleteStatus
	err = storeService.CreatePipelinePut(p)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot put create pipeline into store", "error", err.Error())
		return
	}
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
	p := &gaia.Pipeline{}
	if err := ctx.ReadJSON(p); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// The name could contain a path. Split it up
	path := strings.Split(p.Name, pipelinePathSplitChar)

	// Iterate all objects
	for _, p := range path {
		// Length should be correct
		if len(p) < 1 || len(p) > 50 {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.WriteString(errPathLength.Error())
			return
		}

		// TODO check if pipeline name is already in use
	}
}
