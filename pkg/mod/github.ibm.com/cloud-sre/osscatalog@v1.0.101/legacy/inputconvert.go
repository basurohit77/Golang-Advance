package legacy

import (
	"fmt"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

var validEntryTypes = map[string]ossrecord.EntryType{
	"service":     ossrecord.SERVICE,
	"runtime":     ossrecord.RUNTIME,
	"iaas":        ossrecord.IAAS,
	"boilerplate": ossrecord.TEMPLATE,
	"component":   ossrecord.PLATFORMCOMPONENT,
	"GROUP":       ossrecord.EntryType("*GROUP*"), // Dummy entry
}

// ParseEntryType converts a Legacy EntryType string into the common ossrecord.EntryType type
func ParseEntryType(input string) (ossrecord.EntryType, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil
	}
	if result, ok := validEntryTypes[input]; ok {
		return result, nil
	}
	result, err := ossrecord.ParseEntryType(input)
	if result == "" || err != nil {
		debug.PrintError(`legacy.ParseEntryType(): Invalid input value: "%s"`, input)
		return ossrecord.EntryType(fmt.Sprintf("*%s*", input)), nil
	}
	return result, nil
}

var validOperationalStatuses = map[string]ossrecord.OperationalStatus{
	"GA":           ossrecord.GA,
	"Beta":         ossrecord.BETA,
	"Experimental": ossrecord.EXPERIMENTAL,
	"Third-party":  ossrecord.THIRDPARTY,
	//"Community":    ossrecord.COMMUNITY,
	"Deprecated": ossrecord.DEPRECATED,
	//"Retired":      ossrecord.RETIRED,
}

// ParseOperationalStatus converts a Legacy EntryType string into the common ossrecord.EntryType type
func ParseOperationalStatus(input string) (ossrecord.OperationalStatus, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil
	}
	if result, ok := validOperationalStatuses[input]; ok {
		return result, nil
	}
	result, err := ossrecord.ParseOperationalStatus(input)
	if result == "" || err != nil {
		debug.PrintError(`legacy.ParseOperationalStatus(): Invalid input value: "%s"`, input)
		return ossrecord.OperationalStatus(fmt.Sprintf("*%s*", input)), nil
	}
	return result, nil
}
