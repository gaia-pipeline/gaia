package pipeline

import (
	"context"
	"os"
	"os/exec"
	"time"

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

// PrepareEnvironment prepares the environment before we start the build process.
func (b *BuildPipelineGolang) PrepareEnvironment(p *gaia.CreatePipeline) error {
	// create uuid for destination folder
	uuid := uuid.Must(uuid.NewV4())

	// Create local temp folder for clone
	goPath := gaia.Cfg.HomePath + string(os.PathSeparator) + tmpFolder
	cloneFolder := goPath + string(os.PathSeparator) + srcFolder + string(os.PathSeparator) + uuid.String()
	err := os.MkdirAll(cloneFolder, 0700)
	if err != nil {
		return err
	}

	// Set new generated path in pipeline obj for later usage
	p.Pipeline.Repo.LocalDest = cloneFolder
	return nil
}

// ExecuteBuild executes the golang build process
func (b *BuildPipelineGolang) ExecuteBuild(p *gaia.CreatePipeline) error {
	// Look for golang executeable
	path, err := exec.LookPath(golangBinaryName)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot find go executeable", "error", err.Error())
		return err
	}
	goPath := gaia.Cfg.HomePath + string(os.PathSeparator) + tmpFolder

	// Set command args for get dependencies
	args := []string{
		path,
		"get",
		"./...",
	}
	env := []string{
		"GOPATH=" + goPath,
	}

	// Execute and wait until finish or timeout
	output, err := executeCmd(path, args, env, p.Pipeline.Repo.LocalDest)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot get dependencies", "error", err.Error())
		p.Output = string(output)
		return err
	}

	// Set command args for build
	env = []string{
		"GOPATH=" + goPath,
	}
	args = []string{
		path,
		"build",
		"-o",
		p.Pipeline.Name,
	}

	// Execute and wait until finish or timeout
	output, err = executeCmd(path, args, env, p.Pipeline.Repo.LocalDest)
	p.Output = string(output)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot build pipeline", "error", err.Error())
		return err
	}

	return nil
}

// executeCmd wraps a context around the command and executes it.
func executeCmd(path string, args []string, env []string, dir string) ([]byte, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), maxTimeoutMinutes*time.Minute)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = env
	cmd.Dir = dir

	// Execute command
	return cmd.CombinedOutput()
}

// CopyBinary copies the final compiled archive to the
// destination folder.
func (b *BuildPipelineGolang) CopyBinary(p *gaia.CreatePipeline) error {
	// TODO
	return nil
}
