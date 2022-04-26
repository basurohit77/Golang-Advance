package ossmerge

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/scorecardv1"
)

// SegmentInfo represents the data model for OSS segments information for the oss-catalog tool
type SegmentInfo struct {
	ossrecordextended.OSSSegmentExtended // The full OSS record under construction
	PriorOSS                             ossrecord.OSSSegment
	PriorOSSValidation                   *ossvalidation.OSSValidation // Prios OSSValidation entry for this segment, if any, found in Catalog
	PriorOSSValidationChecksum           string                       // Checksum for the prior OSSValidation entry for this segment, if any, found in Catalog
	SourceScorecardV1                    scorecardv1.SegmentResource
	tribes                               map[ossrecord.TribeID]*TribeInfo
	tribeNames                           map[string]*TribeInfo
	forceDelete                          bool
}

// HasPriorOSS returns true if the PriorOSS sub-record in this SegmentInfo is present, false otherwise
func (seg *SegmentInfo) HasPriorOSS() bool {
	return seg.PriorOSS.SegmentID != ""
}

// GetPriorOSS returns a pointer to the PriorOSS record in this SegmentInfo if it is valid, nil otherwise
func (seg *SegmentInfo) GetPriorOSS() *ossrecord.OSSSegment {
	if seg.HasPriorOSS() {
		return &seg.PriorOSS
	}
	return nil
}

// HasSourceScorecardV1 returns true if the SourceScorecardV1 sub-record in this SegmentInfo is present, false otherwise
func (seg *SegmentInfo) HasSourceScorecardV1() bool {
	return seg.SourceScorecardV1.Name != ""
}

// GetSourceScorecardV1 returns a pointer to the SourceScorecardV1 sub-record in this SegmentInfo if it is valid, nil otherwise
func (seg *SegmentInfo) GetSourceScorecardV1() *scorecardv1.SegmentResource {
	if seg.HasSourceScorecardV1() {
		return &seg.SourceScorecardV1
	}
	return nil
}

// String returns a short string identifier for this SegmentIndo
func (seg *SegmentInfo) String() string {
	return seg.OSSSegment.String()
}

// segmentsById records all the segments by their SegmentID -- primary table for the data model
var segmentsByID = make(map[ossrecord.SegmentID]*SegmentInfo)

var segmentsByName = make(map[string]*SegmentInfo)

// CheckConsistency verifies that this SegmentInfo record is internally consistent
// and properly registered in the various lookup tables
// It returns an error if it finds an issue, or panics if the "panicIfError" flag is true
func (seg *SegmentInfo) CheckConsistency(segmentID ossrecord.SegmentID, segmentName string, panicIfError bool) error {
	issues := strings.Builder{}
	var foundByID = "<none>"
	var foundByName = "<none>"
	if seg.OSSSegment.SegmentID == "" {
		issues.WriteString(fmt.Sprintf(`;  missing segmentID`))
	}
	if segmentID != "" && seg.OSSSegment.SegmentID != segmentID {
		issues.WriteString(fmt.Sprintf(`;  expected segmentID="%s"`, segmentID))
	}
	baseName, _ := osstags.GetOSSTestBaseName(seg.OSSSegment.DisplayName)
	if segmentName != "" && seg.OSSSegment.DisplayName != segmentName {
		baseName2, _ := osstags.GetOSSTestBaseName(segmentName)
		if baseName != baseName2 {
			issues.WriteString(fmt.Sprintf(`;  expected segmentName="%s"`, segmentName))
		}
	}
	if seg.OSSSegment.SegmentID != "" {
		if seg1, found1 := segmentsByID[seg.OSSSegment.SegmentID]; found1 {
			foundByID = seg1.String()
			if seg1 != seg {
				issues.WriteString(fmt.Sprintf(`;  different entry found in segmentID table`))
			}
		} else {
			issues.WriteString(fmt.Sprintf(`;  not found in segmentID table`))
		}
	}
	if baseName != "" {
		if seg1, found1 := segmentsByName[baseName]; found1 {
			foundByName = seg1.String()
			if seg1 != seg {
				issues.WriteString(fmt.Sprintf(`;  different entry found in segmentName table`))
			}
		} else {
			issues.WriteString(fmt.Sprintf(`;  not found in segmentName table`))
		}
	}
	if issues.Len() > 0 {
		err := fmt.Errorf("Consistency issues with SegmentInfo object %s   foundByID=%s   foundByName=%s  %s", seg.String(), foundByID, foundByName, issues.String())
		if panicIfError {
			panic(err)
		} else {
			return err
		}
	}
	return nil
}

