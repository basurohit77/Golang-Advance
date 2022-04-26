package scorecardv1

import (
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// GetOperationalStatus converts the Status field from a ScorecardV1Detail record into a standard ossrecord.OperationalStatus value
func (s *DetailEntry) GetOperationalStatus() (ossrecord.OperationalStatus, bool) {
	switch s.Status {
	case "GA":
		return ossrecord.GA, true
	case "Beta":
		return ossrecord.BETA, true
	case "Experimental":
		return ossrecord.EXPERIMENTAL, true
	case "Deprecated":
		// TODO: Is it possible to be both DEPRECATED and some other status?
		return ossrecord.DEPRECATED, true
	case "NA - Internal Tool":
		return ossrecord.INTERNAL, true
	case "NA - International Tool": // XXX Note typo in Doctor label
		return ossrecord.INTERNAL, true
	case "":
		return ossrecord.OperationalStatusUnknown, true
	// TODO: Need to track THIRDPARTY services in ScorecardV1
	default:
		return ossrecord.OperationalStatusUnknown, false
	}
}
