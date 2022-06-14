package proto

import "sort"

// sortedStrings provides a copy of the given input string, in which
// the elements have been sorted.
//
// It aims to reduce stutter in some of the tests in this package.
func sortedStrings(in []string) []string {
	cp := make([]string, len(in))
	copy(cp, in)
	sort.Strings(cp)
	return cp
}
