package pipeline

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gaia-pipeline/gaia"
)

var (
	// virtualEnvName is the binary name of virtual environment app.
	virtualEnvName = "virtualenv"

	// pythonPipInstallCmd is the command used to install the python distribution
	// package.
	pythonPipInstallCmd = ". bin/activate; python -m pip install %s.tar.gz"
)

// updatePipeline executes update steps dependent on the pipeline type.
// Some pipeline types may don't require this.
func updatePipeline(p *gaia.Pipeline) error {
	switch p.Type {
	case gaia.PTypePython:
		// Remove virtual environment if existend
		virtualEnvPath := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpPythonFolder, p.Name)
		os.RemoveAll(virtualEnvPath)

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
		err = copyFileContents(p.ExecPath, filepath.Join(virtualEnvPath, p.Name+".tar.gz"))
		if err != nil {
			return err
		}

		// install plugin in this environment
		cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf(pythonPipInstallCmd, filepath.Join(virtualEnvPath, p.Name)))
		cmd.Dir = virtualEnvPath
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// Update checksum
	checksum, err := getSHA256Sum(p.ExecPath)
	if err != nil {
		return err
	}
	p.SHA256Sum = checksum

	return nil
}
