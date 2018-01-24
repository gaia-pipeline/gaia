package golang

import "github.com/gaia-pipeline/gaia/proto"

// Jobs new type for wrapper around proto.job
type Jobs []JobsWrapper

// JobsWrapper wraps a function pointer around the
// proto.Job struct.
// The given function corresponds to the job.
type JobsWrapper struct {
	FuncPointer func() error
	Job         proto.Job
}

// Get looks up a job by the given id.
// Returns the job otherwise nil.
func (j *Jobs) Get(uniqueid string) *JobsWrapper {
	for _, job := range *j {
		if job.Job.UniqueId == uniqueid {
			return &job
		}
	}

	return nil
}
