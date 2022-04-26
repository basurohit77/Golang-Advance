package ossmergecontrol

import (
	"crypto/sha1" // #nosec G505
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// OSSMergeControl captures various information that controls how a particular
// service/componentry (or group of entries/names) is to be merged.
type OSSMergeControl struct {
	CanonicalName     string                 `json:"canonical_name"`
	OSSTags           osstags.TagSet         `json:"oss_tags"`
	IgnoredOSSTags    osstags.TagSet         `json:"ignored_oss_tags,omitempty"`
	RawDuplicateNames []string               `json:"raw_duplicate_names"`
	DoNotMergeNames   []string               `json:"do_not_merge_names"`
	Notes             string                 `json:"notes"`
	Overrides         map[string]interface{} `json:"overrides"`
	IgnoredOverrides  map[string]interface{} `json:"ignored_overrides,omitempty"`
	LastUpdate        string                 `json:"last_update"`
	UpdatedBy         string                 `json:"updated_by"`
	Checksum          string                 `json:"checksum"`
	mutex             sync.Mutex
}

// New creates a new, empty OSSMergeControl record
func New(name string) *OSSMergeControl {
	result := &OSSMergeControl{CanonicalName: name}
	// New records start with an empty Checksum (will be set when edited e.g. from osscatviewer)
	result.Checksum = ""
	//	result.RefreshChecksum(true)
	return result
}

// RefreshChecksum checks that the Checksum in this OSSMergeControl record is valid and updates it if appropriate
// Returns true if the existing Checksum correctly represents the contents of the record.
func (ossc *OSSMergeControl) RefreshChecksum(update bool) bool {
	ossc.mutex.Lock()
	defer ossc.mutex.Unlock()
	oldCksum := ossc.Checksum
	ossc.Checksum = ""
	data, _ := json.Marshal(ossc)
	sum := sha1.Sum(data) // #nosec G401
	newCksum := hex.EncodeToString(sum[:len(sum)])
	if newCksum == oldCksum {
		ossc.Checksum = oldCksum
		debug.Debug(debug.MergeControl, "OSSMergeControl.RefreshChecksum(%s) -> %v, %v", ossc.CanonicalName, true, ossc.Checksum)
		return true
	}
	if update {
		ossc.Checksum = newCksum
	} else {
		ossc.Checksum = oldCksum
	}

	if oldCksum == "" && ossc.IsEmpty() {
		debug.Debug(debug.MergeControl, "OSSMergeControl.RefreshChecksum(%s) ignoring null checksum in empty record", ossc.CanonicalName)
		return true
	}
	debug.Debug(debug.MergeControl, "OSSMergeControl.RefreshChecksum(%s) -> %v, %v", ossc.CanonicalName, false, ossc.Checksum)
	return false
}

// IsEmpty returns true if this OSSMergeControl object does not contain any non-zero data
func (ossc *OSSMergeControl) IsEmpty() bool {
	if len(ossc.OSSTags) > 0 ||
		len(ossc.IgnoredOSSTags) > 0 ||
		len(ossc.RawDuplicateNames) > 0 ||
		len(ossc.DoNotMergeNames) > 0 ||
		len(ossc.Overrides) > 0 ||
		len(ossc.IgnoredOverrides) > 0 ||
		len(ossc.Notes) > 0 {
		return false
	}
	return true
}

// IsEmptyExceptNotes returns true if this OSSMergeControl object does not contain any non-zero data except for Notes
func (ossc *OSSMergeControl) IsEmptyExceptNotes() bool {
	if len(ossc.OSSTags) > 0 ||
		len(ossc.IgnoredOSSTags) > 0 ||
		len(ossc.RawDuplicateNames) > 0 ||
		len(ossc.DoNotMergeNames) > 0 ||
		len(ossc.Overrides) > 0 ||
		len(ossc.IgnoredOverrides) > 0 {
		return false
	}
	return true
}

// Header produces a short-multiline header summarizing the OSS MergeControl information
// The output is empty if there is no non-zero OSSMergeControl info
func (ossc *OSSMergeControl) Header() string {
	var result strings.Builder
	if len(ossc.OSSTags) > 0 {
		result.WriteString(fmt.Sprintf("-- OSSTags:            %v\n", ossc.OSSTags))
	}
	if len(ossc.RawDuplicateNames) > 0 {
		result.WriteString(fmt.Sprintf("-- Duplicate Names:    %v\n", ossc.RawDuplicateNames))
	}
	if len(ossc.DoNotMergeNames) > 0 {
		result.WriteString(fmt.Sprintf("-- Do-not-merge Names: %v\n", ossc.DoNotMergeNames))
	}
	return result.String()
}

// JSON returns a JSON-style string representation of this OSSMergeControl record
func (ossc *OSSMergeControl) JSON() string {
	var result strings.Builder
	result.WriteString(`"oss_merge_control": `)
	json, _ := json.MarshalIndent(ossc, "    ", "    ")
	_, err := result.Write(json)
	if err != nil {
		panic(err)
	}
	return result.String()
}

// OneLineString returns a single-line string that represents only the non-zero fields in a OSSMergeControl record
// Used for logging
func (ossc *OSSMergeControl) OneLineString() string {
	var output = struct {
		CanonicalName     string                 `json:"canonical_name"`
		OSSTags           osstags.TagSet         `json:"oss_tags,omitempty"`
		RawDuplicateNames []string               `json:"raw_duplicate_names,omitempty"`
		DoNotMergeNames   []string               `json:"do_not_merge_names,omitempty"`
		Notes             string                 `json:"notes,omitempty"`
		Overrides         map[string]interface{} `json:"overrides,omitempty"`
		LastUpdate        string                 `json:"last_update,omitempty"`
		UpdatedBy         string                 `json:"updated_by,omitempty"`
	}{
		CanonicalName:     ossc.CanonicalName,
		OSSTags:           ossc.OSSTags,
		RawDuplicateNames: ossc.RawDuplicateNames,
		DoNotMergeNames:   ossc.DoNotMergeNames,
		Notes:             ossc.Notes,
		Overrides:         ossc.Overrides,
		LastUpdate:        ossc.LastUpdate,
		UpdatedBy:         ossc.UpdatedBy,
	}

	result, err := json.Marshal(&output)
	if err != nil {
		return fmt.Sprintf(`{"error": "%v"}`, err)
	}
	return string(result)
}

var typeString = reflect.TypeOf(string("foo"))
var typeOperationalStatus = reflect.TypeOf(ossrecord.OperationalStatus(ossrecord.GA))
var typeCRNServiceName = reflect.TypeOf(ossrecord.CRNServiceName("foo"))

var validOverrides = map[string]reflect.Type{
	"StatusPage.Group":                    typeString,
	"StatusPage.CategoryID":               typeString,
	"StatusPage.CategoryParent":           typeCRNServiceName,
	"GeneralInfo.ParentResourceName":      typeCRNServiceName,
	"GeneralInfo.FutureOperationalStatus": typeOperationalStatus,
	//	"ParentResourceName":             typeCRNServiceName,
}

// AddOverride adds an entry to the list of Overrides for this OSSMergeControl record.
// It returns an error if this particular Override is not valid or is a duplicate
func (ossc *OSSMergeControl) AddOverride(name, value string) error {
	// TODO: allow arbitrary JSON types for AddOverride instead of only string
	if ossc.Overrides == nil {
		ossc.Overrides = make(map[string]interface{})
	}
	if type1, found1 := validOverrides[name]; found1 {
		if prior, found2 := ossc.Overrides[name]; !found2 {
			debug.Debug(debug.MergeControl, `Replacing Override "%s=%s" (previous value "%s=%s") for entry %s`, name, value, name, prior, ossc.CanonicalName)
		}
		var typedValue interface{}
		// TODO: use reflection to handle arbitratry types for Overrides, instead of a hard-coded list of types
		switch type1 {
		case typeString:
			typedValue = value
		case typeOperationalStatus:
			typedValue = ossrecord.OperationalStatus(value)
		case typeCRNServiceName:
			typedValue = ossrecord.CRNServiceName(value)
		default:
			return fmt.Errorf(`Unsupported type for Override "%s=%s" (type %s) for entry %s`, name, value, type1, ossc.CanonicalName)
		}
		if typedValue != nil {
			ossc.Overrides[name] = typedValue
		} else {
			return fmt.Errorf(`Invalid value type for Override "%s=%s" (expected %s) for entry %s`, name, value, type1, ossc.CanonicalName)
		}
	} else {
		return fmt.Errorf(`Invalid Override "%s=%s" for entry %s`, name, value, ossc.CanonicalName)
	}
	return nil
}

// GetOverride returns the value for a given named Override in this OSSMergeControl record,
// or nil if this value is not overriden.
// It returns an error if the Override name is invalid
func (ossc *OSSMergeControl) GetOverride(name string) (interface{}, error) {
	var type1 reflect.Type
	var found1 bool
	if type1, found1 = validOverrides[name]; !found1 {
		return nil, fmt.Errorf(`Invalid Override name "%s"`, name)
	}
	if ossc.Overrides == nil {
		return nil, nil
	}
	if val, found2 := ossc.Overrides[name]; found2 {
		type2 := reflect.TypeOf(val)
		// TODO: use reflection to handle arbitratry types for Overrides, instead of a hard-coded list of types
		switch {
		case type1 == typeString && type2 == typeString:
			return val.(string), nil
		case type1 == typeOperationalStatus && type2 == typeString:
			return ossrecord.OperationalStatus(val.(string)), nil
		case type1 == typeCRNServiceName && type2 == typeString:
			return ossrecord.CRNServiceName(val.(string)), nil
		default:
			debug.Debug(debug.MergeControl, `GetOverride(): Unsupported type for Override "%s=%s" (expected type %s) for entry %s`, name, val, type1, ossc.CanonicalName)
			return val, nil
		}
	}
	return nil, nil
}
