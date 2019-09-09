package pipeline

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gaia-pipeline/gaia/helper/filehelper"

	"github.com/gaia-pipeline/gaia"
)

var (
	// virtualEnvName is the binary name of virtual environment app.
	virtualEnvName = "virtualenv"

	// pythonPipInstallCmd is the command used to install the python distribution
	// package.
	pythonPipInstallCmd = ". bin/activate; python -m pip install '%s.tar.gz'"

	// Ruby gem binary name.
	rubyGemName = "gem"

	// Tar binary name.
	tarName = "tar"

	// NPM binary name.
	npmName = "npm"
)

// updatePipeline executes update steps dependent on the pipeline type.
// Some pipeline types may don't require this.
func updatePipeline(p *gaia.Pipeline) error {
	switch p.Type {
	case gaia.PTypePython:
		// Remove virtual environment if exists
		virtualEnvPath := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpPythonFolder, p.Name)
		_ = os.RemoveAll(virtualEnvPath)

		// Create virtual environment
		path, err := exec.LookPath(virtualEnvName)
		if err != nil {
			return errors.New("cannot find virtualenv executable")
		}
		cmd := exec.Command(path, virtualEnvPath)
		if err := cmd.Run(); err != nil {
			return err
		}

		// copy distribution file to environment and remove pipeline type at the end.
		// we have to do this otherwise pip will fail.
		err = filehelper.CopyFileContents(p.ExecPath, filepath.Join(virtualEnvPath, p.Name+".tar.gz"))
		if err != nil {
			return err
		}

		// install plugin in this environment
		cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf(pythonPipInstallCmd, filepath.Join(virtualEnvPath, p.Name)))
		cmd.Dir = virtualEnvPath
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cannot install python plugin: %s", string(out[:]))
		}
	case gaia.PTypeRuby:
		// Find gem binary in path variable.
		path, err := exec.LookPath(rubyGemName)
		if err != nil {
			return err
		}

		// Gem expects that the file suffix is ".gem".
		// Copy gem file to temp folder before we install it.
		tmpFolder := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpRubyFolder)
		err = os.MkdirAll(tmpFolder, 0700)
		if err != nil {
			return err
		}
		pipelineCopyPath := filepath.Join(tmpFolder, filepath.Base(p.ExecPath)+".gem")
		err = filehelper.CopyFileContents(p.ExecPath, pipelineCopyPath)
		if err != nil {
			return err
		}
		defer os.Remove(pipelineCopyPath)

		// Install gem forcefully.
		cmd := exec.Command(path, "install", "-f", pipelineCopyPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cannot install ruby gem: %s", string(out[:]))
		}
	case gaia.PTypeNodeJS:
		// Find tar binary in path
		path, err := exec.LookPath(tarName)
		if err != nil {
			return err
		}

		// Delete old folders if exist
		tmpFolder := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpNodeJSFolder, p.Name)
		_ = os.RemoveAll(tmpFolder)

		// Recreate the temp folder
		if err := os.MkdirAll(tmpFolder, 0700); err != nil {
			return err
		}

		// Unpack it
		cmd := exec.Command(path, "-xzvf", p.ExecPath, "-C", tmpFolder)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cannot unpack nodejs archive: %s", string(out[:]))
		}

		// Find npm binary in path
		path, err = exec.LookPath(npmName)
		if err != nil {
			return err
		}

		// Install dependencies
		cmd = &exec.Cmd{
			Path: path,
			Dir:  tmpFolder,
			Args: []string{path, "install"},
		}
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cannot install dependencies: %s", string(out[:]))
		}
	}

	// Update checksum
	checksum, err := filehelper.GetSHA256Sum(p.ExecPath)
	if err != nil {
		return err
	}
	p.SHA256Sum = checksum

	return nil
}
