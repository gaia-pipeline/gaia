package pipeline

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/satori/go.uuid"
)

const (
	golangBinaryName = "go"
	golangFolder     = "golang"
	srcFolder        = "src"
)

var execCommandContext = exec.CommandContext

// BuildPipelineGolang is the real implementation of BuildPipeline for golang
type BuildPipelineGolang struct {
	Type gaia.PipelineType
}

// PrepareEnvironment prepares the environment before we start the build process.
func (b *BuildPipelineGolang) PrepareEnvironment(p *gaia.CreatePipeline) error {
	// create uuid for destination folder
	uuid := uuid.Must(uuid.NewV4(), nil)

	// Create local temp folder for clone
	goPath := filepath.Join(gaia.Cfg.HomePath, tmpFolder, golangFolder)
	cloneFolder := filepath.Join(goPath, srcFolder, uuid.String())
	err := os.MkdirAll(cloneFolder, 0700)
	if err != nil {
		return err
	}

	// Set new generated path in pipeline obj for later usage
	p.Pipeline.Repo.LocalDest = cloneFolder
	p.Pipeline.UUID = uuid.String()
	return err
}

// ExecuteBuild executes the golang build process
func (b *BuildPipelineGolang) ExecuteBuild(p *gaia.CreatePipeline) error {
	// Look for golang executeable
	path, err := exec.LookPath(golangBinaryName)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot find go executeable", "error", err.Error())
		return err
	}
	goPath := filepath.Join(gaia.Cfg.HomePath, tmpFolder, golangFolder)

	// Set command args for get dependencies
	args := []string{
		"get",
		"-d",
		"./...",
	}

	env := append(os.Environ(), "GOPATH="+goPath)

	// Execute and wait until finish or timeout
	output, err := executeCmd(path, args, env, p.Pipeline.Repo.LocalDest)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot get dependencies", "error", err.Error(), "output", string(output))
		p.Output = string(output)
		return err
	}

	// Set command args for build
	args = []string{
		"build",
		"-o",
		appendTypeToName(p.Pipeline.Name, p.Pipeline.Type),
	}

	// Execute and wait until finish or timeout
	output, err = executeCmd(path, args, env, p.Pipeline.Repo.LocalDest)
	p.Output = string(output)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot build pipeline", "error", err.Error(), "output", string(output))
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
	cmd := execCommandContext(ctx, path, args...)
	cmd.Env = env
	cmd.Dir = dir

	// Execute command
	return cmd.CombinedOutput()
}

// CopyBinary copies the final compiled archive to the
// destination folder.
func (b *BuildPipelineGolang) CopyBinary(p *gaia.CreatePipeline) error {
	// Define src and destination
	src := filepath.Join(p.Pipeline.Repo.LocalDest, appendTypeToName(p.Pipeline.Name, p.Pipeline.Type))
	dest := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Pipeline.Name, p.Pipeline.Type))

	// Copy binary
	if err := copyFileContents(src, dest); err != nil {
		return err
	}

	// Set +x (execution right) for pipeline
	return os.Chmod(dest, 0766)
}

// SavePipeline saves the current pipeline configuration.
func (b *BuildPipelineGolang) SavePipeline(p *gaia.Pipeline) error {
	dest := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Name, p.Type))
	p.ExecPath = dest
	p.Type = gaia.PTypeGolang
	p.Name = strings.TrimSuffix(filepath.Base(dest), typeDelimiter+gaia.PTypeGolang.String())
	p.Created = time.Now()
	// Our pipeline is finished constructing. Save it.
	return storeService.PipelinePut(p)
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
