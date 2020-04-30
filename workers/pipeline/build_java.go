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

var (
	mavenBinaryName = "mvn"
)

const (
	javaFolder        = "java"
	javaFinalJarName  = "plugin-jar-with-dependencies.jar"
	mavenTargetFolder = "target"
)

// BuildPipelineJava is the real implementation of BuildPipeline for java
type BuildPipelineJava struct {
	Type gaia.PipelineType
}

// PrepareEnvironment prepares the environment before we start the build process.
func (b *BuildPipelineJava) PrepareEnvironment(p *gaia.CreatePipeline) error {
	// create uniqueName for destination folder
	v4, err := uuid.NewV4()
	if err != nil {
		return err
	}
	uniqueName := uuid.Must(v4, nil)

	// Create local temp folder for clone
	rootPath := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, javaFolder)
	cloneFolder := filepath.Join(rootPath, srcFolder, uniqueName.String())
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

// ExecuteBuild executes the java build process
func (b *BuildPipelineJava) ExecuteBuild(p *gaia.CreatePipeline) error {
	// Look for maven executable
	path, err := exec.LookPath(mavenBinaryName)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot find maven executeable", "error", err.Error())
		return err
	}
	env := os.Environ()

	// Set command args for build
	args := []string{
		"clean",
		"compile",
		"assembly:single",
	}

	// Set local destination
	localDest := ""
	if p.Pipeline.Repo != nil {
		localDest = p.Pipeline.Repo.LocalDest
	}

	// Execute and wait until finish or timeout
	output, err := executeCmd(path, args, env, localDest)
	p.Output = string(output)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot build pipeline", "error", err.Error(), "output", string(output))
		return err
	}

	// Build has been finished. Set execution path to the build result archive.
	// This will be used during pipeline verification phase which will happen after this step.
	p.Pipeline.ExecPath = filepath.Join(localDest, mavenTargetFolder, javaFinalJarName)

	return nil
}

// CopyBinary copies the final compiled archive to the
// destination folder.
func (b *BuildPipelineJava) CopyBinary(p *gaia.CreatePipeline) error {
	// Define src and destination
	src := filepath.Join(p.Pipeline.Repo.LocalDest, mavenTargetFolder, javaFinalJarName)
	dest := filepath.Join(gaia.Cfg.PipelinePath, pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type))

	// Copy binary
	if err := filehelper.CopyFileContents(src, dest); err != nil {
		return err
	}

	// Set +x (execution right) for pipeline
	return os.Chmod(dest, gaia.ExecutablePermission)
}

// SavePipeline saves the current pipeline configuration.
func (b *BuildPipelineJava) SavePipeline(p *gaia.Pipeline) error {
	dest := filepath.Join(gaia.Cfg.PipelinePath, pipelinehelper.AppendTypeToName(p.Name, p.Type))
	p.ExecPath = dest
	p.Type = gaia.PTypeJava
	p.Name = strings.TrimSuffix(filepath.Base(dest), typeDelimiter+gaia.PTypeJava.String())
	p.Created = time.Now()
	// Our pipeline is finished constructing. Save it.
	storeService, _ := services.StorageService()
	return storeService.PipelinePut(p)
}
