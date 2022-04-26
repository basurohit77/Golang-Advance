package ossmerge

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/servicenow"
)

// Functions related to the merging of ServiceNow info

// normalizeServiceNowEntry normalizes the enumeration-based attributes in one ServiceNow entry contained in one ServiceInfo
func (si *ServiceInfo) normalizeServiceNowEntry(sn *servicenow.ConfigurationItem) {
	var err error
	entryName := sn.CRNServiceName

	sn.GeneralInfo.EntryType, err = servicenow.ParseEntryType(string(sn.GeneralInfo.EntryType))
	if err != nil {
		si.AddValidationIssue(ossvalidation.WARNING, "Error parsing EntryType from ServiceNow record", `%v   (entry="%s")`, err, entryName).TagConsistency()
	}
	sn.GeneralInfo.OperationalStatus, err = servicenow.ParseOperationalStatus(string(sn.GeneralInfo.OperationalStatus))
	if err != nil {
		si.AddValidationIssue(ossvalidation.WARNING, "Error parsing OperationalStatus from ServiceNow record", `%v   (entry="%s")`, err, entryName).TagConsistency()
	}
	sn.Support.ClientExperience, err = servicenow.ParseClientExperience(string(sn.Support.ClientExperience))
	if err != nil {
		si.AddValidationIssue(ossvalidation.WARNING, "Error parsing ClientExperience from ServiceNow record", `%v   (entry="%s")`, err, entryName).TagConsistency()
	}
	sn.Support.Tier2EscalationType, err = servicenow.ParseTier2EscalationType(string(sn.Support.Tier2EscalationType))
	if err != nil {
		si.AddValidationIssue(ossvalidation.WARNING, "Error parsing Support.Tier2EscalationType from ServiceNow record", `%v   (entry="%s")`, err, entryName).TagConsistency()
	}
	sn.Operations.Tier2EscalationType, err = servicenow.ParseTier2EscalationType(string(sn.Operations.Tier2EscalationType))
	if err != nil {
		si.AddValidationIssue(ossvalidation.WARNING, "Error parsing Operations.Tier2EscalationType from ServiceNow record", `%v   (entry="%s")`, err, entryName).TagConsistency()
	}
	debug.Debug(debug.Fine, "normalizeServiceNowEntry(%s, %s) exit", si.String(), sn.String())
}

// processServiceNowImport augments all the ServiceNow entries from API with entries from the csv import file,
// and find any missing ServiceNow entries
// This must happen after we have loaded from all other sources, so that we can accurately report
// if an entry from the ServiceNow import file is not found anywhere else
func processServiceNowImport(pattern *regexp.Regexp) error {
	err := servicenow.ListServiceNowImport(pattern, func(sni *servicenow.SNImport) {
		comparableName := MakeComparableName(sni.Name)
		si, found := LookupService(comparableName, true)
		if found {
			var hasSN bool
			if si.HasSourceServiceNow() {
				if si.GetSourceServiceNow().CRNServiceName == sni.Name {
					hasSN = true
				}
			}
			for _, sn := range si.AdditionalServiceNow {
				if sn.CRNServiceName == sni.Name {
					if hasSN {
						panic(fmt.Sprintf(`processServiceNowImport(%s) - found more than ServiceNow entry with the same name="%s"`, si.String(), sni.Name))
					}
					hasSN = true
				}
			}
			if !hasSN {
				si.AddValidationIssue(ossvalidation.MINOR, "Entry not found in ServiceNow API but found in the ServiceNow csv import file, also found in some other sources (Catalog, ScorecardV1, etc.)", `name=%s`, sni.Name).TagSNOverride()
				sn := &servicenow.ConfigurationItem{}
				si.mergeServiceNowImport(sn, sni)
				// Add the new entry in the AdditionalServiceNow; let the additional sources processing move it to the main entry if appropriate
				si.AdditionalServiceNow = append(si.AdditionalServiceNow, sn)
			} else {
				// We do not override the API results with the ServiceNow import, since they may be more current
			}
		} else {
			si.AddValidationIssue(ossvalidation.MINOR, "Entry not found in ServiceNow API but found in the ServiceNow csv import file, found nowhere else", `name=%s`, sni.Name).TagSNOverride()
			// Make sure we do not pick-up any bad data from SN API
			si.SourceServiceNow = servicenow.ConfigurationItem{}
			si.mergeServiceNowImport(si.GetSourceServiceNow(), sni)
		}
	})
	return err
}

