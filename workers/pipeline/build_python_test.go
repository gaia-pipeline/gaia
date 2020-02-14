package pipeline

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gaia-pipeline/gaia/helper/pipelinehelper"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/store"
	hclog "github.com/hashicorp/go-hclog"
)

func TestPrepareEnvironmentPython(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPrepareEnvironmentPython")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	b := new(BuildPipelinePython)
	p := new(gaia.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err != nil {
		t.Fatal("error was not expected when preparing environment: ", err)
	}
	var expectedDest = regexp.MustCompile(`^/.*/tmp/python/src/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !expectedDest.MatchString(p.Pipeline.Repo.LocalDest) {
		t.Fatalf("expected destination is '%s', but was '%s'", expectedDest, p.Pipeline.Repo.LocalDest)
	}
}

func TestPrepareEnvironmentInvalidPathForMkdirPython(t *testing.T) {
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "/notexists"
	b := new(BuildPipelinePython)
	p := new(gaia.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err == nil {
		t.Fatal("error was expected but none occurred")
	}
}

func TestExecuteBuildPython(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	defer func() {
		execCommandContext = exec.CommandContext
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildPython")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypePython
	p.Pipeline.Repo = &gaia.GitRepo{LocalDest: tmp}
	_ = os.Mkdir(filepath.Join(tmp, "dist"), 0744)
	src := filepath.Join(tmp, "dist", p.Pipeline.Name+".tar.gz")
	f, _ := os.Create(src)
	defer os.RemoveAll(tmp)
	defer f.Close()
	_ = ioutil.WriteFile(src, []byte("testcontent"), 0666)
	b := new(BuildPipelinePython)
	// go must be existent, python maybe not.
	pythonBinaryName = "go"
	err := b.ExecuteBuild(p)
	if err != nil {
		t.Fatalf("error while running executebuild. none was expected but got %s", err.Error())
	}
	expectedBuildArgs := "setup.py,sdist"
	actualArgs := os.Getenv("CMD_ARGS")
	if !strings.Contains(actualArgs, expectedBuildArgs) {
		t.Fatalf("expected args '%s' actual args '%s'", expectedBuildArgs, actualArgs)
	}
}

func TestExecuteBuildContextTimeoutPython(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	buildKillContext = true
	defer func() {
		execCommandContext = exec.CommandContext
		buildKillContext = false
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildContextTimeoutPython")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelinePython)
	p := new(gaia.CreatePipeline)
	// go must be existent, mvn maybe not.
	pythonBinaryName = "go"
	err := b.ExecuteBuild(p)
	if err == nil {
		t.Fatal("no error found while expecting error.")
	}
	if err.Error() != "context deadline exceeded" {
		t.Fatal("context deadline should have been exceeded. was instead: ", err)
	}
}

func TestCopyBinaryPython(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinaryPython")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelinePython)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypePython
	p.Pipeline.Repo = &gaia.GitRepo{LocalDest: tmp}
	_ = os.Mkdir(filepath.Join(tmp, "dist"), 0744)
	src := filepath.Join(tmp, "dist", p.Pipeline.Name+".tar.gz")
	dst := pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type)
	f, _ := os.Create(src)
	defer os.RemoveAll(tmp)
	defer f.Close()
	defer os.Remove(dst)
	_ = ioutil.WriteFile(src, []byte("testcontent"), 0666)
	err := b.CopyBinary(p)
	if err != nil {
		t.Fatal("error was not expected when copying binary: ", err)
	}
	content, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatal("error encountered while reading destination file: ", err)
	}
	if string(content) != "testcontent" {
		t.Fatal("file content did not equal src content. was: ", string(content))
	}
}

func TestCopyBinarySrcDoesNotExistPython(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinarySrcDoesNotExistPython")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelinePython)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypePython
	p.Pipeline.Repo = &gaia.GitRepo{LocalDest: "/noneexistent"}
	err := b.CopyBinary(p)
	if err == nil {
		t.Fatal("error was expected when copying binary but none occurred ")
	}
	if err.Error() != "open /noneexistent/dist: no such file or directory" {
		t.Fatal("a different error occurred then expected: ", err)
	}
}

type pythonMockStorer struct {
	store.GaiaStore
	Error error
}

// PipelinePut is a Mock implementation for pipelines
func (m *pythonMockStorer) PipelinePut(p *gaia.Pipeline) error {
	return m.Error
}

func TestSavePipelinePython(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestSavePipelinePython")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	m := new(pythonMockStorer)
	services.MockStorageService(m)
	defer os.Remove(tmp)
	defer os.Remove("gaia.db")
	gaia.Cfg.PipelinePath = tmp + "/pipelines/"
	defer os.Remove(gaia.Cfg.PipelinePath)
	p := new(gaia.Pipeline)
	p.Name = "main"
	p.Type = gaia.PTypePython
	b := new(BuildPipelinePython)
	err := b.SavePipeline(p)
	if err != nil {
		t.Fatal("something went wrong. wasn't supposed to get error: ", err)
	}
	if p.Name != "main" {
		t.Fatal("name of pipeline didn't equal expected 'main'. was instead: ", p.Name)
	}
	if p.Type != gaia.PTypePython {
		t.Fatal("type of pipeline was not java. instead was: ", p.Type)
	}
}
