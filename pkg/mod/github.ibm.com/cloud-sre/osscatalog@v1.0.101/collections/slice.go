package collections

import (
	"sort"
	"strings"
)

// AppendSliceStringNoDups appends to a slice of strings while suppressing duplicates
// It has the same signature as the built-in append() function,
// except that it returns the number of duplicate values that it suppressed
// to indicate if the new value was a duplicate
// Note: this has the side effect of sorting the slice
func AppendSliceStringNoDups(slice []string, vals ...string) (newVal []string, numDups int) {
	for _, val := range vals {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		var isDup bool
		for _, elem := range slice {
			if elem == val {
				numDups++
				isDup = true
				break
			}
		}
		if !isDup {
			slice = append(slice, val)
		}
	}
	sort.Strings(slice)
	return slice, numDups
}

// CompareSliceString compares two slices of strings and reports on all differences
// Note: this has the side effect of sorting the slice
func CompareSliceString(left, right []string) (leftOnly, rightOnly []string) {
	// Re-sort just to be safe
	sort.Strings(left)
	sort.Strings(right)

	leftOnly = []string{}
	rightOnly = []string{}

	ix := 0
	for _, item := range left {
		for {
			if ix >= len(right) {
				leftOnly = append(leftOnly, item)
				break
			}
			if item == right[ix] {
				// match - advance both slices
				ix++
				break
			} else if item < right[ix] {
				leftOnly = append(leftOnly, item)
				// keep going through items in the left slice
				break
			} else if item > right[ix] {
				rightOnly = append(rightOnly, right[ix])
				// keep going through items in the right slice
				ix++
			}
		}
	}
	for _, item := range right[ix:] {
		rightOnly = append(rightOnly, item)
	}

	return leftOnly, rightOnly
}