// mergeServiceNowImport merges one ServiceNow entry contained in a ServiceInfo, based on information obtained from the ServiceNow csv import file
func (si *ServiceInfo) mergeServiceNowImport(sn *servicenow.ConfigurationItem, sni *servicenow.SNImport) {
	sn.CRNServiceName = sni.Name
	sn.DisplayName = sni.DisplayName
	sn.GeneralInfo.FullCRN = sni.FullCRN
	sn.StatusPage.CategoryID = sni.StatusPageNotificationCategoryID
	sn.ServiceNowInfo.SupportTier1AG = sni.Tier1SupportAssignmentGroup
	sn.ServiceNowInfo.OperationsTier1AG = sni.Tier1OperationsAssignmentGroup
	sn.Support.Tier2EscalationType = ossrecord.Tier2EscalationType(sni.Tier2SupportEscalationType)
	sn.ServiceNowInfo.SupportTier2AG = sni.Tier2SupportAssignmentGroup
	sn.Support.Tier2Repo = ossrecord.GHRepo(sni.Tier2SupportEscalationGitHubRepo)
	sn.Operations.Tier2EscalationType = ossrecord.Tier2EscalationType(sni.Tier2OperationsEscalationType)
	sn.ServiceNowInfo.OperationsTier2AG = sni.Tier2OperationsAssignmentGroup
	sn.Operations.Tier2Repo = ossrecord.GHRepo(sni.Tier2OperationsEscalationGitHubRepo)
	sn.GeneralInfo.EntryType = ossrecord.EntryType(sni.EntryType)
	sn.Support.ClientExperience = ossrecord.ClientExperience(sni.ClientExperience)
	sn.GeneralInfo.ClientFacing = sni.CustomerFacing
	// TODO: Figure out how to interpret the TOCEnbled flag in the original ServiceNow csv import
	//si.overrideServiceNowAttribute(entryName, "TOCEnabled", , sni.TOCEnabled, nil)
	sn.GeneralInfo.OperationalStatus = ossrecord.OperationalStatus(sni.OperationalStatus)
	sn.Ownership.OfferingManager.Name = sni.OfferingManager
	sn.StatusPage.Group = sni.StatusPageNotificationGroup
	//si.overrideServiceNowAttribute(entryName, "CreatedBy", , sni.CreatedBy, nil)
	//si.overrideServiceNowAttribute(entryName, "UpdatedBy", , sni.UpdatedBy, nil)
	sn.Ownership.SegmentName = sni.Segment
	sn.Ownership.TribeName = sni.Tribe
	sn.Operations.Manager.Name = sni.OperationsManager
	sn.Support.Manager.Name = sni.SupportManager
	si.normalizeServiceNowEntry(sn)
}

