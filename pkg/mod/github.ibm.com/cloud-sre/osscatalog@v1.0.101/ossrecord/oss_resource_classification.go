package ossrecord

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// OSSResourceClassification contains metadata describing all resource types supported in the OSS registry
type OSSResourceClassification struct {
	SchemaVersion             string                      `json:"schema_version"` // The version of the schema for this OSSEntry
	OSSOnboardingPhase        OSSOnboardingPhase          `json:"oss_onboarding_phase"`
	OSSOnboardingApprover     Person                      `json:"oss_onboarding_approver"`
	OSSOnboardingApprovalDate string                      `json:"oss_onboarding_approval_date"`
	TypeDefinitions           []OSSResourceTypeDefinition `json:"type_definitions"`
}

// OSSResourceTypeDefinition represents the metadata describing one specific resource type supported in the OSS registry
type OSSResourceTypeDefinition struct {
	EntryType     EntryType `json:"entry_type"`
	Description   string    `json:"description"`
	CMDBClass     string    `json:"cmdb_class"`
	LegacyRMCType string    `json:"legacy_rmc_type"`
	LegacyOSSType EntryType `json:"legacy_oss_type"`
}

// CheckSchemaVersion checks if the SchemaVersion in this entry is compatible with the current library version.
// If it is compatible, it returns nil
// if it is not compatible, it return a descriptive error AND UPDATES the SchemaVersion in the entry to mark the fact
// that it is not compatible
func (r *OSSResourceClassification) CheckSchemaVersion() error {
	return checkSchemaVersion(r, &r.SchemaVersion)
}

// CleanEntryForCompare cleans the content of an OSSEntry so that it can be compared with other entries
// without flagging irrelevant differences
func (r *OSSResourceClassification) CleanEntryForCompare() {
	// nothing to clean
}

// GetOSSEntryID returns a unique identifier for this record
func (r *OSSResourceClassification) GetOSSEntryID() OSSEntryID {
	// Hard-coded ID for this singleton entry
	return `oss_resource_classification`
}

// GetOSSOnboardingPhase returns the current OSS onboarding phase associated with this OSSEntry
func (r *OSSResourceClassification) GetOSSOnboardingPhase() OSSOnboardingPhase {
	return r.OSSOnboardingPhase
}

// GetOSSTags returns the OSS tags associated with this OSSEntry
func (r *OSSResourceClassification) GetOSSTags() *osstags.TagSet {
	// Hard-coded for this singleton entry
	return &osstags.TagSet{}
}

// String returns a short string identifier for this OSSEntry
func (r *OSSResourceClassification) String() string {
	// Hard-coded for this singleton entry
	return fmt.Sprintf("oss_schema(%s)", OSSResourceClassificationEntryName)
}

// Header returns a one-line string representing this OSSEntry record
func (r *OSSResourceClassification) Header() string {
	// Hard-coded for this singleton entry
	return `OSS Resource Classification singleton object (oss_schema)\n`
}

// IsUpdatable returns true if this OSS record can be updated in the Catalog (i.e. it is a real standalone OSS record, not a placeholder for a native Catalog entry)
func (r *OSSResourceClassification) IsUpdatable() bool {
	return true
}

// JSON returns a JSON-style string representation of the entire OSSResourceClassification record (with no header), multi-line
func (r *OSSResourceClassification) JSON() string {
	var result strings.Builder
	result.WriteString("  {\n")
	result.WriteString(`    "oss_resource_classification": `)
	json, _ := json.MarshalIndent(r, "    ", "    ")
	_, err := result.Write(json)
	if err != nil {
		panic(err)
	}
	result.WriteString("\n")
	result.WriteString("  }")
	return result.String()
}

// SetTimes sets the Created and Updated times of this OSSEntry
func (r *OSSResourceClassification) SetTimes(created string, updated string) {
	// No-op: Created and Updated times not tracked for non-extended versions of OSS entries
}

// OSSResourceClassificationEntryName is the name to use in the Catalog for the Resource Classification singleton entry
var OSSResourceClassificationEntryName = `oss_resource_classification`

var _ OSSEntry = &OSSResourceClassification{} // verify
