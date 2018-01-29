package pipeline

import (
	"os"
	"os/exec"

	"github.com/gaia-pipeline/gaia"
	"github.com/satori/go.uuid"
)

const (
	golangBinaryName = "go"
	srcFolder        = "src"
)

// BuildPipelineGolang is the real implementation of BuildPipeline for golang
type BuildPipelineGolang struct {
	Type gaia.PipelineType
}

// PrepareBuild prepares the environment and command before
// starting the build process.
func (b *BuildPipelineGolang) PrepareBuild(p *gaia.Pipeline) (*exec.Cmd, error) {
	// create uuid for destination folder
	uuid := uuid.Must(uuid.NewV4())

	// Create local temp folder for clone
	goPath := gaia.Cfg.HomePath + string(os.PathSeparator) + tmpFolder
	cloneFolder := goPath + string(os.PathSeparator) + srcFolder + string(os.PathSeparator) + uuid.String()
	err := os.MkdirAll(cloneFolder, 0700)
	if err != nil {
		return nil, err
	}

	// Set new generated path in pipeline obj for later usage
	p.Repo.LocalDest = cloneFolder

	// Create empty command
	c := &exec.Cmd{}

	// Look for golang executeable
	path, err := exec.LookPath(golangBinaryName)
	if err != nil {
		return nil, err
	}

	// Set command args
	c.Path = path
	c.Dir = cloneFolder
	c.Env = []string{
		"GOPATH=" + goPath,
	}
	c.Args = []string{
		"build",
		"-x",
	}

	// return command
	return c, nil
}

// ExecuteBuild executes the golang build process
func (b *BuildPipelineGolang) ExecuteBuild(cmd *exec.Cmd) error {
	// TODO
	return nil
}

// CopyBinary copies the final compiled archive to the
// destination folder.
func (b *BuildPipelineGolang) CopyBinary(p *gaia.Pipeline) error {
	// TODO
	return nil
}