func (si *ServiceInfo) mergeServiceNowInfo() {
	oss := &si.OSSService.ServiceNowInfo
	sn := &si.SourceServiceNow.ServiceNowInfo
	prior := &si.PriorOSS.ServiceNowInfo
	var err error

	oss.SupportTier1AG = si.MergeValues("SupportTier1AG", ServiceNow{sn.SupportTier1AG}, PriorOSS{prior.SupportTier1AG}).(string)
	if si.OSSService.Support.Tier2EscalationType == ossrecord.SERVICENOW {
		oss.SupportTier2AG = si.MergeValues("SupportTier2AG", ServiceNow{sn.SupportTier2AG}, PriorOSS{prior.SupportTier2AG}).(string)
	} else {
		oss.SupportTier2AG = si.MergeValues("SupportTier2AG", SeverityIfMissing{ossvalidation.IGNORE}, ServiceNow{sn.SupportTier2AG}, PriorOSS{prior.SupportTier2AG}).(string)
	}
	oss.OperationsTier1AG = si.MergeValues("OperationsTier1AG", ServiceNow{sn.OperationsTier1AG}, PriorOSS{prior.OperationsTier1AG}).(string)
	if si.OSSService.Operations.Tier2EscalationType == ossrecord.SERVICENOW {
		oss.OperationsTier2AG = si.MergeValues("OperationsTier2AG", ServiceNow{sn.OperationsTier2AG}, PriorOSS{prior.OperationsTier2AG}).(string)
	} else {
		oss.OperationsTier2AG = si.MergeValues("OperationsTier2AG", SeverityIfMissing{ossvalidation.IGNORE}, ServiceNow{sn.OperationsTier2AG}, PriorOSS{prior.OperationsTier2AG}).(string)
	}
	oss.RCAApprovalGroup = si.MergeValues("RCAApprovalGroup", ServiceNow{sn.RCAApprovalGroup}, PriorOSS{prior.RCAApprovalGroup}).(string)
	oss.ERCAApprovalGroup = si.MergeValues("ERCAApprovalGroup", ServiceNow{sn.ERCAApprovalGroup}, PriorOSS{prior.ERCAApprovalGroup}).(string)
	oss.TargetedCommunication = si.MergeValues("TargetedCommunication", ServiceNow{sn.TargetedCommunication}, PriorOSS{prior.TargetedCommunication}).(string)
	oss.CIEPageout = si.MergeValues("CIEPageout", ServiceNow{sn.CIEPageout}, PriorOSS{prior.CIEPageout}).(string)
	if sn.SupportNotApplicable != "" {
		oss.SupportNotApplicable, err = strconv.ParseBool(sn.SupportNotApplicable)
		if err != nil {
			si.AddValidationIssue(ossvalidation.WARNING, "Error parsing SupportNotApplicable from ServiceNow record", `%v   (entry="%s")`, err, si.SourceServiceNow.CRNServiceName).TagConsistency()
			oss.SupportNotApplicable = false
		}
	} else {
		oss.SupportNotApplicable = false
	}
	if sn.OperationsNotApplicable != "" {
		oss.OperationsNotApplicable, err = strconv.ParseBool(sn.OperationsNotApplicable)
		if err != nil {
			si.AddValidationIssue(ossvalidation.WARNING, "Error parsing OperationsNotApplicable from ServiceNow record", `%v   (entry="%s")`, err, si.SourceServiceNow.CRNServiceName).TagConsistency()
			oss.OperationsNotApplicable = false
		}
	} else {
		oss.OperationsNotApplicable = false
	}
}

