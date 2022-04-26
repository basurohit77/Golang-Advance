package testutils

import (
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-rest-test/lg"
)

// CheckTime is used as a basic check of a timestamp
func CheckTime(fct, label, value string) {
	METHOD := fct + "->CheckTime"

	if value == "" {
		lg.Err(METHOD, nil, "No %s (CheckTime)", label)
	}

	_, err := db.VerifyAndConvertTimestamp(value)
	if err != nil {
		lg.Err(METHOD, err, "Bad time format received %s", value)
	}

}

// CheckNoBlankValue will ensure a value exists for a string value
func CheckNoBlankValue(fct, label, value string) {
	METHOD := fct + "->CheckNoBlankValue"
	if value == "" {
		lg.Err(METHOD, nil, "No value for %s", label)
	}
}

// CheckEnum will check against enum values
func CheckEnum(fct, label, value string, enumVals ...string) {
	METHOD := fct + "->CheckEnum"

	for _, e := range enumVals {
		if value == e {
			return
		}
	}

	lg.Err(METHOD, nil, "Invalid value for %s. Value is %s", label, value)
}
