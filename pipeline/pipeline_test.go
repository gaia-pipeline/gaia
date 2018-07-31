package pipeline

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
)

var killContext = false
var killOnBuild = false

func fakeExecCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	if killContext {
		c, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()
		ctx = c
	}
	cs := []string{"-test.run=TestExecCommandContextHelper", "--", name}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	arg := strings.Join(cs, ",")
	envArgs := os.Getenv("CMD_ARGS")
	if len(envArgs) != 0 {
		envArgs += ":" + arg
	} else {
		envArgs = arg
	}
	os.Setenv("CMD_ARGS", envArgs)
	return cmd
}

func TestExecCommandContextHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

func TestAppend(t *testing.T) {
	ap := NewActivePipelines()

	p1 := gaia.Pipeline{
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	ret := ap.GetByName("Pipeline A")

	if p1.Name != ret.Name || p1.Type != ret.Type {
		t.Fatalf("Appended pipeline is not present in active pipelines.")
	}

}

func TestUpdate(t *testing.T) {
	ap := NewActivePipelines()

	p1 := gaia.Pipeline{
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := gaia.Pipeline{
		Name:    "Pipeline B",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	ap.Update(0, p2)

	ret := ap.GetByName("Pipeline B")

	if p2.Name != ret.Name {
		t.Fatalf("Pipeline should have been updated.")
	}

}

func TestRemove(t *testing.T) {
	ap := NewActivePipelines()

	p1 := gaia.Pipeline{
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := gaia.Pipeline{
		Name:    "Pipeline B",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p2)

	ap.Remove(1)

	count := 0
	for pipeline := range ap.Iter() {
		count++
		if pipeline.Name == "Pipeline B" {
			t.Fatalf("Pipeline B still exists. It should have been removed.")
		}
	}

	if count != 1 {
		t.Fatalf("Expected pipeline count to be %v. Got %v.", 1, count)
	}
}

func TestGetByName(t *testing.T) {
	ap := NewActivePipelines()

	p1 := gaia.Pipeline{
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	ret := ap.GetByName("Pipeline A")

	if p1.Name != ret.Name || p1.Type != ret.Type {
		t.Fatalf("Pipeline A should have been retrieved.")
	}

	ret = ap.GetByName("Pipeline B")
	if ret != nil {
		t.Fatalf("Pipeline B should not have been retrieved.")
	}
}

func TestReplace(t *testing.T) {
	ap := NewActivePipelines()

	p1 := gaia.Pipeline{
		Name: "Pipeline A",
		Type: gaia.PTypeGolang,
		Repo: gaia.GitRepo{
			URL:       "https://github.com/gaia-pipeline/go-test-example-1",
			LocalDest: "tmp",
		},
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := gaia.Pipeline{
		Name: "Pipeline A",
		Type: gaia.PTypeGolang,
		Repo: gaia.GitRepo{
			URL:       "https://github.com/gaia-pipeline/go-test-example-2",
			LocalDest: "tmp",
		},
		Created: time.Now(),
	}
	ap.Append(p2)

	ret := ap.Replace(p2)
	if !ret {
		t.Fatalf("The pipeline could not be replaced")
	}

	p := ap.GetByName("Pipeline A")
	if p.Repo.URL != "https://github.com/gaia-pipeline/go-test-example-2" {
		t.Fatalf("The pipeline repo URL should have been replaced")
	}
}

func TestReplaceByName(t *testing.T) {
	ap := NewActivePipelines()

	p1 := gaia.Pipeline{
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := gaia.Pipeline{
		Name:    "Pipeline B",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	ap.ReplaceByName("Pipeline A", p2)

	ret := ap.GetByName("Pipeline B")

	if p2.Name != ret.Name {
		t.Fatalf("Pipeline should have been updated.")
	}
}

func TestIter(t *testing.T) {
	ap := NewActivePipelines()

	var pipelineNames = []string{"Pipeline A", "Pipeline B", "Pipeline C"}
	var retrievedNames []string

	for _, n := range pipelineNames {
		p := gaia.Pipeline{
			Name:    n,
			Type:    gaia.PTypeGolang,
			Created: time.Now(),
		}
		ap.Append(p)
	}

	count := 0
	for pipeline := range ap.Iter() {
		count++
		retrievedNames = append(retrievedNames, pipeline.Name)
	}

	if count != len(pipelineNames) {
		t.Fatalf("Expected %d pipelines. Got %d.", len(pipelineNames), count)
	}

	for i := range retrievedNames {
		if pipelineNames[i] != retrievedNames[i] {
			t.Fatalf("The pipeline names do not match")
		}
	}
}

func TestContains(t *testing.T) {
	ap := NewActivePipelines()

	p1 := gaia.Pipeline{
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	ret := ap.Contains("Pipeline A")
	if !ret {
		t.Fatalf("Expected Pipeline A to be present in active pipelines.")
	}
}

func TestRemoveDeletedPipelines(t *testing.T) {
	ap := NewActivePipelines()

	p1 := gaia.Pipeline{
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := gaia.Pipeline{
		Name:    "Pipeline B",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p2)

	p3 := gaia.Pipeline{
		Name:    "Pipeline C",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p3)

	// Let's assume Pipeline B was deleted.
	existingPipelineNames := []string{"Pipeline A", "Pipeline C"}

	ap.RemoveDeletedPipelines(existingPipelineNames)

	for pipeline := range ap.Iter() {
		if pipeline.Name == "Pipeline B" {
			t.Fatalf("Pipeline B still exists. It should have been removed.")
		}
	}

}

func TestRenameBinary(t *testing.T) {
	tmp := os.TempDir()
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.PipelinePath = tmp

	p := gaia.Pipeline{
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	newName := "Pipeline B"

	src := filepath.Join(tmp, appendTypeToName(p.Name, p.Type))
	dst := filepath.Join(tmp, appendTypeToName(newName, p.Type))
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(src)
	defer os.Remove(dst)

	ioutil.WriteFile(src, []byte("testcontent"), 0666)

	err := RenameBinary(p, newName)
	if err != nil {
		t.Fatal("an error occured while renaming the binary: ", err)
	}

	content, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatal("an error occured while reading destination file: ", err)
	}
	if string(content) != "testcontent" {
		t.Fatal("file content does not equal src content. was: ", string(content))
	}

}

func TestDeleteBinary(t *testing.T) {
	tmp := os.TempDir()
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.PipelinePath = tmp

	p := gaia.Pipeline{
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	src := filepath.Join(tmp, appendTypeToName(p.Name, p.Type))
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(src)

	ioutil.WriteFile(src, []byte("testcontent"), 0666)

	err := DeleteBinary(p)
	if err != nil {
		t.Fatal("an error occured while deleting the binary: ", err)
	}

	_, err = os.Stat(src)
	if !os.IsNotExist(err) {
		t.Fatal("the binary file still exists. It should have been deleted")
	}
}

func TestGetExecPath(t *testing.T) {
	tmp := os.TempDir()
	gaia.Cfg = new(gaia.Config)
	gaia.Cfg.PipelinePath = tmp

	p := gaia.Pipeline{
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	expectedPath := filepath.Join(tmp, appendTypeToName(p.Name, p.Type))
	execPath := GetExecPath(p)

	if execPath != expectedPath {
		t.Fatalf("expected execpath to be %s. got %s", expectedPath, execPath)
	}
}
