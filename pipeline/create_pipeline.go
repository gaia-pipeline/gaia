package pipeline

import (
	"fmt"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
)

const (
	// Percent of pipeline creation progress after git clone
	pipelineCloneStatus = 25

	// Percent of pipeline creation progress after compile process done
	pipelineCompileStatus = 75

	// Completed percent progress
	pipelineCompleteStatus = 100
)

// CreatePipeline is the main function which executes step by step the creation
// of a plugin.
// After each step, the status is written to store and can be retrieved via API.
func CreatePipeline(p *gaia.CreatePipeline) {
	storeService, _ := services.StorageService()
	// Define build process for the given type
	bP := newBuildPipeline(p.Pipeline.Type)
	if bP == nil {
		// Pipeline type is not supported
		p.StatusType = gaia.CreatePipelineFailed
		p.Output = fmt.Sprintf("create pipeline failed. Pipeline type is not supported %s is not supported", p.Pipeline.Type)
		storeService.CreatePipelinePut(p)
		return
	}

	// Setup environment before cloning repo and command
	err := bP.PrepareEnvironment(p)
	if err != nil {
		p.StatusType = gaia.CreatePipelineFailed
		p.Output = fmt.Sprintf("cannot prepare build: %s", err.Error())
		storeService.CreatePipelinePut(p)
		return
	}

	// Clone git repo
	err = gitCloneRepo(&p.Pipeline.Repo)
	if err != nil {
		p.StatusType = gaia.CreatePipelineFailed
		p.Output = fmt.Sprintf("cannot prepare build: %s", err.Error())
		storeService.CreatePipelinePut(p)
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
	err = bP.ExecuteBuild(p)
	if err != nil {
		p.StatusType = gaia.CreatePipelineFailed
		storeService.CreatePipelinePut(p)
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
		p.StatusType = gaia.CreatePipelineFailed
		p.Output = fmt.Sprintf("cannot copy compiled binary: %s", err.Error())
		storeService.CreatePipelinePut(p)
		return
	}

	// Save the generated pipeline data
	err = bP.SavePipeline(&p.Pipeline)
	if err != nil {
		p.StatusType = gaia.CreatePipelineFailed
		p.Output = fmt.Sprintf("failed to save the created pipeline: %s", err.Error())
		storeService.CreatePipelinePut(p)
		return
	}

	// Set create pipeline status to complete
	p.Status = pipelineCompleteStatus
	p.StatusType = gaia.CreatePipelineSuccess
	err = storeService.CreatePipelinePut(p)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot put create pipeline into store", "error", err.Error())
		return
	}
}
