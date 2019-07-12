package pipeline

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/hashicorp/go-hclog"
)

func TestPrepareEnvironmentCpp(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPrepareEnvironmentGo")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	b := new(BuildPipelineCpp)
	p := new(gaia.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err != nil {
		t.Fatal("error was not expected when preparing environment: ", err)
	}
	var expectedDest = regexp.MustCompile(`^/.*/tmp/cpp/src/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !expectedDest.MatchString(p.Pipeline.Repo.LocalDest) {
		t.Fatalf("expected destination is '%s', but was '%s'", expectedDest, p.Pipeline.Repo.LocalDest)
	}
}

func TestPrepareEnvironmentInvalidPathForMkdirCpp(t *testing.T) {
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "/notexists"
	b := new(BuildPipelineCpp)
	p := new(gaia.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err == nil {
		t.Fatal("error was expected but none occurred")
	}
}

func TestExecuteBuildCpp(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	defer func() {
		execCommandContext = exec.CommandContext
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildCpp")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	b := new(BuildPipelineCpp)
	p := new(gaia.CreatePipeline)
	err := b.ExecuteBuild(p)
	if err != nil {
		t.Fatal("error while running executebuild. none was expected")
	}
	expectedBuildArgs := ""
	actualArgs := os.Getenv("CMD_ARGS")
	if !strings.Contains(actualArgs, expectedBuildArgs) {
		t.Fatalf("expected args '%s' actual args '%s'", expectedBuildArgs, actualArgs)
	}
}

func TestExecuteBuildContextTimeoutCpp(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	buildKillContext = true
	defer func() {
		execCommandContext = exec.CommandContext
		buildKillContext = false
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildContextTimeoutCpp")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineCpp)
	p := new(gaia.CreatePipeline)
	err := b.ExecuteBuild(p)
	if err == nil {
		t.Fatal("no error found while expecting error.")
	}
	if err.Error() != "context deadline exceeded" {
		t.Fatal("context deadline should have been exceeded. was instead: ", err)
	}
}

func TestExecuteBuildBinaryNotFoundErrorCpp(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildBinaryNotFoundErrorCpp")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	currentPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", currentPath) }()
	_ = os.Setenv("PATH", "")
	b := new(BuildPipelineCpp)
	p := new(gaia.CreatePipeline)
	err := b.ExecuteBuild(p)
	if err == nil {
		t.Fatal("no error found while expecting error.")
	}
	if err.Error() != "exec: \"make\": executable file not found in $PATH" {
		t.Fatal("the error wasn't what we expected. instead it was: ", err)
	}
}

func TestCopyBinaryCpp(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinaryCpp")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineCpp)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypeCpp
	p.Pipeline.Repo = &gaia.GitRepo{LocalDest: tmp}
	src := filepath.Join(tmp, cppFinalBinaryName)
	dst := appendTypeToName(p.Pipeline.Name, p.Pipeline.Type)
	f, _ := os.Create(src)
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

func TestCopyBinarySrcDoesNotExistCpp(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinarySrcDoesNotExistCpp")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineCpp)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypeCpp
	p.Pipeline.Repo = &gaia.GitRepo{LocalDest: "/noneexistent"}
	err := b.CopyBinary(p)
	if err == nil {
		t.Fatal("error was expected when copying binary but none occurred ")
	}
	if err.Error() != "open /noneexistent/"+cppFinalBinaryName+": no such file or directory" {
		t.Fatal("a different error occurred then expected: ", err)
	}
}

type cppMockStorer struct {
	store.GaiaStore
	Error error
}

// PipelinePut is a Mock implementation for pipelines
func (m *cppMockStorer) PipelinePut(p *gaia.Pipeline) error {
	return m.Error
}

func TestSavePipelineCpp(t *testing.T) {
	defer os.Remove("gaia.db")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "/tmp"
	gaia.Cfg.PipelinePath = "/tmp/pipelines/"
	// Initialize shared logger
	p := new(gaia.Pipeline)
	p.Name = "main"
	p.Type = gaia.PTypeCpp
	b := new(BuildPipelineCpp)
	m := new(cppMockStorer)
	services.MockStorageService(m)
	err := b.SavePipeline(p)
	if err != nil {
		t.Fatal("something went wrong. wasn't supposed to get error: ", err)
	}
	if p.Name != "main" {
		t.Fatal("name of pipeline didn't equal expected 'main'. was instead: ", p.Name)
	}
	if p.Type != gaia.PTypeCpp {
		t.Fatal("type of pipeline was not cpp. instead was: ", p.Type)
	}
}

func TestSavePipelineSaveErrorsCpp(t *testing.T) {
	defer os.Remove("gaia.db")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "/tmp"
	gaia.Cfg.PipelinePath = "/tmp/pipelines/"
	// Initialize shared logger
	p := new(gaia.Pipeline)
	p.Name = "main"
	p.Type = gaia.PTypeCpp
	b := new(BuildPipelineCpp)
	m := new(cppMockStorer)
	m.Error = errors.New("database error")
	services.MockStorageService(m)
	err := b.SavePipeline(p)
	if err == nil {
		t.Fatal("expected error which did not occur")
	}
	if err.Error() != "database error" {
		t.Fatal("error message was not the expected message. was: ", err.Error())
	}
}
