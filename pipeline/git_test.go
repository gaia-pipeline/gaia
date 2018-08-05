package pipeline

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gaia-pipeline/gaia"
	hclog "github.com/hashicorp/go-hclog"
)

func TestGitCloneRepo(t *testing.T) {
	repo := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/pipeline-test",
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
	tmp, _ := ioutil.TempDir("", "TestUpdateAllPipelinesRepositoryNotFound")
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = tmp
	// Initialize shared logger
	b := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: b,
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
	b := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: b,
		Name:   "Gaia",
	})
	repo := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/pipeline-test",
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
	b := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: b,
		Name:   "Gaia",
	})
	repo := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/pipeline-test",
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
	defer func() { GlobalActivePipelines = nil }()
	GlobalActivePipelines.Append(*p1)
	GlobalActivePipelines.Append(*p2)
	updateAllCurrentPipelines()
	if !strings.Contains(b.String(), "already up-to-date") {
		t.Fatal("log output did not contain error message that the repo is up-to-date.: ", b.String())
	}
}

func TestUpdateAllPipelinesHundredPipelines(t *testing.T) {
	if _, ok := os.LookupEnv("GAIA_RUN_HUNDRED_PIPELINE_TEST"); !ok {
		t.Skip()
	}
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.HomePath = "tmp"
	// Initialize shared logger
	b := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: b,
		Name:   "Gaia",
	})
	repo := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/pipeline-test",
		LocalDest: "tmp",
	}
	// always ensure that tmp folder is cleaned up
	defer os.RemoveAll("tmp")
	err := gitCloneRepo(repo)
	if err != nil {
		t.Fatal(err)
	}

	GlobalActivePipelines = NewActivePipelines()
	for i := 1; i < 100; i++ {
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

func TestGetAuthInfoWithUsernameAndPassword(t *testing.T) {
	repoWithUsernameAndPassword := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/pipeline-test",
		LocalDest: "tmp",
		Username:  "username",
		Password:  "password",
	}

	auth, _ := getAuthInfo(repoWithUsernameAndPassword)
	if auth == nil {
		t.Fatal("auth should not be nil when username and password is provided")
	}
}

func TestGetAuthInfoWithPrivateKey(t *testing.T) {
	samplePrivateKey := `
-----BEGIN RSA PRIVATE KEY-----
MD8CAQACCQDB9DczYvFuZQIDAQABAgkAtqAKvH9QoQECBQDjAl9BAgUA2rkqJQIE
Xbs5AQIEIzWnmQIFAOEml+E=
-----END RSA PRIVATE KEY-----
`
	repoWithValidPrivateKey := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/pipeline-test",
		LocalDest: "tmp",
		PrivateKey: gaia.PrivateKey{
			Key:      samplePrivateKey,
			Username: "username",
			Password: "password",
		},
	}
	_, err := getAuthInfo(repoWithValidPrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	repoWithInvalidPrivateKey := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/pipeline-test",
		LocalDest: "tmp",
		PrivateKey: gaia.PrivateKey{
			Key:      "random_key",
			Username: "username",
			Password: "password",
		},
	}
	auth, _ := getAuthInfo(repoWithInvalidPrivateKey)
	if auth != nil {
		t.Fatal("auth should be nil for invalid private key")
	}
}

func TestGetAuthInfoEmpty(t *testing.T) {
	repoWithoutAuthInfo := &gaia.GitRepo{
		URL:       "https://github.com/gaia-pipeline/pipeline-test",
		LocalDest: "tmp",
	}
	auth, _ := getAuthInfo(repoWithoutAuthInfo)
	if auth != nil {
		t.Fatal("auth should be nil when no authentication info is provided")
	}
}
