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

	uuid "github.com/satori/go.uuid"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/hashicorp/go-hclog"
)

func TestPrepareEnvironmentNodeJS(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestPrepareEnvironmentNodeJS")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	b := new(BuildPipelineNodeJS)
	p := new(gaia.CreatePipeline)
	err = b.PrepareEnvironment(p)
	if err != nil {
		t.Fatal("error was not expected when preparing environment: ", err)
	}
	var expectedDest = regexp.MustCompile(`^/.*/tmp/nodejs/src/.*`)
	if !expectedDest.MatchString(p.Pipeline.Repo.LocalDest) {
		t.Fatalf("expected destination is '%s', but was '%s'", expectedDest, p.Pipeline.Repo.LocalDest)
	}
}

func TestPrepareEnvironmentInvalidPathForMkdirNodeJS(t *testing.T) {
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "/notexists"
	b := new(BuildPipelineNodeJS)
	p := new(gaia.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err == nil {
		t.Fatal("error was expected but none occurred")
	}
}

func TestExecuteBuildNodeJS(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	defer func() {
		execCommandContext = exec.CommandContext
	}()
	tmp, err := ioutil.TempDir("", "TestExecuteBuildNodeJS")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	pipelineID := uuid.Must(uuid.NewV4(), nil)
	buildDir := filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, gaia.TmpNodeJSFolder, srcFolder, pipelineID.String())
	if err := os.MkdirAll(buildDir, 0700); err != nil {
		t.Fatal(err)
	}
	b := new(BuildPipelineNodeJS)
	p := new(gaia.CreatePipeline)
	p.Pipeline.UUID = pipelineID.String()
	p.Pipeline.Repo = &gaia.GitRepo{}
	err = b.ExecuteBuild(p)
	if err != nil {
		t.Fatal("error while running executebuild. none was expected")
	}
	expectedBuildArgs := ""
	actualArgs := os.Getenv("CMD_ARGS")
	if !strings.Contains(actualArgs, expectedBuildArgs) {
		t.Fatalf("expected args '%s' actual args '%s'", expectedBuildArgs, actualArgs)
	}
}

func TestExecuteBuildContextTimeoutNodeJS(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	buildKillContext = true
	defer func() {
		execCommandContext = exec.CommandContext
		buildKillContext = false
	}()
	tmp, err := ioutil.TempDir("", "TestExecuteBuildContextTimeoutNodeJS")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineNodeJS)
	p := new(gaia.CreatePipeline)
	err = b.ExecuteBuild(p)
	if err == nil {
		t.Fatal("no error found while expecting error.")
	}
	if err.Error() != "context deadline exceeded" {
		t.Fatal("context deadline should have been exceeded. was instead: ", err)
	}
}

func TestExecuteBuildBinaryNotFoundErrorNodeJS(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestExecuteBuildBinaryNotFoundErrorNodeJS")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
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
	b := new(BuildPipelineNodeJS)
	p := new(gaia.CreatePipeline)
	err = b.ExecuteBuild(p)
	if err == nil {
		t.Fatal("no error found while expecting error.")
	}
	if err.Error() != "exec: \"tar\": executable file not found in $PATH" {
		t.Fatal("the error wasn't what we expected. instead it was: ", err)
	}
}

func TestCopyBinaryNodeJS(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestCopyBinaryNodeJS")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	dest, err := ioutil.TempDir("", "TestCopyBinaryNodeJSDest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dest)
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	gaia.Cfg.PipelinePath = dest
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineNodeJS)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypeNodeJS
	p.Pipeline.Repo = &gaia.GitRepo{LocalDest: tmp}
	src := filepath.Join(tmp, appendTypeToName(p.Pipeline.Name, p.Pipeline.Type))
	dst := filepath.Join(dest, appendTypeToName(p.Pipeline.Name, p.Pipeline.Type))
	if err := ioutil.WriteFile(src, []byte("testcontent"), 0666); err != nil {
		t.Fatal(err)
	}
	err = b.CopyBinary(p)
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

func TestCopyBinarySrcDoesNotExistNodeJS(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestCopyBinarySrcDoesNotExistNodeNS")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineNodeJS)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypeNodeJS
	p.Pipeline.Repo = &gaia.GitRepo{LocalDest: "/noneexistent"}
	err = b.CopyBinary(p)
	if err == nil {
		t.Fatal("error was expected when copying binary but none occurred ")
	}
	if err.Error() != "open /noneexistent/"+appendTypeToName(p.Pipeline.Name, p.Pipeline.Type)+": no such file or directory" {
		t.Fatal("a different error occurred then expected: ", err)
	}
}

type nodeJSMockStorer struct {
	store.GaiaStore
	Error error
}

// PipelinePut is a Mock implementation for pipelines
func (m *nodeJSMockStorer) PipelinePut(p *gaia.Pipeline) error {
	return m.Error
}

func TestSavePipelineNodeJS(t *testing.T) {
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "/tmp"
	gaia.Cfg.PipelinePath = "/tmp/pipelines/"
	// Initialize shared logger
	p := new(gaia.Pipeline)
	p.Name = "main"
	p.Type = gaia.PTypeNodeJS
	b := new(BuildPipelineNodeJS)
	m := new(nodeJSMockStorer)
	services.MockStorageService(m)
	defer services.MockStorageService(nil)
	err := b.SavePipeline(p)
	if err != nil {
		t.Fatal("something went wrong. wasn't supposed to get error: ", err)
	}
	if p.Name != "main" {
		t.Fatal("name of pipeline didn't equal expected 'main'. was instead: ", p.Name)
	}
	if p.Type != gaia.PTypeNodeJS {
		t.Fatal("type of pipeline was not nodejs. instead was: ", p.Type)
	}
}

func TestSavePipelineSaveErrorsNodeJS(t *testing.T) {
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "/tmp"
	gaia.Cfg.PipelinePath = "/tmp/pipelines/"
	// Initialize shared logger
	p := new(gaia.Pipeline)
	p.Name = "main"
	p.Type = gaia.PTypeNodeJS
	b := new(BuildPipelineNodeJS)
	m := new(nodeJSMockStorer)
	m.Error = errors.New("database error")
	services.MockStorageService(m)
	defer services.MockStorageService(nil)
	err := b.SavePipeline(p)
	if err == nil {
		t.Fatal("expected error which did not occur")
	}
	if err.Error() != "database error" {
		t.Fatal("error message was not the expected message. was: ", err.Error())
	}
}
