package gaiascheduler

import (
	"os/exec"
	"path/filepath"

	"github.com/gaia-pipeline/gaia"
	"gopkg.in/yaml.v2"
)

// createPipelineCmd creates the execute command for the plugin system
// dependent on the plugin type.
func createPipelineCmd(p *gaia.Pipeline) *exec.Cmd {
	if p == nil {
		return nil
	}
	c := &exec.Cmd{}

	// Dependent on the pipeline type
	switch p.Type {
	case gaia.PTypeGolang:
		c.Path = p.ExecPath
	case gaia.PTypeJava:
		// Look for java executable
		path, err := exec.LookPath(javaExecName)
		if err != nil {
			gaia.Cfg.Logger.Error("cannot find java executable", "error", err.Error())
			return nil
		}

		// Build start command
		c.Path = path
		c.Args = []string{
			path,
			"-jar",
			p.ExecPath,
		}
	case gaia.PTypePython:
		// Build start command
		c.Path = "/bin/sh"
		c.Args = []string{
			"/bin/sh",
			"-c",
			". bin/activate; exec " + pythonExecName + " -c \"import pipeline; pipeline.main()\"",
		}
		c.Dir = filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpPythonFolder, p.Name)
	case gaia.PTypeCpp:
		c.Path = p.ExecPath
	case gaia.PTypeRuby:
		// Look for ruby executable
		path, err := exec.LookPath(rubyExecName)
		if err != nil {
			gaia.Cfg.Logger.Error("cannot find ruby executable", "error", err.Error())
			return nil
		}

		// Get the gem name from the gem file.
		gemName, err := findRubyGemName(p.ExecPath)
		if err != nil {
			gaia.Cfg.Logger.Error("cannot find the gem name from the gem file", "error", err.Error())
			return nil
		}

		// Build start command
		c.Path = path
		c.Args = []string{
			path,
			"-r",
			gemName,
			"-e",
			"Main.main",
		}
	case gaia.PTypeNodeJS:
		// Look for node executable
		path, err := exec.LookPath(nodeJSExecName)
		if err != nil {
			gaia.Cfg.Logger.Error("cannot find NodeJS executable", "error", err)
			return nil
		}

		// Define the folder where the nodejs plugin is located unpacked
		unpackedFolder := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpNodeJSFolder, p.Name)

		// Build start command
		c.Path = path
		c.Args = []string{
			path,
			"index.js",
		}
		c.Dir = unpackedFolder
	default:
		c = nil
	}

	return c
}

var findRubyGemCommands = []string{"specification", "--yaml"}

// findRubyGemName finds the gem name of a ruby gem file.
func findRubyGemName(execPath string) (name string, err error) {
	// Find the gem binary path.
	path, err := exec.LookPath(rubyGemName)
	if err != nil {
		return
	}

	// Get the gem specification in YAML format.
	gemCommands := append(findRubyGemCommands, execPath)
	cmd := exec.Command(path, gemCommands...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		gaia.Cfg.Logger.Debug("output", "output", string(output[:]))
		return
	}

	// Struct helper to filter for what we need.
	type gemSpecOutput struct {
		Name string
	}

	// Transform and filter the gem specification.
	gemSpec := gemSpecOutput{}
	err = yaml.Unmarshal(output, &gemSpec)
	if err != nil {
		return
	}
	name = gemSpec.Name
	return
}
