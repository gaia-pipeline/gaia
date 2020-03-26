package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia/helper/filehelper"
	"github.com/gaia-pipeline/gaia/helper/pipelinehelper"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gofrs/uuid"
)

const nodeJSInternalCloneFolder = "jsclone"

// BuildPipelineNodeJS is the real implementation of BuildPipeline for NodeJS
type BuildPipelineNodeJS struct {
	Type gaia.PipelineType
}

// PrepareEnvironment prepares the environment before we start the build process.
func (b *BuildPipelineNodeJS) PrepareEnvironment(p *gaia.CreatePipeline) error {
	// create uniqueName for destination folder
	v4, err := uuid.NewV4()
	uniqueName := uuid.Must(v4, nil)

	// Create local temp folder for clone
	cloneFolder := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpNodeJSFolder, srcFolder, uniqueName.String(), nodeJSInternalCloneFolder)
	err = os.MkdirAll(cloneFolder, 0700)
	if err != nil {
		return err
	}

	// Set new generated path in pipeline obj for later usage
	if p.Pipeline.Repo == nil {
		p.Pipeline.Repo = &gaia.GitRepo{}
	}
	p.Pipeline.Repo.LocalDest = cloneFolder
	p.Pipeline.UUID = uniqueName.String()
	return nil
}

// ExecuteBuild executes the NodeJS build process
func (b *BuildPipelineNodeJS) ExecuteBuild(p *gaia.CreatePipeline) error {
	// Look for Tar binary executable
	path, err := exec.LookPath(tarName)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot find tar binary executable", "error", err.Error())
		return err
	}

	// Set local destination
	localDest := ""
	if p.Pipeline.Repo != nil {
		localDest = p.Pipeline.Repo.LocalDest
	}

	// Set command args for archive process
	pipelineFileName := pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type)
	args := []string{
		"--exclude=.git",
		"-czvf",
		pipelineFileName,
		"-C",
		localDest,
		".",
	}

	// Execute and wait until finish or timeout
	uniqueFolder := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpNodeJSFolder, srcFolder, p.Pipeline.UUID)
	output, err := executeCmd(path, args, os.Environ(), uniqueFolder)
	p.Output = string(output[:])
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot build pipeline", "error", err.Error(), "output", string(output[:]))
		return err
	}

	// Build has been finished. Set execution path to the build result archive.
	// This will be used during pipeline verification phase which will happen after this step.
	p.Pipeline.ExecPath = filepath.Join(uniqueFolder, pipelineFileName)

	// Set the the local destination variable to the unique folder because this is now the place
	// where our binary is located.
	p.Pipeline.Repo.LocalDest = uniqueFolder

	return nil
}

// CopyBinary copies the final compiled binary to the
// destination folder.
func (b *BuildPipelineNodeJS) CopyBinary(p *gaia.CreatePipeline) error {
	// Define src and destination
	src := filepath.Join(p.Pipeline.Repo.LocalDest, pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type))
	dest := filepath.Join(gaia.Cfg.PipelinePath, pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type))

	// Copy binary
	if err := filehelper.CopyFileContents(src, dest); err != nil {
		return err
	}

	// Set +x (execution right) for pipeline
	return os.Chmod(dest, gaia.ExecutablePermission)
}

// SavePipeline saves the current pipeline configuration.
func (b *BuildPipelineNodeJS) SavePipeline(p *gaia.Pipeline) error {
	dest := filepath.Join(gaia.Cfg.PipelinePath, pipelinehelper.AppendTypeToName(p.Name, p.Type))
	p.ExecPath = dest
	p.Type = gaia.PTypeNodeJS
	p.Name = strings.TrimSuffix(filepath.Base(dest), typeDelimiter+gaia.PTypeNodeJS.String())
	p.Created = time.Now()
	// Our pipeline is finished constructing. Save it.
	storeService, _ := services.StorageService()
	return storeService.PipelinePut(p)
}
