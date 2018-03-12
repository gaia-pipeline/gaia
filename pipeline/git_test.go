package pipeline

import (
	"os"
	"testing"

	"github.com/gaia-pipeline/gaia"
)

func TestGitCloneRepo(t *testing.T) {
	repo := &gaia.GitRepo{
		URL: "https://github.com/gaia-pipeline/gaia",
	}
	err := gitCloneRepo(repo)
	if err != nil {
		t.Fatal(err)
	}

	// clean up
	err = os.RemoveAll("tmp")
	if err != nil {
		t.Fatal(err)
	}
}
