package pipeline

import (
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
)

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
