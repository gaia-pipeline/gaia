package pipeline

import (
	"strings"

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
	var auth transport.AuthMethod
	if repo.Username != "" && repo.Password != "" {
		// Basic auth provided
		auth = &http.BasicAuth{
			Username: repo.Username,
			Password: repo.Password,
		}
	} else if repo.PrivateKey.Key != "" {
		auth, err = ssh.NewPublicKeys(repo.PrivateKey.Username, []byte(repo.PrivateKey.Key), repo.PrivateKey.Password)
		if err != nil {
			return err
		}
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

// gitCloneRepo clones the given repo to a local folder.
// The destination will be attached to the given repo obj.
func gitCloneRepo(repo *gaia.GitRepo) error {
	// Check if credentials were provided
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
			return err
		}
	}

	// Clone repo
	_, err := git.PlainClone(repo.LocalDest, false, &git.CloneOptions{
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
	// Get all active pipelines
	for pipeline := range GlobalActivePipelines.Iter() {
		r, err := git.PlainOpen(pipeline.Repo.LocalDest)
		if err != nil {
			// ignore for now
			return
		}

		beforPull, _ := r.Head()
		tree, _ := r.Worktree()
		tree.Pull(&git.PullOptions{
			RemoteName: "origin",
		})
		afterPull, _ := r.Head()
		gaia.Cfg.Logger.Debug("no need to update pipeline: ", pipeline.Name)
		// if there are no changes...
		if beforPull.Hash() == afterPull.Hash() {
			continue
		}
		gaia.Cfg.Logger.Debug("updating pipeline: ", pipeline.Name)
		// otherwise build the pipeline
		b := newBuildPipeline(pipeline.Type)
		createPipeline := &gaia.CreatePipeline{}
		createPipeline.Pipeline = pipeline
		b.ExecuteBuild(createPipeline)
		gaia.Cfg.Logger.Debug("successfully updated: ", pipeline.Name)
	}
}