// LookupSegment returns the SegmentInfo record associated with a given SegmentID or creates a new record if appropriate.
// If no record exists and the parameter 'createIfNeeded' is false, 'nil' is returned.
func LookupSegment(segmentID ossrecord.SegmentID, createIfNeeded bool) (seg *SegmentInfo, found bool) {
	if seg, found = segmentsByID[segmentID]; found {
		_ = seg.CheckConsistency(segmentID, "", true)
		return seg, true
	}
	if createIfNeeded {
		seg := new(SegmentInfo)
		seg.OSSSegment.SegmentID = segmentID
		seg.tribes = make(map[ossrecord.TribeID]*TribeInfo)
		seg.tribeNames = make(map[string]*TribeInfo)
		seg.OSSValidation = ossvalidation.New("", options.GlobalOptions().LogTimeStamp)
		segmentsByID[segmentID] = seg
		_ = seg.CheckConsistency(segmentID, "", true)
		return seg, true
	}
	return nil, false
}

// SetName sets the Display Name for this Segment, and indexes it so that it can be found by LookupSegmentName()
func (seg *SegmentInfo) SetName(segmentName string) {
	if seg.OSSSegment.DisplayName != "" {
		panic(fmt.Sprintf(`Duplicate call to SegmentInfo.SetName(%s) for existing %s`, segmentName, seg.String()))
	}
	baseName, _ := osstags.GetOSSTestBaseName(segmentName)
	if seg0, found0 := segmentsByName[baseName]; found0 {
		debug.PrintError(`Duplicate Segment name: "%s" - prior=%s - new=%s - ignoring the new entry`, segmentName, seg0.String(), seg.String())
	} else {
		seg.OSSSegment.DisplayName = segmentName
		segmentsByName[baseName] = seg
		_ = seg.CheckConsistency("", segmentName, true)
	}
}

// LookupSegmentName finds a SegmentInfo entry by its Display Name (must already exist)
func LookupSegmentName(segmentName string) (seg *SegmentInfo, found bool) {
	baseName, _ := osstags.GetOSSTestBaseName(segmentName)
	seg, found = segmentsByName[baseName]
	if found {
		_ = seg.CheckConsistency("", segmentName, true)
	}
	return seg, found
}

// ListAllSegments invokes the handler function on all SegmentInfo records whose display_name matches the specified pattern
// (or all records, if the pattern is empty)
func ListAllSegments(pattern *regexp.Regexp, handler func(seg *SegmentInfo)) error {
	for _, seg := range segmentsByID {
		name := seg.OSSSegment.DisplayName
		if pattern == nil || pattern.FindString(name) != "" {
			handler(seg)
		}
	}
	return nil
}

// ListAllTribes invokes the handler function on all TribeInfo records belonging to the given Segment,
// whose display_name matches the specified pattern
// (or all records, if the pattern is empty)
func (seg *SegmentInfo) ListAllTribes(pattern *regexp.Regexp, handler func(tr *TribeInfo)) error {
	for _, tr := range seg.tribes {
		name := tr.OSSTribe.DisplayName
		if pattern == nil || pattern.FindString(name) != "" {
			handler(tr)
		}
	}
	return nil
}

// OSSEntry returns the contents of this SegmentInfo as a OSSEntry type
func (seg *SegmentInfo) OSSEntry() ossrecord.OSSEntry {
	return &seg.OSSSegmentExtended
}

// PriorOSSEntry returns pointer to the PriorOSS record in this SegmentInfo as a OSSEntry type, if it is valid, nil otherwise
func (seg *SegmentInfo) PriorOSSEntry() ossrecord.OSSEntry {
	return seg.GetPriorOSS()
}

