package ossmerge

import (
	"fmt"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/clearinghouse"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/monitoringinfo"
	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
)

type mergePhaseType string

const (
	mergePhaseStart           mergePhaseType = ""
	mergePhaseSegments        mergePhaseType = "SEGMENTS"
	mergePhaseEnvironmentsOne mergePhaseType = "ENVIRONMENTS-ONE"
	mergePhaseEnvironmentsTwo mergePhaseType = "ENVIRONMENTS-TWO"
	mergePhaseServicesOne     mergePhaseType = "SERVICES-ONE"
	mergePhaseServicesTwo     mergePhaseType = "SERVICES-TWO"
	mergePhaseServicesThree   mergePhaseType = "SERVICES-THREE"
	mergePhaseFinalized       mergePhaseType = "FINALIZED"
)

var globalMergePhase = mergePhaseStart

// advanceGlobalMergePhase advances to the next global merge phase, while performing sanity checks to ensure that the phases are executed in the right sequence
func advanceGlobalMergePhase(nextPhase mergePhaseType) {
	debug.Debug(debug.Merge, `ossmerge.advanceGlobalMergePhase(%s) -- previous globalMergePhase=%s`, nextPhase, globalMergePhase)
	switch nextPhase {
	case mergePhaseSegments:
		if globalMergePhase != mergePhaseStart {
			panic(fmt.Sprintf(`ossmerge.advanceGlobalMergePhase(): cannot advance to merge phase Segments - expecting previous phase %v got %v`, mergePhaseStart, globalMergePhase))
		}
	case mergePhaseEnvironmentsOne:
		if globalMergePhase != mergePhaseSegments {
			panic(fmt.Sprintf(`ossmerge.advanceGlobalMergePhase(): cannot advance to merge phase EnvironmentsOne - expecting previous phase %v got %v`, mergePhaseSegments, globalMergePhase))
		}
	case mergePhaseEnvironmentsTwo:
		if globalMergePhase != mergePhaseEnvironmentsOne {
			panic(fmt.Sprintf(`ossmerge.advanceGlobalMergePhase(): cannot advance to merge phase EnvironmentsTwo - expecting previous phase %v got %v`, mergePhaseEnvironmentsOne, globalMergePhase))
		}
	case mergePhaseServicesOne:
		if globalMergePhase != mergePhaseEnvironmentsTwo {
			panic(fmt.Sprintf(`ossmerge.advanceGlobalMergePhase(): cannot advance to merge phase ServicesOne - expecting previous phase %v got %v`, mergePhaseEnvironmentsTwo, globalMergePhase))
		}
	case mergePhaseServicesTwo:
		if globalMergePhase != mergePhaseServicesOne {
			panic(fmt.Sprintf(`ossmerge.advanceGlobalMergePhase(): cannot advance to merge phase ServicesTwo - expecting previous phase %v got %v `, mergePhaseServicesOne, globalMergePhase))
		}
	case mergePhaseServicesThree:
		if globalMergePhase != mergePhaseServicesTwo {
			panic(fmt.Sprintf(`ossmerge.advanceGlobalMergePhase(): cannot advance to merge phase ServicesThree - expecting previous phase %v got %v`, mergePhaseServicesTwo, globalMergePhase))
		}
	case mergePhaseFinalized:
		if globalMergePhase != mergePhaseServicesThree {
			panic(fmt.Sprintf(`ossmerge.advanceGlobalMergePhase(): cannot advance to merge phase Finalized - expecting previous phase %v got %v`, mergePhaseServicesThree, globalMergePhase))
		}
	default:
		panic(fmt.Sprintf(`ossmerge.advanceGlobalMergePhase(): unexpected next merge phase %v - (current phase %v)`, nextPhase, globalMergePhase))
	}
	globalMergePhase = nextPhase
}

// checkMergePhase verifies that we are in the correct phase and panics if not
func (si *ServiceInfo) checkMergePhase(expectedPhase mergePhaseType) {
	if globalMergePhase != expectedPhase {
		panic(fmt.Sprintf(`ossmerge.checkMergePhase(): expecting global merge phase %v got %v for entry ServiceInfo(%s)`, expectedPhase, globalMergePhase, si.String()))
	}
	if si.mergeWorkArea.mergePhase != expectedPhase {
		panic(fmt.Sprintf(`ossmerge.checkMergePhase(): expecting entry merge phase %v got %v for entry ServiceInfo(%s)`, expectedPhase, si.mergeWorkArea.mergePhase, si.String()))
	}
}

