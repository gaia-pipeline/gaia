package pipeline

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gaia-pipeline/gaia"
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
		path, err := exec.LookPath("virtualenv")
		if err != nil {
			return errors.New("cannot find virtualenv executeable")
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
		cmd = exec.Command("/bin/sh", "-c", "source bin/activate; python -m pip install "+filepath.Join(virtualEnvPath, p.Name+".tar.gz"))
		cmd.Dir = virtualEnvPath
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
