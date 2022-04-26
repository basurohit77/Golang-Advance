package ossrecord

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// TribeID represents the type for the unique identifer of a tribe.
// It may be a name, but note that it is distinct from the DisplayName and it is immutable
type TribeID string

// OSSTribe represents the OSS info about one Tribe
type OSSTribe struct {
	SchemaVersion             string             `json:"schema_version"` // The version of the schema for this OSSTribe record entry
	TribeID                   TribeID            `json:"tribe_id"`
	OSSTags                   osstags.TagSet     `json:"oss_tags"`
	OSSOnboardingPhase        OSSOnboardingPhase `json:"oss_onboarding_phase"`
	OSSOnboardingApprover     Person             `json:"oss_onboarding_approver"`      // Person who approved this OSS record in RMC
	OSSOnboardingApprovalDate string             `json:"oss_onboarding_approval_date"` // Date that the person approved this OSS record in RMC
	DisplayName               string             `json:"display_name"`
	SegmentID                 SegmentID          `json:"segment_id"`
	Owner                     Person             `json:"owner"`
	ChangeApprovers           []*PersonListEntry `json:"change_approvers"`
}

// MakeOSSTribeID creates a OSSEntryID for a OSS
func MakeOSSTribeID(id TribeID) OSSEntryID {
	if id != "" {
		return OSSEntryID("oss_tribe." + string(id))
	}
	return ""
}

// GetOSSEntryID returns a unique identifier for this record
func (tr *OSSTribe) GetOSSEntryID() OSSEntryID {
	return MakeOSSTribeID(tr.TribeID)
}

// String returns a short string identifier for this TribeInfo
func (tr *OSSTribe) String() string {
	return fmt.Sprintf(`Tribe(%s[%s])`, tr.DisplayName, tr.TribeID)
}

// Header returns a one-line string representing this OSSTribe record
func (tr *OSSTribe) Header() string {
	var onboardingPhase string
	if tr.OSSOnboardingPhase != "" {
		onboardingPhase = fmt.Sprintf("OSSOnboardingPhase=%s", tr.OSSOnboardingPhase)
	}
	return fmt.Sprintf("%s %s OSSTAGS=%s\n", tr.String(), onboardingPhase, tr.OSSTags)
}

// JSON returns a JSON-style string representation of the entire OSSTribe record (with no header), multi-line
func (tr *OSSTribe) JSON() string {
	var result strings.Builder
	result.WriteString("  {\n")
	result.WriteString(`    "oss_tribe": `)
	json, _ := json.MarshalIndent(tr, "    ", "    ")
	_, err := result.Write(json)
	if err != nil {
		panic(err)
	}
	result.WriteString("\n")
	result.WriteString("  }")
	return result.String()
}

// IsUpdatable returns true if this OSS record can be updated in the Catalog (i.e. it is a real standalone OSS record, not a placeholder for a native Catalog entry)
func (tr *OSSTribe) IsUpdatable() bool {
	return true
}

// SetTimes sets the Created and Updated times of this OSSEntry
func (tr *OSSTribe) SetTimes(created string, updated string) {
	// No-op: Created and Updated times not tracked for non-extended versions of OSS entries
}

// GetOSSTags returns the OSS tags associated with this OSSEntry
func (tr *OSSTribe) GetOSSTags() *osstags.TagSet {
	return &tr.OSSTags
}

// ResetForRMC resets the validation info (status and issues) that might become obsolete when an entry is being edited in RMC
func (tr *OSSTribe) ResetForRMC() {
	tr.OSSTags = tr.OSSTags.WithoutPureStatus().Copy()
}

// GetOSSOnboardingPhase returns the current OSS onboarding phase associated with this OSSEntry
func (tr *OSSTribe) GetOSSOnboardingPhase() OSSOnboardingPhase {
	return tr.OSSOnboardingPhase
}

// CheckSchemaVersion checks if the SchemaVersion in this entry is compatible with the current library version.
// If it is compatible, it returns nil
// if it is not compatible, it return a descriptive error AND UPDATES the SchemaVersion in the entry to mark the fact
// that it is not compatible
func (tr *OSSTribe) CheckSchemaVersion() error {
	return checkSchemaVersion(tr, &tr.SchemaVersion)
}

// CleanEntryForCompare cleans the content of an OSSEntry so that it can be compared with other entries
// without flagging irrelevant differences
func (tr *OSSTribe) CleanEntryForCompare() {
	// nothing to clean
	return
}

// DeepCopy performs a deep copy of this OSSTribe object, returning a new OSSTribe object
func (tr *OSSTribe) DeepCopy() *OSSTribe {
	buffer, err := json.Marshal(tr)
	if err != nil {
		panic(err)
	}
	dest := new(OSSTribe)
	err = json.Unmarshal(buffer, dest)
	if err != nil {
		panic(err)
	}
	return dest
}

var _ OSSEntry = &OSSTribe{} // verify
