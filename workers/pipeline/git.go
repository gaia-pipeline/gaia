package pipeline

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	gohttp "net/http"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/google/go-github/github"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/client"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	refHead = "refs/heads"
)

// GitLSRemote get remote branches from a git repo
// without actually cloning the repo. This is great
// for looking if we have access to this repo.
func GitLSRemote(repo *gaia.GitRepo) error {
	// Validate provided git url
	if strings.Contains(repo.URL, "@") {
		return errors.New("git url should not include username and/or password")
	}

	// Create new endpoint
	ep, err := transport.NewEndpoint(repo.URL)
	if err != nil {
		return err
	}

	// Attach credentials if provided
	auth, err := getAuthInfo(repo, nil)
	if err != nil {
		return err
	}

	// Create client
	cl, err := client.NewClient(ep)
	if err != nil {
		return err
	}

	// Open new session
	s, err := cl.NewUploadPackSession(ep, auth)
	if err != nil {
		if strings.Contains(err.Error(), "knownhosts: key is unknown") {
			gaia.Cfg.Logger.Warn("Warning: Unknown host key.", "error", err.Error(), "URL", repo.URL)
			auth, err := getAuthInfo(repo, gossh.InsecureIgnoreHostKey())
			if err != nil {
				return err
			}
			s, err = cl.NewUploadPackSession(ep, auth)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer s.Close()

	// Get advertised references (e.g. branches)
	ar, err := s.AdvertisedReferences()
	if err != nil {
		return err
	}

	// Iterate all references
	repo.Branches = []string{}
	for ref := range ar.References {
		// filter for head refs which is a branch
		if strings.Contains(ref, refHead) {
			repo.Branches = append(repo.Branches, ref)
		}
	}

	return nil
}

// UpdateRepository takes a git type repository and updates
// it by pulling in new code if it's available.
func UpdateRepository(pipe *gaia.Pipeline) error {
	r, err := git.PlainOpen(pipe.Repo.LocalDest)
	if err != nil {
		// We don't stop gaia working because of an automated update failed.
		// So we just move on.
		gaia.Cfg.Logger.Error("error while opening repo: ", pipe.Repo.LocalDest, err.Error())
		return err
	}
	gaia.Cfg.Logger.Debug("checking pipeline: ", "message", pipe.Name)
	gaia.Cfg.Logger.Debug("selected branch: ", "message", pipe.Repo.SelectedBranch)
	auth, err := getAuthInfo(pipe.Repo, nil)
	if err != nil {
		// It's also an error if the repo is already up to date so we just move on.
		gaia.Cfg.Logger.Error("error getting auth info while doing a pull request: ", "error", err.Error())
		return err
	}
	tree, _ := r.Worktree()
	o := &git.PullOptions{
		ReferenceName: plumbing.ReferenceName(pipe.Repo.SelectedBranch),
		SingleBranch:  true,
		RemoteName:    "origin",
		Auth:          auth,
	}
	err = tree.Pull(o)
	if err != nil {
		if strings.Contains(err.Error(), "knownhosts: key is unknown") {
			gaia.Cfg.Logger.Warn("Warning: Unknown host key.", "error", err.Error(), "host", "URL", pipe.Repo.URL)
			auth, err = getAuthInfo(pipe.Repo, gossh.InsecureIgnoreHostKey())
			if err != nil {
				return err
			}
			o.Auth = auth
			err = tree.Pull(o)
			if err != nil {
				return err
			}
		} else {
			// It's also an error if the repo is already up to date so we just move on.
			gaia.Cfg.Logger.Error("error while doing a pull request: ", "error", err.Error())
			return err
		}
	}

	gaia.Cfg.Logger.Debug("updating pipeline: ", "message", pipe.Name)
	b := newBuildPipeline(pipe.Type)
	createPipeline := &gaia.CreatePipeline{
		Pipeline: gaia.Pipeline{
			Repo: &gaia.GitRepo{},
		},
	}
	createPipeline.Pipeline = *pipe
	_ = b.ExecuteBuild(createPipeline)
	_ = b.CopyBinary(createPipeline)
	gaia.Cfg.Logger.Debug("successfully updated: ", "message", pipe.Name)
	return nil
}

// gitCloneRepo clones the given repo to a local folder.
// The destination will be attached to the given repo obj.
func gitCloneRepo(repo *gaia.GitRepo) error {
	// Check if credentials were provided
	auth, err := getAuthInfo(repo, nil)
	if err != nil {
		return err
	}
	o := &git.CloneOptions{
		Auth:              auth,
		URL:               repo.URL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		SingleBranch:      true,
		ReferenceName:     plumbing.ReferenceName(repo.SelectedBranch),
	}
	// Clone repo
	_, err = git.PlainClone(repo.LocalDest, false, o)
	if err != nil {
		if strings.Contains(err.Error(), "knownhosts: key is unknown") {
			gaia.Cfg.Logger.Warn("Warning: Unknown host key.", "error", err.Error(), "URL", repo.URL)
			auth, err = getAuthInfo(repo, gossh.InsecureIgnoreHostKey())
			if err != nil {
				return err
			}
			o.Auth = auth
			// Clone repo again with no host key verification.
			_, err = git.PlainClone(repo.LocalDest, false, o)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func updateAllCurrentPipelines() {
	gaia.Cfg.Logger.Debug("starting updating of pipelines...")
	allPipelines := GlobalActivePipelines.GetAll()
	var wg sync.WaitGroup
	sem := make(chan int, 4)
	for _, p := range allPipelines {
		wg.Add(1)
		go func(pipe gaia.Pipeline) {
			defer wg.Done()
			sem <- 1
			defer func() { <-sem }()
			_ = UpdateRepository(&pipe)
		}(p)
	}
	wg.Wait()
}

// GithubRepoService is an interface defining the Wrapper Interface
// needed to test the github client.
type GithubRepoService interface {
	CreateHook(ctx context.Context, owner, repo string, hook *github.Hook) (*github.Hook, *github.Response, error)
}

// GithubClient is a client that has the ability to replace the actual
// git client.
type GithubClient struct {
	Repositories GithubRepoService
	*github.Client
}

// NewGithubClient creates a wrapper around the github client. This is
// needed in order to decouple gaia from github client to be
// able to unit test createGithubWebhook and ultimately have
// the ability to replace github with anything else.
func NewGithubClient(httpClient *gohttp.Client, repoMock GithubRepoService) GithubClient {
	if repoMock != nil {
		return GithubClient{
			Repositories: repoMock,
		}
	}
	githubClient := github.NewClient(httpClient)

	return GithubClient{
		Repositories: githubClient.Repositories,
	}
}

func createGithubWebhook(token string, repo *gaia.GitRepo, gitRepo GithubRepoService) error {
	vault, err := services.VaultService(nil)
	if err != nil {
		gaia.Cfg.Logger.Error("unable to initialize vault: ", "error", err.Error())
		return err
	}

	err = vault.LoadSecrets()
	if err != nil {
		gaia.Cfg.Logger.Error("unable to open vault: ", "error", err.Error())
		return err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	config := make(map[string]interface{})
	config["url"] = gaia.Cfg.Hostname + "/api/" + gaia.APIVersion + "/pipeline/githook"
	secret, err := vault.Get("GITHUB_WEBHOOK_SECRET")
	if err != nil {
		secret = []byte(generateWebhookSecret())
		vault.Add("GITHUB_WEBHOOK_SECRET", secret)
		err = vault.SaveSecrets()
		if err != nil {
			return err
		}
	}
	config["secret"] = string(secret)
	config["content_type"] = "json"

	githubClient := NewGithubClient(tc, gitRepo)
	repoName := path.Base(repo.URL)
	repoName = strings.TrimSuffix(repoName, ".git")
	// var repoLocation string
	re := regexp.MustCompile("^(https|git)(://|@)([^/:]+)[/:]([^/:]+)/(.+)$")
	m := re.FindAllStringSubmatch(repo.URL, -1)
	if m == nil {
		return errors.New("failed to extract url parameters from git url")
	}
	repoUser := m[0][4]
	hook, resp, err := githubClient.Repositories.CreateHook(context.Background(), repoUser, repoName, &github.Hook{
		Events: []string{"push"},
		Name:   github.String("web"),
		Active: github.Bool(true),
		Config: config,
	})
	if err != nil {
		gaia.Cfg.Logger.Error("error while trying to create webhook: ", "error", err.Error())
		return err
	}
	gaia.Cfg.Logger.Info("hook created: ", github.Stringify(hook.Name), resp.Status)
	gaia.Cfg.Logger.Info("hook url: ", "url", hook.GetURL())
	return nil
}

func generateWebhookSecret() string {
	secret := make([]byte, 16)
	_, _ = rand.Read(secret)
	based := base64.URLEncoding.EncodeToString(secret)
	return strings.TrimSuffix(based, "==")
}

func getAuthInfo(repo *gaia.GitRepo, callBack gossh.HostKeyCallback) (transport.AuthMethod, error) {
	var auth transport.AuthMethod
	if repo.Username != "" && repo.Password != "" {
		// Basic auth provided
		auth = &http.BasicAuth{
			Username: repo.Username,
			Password: repo.Password,
		}
	} else if repo.PrivateKey.Key != "" {
		var err error
		auth, err = ssh.NewPublicKeys(repo.PrivateKey.Username, []byte(repo.PrivateKey.Key), repo.PrivateKey.Password)
		if err != nil {
			return nil, err
		}

		if callBack == nil {
			callBack, err = ssh.NewKnownHostsCallback()
			if err != nil {
				return nil, err
			}
		}
		auth.(*ssh.PublicKeys).HostKeyCallback = callBack
	}
	return auth, nil
}
