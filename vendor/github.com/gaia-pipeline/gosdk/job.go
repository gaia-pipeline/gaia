package golang

import (
	"github.com/gaia-pipeline/protobuf"
)

// Jobs is a collection of job
type Jobs []Job

// Job represents a single job which should be executed during pipeline run.
// Handler is the function pointer to the function which will be executed.
type Job struct {
	Handler     func() error
	Title       string
	Description string
	Priority    int64
	Args        map[string]string
}

// jobsWrapper wraps a function pointer around the
// proto.Job struct.
// The given function corresponds to the job.
type jobsWrapper struct {
	funcPointer func() error
	job         proto.Job
}

// Get looks up a job by the given id.
// Returns the job otherwise nil.
func getJob(hash uint32) *jobsWrapper {
	for _, job := range cachedJobs {
		if job.job.UniqueId == hash {
			return &job
		}
	}

	return nil
}
