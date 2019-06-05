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
