package pipeline

import (
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
)

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
