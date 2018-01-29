package pipeline

import (
	"os/exec"

	"github.com/gaia-pipeline/gaia"
)

// BuildPipeline is the interface for pipelines which
// are not yet compiled.
type BuildPipeline interface {
	// PrepareBuild prepares the environment and command before
	// the build process is about to start.
	PrepareBuild(*gaia.Pipeline) (*exec.Cmd, error)

	// ExecuteBuild executes the compiler and tracks the status of
	// the compiling process.
	ExecuteBuild(*exec.Cmd) error

	// CopyBinary copies the result from the compile process
	// to the plugins folder.
	CopyBinary(*gaia.Pipeline) error
}

const (
	// Temp folder where we store our temp files during build pipeline.
	tmpFolder = "tmp"
)

// NewBuildPipeline creates a new build pipeline for the given
// pipeline type.
func NewBuildPipeline(t gaia.PipelineType) BuildPipeline {
	var bP BuildPipeline

	// Create build pipeline for given pipeline type
	switch t {
	case gaia.GOLANG:
		bP = &BuildPipelineGolang{
			Type: t,
		}
	}

	return bP
}
