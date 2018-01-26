package handlers

import (
	"errors"
	"strings"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/pipeline"
	"github.com/kataras/iris"
)

var (
	errPathLength = errors.New("name of pipeline is empty or one of the path elements length exceeds 50 characters")
)

const (
	pipelinePathSplitChar = "/"
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

	// Check for remote branches
	err := pipeline.GitLSRemote(repo)
	if err != nil {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.WriteString(err.Error())
		return
	}

	// Return branches
	ctx.JSON(repo)
}

// PipelineBuildFromSource clones a given git repo and
// compiles the included source file to a plugin.
func PipelineBuildFromSource(ctx iris.Context) {
	p := &gaia.Pipeline{}
	if err := ctx.ReadJSON(p); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// Clone git repo
	err := pipeline.GitCloneRepo(&p.Repo)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return
	}

	// Start compiling process for given plugin type
	switch p.Type {
	case gaia.GOLANG:
		// Start compile process for golang TODO
	}

	// copy compiled binary to plugins folder
}

// PipelineNameAvailable looks up if the given pipeline name is
// available and valid.
func PipelineNameAvailable(ctx iris.Context) {
	p := &gaia.Pipeline{}
	if err := ctx.ReadJSON(p); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	// The name could contain a path. Split it up
	path := strings.Split(p.Name, pipelinePathSplitChar)

	// Iterate all objects
	for _, p := range path {
		// Length should be correct
		if len(p) < 1 || len(p) > 50 {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.WriteString(errPathLength.Error())
			return
		}

		// TODO check if pipeline name is already in use
	}
}
