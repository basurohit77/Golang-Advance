package ossmergemodel

import (
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// Common methods for all implementations of the Model interface

// SkipMerge returns true if we should skip the OSS merge for this record and simply copy it from
// the prior OSS record.
// If not empty, the "reason" string provides details on why this record should or should not be merged
func SkipMerge(m Model) (skip bool, reason string) {
	tags := m.OSSEntry().GetOSSTags()
	ossval := m.GetOSSValidation()
	if tags.Contains(osstags.OSSTest) {
		if options.GlobalOptions().TestMode {
			if tags.Contains(osstags.OSSOnly) {
				if ossval == nil || ossval.NumTrueSources() == 0 {
					if m.HasPriorOSS() {
						skip = true
						reason = "OSS-Only+OSS-Test record"
					} else {
						skip = false
						reason = "OSS-Only+OSS-Test record but no prior OSS record exists"
					}
				} else {
					skip = false
					reason = "OSS-Only+OSS-Test record but has some true sources other than the prior OSS record itself"
				}
			} else {
				skip = false
				reason = ""
			}
		} else {
			if m.HasPriorOSS() {
				skip = true
				reason = "OSS-Test record but not running in test mode"
			} else {
				skip = false
				reason = "OSS-Test record, not running in test mode but no prior OSS record exists"
			}
		}
	} else {
		if tags.Contains(osstags.OSSOnly) {
			if ossval == nil || ossval.NumTrueSources() == 0 {
				if m.HasPriorOSS() {
					skip = true
					reason = "OSS-Only record"
				} else {
					skip = false
					reason = "OSS-Only record but no prior OSS record exists"
				}
			} else {
				skip = false
				reason = "OSS-Only record but has some true sources other than the prior OSS record itself"
			}
		} else {
			skip = false
			reason = ""
		}
	}

	return skip, reason
}
