package options

import "fmt"

// RunMode reflects the mode of a run of osscatimporter/osscatpublisher: read-only, read-write, interactive
type RunMode string

// Constant definitions for RunMode
const (
	RunModeRO          RunMode = "read-only"
	RunModeRW          RunMode = "READ-WRITE"
	RunModeInteractive RunMode = "interactive"
)

// String returns a string representation of this RunMode (suitable for use in log messages)
func (r RunMode) String() string {
	return string(r)
}

// ShortString returns a short string representation of thid RunMode, suitable for use in log file names
func (r RunMode) ShortString() string {
	switch r {
	case RunModeRO:
		return "ro"
	case RunModeRW:
		return "rw"
	case RunModeInteractive:
		return "interactive"
	default:
		panic(fmt.Sprintf(`Unknown RunMode: "%v"`, r))
	}
}
