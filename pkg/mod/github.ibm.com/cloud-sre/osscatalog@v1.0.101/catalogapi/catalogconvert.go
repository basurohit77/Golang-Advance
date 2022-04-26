package catalogapi

import (
	"fmt"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// GetEntryType converts the Kind field from a Catalog Resource record into a standard ossrecord.EntryType value
func (r *Resource) GetEntryType() (result ossrecord.EntryType, ok bool) {
	switch r.Kind {
	case "service":
		return ossrecord.SERVICE, true
	case "runtime": // TODO: What to do about "buildpack" type
		return ossrecord.RUNTIME, true
	case "template", "boilerplate":
		return ossrecord.TEMPLATE, true
	case "iaas":
		return ossrecord.IAAS, true
	case "platform_component", "platform_service", "platform_service_environment": // TODO: What to do about "platform_service_environment" type
		return ossrecord.PLATFORMCOMPONENT, true
	case "composite":
		return ossrecord.COMPOSITE, true
	default:
		return "", false
	}
}

// GetOperationalStatus extracts the Operational Status info from the tags in a Catalog Resource and converts it into a standard ossrecord.OperationalStatus value
func (r *Resource) GetOperationalStatus() (ossrecord.OperationalStatus, error) {
	var sourceTags []string
	var isThirdParty bool
	var isCommunity bool
	var isIBM bool
	var isDeprecated bool
	var levelTags []string
	var tmpStatus ossrecord.OperationalStatus

	for _, t := range r.Tags {
		switch t {
		case "ibm_deprecated", "deprecated":
			levelTags = append(levelTags, t)
			tmpStatus = ossrecord.DEPRECATED
			isDeprecated = true
		case "ibm_ga": // XXX "ibm_ga" does not really exist as a tag? -- check just in case
			levelTags = append(levelTags, t)
			tmpStatus = ossrecord.GA
		case "ibm_beta":
			levelTags = append(levelTags, t)
			tmpStatus = ossrecord.BETA
		case "ibm_experimental":
			levelTags = append(levelTags, t)
			tmpStatus = ossrecord.EXPERIMENTAL
		case "ibm_third_party":
			sourceTags = append(sourceTags, t)
			isThirdParty = true
		case "ibm_community", "community":
			sourceTags = append(sourceTags, t)
			isCommunity = true
		case "ibm_created":
			sourceTags = append(sourceTags, t)
			isIBM = true
		}
	}

	if isDeprecated { // "deprecated" overrides all other tags
		if len(levelTags) > 1 {
			return ossrecord.DEPRECATED, fmt.Errorf(`Conflicting Catalog tags for support level: %q -- "deprecated" takes precedence"`, r.Tags /*append(sourceTags, levelTags...)*/)
		}
		return ossrecord.DEPRECATED, nil
	}

	if len(levelTags) > 1 {
		return ossrecord.OperationalStatusUnknown, fmt.Errorf(`Conflicting Catalog tags for support level: %q -- resetting`, r.Tags /*append(sourceTags, levelTags...)*/)
	}

	if tmpStatus == "" {
		tmpStatus = ossrecord.GA // We assume GA if not otherwise specified
	}

	if isIBM {
		if len(sourceTags) > 1 {
			return tmpStatus, fmt.Errorf(`Catalog tags specify both "ibm_created" and other sources: %q -- "ibm_created" takes precedence`, r.Tags /*append(sourceTags, levelTags...)*/)
		}
		return tmpStatus, nil
	} else if isThirdParty {
		if len(sourceTags) > 1 || len(levelTags) > 0 {
			return ossrecord.THIRDPARTY, fmt.Errorf(`Catalog tags specify both "ibm_third_party" and other sources/levels: %q -- "ibm_third_party" takes precedence`, r.Tags /*append(sourceTags, levelTags...)*/)
		}
		return ossrecord.THIRDPARTY, nil
	} else if isCommunity {
		if len(sourceTags) > 1 {
			return ossrecord.OperationalStatusUnknown, fmt.Errorf(`Catalog tags specify "ibm_community" and other unknown sources: %q -- resetting`, r.Tags /*append(sourceTags, levelTags...)*/)
		}
		if len(levelTags) > 0 {
			return ossrecord.COMMUNITY, fmt.Errorf(`Catalog tags specify both "ibm_community" and some levels: %q -- "ibm_community" takes precedence`, r.Tags /*append(sourceTags, levelTags...)*/)
		}
		return ossrecord.COMMUNITY, nil
	} else {
		if len(sourceTags) > 1 {
			return ossrecord.OperationalStatusUnknown, fmt.Errorf(`Catalog tags specify unknown sources: %q`, r.Tags /*append(sourceTags, levelTags...)*/)
		}
		// We assume IBM if not otherwise specified
		return tmpStatus, nil
	}
}

// ParseVisibilityRestrictions parses a string that represents a Catalog visibility restrictions and returns VisibilityRestriction value
func ParseVisibilityRestrictions(input string) (VisibilityRestrictions, error) {
	switch input {
	case string(VisibilityPublic):
		return VisibilityPublic, nil
	case string(VisibilityIBMOnly):
		return VisibilityIBMOnly, nil
	case string(VisibilityPrivate):
		return VisibilityPrivate, nil
	default:
		return "", fmt.Errorf("Cannot parse Catalog visibility restriction: %v", input)
	}
}

// IsPublicVisible returns true if this Catalog entry is publicly visible (even outside IBM),
// based on the various flags and criteria that control visibility
func (r *Resource) IsPublicVisible() bool {
	if r.Active && !r.Disabled && !r.ObjectMetaData.UI.Hidden && r.EffectiveVisibility.Restrictions == string(VisibilityPublic) {
		return true
	}
	return false
}

// IsPublicVisibleInactiveOK returns true if this Catalog entry is publicly visible (even outside IBM),
// based on the various flags and criteria that control visibility.
// This variant of the function does not take the "Active" flag into consideration, i.e. an entry is considered
// Public even if it is inactive (used for Environments)
func (r *Resource) IsPublicVisibleInactiveOK() bool {
	if /* r.Active && */ !r.Disabled && /* !r.ObjectMetaData.UI.Hidden && */ r.EffectiveVisibility.Restrictions == string(VisibilityPublic) {
		return true
	}
	return false
}

// IsPublicVisibleHiddenOK returns true if this Catalog entry is publicly visible (even outside IBM),
// based on the various flags and criteria that control visibility.
// This variant of the function does not take the "UI.Hidden" flag into consideration, i.e. an entry is considered
// Public even if it is hidden in the default UI (used for VMware offerings and others)
func (r *Resource) IsPublicVisibleHiddenOK() bool {
	if r.Active && !r.Disabled && /* !r.ObjectMetaData.UI.Hidden && */ r.EffectiveVisibility.Restrictions == string(VisibilityPublic) {
		return true
	}
	return false
}
