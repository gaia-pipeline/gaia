package pipeline

import (
	"os"
	"testing"

	"github.com/gaia-pipeline/gaia"
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
