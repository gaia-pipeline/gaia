package pipeline

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/google/go-github/github"
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

type MockGitVaultStorer struct {
	Error error
}

var gitStore []byte

func (mvs *MockGitVaultStorer) Init() error {
	gitStore = make([]byte, 0)
	return mvs.Error
}

func (mvs *MockGitVaultStorer) Read() ([]byte, error) {
	return gitStore, mvs.Error
}

func (mvs *MockGitVaultStorer) Write(data []byte) error {
	gitStore = data
	return mvs.Error
}

type MockGithubRepositoryService struct {
	Hook     *github.Hook
	Response *github.Response
	Error    error
}

func (mgc *MockGithubRepositoryService) CreateHook(ctx context.Context, owner, repo string, hook *github.Hook) (*github.Hook, *github.Response, error) {
	return mgc.Hook, mgc.Response, mgc.Error
}

func TestCreateGithubWebhook(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreateGithubWebhook")
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.VaultPath = tmp
	gaia.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Gaia",
	})
	_, err := services.CertificateService()
	if err != nil {
		t.Fatalf("cannot initialize certificate service: %v", err.Error())
	}

	m := new(MockGitVaultStorer)
	v, _ := services.VaultService(m)
	v.Add("GITHUB_WEBHOOK_SECRET", []byte("superawesomesecretgithubpassword"))
	defer func() {
		services.MockVaultService(nil)
		services.MockCertificateService(nil)
	}()
	repo := gaia.GitRepo{
		URL: "https://github.com/gaia-pipeline/gaia",
	}

	mock := new(MockGithubRepositoryService)
	mock.Hook = &github.Hook{
		Name: github.String("test hook"),
		URL:  github.String("https://github.com/gaia-pipline/testurl"),
	}
	mock.Response = &github.Response{
		Response: &http.Response{
			Status: "Ok",
		},
	}

	err = createGithubWebhook("asdf", &repo, mock)
	if err != nil {
		t.Fatal("did not expect error to occur. was: ", err)
	}
	body, _ := ioutil.ReadAll(buf)
	expectedStatusMessage := []byte("hook created: : \"test hook\"=Ok")
	expectedHookURL := []byte("https://github.com/gaia-pipline/testurl")
	if !bytes.Contains(body, expectedStatusMessage) {
		t.Fatalf("expected status message not found in logs. want:'%s', got: '%s'", expectedStatusMessage, body)
	}
	if !bytes.Contains(body, expectedHookURL) {
		t.Fatalf("expected hook url not found in logs. want:'%s', got: '%s'", expectedHookURL, body)
	}
}
