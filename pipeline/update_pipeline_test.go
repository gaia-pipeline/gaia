package pipeline

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
	hclog "github.com/hashicorp/go-hclog"
)

func TestUpdatePipelinePython(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestUpdatePipelinePython")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})

	p1 := gaia.Pipeline{
		Name:    "PipelinA",
		Type:    gaia.PTypePython,
		Created: time.Now(),
	}

	// Create fake virtualenv folder with temp file
	virtualEnvPath := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpPythonFolder, p1.Name)
	err := os.MkdirAll(virtualEnvPath, 0700)
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(tmp, "PipelineA_python")
	p1.ExecPath = src
	defer os.RemoveAll(tmp)
	ioutil.WriteFile(src, []byte("testcontent"), 0666)

	// fake execution commands
	virtualEnvName = "mkdir"
	pythonPipInstallCmd = "echo %s"

	// run
	err = updatePipeline(&p1)
	if err != nil {
		t.Fatal(err)
	}

	// check if file has been copied to correct place
	if _, err = os.Stat(filepath.Join(virtualEnvPath, p1.Name+".tar.gz")); err != nil {
		t.Fatalf("distribution file does not exist: %s", err.Error())
	}
}