// checkGlobalMergePhaseMultiple verifies that we are in one of the given global merge phases and panics if not
// Needed during phase TWO because we might be cross-referencing entries that are stil in phase ONE
func (si *ServiceInfo) checkGlobalMergePhaseMultiple(expectedGlobalPhases ...mergePhaseType) {
	for _, e := range expectedGlobalPhases {
		if globalMergePhase == e {
			return
		}
	}
	panic(fmt.Sprintf(`ossmerge.checkGlobalMergePhaseMultiple(): expecting global merge phase %v got %v for entry ServiceInfo(%s)`, expectedGlobalPhases, globalMergePhase, si.String()))
}

// checkEntryMergePhaseMultiple verifies that this ServiceInfo is in one of the given phases and panics if not
// Needed during phase TWO because we might be cross-referencing entries that are stil in phase ONE
func (si *ServiceInfo) checkEntryMergePhaseMultiple(expectedEntryPhases ...mergePhaseType) {
	for _, e := range expectedEntryPhases {
		if si.mergeWorkArea.mergePhase == e {
			return
		}
	}
	panic(fmt.Sprintf(`ossmerge.checkEntryMergePhaseMultiple(): expecting entry merge phase %v got %v for entry ServiceInfo(%s)`, expectedEntryPhases, si.mergeWorkArea.mergePhase, si.String()))
}

// advanceMergePhase advances this ServiceInfo record to the next merge phase, while performing sanity checks to ensure that the phases are executed in the right sequence
func (si *ServiceInfo) advanceMergePhase(nextPhase mergePhaseType) {
	debug.Debug(debug.Merge, `ossmerge.advanceMergePhase(%s, ServiceInfo(%s)) -- current globalMergePhase=%s / previous entry mergePhase=%s`, nextPhase, si.String(), globalMergePhase, si.mergeWorkArea.mergePhase)
	switch nextPhase {
	case mergePhaseServicesOne:
		if globalMergePhase != mergePhaseServicesOne {
			panic(fmt.Sprintf(`ossmerge.advanceMergePhase(): expecting global merge phase %v got %v for entry ServiceInfo(%s)`, mergePhaseServicesOne, globalMergePhase, si.String()))
		}
		if si.mergeWorkArea.mergePhase != mergePhaseStart { // Note no mergePhaseSegments  or mergePhaseEnvironments for ServiceInfo
			panic(fmt.Sprintf(`ossmerge.advanceMergePhase(): expecting entry merge phase %v got %v for entry ServiceInfo(%s)`, mergePhaseStart, si.mergeWorkArea.mergePhase, si.String()))
		}
	case mergePhaseServicesTwo:
		if globalMergePhase != mergePhaseServicesTwo {
			panic(fmt.Sprintf(`ossmerge.advanceMergePhase(): expecting global merge phase %v got %v for entry ServiceInfo(%s)`, mergePhaseServicesTwo, globalMergePhase, si.String()))
		}
		if si.mergeWorkArea.mergePhase != mergePhaseServicesOne {
			panic(fmt.Sprintf(`ossmerge.advanceMergePhase(): expecting entry merge phase %v got %v for entry ServiceInfo(%s)`, mergePhaseServicesOne, si.mergeWorkArea.mergePhase, si.String()))
		}
	case mergePhaseServicesThree:
		if globalMergePhase != mergePhaseServicesThree {
			panic(fmt.Sprintf(`ossmerge.advanceMergePhase(): expecting global merge phase %v got %v for entry ServiceInfo(%s)`, mergePhaseServicesThree, globalMergePhase, si.String()))
		}
		if si.mergeWorkArea.mergePhase != mergePhaseServicesTwo {
			panic(fmt.Sprintf(`ossmerge.advanceMergePhase(): expecting entry merge phase %v got %v for entry ServiceInfo(%s)`, mergePhaseServicesTwo, si.mergeWorkArea.mergePhase, si.String()))
		}
	case mergePhaseFinalized:
		if globalMergePhase != mergePhaseServicesThree { // Note we advance ServiceInfo to mergePhaseFinalized during the laat step in global mergePhaseThree
			panic(fmt.Sprintf(`ossmerge.advanceMergePhase(): expecting global merge phase %v got %v for entry ServiceInfo(%s)`, mergePhaseServicesThree, globalMergePhase, si.String()))
		}
		if si.mergeWorkArea.mergePhase != mergePhaseServicesThree {
			panic(fmt.Sprintf(`ossmerge.advanceMergePhase(): expecting entry merge phase %v got %v for entry ServiceInfo(%s)`, mergePhaseServicesThree, si.mergeWorkArea.mergePhase, si.String()))
		}
	default:
		panic(fmt.Sprintf(`ossmerge.advanceMergePhase(): unexpected next merge phase %v - (current phase %v) for entry ServiceInfo(%s)`, nextPhase, si.mergeWorkArea.mergePhase, si.String()))
	}
	si.mergeWorkArea.mergePhase = nextPhase
}

