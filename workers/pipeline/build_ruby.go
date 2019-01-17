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
	gemBinaryName = "gem"
)

// BuildPipelineRuby is the real implementation of BuildPipeline for Ruby
type BuildPipelineRuby struct {
	Type gaia.PipelineType
}

// PrepareEnvironment prepares the environment before we start the build process.
func (b *BuildPipelineRuby) PrepareEnvironment(p *gaia.CreatePipeline) error {
	// create uuid for destination folder
	uuid := uuid.Must(uuid.NewV4(), nil)

	// Create local temp folder for clone
	cloneFolder := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpRubyFolder, srcFolder, uuid.String())
	err := os.MkdirAll(cloneFolder, 0700)
	if err != nil {
		return err
	}

	// Set new generated path in pipeline obj for later usage
	p.Pipeline.Repo.LocalDest = cloneFolder
	p.Pipeline.UUID = uuid.String()
	return err
}

// ExecuteBuild executes the ruby build process
func (b *BuildPipelineRuby) ExecuteBuild(p *gaia.CreatePipeline) error {
	// Look for gem binary executable
	path, err := exec.LookPath(gemBinaryName)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot find gem binary executable", "error", err.Error())
		return err
	}

	// Get all gemspec files in cloned folder.
	gemspec, err := filterPathContentBySuffix(p.Pipeline.Repo.LocalDest, ".gemspec")
	if err != nil {
		gaia.Cfg.Logger.Error("cannot find gemspec file in cloned repository folder", "path", p.Pipeline.Repo.LocalDest)
		return err
	}

	// if we found more or less than one gemspec we have a problem.
	if len(gemspec) != 1 {
		gaia.Cfg.Logger.Debug("cannot find gemspec file in cloned repo", "foundGemspecs", len(gemspec), "gemspecs", gemspec)
		return errors.New("cannot find gemspec file in cloned repo")
	}

	// Set command args for build
	args := []string{
		"build",
		gemspec[0],
	}

	// Execute and wait until finish or timeout
	output, err := executeCmd(path, args, os.Environ(), p.Pipeline.Repo.LocalDest)
	p.Output = string(output)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot build pipeline", "error", err.Error(), "output", string(output))
		return err
	}

	// Search for resulting gem file.
	gemfile, err := filterPathContentBySuffix(p.Pipeline.Repo.LocalDest, ".gem")
	if err != nil {
		gaia.Cfg.Logger.Error("cannot find final gem file after build", "path", p.Pipeline.Repo.LocalDest)
		return err
	}

	// if we found more or less than one gem file then we have a problem.
	if len(gemfile) != 1 {
		gaia.Cfg.Logger.Debug("cannot find gem file in cloned repo", "foundGemFiles", len(gemfile), "gems", gemfile)
		return errors.New("cannot find gem file in cloned repo")
	}

	// Build has been finished. Set execution path to the build result archive.
	// This will be used during pipeline verification phase which will happen after this step.
	p.Pipeline.ExecPath = gemfile[0]

	return nil
}

// filterPathContentBySuffix reads the whole directory given by path and
// returns all files with full path which have the same suffix like provided.
func filterPathContentBySuffix(path, suffix string) ([]string, error) {
	filteredFiles := []string{}

	// Read complete directory.
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return filteredFiles, err
	}

	// filter for files ending with given suffix.
	for _, file := range files {
		if strings.HasSuffix(file.Name(), suffix) {
			filteredFiles = append(filteredFiles, filepath.Join(path, file.Name()))
		}
	}
	return filteredFiles, nil
}

// CopyBinary copies the final compiled binary to the
// destination folder.
func (b *BuildPipelineRuby) CopyBinary(p *gaia.CreatePipeline) error {
	// Search for resulting gem file.
	gemfile, err := filterPathContentBySuffix(p.Pipeline.Repo.LocalDest, ".gem")
	if err != nil {
		gaia.Cfg.Logger.Error("cannot find final gem file during copy", "path", p.Pipeline.Repo.LocalDest)
		return err
	}

	// Define src and destination
	src := gemfile[0]
	dest := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Pipeline.Name, p.Pipeline.Type))

	// Copy binary
	if err := copyFileContents(src, dest); err != nil {
		return err
	}

	// Set +x (execution right) for pipeline
	return os.Chmod(dest, 0766)
}

// SavePipeline saves the current pipeline configuration.
func (b *BuildPipelineRuby) SavePipeline(p *gaia.Pipeline) error {
	dest := filepath.Join(gaia.Cfg.PipelinePath, appendTypeToName(p.Name, p.Type))
	p.ExecPath = dest
	p.Type = gaia.PTypeRuby
	p.Name = strings.TrimSuffix(filepath.Base(dest), typeDelimiter+gaia.PTypeRuby.String())
	p.Created = time.Now()
	// Our pipeline is finished constructing. Save it.
	storeService, _ := services.StorageService()
	return storeService.PipelinePut(p)
}
