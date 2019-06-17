package stringhelper

import (
	"sort"
	"strings"
)

// IsContainedInSlice checks if the given string is contained in the slice.
// If case insensitive is enabled, checks are case insensitive.
func IsContainedInSlice(s []string, c string, caseIns bool) bool {
	for _, curr := range s {
		if caseIns && strings.EqualFold(curr, c) {
			return true
		} else if curr == c {
			return true
		}
	}
	return false
}

// DiffSlices returns A - B set difference of the two given slices.
// It also sorts the returned slice.
func DiffSlices(a []string, b []string, caseIns bool) []string {
	m := map[string]bool{}
	for _, aItem := range a {
		if caseIns {
			aItem = strings.ToLower(aItem)
		}

		m[aItem] = true
	}

	// Check if equal
	for _, bItem := range b {
		if _, ok := m[bItem]; ok {
			m[bItem] = false
		}
	}

	var items []string
	for item, exists := range m {
		if exists {
			items = append(items, item)
		}
	}
	sort.Strings(items)
	return items
}