// Finalized return true if this ServiceInfo record has completed all phases of ossmerge
func (si *ServiceInfo) Finalized() bool {
	return si.mergeWorkArea.mergePhase == mergePhaseFinalized
}

// SetFinalized forces the merge phase of this ServiceInfo record to "Finalized", regardless of its current phase.
// DANGEROUS - used only for testing
func (si *ServiceInfo) SetFinalized() {
	debug.Debug(debug.Merge, `ossmerge.SetFinalized(ServiceInfo(%s)) -- current globalMergePhase=%s / previous entry mergePhase=%s`, si.String(), globalMergePhase, si.mergeWorkArea.mergePhase)
	if globalMergePhase != mergePhaseStart && globalMergePhase != mergePhaseFinalized {
		panic(fmt.Sprintf(`ossmerge.SetFinalized(): cannot set global merge phase to "Finalized" - expecting previous phase %v got %v (for entry ServiceInfo(%s))`, mergePhaseStart, globalMergePhase, si.String()))
	}
	if si.mergeWorkArea.mergePhase != mergePhaseStart {
		panic(fmt.Sprintf(`ossmerge.SetFinalized(): cannot set entry merge phase to "Finalized" - expecting previous phase %v got %v for entry ServiceInfo(%s)`, mergePhaseStart, si.mergeWorkArea.mergePhase, si.String()))
	}
	globalMergePhase = mergePhaseFinalized
	si.mergeWorkArea.mergePhase = mergePhaseFinalized
}

