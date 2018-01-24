package handlers

import (
	"github.com/kataras/iris"
	"github.com/michelvocks/gaia"
	"github.com/michelvocks/gaia/pipeline"
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
