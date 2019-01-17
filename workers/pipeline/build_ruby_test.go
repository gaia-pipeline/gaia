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
	hclog "github.com/hashicorp/go-hclog"
)

func TestPrepareEnvironmentRuby(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPrepareEnvironmentRuby")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	b := new(BuildPipelineRuby)
	p := new(gaia.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err != nil {
		t.Fatal("error was not expected when preparing environment: ", err)
	}
	var expectedDest = regexp.MustCompile(`^/.*/tmp/ruby/src/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !expectedDest.MatchString(p.Pipeline.Repo.LocalDest) {
		t.Fatalf("expected destination is '%s', but was '%s'", expectedDest, p.Pipeline.Repo.LocalDest)
	}
}

func TestPrepareEnvironmentInvalidPathForMkdirRuby(t *testing.T) {
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "/notexists"
	b := new(BuildPipelineRuby)
	p := new(gaia.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err == nil {
		t.Fatal("error was expected but none occurred")
	}
}

func TestExecuteBuildRuby(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	defer func() {
		execCommandContext = exec.CommandContext
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildRuby")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	b := new(BuildPipelineRuby)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypeRuby
	p.Pipeline.Repo.LocalDest = tmp
	src := filepath.Join(tmp, p.Pipeline.Name+".gemspec")
	f, _ := os.Create(src)
	defer os.RemoveAll(tmp)
	defer f.Close()
	ioutil.WriteFile(src, []byte("testcontent"), 0666)
	dst := filepath.Join(tmp, p.Pipeline.Name+".gem")
	resultFile, _ := os.Create(dst)
	defer resultFile.Close()
	ioutil.WriteFile(dst, []byte("testcontent"), 0666)
	err := b.ExecuteBuild(p)
	if err != nil {
		t.Fatalf("error while running executebuild. none was expected: %s", err.Error())
	}
	expectedBuildArgs := ""
	actualArgs := os.Getenv("CMD_ARGS")
	if !strings.Contains(actualArgs, expectedBuildArgs) {
		t.Fatalf("expected args '%s' actual args '%s'", expectedBuildArgs, actualArgs)
	}
}

func TestExecuteBuildContextTimeoutRuby(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	buildKillContext = true
	defer func() {
		execCommandContext = exec.CommandContext
		buildKillContext = false
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildContextTimeoutRuby")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineRuby)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypeRuby
	p.Pipeline.Repo.LocalDest = tmp
	src := filepath.Join(tmp, p.Pipeline.Name+".gemspec")
	f, _ := os.Create(src)
	defer os.RemoveAll(tmp)
	defer f.Close()
	err := b.ExecuteBuild(p)
	if err == nil {
		t.Fatal("no error found while expecting error.")
	}
	if err.Error() != "context deadline exceeded" {
		t.Fatal("context deadline should have been exceeded. was instead: ", err)
	}
}

func TestExecuteBuildBinaryNotFoundErrorRuby(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildBinaryNotFoundErrorRuby")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	currentPath := os.Getenv("PATH")
	defer func() { os.Setenv("PATH", currentPath) }()
	os.Setenv("PATH", "")
	b := new(BuildPipelineRuby)
	p := new(gaia.CreatePipeline)
	err := b.ExecuteBuild(p)
	if err == nil {
		t.Fatal("no error found while expecting error.")
	}
	if err.Error() != "exec: \"gem\": executable file not found in $PATH" {
		t.Fatal("the error wasn't what we expected. instead it was: ", err)
	}
}

func TestCopyBinaryRuby(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinaryRuby")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineRuby)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypeRuby
	p.Pipeline.Repo.LocalDest = tmp
	src := filepath.Join(tmp, "test.gem")
	dst := appendTypeToName(p.Pipeline.Name, p.Pipeline.Type)
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(dst)
	ioutil.WriteFile(src, []byte("testcontent"), 0666)
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

func TestCopyBinarySrcDoesNotExistRuby(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinarySrcDoesNotExistRuby")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineRuby)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypeRuby
	p.Pipeline.Repo.LocalDest = "/noneexistent"
	err := b.CopyBinary(p)
	if err == nil {
		t.Fatal("error was expected when copying binary but none occurred ")
	}
	if err.Error() != "open /noneexistent: no such file or directory" {
		t.Fatal("a different error occurred then expected: ", err)
	}
}

type rubyMockStorer struct {
	store.GaiaStore
	Error error
}

// PipelinePut is a Mock implementation for pipelines
func (m *rubyMockStorer) PipelinePut(p *gaia.Pipeline) error {
	return m.Error
}

func TestSavePipelineRuby(t *testing.T) {
	defer os.Remove("gaia.db")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "/tmp"
	gaia.Cfg.PipelinePath = "/tmp/pipelines/"
	// Initialize shared logger
	p := new(gaia.Pipeline)
	p.Name = "main"
	p.Type = gaia.PTypeRuby
	b := new(BuildPipelineRuby)
	m := new(rubyMockStorer)
	services.MockStorageService(m)
	err := b.SavePipeline(p)
	if err != nil {
		t.Fatal("something went wrong. wasn't supposed to get error: ", err)
	}
	if p.Name != "main" {
		t.Fatal("name of pipeline didn't equal expected 'main'. was instead: ", p.Name)
	}
	if p.Type != gaia.PTypeRuby {
		t.Fatal("type of pipeline was not ruby. instead was: ", p.Type)
	}
}

func TestSavePipelineSaveErrorsRuby(t *testing.T) {
	defer os.Remove("gaia.db")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "/tmp"
	gaia.Cfg.PipelinePath = "/tmp/pipelines/"
	// Initialize shared logger
	p := new(gaia.Pipeline)
	p.Name = "main"
	p.Type = gaia.PTypeCpp
	b := new(BuildPipelineRuby)
	m := new(rubyMockStorer)
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
