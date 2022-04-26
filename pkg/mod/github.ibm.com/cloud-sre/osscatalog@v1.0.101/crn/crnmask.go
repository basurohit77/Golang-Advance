package crn

import (
	"fmt"
	"strings"

	"github.ibm.com/cloud-sre/oss-globals/consts"
)

// Mask represents one full CRN mask, possibly with some segments left blank
// (e.g. to represent an environment)
type Mask struct {
	CName           string
	CType           string
	ServiceName     string
	Location        string
	Scope           string
	ServiceInstance string
	ResourceType    string
	Resource        string
}

// MaskAll represents one full CRN mask, possibly with some segments left blank
// Adds CRN and Version not included in the old version Mask
type MaskAll struct {
	Scheme          string
	Version         string
	CName           string
	CType           string
	ServiceName     string
	Location        string
	Scope           string
	ServiceInstance string
	ResourceType    string
	Resource        string
}

// String represents a CRN mask encoded as a string
type String string

/*
CRN string format: crn:version:cname:ctype:service-name:location:scope:service-instance:resource-type:resource
             		0   1       2     3     4            5        6     7                8             9
*/

// ToCRNString returns this CRN Mask as a string in standard format (colon-separated)
func (crn Mask) ToCRNString() String {

	return (String(fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s:%s:%s", consts.CrnScheme, consts.CrnVersion, crn.CName, crn.CType, crn.ServiceName, crn.Location, crn.Scope, crn.ServiceInstance, crn.ResourceType, crn.Resource)))
}

// ToCRNString returns this CRN Mask as a string in standard format (colon-separated)
func (crn MaskAll) ToCRNString() String {
	return (String(fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s:%s:%s", crn.Scheme, crn.Version, crn.CName, crn.CType, crn.ServiceName, crn.Location, crn.Scope, crn.ServiceInstance, crn.ResourceType, crn.Resource)))
}

// IsZero returns true if the specified CRN Mask is the zero value
func (crn Mask) IsZero() bool {
	var zero Mask
	if crn == zero {
		return true
	}
	return false
}

// Parse parses a string into a CRN Mask (expecting the standard colon-separated format)
func Parse(input string) (crnMask Mask, err error) {
	comps := strings.Split(strings.ToLower(input), consts.CrnSeparator) //converts CRN to lowercase
	if len(comps) != consts.CrnLen {
		return Mask{}, fmt.Errorf(`ParseCRNMask() expected %d components but got %d  (input="%s")`, consts.CrnLen, len(comps), input)
	}
	if comps[consts.Schema] != consts.CrnScheme || comps[consts.Version] != consts.CrnVersion {
		return Mask{}, fmt.Errorf(`ParseCRNMask() expected the CRN to start with "crn:v1:" but got "%s"`, input)
	}
	ret := Mask{
		CName:           comps[2],
		CType:           comps[3],
		ServiceName:     comps[4],
		Location:        comps[5],
		Scope:           comps[6],
		ServiceInstance: comps[7],
		ResourceType:    comps[8],
		Resource:        comps[9],
	}
	return ret, nil
}

// ParseAll parses a string into a CRN Mask (expecting the standard colon-separated format) will return all CRN Mask values including CRNShema and Version
func ParseAll(input string) (crnMask MaskAll, err error) {
	comps := strings.Split(strings.ToLower(input), consts.CrnSeparator) //converts CRN to lowercase
	if len(comps) != consts.CrnLen {
		return MaskAll{}, fmt.Errorf(`ParseCRNMask() expected %d components but got %d  (input="%s")`, consts.CrnLen, len(comps), input)
	}
	if comps[consts.Schema] != consts.CrnScheme || comps[consts.Version] != consts.CrnVersion {
		return MaskAll{}, fmt.Errorf(`ParseCRNMask() expected the CRN to start with "crn:v1:" but got "%s"`, input)
	}
	ret := MaskAll{
		Scheme:          comps[consts.Schema],
		Version:         comps[consts.Version],
		CName:           comps[consts.Cname],
		CType:           comps[consts.Ctype],
		ServiceName:     comps[consts.ServiceName],
		Location:        comps[consts.Location],
		Scope:           comps[consts.Scope],
		ServiceInstance: comps[consts.ServiceInstance],
		ResourceType:    comps[consts.ResourceType],
		Resource:        comps[consts.Resource],
	}
	return ret, nil
}

// ValidationKey represent one key for a particular aspect of CRN Mask to validate
// in Validate() or ValidateOrPanic()
type ValidationKey int

// Possible values for ValidationKey
const (
	ExpectEnvironment ValidationKey = iota
	ExpectEnvironmentPublicCloud
	ExpectServiceName
	AllowSoftLayerCName
	AllowBlankCName
	AllowOtherSegments
	StripOtherSegments
)

// Validate checks that a given CRN Mask is valid and returns an error if not
// The set of ValidationKeys passed as parameters further specifies what is expected
func (crn Mask) Validate(keys ...ValidationKey) error {
	_, err := crn.Normalize(keys...)
	return err
}

// Normalize normalizes the given CRN Mask into a standard format.
// It returns an error if the input CRN Mask contains invalid segments that cannot be simply normalized.
// The set of ValidationKeys passed as parameters further specifies what is expected
func (crn Mask) Normalize(keys ...ValidationKey) (Mask, error) {
	ret := Mask{}
	var invalid []string
	var expectEnvironment, expectEnvironmentPublicCloud, expectServiceName, allowSoftLayerCName, allowBlankCName, allowOtherSegments, stripOtherSegments bool
	for _, key := range keys {
		switch key {
		case ExpectEnvironment:
			expectEnvironment = true
		case ExpectEnvironmentPublicCloud:
			expectEnvironmentPublicCloud = true
		case ExpectServiceName:
			expectServiceName = true
		case AllowSoftLayerCName:
			allowSoftLayerCName = true
		case AllowBlankCName:
			allowBlankCName = true
		case AllowOtherSegments:
			allowOtherSegments = true
		case StripOtherSegments:
			stripOtherSegments = true
		default:
			panic(fmt.Errorf(`Unknown CRN validation key: %v`, key))
		}
	}

	if expectEnvironment || expectEnvironmentPublicCloud {
		if crn.Location != "" {
			// TODO: check that the Location is valid
			if (crn.CName == "" && crn.CType == "") && allowBlankCName {
				ret.CName = consts.Bluemix
				ret.CType = consts.CrnPublic
				ret.Location = crn.Location
			} else if (crn.CName == consts.SofLayer && crn.CType == consts.CrnPublic) && allowSoftLayerCName {
				ret.CName = consts.Bluemix
				ret.CType = consts.CrnPublic
				ret.Location = crn.Location
			} else if crn.CName == consts.IBMCloud && crn.CType == consts.CrnPublic {
				ret.CName = consts.Bluemix
				ret.CType = consts.CrnPublic
				ret.Location = crn.Location
			} else if crn.CName != "" && crn.CType != "" {
				ret.CName = crn.CName
				ret.CType = crn.CType
				ret.Location = crn.Location
			} else {
				invalid = append(invalid, `All 3 of cname/ctype/location must be set, or none of them`)
			}
		} else {
			invalid = append(invalid, `Missing Environment information (cname/ctype/location)`)
		}
		if expectEnvironmentPublicCloud && !ret.IsIBMPublicCloud() {
			invalid = append(invalid, `Invalid Environment information (cname/ctype/location): expected IBM Public Cloud`)
		}
	} else {
		if crn.CName != "" || crn.CType != "" || crn.Location != "" {
			invalid = append(invalid, `Unexpected Environment (cname/ctype/location) segments`)
		}
	}

	if expectServiceName {
		if crn.ServiceName != "" {
			ret.ServiceName = crn.ServiceName
		} else {
			invalid = append(invalid, `Missing "service-name" segment`)
		}
	} else {
		if crn.ServiceName != "" {
			invalid = append(invalid, `Unexpected "service-name" segment`)
		}
	}

	if crn.Scope != "" && !stripOtherSegments {
		if allowOtherSegments {
			ret.Scope = crn.Scope
		} else {
			invalid = append(invalid, `Unexpected "scope" segment`)
		}
	}
	if crn.ServiceInstance != "" && !stripOtherSegments {
		if allowOtherSegments {
			ret.ServiceInstance = crn.ServiceInstance
		} else {
			invalid = append(invalid, `Unexpected "service-instance"segment`)
		}
	}
	if crn.ResourceType != "" && !stripOtherSegments {
		if allowOtherSegments {
			ret.ResourceType = crn.ResourceType
		} else {
			invalid = append(invalid, `Unexpected "resource-type" segment`)
		}
	}
	if crn.Resource != "" && !stripOtherSegments {
		if allowOtherSegments {
			ret.Resource = crn.Resource
		} else {
			invalid = append(invalid, `Unexpected "resource" segment`)
		}
	}

	if len(invalid) > 0 {
		return ret, fmt.Errorf(`Invalid CRN Mask "%s": %q`, crn.ToCRNString(), invalid)
	}
	return ret, nil
}

// ParseAndNormalize parses a string into a CRN Mask (expecting the standard colon-separated format)
// and then normalizes it (using Normalize())
// It returns a single error if there was anything wrong during parsing or normalization
func ParseAndNormalize(input string, keys ...ValidationKey) (crn Mask, err error) {
	parsed, err := Parse(input)
	if err != nil {
		return Mask{}, err
	}
	normalized, err := parsed.Normalize(keys...)
	return normalized, err
}

// SetPublicCloud adds information to this CRN Mask to specify a IBM Public Cloud region with the given location
// Returns a new CRN Mask with the added information
// Panics if this
func (crn Mask) SetPublicCloud(location string) Mask {
	if crn.CName != "" || crn.CType != "" || crn.Location != "" {
		panic(fmt.Errorf(`Cannot set new CRN Mask as Public Cloud - already contains some of CName/CType/Location (input="%s")`, crn.ToCRNString()))
	}
	if location == "" {
		panic(fmt.Errorf(`Cannot set new CRN Mask as Public Cloud with a null Location`))
		// TODO: check that the Location is valid
	}
	ret := crn
	ret.CName = consts.Bluemix
	ret.CType = consts.CrnPublic
	ret.Location = location
	return ret
}

// IsIBMPublicCloud returns true if this CRN Mask is associated with one of the IBM Public Cloud regions or locations
func (crn Mask) IsIBMPublicCloud() bool {
	if crn.CName == consts.Bluemix && crn.CType == consts.CrnPublic {
		return true
		// TODO: check that the Location is valid
	}
	return false
}

// IsIBMCloudDedicated returns true if this CRN Mask is associated with one of the IBM Cloud Dedicated environments
func (crn Mask) IsIBMCloudDedicated() bool {
	if crn.CType == consts.CrnDedicated {
		return true
		// TODO: check that the CName and Location are valid
	}
	return false
}

// IsIBMCloudLocal returns true if this CRN Mask is associated with one of the IBM Cloud Local environments
func (crn Mask) IsIBMCloudLocal() bool {
	if crn.CType == consts.CrnLocal {
		return true
		// TODO: check that the CName and Location are valid
	}
	return false
}

// IsIBMCloudStaging returns true if this CRN Mask is associated with one of the IBM Cloud Staging environments
func (crn Mask) IsIBMCloudStaging() bool {
	if crn.CType == consts.CrnPublic {
		// TODO: check that the CName and Location are valid
		switch crn.CName {
		case "staging":
			return true
		case "ys0-dallas", "ys0-london", "yf-dallas", "yf-london":
			// Special case for YS0 and YF environemnts - see GHE issue #315
			return true
		default:
			return false
		}
	}
	return false
}

// IsAnyIBMCloud returns true if this CRN Mask is associated with any IBM Cloud environment (Public, Dedicated, Local or Staging)
func (crn Mask) IsAnyIBMCloud() bool {
	if crn.IsIBMPublicCloud() ||
		crn.IsIBMCloudDedicated() ||
		crn.IsIBMCloudLocal() ||
		crn.IsIBMCloudStaging() {
		return true
	}
	return false
}