// mergePhaseOne consolidates and merges the information from various sources in a
// ServiceInfo record and constructs the OSSService and related OSS records for that service/component.
// First phase of the overall merge.
// This phase operates on individual records; while operating in one record, we cannot assume anything about the state of
// any other records.
func (si *ServiceInfo) mergePhaseOne() {
	si.advanceMergePhase(mergePhaseServicesOne)

	if si.DuplicateOf != "" {
		debug.Debug(debug.Merge, `ossmerge.PhaseOne(): Skipping merge of entry "%s" that is marked as a duplicate of entry "%s"`, si.GetServiceName(), si.DuplicateOf)
		return
	}

	// Create a OSSMergeControl record. Needed only the first time, if we did not load
	// a prior OSS record from the Catalog
	if si.OSSMergeControl == nil {
		si.OSSMergeControl = ossmergecontrol.New("")
	}

	// Start with the Prior OSS entry as a baseline.
	// Note that we may subsequently overwrite some of the attributes with new values based on the merge.
	if si.HasPriorOSS() {
		si.OSSService = *si.GetPriorOSS().DeepCopy()
	}

	// Normalize the OSSMergeControl tags
	err := si.OSSMergeControl.OSSTags.Validate(false)
	if err != nil {
		debug.Warning("Found invalid OSSTags in OSSMergeControl for entry %s: %v", si.String(), err)
	}

	if si.HasPriorOSS() && si.GetPriorOSS().GeneralInfo.OSSOnboardingPhase != "" {
		prior := si.GetPriorOSS()
		// Copy the OSSTags up front to make sure we have them available before any other merge action that they might influence

		// Copy any tags not coming from OSSMergeControl from the PriorOSS record written by RMC,
		// but remove status/generated tags
		si.OSSService.GeneralInfo.OSSTags = nil
		for _, t := range prior.GeneralInfo.OSSTags {
			if (t == osstags.OSSTest || t == osstags.OSSOnly || t.IsRMCManaged()) && !t.IsStatusTag() {
				si.OSSService.GeneralInfo.OSSTags.AddTag(t)
			}
		}
		// Must remove any OSSTags from OSSMergeControl that are now superseded by RMC
		for _, t := range si.OSSMergeControl.OSSTags {
			if t.IsRMCManaged() {
				si.OSSMergeControl.OSSTags.RemoveTag(t)
				si.OSSMergeControl.IgnoredOSSTags.AddTag(t)
				// Do not copy into main OSS record
			} else {
				si.OSSService.GeneralInfo.OSSTags.AddTag(t)
			}
		}
		if len(si.OSSMergeControl.IgnoredOSSTags) > 0 {
			si.AddValidationIssue(ossvalidation.MINOR, "One or more OSSMergeControl OSSTags ignored because this record is managed by RMC", "%s", si.OSSMergeControl.IgnoredOSSTags.String()).TagControlOverride()
		}
		// Must remove all Overrides from OSSMergeControl
		if si.OSSMergeControl.Overrides != nil {
			if si.OSSMergeControl.IgnoredOverrides == nil {
				si.OSSMergeControl.IgnoredOverrides = make(map[string]interface{})
			}
			for k, v := range si.OSSMergeControl.Overrides {
				si.OSSMergeControl.IgnoredOverrides[k] = v
			}
			si.OSSMergeControl.Overrides = nil
		}
		if len(si.OSSMergeControl.IgnoredOverrides) > 0 {
			si.AddValidationIssue(ossvalidation.MINOR, "One or more OSSMergeControl Overrides ignored because this record is managed by RMC", "%v", si.OSSMergeControl.IgnoredOverrides).TagControlOverride()
		}
	} else {
		si.OSSService.GeneralInfo.OSSOnboardingPhase = ""
		// Copy the OSSTags up front to make sure we have them available before any other merge action that they might influence
		// Obliterate any other OSSTags found in the PriorOSS record
		si.OSSService.GeneralInfo.OSSTags = si.OSSMergeControl.OSSTags.Copy()
		if len(si.OSSMergeControl.IgnoredOSSTags) > 0 {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Record has one or more OSSMergeControl OSSTags ignored, but is not managed by RMC", "%s", si.OSSMergeControl.IgnoredOSSTags.String()).TagControlOverride()
		}
		if len(si.OSSMergeControl.IgnoredOverrides) > 0 {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Record has one or more OSSMergeControl Overrides ignored, but is not managed by RMC", "%v", si.OSSMergeControl.IgnoredOverrides).TagControlOverride()
		}
	}

	if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) {
		si.OSSMergeControl.OSSTags.AddTag(osstags.OSSTest)
	}
	if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSOnly) {
		si.OSSMergeControl.OSSTags.AddTag(osstags.OSSOnly)
	}
	si.checkExpiredOSSTags()

	si.OSSService.SchemaVersion = ossrecord.OSSCurrentSchema

	// Pick the best record from each source, if there is more than one
	si.checkAllAdditionalSources()

	// Get rid of the Main Catalog entry if we are not supposed to include it in the merge
	si.checkIgnoreMainCatalog()

	si.OSSService.ReferenceResourceName = si.mergeReferenceResourceName()

	// Note: if the entry was private in Catalog and not found anywhere else, we might end-up with no name at all
	if si.OSSService.ReferenceResourceName == "" {
		debug.Debug(debug.Merge, `ossmerge.PhaseOne(): Skipping remaining merge of entry "%s" that has no valid name`, si.String())
		return
	}

	var skip bool
	var reason string
	if skip, reason = si.SkipMerge(); skip {
		if reason != "" {
			debug.Info("Copying entire Prior OSSService record (no merge): %s: %s", reason, si.String())
		} else {
			panic(fmt.Sprintf("No reason provided for copying entire Prior OSSService record: %s", si.String()))
		}
		if !si.HasPriorOSS() {
			panic(fmt.Sprintf("SkipMerge=true but no PriorOSS record: %s: %s", reason, si.String()))
		}
		// re-copy the PriorOSS() record wholesale (note we probably already copied it earlier, but we want to reset the OSSTags)
		si.OSSService = *si.GetPriorOSS().DeepCopy()
		si.registerStatusCategoryParent()
		if si.HasPriorOSSValidation() {
			si.OSSValidation = si.GetPriorOSSValidation()
		}
		//si.AddValidationIssue(ossvalidation.INFO, "OSS-only or OSS-test entry copied entirely from Prior OSS record without a full merge - Validation Issues maybe be incomplete", "").TagConsistency()
		return
	}
	if reason != "" {
		debug.Warning("Force merging of OSSService record: %s: %s", reason, si.String())
	}

	// Update the CanonicalName in the OSSMergeControl
	// (we got it either from a prior OSS record, or we initialize it now if this is the first time we create a OSS record)
	if si.OSSMergeControl.CanonicalName != "" {
		if si.OSSMergeControl.CanonicalName != string(si.OSSService.ReferenceResourceName) {
			// TODO: How should we handle a change of ReferenceResourceName for an existing OSS record?
			panic(fmt.Sprintf(`Mismatched prior CanonicalName "%s" in OSS record for "%s"`, si.OSSMergeControl.CanonicalName, si.OSSService.ReferenceResourceName))
		}
	} else {
		si.OSSMergeControl.CanonicalName = string(si.OSSService.ReferenceResourceName)
	}

	si.OSSService.ReferenceDisplayName = si.mergeReferenceDisplayName()

	// Must be called before we compute the EntryType, because we may use the Segment Info to determine the EntryType (for GaaS segments)
	si.mergeOwnership()

	si.OSSService.GeneralInfo.EntryType = si.mergeEntryType()

	// Must be called after we determined the EntryType
	si.OSSService.GeneralInfo.OperationalStatus = si.mergeOperationalStatus()
	si.OSSService.GeneralInfo.FutureOperationalStatus = si.MergeValues("FutureOperationalStatus", SeverityIfMissing{ossvalidation.IGNORE}, SeverityIfMismatch{ossvalidation.WARNING}, OverrideProperty{"GeneralInfo.FutureOperationalStatus"}, RMCOSS{si.PriorOSS.GeneralInfo.FutureOperationalStatus}, PriorOSS{si.PriorOSS.GeneralInfo.FutureOperationalStatus}).(ossrecord.OperationalStatus)
	/* TODO: decide if we should clear FutureOperationalStatus: https://github.ibm.com/cloud-sre/osscatalog/issues/370
	if si.OSSService.GeneralInfo.FutureOperationalStatus == si.OSSService.GeneralInfo.OperationalStatus {
		si.OSSService.GeneralInfo.FutureOperationalStatus = ""
		si.AddValidationIssue(ossvalidation.INFO, "Resetting FutureOperationalStatus as it is identical to currentOperationalStatus", "%v", si.OSSService.GeneralInfo.OperationalStatus).TagCRN()
	}
	*/

	// Must be called after we determined the EntryType and OperationalStatus
	si.checkCatalogVisibility()
	si.mergeClientFacing()

	// Must happen before mergeSupport() and mergeOperations() because we need the value of SupportNotApplicable, OperationsNotApplicable
	si.mergeServiceNowInfo()

	si.mergeGeneralInfo()
	si.mergeSupport()
	si.mergeOperations()
	si.mergeCompliance() // mergeCompliance must happen before mergeStatusPage because we need the ServiceNowOnboarded flag
	si.mergeStatusPage()
	// TODO: Missing AdditionalContacts from ServiceNow
	// si.OSSService.AdditionalContacts = si.MergeValues("AdditionalContacts", SeverityIfMissing{IGNORE}, ServiceNow{si.SourceServiceNow.AdditionalContacts}, PriorOSS{si.PriorOSS.AdditionalContacts}).(string)
	si.mergeCatalogInfo()
	si.mergeProductInfoPhaseOne()

	// OSSMergeControl information from a pre-existing OSSValidation record is automatically copied when loading the ServiceInfo record

	// Check for missing sources and visibility. We cannot do this before we completed all the other merges.
	si.checkMissingSources()

	// TODO: Additional validation not directly related to merging, e.g. Client-facing status, IAM, RC, etc.

	si.checkOwnership()

	si.checkServiceNowEnrollment()

	si.checkIAM()

	if ossrunactions.Monitoring.IsEnabled() {
		si.OSSValidation.RecordRunAction(ossrunactions.Monitoring)
		monitoringinfo.MergeOneService(si)
	} else {
		si.OSSValidation.CopyRunAction(si.GetPriorOSSValidation(), ossrunactions.Monitoring)
		if si.HasPriorOSS() {
			si.OSSService.MonitoringInfo = si.GetPriorOSS().MonitoringInfo
		}
	}
}

