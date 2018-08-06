package pipeline

import (
	"strings"
	"sync"

	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/gaia-pipeline/gaia"
	git "gopkg.in/src-d/go-git.v4"
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
	// Create new endpoint
	ep, err := transport.NewEndpoint(repo.URL)
	if err != nil {
		return err
	}

	// Attach credentials if provided
	auth, err := getAuthInfo(repo)
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
		return err
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
	gaia.Cfg.Logger.Debug("checking pipeline: ", pipe.Name)
	gaia.Cfg.Logger.Debug("selected branch : ", pipe.Repo.SelectedBranch)
	auth, err := getAuthInfo(&pipe.Repo)
	if err != nil {
		// It's also an error if the repo is already up to date so we just move on.
		gaia.Cfg.Logger.Error("error getting auth info while doing a pull request : ", err.Error())
		return err
	}
	tree, _ := r.Worktree()
	err = tree.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	if err != nil {
		// It's also an error if the repo is already up to date so we just move on.
		gaia.Cfg.Logger.Error("error while doing a pull request : ", err.Error())
		return err
	}

	gaia.Cfg.Logger.Debug("updating pipeline: ", pipe.Name)
	b := newBuildPipeline(pipe.Type)
	createPipeline := &gaia.CreatePipeline{}
	createPipeline.Pipeline = *pipe
	b.ExecuteBuild(createPipeline)
	b.CopyBinary(createPipeline)
	gaia.Cfg.Logger.Debug("successfully updated: ", pipe.Name)
	return nil
}

// gitCloneRepo clones the given repo to a local folder.
// The destination will be attached to the given repo obj.
func gitCloneRepo(repo *gaia.GitRepo) error {
	// Check if credentials were provided
	auth, err := getAuthInfo(repo)
	if err != nil {
		return err
	}

	// Clone repo
	_, err = git.PlainClone(repo.LocalDest, false, &git.CloneOptions{
		Auth:              auth,
		URL:               repo.URL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		SingleBranch:      true,
		ReferenceName:     plumbing.ReferenceName(repo.SelectedBranch),
	})
	if err != nil {
		return err
	}

	return nil
}

func updateAllCurrentPipelines() {
	gaia.Cfg.Logger.Debug("starting updating of pipelines...")
	var allPipelines []gaia.Pipeline
	var wg sync.WaitGroup
	sem := make(chan int, 4)
	for pipeline := range GlobalActivePipelines.Iter() {
		allPipelines = append(allPipelines, pipeline)
	}
	for _, p := range allPipelines {
		wg.Add(1)
		go func(pipe gaia.Pipeline) {
			defer wg.Done()
			sem <- 1
			defer func() { <-sem }()
			UpdateRepository(&pipe)
		}(p)
	}
	wg.Wait()
}

func getAuthInfo(repo *gaia.GitRepo) (transport.AuthMethod, error) {
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
	}
	return auth, nil
}
