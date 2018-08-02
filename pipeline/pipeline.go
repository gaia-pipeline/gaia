package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

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

	// SavePipeline the pipeline in its current format
	SavePipeline(*gaia.Pipeline) error
}

// ActivePipelines holds all active pipelines.
// ActivePipelines can be safely shared between goroutines.
type ActivePipelines struct {
	sync.RWMutex

	// All active pipelines
	Pipelines []gaia.Pipeline
}

const (
	// Temp folder where we store our temp files during build pipeline.
	tmpFolder = "tmp"

	// Max minutes until the build process will be interrupted and marked as failed
	maxTimeoutMinutes = 60

	// typeDelimiter defines the delimiter in the file name to define
	// the pipeline type.
	typeDelimiter = "_"
)

var (
	// GlobalActivePipelines holds globally all current active pipleines.
	GlobalActivePipelines *ActivePipelines

	// errMissingType is the error thrown when a pipeline is missing the type
	// in the file name.
	errMissingType = errors.New("couldnt find pipeline type definition")
)

// execution command context used for build
var execCommandContext = exec.CommandContext

// Source folder name where the sources are stored
const srcFolder = "src"

// newBuildPipeline creates a new build pipeline for the given
// pipeline type.
func newBuildPipeline(t gaia.PipelineType) BuildPipeline {
	var bP BuildPipeline

	// Create build pipeline for given pipeline type
	switch t {
	case gaia.PTypeGolang:
		bP = &BuildPipelineGolang{
			Type: t,
		}
	case gaia.PTypeJava:
		bP = &BuildPipelineJava{
			Type: t,
		}
	}

	return bP
}

// NewActivePipelines creates a new instance of ActivePipelines
func NewActivePipelines() *ActivePipelines {
	ap := &ActivePipelines{
		Pipelines: make([]gaia.Pipeline, 0),
	}

	return ap
}

// Append appends a new pipeline to ActivePipelines.
func (ap *ActivePipelines) Append(p gaia.Pipeline) {
	ap.Lock()
	defer ap.Unlock()

	ap.Pipelines = append(ap.Pipelines, p)
}

// Update updates a pipeline at the given index with the given pipeline.
func (ap *ActivePipelines) Update(index int, p gaia.Pipeline) {
	ap.Lock()
	defer ap.Unlock()

	ap.Pipelines[index] = p
}

// Remove removes a pipeline at the given index from ActivePipelines.
func (ap *ActivePipelines) Remove(index int) {
	ap.Lock()
	defer ap.Unlock()

	ap.Pipelines = append(ap.Pipelines[:index], ap.Pipelines[index+1:]...)
}

// GetByName looks up the pipeline by the given name.
func (ap *ActivePipelines) GetByName(n string) *gaia.Pipeline {
	var foundPipeline gaia.Pipeline
	for pipeline := range ap.Iter() {
		if pipeline.Name == n {
			foundPipeline = pipeline
		}
	}

	if foundPipeline.Name == "" {
		return nil
	}

	return &foundPipeline
}

// Replace takes the given pipeline and replaces it in the ActivePipelines
// slice. Return true when success otherwise false.
func (ap *ActivePipelines) Replace(p gaia.Pipeline) bool {
	ap.Lock()
	defer ap.Unlock()

	// Search for the id
	var i = -1
	for id, pipeline := range ap.Pipelines {
		if pipeline.Name == p.Name {
			i = id
			break
		}
	}

	// We got it?
	if i == -1 {
		return false
	}

	// Yes
	ap.Pipelines[i] = p
	return true
}

// ReplaceByName replaces the pipeline that has the given name with the given pipeline.
func (ap *ActivePipelines) ReplaceByName(n string, p gaia.Pipeline) bool {

	var index int
	var pipelineIndex int
	var found bool

	for pipeline := range ap.Iter() {
		if pipeline.Name == n {
			found = true
			pipelineIndex = index
		}
		index++
	}

	if found {
		ap.Update(pipelineIndex, p)
	}

	return found

}

// Iter iterates over the pipelines in the concurrent slice.
func (ap *ActivePipelines) Iter() <-chan gaia.Pipeline {
	c := make(chan gaia.Pipeline)

	go func() {
		ap.RLock()
		defer ap.RUnlock()
		for _, pipeline := range ap.Pipelines {
			c <- pipeline
		}
		close(c)
	}()

	return c
}

// Contains checks if the given pipeline name has been already appended
// to the given ActivePipelines instance.
func (ap *ActivePipelines) Contains(n string) bool {
	var foundPipeline bool
	for pipeline := range ap.Iter() {
		if pipeline.Name == n {
			foundPipeline = true
		}
	}

	return foundPipeline
}

// RemoveDeletedPipelines removes the pipelines whose names are NOT
// present in `existingPipelineNames` from the given ActivePipelines instance.
func (ap *ActivePipelines) RemoveDeletedPipelines(existingPipelineNames []string) {
	var deletedPipelineIndices []int
	var index int
	for pipeline := range ap.Iter() {
		found := false
		for _, name := range existingPipelineNames {
			if pipeline.Name == name {
				found = true
				break
			}
		}
		if !found {
			deletedPipelineIndices = append(deletedPipelineIndices, index)
		}
		index++
	}
	for _, idx := range deletedPipelineIndices {
		ap.Remove(idx)
	}
}

// RenameBinary renames the binary file for the given pipeline.
func RenameBinary(p gaia.Pipeline, newName string) error {
	currentBinaryName := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Name, p.Type))
	newBinaryName := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(newName, p.Type))
	return os.Rename(currentBinaryName, newBinaryName)
}

// DeleteBinary deletes the binary for the given pipeline.
func DeleteBinary(p gaia.Pipeline) error {
	binaryFile := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Name, p.Type))
	return os.Remove(binaryFile)
}

// GetExecPath returns the path to the executable for the given pipeline.
func GetExecPath(p gaia.Pipeline) string {
	return filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Name, p.Type))
}

// appendTypeToName appends the type to the output binary name.
// This allows us later to define the pipeline type by the name.
func appendTypeToName(n string, pType gaia.PipelineType) string {
	return fmt.Sprintf("%s%s%s", n, typeDelimiter, pType.String())
}

// executeCmd wraps a context around the command and executes it.
func executeCmd(path string, args []string, env []string, dir string) ([]byte, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), maxTimeoutMinutes*time.Minute)
	defer cancel()

	// Create command
	cmd := execCommandContext(ctx, path, args...)
	cmd.Env = env
	cmd.Dir = dir

	// Execute command
	return cmd.CombinedOutput()
}

// copyFileContents copies the content from source to destination.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