// mergePhaseTwo continues the merge. It handles relationships between multiple ServiceInfo records,
// and relies on the fact that each individual record has already been constructed (without relationships) in Phase ServicesOne
func (si *ServiceInfo) mergePhaseTwo() {
	si.advanceMergePhase(mergePhaseServicesTwo)

	if skip, _ := si.SkipMerge(); skip {
		debug.Debug(debug.Merge, `ossmerge.PhaseTwo(): Skipping merge of OSSService record "%s" that has the SkipMerge flag`, si.String())
		return
	}

	if si.IsDeletable() {
		debug.Debug(debug.Merge, `ossmerge.PhaseTwo(): Skipping merge of entry "%s" that is deletable`, si.String())
		return
	}

	// Handle Catalog "composite" objects and their children
	si.checkComposite()

	// Look for issues related to the Status Page Category ID/Parent information
	si.checkStatusCategoryParent()

	// Setup the parent OSS entry link (if any)
	// XXX This must happen last, because other operations, for example checkComposite(), may contribute some parent information
	si.mergeParentResourceName()

	// Setup the ClearingHouse links
	si.mergeClearingHouse()

	// Figure out additional ProductInfo
	// Must be executed after all records have been scanned once (i.e. in Phase 2)
	// and after we have loaded all ClearingHouse links, if any
	si.mergeProductInfoPhaseTwo()

	// Preliminary computation of the CRN OSS validation status (after merge/validation)
	// XXX This must happen *before* checkPnPEnablement() because that check requires the CRN validation status
	validationCRNStatus := si.OSSValidation.SummaryCRNStatus(si.OSSService.GeneralInfo.OSSTags)
	si.OSSService.GeneralInfo.OSSTags.SetCRNStatus(validationCRNStatus)
}

