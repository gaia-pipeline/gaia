package pipeline

import (
	"os"
	"testing"

	"github.com/michelvocks/gaia"
)

func TestGitCloneRepo(t *testing.T) {
	repo := &gaia.GitRepo{
		URL: "https://github.com/michelvocks/gaia",
	}
	err := GitCloneRepo(repo)
	if err != nil {
		t.Fatal(err)
	}

	// clean up
	err = os.RemoveAll("tmp")
	if err != nil {
		t.Fatal(err)
	}
}