// checkServiceNowEnrollment checks for any internal consistency issues with the ServiceNow enrollment data
func (si *ServiceInfo) checkServiceNowEnrollment() {
	if !si.HasSourceServiceNow() {
		return
	}
	sn := si.GetSourceServiceNow()
	if sn.GeneralInfo.OperationalStatus == ossrecord.RETIRED {
		si.AddValidationIssue(ossvalidation.INFO, "Omitting internal ServiceNow enrollment validation for Retired CI", "ServiceNow Status=%v", sn.GeneralInfo.OperationalStatus).TagSNEnrollment()
		return
	}

	sni := &sn.ServiceNowInfo

	if sni.SupportTier1AG == "" && sni.OperationsTier1AG == "" {
		si.AddValidationIssue(ossvalidation.SEVERE, "ServiceNow Tier1 Assignment Groups for Support and Operations are both empty -- this CI does not appear to really be used", "").TagSNEnrollment()
	}

	if !si.OSSService.ServiceNowInfo.SupportNotApplicable {
		if sni.SupportTier1AG == "" {
			si.AddValidationIssue(ossvalidation.SEVERE, "Tier1 Support Assignment Group is missing", "").TagSNEnrollment()
		}
		if isACSAssignmentGroup(sni.SupportTier2AG) {
			si.AddValidationIssue(ossvalidation.SEVERE, "Service is using ACS (IBM Cloud Support) as a Tier2 for Support", "SupportTier2AG=%s", sni.SupportTier2AG).TagSNEnrollment()
		}
		if isTOCAssignmentGroup(sni.SupportTier1AG) {
			si.AddValidationIssue(ossvalidation.SEVERE, "Service is using TOC as a front-end (Tier1) for Support", "SupportTier1AG=%s", sni.SupportTier1AG).TagSNEnrollment()
		}
		if isTOCAssignmentGroup(sni.SupportTier2AG) {
			si.AddValidationIssue(ossvalidation.SEVERE, "Service is using TOC as a Tier2 for Support", "SupportTier2AG=%s", sni.SupportTier2AG).TagSNEnrollment()
		}
		sns := &sn.Support
		switch sns.ClientExperience {
		case "":
			si.AddValidationIssue(ossvalidation.MINOR, "ServiceNow Client Experience attribute is empty", "").TagSNEnrollment()
		case ossrecord.ACSSUPPORTED:
			if !isACSAssignmentGroup(sni.SupportTier1AG) {
				si.AddValidationIssue(ossvalidation.MINOR, `ServiceNow Client Experience is "ACS Supported" but ACS is not the Tier1 Support Assignment Group`, "SupportTier1AG=%s", sni.SupportTier1AG).TagSNEnrollment()
			}
		case ossrecord.TRIBESUPPORTED:
			if isACSAssignmentGroup(sni.SupportTier1AG) {
				si.AddValidationIssue(ossvalidation.MINOR, `ServiceNow Client Experience is "Tribe Supported" but has ACS as the Tier1 Support Assignment Group`, "SupportTier1AG=%s", sni.SupportTier1AG).TagSNEnrollment()
			}
		default:
			si.AddValidationIssue(ossvalidation.MINOR, "Unknown ServiceNow Client Experience attribute", "Client Experience=%s", sns.ClientExperience).TagSNEnrollment()
		}
		switch sns.Tier2EscalationType {
		case ossrecord.SERVICENOW:
			/* deprecated
			if sni.SupportTier2AG == "" {
				si.AddValidationIssue(ossvalidation.SEVERE, `Tier2 Support Assignment Group cannot be empty if Escalation Type is "ServiceNow"`, "").TagSNEnrollment()
			}
			*/
		case ossrecord.GITHUB:
			if err := isValidGitHubRepo(string(sns.Tier2Repo)); err != nil {
				si.AddValidationIssue(ossvalidation.SEVERE, `ServiceNow Support GitHub repo is invalid but Support Escalation Type is "GITHUB"`, "%v", err).TagSNEnrollment()
			}
		case ossrecord.RTC:
			si.AddValidationIssue(ossvalidation.SEVERE, `ServiceNow Support Tier2 Escalation Type "RTC" is deprecated`, "Support Tier2 Escalation Type=%s", sns.Tier2EscalationType).TagSNEnrollment()
		case ossrecord.OTHERESCALATION:
			si.AddValidationIssue(ossvalidation.INFO, `ServiceNow Support Tier2 Escalation Type is "OTHER" - this is rare but probably OK`, "Support Tier2 Escalation Type=%s", sns.Tier2EscalationType).TagSNEnrollment()
		case "":
			if isACSAssignmentGroup(sni.SupportTier1AG) {
				si.AddValidationIssue(ossvalidation.SEVERE, `Tier2 Support escalation type must not be empty if ACS is the Tier1 Support Assignment Group`, "SupportTier1AG=%s", sni.SupportTier1AG).TagSNEnrollment()
			} else {
				si.AddValidationIssue(ossvalidation.MINOR, `Tier2 Support escalation type is empty but the Tier1 Support Assignment Group is not ACS - this may be OK depending on the Tribe`, "SupportTier1AG=%s", sni.SupportTier1AG).TagSNEnrollment()
			}
		default:
			si.AddValidationIssue(ossvalidation.CRITICAL, "Unknown ServiceNow Support Tier2 Escalation Type attribute", "Support Tier2 Escalation Type=%s", sns.Tier2EscalationType).TagSNEnrollment()
		}
	}

	if !si.OSSService.ServiceNowInfo.OperationsNotApplicable {
		if sni.OperationsTier1AG == "" {
			si.AddValidationIssue(ossvalidation.SEVERE, "Tier1 Operations Assignment Group is missing", "").TagSNEnrollment()
		}
		if isACSAssignmentGroup(sni.OperationsTier1AG) {
			si.AddValidationIssue(ossvalidation.SEVERE, "Service is using ACS (IBM Cloud Support) as a front-end (Tier1) for Operations", "OperationsTier1AG=%s", sni.OperationsTier1AG).TagSNEnrollment()
		}
		if isACSAssignmentGroup(sni.OperationsTier2AG) {
			si.AddValidationIssue(ossvalidation.MINOR, "Service is using ACS (IBM Cloud Support) as a Tier2 for Operations", "OperationsTier2AG=%s", sni.OperationsTier2AG).TagSNEnrollment()
		}
		if isTOCAssignmentGroup(sni.OperationsTier1AG) {
			si.AddValidationIssue(ossvalidation.SEVERE, "Service is using TOC as a front-end (Tier1) for Operations", "OperationsTier1AG=%s", sni.OperationsTier1AG).TagSNEnrollment()
		}
		if isTOCAssignmentGroup(sni.OperationsTier2AG) {
			si.AddValidationIssue(ossvalidation.MINOR, "Service is using TOC as a Tier2 for Operations", "OperationsTier2AG=%s", sni.OperationsTier2AG).TagSNEnrollment()
		}
		/* deprecated
		sno := &sn.Operations
		switch sno.Tier2EscalationType {
		case ossrecord.SERVICENOW:
			if sni.OperationsTier2AG == "" {
				si.AddValidationIssue(ossvalidation.MINOR, `Tier2 Operations Assignment Group cannot be empty if Escalation Type is "ServiceNow"`, "").TagSNEnrollment()
			}
		case ossrecord.GITHUB:
			if err := isValidGitHubRepo(string(sno.Tier2Repo)); err != nil {
				si.AddValidationIssue(ossvalidation.MINOR, `ServiceNow Operations GitHub repo is invalid but Operations Escalation Type is "GITHUB"`, "%v", err).TagSNEnrollment()
			}
		case ossrecord.RTC:
			si.AddValidationIssue(ossvalidation.MINOR, `ServiceNow Operations Tier2 Escalation Type "RTC" is deprecated`, "Operations Tier2 Escalation Type=%s", sno.Tier2EscalationType).TagSNEnrollment()
		case ossrecord.OTHERESCALATION:
			si.AddValidationIssue(ossvalidation.INFO, `ServiceNow Operations Tier2 Escalation Type is "OTHER" - this is rare but probably OK`, "Operations Tier2 Escalation Type=%s", sno.Tier2EscalationType).TagSNEnrollment()
		case "":
			si.AddValidationIssue(ossvalidation.MINOR, `Tier2 Operations escalation type is empty - this may be OK depending on the Tribe`, "SupportTier1AG=%s", sni.SupportTier1AG).TagSNEnrollment()
		default:
			si.AddValidationIssue(ossvalidation.MINOR, "Unknown ServiceNow Operations Tier2 Escalation Type attribute", "Operations Tier2 Escalation Type=%s", sno.Tier2EscalationType).TagSNEnrollment()
		}
		*/
	}
}

// isACSAssignmentGroup returns true is the specified Assignment Group belongs to ACS (IBM Cloud Support)
// FIXME: check AG prefixes for ACS
func isACSAssignmentGroup(ag string) bool {
	return strings.HasPrefix(ag, "acs-")
}

// isTOCssignmentGroup returns true is the specified Assignment Group belongs to TOC
// FIXME: check AG prefixes for TOC
func isTOCAssignmentGroup(ag string) bool {
	return strings.HasPrefix(ag, "TOC")
}

// isValidGitHubRepo returns an error if the specifed GitHub repo is invalid or empty
func isValidGitHubRepo(gh string) error {
	if !gitHubPattern.MatchString(gh) {
		return fmt.Errorf(`Invalid GitHub repo "%s"  (valid pattern="%s")`, gh, gitHubPattern.String())
	}
	return nil
}

var gitHubPattern = regexp.MustCompile(`^https://github.ibm.com/`)
