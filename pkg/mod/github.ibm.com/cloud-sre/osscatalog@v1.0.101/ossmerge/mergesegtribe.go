package ossmerge

import (
	"fmt"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/scorecardv1"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// Functions for merging Segment and Tribe information
//
// The methods in this file are mostly placeholders, since currently ScorecardV1 is the single
// source of information for segments and tribes.

// LoadScorecardV1SegmentTribes loads all the Segment and Tribe information from ScorecardV1 into the model
func LoadScorecardV1SegmentTribes() error {
	var numSegments = 0
	var numTribes = 0
	var numSegOrTribeEntries = 0
	debug.Info("Loading one batch of Segment/Tribe entries from ScorecardV1 (%d entries so far)", numSegOrTribeEntries)
	err := scorecardv1.ListSegments(func(s *scorecardv1.SegmentResource) error {
		numSegments++
		numSegOrTribeEntries++
		if (numSegOrTribeEntries % 30) == 0 {
			debug.Info("Loading one batch of Segment/Tribe entries from ScorecardV1 (%d entries so far)", numSegOrTribeEntries)
		}
		segID := s.GetSegmentID()
		segment, _ := LookupSegment(segID, true)
		if !segment.HasSourceScorecardV1() {
			segment.SetName(s.Name)
			segment.SourceScorecardV1 = *s
			err := scorecardv1.ListTribes(s, func(t *scorecardv1.TribeResource) error {
				numTribes++
				numSegOrTribeEntries++
				if (numSegOrTribeEntries % 30) == 0 {
					debug.Info("Loading one batch of Segment/Tribe entries from ScorecardV1 (%d entries so far)", numSegOrTribeEntries)
				}
				tribeID := t.GetTribeID()
				tribe, _ := segment.LookupTribe(tribeID, true)
				if !tribe.HasSourceScorecardV1() {
					tribe.SetName(t.Name)
					tribe.SourceScorecardV1 = *t
				} else {
					debug.PrintError(`Duplicate entry for Tribe from %s from ScorecardV1 (ignoring second entry) Tribe(%s[%s]) - prior=%s`, segment.String(), s.Name, segID, tribe.String())
				}
				return nil
			})
			if err != nil {
				return debug.WrapError(err, "Error listing Tribes from %s from ScorecardV1", segment.String())
			}
		} else {
			debug.PrintError(`Duplicate entry for Segment from ScorecardV1 (ignoring second entry) Segment(%s[%s]) - prior=%s`, s.Name, segID, segment.String())
		}
		return nil
	})
	if err != nil {
		return (debug.WrapError(err, "Error listing Segments from ScorecardV1"))
	}
	debug.Info("Read %d Segment and %d Tribe entries from ScorecardV1", numSegments, numTribes)
	return nil
}

// ConstructOSSSegment consolidates and merges the information from availables sources in a
// SegmentInfo record and constructs the OSSSegment record.
// Currently, ScorecardV1 is the single source
func (seg *SegmentInfo) ConstructOSSSegment() {
	seg.OSSSegment.SchemaVersion = ossrecord.OSSCurrentSchema

	// Set OSSTags. PriorOSS is the only possible source. No OSS tags or segment type info in ScorecardV1.
	if seg.HasPriorOSS() {
		seg.OSSSegment.OSSTags = seg.GetPriorOSS().OSSTags.Copy()
		seg.OSSValidation.AddSource(string(seg.GetPriorOSS().SegmentID), ossvalidation.PRIOROSS)
		if !ossrunactions.Doctor.IsEnabled() && seg.GetPriorOSS().OSSOnboardingPhase == "" {
			seg.OSSValidation.AddSource(string(seg.GetPriorOSS().SegmentID), ossvalidation.SCORECARDV1DISABLED)
		}
		seg.OSSSegment.OSSOnboardingPhase = seg.GetPriorOSS().OSSOnboardingPhase
		seg.OSSSegment.OSSOnboardingApprover = seg.GetPriorOSS().OSSOnboardingApprover
		seg.OSSSegment.OSSOnboardingApprovalDate = seg.GetPriorOSS().OSSOnboardingApprovalDate
	}

	if seg.HasSourceScorecardV1() {
		seg.OSSValidation.AddSource(string(seg.GetSourceScorecardV1().GetSegmentID()), ossvalidation.SCORECARDV1)
	}

	// SegmentID is already set during loading of the SegmentInfo

	// set SegmentType
	if seg.HasPriorOSS() {
		if seg.OSSSegment.OSSTags.Contains(osstags.TypeGaaS) {
			seg.OSSSegment.SegmentType = ossrecord.SegmentTypeGaaS
		} else {
			seg.OSSSegment.SegmentType = seg.GetPriorOSS().SegmentType
		}
	}
	if seg.OSSSegment.SegmentType == "" {
		// Default value - assume SegmentTypeIBMPublicCloud unless told otherwise
		seg.OSSSegment.SegmentType = ossrecord.SegmentTypeIBMPublicCloud
	}

	// set DisplayName
	if seg.HasSourceScorecardV1() {
		// SetName was already called while loading from ScorecardV1
	} else {
		if seg.HasPriorOSS() {
			seg.SetName(seg.GetPriorOSS().DisplayName)
		}
	}

	// set Owner
	if seg.HasSourceScorecardV1() {
		sc := seg.GetSourceScorecardV1()
		seg.OSSSegment.Owner = ossrecord.Person{Name: sc.MgmtName, W3ID: sc.MgmtEmail}
	}
	if !seg.OSSSegment.Owner.IsValid() && seg.HasPriorOSS() && !seg.GetPriorOSS().Owner.IsEmpty() {
		if !seg.OSSSegment.Owner.IsEmpty() {
			seg.OSSValidation.AddIssue(ossvalidation.SEVERE, "Overwriting invalid Owner info with data from prior OSS record", "invalid=%+v / prior=%+v", seg.OSSSegment.Owner, seg.GetPriorOSS().Owner).TagRunAction(ossrunactions.Tribes).TagDataMismatch()
			//debug.PrintError(`OSSSegment(type=%-18s) overwriting invalid Owner info with data from prior OSS record: %s (invalid=%+v / prior=%+v)`, seg.OSSSegment.SegmentType, seg.String(), seg.OSSSegment.Owner, seg.GetPriorOSS().Owner)
		}
		seg.OSSSegment.Owner = seg.GetPriorOSS().Owner
	}
	if !seg.OSSSegment.Owner.IsValid() {
		seg.OSSValidation.AddIssue(ossvalidation.WARNING, "Segment does not have a valid Owner", "%+v", seg.OSSSegment.Owner).TagRunAction(ossrunactions.Tribes).TagDataMissing()
		//debug.Warning(`OSSSegment(type=%-18s) does not contain a valid Owner            : %s (%+v)`, seg.OSSSegment.SegmentType, seg.String(), seg.OSSSegment.Owner)
		//		seg.forceDelete = true // XXX Cannot force delete because too many segments are non-compliant
	}

	// set Technical Contact
	if seg.HasSourceScorecardV1() {
		sc := seg.GetSourceScorecardV1()
		if sc.TechEmail != "" && strings.ToLower(sc.TechEmail) != "none" {
			seg.OSSSegment.TechnicalContact = ossrecord.Person{Name: sc.TechName, W3ID: sc.TechEmail}
		}
	}
	if !seg.OSSSegment.TechnicalContact.IsValid() && seg.HasPriorOSS() && !seg.GetPriorOSS().TechnicalContact.IsEmpty() {
		if !seg.OSSSegment.TechnicalContact.IsEmpty() {
			seg.OSSValidation.AddIssue(ossvalidation.SEVERE, "Overwriting invalid Technical Contact info with data from prior OSS record", "invalid=%+v / prior=%+v", seg.OSSSegment.TechnicalContact, seg.GetPriorOSS().TechnicalContact).TagRunAction(ossrunactions.Tribes).TagDataMismatch()
			//debug.PrintError(`OSSSegment(type=%-18s) overwriting invalid Technical Contact info with data from prior OSS record: %s (invalid=%+v / prior=%+v)`, seg.OSSSegment.SegmentType, seg.String(), seg.OSSSegment.TechnicalContact, seg.GetPriorOSS().TechnicalContact)
		}
		seg.OSSSegment.TechnicalContact = seg.GetPriorOSS().TechnicalContact
	}
	if !seg.OSSSegment.TechnicalContact.IsValid() {
		seg.OSSValidation.AddIssue(ossvalidation.WARNING, "Segment does not have a valid Technical Contact", "%+v", seg.OSSSegment.TechnicalContact).TagRunAction(ossrunactions.Tribes).TagDataMissing()
		//debug.Warning(`OSSSegment(type=%-18s) does not contain a valid Technical Contact: %s (%+v)`, seg.OSSSegment.SegmentType, seg.String(), seg.OSSSegment.TechnicalContact)
		//		seg.forceDelete = true // XXX Cannot force delete because too many segments are non-compliant
	}

	// set ERCAApprovers
	if len(seg.OSSSegment.ERCAApprovers) == 0 && seg.HasPriorOSS() {
		seg.OSSSegment.ERCAApprovers = seg.GetPriorOSS().ERCAApprovers
	}

	// set ChangeCommApprovers
	if len(seg.OSSSegment.ChangeCommApprovers) == 0 && seg.HasPriorOSS() {
		seg.OSSSegment.ChangeCommApprovers = seg.GetPriorOSS().ChangeCommApprovers
	}

	// Check consistency
	if seg.HasSourceScorecardV1() {
		sc := seg.GetSourceScorecardV1()
		_ = seg.CheckConsistency(sc.GetSegmentID(), sc.Name, true)
	} else if seg.HasPriorOSS() {
		prior := seg.GetPriorOSS()
		_ = seg.CheckConsistency(prior.SegmentID, prior.DisplayName, true)
	} else {
		_ = seg.CheckConsistency("", "", true)

	}

	// Record the IBMCloudDefaultSegment
	if seg.OSSSegment.OSSTags.Contains(osstags.IBMCloudDefaultSegment) /* || seg.OSSSegment.DisplayName == "Core Platform" /* XXX */ {
		IBMCloudDefaultSegment = append(IBMCloudDefaultSegment, seg)
		if len(IBMCloudDefaultSegment) > 1 {
			seg.OSSValidation.AddIssue(ossvalidation.SEVERE, fmt.Sprintf(`Found more than one OSSSegment tagged "%s"`, osstags.IBMCloudDefaultSegment), "other=%v", IBMCloudDefaultSegment).TagRunAction(ossrunactions.Tribes).TagDataMismatch()
			debug.PrintError(`Found more than one OSSSegment tagged "%s": %v`, osstags.IBMCloudDefaultSegment, IBMCloudDefaultSegment)
		}
	}

	if seg.OSSValidation != nil {
		seg.OSSValidation.SetSourceNameCanonical(string(seg.SegmentID))
		seg.OSSValidation.AddIssue(ossvalidation.INFO, "Merge Sources", "%v", seg.OSSValidation.AllSources()).TagSegTribes()
	}

	// Check for test records
	osstags.CheckOSSTestTag(&seg.DisplayName, &seg.OSSTags)

	// Take care of any side effects of IsDeletable() (e.g. new ValidationIssues)
	_ = seg.IsDeletable()

	// Sort the ValidationIssues and the tags
	if seg.OSSValidation != nil {
		seg.OSSValidation.Sort()
	}
	err := seg.OSSSegment.OSSTags.Validate(true)
	if err != nil {
		panic(err)
	}

	if seg.OSSSegment.OSSOnboardingPhase == ossrecord.EDIT {
		seg.OSSSegmentExtended.ResetForRMC()
	}
}

// ConstructOSSTribe consolidates and merges the information from availables sources in a
// TribeInfo record and constructs the OSSTribe record.
// Currently, ScorecardV1 is the single source
func (tr *TribeInfo) ConstructOSSTribe() {
	tr.OSSTribe.SchemaVersion = ossrecord.OSSCurrentSchema

	// Set OSSTags. PriorOSS is the only possible source. No OSS tags or segment type info in ScorecardV1.
	if tr.HasPriorOSS() {
		tr.OSSTribe.OSSTags = tr.GetPriorOSS().OSSTags.Copy()
		tr.OSSValidation.AddSource(string(tr.GetPriorOSS().TribeID), ossvalidation.PRIOROSS)
		if !ossrunactions.Doctor.IsEnabled() && tr.GetPriorOSS().OSSOnboardingPhase == "" {
			tr.OSSValidation.AddSource(string(tr.GetPriorOSS().TribeID), ossvalidation.SCORECARDV1DISABLED)
		}
		tr.OSSTribe.OSSOnboardingPhase = tr.GetPriorOSS().OSSOnboardingPhase
		tr.OSSTribe.OSSOnboardingApprover = tr.GetPriorOSS().OSSOnboardingApprover
		tr.OSSTribe.OSSOnboardingApprovalDate = tr.GetPriorOSS().OSSOnboardingApprovalDate
	}

	if tr.HasSourceScorecardV1() {
		tr.OSSValidation.AddSource(string(tr.GetSourceScorecardV1().GetTribeID()), ossvalidation.SCORECARDV1)
	}

	// TribeID and SegmentID are already set during loading of the TribeInfo

	// set DisplayName
	if tr.HasSourceScorecardV1() {
		// SetName was already called while loading from ScorecardV1
	} else {
		if tr.HasPriorOSS() {
			tr.SetName(tr.GetPriorOSS().DisplayName)
		}
	}

	// set Owner
	if tr.HasSourceScorecardV1() {
		sc := tr.GetSourceScorecardV1()
		tr.OSSTribe.Owner = ossrecord.Person{Name: sc.OwnerContact, W3ID: sc.OwnerEmail}
	}
	if !tr.OSSTribe.Owner.IsValid() && tr.HasPriorOSS() && !tr.GetPriorOSS().Owner.IsEmpty() {
		if !tr.OSSTribe.Owner.IsEmpty() {
			tr.OSSValidation.AddIssue(ossvalidation.SEVERE, "Overwriting invalid Owner info with data from prior OSS record", "invalid=%+v / prior=%+v", tr.OSSTribe.Owner, tr.GetPriorOSS().Owner).TagRunAction(ossrunactions.Tribes).TagDataMismatch()
			//debug.Warning(`OSSTribe overwriting invalid Owner info with data from prior OSS record: %s (invalid=%+v / prior=%+v)`, tr.String(), tr.OSSTribe.Owner, tr.GetPriorOSS().Owner)
		}
		tr.OSSTribe.Owner = tr.GetPriorOSS().Owner
	}
	if !tr.OSSTribe.Owner.IsValid() {
		tr.OSSValidation.AddIssue(ossvalidation.WARNING, "Tribe does not have a valid Owner", "%+v", tr.OSSTribe.Owner).TagRunAction(ossrunactions.Tribes).TagDataMissing()
		//debug.Warning(`OSSTribe does not contain a valid Owner            : %s (%+v)`, tr.String(), tr.OSSTribe.Owner)
	}

	// set ChangeApprovers
	if tr.HasSourceScorecardV1() {
		sc := tr.GetSourceScorecardV1()
		tr.OSSTribe.ChangeApprovers = make([]*ossrecord.PersonListEntry, 0, len(sc.ChangeApprovers))
		for _, ca := range sc.ChangeApprovers {
			tags := make([]string, len(ca.Tags))
			copy(tags, ca.Tags)
			tr.OSSTribe.ChangeApprovers = append(tr.OSSTribe.ChangeApprovers, &ossrecord.PersonListEntry{Member: ca.Member, Tags: tags})
		}
	}
	if len(tr.OSSTribe.ChangeApprovers) == 0 && tr.HasPriorOSS() {
		tr.OSSTribe.ChangeApprovers = tr.GetPriorOSS().ChangeApprovers
	}

	// Check consistency
	if tr.HasSourceScorecardV1() {
		sc := tr.GetSourceScorecardV1()
		_ = tr.CheckConsistency(sc.GetTribeID(), sc.Name, true)
	} else if tr.HasPriorOSS() {
		prior := tr.GetPriorOSS()
		_ = tr.CheckConsistency(prior.TribeID, prior.DisplayName, true)
	} else {
		_ = tr.CheckConsistency("", "", true)
	}

	if tr.OSSValidation != nil {
		tr.OSSValidation.SetSourceNameCanonical(string(tr.TribeID))
		tr.OSSValidation.AddIssue(ossvalidation.INFO, "Merge Sources", "%v", tr.OSSValidation.AllSources()).TagSegTribes()
	}

	// Check for test records
	osstags.CheckOSSTestTag(&tr.DisplayName, &tr.OSSTags)

	// Take care of any side effects of IsDeletable() (e.g. new ValidationIssues)
	_ = tr.IsDeletable()

	// Sort the ValidationIssues and the tags
	if tr.OSSValidation != nil {
		tr.OSSValidation.Sort()
	}
	err := tr.OSSTribe.OSSTags.Validate(true)
	if err != nil {
		panic(err)
	}

	if tr.OSSTribe.OSSOnboardingPhase == ossrecord.EDIT {
		tr.OSSTribeExtended.ResetForRMC()
	}
}

// IBMCloudDefaultSegment keeps track of the SegmentInfo that is tagged with the osstags.IBMCloudDefaultSegment tag,
// to be used as default segment for all environment records that are part of IBM Public Cloud or legacy Dedicated/Local
// environments (but not GaaS)
// This variable is a slice to allow for the fact that we might find more than one tagge SegmentInfo, but it is
// an error if there is not exactly one
var IBMCloudDefaultSegment []*SegmentInfo
