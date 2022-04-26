package ossrecord

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// SegmentID represents the type for the unique identifer of a segment.
// It may be a name, but note that it is distinct from the DisplayName and it is immutable
type SegmentID string

// OSSSegment represents the OSS info about one Segment
type OSSSegment struct {
	SchemaVersion             string             `json:"schema_version"` // The version of the schema for this OSSSegment record entry
	SegmentID                 SegmentID          `json:"segment_id"`
	OSSTags                   osstags.TagSet     `json:"oss_tags"`
	OSSOnboardingPhase        OSSOnboardingPhase `json:"oss_onboarding_phase"`
	OSSOnboardingApprover     Person             `json:"oss_onboarding_approver"`      // Person who approved this OSS record in RMC
	OSSOnboardingApprovalDate string             `json:"oss_onboarding_approval_date"` // Date that the person approved this OSS record in RMC
	SegmentType               SegmentType        `json:"segment_type"`
	DisplayName               string             `json:"display_name"`
	Owner                     Person             `json:"owner"`
	TechnicalContact          Person             `json:"technical_contact"` // TODO: Do we need a Technical Contact at the Segment level?
	ERCAApprovers             []*PersonListEntry `json:"erca_approvers"`
	ChangeCommApprovers       []*PersonListEntry `json:"change_comm_approvers"`
}

// SegmentType represents the type of a segment represented
type SegmentType string

const (
	// SegmentTypeIBMPublicCloud is the SegmerntType for a Segment that is part of the IBM Public Cloud (default)
	SegmentTypeIBMPublicCloud SegmentType = "IBM_PUBLIC_CLOUD"
	// SegmentTypeGaaS is the SegmerntType for a Segment that is not part of the IBM Public Cloud catalog proper but that is nonetheless tracked through IBM Cloud OSS,
	// usihg the Global Ops as a Service (GaaS) framework.
	SegmentTypeGaaS SegmentType = "GAAS"
)

// MakeOSSSegmentID creates a OSSEntryID for a OSS Segment entry in
func MakeOSSSegmentID(id SegmentID) OSSEntryID {
	if id != "" {
		return OSSEntryID("oss_segment." + string(id))
	}
	return ""
}

// GetOSSEntryID returns a unique identifier for this record
func (seg *OSSSegment) GetOSSEntryID() OSSEntryID {
	return MakeOSSSegmentID(seg.SegmentID)
}

// String returns a short string identifier for this SegmentIndo
func (seg *OSSSegment) String() string {
	return fmt.Sprintf(`Segment(%s[%s])`, seg.DisplayName, seg.SegmentID)
}

// Header returns a one-line string representing this OSSSegment record
func (seg *OSSSegment) Header() string {
	var onboardingPhase string
	if seg.OSSOnboardingPhase != "" {
		onboardingPhase = fmt.Sprintf("OSSOnboardingPhase=%s", seg.OSSOnboardingPhase)
	}
	return fmt.Sprintf("%s  %s %s OSSTAGS=%s\n", seg.String(), seg.SegmentType, onboardingPhase, seg.OSSTags)
}

// JSON returns a JSON-style string representation of the entire OSSSegment record (with no header), multi-line
func (seg *OSSSegment) JSON() string {
	var result strings.Builder
	result.WriteString("  {\n")
	result.WriteString(`    "oss_segment": `)
	json, _ := json.MarshalIndent(seg, "    ", "    ")
	_, err := result.Write(json)
	if err != nil {
		panic(err)
	}
	result.WriteString("\n")
	result.WriteString("  }")
	return result.String()
}

// CleanEntryForCompare cleans the content of an OSSEntry so that it can be compared with other entries
// without flagging irrelevant differences
func (seg *OSSSegment) CleanEntryForCompare() {
	// nothing to clean
	return
}

// IsUpdatable returns true if this OSS record can be updated in the Catalog (i.e. it is a real standalone OSS record, not a placeholder for a native Catalog entry)
func (seg *OSSSegment) IsUpdatable() bool {
	return true
}

// SetTimes sets the Created and Updated times of this OSSEntry
func (seg *OSSSegment) SetTimes(created string, updated string) {
	// No-op: Created and Updated times not tracked for non-extended versions of OSS entries
}

// GetOSSTags returns the OSS tags associated with this OSSEntry
func (seg *OSSSegment) GetOSSTags() *osstags.TagSet {
	return &seg.OSSTags
}

// ResetForRMC resets the validation info (status and issues) that might become obsolete when an entry is being edited in RMC
func (seg *OSSSegment) ResetForRMC() {
	seg.OSSTags = seg.OSSTags.WithoutPureStatus().Copy()
}

// GetOSSOnboardingPhase returns the current OSS onboarding phase associated with this OSSEntry
func (seg *OSSSegment) GetOSSOnboardingPhase() OSSOnboardingPhase {
	return seg.OSSOnboardingPhase
}

// CheckSchemaVersion checks if the SchemaVersion in this entry is compatible with the current library version.
// If it is compatible, it returns nil
// if it is not compatible, it return a descriptive error AND UPDATES the SchemaVersion in the entry to mark the fact
// that it is not compatible
func (seg *OSSSegment) CheckSchemaVersion() error {
	return checkSchemaVersion(seg, &seg.SchemaVersion)
}

// DeepCopy performs a deep copy of this OSSSegment object, returning a new OSSSegment object
func (seg *OSSSegment) DeepCopy() *OSSSegment {
	buffer, err := json.Marshal(seg)
	if err != nil {
		panic(err)
	}
	dest := new(OSSSegment)
	err = json.Unmarshal(buffer, dest)
	if err != nil {
		panic(err)
	}
	return dest
}

var _ OSSEntry = &OSSSegment{} // verify
