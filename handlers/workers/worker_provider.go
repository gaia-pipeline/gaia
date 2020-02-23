package workers

import "github.com/gaia-pipeline/gaia/workers/scheduler/service"

type Dependencies struct {
	Scheduler service.GaiaScheduler
}

type workerProvider struct {
	deps Dependencies
}

func NewWorkerProvider(deps Dependencies) *workerProvider {
	return &workerProvider{
		deps: deps,
	}
}
