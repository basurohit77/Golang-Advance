package gitnotice

import (
	"log"
	"regexp"
	"strings"
)

const (
	// MsgInvalidCharacters Indicates invalid characters found in the CRN
	MsgInvalidCharacters = "ERROR: %s: CRN has an invalid characters."
	// MsgWrongFormat indicates that the CRN has an unparsable format for some reason. This probably rarely happens.
	MsgWrongFormat = "ERROR: %s: CRN has an invalid format. %s"
	// MsgWrongCRNStart indicates the CRN doesn't start in the expected way
	MsgWrongCRNStart = "ERROR: %s: CRN does not appear correct. Does not start with crn:v1"
	// MsgWrongSegmentCount indicates the CRN didn't contain the expected number of segments
	MsgWrongSegmentCount = "ERROR: %s: CRN does not have enough segments: %d"
)

// CRN is a type for holding CRN values
type CRN string

// String will convert a CRN to a simple string
func (crn CRN) String() string {
	return string(crn)
}

// ToCRN will convert a string to a CRN. It will do some error checking in the process
func ToCRN(input string) (CRN, *Error) {
	bits := strings.Split(input, ":")
	if len(bits) != 10 {
		return "", NewError(nil, MsgWrongSegmentCount, input, len(bits))
	}

	if bits[0] != "crn" || bits[1] != "v1" {
		return "", NewError(nil, MsgWrongCRNStart, input)
	}

	// First validate that the crn meets all the basic character requirements
	matched, err := regexp.MatchString(`^[a-z0-9:-]+$`, input)
	if err != nil {
		return "", NewError(nil, MsgWrongFormat, input, err.Error())
	}
	if !matched {
		log.Println("CRN has invalid characters")
		return "", NewError(nil, MsgInvalidCharacters, input)
	}

	return CRN(input), nil
}
