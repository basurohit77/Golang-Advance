package targurl

import (
	"errors"
	"net/url"
	"strings"
)

const (
	tn = `[targeted notification]`
)

var (
	// ErrLongDescEmpty is returned when the long desctioption field of an incident is empty
	ErrLongDescEmpty = errors.New("targurl: Empty long description field")

	// ErrLongDescNoMatch is returned when the constant tn is not part of the long description field
	ErrLongDescNoMatch = errors.New("targurl: Targeted Notification missing from long description field")

	// ErrStringEmpty is returned as part of URL validation when the supplied URL string is empty
	ErrStringEmpty = errors.New("targurl: Supplied URL string is empty")

	// ErrURLInvalid is returned if the supplied URL cannot be parsed syntactically. No attempt
	// is made to validate the URL's fitness to get to its destination.
	ErrURLInvalid = errors.New("targurl: Supplied URL is invalid")
)

// URLFromLongDescription obtains a URL from the long desctiption field of
// an incoming Service Now (SN) incident.
// The URL is entered by SN users manually in markdown format:
// [targeted notification](https://url.to.parse)
func URLFromLongDescription(ld string) (string, error) {
	if ld == "" {
		return "", ErrLongDescEmpty
	}

	// We are looking for the string [targeted notification] in the incoming string
	// return ErrLongDescNoMatch if not found
	if !strings.Contains(ld, tn) {
		return "", ErrLongDescNoMatch
	}

	var targURL, val string
	val = tn + "("
	targURL = Between(ld, val, ")")
	if err := validateURL(targURL); err != nil {
		return "", err
	}
	return targURL, nil
}

// Between gets a substring between the two strings passed
func Between(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}

	// Grab all text begining at position of [targeted notification] and store it in
	// valueb
	valueb := value[posFirst:]
	posLast := strings.Index(valueb, b)
	if posLast == -1 {
		return ""
	}

	// Adjust first and last positions to grab everything in between '[trageted notification]('
	// and closing ')'
	posFirstAdjusted := posFirst + len(a)
	postLastAdjusted := posFirst + posLast
	if posFirstAdjusted >= postLastAdjusted {
		return ""
	}

	return value[posFirstAdjusted:postLastAdjusted]
}

// between gets a substring between the two strings passed keep this not poblic funtion to maintain compatibility with current code
// once new code from pnp-status gets moved this func can be removed ATR 25-Nov-2019
func between(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}

	// Grab all text begining at position of [targeted notification] and store it in
	// valueb
	valueb := value[posFirst:]
	posLast := strings.Index(valueb, b)
	if posLast == -1 {
		return ""
	}

	// Adjust first and last positions to grab everything in between '[trageted notification]('
	// and closing ')'
	posFirstAdjusted := posFirst + len(a)
	postLastAdjusted := posFirst + posLast
	if posFirstAdjusted >= postLastAdjusted {
		return ""
	}

	return value[posFirstAdjusted:postLastAdjusted]
}
func validateURL(s string) error {
	if s == "" {
		return ErrStringEmpty
	}
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return ErrURLInvalid
	}
	return nil
}
