package options

import (
	"os"
)

// InputFileFlag represents a flag for the "flags" package, that contains a file name that is valid for input
type InputFileFlag string

// String returns a string representation of the file name value of this InputFileFlag
func (f *InputFileFlag) String() string {
	return string(*f)
}

// Set specifies the file name value for this InputFileFlag and check that it is a valid readable file
func (f *InputFileFlag) Set(v string) error {
	fd, err := os.Open(v) // #nosec G304
	fd.Close()
	if err != nil {
		return err
	}
	*f = InputFileFlag(v)
	return nil
}
