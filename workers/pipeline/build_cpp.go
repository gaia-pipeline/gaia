package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/satori/go.uuid"
)

const (
	cppBinaryName      = "make"
	cppFinalBinaryName = "pipeline.out"
)

// BuildPipelineCpp is the real implementation of BuildPipeline for C++
type BuildPipelineCpp struct {
	Type gaia.PipelineType
}

// PrepareEnvironment prepares the environment before we start the build process.
func (b *BuildPipelineCpp) PrepareEnvironment(p *gaia.CreatePipeline) error {
	// create uuid for destination folder
	uuid := uuid.Must(uuid.NewV4(), nil)

	// Create local temp folder for clone
	cloneFolder := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpCppFolder, srcFolder, uuid.String())
	err := os.MkdirAll(cloneFolder, 0700)
	if err != nil {
		return err
	}

	// Set new generated path in pipeline obj for later usage
	p.Pipeline.Repo.LocalDest = cloneFolder
	p.Pipeline.UUID = uuid.String()
	return err
}

// ExecuteBuild executes the c++ build process
func (b *BuildPipelineCpp) ExecuteBuild(p *gaia.CreatePipeline) error {
	// Look for c++ binary executable
	path, err := exec.LookPath(cppBinaryName)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot find c++ binary executable", "error", err.Error())
		return err
	}

	// Set command args for build
	args := []string{}

	// Execute and wait until finish or timeout
	output, err := executeCmd(path, args, os.Environ(), p.Pipeline.Repo.LocalDest)
	p.Output = string(output)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot build pipeline", "error", err.Error(), "output", string(output))
		return err
	}

	// Build has been finished. Set execution path to the build result archive.
	// This will be used during pipeline verification phase which will happen after this step.
	p.Pipeline.ExecPath = filepath.Join(p.Pipeline.Repo.LocalDest, cppFinalBinaryName)

	return nil
}

// CopyBinary copies the final compiled binary to the
// destination folder.
func (b *BuildPipelineCpp) CopyBinary(p *gaia.CreatePipeline) error {
	// Define src and destination
	src := filepath.Join(p.Pipeline.Repo.LocalDest, cppFinalBinaryName)
	dest := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Pipeline.Name, p.Pipeline.Type))

	// Copy binary
	if err := copyFileContents(src, dest); err != nil {
		return err
	}

	// Set +x (execution right) for pipeline
	return os.Chmod(dest, 0766)
}

// SavePipeline saves the current pipeline configuration.
func (b *BuildPipelineCpp) SavePipeline(p *gaia.Pipeline) error {
	dest := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Name, p.Type))
	p.ExecPath = dest
	p.Type = gaia.PTypeCpp
	p.Name = strings.TrimSuffix(filepath.Base(dest), typeDelimiter+gaia.PTypeCpp.String())
	p.Created = time.Now()
	// Our pipeline is finished constructing. Save it.
	storeService, _ := services.StorageService()
	return storeService.PipelinePut(p)
}
