package ossrecord

import (
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// OSSEntry is a common type to identify different kinds of OSS entries in Catalog: OSSService, OSSSegment, OSSTribe,
// OSSServiceExtended, OSSSegmentExtended, OSSTribeExtended, OSSEnvironmentExtended
type OSSEntry interface {
	GetOSSEntryID() OSSEntryID                 // Return a unique identifier (and presence of this method confirms that this object implements the OSSEntry interface)
	GetOSSTags() *osstags.TagSet               // Return the OSS tags associated with this OSSEntry
	GetOSSOnboardingPhase() OSSOnboardingPhase // Return the current OSS onboarding phase for this OSSEntry
	CheckSchemaVersion() error                 // Check if the SchemaVersion in this entry is compatible with the current library version
	String() string                            // Returns a short string identifier
	Header() string                            // Returns a one-line string representing this entry
	JSON() string                              // Returns a JSON-style string representation of the entire entry (with no header), multi-line
	CleanEntryForCompare()                     // Cleans the content of an OSSEntry so that it can be compared with other entries
	IsUpdatable() bool                         // returns true if this OSS record can be updated in the Catalog (i.e. it is a real standalone OSS record, not a placeholder for a native Catalog entry)
	SetTimes(string, string)                   // Set the Created and Updated times of this OSSEntry
}

// OSSEntryID is a common type for a unique ID to identify any concrete instance of the OSSEntry interface
type OSSEntryID string
