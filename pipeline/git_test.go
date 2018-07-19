package pipeline

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gaia-pipeline/gaia"
	hclog "github.com/hashicorp/go-hclog"
)

func TestGitCloneRepo(t *testing.T) {
	repo := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/go-test-example",
		LocalDest: "tmp",
	}
	// always ensure that tmp folder is cleaned up
	defer os.RemoveAll("tmp")
	err := gitCloneRepo(repo)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateAllPipelinesRepositoryNotFound(t *testing.T) {
	tmp := os.TempDir()
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	var b strings.Builder
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: &b,
		Name:   "Gaia",
	})

	p := new(gaia.Pipeline)
	p.Repo.LocalDest = tmp
	GlobalActivePipelines = NewActivePipelines()
	GlobalActivePipelines.Append(*p)
	updateAllCurrentPipelines()
	if !strings.Contains(b.String(), "repository does not exist") {
		t.Fatal("error message not found in logs: ", b.String())
	}
}

func TestUpdateAllPipelinesAlreadyUpToDate(t *testing.T) {
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "tmp"
	// Initialize shared logger
	var b strings.Builder
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: &b,
		Name:   "Gaia",
	})
	repo := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/go-test-example",
		LocalDest: "tmp",
	}
	// always ensure that tmp folder is cleaned up
	defer os.RemoveAll("tmp")
	err := gitCloneRepo(repo)
	if err != nil {
		t.Fatal(err)
	}

	p := new(gaia.Pipeline)
	p.Name = "main"
	p.Repo.SelectedBranch = "master"
	p.Repo.LocalDest = "tmp"
	GlobalActivePipelines = NewActivePipelines()
	GlobalActivePipelines.Append(*p)
	updateAllCurrentPipelines()
	if !strings.Contains(b.String(), "already up-to-date") {
		t.Fatal("log output did not contain error message that the repo is up-to-date.: ", b.String())
	}
}

func TestUpdateAllPipelinesAlreadyUpToDateWithMoreThanOnePipeline(t *testing.T) {
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "tmp"
	// Initialize shared logger
	var b strings.Builder
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: &b,
		Name:   "Gaia",
	})
	repo := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/go-test-example",
		LocalDest: "tmp",
	}
	// always ensure that tmp folder is cleaned up
	defer os.RemoveAll("tmp")
	err := gitCloneRepo(repo)
	if err != nil {
		t.Fatal(err)
	}

	p1 := new(gaia.Pipeline)
	p1.Name = "main"
	p1.Repo.SelectedBranch = "master"
	p1.Repo.LocalDest = "tmp"
	p2 := new(gaia.Pipeline)
	p2.Name = "main"
	p2.Repo.SelectedBranch = "master"
	p2.Repo.LocalDest = "tmp"
	GlobalActivePipelines = NewActivePipelines()
	GlobalActivePipelines.Append(*p1)
	GlobalActivePipelines.Append(*p2)
	updateAllCurrentPipelines()
	if !strings.Contains(b.String(), "already up-to-date") {
		t.Fatal("log output did not contain error message that the repo is up-to-date.: ", b.String())
	}
}

func TestUpdateAllPipelinesTenPipelines(t *testing.T) {
	if _, ok := os.LookupEnv("GAIA_RUN_TEN_PIPELINE_TEST"); !ok {
		t.Skip()
	}
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "tmp"
	// Initialize shared logger
	var b strings.Builder
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: &b,
		Name:   "Gaia",
	})
	repo := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/go-test-example",
		LocalDest: "tmp",
	}
	// always ensure that tmp folder is cleaned up
	defer os.RemoveAll("tmp")
	err := gitCloneRepo(repo)
	if err != nil {
		t.Fatal(err)
	}

	GlobalActivePipelines = NewActivePipelines()
	for i := 1; i < 10; i++ {
		p := new(gaia.Pipeline)
		name := strconv.Itoa(i)
		p.Name = "main" + name
		p.Repo.SelectedBranch = "master"
		p.Repo.LocalDest = "tmp"
		GlobalActivePipelines.Append(*p)
	}
	updateAllCurrentPipelines()
	if !strings.Contains(b.String(), "already up-to-date") {
		t.Fatal("log output did not contain error message that the repo is up-to-date.: ", b.String())
	}
}
