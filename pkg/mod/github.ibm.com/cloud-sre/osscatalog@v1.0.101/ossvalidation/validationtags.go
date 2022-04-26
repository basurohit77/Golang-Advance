package ossvalidation

import (
	"fmt"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
)

// Tag represents the set of tags used to categorize ValidationIssues
type Tag string

// Possible tags for ValidationIssues
const (
	TagCRN                Tag = "CRN"                // Issues associated with basic registration of entries in the various tools and consistency of CRN service-names
	TagDataMismatch       Tag = "DataMismatch"       // Issues associated with mismatches between data in multiple tools (other than CRN service-names)
	TagDataMissing        Tag = "DataMissing"        // Issues associated with missing data in various tools
	TagSNOverride         Tag = "SNOverride"         // Issues associated with overriding ServiceNow data from a separate ServiceNow csv import file
	TagSNEnrollment       Tag = "SNEnrollment"       // Issues associated with the internal consistency of ServiceNow enrollment data
	TagCatalogVisibility  Tag = "CatalogVisibility"  // Issues associated with the visibility of entries in the Global Catalog
	TagNewProperty        Tag = "NewProperty"        // Issues associated with new properties for EntryTypes and OperationalStatus, not supported in the legacy tools (ServiceNow, ScorecardV1, etc.)
	TagControlOverride    Tag = "ControlOverride"    // Issues associated with overriding entries through the OSSMergeControl.Overrides
	TagStatusPage         Tag = "StatusPage"         // Issues associated with the status page / notifications configuration
	TagConsistency        Tag = "Consistency"        // General issues associated with the consistency of various registration attributes
	TagExpired            Tag = "ExpiredOSSTag"      // Issues related to expired OSS Tags
	TagCatalogComposite   Tag = "CatalogComposite"   // Issues related to the format of Catalog "composite" entries
	TagProductInfo        Tag = "ProductInfo"        // Issues related to product information (part numbers, PID, ClearingHouse, etc.)
	TagDependencies       Tag = "Dependencies"       // Issues related to product dependency information
	TagCatalogConsistency Tag = "CatalogConsistency" // Issues related to the overall registration in Global Catalog (e.g  bad plans, deployments, etc.)
	TagIAM                Tag = "IAM"                // Issues related to the registration of a service in IAM
	TagPriorOSS           Tag = "PriorOSS"           // Issue related to copying data from a PriorOSS record rather than from new sources
	TagTest               Tag = "Test"               // Issue is related to test mode or test entries
)

var allValidTags = make(map[string]Tag)

func init() {
	registerTag(TagCRN)
	registerTag(TagDataMismatch)
	registerTag(TagDataMissing)
	registerTag(TagSNOverride)
	registerTag(TagSNEnrollment)
	registerTag(TagCatalogVisibility)
	registerTag(TagNewProperty)
	registerTag(TagControlOverride)
	registerTag(TagStatusPage)
	registerTag(TagConsistency)
	registerTag(TagExpired)
	registerTag(TagCatalogComposite)
	registerTag(TagProductInfo)
	registerTag(TagDependencies)
	registerTag(TagCatalogConsistency)
	registerTag(TagIAM)
	registerTag(TagPriorOSS)
	registerTag(TagTest)

	for _, ra := range ossrunactions.ListValidRunActions() {
		if ra.Parent() == nil {
			registerTag(RunActionToTag(ra))
		}
	}
}

func registerTag(t Tag) {
	folded := strings.TrimSpace(strings.ToLower(string(t)))
	if _, found := allValidTags[folded]; found {
		panic(fmt.Sprintf("Found duplicate ossvalidation.Tag: %s (%s)", t, folded))
	}
	allValidTags[folded] = t
}

// ParseTag parses a string that represents a OSS ValidationTag and returns the actual Tag
func ParseTag(s string) (t Tag, ok bool) {
	folded := strings.TrimSpace(strings.ToLower(s))
	t, ok = allValidTags[folded]
	return t, ok
}

// AddTag adds a tag to a ValidationIssue
func (v *ValidationIssue) AddTag(t ...Tag) *ValidationIssue {
	// Prevent duplicate tags. This is necessary because AddIssueIgnoreDup() may add the same tags twice to the same issue
	for _, t0 := range t {
		found := false
		for _, t1 := range v.Tags {
			if t1 == t0 {
				found = true
				break
			}
		}
		if !found {
			v.Tags = append(v.Tags, t0)
		}
	}
	return v
}

// TagCRN is a utility function to add the "TagCRN" tag to a ValidationIssue
func (v *ValidationIssue) TagCRN() *ValidationIssue {
	return v.AddTag(TagCRN)
}

