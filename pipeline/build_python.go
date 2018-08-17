package pipeline

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/satori/go.uuid"
)

var (
	pythonBinaryName = "python"
)

// BuildPipelinePython is the real implementation of BuildPipeline for python
type BuildPipelinePython struct {
	Type gaia.PipelineType
}

// PrepareEnvironment prepares the environment before we start the build process.
func (b *BuildPipelinePython) PrepareEnvironment(p *gaia.CreatePipeline) error {
	// create uuid for destination folder
	uuid := uuid.Must(uuid.NewV4(), nil)

	// Create local temp folder for clone
	rootPath := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpPythonFolder)
	cloneFolder := filepath.Join(rootPath, srcFolder, uuid.String())
	err := os.MkdirAll(cloneFolder, 0700)
	if err != nil {
		return err
	}

	// Set new generated path in pipeline obj for later usage
	p.Pipeline.Repo.LocalDest = cloneFolder
	p.Pipeline.UUID = uuid.String()
	return err
}

// ExecuteBuild executes the python build process
func (b *BuildPipelinePython) ExecuteBuild(p *gaia.CreatePipeline) error {
	// Look for python executeable
	path, err := exec.LookPath(pythonBinaryName)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot find python executeable", "error", err.Error())
		return err
	}

	// Set command args for build distribution package
	args := []string{
		"setup.py",
		"sdist",
	}

	// Execute and wait until finish or timeout
	output, err := executeCmd(path, args, os.Environ(), p.Pipeline.Repo.LocalDest)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot generate python distribution package", "error", err.Error(), "output", string(output))
		p.Output = string(output)
		return err
	}

	return nil
}

// CopyBinary copies the final compiled archive to the
// destination folder.
func (b *BuildPipelinePython) CopyBinary(p *gaia.CreatePipeline) error {
	// find all files in dist folder
	distFolder := filepath.Join(p.Pipeline.Repo.LocalDest, "dist")
	files, err := ioutil.ReadDir(distFolder)
	if err != nil {
		return err
	}

	// filter for archives
	archive := []os.FileInfo{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tar.gz") {
			archive = append(archive, file)
		}
	}

	// if we found more or less than one archive we have a problem
	if len(archive) != 1 {
		gaia.Cfg.Logger.Debug("cannot copy python package", "foundPackages", len(archive), "archives", files)
		return errors.New("cannot copy python package: not found")
	}

	// Define src and destination
	src := filepath.Join(distFolder, archive[0].Name())
	dest := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Pipeline.Name, p.Pipeline.Type))

	// Copy binary
	if err := copyFileContents(src, dest); err != nil {
		return err
	}

	// Set +x (execution right) for pipeline
	return os.Chmod(dest, 0766)
}

// SavePipeline saves the current pipeline configuration.
func (b *BuildPipelinePython) SavePipeline(p *gaia.Pipeline) error {
	dest := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Name, p.Type))
	p.ExecPath = dest
	p.Type = gaia.PTypePython
	p.Name = strings.TrimSuffix(filepath.Base(dest), typeDelimiter+gaia.PTypePython.String())
	p.Created = time.Now()
	// Our pipeline is finished constructing. Save it.
	storeService, _ := services.StorageService()
	return storeService.PipelinePut(p)
}