// mergePhaseThree finalizes the OSS Validation issues and status associated with this ServiceInfo record.
// It must be run *after* PhaseOne and PhaseTwo has been executed for *all* available ServiceInfo records,
// in order to be able to detect issues that can only be detected by examining relations between multiple records.
func (si *ServiceInfo) mergePhaseThree() {
	si.advanceMergePhase(mergePhaseServicesThree)

	if skip, _ := si.SkipMerge(); !skip {
		if !si.IsDeletable() {
			// Figure out additional ProductInfo
			// Must be executed after all records have been scanned and all links have been set-up (i.e. in Phase 3)
			si.mergeProductInfoPhaseThree()

			// Get dependency information from ClearingHouse
			si.mergeCHDependencies()

			// Check if the entry can be enabled for the PnP status page
			// XXX Must happen after the computation of CRN validation status
			si.checkPnPEnablement()

			// Sanity check: make sure the CRN OSS validation status has not changed between the initial and final merge phases
			validationCRNStatus := si.OSSService.GeneralInfo.OSSTags.GetCRNStatus()
			newValidationCRNStatus := si.OSSValidation.SummaryCRNStatus(si.OSSService.GeneralInfo.OSSTags)
			if validationCRNStatus != newValidationCRNStatus {
				debug.Warning(`CRN Validation Status changed from "%s" to "%s" for entry "%s"`, validationCRNStatus, newValidationCRNStatus, si.GetServiceName())
				si.OSSService.GeneralInfo.OSSTags.SetCRNStatus(newValidationCRNStatus)
			}

			// Compute and set the overall validation status (after merge/validation)
			validationOverallStatus := si.OSSValidation.SummaryOverallStatus(si.OSSService.GeneralInfo.OSSTags)
			si.OSSService.GeneralInfo.OSSTags.SetOverallStatus(validationOverallStatus)
		} else {
			debug.Debug(debug.Merge, `ossmerge.PhaseThree(): Skipping merge of entry "%s" that is deletable`, si.String())
		}
	} else {
		debug.Debug(debug.Merge, `ossmerge.PhaseThree(): Skipping merge of OSSService record "%s" that has the SkipMerge flag`, si.String())
	}

	if si.OSSService.GeneralInfo.OSSOnboardingPhase == ossrecord.EDIT {
		si.OSSServiceExtended.ResetForRMC()
	}

	// Sort the ValidationIssues and the tags
	si.OSSValidation.Sort()
	err := si.OSSService.GeneralInfo.OSSTags.Validate(true)
	if err != nil {
		panic(err)
	}
}

