package agent

import (
	"github.com/gaia-pipeline/gaia/helper/stringhelper"
	"testing"
)

func TestFindLocalBinaries(t *testing.T) {
	testTags := []string{"tag1", "tag2", "tag3"}

	// Simple lookup. Go binary should always exist in dev/test environments.
	tags := findLocalBinaries(testTags)

	// Check output
	if len(tags) < 4 {
		t.Fatalf("expected at least 4 tags but got %d", len(tags))
	}
	if len(stringhelper.DiffSlices(append(testTags, "golang"), tags, false)) != 0 {
		t.Fatalf("expected different output: %#v", tags)
	}

	// Negative language tag
	testTags = append(testTags, "-golang")
	tags = findLocalBinaries(testTags)

	// Check output
	if len(tags) < 3 {
		t.Fatalf("expected at least 3 tags but got %d", len(tags))
	}
	if stringhelper.IsContainedInSlice(tags, "golang", false) {
		t.Fatalf("golang should not be included: %#v", tags)
	}
}

