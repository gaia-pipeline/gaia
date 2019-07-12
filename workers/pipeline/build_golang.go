package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	uuid "github.com/satori/go.uuid"
)

const (
	golangBinaryName = "go"
)

// BuildPipelineGolang is the real implementation of BuildPipeline for golang
type BuildPipelineGolang struct {
	Type gaia.PipelineType
}

// PrepareEnvironment prepares the environment before we start the build process.
func (b *BuildPipelineGolang) PrepareEnvironment(p *gaia.CreatePipeline) error {
	// create uuid for destination folder
	uuid := uuid.Must(uuid.NewV4(), nil)

	// Create local temp folder for clone
	goPath := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpGoFolder)
	cloneFolder := filepath.Join(goPath, srcFolder, uuid.String())
	err := os.MkdirAll(cloneFolder, 0700)
	if err != nil {
		return err
	}

	// Set new generated path in pipeline obj for later usage
	if p.Pipeline.Repo == nil {
		p.Pipeline.Repo = &gaia.GitRepo{}
	}
	p.Pipeline.Repo.LocalDest = cloneFolder
	p.Pipeline.UUID = uuid.String()
	return nil
}

// ExecuteBuild executes the golang build process
func (b *BuildPipelineGolang) ExecuteBuild(p *gaia.CreatePipeline) error {
	// Look for golang executable
	path, err := exec.LookPath(golangBinaryName)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot find go executable", "error", err.Error())
		return err
	}
	goPath := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpGoFolder)

	// Set command args for get dependencies
	args := []string{
		"get",
		"-d",
		"./...",
	}

	env := append(os.Environ(), "GOPATH="+goPath)

	// Set local destination
	localDest := ""
	if p.Pipeline.Repo != nil {
		localDest = p.Pipeline.Repo.LocalDest
	}

	// Execute and wait until finish or timeout
	output, err := executeCmd(path, args, env, localDest)
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
	output, err = executeCmd(path, args, env, localDest)
	p.Output = string(output)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot build pipeline", "error", err.Error(), "output", string(output))
		return err
	}

	// Build has been finished. Set execution path to the build result archive.
	// This will be used during pipeline verification phase which will happen after this step.
	p.Pipeline.ExecPath = filepath.Join(localDest, appendTypeToName(p.Pipeline.Name, p.Pipeline.Type))

	return nil
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
	return os.Chmod(dest, gaia.ExecutablePermission)
}

// SavePipeline saves the current pipeline configuration.
func (b *BuildPipelineGolang) SavePipeline(p *gaia.Pipeline) error {
	dest := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Name, p.Type))
	p.ExecPath = dest
	p.Type = gaia.PTypeGolang
	p.Name = strings.TrimSuffix(filepath.Base(dest), typeDelimiter+gaia.PTypeGolang.String())
	p.Created = time.Now()
	// Our pipeline is finished constructing. Save it.
	storeService, _ := services.StorageService()
	return storeService.PipelinePut(p)
}
