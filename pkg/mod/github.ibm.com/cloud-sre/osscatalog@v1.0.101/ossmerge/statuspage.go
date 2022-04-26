package ossmerge

import (
	"fmt"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// XXX Special flag to force resetting the PnPCandidate tag in all entries where it is not needed
var resetPnPCandidate = false

// XXX Special flag to force adding the PnPCandidate tag in all entries that become automatically PnPEnabled
var addPnPCandidate = false

type pnpStatusType int

const (
	pnpUnknown pnpStatusType = iota
	pnpChecking
	pnpDisabled
	pnpEnabled
)

func (si *ServiceInfo) mergeStatusPage() {
	oss := &si.OSSService.StatusPage
	sn := &si.SourceServiceNow.StatusPage
	prior := &si.PriorOSS.StatusPage

	oss.Group = si.MergeValues("Status Page Notification Group", SeverityIfMissing{ossvalidation.MINOR}, SeverityIfMismatch{ossvalidation.WARNING}, OverrideProperty{"StatusPage.Group"}, RMCOSS{prior.Group}, ServiceNow{sn.Group}, PriorOSS{prior.Group}).(string)
	oss.CategoryID = si.MergeValues("Status Page Notification CategoryID", SeverityIfMissing{ossvalidation.MINOR}, SeverityIfMismatch{ossvalidation.WARNING}, OverrideProperty{"StatusPage.CategoryID"}, RMCOSS{prior.CategoryID}, ServiceNow{sn.CategoryID}, PriorOSS{prior.CategoryID}).(string)
	oss.CategoryIDMisspelled = oss.CategoryID
	oss.CategoryParent = si.MergeValues("Status Page Notification Category Parent", SeverityIfMissing{ossvalidation.IGNORE}, SeverityIfMismatch{ossvalidation.IGNORE}, OverrideProperty{"StatusPage.CategoryParent"}, RMCOSS{prior.CategoryParent}, PriorOSS{prior.CategoryParent}).(ossrecord.CRNServiceName)
	if oss.CategoryID != "" {
		si.registerStatusCategoryParent()
	} else if oss.CategoryParent != "" {
		si.AddValidationIssue(ossvalidation.SEVERE, "Status Page Notification Category ID is empty but Category Parent is not empty (deleting)", string(oss.CategoryParent)).TagStatusPage()
		oss.CategoryParent = ""
	}
}

type statusCategoryInfo struct {
	numEntries int
	parents    []ossrecord.CRNServiceName
}

var allStatusCategoryParents = make(map[string]*statusCategoryInfo)

const emptyCategoryParent ossrecord.CRNServiceName = "<empty>"

func (si *ServiceInfo) registerStatusCategoryParent() {
	if !si.OSSService.Compliance.ServiceNowOnboarded || si.OSSService.StatusPage.CategoryID == "" {
		return
	}
	categoryID := si.OSSService.StatusPage.CategoryID
	parent := si.OSSService.StatusPage.CategoryParent
	if parent == "" {
		parent = emptyCategoryParent
	}
	if sci, found := allStatusCategoryParents[categoryID]; found {
		sci.numEntries++
		var found0 bool
		for _, p0 := range sci.parents {
			if p0 == parent {
				found0 = true
				break
			}
		}
		if !found0 {
			sci.parents = append(sci.parents, parent)
		}
	} else {
		sci = &statusCategoryInfo{
			numEntries: 1,
			parents:    []ossrecord.CRNServiceName{parent},
		}
		allStatusCategoryParents[categoryID] = sci
	}
}

// checkStatusCategoryParent detects inconsistencies in the settings of the Status.Page.CategoryParent
// attributes across all ServiceInfo entries.
// This method must be called *after* all ServiceInfo records have been individually merged.
func (si *ServiceInfo) checkStatusCategoryParent() {
	si.checkMergePhase(mergePhaseServicesTwo)

	if !si.OSSService.Compliance.ServiceNowOnboarded {
		si.AddValidationIssue(ossvalidation.INFO, "Skipping computation of Status Page Category Parent for entry that is not ServiceNowOnboarded", "ServiceNow.Status=%s", si.SourceServiceNow.GeneralInfo.OperationalStatus).TagStatusPage()
		si.mergeWorkArea.pnpCategoryParentIssues++
		return
	}

	recordIssue := func() {
		if si.OSSService.GeneralInfo.OSSOnboardingPhase == "" {
			registerDeferredFunction(func() {
				si.OSSService.StatusPage.CategoryID = ""
				si.OSSService.StatusPage.CategoryParent = ""
			})
		} else {
			si.mergeWorkArea.pnpCategoryParentIssues++
		}
	}

	categoryID := si.OSSService.StatusPage.CategoryID
	if si.OSSService.StatusPage.CategoryID == "" {
		if si.OSSService.StatusPage.CategoryParent != "" {
			panic(fmt.Sprintf("checkStatusCategoryParent(): Status Page Notification Category ID is empty but Category Parent (%s) is not empty whie processing entry %s", si.OSSService.StatusPage.CategoryParent, si.String()))
		}
		return
	}
	if sci, found := allStatusCategoryParents[categoryID]; found {
		if len(sci.parents) == 0 {
			panic(fmt.Sprintf(`checkStatusCategoryParent() empty parents list for CategoryID="%s" in master list while processing entry %s`, categoryID, si.String()))
		}
		si.OSSValidation.StatusCategoryCount = sci.numEntries
		if len(sci.parents) > 1 {
			si.AddValidationIssue(ossvalidation.SEVERE, "More than one Status Page Category Parent for this CategoryID (voiding)", `CategoryID="%s"  parents=%v`, categoryID, sci.parents).TagStatusPage()
			recordIssue()
			return
		}
		parentName := si.OSSService.StatusPage.CategoryParent
		if parentName != "" {
			if parentName != sci.parents[0] {
				panic(fmt.Sprintf(`checkStatusCategoryParent() Parent Name "%s" does not match the name from the master list: "%s" for CategoryID="%s" in master list while processing entry %s`, parentName, sci.parents[0], categoryID, si.String()))
			} else if parentName == si.OSSService.ReferenceResourceName {
				si.AddNamedValidationIssue(ossvalidation.StatusCategoryParent, `CategoryID="%s"  number of entries with this CategoryID=%d`, categoryID, sci.numEntries)
			} else {
				parentComparableName := MakeComparableName(string(parentName))
				parentSI, found1 := LookupService(parentComparableName, false)
				if found1 {
					if parentName != parentSI.OSSService.ReferenceResourceName {
						si.AddValidationIssue(ossvalidation.SEVERE, "Status Page Notification Category Parent is found but does not use the canonical name (voiding)", `CategoryParent="%s"  CanonicalName="%s"`, parentName, parentSI.OSSService.ReferenceResourceName).TagStatusPage()
						recordIssue()
					}
					if categoryID != parentSI.OSSService.StatusPage.CategoryID {
						si.AddValidationIssue(ossvalidation.SEVERE, "Status Page Notification CategoryID does not match the CategoryID of the CategoryParent (voiding)", `this.CategoryID="%s"  parent.CategoryID="%s"`, categoryID, parentSI.OSSService.StatusPage.CategoryID).TagStatusPage()
						recordIssue()
					}
				} else {
					si.AddValidationIssue(ossvalidation.SEVERE, "Status Page Notification Category Parent not found (voiding)", string(parentName)).TagStatusPage()
					recordIssue()
				}
			}
		} else if sci.numEntries > 1 {
			si.AddValidationIssue(ossvalidation.SEVERE, "Status Page Notification CategoryID is used in more than one entry but Category Parent is blank (voiding)", `CategoryID="%s"  number of entries with same ID=%d`, categoryID, sci.numEntries).TagStatusPage()
			recordIssue()
		}
	} else {
		si.AddValidationIssue(ossvalidation.CRITICAL, `CategoryID not found in master list while processing entry`, `CategoryID="%s"  CategoryParent="%s"`, si.OSSService.StatusPage.CategoryID, si.OSSService.StatusPage.CategoryParent).TagStatusPage()
		debug.PrintError(`checkStatusCategoryParent() CategoryID="%s" not found in master list while processing entry %s`, categoryID, si.String())
		si.mergeWorkArea.pnpCategoryParentIssues++
	}
}

// checkPnPEnablement checks if all conditions are met to enable the use of this service/component with the PnP system.
// If yes, this function returns true and adds the "PnPEnabled" tag in the OSSTags attribute of the OSS entry
func (si *ServiceInfo) checkPnPEnablement() (result bool) {
	defer func() {
		if si.OSSService.GeneralInfo.OSSOnboardingPhase == ossrecord.EDIT {
			if result != si.OSSService.GeneralInfo.OSSTags.Contains(osstags.PnPEnabled) {
				si.AddValidationIssue(ossvalidation.SEVERE, `Deferring a change to the "pnp_enabled" setting because this entry is currently being edited in RMC`, `proposed_new_value=%v`, result).TagStatusPage().TagCRN()
			}
			result = si.OSSService.GeneralInfo.OSSTags.Contains(osstags.PnPEnabled)
		}
	}()

	switch si.mergeWorkArea.pnpStatus {
	case pnpUnknown:
		// This is our first time in this function for this ServiceInfo
		// keep going and compute the PnP enablement
		si.mergeWorkArea.pnpStatus = pnpChecking
	case pnpChecking:
		debug.PrintError("Recursive call to CheckPnpEnabled() for %s -- possible loop in StatusPage.CategoryParent=%s  /  StatusPage.CategoryID=%s ?", si.String(), si.OSSService.StatusPage.CategoryParent, si.OSSService.StatusPage.CategoryID)
		si.AddValidationIssue(ossvalidation.CRITICAL, "Recursive call to CheckPnpEnabled()  -- possible loop in StatusPage.CategoryParent?", "StatusPage.CategoryParent=%s  StatusPage.CategoryID=%s", si.OSSService.StatusPage.CategoryParent, si.OSSService.StatusPage.CategoryID).TagStatusPage()
		si.mergeWorkArea.pnpStatus = pnpDisabled
		return false
	case pnpDisabled:
		return false
	case pnpEnabled:
		return true
	default:
		panic(fmt.Sprintf("Unknown pnpStatus=%v in entry %s", si.mergeWorkArea.pnpStatus, si.String()))
	}

	si.checkGlobalMergePhaseMultiple(mergePhaseServicesThree)
	si.checkEntryMergePhaseMultiple(mergePhaseServicesTwo, mergePhaseServicesThree) // Note: some parent entries may be checked from a call from checkPnPEnablement in a child entry and may not themselves be in phase THREE yet

	// Check the PnP enablement control tags present
	var numPnPTags = 0
	var tagPnPCandidate = si.OSSService.GeneralInfo.OSSTags.Contains(osstags.PnPCandidate)
	if tagPnPCandidate {
		numPnPTags++
	}
	var tagPnPInclude = si.OSSService.GeneralInfo.OSSTags.Contains(osstags.PnPInclude)
	if tagPnPInclude {
		numPnPTags++
	}
	var tagPnPExclude = si.OSSService.GeneralInfo.OSSTags.Contains(osstags.PnPExclude)
	if tagPnPExclude {
		numPnPTags++
	}
	if numPnPTags > 1 {
		si.AddValidationIssue(ossvalidation.CRITICAL, "More than one PnP enablement control tag specified -- ignoring and not enabling for PnP", "pnp_candiate=%v  pnp_include=%v  pnp_exclude=%v", tagPnPCandidate, tagPnPInclude, tagPnPExclude).TagStatusPage()
		si.mergeWorkArea.pnpStatus = pnpDisabled
	}

	var issuesBasicCriteria []*ossvalidation.ValidationIssue
	var issuesNonOverridable []*ossvalidation.ValidationIssue
	var issuesOverridable []*ossvalidation.ValidationIssue
	var numIssues = -1 // We start with -1, meaning we have not yet started to count the true number of issues

	// Check basic criteria
	if !(si.OSSService.GeneralInfo.EntryType == ossrecord.SERVICE ||
		si.OSSService.GeneralInfo.EntryType == ossrecord.RUNTIME ||
		si.OSSService.GeneralInfo.EntryType == ossrecord.TEMPLATE ||
		si.OSSService.GeneralInfo.EntryType == ossrecord.IAAS ||
		si.OSSService.GeneralInfo.EntryType == ossrecord.PLATFORMCOMPONENT ||
		si.OSSService.GeneralInfo.EntryType == ossrecord.SUPERCOMPONENT ||
		si.OSSService.GeneralInfo.EntryType == ossrecord.SUBCOMPONENT ||
		si.OSSService.GeneralInfo.EntryType == ossrecord.CONSULTING ||
		si.OSSService.GeneralInfo.EntryType == ossrecord.CONTENT) {
		/* We do not expect pnp_enabled for INTERNALSERVICE */
		issuesBasicCriteria = append(issuesBasicCriteria, ossvalidation.NewIssue(ossvalidation.WARNING, `Entry does not meet normal criteria for being "pnp_enabled" (but this can be overridden): type is not SERVICE, RUNTIME, TEMPLATE, IAAS, PLATFORMCOMPONENT or SUBCOMPONENT`, "EntryType=%v", si.OSSService.GeneralInfo.EntryType).TagStatusPage())
	}
	if !(si.OSSService.GeneralInfo.OperationalStatus == ossrecord.GA ||
		si.OSSService.GeneralInfo.OperationalStatus == ossrecord.BETA ||
		si.OSSService.GeneralInfo.OperationalStatus == ossrecord.DEPRECATED ||
		si.OSSService.GeneralInfo.OperationalStatus == ossrecord.COMMUNITY) {
		issuesBasicCriteria = append(issuesBasicCriteria, ossvalidation.NewIssue(ossvalidation.WARNING, `Entry does not meet normal criteria for being "pnp_enabled" (but this can be overridden): status is not GA, BETA, DEPRECATED or COMMUNITY`, "OperationalStatus=%v", si.OSSService.GeneralInfo.OperationalStatus).TagStatusPage())
	}

	if issuesBasicCriteria != nil && numPnPTags == 0 {
		// No chance we would pnp_enable - bail out
		si.AddNamedValidationIssue(ossvalidation.PnPNotEnabledMissBasicCriteria, "")
		si.mergeWorkArea.pnpStatus = pnpDisabled
	}

	if si.mergeWorkArea.pnpStatus == pnpChecking { // Check that we did not "bail out" early and it is still useful to check for issues

		// Check non-overridable issues
		if (!si.HasSourceServiceNow() || si.GetSourceServiceNow().IsRetired()) && !si.GeneralInfo.OSSTags.Contains(osstags.OSSOnly) {
			issuesNonOverridable = append(issuesNonOverridable, ossvalidation.NewIssue(ossvalidation.SEVERE, `Entry cannot be "pnp_enabled" (and this cannot be overridden): entry not in ServiceNow or RETIRED`, "").TagStatusPage())
		} else if si.OSSService.StatusPage.CategoryID == "" {
			issuesNonOverridable = append(issuesNonOverridable, ossvalidation.NewIssue(ossvalidation.SEVERE, `Entry cannot be "pnp_enabled" (and this cannot be overridden): StatusPage.CategoryID is empty`, "").TagStatusPage())
		} else if si.mergeWorkArea.pnpCategoryParentIssues > 0 {
			issuesNonOverridable = append(issuesNonOverridable, ossvalidation.NewIssue(ossvalidation.SEVERE, `Entry cannot be "pnp_enabled" (and this cannot be overridden): CategoryParent is inconsistent`, "").TagStatusPage())
		} else if si.OSSService.StatusPage.CategoryParent != "" && si.OSSService.StatusPage.CategoryParent != si.OSSService.ReferenceResourceName {
			parent, found := LookupService(MakeComparableName(string(si.OSSService.StatusPage.CategoryParent)), false)
			if found {
				ok := parent.checkPnPEnablement()
				if !ok {
					issuesNonOverridable = append(issuesNonOverridable, ossvalidation.NewIssue(ossvalidation.SEVERE, `Entry cannot be "pnp_enabled" (and this cannot be overridden): StatusPage.CategoryParent is not itself "pnp_enabled"`, `StatusPage.CategoryParent="%v"`, si.OSSService.StatusPage.CategoryParent).TagStatusPage())
				}
			} else {
				issuesNonOverridable = append(issuesNonOverridable, ossvalidation.NewIssue(ossvalidation.SEVERE, `Entry cannot be "pnp_enabled" (and this cannot be overridden): StatusPage.CategoryParent is not found`, `StatusPage.CategoryParent="%v"`, si.OSSService.StatusPage.CategoryParent).TagStatusPage())
			}
		}

		// Check overridable issues
		if si.OSSService.StatusPage.CategoryParent == "" || si.OSSService.StatusPage.CategoryParent == si.OSSService.ReferenceResourceName {
			// Standalone entry
			if !si.OSSService.GeneralInfo.OSSTags.Contains(osstags.PnPEnabledIaaS) && !si.OSSService.GeneralInfo.ClientFacing {
				issuesOverridable = append(issuesOverridable, ossvalidation.NewIssue(ossvalidation.WARNING, `Entry does not meet normal criteria for being "pnp_enabled" (but this can be overridden): not marked "client-facing" nor "pnp_enabled_iaas"`, "client-facing=%v  pnp_enabled_iaas=%v", si.OSSService.GeneralInfo.ClientFacing, si.OSSService.GeneralInfo.OSSTags.Contains(osstags.PnPEnabledIaaS)).TagStatusPage())
			}
			if !si.OSSService.Compliance.ServiceNowOnboarded && si.HasSourceServiceNow() {
				if si.HasSourceScorecardV1Detail() {
					issuesOverridable = append(issuesOverridable, ossvalidation.NewIssue(ossvalidation.WARNING, `Entry does not meet normal criteria for being "pnp_enabled" (but this can be overridden): not marked "ServiceNowOnboarded" in ScorecardV1`, "").TagStatusPage())
				} else if ossrunactions.ScorecardV1.IsEnabled() {
					if si.OSSService.GeneralInfo.EntryType != ossrecord.SUBCOMPONENT &&
						si.OSSService.GeneralInfo.EntryType != ossrecord.SUPERCOMPONENT &&
						si.OSSService.GeneralInfo.EntryType != ossrecord.CONSULTING &&
						si.OSSService.GeneralInfo.EntryType != ossrecord.CONTENT &&
						si.OSSService.GeneralInfo.EntryType != ossrecord.INTERNALSERVICE {
						issuesOverridable = append(issuesOverridable, ossvalidation.NewIssue(ossvalidation.WARNING, `Entry does not meet normal criteria for being "pnp_enabled" (but this can be overridden): entry not in ScorecardV1 (and thus not marked "ServiceNowOnboarded" in ScorecardV1)`, "").TagStatusPage())
					}
				}
			}
		} else {
			// Entry under a CategoryParent
		}
		crnStatus := si.OSSService.GeneralInfo.OSSTags.GetCRNStatus()
		switch crnStatus {
		case osstags.StatusCRNGreen:
			// All good
			break
		case osstags.StatusCRNYellow:
			if si.HasPriorOSS() && si.GetPriorOSS().GetOSSTags().Contains(osstags.PnPEnabled) {
				si.AddValidationIssue(ossvalidation.WARNING, `Entry does not meet normal criteria for being "pnp_enabled" (CRN Validation Status is not "green") but it was already "pnp_enabled" previously - ignoring this issue`, "CRNStatus=%s", si.OSSService.GeneralInfo.OSSTags.GetCRNStatus()).TagStatusPage()
				break
			}
			fallthrough
		case osstags.StatusCRNRed:
			issuesOverridable = append(issuesOverridable, ossvalidation.NewIssue(ossvalidation.WARNING, `Entry does not meet normal criteria for being "pnp_enabled" (but this can be overridden): CRN Validation Status is not "green"`, "CRNStatus=%s", si.OSSService.GeneralInfo.OSSTags.GetCRNStatus()).TagStatusPage())
		default:
			panic(fmt.Sprintf("checkPnPEnablement() - CRN Validation Status invalid for %s - status=%q", si.String(), crnStatus))
		}

		// Take stock of all the criteria and decide what to do
		numIssues = len(issuesBasicCriteria) + len(issuesNonOverridable) + len(issuesOverridable)
		switch {
		case tagPnPCandidate:
			if numIssues == 0 {
				si.AddNamedValidationIssue(ossvalidation.PnPEnabledWithUnnecessaryCandidate, "%d issues", numIssues)
				si.mergeWorkArea.pnpStatus = pnpEnabled
				if resetPnPCandidate {
					si.AddValidationIssue(ossvalidation.INFO, fmt.Sprintf(`resetPnPCandidate: removing the "%s" tag`, osstags.PnPCandidate), "").TagStatusPage()
					si.OSSMergeControl.OSSTags.RemoveTag(osstags.PnPCandidate)
				}
			} else {
				si.AddNamedValidationIssue(ossvalidation.PnPNotEnabledButCandidate, "%d issues", numIssues)
				si.mergeWorkArea.pnpStatus = pnpDisabled
			}
		case tagPnPInclude:
			if issuesNonOverridable == nil {
				if issuesBasicCriteria == nil && issuesOverridable == nil {
					si.AddNamedValidationIssue(ossvalidation.PnPEnabledWithUnnecessaryInclude, "%d issues", numIssues)
					si.mergeWorkArea.pnpStatus = pnpEnabled
				} else {
					si.AddNamedValidationIssue(ossvalidation.PnPEnabledWithInclude, "%d issues", numIssues)
					si.mergeWorkArea.pnpStatus = pnpEnabled
				}
			} else {
				si.AddNamedValidationIssue(ossvalidation.PnPNotEnabledButInclude, "%d issues", numIssues)
				si.mergeWorkArea.pnpStatus = pnpDisabled
			}
		case tagPnPExclude:
			if numIssues == 0 {
				si.AddNamedValidationIssue(ossvalidation.PnPNotEnabledWithExclude, "%d issues", numIssues)
				si.mergeWorkArea.pnpStatus = pnpDisabled
			} else {
				si.AddNamedValidationIssue(ossvalidation.PnPNotEnabledWithUnnecessaryExclude, "%d issues", numIssues)
				si.mergeWorkArea.pnpStatus = pnpDisabled
			}
		default:
			if numIssues == 0 {
				si.AddNamedValidationIssue(ossvalidation.PnPEnabledWithoutCandidate, "%d issues", numIssues)
				si.mergeWorkArea.pnpStatus = pnpEnabled
				if addPnPCandidate {
					si.AddValidationIssue(ossvalidation.INFO, fmt.Sprintf(`addPnPCandidate: adding the "%s" tag`, osstags.PnPCandidate), "").TagStatusPage()
					si.OSSMergeControl.OSSTags.AddTag(osstags.PnPCandidate)
				}
			} else {
				si.AddNamedValidationIssue(ossvalidation.PnPNotEnabledWithoutCandidate, "%d issues", numIssues)
				si.mergeWorkArea.pnpStatus = pnpDisabled
			}
		}

		// Append all the ValidationIssues
		for _, v := range issuesBasicCriteria {
			si.OSSValidation.AddIssuePreallocated(v)
		}
		for _, v := range issuesNonOverridable {
			si.OSSValidation.AddIssuePreallocated(v)
		}
		for _, v := range issuesOverridable {
			si.OSSValidation.AddIssuePreallocated(v)
		}
	}

	// Generate the final return code
	switch si.mergeWorkArea.pnpStatus {
	case pnpDisabled:
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.PnPEnabledIaaS) {
			si.OSSValidation.AddNamedIssue(ossvalidation.PnPNotEnabledButEnabledIaaS, "%d issues", numIssues)
		}
		if si.HasPriorOSS() && si.GetPriorOSS().GeneralInfo.OSSTags.Contains(osstags.PnPEnabled) {
			si.AddNamedValidationIssue(ossvalidation.PnPEnabledRemoved, "%s", si.String())
		}
		return false
	case pnpEnabled:
		si.OSSService.GeneralInfo.OSSTags.AddTag(osstags.PnPEnabled)
		if !si.HasPriorOSS() || !si.GetPriorOSS().GeneralInfo.OSSTags.Contains(osstags.PnPEnabled) {
			si.AddNamedValidationIssue(ossvalidation.PnPEnabledAdded, "%s  -- categoryParent=%s   categoryId=%s", si.String(), si.OSSService.StatusPage.CategoryParent, si.OSSService.StatusPage.CategoryID)
		}
		return true
	default:
		panic(fmt.Sprintf("Unexpected pnpStatus=%v at end of checkPnPEnablement() for entry %s", si.mergeWorkArea.pnpStatus, si.String()))
	}
}