// MergeAllEntries merges+validates all the entries loaded into the model (segments, tribes, services/components)
// It returns a list of all entries successfully merged
func MergeAllEntries(pattern *regexp.Regexp) (allMergedEntries []ossmergemodel.Model, err error) {
	allMergedEntries = make([]ossmergemodel.Model, 0, len(serviceByComparableName)+(3*len(segmentsByName)))

	// Merge phase: SEGMENTS
	if ossrunactions.Tribes.IsEnabled() {
		advanceGlobalMergePhase(mergePhaseSegments)
		debug.Info("Starting merge for Segments and Tribes")
		err := ListAllSegments(nil, func(seg *SegmentInfo) {
			// Set OSSTags
			// Do this early, so that we can refer to those tags to determine other attributes
			if seg.HasPriorOSS() {
				seg.OSSSegment.OSSTags = seg.GetPriorOSS().OSSTags.Copy()
			}
			// Check if this is a valid merge
			var skip bool
			var reason string
			if skip, reason = seg.SkipMerge(); skip {
				if reason != "" {
					debug.Info("Copying entire Prior OSSSegment record (no merge): %s: %s", reason, seg.String())
				} else {
					panic(fmt.Sprintf("No reason provided for copying entire Prior OSSSegment record: %s", seg.String()))
				}
				tags := seg.OSSSegment.OSSTags.Copy()
				seg.OSSSegment = *seg.GetPriorOSS().DeepCopy()
				seg.OSSSegment.OSSTags = tags
				name := seg.OSSSegment.DisplayName
				if !seg.HasSourceScorecardV1() && name != "" {
					seg.OSSSegment.DisplayName = ""
					seg.SetName(name)
				}
				seg.OSSValidation = seg.PriorOSSValidation
				if seg.OSSSegment.OSSOnboardingPhase == ossrecord.EDIT {
					seg.OSSSegmentExtended.ResetForRMC()
				}
			} else {
				if reason != "" {
					debug.Warning("Force merging of OSSSegment record: %s: %s", reason, seg.String())
				}
				seg.ConstructOSSSegment()
			}
			allMergedEntries = append(allMergedEntries, seg)
			err := seg.ListAllTribes(nil, func(tr *TribeInfo) {
				// Set OSSTags
				// Do this early, so that we can refer to those tags to determine other attributes
				if tr.HasPriorOSS() {
					tr.OSSTribe.OSSTags = tr.GetPriorOSS().OSSTags.Copy()
				}
				// Check if this is a valid merge
				var skip bool
				var reason string
				if skip, reason = tr.SkipMerge(); skip {
					if reason != "" {
						debug.Info("Copying entire Prior OSSTribe record (no merge): %s: %s", reason, tr.String())
					} else {
						panic(fmt.Sprintf("No reason provided for copying entire Prior OSSTribe record: %s", tr.String()))
					}
					tags := tr.OSSTribe.OSSTags.Copy()
					tr.OSSTribe = *tr.GetPriorOSS().DeepCopy()
					tr.OSSTribe.OSSTags = tags
					name := tr.OSSTribe.DisplayName
					if !tr.HasSourceScorecardV1() && name != "" {
						tr.OSSTribe.DisplayName = ""
						tr.SetName(name)
					}
					tr.OSSValidation = tr.PriorOSSValidation
					if tr.OSSTribe.OSSOnboardingPhase == ossrecord.EDIT {
						tr.OSSTribeExtended.ResetForRMC()
					}
				} else {
					if reason != "" {
						debug.Warning("Force merging of OSSTribe record: %s: %s", reason, tr.String())
					}
					tr.ConstructOSSTribe()
				}
				allMergedEntries = append(allMergedEntries, tr)
			})
			if err != nil {
				// TODO: Should exit the loop at this point
				//debug.PrintError("Error while merging all OSSTribe records for Segment \"%s\": %v", seg.String(), err)
				panic(fmt.Sprintf("Error while merging all OSSTribe records for Segment \"%s\": %v", seg.String(), err))
			}
		})
		if err != nil {
			return nil, debug.WrapError(err, "Error while merging all OSSSegment records for pattern \"%s\"", pattern)
		}
	} else {
		advanceGlobalMergePhase(mergePhaseSegments)
	}

	// Merge phase: ENVIRONMENTS ONE and TWO
	if ossrunactions.Environments.IsEnabled() || ossrunactions.EnvironmentsNative.IsEnabled() {
		advanceGlobalMergePhase(mergePhaseEnvironmentsOne)
		debug.Info("Starting merge phase ONE for Environments")
		err := ListAllEnvironments(nil, func(env *EnvironmentInfo) {
			// Set OSSTags
			// Do this early, so that we can refer to those tags to determine other attributes
			if env.HasPriorOSS() {
				env.OSSEnvironment.OSSTags = env.GetPriorOSS().OSSTags.Copy()
			}
			var skip bool
			var reason string
			if skip, reason = env.SkipMerge(); skip {
				if reason != "" {
					debug.Info("Copying entire Prior OSSEnvironment record (no merge): %s: %s", reason, env.String())
				} else {
					panic(fmt.Sprintf("No reason provided for copying entire Prior OSSEnvironment record: %s", env.String()))
				}
				tags := env.OSSEnvironment.OSSTags.Copy()
				env.OSSEnvironment = *env.GetPriorOSS().DeepCopy()
				env.OSSEnvironment.OSSTags = tags
				if env.HasPriorOSSValidation() {
					env.OSSValidation = env.GetPriorOSSValidation()
				}
				/* XXX Why change the record at all if skipping?
				if env.OSSEnvironment.OSSOnboardingPhase == ossrecord.EDIT {
					env.OSSEnvironmentExtended.ResetForRMC()
				}
				*/
			} else {
				if reason != "" {
					debug.Warning("Force merging of OSSEnvironment record: %s: %s", reason, env.String())
				}
				env.mergePhaseOne()
			}
			allMergedEntries = append(allMergedEntries, env)
		})
		if err != nil {
			return nil, debug.WrapError(err, "Error while merging all OSSEnvironment records (phase ONE) for pattern \"%s\"", pattern)
		}

		advanceGlobalMergePhase(mergePhaseEnvironmentsTwo)
		debug.Info("Starting merge phase TWO for Environments")
		err = ListAllEnvironments(nil, func(env *EnvironmentInfo) {
			env.mergePhaseTwo()
		})
		if err != nil {
			return nil, debug.WrapError(err, "Error while merging all OSSEnvironment records (phase TWO) for pattern \"%s\"", pattern)
		}
	} else {
		advanceGlobalMergePhase(mergePhaseEnvironmentsOne)
		advanceGlobalMergePhase(mergePhaseEnvironmentsTwo)
	}

	if ossrunactions.Services.IsEnabled() {

		// Merge phase: ONE
		advanceGlobalMergePhase(mergePhaseServicesOne)
		debug.Info("Starting merge for Services/Components: Phase 1: merging information from all basic sources")
		err = ListAllServices(pattern, func(si *ServiceInfo) {
			si.mergePhaseOne()
		})
		if err != nil {
			return nil, debug.WrapError(err, "Error during merge phase One for all service/component records for pattern \"%s\"", pattern)
		}

		// XXX Should we use a nil pattern here to catch all records during the later merge phases?

		// Merge phase: TWO
		advanceGlobalMergePhase(mergePhaseServicesTwo)
		//		defer profile.Start().Stop() // XXX

		// EXPERIMENTAL: attempt to match names with ClearingHouse
		if clearinghouse.HasCHInfo() {
			debug.Info("Starting merge for Services/Components: Phase 2a: ClearingHouse entry names")
			err := loadAllNames()
			if err != nil {
				return nil, debug.WrapError(err, "Loading all names for matching ClearingHouse entries")
			}
			err = mergeNameGroups()
			if err != nil {
				return nil, debug.WrapError(err, "Merging names of matching ClearingHouse entries")
			}
		} else {
			debug.Info("Skipping merge for Services/Components: Phase 2a: ClearingHouse entry names")
		}

		debug.Info("Starting merge for Services/Components: Phase 2b: cross-entry info: composites, parent resource, status category parent, ClearingHouse links")
		err = ListAllServices(pattern, func(si *ServiceInfo) {
			si.mergePhaseTwo()
		})
		if err != nil {
			return nil, debug.WrapError(err, "Error during merge phase Two for all service/component records for pattern \"%s\"", pattern)
		}
		runAllDeferredFunctions()

		// Check for issues related to multiple entries sharing the same ProductInfo
		debug.Info("Starting merge for Services/Components: Phase 2c: consistency of product (PID) and ClearingHouse information")
		checkProductInfo()

		// Merge phase: THREE
		advanceGlobalMergePhase(mergePhaseServicesThree)
		debug.Info("Starting merge for Services/Components: Phase 3: finalizing merge status and PnP enablement")
		err = ListAllServices(pattern, func(si *ServiceInfo) {
			si.mergePhaseThree()
			allMergedEntries = append(allMergedEntries, si)
			si.advanceMergePhase(mergePhaseFinalized)
		})
		if err != nil {
			return nil, debug.WrapError(err, "Error during merge phase Three(finalization) for all service/component records for pattern \"%s\"", pattern)
		}
	} else {
		advanceGlobalMergePhase(mergePhaseServicesOne)
		advanceGlobalMergePhase(mergePhaseServicesTwo)
		advanceGlobalMergePhase(mergePhaseServicesThree)
	}

	// Merge phase: FINAL
	advanceGlobalMergePhase(mergePhaseFinalized)
	debug.Info("All merge operations completed")

	return allMergedEntries, nil
}

// MergeOneService attempts to merge the info for a single service, as much as possible
// XXX This function cannot do a complete merge, because that might require establishing links and correlating information between multiple services
// It should be used only for testing and maintenance of the tool
func MergeOneService(si *ServiceInfo) error {
	advanceGlobalMergePhase(mergePhaseSegments)
	advanceGlobalMergePhase(mergePhaseEnvironmentsOne)
	advanceGlobalMergePhase(mergePhaseEnvironmentsTwo)

	advanceGlobalMergePhase(mergePhaseServicesOne)
	si.mergePhaseOne()

	advanceGlobalMergePhase(mergePhaseServicesTwo)
	si.mergePhaseTwo()
	runAllDeferredFunctions()

	advanceGlobalMergePhase(mergePhaseServicesThree)
	si.mergePhaseThree()
	si.SetFinalized()

	advanceGlobalMergePhase(mergePhaseFinalized)

	return nil
}
