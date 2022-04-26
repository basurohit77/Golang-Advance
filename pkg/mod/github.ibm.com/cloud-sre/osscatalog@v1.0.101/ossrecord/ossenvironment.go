package ossrecord

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// OSSEnvironment represents all the OSS information maintained about a Region, Zone or Datacenter in the IBM Public Cloud,
// or a Dedicated or Local Cloud region.
// Dedicated and Local regions are essentially deprecated, but many are still in existence and actively used by Clients,
// hence they must be tracked.
type OSSEnvironment struct {
	SchemaVersion             string             `json:"schema_version"` // The version of the schema for this OSSEnvironment record entry
	EnvironmentID             EnvironmentID      `json:"id"`             // Unique identifier for this OSSEnvironment -- in the form of a CRN mask
	OSSTags                   osstags.TagSet     `json:"oss_tags"`
	OSSOnboardingPhase        OSSOnboardingPhase `json:"oss_onboarding_phase"`
	OSSOnboardingApprover     Person             `json:"oss_onboarding_approver"`      // Person who approved this OSS record in RMC
	OSSOnboardingApprovalDate string             `json:"oss_onboarding_approval_date"` // Date that the person approved this OSS record in RMC
	ParentID                  EnvironmentID      `json:"parent_id"`                    // Parent of this OSSEnvironment, if any (used to represent Zones or Datacenters that are inside a Multi-Zone Region)
	DisplayName               string             `json:"display_name"`                 //
	Type                      EnvironmentType    `json:"type"`                         //
	Status                    EnvironmentStatus  `json:"status"`                       //
	ReferenceCatalogID        CatalogID          `json:"reference_catalog_id"`         // ID of the underlying Global Catalog entry that represents this environment
	ReferenceCatalogPath      string             `json:"reference_catalog_path"`       // path of the main Global Catalog entry (if any) that represents this environment for which this entry contains the OSS information
	OwningSegment             SegmentID          `json:"owning_segment"`               // The OSSSegment that owns this environment (can be empty for IBM Cloud environments)
	OwningClient              string             `json:"owning_client"`                // Optional: client that owns a group of Dedicated/Local environments (for access control / Blink)
	Description               string             `json:"description"`                  //
	LegacyIMSID               string             `json:"ims_id"`                       // Optional: Softlayer IMS identifier for a datacenter (numeric)
	LegacyMCCPID              string             `json:"mccp_id"`                      // Optional: Bluemix MCCP ID (not available for newer environments)
	LegacyDoctorCRN           string             `json:"legacy_doctor_crn"`            // Optional: the old-style CRN Mask used in Doctor, prior to the uniformization of all CRN Masks as Environment IDs
}

// EnvironmentID is the unique identifier for a OSSEnvironment - represented as a CRN mask
type EnvironmentID string

// EnvironmentType represents the type of a OSSEnvironment - basically either Public, Dedicated or Local
type EnvironmentType string

const (
	// EnvironmentIBMCloudRegion is the EnvironmentType representing a Multi-Zone Region in the IBM Public Cloud
	EnvironmentIBMCloudRegion EnvironmentType = "IBMCLOUD_REGION"
	// EnvironmentIBMCloudDatacenter is the EnvironmentType representing a standalone Datacenter in the IBM Public Cloud
	EnvironmentIBMCloudDatacenter EnvironmentType = "IBMCLOUD_DATACENTER"
	// EnvironmentIBMCloudZone is the EnvironmentType representing an Availability Zone in the IBM Public Cloud
	EnvironmentIBMCloudZone EnvironmentType = "IBMCLOUD_ZONE"
	// EnvironmentIBMCloudPOP is the EnvironmentType representing a POP (point of presence) within the IBM Public Cloud
	EnvironmentIBMCloudPOP EnvironmentType = "IBMCLOUD_POP"
	// EnvironmentIBMCloudSatellite is the EnvironmentType representing a Satellite container in the IBM Public Cloud
	EnvironmentIBMCloudSatellite EnvironmentType = "IBMCLOUD_SATELLITE"
	// EnvironmentIBMCloudDedicated is the EnvironmentType representing a Dedicated Cloud region
	EnvironmentIBMCloudDedicated EnvironmentType = "IBMCLOUD_DEDICATED"
	// EnvironmentIBMCloudLocal is the EnvironmentType representing a Local Cloud region
	EnvironmentIBMCloudLocal EnvironmentType = "IBMCLOUD_LOCAL"
	// EnvironmentIBMCloudStaging is the EnvironmentType representing a Staging (test) Cloud region
	EnvironmentIBMCloudStaging EnvironmentType = "IBMCLOUD_STAGING"
	// EnvironmentGAAS is the EnvironmentType representing an environment outside IBM Cloud (GaaS project)
	// The OwningSegment attribute of the OSSEnvironment record indicates which non-IBM Cloud entity controls this environment
	EnvironmentGAAS EnvironmentType = "GAAS"
	// EnvironmentTypeSpecial is the EnvironmentType for a special entry try that may not exactly be a true environment, but is sometimes used
	// as the "location" field in CRNs and needs to be tracked for purposes of Change Management, Status Page, etc.  -- e.g. "global"
	EnvironmentTypeSpecial EnvironmentType = "SPECIAL"
	// EnvironmentTypeUnknown indicates that we were unable to determine a valid EnvironmentType
	EnvironmentTypeUnknown EnvironmentType = "<unknown>"
)

// EnvironmentStatus represents the status of a OSSEnvironment - basically either active or decommissioned (for some Dedicated/Local regions)
type EnvironmentStatus string