// IsDeletable returns true if there is no reason to keep this OSS record in the Catalog (i.e it has no source, and no non-zero OSSControl or RMC data)
func (seg *SegmentInfo) IsDeletable() bool {
	if seg.forceDelete || (!seg.HasSourceScorecardV1() && !seg.OSSSegment.OSSTags.Contains(osstags.OSSOnly) && !seg.OSSSegment.OSSTags.Contains(osstags.OSSTest) && seg.OSSSegment.OSSOnboardingPhase == "") {
		if !ossrunactions.Doctor.IsEnabled() && seg.HasPriorOSS() && seg.GetPriorOSS().OSSOnboardingPhase == "" {
			if seg.OSSValidation != nil {
				seg.OSSValidation.AddIssueIgnoreDup(ossvalidation.INFO, `Cannot confirm if entry is deletable when Doctor is disabled`, ``).TagPriorOSS()
			} else {
				debug.Info(`Cannot confirm if entry is deletable when Doctor is disabled: %s`, seg.String())
			}
			return false
		}
		return true
	}
	return false
}

// IsUpdatable returns true if this OSS record can be updated in the Catalog (i.e. it is a real standalone OSS record, not a placeholder for a native Catalog entry)
func (seg *SegmentInfo) IsUpdatable() bool {
	return seg.OSSSegment.IsUpdatable()
}

// IsValid returns true if this SegmentInfo record represents a valid entry that
// can be updated or merged into the Catalog
func (seg *SegmentInfo) IsValid() bool {
	return seg.OSSSegment.SegmentID != ""
}

// Header returns a short header representing this SegmentInfo record, suitable for printing in logs and reports
func (seg *SegmentInfo) Header() string {
	result := strings.Builder{}
	result.WriteString(seg.OSSSegment.Header())
	if seg.OSSValidation != nil {
		result.WriteString(seg.OSSValidation.Header())
	}
	return result.String()
}

// Details returns a long text representing this SegmentInfo record (including a Header), suitable for printing in logs and reports
func (seg *SegmentInfo) Details() string {
	result := strings.Builder{}
	result.WriteString(seg.OSSSegment.Header())
	if seg.OSSValidation != nil {
		result.WriteString(seg.OSSValidation.Details())
	}
	return result.String()
}

// Diffs returns a list of differences between the new and priorOSS versions of the OSSSegment record in this SegmentInfo
func (seg *SegmentInfo) Diffs() *compare.Output {
	out := compare.Output{}
	compare.DeepCompare("before", &seg.PriorOSS, "after", &seg.OSSSegment, &out)
	if (seg.OSSValidation == nil && seg.PriorOSSValidation != nil) || (seg.OSSValidation != nil && seg.OSSValidation.Checksum() != seg.PriorOSSValidationChecksum) {
		compare.DeepCompare("before.OSSValidation", seg.PriorOSSValidation, "after.OSSValidation", seg.OSSValidation, &out)
		//out.AddOtherDiff("OSSValidation updated")
	}
	return &out
}

// GlobalSortKey returns a global key for sorting across all ossmerge.Model entries of all types
func (seg *SegmentInfo) GlobalSortKey() string {
	return "1." + seg.String()
}

// GetAccessor returns the Accessor function for a given sub-record within the SegmentInfo
func (seg *SegmentInfo) GetAccessor(t reflect.Type) ossmergemodel.AccessorFunc {
	panic("GetAccessor() not defined for type ossmerge.SegmentInfo")
}

// GetOSSValidation returns the OSSValidation record in this SegmentInfo, or nil if there is none
func (seg *SegmentInfo) GetOSSValidation() *ossvalidation.OSSValidation {
	return seg.OSSValidation
}

// SkipMerge returns true if we should skip the OSS merge for this record and simply copy it from
// the prior OSS record.
// If not empty, the "reason" string provides details on why this record should or should not be merged
func (seg *SegmentInfo) SkipMerge() (skip bool, reason string) {
	return ossmergemodel.SkipMerge(seg)
}

var _ ossmergemodel.Model = &SegmentInfo{} // verify
