package handlers

import (
	"strings"

	"github.com/kataras/iris"
	"github.com/michelvocks/gaia"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/client"
)

const (
	refHead = "refs/heads"
)

// PipelineGitLSRemote checks for available git remote branches.
// This is the perfect way to check if we have access to a given repo.
func PipelineGitLSRemote(ctx iris.Context) {
	repo := &gaia.GitRepo{}
	if err := ctx.ReadJSON(repo); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// Create new endpoint
	ep, err := transport.NewEndpoint(repo.URL)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// Attach credentials if provided
	if repo.Username != "" && repo.Password != "" {
		ep.User = repo.Username
		ep.Password = repo.Password
	} else if repo.PrivateKey != "" {

	}

	// Create client
	cl, err := client.NewClient(ep)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// Open new session
	s, err := cl.NewUploadPackSession(ep, nil)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
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
	if err == transport.ErrAuthenticationRequired {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.WriteString(err.Error())
		return
	} else if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// Iterate all references
	repo.Branches = []string{}
	for ref := range ar.References {
		// filter for head refs which is a branch
		if strings.Contains(ref, refHead) {
			repo.Branches = append(repo.Branches, ref)
		}
	}

	// Return branches
	ctx.JSON(repo)
}
