package pipeline

import (
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/gaia-pipeline/gaia"
	"github.com/satori/go.uuid"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/client"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	refHead   = "refs/heads"
	tmpFolder = "tmp"
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
		ep.User = repo.Username
		ep.Password = repo.Password
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
	// We have to reset the username and password to
	// prevent go-git setting the credentials in the URL
	// which will not be URL encoded.
	// https://github.com/src-d/go-git/issues/723
	ep.User = ""
	ep.Password = ""
	repo.Password = ""
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

// GitCloneRepo clones the given repo to a local folder.
// The destination will be attached to the given repo obj.
func GitCloneRepo(repo *gaia.GitRepo) error {
	// create uuid for clone folder
	uuid := uuid.Must(uuid.NewV4())

	// Create local temp folder for clone
	folder := tmpFolder + string(os.PathSeparator) + uuid.String()
	err := os.MkdirAll(folder, 0700)
	if err != nil {
		return err
	}

	// Check if credentials were provided
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

	// Clone repo
	_, err = git.PlainClone(folder, false, &git.CloneOptions{
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
