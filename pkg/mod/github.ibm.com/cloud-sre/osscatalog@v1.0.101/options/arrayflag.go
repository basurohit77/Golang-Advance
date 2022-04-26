package options

import "strings"

// ArrayFlag represents a flag for the "flags" package, that contains a list of string values
type ArrayFlag []string

// String returns a string representation of all the individual flag values contained in this ArrayFlag
func (af *ArrayFlag) String() string {
	return strings.Join(*af, ",")
}

// Set adds one or more flag values into this ArrayFlag (provided as a comma-separated list)
func (af *ArrayFlag) Set(v string) error {
	av := strings.Split(v, ",")
	for _, v0 := range av {
		*af = append(*af, strings.TrimSpace(v0))
	}
	return nil
}
