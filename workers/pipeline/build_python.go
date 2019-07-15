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
	uuid "github.com/satori/go.uuid"
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
	// create uniqueName for destination folder
	uniqueName := uuid.Must(uuid.NewV4(), nil)

	// Create local temp folder for clone
	rootPath := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpPythonFolder)
	cloneFolder := filepath.Join(rootPath, srcFolder, uniqueName.String())
	err := os.MkdirAll(cloneFolder, 0700)
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

	// Set local destination
	localDest := ""
	if p.Pipeline.Repo != nil {
		localDest = p.Pipeline.Repo.LocalDest
	}

	// Execute and wait until finish or timeout
	output, err := executeCmd(path, args, os.Environ(), localDest)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot generate python distribution package", "error", err.Error(), "output", string(output))
		p.Output = string(output)
		return err
	}

	// Build has been finished. Set execution path to the build result archive.
	// This will be used during pipeline verification phase which will happen after this step.
	p.Pipeline.ExecPath, err = findPythonArchivePath(p)
	if err != nil {
		return err
	}

	return nil
}

// findPythonArchivePath filters the archives in the generated dist folder
// and looks for the final archive. It will return an error if less or more
// than one file(s) are found otherwise the full path to the file.
func findPythonArchivePath(p *gaia.CreatePipeline) (src string, err error) {
	// find all files in dist folder
	distFolder := filepath.Join(p.Pipeline.Repo.LocalDest, "dist")
	files, err := ioutil.ReadDir(distFolder)
	if err != nil {
		return
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
		gaia.Cfg.Logger.Debug("cannot find python package", "foundPackages", len(archive), "archives", files)
		err = errors.New("cannot find python package")
		return
	}

	// Return full path
	src = filepath.Join(distFolder, archive[0].Name())
	return
}

// CopyBinary copies the final compiled archive to the
// destination folder.
func (b *BuildPipelinePython) CopyBinary(p *gaia.CreatePipeline) error {
	// Define src and destination
	src, err := findPythonArchivePath(p)
	if err != nil {
		return err
	}
	dest := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Pipeline.Name, p.Pipeline.Type))

	// Copy binary
	if err := copyFileContents(src, dest); err != nil {
		return err
	}

	// Set +x (execution right) for pipeline
	return os.Chmod(dest, gaia.ExecutablePermission)
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
