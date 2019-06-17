package stringhelper

import "testing"

func TestIsContainedInSlice(t *testing.T) {
	slice := []string{"ITEM1", "item1", "ITEM2"}

	// Simple contains check
	if !IsContainedInSlice(slice, "item1", false) {
		t.Fatal("item1 not contained in slice")
	}

	// Check without case insensitive
	if IsContainedInSlice(slice, "Item2", false) {
		t.Fatal("Item2 is contained but should not")
	}

	// Check with case insensitive
	if !IsContainedInSlice(slice, "item2", true) {
		t.Fatal("item2 case insensitive check failed")
	}

	// Should also work for duplicates
	if !IsContainedInSlice(slice, "item1", true) {
		t.Fatal("item1 not contained")
	}
}

func TestDiffSlices(t *testing.T) {
	a := []string{"ITEM1", "item1", "item3", "item2"}
	b := []string{"item1"}

	// Simple case insensitive check
	out := DiffSlices(a, b, true)
	for _, item := range DiffSlices(a, b, true) {
		if item == "ITEM1" || item == "item1" {
			t.Fatalf("item1 should be non-existend: %s", item)
		}
	}

	// Check if it is sorted
	if len(out) != 2 {
			t.Fatalf("expected 2 but got %d", len(out))
	}
	if out[1] != "item3" {
		t.Fatalf("expected '%s' but got '%s'", "item3", out[1])
	}


	// Multiple different values
	b = append(b, "nonexistend")
	b = append(b, "item2")
	out = DiffSlices(a, b, false)

	// Check
	if len(out) != 2 {
		t.Fatalf("expected 2 but got %d", len(out))
	}
	if out[0] != "ITEM1" {
		t.Fatalf("expected '%s' but got '%s'", "ITEM1", out[0])
	}
	if out[1] != "item3" {
		t.Fatalf("expected '%s' but got '%s'", "item3", out[1])
	}
}
