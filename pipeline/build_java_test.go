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

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/store"
	hclog "github.com/hashicorp/go-hclog"
)

func TestPrepareEnvironmentJava(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPrepareEnvironmentJava")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	b := new(BuildPipelineJava)
	p := new(gaia.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err != nil {
		t.Fatal("error was not expected when preparing environment: ", err)
	}
	var expectedDest = regexp.MustCompile(`^/.*/tmp/java/src/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !expectedDest.MatchString(p.Pipeline.Repo.LocalDest) {
		t.Fatalf("expected destination is '%s', but was '%s'", expectedDest, p.Pipeline.Repo.LocalDest)
	}
}

func TestPrepareEnvironmentInvalidPathForMkdirJava(t *testing.T) {
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "/notexists"
	b := new(BuildPipelineJava)
	p := new(gaia.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err == nil {
		t.Fatal("error was expected but none occurred")
	}
}

func TestExecuteBuildJava(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	defer func() {
		execCommandContext = exec.CommandContext
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildJava")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	b := new(BuildPipelineJava)
	p := new(gaia.CreatePipeline)
	// go must be existent, mvn maybe not.
	mavenBinaryName = "go"
	err := b.ExecuteBuild(p)
	if err != nil {
		t.Fatal("error while running executebuild. none was expected")
	}
	expectedBuildArgs := "clean,compile,assembly:single"
	actualArgs := os.Getenv("CMD_ARGS")
	if !strings.Contains(actualArgs, expectedBuildArgs) {
		t.Fatalf("expected args '%s' actual args '%s'", expectedBuildArgs, actualArgs)
	}
}

func TestExecuteBuildContextTimeoutJava(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	buildKillContext = true
	defer func() {
		execCommandContext = exec.CommandContext
		buildKillContext = false
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildContextTimeoutJava")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineJava)
	p := new(gaia.CreatePipeline)
	// go must be existent, mvn maybe not.
	mavenBinaryName = "go"
	err := b.ExecuteBuild(p)
	if err == nil {
		t.Fatal("no error found while expecting error.")
	}
	if err.Error() != "context deadline exceeded" {
		t.Fatal("context deadline should have been exceeded. was instead: ", err)
	}
}

func TestCopyBinaryJava(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinaryJava")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineJava)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypeJava
	p.Pipeline.Repo.LocalDest = tmp
	os.Mkdir(filepath.Join(tmp, mavenTargetFolder), 0744)
	src := filepath.Join(tmp, mavenTargetFolder, javaFinalJarName)
	dst := appendTypeToName(p.Pipeline.Name, p.Pipeline.Type)
	f, _ := os.Create(src)
	defer os.Remove(filepath.Join(tmp, mavenTargetFolder))
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

func TestCopyBinarySrcDoesNotExistJava(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinarySrcDoesNotExistJava")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	b := new(BuildPipelineJava)
	p := new(gaia.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = gaia.PTypeJava
	p.Pipeline.Repo.LocalDest = "/noneexistent"
	err := b.CopyBinary(p)
	if err == nil {
		t.Fatal("error was expected when copying binary but none occurred ")
	}
	if err.Error() != "open /noneexistent/target/plugin-jar-with-dependencies.jar: no such file or directory" {
		t.Fatal("a different error occurred then expected: ", err)
	}
}

type javaMockStorer struct {
	store.GaiaStore
	Error error
}

// PipelinePut is a Mock implementation for pipelines
func (m *javaMockStorer) PipelinePut(p *gaia.Pipeline) error {
	return m.Error
}

func TestSavePipelineJava(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestSavePipelineJava")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	m := new(javaMockStorer)
	services.MockStorageService(m)
	defer os.Remove(tmp)
	defer os.Remove("gaia.db")
	gaia.Cfg.PipelinePath = tmp + "/pipelines/"
	defer os.Remove(gaia.Cfg.PipelinePath)
	p := new(gaia.Pipeline)
	p.Name = "main"
	p.Type = gaia.PTypeJava
	b := new(BuildPipelineJava)
	err := b.SavePipeline(p)
	if err != nil {
		t.Fatal("something went wrong. wasn't supposed to get error: ", err)
	}
	if p.Name != "main" {
		t.Fatal("name of pipeline didn't equal expected 'main'. was instead: ", p.Name)
	}
	if p.Type != gaia.PTypeJava {
		t.Fatal("type of pipeline was not java. instead was: ", p.Type)
	}
}
