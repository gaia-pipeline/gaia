package service

import "github.com/gaia-pipeline/gaia"

// GaiaScheduler is a job scheduler for gaia pipeline runs.
type GaiaScheduler interface {
	Init()
	SchedulePipeline(p *gaia.Pipeline, args []*gaia.Argument) (*gaia.PipelineRun, error)
	SetPipelineJobs(p *gaia.Pipeline) error
	StopPipelineRun(p *gaia.Pipeline, runID int) error
	GetFreeWorkers() int32
	CountScheduledRuns() int
}
