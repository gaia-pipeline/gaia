package stringhelper

import "strings"

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
