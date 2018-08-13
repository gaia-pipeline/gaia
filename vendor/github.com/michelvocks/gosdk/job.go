package golang

import (
	"github.com/michelvocks/protobuf"
)

// InputType represents the available input types.
type InputType string

const (
	// TextFieldInp text field input
	TextFieldInp InputType = "textfield"

	// TextAreaInp text area input
	TextAreaInp InputType = "textarea"

	// BoolInp boolean input
	BoolInp InputType = "boolean"

	// VaultInp vault automatic input
	VaultInp InputType = "vault"
)

// Jobs is a collection of job
type Jobs []Job

// Job represents a single job which should be executed during pipeline run.
// Handler is the function pointer to the function which will be executed.
type Job struct {
	Handler     func(Arguments) error
	Title       string
	Description string
	DependsOn   []string
	Args        Arguments
	Interaction *ManualInteraction
}

// Arguments is a collection of argument
type Arguments []Argument

// Argument represents a single argument.
type Argument struct {
	Description string
	Type        InputType
	Key         string
	Value       string
}

// ManualInteraction represents a manual interaction which can be set per job.
// Before the related job is executed, the manual interaction is displayed to
// the Gaia user.
type ManualInteraction struct {
	Description string
	Type        InputType
	Value       string
}

// jobsWrapper wraps a function pointer around the
// proto.Job struct.
// The given function corresponds to the job.
type jobsWrapper struct {
	funcPointer func(Arguments) error
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

// String returns a input type string back
func (i InputType) String() string {
	return string(i)
}