// TagDataMismatch is a utility function to add the "TagDataMismatch" tag to a ValidationIssue
func (v *ValidationIssue) TagDataMismatch() *ValidationIssue {
	return v.AddTag(TagDataMismatch)
}

// TagDataMissing is a utility function to add the "TagDataMissing" tag to a ValidationIssue
func (v *ValidationIssue) TagDataMissing() *ValidationIssue {
	return v.AddTag(TagDataMissing)
}

// TagSNOverride is a utility function to add the "TagSNOverride" tag to a ValidationIssue
func (v *ValidationIssue) TagSNOverride() *ValidationIssue {
	return v.AddTag(TagSNOverride)
}

// TagSNEnrollment is a utility function to add the "TagSNEnrollment" tag to a ValidationIssue
func (v *ValidationIssue) TagSNEnrollment() *ValidationIssue {
	return v.AddTag(TagSNEnrollment)
}

// TagCatalogVisibility is a utility function to add the "TagCatalogVisibility" tag to a ValidationIssue
func (v *ValidationIssue) TagCatalogVisibility() *ValidationIssue {
	return v.AddTag(TagCatalogVisibility)
}

// TagNewProperty is a utility function to add the "TagNewProperty" tag to a ValidationIssue
func (v *ValidationIssue) TagNewProperty() *ValidationIssue {
	return v.AddTag(TagNewProperty)
}

// TagControlOverride is a utility function to add the "TagControlOverride" tag to a ValidationIssue
func (v *ValidationIssue) TagControlOverride() *ValidationIssue {
	return v.AddTag(TagControlOverride)
}

// TagStatusPage is a utility function to add the "TagStatusPage" tag to a ValidationIssue
func (v *ValidationIssue) TagStatusPage() *ValidationIssue {
	return v.AddTag(TagStatusPage)
}

// TagConsistency is a utility function to add the "TagConsistency" tag to a ValidationIssue
func (v *ValidationIssue) TagConsistency() *ValidationIssue {
	return v.AddTag(TagConsistency)
}

// TagExpired is a utility function to add the "TagExpired" tag to a ValidationIssue
func (v *ValidationIssue) TagExpired() *ValidationIssue {
	return v.AddTag(TagExpired)
}

// TagCatalogComposite is a utility function to add the "TagCatalogComposite" tag to a ValidationIssue
func (v *ValidationIssue) TagCatalogComposite() *ValidationIssue {
	return v.AddTag(TagCatalogComposite)
}

// TagProductInfo is a utility function to add the "TagProductInfo" tag to a ValidationIssue
func (v *ValidationIssue) TagProductInfo() *ValidationIssue {
	return v.AddTag(TagProductInfo)
}

// TagDependencies is a utility function to add the "TagDependencies" tag to a ValidationIssue
func (v *ValidationIssue) TagDependencies() *ValidationIssue {
	return v.AddTag(TagDependencies)
}

// TagCatalogConsistency is a utility function to add the "TagCatalogConsistency" tag to a ValidationIssue
func (v *ValidationIssue) TagCatalogConsistency() *ValidationIssue {
	return v.AddTag(TagCatalogConsistency)
}

// TagMonitoring is a utility function to add the "TagMonitoring" tag to a ValidationIssue
func (v *ValidationIssue) TagMonitoring() *ValidationIssue {
	return v.AddTag(RunActionToTag(ossrunactions.Monitoring))
}

// TagEnvironments is a utility function to add the "TagEnvironments" tag to a ValidationIssue
func (v *ValidationIssue) TagEnvironments() *ValidationIssue {
	return v.AddTag(RunActionToTag(ossrunactions.Environments))
}

// TagSegTribes is a utility function to add the "Tribes" tag to a ValidationIssue
func (v *ValidationIssue) TagSegTribes() *ValidationIssue {
	return v.AddTag(RunActionToTag(ossrunactions.Tribes))
}

// TagIAM is a utility function to add the "TagIAM" tag to a ValidationIssue
func (v *ValidationIssue) TagIAM() *ValidationIssue {
	return v.AddTag(TagIAM)
}

// TagPriorOSS is a utility function to add the "TagPriorOSS" tag to a ValidationIssue
func (v *ValidationIssue) TagPriorOSS() *ValidationIssue {
	return v.AddTag(TagPriorOSS)
}

// TagTest is a utility function to add the "TagTest" tag to a ValidationIssue
func (v *ValidationIssue) TagTest() *ValidationIssue {
	return v.AddTag(TagTest)
}
