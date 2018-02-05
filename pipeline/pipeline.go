package pipeline

import (
	"github.com/gaia-pipeline/gaia"
)

// BuildPipeline is the interface for pipelines which
// are not yet compiled.
type BuildPipeline interface {
	// PrepareEnvironment prepares the environment before we start the
	// build process.
	PrepareEnvironment(*gaia.CreatePipeline) error

	// ExecuteBuild executes the compiler and tracks the status of
	// the compiling process.
	ExecuteBuild(*gaia.CreatePipeline) error

	// CopyBinary copies the result from the compile process
	// to the plugins folder.
	CopyBinary(*gaia.CreatePipeline) error
}

const (
	// Temp folder where we store our temp files during build pipeline.
	tmpFolder = "tmp"

	// Max minutes until the build process will be interrupted and marked as failed
	maxTimeoutMinutes = 60
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