const (
	// EnvironmentActive is the EnvironmentStatus representing an active Region, Datacenter or other Environment (Public, Dedicated, Local, GAAS)
	EnvironmentActive EnvironmentStatus = "ACTIVE"
	// EnvironmentDecommissioned is the EnvironmentStatus representing a decomissioned Dedicated or Local Region
	EnvironmentDecommissioned EnvironmentStatus = "DECOMMISSIONED"
	// EnvironmentNotReady is the EnvironmentStatus representing a Region, Datacenter or other Environment that is not yet ready for use (registration/OSS status may be incomplete)
	EnvironmentNotReady EnvironmentStatus = "NOTREADY"
	// EnvironmentSelectAvailability is the EnvironmentStatus for a Region, Datacenter or other Environment that is available to a selected set of Clients in the IBM Cloud catalog, but not generally available
	EnvironmentSelectAvailability EnvironmentStatus = "SELECTAVAILABILITY"
	// EnvironmentStatusUnknown indicates that we were unable to determine a valid EnvironmentStatus
	EnvironmentStatusUnknown EnvironmentStatus = "<unknown>"
)

const ossEnvironmentIDPrefix = "oss_environment."

// MakeOSSEnvironmentID returns a global unique ID for a OSSEnvironment record, given only its (internal) EnvironmentID
// This function is used when we do not have access to the full OSSEnvironment record,
// for example when dealing with the parent ID of an another OSSEnvironment
func MakeOSSEnvironmentID(id EnvironmentID) OSSEntryID {
	if id != "" {
		return OSSEntryID(ossEnvironmentIDPrefix + string(id))
	}
	return ""
}

// IsOSSEnvironmentID returns true if this string looks like a valid OSSEntryID for a OSSEnvironment object
func IsOSSEnvironmentID(id OSSEntryID) bool {
	return strings.HasPrefix(string(id), ossEnvironmentIDPrefix)
}

// GetOSSEntryID returns a unique identifier for this record
func (env *OSSEnvironment) GetOSSEntryID() OSSEntryID {
	return MakeOSSEnvironmentID(env.EnvironmentID)
}

// String returns a short string identifier for this OSSEnvironment
func (env *OSSEnvironment) String() string {
	return fmt.Sprintf(`Environment(%s[%s])`, env.DisplayName, env.EnvironmentID)
}

// Header returns a one-line string representing this OSSEnvironment record
func (env *OSSEnvironment) Header() string {
	var onboardingPhase string
	if env.OSSOnboardingPhase != "" {
		onboardingPhase = fmt.Sprintf("OSSOnboardingPhase=%s", env.OSSOnboardingPhase)
	}
	return fmt.Sprintf("%s \"%s\"  %s/%s %s OSSTAGS=%s  parent=%s  catalog_id=%s  catalog_path=%s\n", env.String(), env.DisplayName, env.Type, env.Status, onboardingPhase, &env.OSSTags, env.ParentID, env.ReferenceCatalogID, env.ReferenceCatalogPath)
}

// JSON returns a JSON-style string representation of the entire OSSEnvironment record (with no header), multi-line
func (env *OSSEnvironment) JSON() string {
	var result strings.Builder
	result.WriteString("  {\n")
	result.WriteString(`    "oss_environment": `)
	json, _ := json.MarshalIndent(env, "    ", "    ")
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
func (env *OSSEnvironment) CleanEntryForCompare() {
	// nothing to clean
	return
}

// IsUpdatable returns true if this OSS record can be updated in the Catalog (i.e. it is a real standalone OSS record, not a placeholder for a native Catalog entry)
func (env *OSSEnvironment) IsUpdatable() bool {
	return !env.OSSTags.Contains(osstags.CatalogNative)
}

// SetTimes sets the Created and Updated times of this OSSEntry
func (env *OSSEnvironment) SetTimes(created string, updated string) {
	// No-op: Created and Updated times not tracked for non-extended versions of OSS entries
}

// GetOSSTags returns the OSS tags associated with this OSSEntry
func (env *OSSEnvironment) GetOSSTags() *osstags.TagSet {
	return &env.OSSTags
}

// ResetForRMC resets the validation info (status and issues) that might become obsolete when an entry is being edited in RMC
func (env *OSSEnvironment) ResetForRMC() {
	env.OSSTags = env.OSSTags.WithoutPureStatus().Copy()
}

// GetOSSOnboardingPhase returns the current OSS onboarding phase associated with this OSSEntry
func (env *OSSEnvironment) GetOSSOnboardingPhase() OSSOnboardingPhase {
	return env.OSSOnboardingPhase
}

// CheckSchemaVersion checks if the SchemaVersion in this entry is compatible with the current library version.
// If it is compatible, it returns nil
// if it is not compatible, it return a descriptive error AND UPDATES the SchemaVersion in the entry to mark the fact
// that it is not compatible
func (env *OSSEnvironment) CheckSchemaVersion() error {
	return checkSchemaVersion(env, &env.SchemaVersion)
}

// DeepCopy performs a deep copy of this OSSEnvironment object, returning a new OSSEnvironment object
func (env *OSSEnvironment) DeepCopy() *OSSEnvironment {
	buffer, err := json.Marshal(env)
	if err != nil {
		panic(err)
	}
	dest := new(OSSEnvironment)
	err = json.Unmarshal(buffer, dest)
	if err != nil {
		panic(err)
	}
	return dest
}

var _ OSSEntry = &OSSEnvironment{} // verify that this is a valid OSSEntry
