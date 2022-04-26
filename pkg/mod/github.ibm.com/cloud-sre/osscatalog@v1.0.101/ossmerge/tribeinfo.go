package ossmerge

import (
	"fmt"
	"reflect"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/scorecardv1"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// TribeInfo represents the data model for OSS tribes information for the oss-catalog tool
type TribeInfo struct {
	ossrecordextended.OSSTribeExtended // The full OSS record under construction
	PriorOSS                           ossrecord.OSSTribe
	PriorOSSValidation                 *ossvalidation.OSSValidation // Prios OSSValidation entry for this tribe, if any, found in Catalog
	PriorOSSValidationChecksum         string                       // Checksum for the prior OSSValidation entry for this tribe, if any, found in Catalog
	SourceScorecardV1                  scorecardv1.TribeResource
	segment                            *SegmentInfo
}

// HasPriorOSS returns true if the PriorOSS sub-record in this TribeInfo is present, false otherwise
func (tr *TribeInfo) HasPriorOSS() bool {
	return tr.PriorOSS.TribeID != ""
}

// GetPriorOSS returns a pointer to the PriorOSS record in this TribeInfo if it is valid, nil otherwise
func (tr *TribeInfo) GetPriorOSS() *ossrecord.OSSTribe {
	if tr.HasPriorOSS() {
		return &tr.PriorOSS
	}
	return nil
}

// HasSourceScorecardV1 returns true if the SourceScorecardV1 sub-record in this TribeInfo is present, false otherwise
func (tr *TribeInfo) HasSourceScorecardV1() bool {
	return tr.SourceScorecardV1.Name != ""
}

// GetSourceScorecardV1 returns a pointer to the SourceScorecardV1 sub-record in this TribeInfo if it is valid, nil otherwise
func (tr *TribeInfo) GetSourceScorecardV1() *scorecardv1.TribeResource {
	if tr.HasSourceScorecardV1() {
		return &tr.SourceScorecardV1
	}
	return nil
}

// String returns a short string identifier for this TribeIndo
func (tr *TribeInfo) String() string {
	if tr.segment != nil {
		return fmt.Sprintf(`%s/%s`, tr.OSSTribe.String(), tr.segment.String())
	}
	return fmt.Sprintf(`%s/Segment(none))`, tr.OSSTribe.String())
}

// allTribesById records all the tribes across all segments by their TribeID -- to detect duplicate TribeIDs
var allTribesByID = make(map[ossrecord.TribeID]*TribeInfo)

// CheckConsistency verifies that this TribeInfo record is internally consistent
// and properly registered in the various lookup tables
// It returns an error if it finds an issue, or panics if the "panicIfError" flag is true
func (tr *TribeInfo) CheckConsistency(tribeID ossrecord.TribeID, tribeName string, panicIfError bool) error {
	issues := strings.Builder{}
	var foundByID = "<none>"
	var foundByName = "<none>"
	if tr.OSSTribe.TribeID == "" {
		issues.WriteString(fmt.Sprintf(`;  missing tribeID`))
	}
	if tr.OSSTribe.SegmentID == "" {
		issues.WriteString(fmt.Sprintf(`;  missing segmentID`))
	} else if tr.OSSTribe.SegmentID != tr.segment.OSSSegment.SegmentID {
		if seg1, found1 := segmentsByID[tr.OSSTribe.SegmentID]; found1 {
			issues.WriteString(fmt.Sprintf(`;  invalid segmentID %s = expected %s`, seg1.String(), tr.segment.String()))
			err1 := seg1.CheckConsistency(tr.OSSTribe.SegmentID, "", false)
			if err1 != nil {
				issues.WriteString(fmt.Sprintf(`;  ALSO((%s))`, err1.Error()))
			}
		} else {
			issues.WriteString(fmt.Sprintf(`;  invalid segmentID (not found): Segment([%s]) = expected %s`, tr.OSSTribe.SegmentID, tr.segment.String()))
		}
	}
	if tribeID != "" && tr.OSSTribe.TribeID != tribeID {
		issues.WriteString(fmt.Sprintf(`;  expected tribeID="%s"`, tribeID))
	}
	baseName, _ := osstags.GetOSSTestBaseName(tr.OSSTribe.DisplayName)
	if tribeName != "" && tr.OSSTribe.DisplayName != tribeName && !debug.CompareDuplicateNames(tribeName, tr.OSSTribe.DisplayName) {
		baseName2, _ := osstags.GetOSSTestBaseName(tribeName)
		if baseName != baseName2 {
			issues.WriteString(fmt.Sprintf(`;  expected tribeName="%s"`, tribeName))
		}
	}
	if tr.OSSTribe.TribeID != "" {
		if tr1, found1 := tr.segment.tribes[tr.OSSTribe.TribeID]; found1 {
			foundByID = tr1.String()
			if tr1 != tr {
				issues.WriteString(fmt.Sprintf(`;  different entry found in tribes table for this segment`))
			}
		} else {
			issues.WriteString(fmt.Sprintf(`;  not found in tribes table for this segment`))
		}
		if tr1, found1 := allTribesByID[tr.OSSTribe.TribeID]; found1 {
			if tr1 != tr {
				issues.WriteString(fmt.Sprintf(`;  duplicate TribeID in global table - prior=%s`, tr1.String()))
			}
		} else {
			issues.WriteString(fmt.Sprintf(`;  TribeID not found in global table`))
		}
	}
	if baseName != "" {
		if tr1, found1 := tr.segment.tribeNames[baseName]; found1 {
			foundByName = tr1.String()
			if tr1 != tr {
				issues.WriteString(fmt.Sprintf(`;  different entry found in tribeNames table for this segment`))
			}
		} else {
			issues.WriteString(fmt.Sprintf(`;  not found in tribeNames table for this segment`))
		}
	}
	if issues.Len() > 0 {
		err := fmt.Errorf("Consistency issues with TribeInfo object %s   foundByID=%s   foundByName=%s  %s", tr.String(), foundByID, foundByName, issues.String())
		if panicIfError {
			panic(err)
		} else {
			return err
		}
	}
	return nil
}

// LookupTribe returns the TribeInfo record associated with a given TribeID within this SegmentInfo or creates a new record if appropriate.
// If no record exists and the parameter 'createIfNeeded' is false, 'nil' is returned.
func (seg *SegmentInfo) LookupTribe(tribeID ossrecord.TribeID, createIfNeeded bool) (tr *TribeInfo, found bool) {
	if tr, found = seg.tribes[tribeID]; found {
		_ = tr.CheckConsistency(tribeID, "", true)
		return tr, true
	}
	if tr0, found0 := allTribesByID[tribeID]; found0 {
		panic(fmt.Sprintf(`LookupTribe(%s) in the wrong segment %s - tribeID already found in %s`, tribeID, seg.String(), tr0.String()))
	} else if createIfNeeded {
		tr := new(TribeInfo)
		tr.OSSTribe.SegmentID = seg.OSSSegment.SegmentID
		tr.segment = seg
		tr.OSSTribe.TribeID = tribeID
		tr.OSSValidation = ossvalidation.New("", options.GlobalOptions().LogTimeStamp)
		seg.tribes[tribeID] = tr
		allTribesByID[tribeID] = tr
		_ = tr.CheckConsistency(tribeID, "", true)
		return tr, true
	}
	return nil, false
}

// SetName sets the Display Name for this Tribe, and indexes it so that it can be found by LookupTribeName()
func (tr *TribeInfo) SetName(tribeName string) {
	if tr.OSSTribe.DisplayName != "" {
		panic(fmt.Sprintf(`Duplicate call to TribeInfo.SetName() for existing %s - new name=%s`, tr.String(), tribeName))
	}
	baseName, _ := osstags.GetOSSTestBaseName(tribeName)
	if tr0, found0 := tr.segment.tribeNames[baseName]; found0 {
		for found0 {
			newName := debug.MakeDuplicateName(tribeName)
			debug.PrintError(`Duplicate Tribe name: "%s" - prior=%s - new=%s - converting it to "%s"`, tribeName, tr0.String(), tr.String(), newName)
			tribeName = newName
			baseName, _ = osstags.GetOSSTestBaseName(newName)
			tr0, found0 = tr.segment.tribeNames[baseName]
		}
	}
	tr.OSSTribe.DisplayName = tribeName
	tr.segment.tribeNames[baseName] = tr
	_ = tr.CheckConsistency("", tribeName, true) // XXX Do this earlier in case tr.segment is nil ?
}

// LookupTribeName finds a TribeInfo entry by its Display Name (must already exist)
func (seg *SegmentInfo) LookupTribeName(tribeName string) (tr *TribeInfo, found bool) {
	baseName, _ := osstags.GetOSSTestBaseName(tribeName)
	tr, found = seg.tribeNames[baseName]
	if found {
		_ = tr.CheckConsistency("", tribeName, true)
	}
	return tr, found
}

// OSSEntry returns the contents of this TribeInfo as a OSSEntry type
func (tr *TribeInfo) OSSEntry() ossrecord.OSSEntry {
	return &tr.OSSTribeExtended
}

// PriorOSSEntry returns pointer to the PriorOSS record in this TribeInfo as a OSSEntry type, if it is valid, nil otherwise
func (tr *TribeInfo) PriorOSSEntry() ossrecord.OSSEntry {
	return tr.GetPriorOSS()
}

// IsDeletable returns true if there is no reason to keep this OSS record in the Catalog (i.e it has no source, and no non-zero OSSControl or RMC data)
func (tr *TribeInfo) IsDeletable() bool {
	if !tr.HasSourceScorecardV1() && !tr.OSSTribe.OSSTags.Contains(osstags.OSSOnly) && !tr.OSSTribe.OSSTags.Contains(osstags.OSSTest) && tr.OSSTribe.OSSOnboardingPhase == "" {
		if !ossrunactions.Doctor.IsEnabled() && tr.HasPriorOSS() && tr.GetPriorOSS().OSSOnboardingPhase == "" {
			if tr.OSSValidation != nil {
				tr.OSSValidation.AddIssueIgnoreDup(ossvalidation.INFO, `Cannot confirm if entry is deletable when Doctor is disabled`, ``).TagPriorOSS()
			} else {
				debug.Info(`Cannot confirm if entry is deletable when Doctor is disabled: %s`, tr.String())
			}
			return false
		}
		return true
	}
	return false
}

// IsUpdatable returns true if this OSS record can be updated in the Catalog (i.e. it is a real standalone OSS record, not a placeholder for a native Catalog entry)
func (tr *TribeInfo) IsUpdatable() bool {
	return tr.OSSTribe.IsUpdatable()
}

// IsValid returns true if this TribeInfo record represents a valid entry that
// can be updated or merged into the Catalog
func (tr *TribeInfo) IsValid() bool {
	return tr.OSSTribe.TribeID != ""
}

// Header returns a short header representing this TribeInfo record, suitable for printing in logs and reports
func (tr *TribeInfo) Header() string {
	result := strings.Builder{}
	result.WriteString(tr.OSSTribe.Header())
	if tr.OSSValidation != nil {
		result.WriteString(tr.OSSValidation.Header())
	}
	return result.String()
}

// Details returns a long text representing this TribeInfo record (including a Header), suitable for printing in logs and reports
func (tr *TribeInfo) Details() string {
	result := strings.Builder{}
	result.WriteString(tr.OSSTribe.Header())
	if tr.OSSValidation != nil {
		result.WriteString(tr.OSSValidation.Details())
	}
	return result.String()
}

// Diffs returns a list of differences between the new and priorOSS versions of the OSSTribe record in this TribeInfo
func (tr *TribeInfo) Diffs() *compare.Output {
	out := compare.Output{}
	compare.DeepCompare("before", &tr.PriorOSS, "after", &tr.OSSTribe, &out)
	if (tr.OSSValidation == nil && tr.PriorOSSValidation != nil) || (tr.OSSValidation != nil && tr.OSSValidation.Checksum() != tr.PriorOSSValidationChecksum) {
		compare.DeepCompare("before.OSSValidation", tr.PriorOSSValidation, "after.OSSValidation", tr.OSSValidation, &out)
		//out.AddOtherDiff("OSSValidation updated")
	}
	return &out
}

// GlobalSortKey returns a global key for sorting across all ossmerge.Model entries of all types
func (tr *TribeInfo) GlobalSortKey() string {
	return "2." + tr.String()
}

// GetAccessor returns the Accessor function for a given sub-record within the TribeInfo
func (tr *TribeInfo) GetAccessor(t reflect.Type) ossmergemodel.AccessorFunc {
	panic("GetAccessor() not defined for type ossmerge.TribeInfo")
}

// GetOSSValidation returns the OSSValidation record in this TribeInfo, or nil if there is none
func (tr *TribeInfo) GetOSSValidation() *ossvalidation.OSSValidation {
	return tr.OSSValidation
}

// SkipMerge returns true if we should skip the OSS merge for this record and simply copy it from
// the prior OSS record.
// If not empty, the "reason" string provides details on why this record should or should not be merged
func (tr *TribeInfo) SkipMerge() (skip bool, reason string) {
	return ossmergemodel.SkipMerge(tr)
}

var _ ossmergemodel.Model = &TribeInfo{} // verify
