package debug

import (
	"fmt"
	"regexp"
	"strconv"
)

// Utlity functions for error recovery

var dupNamePattern = regexp.MustCompile("^(.*?) \\(([0-99]+)\\)$")

// MakeDuplicateName takes a name and appends "(1)" or "(2)" etc. to distinguish it from potential duplicate names
func MakeDuplicateName(s string) string {
	m := dupNamePattern.FindStringSubmatch(s)
	if m != nil {
		i, err := strconv.Atoi(m[2])
		if err == nil {
			return fmt.Sprintf("%s (%d)", m[1], i+1)
		}
	}
	return s + " (1)"
}

// CompareDuplicateNames compares two names that might differ only in a "(1)", "(2)" etc. suffix
func CompareDuplicateNames(s1, s2 string) bool {
	m1 := dupNamePattern.FindStringSubmatch(s1)
	if m1 != nil {
		s1 = m1[1]
	}

	m2 := dupNamePattern.FindStringSubmatch(s2)
	if m2 != nil {
		s2 = m2[1]
	}

	return s1 == s2
}
