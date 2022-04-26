package ossmerge

import (
	"fmt"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/rest"

	"github.ibm.com/cloud-sre/osscatalog/rmc"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// checkMissingSources checks for missing sources and generates appropriate ValidationIssues.
// This method also takes into account a number of exception cases.
//
// Note: we cannot do this before we completed all the other merges,
// because some of the merged information (e.g. EntryType, OperationalStatus) might be used
// to determine which entries might not need to be registered in some
// of the sources
// TODO: Use attributes from Catalog even for non-visible entries
func (si *ServiceInfo) checkMissingSources() {
	entryType := si.OSSService.GeneralInfo.EntryType
	status := si.OSSService.GeneralInfo.OperationalStatus
	ossTags := &si.OSSService.GeneralInfo.OSSTags

	if !si.HasSourceRMC() {
		var isMinor bool
	allChecks:
		for {
			switch status {
			case ossrecord.DEPRECATED,
				ossrecord.RETIRED,
				ossrecord.NOTREADY,
				ossrecord.THIRDPARTY,
				ossrecord.COMMUNITY:
				isMinor = true
				break allChecks
			}

			switch entryType {
			case ossrecord.OTHEROSS,
				ossrecord.GAAS,
				ossrecord.SUBCOMPONENT,
				ossrecord.SUPERCOMPONENT,
				ossrecord.IAMONLY,
				ossrecord.IAAS,
				ossrecord.VMWARE,
				ossrecord.CONTENT,
				ossrecord.CONSULTING,
				ossrecord.INTERNALSERVICE,
				ossrecord.EntryTypeUnknown:
				isMinor = true
				break allChecks

			case ossrecord.PLATFORMCOMPONENT:
				if !si.HasSourceMainCatalog() {
					isMinor = true
					break allChecks
				}

			}
			break allChecks // Ensure we go through only once
		}
		if isMinor {
			si.AddValidationIssue(ossvalidation.MINOR, "Entry is missing from RMC -- for an entry that might legitimately be OSS-only in RMC", "EntryType=%v  Status=%v", entryType, status).TagCRN().TagRunAction(ossrunactions.RMC)
		} else {
			si.AddValidationIssue(ossvalidation.WARNING, "Entry is missing from RMC -- for an entry that is normally required as a main RMC entry", "EntryType=%v  Status=%v", entryType, status).TagCRN().TagRunAction(ossrunactions.RMC)
		}
	} else if si.OSSService.GetOSSOnboardingPhase() == ossrecord.INVALID {
		si.AddValidationIssue(ossvalidation.WARNING, "Entry is in RMC but has OSSOnboardingPhase=INVALID", "EntryType=%v  Status=%v", entryType, status).TagCRN().TagRunAction(ossrunactions.RMC)
	}

	if !si.HasSourceMainCatalog() {
		switch {
		case ossTags.Contains(osstags.NotReady):
		case status == ossrecord.DEPRECATED: // TODO: should check for Catalog entry with "deprecated" tag
		case status == ossrecord.RETIRED:
		case status == ossrecord.NOTREADY:
		case entryType == ossrecord.OTHEROSS:
		case entryType == ossrecord.GAAS:
		case entryType == ossrecord.SUBCOMPONENT:
		case entryType == ossrecord.SUPERCOMPONENT:
		case entryType == ossrecord.IAMONLY:
		case entryType == ossrecord.CONTENT:
		case entryType == ossrecord.INTERNALSERVICE:
			//case ossTags.Contains(osstags.LimitedAvailability):
		// in none of the cases above do we require a Catalog entry
		// break (not needed as no automatic fallthrough for any of the case clauses)

		case entryType == ossrecord.PLATFORMCOMPONENT:
			si.AddValidationIssue(ossvalidation.WARNING, "PLATFORMCOMPONENT entry is missing from Global Catalog (or private/inactive)", "").TagCRN()
		default:
			si.AddValidationIssue(ossvalidation.SEVERE, "Entry is missing from Global Catalog (or private/inactive)", "").TagCRN()
		}
		/*
				group_not_in_catalog = [
			        DEPRECATED,
			        NOT_IN_IBM_CLOUD,
			    #    IAAS,
			    #    IAAS_LEGACY,
			    #    IAAS_NEW,
			        NOTREADY,
			        COMPONENT,
					SUBCOMPONENT,
					SUPERCOMPONENT,
			        SUBCOMPONENT_CLIENT_FACING,
			        NOT_A_SERVICE,
			        DARK_LAUNCH,
			        RC_NOT_CATALOG,
			        CF_OBSOLETE
					]
		*/
	} else {
		switch entryType {
		case ossrecord.GAAS, ossrecord.OTHEROSS, ossrecord.SUBCOMPONENT, ossrecord.SUPERCOMPONENT, ossrecord.CONTENT, ossrecord.CONSULTING, ossrecord.INTERNALSERVICE:
			si.AddValidationIssue(ossvalidation.WARNING, "Entry found in Global Catalog, not expected for this type", "type=%s", entryType).TagCRN()
		}
	}

	if ossrunactions.ScorecardV1.IsEnabled() {
		si.OSSValidation.RecordRunAction(ossrunactions.ScorecardV1)
		if !si.HasSourceScorecardV1Detail() {
			switch {
			case ossTags.Contains(osstags.NotReady):
			case status == ossrecord.DEPRECATED:
			case status == ossrecord.RETIRED:
			case status == ossrecord.THIRDPARTY:
			case status == ossrecord.COMMUNITY:
			//		case status == ossrecord.LIMITEDAVAILABILIY:
			case entryType == ossrecord.SUBCOMPONENT:
			case entryType == ossrecord.SUPERCOMPONENT:
			//		case entryType == ossrecord.COMPOSITE:
			case entryType == ossrecord.OTHEROSS:
			case entryType == ossrecord.IAMONLY:
			case entryType == ossrecord.CONTENT:
			case entryType == ossrecord.CONSULTING:
			case entryType == ossrecord.INTERNALSERVICE:
				// break (not needed as no automatic fallthrough for any of the case clauses)

			case entryType == ossrecord.GAAS:
				si.AddValidationIssue(ossvalidation.WARNING, "GAAS entry is missing from ScorecardV1", "").TagCRN()
			case entryType == ossrecord.PLATFORMCOMPONENT:
				si.AddValidationIssue(ossvalidation.WARNING, "PLATFORMCOMPONENT entry is missing from ScorecardV1", "").TagCRN()
			default:
				si.AddValidationIssue(ossvalidation.SEVERE, "Entry is missing from ScorecardV1", "").TagCRN()
			}
			/*
					group_not_in_scorecardv1 = [
				        DEPRECATED,
				        NOT_IN_IBM_CLOUD,
				    #    IAAS,
				    #    IAAS_LEGACY,
				    #   IAAS_NEW,
				        NOTREADY,
				        COMPONENT,
						SUBCOMPONENT,
						SUPERCOMPONENT,
				        SUBCOMPONENT_CLIENT_FACING,
				        NOT_A_SERVICE,
				    #   DARK_LAUNCH,
				        RC_NOT_CATALOG,
				        CF_OBSOLETE
						]
			*/
		}
	} else {
		si.OSSValidation.CopyRunAction(si.GetPriorOSSValidation(), ossrunactions.ScorecardV1)
		if si.HasPriorOSS() && si.HasPriorOSSValidation() {
			// Copy the ScorecardV1 source names from prior run
			prior := si.GetPriorOSSValidation()
			names := collections.NewStringSet()
			names.Add(prior.SourceNames(ossvalidation.SCORECARDV1)...)
			names.Add(prior.SourceNames(ossvalidation.SCORECARDV1DISABLED)...)
			if names.Len() > 0 {
				si.AddValidationIssue(ossvalidation.INFO, "Copying prior source names from ScorecardV1 (but not actually fetching new data from ScorecardV1)", "%q", names.Slice()).TagCRN().TagPriorOSS()
				for _, n := range names.Slice() {
					si.OSSValidation.AddSource(n, ossvalidation.SCORECARDV1DISABLED)
				}
			}
		}
	}

	if !si.HasSourceServiceNow() || si.GetSourceServiceNow().IsRetired() {
		switch {
		case ossTags.Contains(osstags.NotReady):
		case status == ossrecord.THIRDPARTY:
		case status == ossrecord.COMMUNITY:
		case status == ossrecord.RETIRED:
			// break (not needed as no automatic fallthrough for any of the case clauses)
		default:
			if !si.HasSourceServiceNow() {
				si.AddValidationIssue(ossvalidation.SEVERE, "Entry is missing from ServiceNow", "").TagCRN()
			} else if si.GetSourceServiceNow().IsRetired() {
				si.AddValidationIssue(ossvalidation.SEVERE, "Entry is in ServiceNow but in RETIRED state", "").TagCRN()
			} else {
				panic("Logic error in checkMissingSources")
			}
		}
		/*
				group_not_in_servicenow = [
			    #    IAAS,
			        NOTREADY,
			        RC_NOT_CATALOG,
			        CF_OBSOLETE
					]
		*/
	}

	if si.OSSService.GeneralInfo.EntryType == ossrecord.IAMONLY && (si.HasSourceServiceNow() || si.HasSourceScorecardV1Detail()) {
		si.AddValidationIssue(ossvalidation.CRITICAL, "Entry is of type IAM_ONLY but it has source records in ServiceNow or ScorecardV1", `ServiceNow=%q   ScorecardV1=%q`, si.SourceServiceNow.CRNServiceName, si.SourceScorecardV1Detail.Name).TagCRN().TagIAM()
	}

}

// mergeReferenceResourceName examines all four possible record sources (PriorOSS, Catalog, ServiceNow and ScorecardV1)
// and returns the best possible resource name between them. It generates ValidationIssues if any of the
// source are missing, have mismatched names, or have names that do not conform to the canonical format.
func (si *ServiceInfo) mergeReferenceResourceName() ossrecord.CRNServiceName {
	var newName ossrecord.CRNServiceName
	var mustCheckCatalogGroup bool

	if si.HasPriorOSS() {
		if si.PriorOSS.ReferenceResourceName != "" {
			canonical := MakeCanonicalName(string(si.PriorOSS.ReferenceResourceName))
			if si.PriorOSS.ReferenceResourceName == canonical {
				newName = si.PriorOSS.ReferenceResourceName
			} else {
				si.AddValidationIssue(ossvalidation.CRITICAL, "Entry name in prior OSS record in Catalog is not in canonical format", `name="%s" - converting to "%s" for new OSS record`, si.PriorOSS.ReferenceResourceName, canonical).TagCRN()
				newName = canonical
			}
			si.OSSValidation.AddSource(string(si.PriorOSS.ReferenceResourceName), ossvalidation.PRIOROSS)
		} else {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Found a prior OSS record in Catalog but its ReferenceResourceName is missing", "").TagCRN()
		}
	}

	if si.HasSourceMainCatalog() {
		if si.SourceMainCatalog.Name != "" {
			if newName != "" {
				canonical, isComposite := ConvertCompositeToCanonicalName(si.SourceMainCatalog.Name)
				if isComposite {
					if canonical != newName {
						si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in Catalog looks like a Composite child name but does not otherwise match the canonical name format", `name="%s" - canonical_name="%s`, si.GetSourceMainCatalog().Name, newName).TagCRN()
					}
				} else {
					if ossrecord.CRNServiceName(si.SourceMainCatalog.Name) != newName {
						si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in main Catalog entry does not match the canonical name from the prior OSS record (keeping the name from prior OSS record)", `name="%s" - canonical_name="%s"`, si.GetSourceMainCatalog().Name, newName).TagCRN()
					}
				}
			} else {
				canonical, isComposite := ConvertCompositeToCanonicalName(si.SourceMainCatalog.Name)
				if isComposite {
					newName = canonical
				} else {
					if ossrecord.CRNServiceName(si.SourceMainCatalog.Name) == canonical {
						newName = canonical
					} else {
						si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in Catalog is not in canonical format", `name="%s" - converting to "%s" for OSS record`, si.GetSourceMainCatalog().Name, canonical).TagCRN()
						newName = canonical
					}
				}
			}
			if si.SourceMainCatalog.Group {
				// Hold off on committing to this Catalog source until we determine if this Catalog Group entry should be ignored,
				// which itself depends on whether we found other valid sources
				mustCheckCatalogGroup = true
			} else {
				si.OSSValidation.AddSource(string(si.SourceMainCatalog.Name), ossvalidation.CATALOG)
			}
			if si.SourceMainCatalog.ID != "" {
				si.OSSService.ReferenceCatalogID = ossrecord.CatalogID(si.SourceMainCatalog.ID)
				si.OSSService.ReferenceCatalogPath = si.SourceMainCatalog.CatalogPath
			} else {
				si.AddValidationIssue(ossvalidation.CRITICAL, "ID in main Catalog entry is empty", "").TagCRN()
			}
		} else {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Entry name in main Catalog entry is empty", "").TagCRN()
		}
	}

	if si.HasSourceServiceNow() {
		if si.SourceServiceNow.CRNServiceName != "" {
			if newName != "" {
				if ossrecord.CRNServiceName(si.SourceServiceNow.CRNServiceName) != newName {
					si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in ServiceNow does not match the canonical name", `name="%s" - canonical_name="%s"`, si.SourceServiceNow.CRNServiceName, newName).TagCRN()
				}
			} else {
				canonical := MakeCanonicalName(si.SourceServiceNow.CRNServiceName)
				if ossrecord.CRNServiceName(si.SourceServiceNow.CRNServiceName) != canonical {
					si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in ServiceNow is not in canonical format", `name="%s" - converting to "%s" for OSS record`, si.SourceServiceNow.CRNServiceName, canonical).TagCRN()
				}
				si.AddValidationIssue(ossvalidation.INFO, "Did not get an entry name from Catalog - using name from ServiceNow entry instead", `name="%s"`, canonical).TagCRN()
				newName = canonical
			}
			if si.GetSourceServiceNow().IsRetired() {
				si.OSSValidation.AddSource(si.SourceServiceNow.CRNServiceName, ossvalidation.SERVICENOWRETIRED)
			} else {
				si.OSSValidation.AddSource(si.SourceServiceNow.CRNServiceName, ossvalidation.SERVICENOW)
			}
			si.OSSService.GeneralInfo.OSSTags.AddTag(osstags.ServiceNowApproved)
		} else {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Entry name in ServiceNow is empty", "").TagCRN()
		}
	}

	if si.HasSourceScorecardV1Detail() {
		if si.SourceScorecardV1Detail.Name != "" {
			if newName != "" {
				if ossrecord.CRNServiceName(si.SourceScorecardV1Detail.Name) != newName {
					si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in ScorecardV1 does not match the canonical name", `name="%s" - canonical_name="%s"`, si.SourceScorecardV1Detail.Name, newName).TagCRN()
				}
			} else {
				canonical := MakeCanonicalName(si.SourceScorecardV1Detail.Name)
				if ossrecord.CRNServiceName(si.SourceScorecardV1Detail.Name) != canonical {
					si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in ScorecardV1 is not in canonical format", `name="%s" - converting to "%s" for OSS record`, si.SourceServiceNow.CRNServiceName, canonical).TagCRN()
				}
				si.AddValidationIssue(ossvalidation.INFO, "Did not get an entry name from Catalog or ServiceNow - using name from ScorecardV1 entry instead", `name="%s"`, canonical).TagCRN()
				newName = canonical
			}
			si.OSSValidation.AddSource(string(si.SourceScorecardV1Detail.Name), ossvalidation.SCORECARDV1)
		} else {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Entry name in ScorecardV1 is empty", "").TagCRN()
		}
	}

	if si.HasSourceIAM() {
		if si.SourceIAM.Name != "" {
			if si.HasSourceMainCatalog() {
				if si.GetSourceMainCatalog().Name != si.SourceIAM.Name {
					si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in IAM does not match the Catalog name", `IAM_name="%s" - Catalog_name="%s"`, si.SourceIAM.Name, si.GetSourceMainCatalog().Name).TagCRN().TagIAM()
				}
			} else if newName != "" {
				canonical, isComposite := ConvertCompositeToCanonicalName(si.SourceIAM.Name)
				if isComposite {
					if canonical != newName {
						si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in IAM looks like a Composite child name but does not otherwise match the canonical name format", `IAM_name="%s" - canonical_name="%s`, si.GetSourceIAM().Name, newName).TagCRN().TagIAM()
					}
				} else {
					if ossrecord.CRNServiceName(si.SourceIAM.Name) != newName {
						si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in IAM does not match the canonical name", `IAM_name="%s" - canonical_name="%s"`, si.SourceIAM.Name, newName).TagCRN().TagIAM()
					}
				}
			} else {
				canonical, isComposite := ConvertCompositeToCanonicalName(si.SourceIAM.Name)
				if isComposite {
					// nothing to do
				} else {
					if ossrecord.CRNServiceName(si.SourceIAM.Name) != canonical {
						si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in IAM is not in canonical format", `IAM_name="%s" - converting to "%s" for OSS record`, si.SourceIAM.Name, canonical).TagCRN().TagIAM()
					}
				}
				/* IAM-only entries will be processed later on in mergeEntryType()
				debug.Info(`Ignoring IAM-only entry (no sources in Catalog, ServiceNow, Scorecard): %q`, si.SourceIAM.Name)
				si.AddValidationIssue(ossvalidation.INFO, "Only name found is in IAM (no sources in Catalog, ServiceNow, Scorecard) -- ignoring this entry", `IAM_name="%s"`, si.SourceIAM.Name).TagCRN().TagIAM()
				//newName = canonical
				*/
			}
			if si.SourceIAM.Enabled {
				si.OSSValidation.AddSource(string(si.SourceIAM.Name), ossvalidation.IAM)
			} else {
				si.OSSValidation.AddSource(string(si.SourceIAM.Name), ossvalidation.IAMDISABLED)
			}
		} else {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Entry name in IAM is empty", "").TagCRN().TagIAM()
		}
	}

	// TODO: Check and merge additional (duplicate) sources beyond the normal 3

	if si.AdditionalMainCatalog != nil {
		var names []string
		var namesIgnored []string
		for _, r := range si.AdditionalMainCatalog {
			var ignored bool
			if si.isMainCatalogIgnored(r) {
				si.OSSValidation.AddSource(string(r.Name), ossvalidation.CATALOGIGNORED)
				namesIgnored = append(namesIgnored, string(r.String()))
				ignored = true
			} else {
				si.OSSValidation.AddSource(string(r.Name), ossvalidation.CATALOG)
				names = append(names, string(r.String()))
			}
			var visibilityRestrictions string
			if r.EffectiveVisibility.Restrictions != r.Visibility.Restrictions {
				visibilityRestrictions = fmt.Sprintf("effective=%s/local=%s", r.EffectiveVisibility.Restrictions, r.Visibility.Restrictions)
			} else {
				visibilityRestrictions = r.Visibility.Restrictions
			}
			si.AddValidationIssue(ossvalidation.INFO, "Found additional Global Catalog entry", `%s  Ignored=%v  Restrictions:%v  Active:%v  Hidden:%v  Disabled:%v`, r.String(), ignored, visibilityRestrictions, r.Active, r.ObjectMetaData.UI.Hidden, r.Disabled).TagCRN()
		}
		if len(names) > 0 {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Found more than one entry with similar names in Global Catalog", `main=%s  additional=%v  ignored=%v`, si.SourceMainCatalog.String(), names, namesIgnored).TagCRN()
		} else {
			si.AddValidationIssue(ossvalidation.WARNING, "Found more than one entry with similar names in Global Catalog (but all duplicates are ignored/not publicly visible)", `main=%s  additional=%v  ignored=%v`, si.SourceMainCatalog.String(), names, namesIgnored).TagCRN()
		}
	}
	if si.AdditionalServiceNow != nil {
		var names []string
		var namesIgnored []string
		for _, r := range si.AdditionalServiceNow {
			var ignored bool
			if r.IsRetired() {
				si.OSSValidation.AddSource(string(r.CRNServiceName), ossvalidation.SERVICENOWRETIRED)
				namesIgnored = append(namesIgnored, string(r.CRNServiceName))
				ignored = true
			} else {
				si.OSSValidation.AddSource(string(r.CRNServiceName), ossvalidation.SERVICENOW)
				names = append(names, string(r.CRNServiceName))
			}
			si.AddValidationIssue(ossvalidation.INFO, "Found additional ServiceNow entry", `name="%s"  Ignored=%v  Status:%v   ID=%v`, r.CRNServiceName, ignored, r.GeneralInfo.OperationalStatus, r.SysID).TagCRN()
		}
		if len(names) > 0 {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Found more than one entry with similar names in ServiceNow", `main_name="%s"  additional=%v  ignored=%v`, si.SourceServiceNow.CRNServiceName, names, namesIgnored).TagCRN()
		} else {
			si.AddValidationIssue(ossvalidation.INFO, "Found more than one entry with similar names in ServiceNow (but all duplicates are ignored/retired)", `main_name="%s"  additional=%v  ignored=%v`, si.SourceServiceNow.CRNServiceName, names, namesIgnored).TagCRN()
		}
	}
	if si.AdditionalScorecardV1Detail != nil {
		var names []string
		var ignored bool
		for _, r := range si.AdditionalScorecardV1Detail {
			si.OSSValidation.AddSource(string(r.Name), ossvalidation.SCORECARDV1)
			names = append(names, string(r.Name))
			si.AddValidationIssue(ossvalidation.INFO, "Found additional ScorecardV1 entry", `name="%s"  Ignored=%v  Status:%v`, r.Name, ignored, r.Status).TagCRN()
		}
		si.AddValidationIssue(ossvalidation.CRITICAL, "Found more than one entry with similar names in ScorecardV1", `main_name="%s"  additional=%v`, si.SourceScorecardV1Detail.Name, names).TagCRN()
	}
	if si.AdditionalIAM != nil {
		var namesEnabled []string
		var namesDisabled []string
		for _, r := range si.AdditionalIAM {
			if r.Enabled {
				si.OSSValidation.AddSource(string(r.Name), ossvalidation.IAM)
				namesEnabled = append(namesEnabled, string(r.Name))
			} else {
				si.OSSValidation.AddSource(string(r.Name), ossvalidation.IAMDISABLED)
				namesDisabled = append(namesDisabled, string(r.Name))
			}
			si.AddValidationIssue(ossvalidation.INFO, "Found additional IAM entry", `name="%s"  Enabled:%v`, r.Name, r.Enabled).TagCRN().TagIAM()
		}
		if len(namesEnabled) > 0 {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Found more than one entry with similar names in IAM - at least some of them enabled", `main_name="%s"  additional_enabled=%q  additional_disabled=%q`, si.SourceIAM.Name, namesEnabled, namesDisabled).TagCRN().TagIAM()
		} else {
			si.AddValidationIssue(ossvalidation.INFO, "Found more than one entry with similar names in IAM - all additional entries disabled", `main_name="%s"  additional_enabled=%q  additional_disabled=%q`, si.SourceIAM.Name, namesEnabled, namesDisabled).TagCRN().TagIAM()
		}
	}

	// Check for ignored entries
	if si.IgnoredMainCatalog != nil {
		si.OSSValidation.AddSource(si.IgnoredMainCatalog.Name, ossvalidation.CATALOGIGNORED)
	}

	// Check Catalog Group entry
	if mustCheckCatalogGroup {
		if si.checkIgnoreCatalogGroup() {
			// blank the Catalog entry
			// XXX Note that even though we decided to ignore this Catalog entry, we already selected its name as the best possible canonical name.
			// This is not an issue if there are no other sources (which is normally the case if we decide to ignore, but it may change in the future)
			si.OSSService.ReferenceCatalogID = ""
			si.OSSService.ReferenceCatalogPath = ""
			saved := *si.GetSourceMainCatalog() // Make a copy
			if si.IgnoredMainCatalog == nil {
				si.IgnoredMainCatalog = &saved
			} else {
				si.AddValidationIssue(ossvalidation.WARNING, "Ignoring Catalog group entry but there is already another ignored Catalog Entry", "group entry=%s   prior ignored entry=%s", saved.String(), si.IgnoredMainCatalog.String())
			}
			si.SourceMainCatalog = catalogapi.Resource{}
			si.OSSValidation.AddSource(si.SourceMainCatalog.Name, ossvalidation.CATALOGIGNORED)
		} else {
			si.OSSValidation.AddSource(string(si.SourceMainCatalog.Name), ossvalidation.CATALOG)
		}
	}

	if newName != "" {
		si.OSSValidation.SetSourceNameCanonical(string(newName))
		if isNameNotMergeable(string(newName)) {
			si.AddValidationIssue(ossvalidation.WARNING, "Not merging this entry with other entries with a \"comparable name\" because it is listed as a \"do-not-merge\" name", "").TagCRN()
			if !IsNameCanonical(string(newName)) {
				si.AddValidationIssue(ossvalidation.WARNING, "Special \"do-not-merge\" name for this entry is not in canonical form", "").TagCRN()
			}
		}
		if err := checkValidCRNServiceName(newName); err != nil {
			if si.GeneralInfo.OSSTags.Contains(osstags.LenientCRNName) {
				si.AddValidationIssue(ossvalidation.MINOR, "CRN service-name does not meet IBM Cloud naming standards (allowed with lenient mode enabled for this entry)", "%v", err).TagCRN()
			} else {
				si.AddValidationIssue(ossvalidation.WARNING, "CRN service-name does not meet IBM Cloud naming standards", "%v", err).TagCRN()
			}
		}
	} else {
		if si.Legacy != nil {
			si.AddValidationIssue(ossvalidation.WARNING, "Entry is found in the Legacy validation report but not in any of the current sources", "Legacy sources: %v", si.Legacy.AllSources).TagCRN()
		}
	}

	// Check for RMC entries
	// TODO: skip searching for third-party services
	var rmcPriorNames []string
	var allPriorNames []string
	if si.HasPriorOSSValidation() {
		rmcPriorNames = si.GetPriorOSSValidation().SourceNames(ossvalidation.RMC)
		allPriorNames = si.GetPriorOSSValidation().AllNames()
	}
	if ossrunactions.RMCRescan.IsEnabled() || (ossrunactions.RMC.IsEnabled() && len(rmcPriorNames) > 0) {
		si.OSSValidation.RecordRunAction(ossrunactions.RMC) // Note that we treat ossrunctions.RMCRescan as just a special case of ossrunactions.RMC
		sources := si.OSSValidation.AllSources()
		if len(sources) > 1 || (len(sources) == 1 && sources[0] != ossvalidation.IAM && sources[0] != ossvalidation.IAMDISABLED) {
			// Load any RMC entry that matches any of the known names
			names := collections.NewStringSet(si.OSSValidation.AllNames()...)
			names.Add(allPriorNames...)
			var potentialCompositeName string
			if si.HasPriorOSS() && si.GeneralInfo.OSSOnboardingPhase != "" {
				// Check for a potential composite name e.g. "is-instance"
				ref := string(si.GetPriorOSS().ReferenceResourceName)
				if ix := strings.Index(ref, "-"); ix > 0 {
					potentialCompositeName = fmt.Sprintf("%s.%s", ref[0:ix], ref[ix+1:])
					names.Add(potentialCompositeName)
				}
			}
			var foundPotentialCompositeName bool
			for _, n := range names.Slice() {
				if foundPotentialCompositeName && n == potentialCompositeName {
					continue
				}
				r, err := rmc.ReadRMCSummaryEntry(ossrecord.CRNServiceName(n), false)
				if err == nil {
					if n == potentialCompositeName {
						foundPotentialCompositeName = true
					}
					if !si.HasSourceRMC() {
						si.SourceRMC = *r
						si.OSSValidation.AddSource(string(r.CRNServiceName), ossvalidation.RMC)
					} else {
						var prior *rmc.SummaryEntry
						if si.SourceRMC.CRNServiceName == r.CRNServiceName {
							prior = si.GetSourceRMC()
						} else {
							for _, r0 := range si.AdditionalRMC {
								if r0.CRNServiceName == r.CRNServiceName {
									prior = r0
									break
								}
							}
						}
						if prior != nil {
							if prior.ID != r.ID {
								si.AddValidationIssue(ossvalidation.CRITICAL, "Found two RMC records with the same CRNServiceName but different IDs", "name=%q  id1=%q  id2=%q",
									prior.CRNServiceName, prior.ID, r.ID).TagCRN().TagCRN().TagRunAction(ossrunactions.RMC)
							}
							// Do not record the RMC entry for a second time
						} else {
							si.AdditionalRMC = append(si.AdditionalRMC, r)
							si.OSSValidation.AddSource(string(r.CRNServiceName), ossvalidation.RMC)
						}
					}
				} else if rest.IsEntryNotFound(err) {
					// ignore
				} else {
					debug.PrintError(`Error reading RMC record "%s" for entry %s: %v`, n, si.String(), err)
				}
			}

			// Look for RMC entries in the test server
			if options.GlobalOptions().TestMode {
				if si.HasPriorOSS() && si.GeneralInfo.OSSOnboardingPhase != "" {
					var names []string
					if potentialCompositeName != "" {
						names = []string{potentialCompositeName, string(si.GetPriorOSS().ReferenceResourceName)}
					} else {
						names = []string{string(si.GetPriorOSS().ReferenceResourceName)}
					}
					var foundPotentialCompositeName bool
					for _, n := range names {
						if foundPotentialCompositeName && n == potentialCompositeName {
							continue
						}
						r, err := rmc.ReadRMCSummaryEntry(ossrecord.CRNServiceName(n), true)
						if err == nil {
							if n == potentialCompositeName {
								foundPotentialCompositeName = true
							}
							if si.HasSourceRMC() {
								debug.Warning("Test Mode: ignoring test entry from RMC that also exists in main(staging) RMC: %s", n)
								si.AddValidationIssue(ossvalidation.WARNING, "Ignoring test entry from RMC that also exists in main(staging) RMC", "").TagTest()
							} else {
								si.SourceRMC = *r
								si.OSSValidation.AddSource(string(r.CRNServiceName), ossvalidation.RMC)
								si.OSSMergeControl.OSSTags.AddTag(osstags.OSSTest)
								si.OSSService.GeneralInfo.OSSTags.AddTag(osstags.OSSTest)
								debug.Info("Test Mode: loading test entry from RMC: %s", n)
								si.AddValidationIssue(ossvalidation.INFO, "Entry loaded from test instance of RMC", "").TagTest()
								// Stop at the first match in Test; do not try to accumulate AdditionalRMC entries
								break
							}
						} else if rest.IsEntryNotFound(err) {
							// ignore
						} else {
							debug.PrintError(`(Test Mode): Error reading RMC record "%s" for entry %s: %v`, n, si.String(), err)
						}
					}
				}
			}

			// Consistency checks
			if si.HasSourceRMC() {
				// TODO: check if RMC entry is deleted/retired
				if newName != "" {
					canonical, isComposite := ConvertCompositeToCanonicalName(string(si.GetSourceRMC().CRNServiceName))
					if isComposite {
						if canonical != newName {
							si.AddValidationIssue(ossvalidation.SEVERE, "Entry name in RMC looks like a Composite child name but does not otherwise match the canonical name format", `RMC="%s" - canonical_name="%s"`, si.SourceRMC.CRNServiceName, newName).TagCRN().TagRunAction(ossrunactions.RMC)
						}
					} else {
						if ossrecord.CRNServiceName(si.GetSourceRMC().CRNServiceName) != newName {
							si.AddValidationIssue(ossvalidation.SEVERE, "RMC entry: CRN service-name in RMC does not match the canonical name", `RMC="%s" - canonical_name="%s"`, si.GetSourceRMC().CRNServiceName, newName).TagCRN().TagRunAction(ossrunactions.RMC)
						}
					}
				}
				if string(si.GetSourceRMC().CRNServiceName) != si.GetSourceRMC().Name {
					si.AddValidationIssue(ossvalidation.MINOR, "RMC entry: CRN service-name in RMC does not match the RMC name attribute", `RMC CRN service-name="%s" - name="%s"`, si.GetSourceRMC().CRNServiceName, si.GetSourceRMC().Name).TagCRN().TagRunAction(ossrunactions.RMC)
				}
				if si.OSSService.ReferenceCatalogID != "" && string(si.OSSService.ReferenceCatalogID) != si.GetSourceRMC().ID {
					si.AddValidationIssue(ossvalidation.CRITICAL, "ReferenceCatalogID has different value in Catalog than RMC (Catalog prevails)",
						`Catalog="%v"   RMC="%v"`, si.OSSService.ReferenceCatalogID, si.GetSourceRMC().ID).TagDataMismatch().TagRunAction(ossrunactions.RMC)
				}
			} else {
				if si.GeneralInfo.OSSOnboardingPhase != "" {
					if si.GeneralInfo.OSSOnboardingPhase == ossrecord.INVALID {
						si.AddValidationIssue(ossvalidation.INFO, "Entry has OSSOnboardingPhase=INVALID but is not found in RMC", "").TagCRN().TagRunAction(ossrunactions.RMC)
					} else {
						si.AddValidationIssue(ossvalidation.SEVERE, "Entry has non-empty OSSOnboardingPhase but is not found in RMC - setting to INVALID", "prior OSSOnboardingPhase=%s", si.GeneralInfo.OSSOnboardingPhase).TagCRN().TagRunAction(ossrunactions.RMC)
						si.GeneralInfo.OSSOnboardingPhase = ossrecord.INVALID
					}
				}
			}
			if si.AdditionalRMC != nil {
				var names []string
				var namesIgnored []string
				for _, r := range si.AdditionalRMC {
					var ignored bool
					if false { // TODO: check if RMC entry is deleted/retired
						si.OSSValidation.AddSource(string(r.CRNServiceName), ossvalidation.RMC)
						namesIgnored = append(namesIgnored, string(r.CRNServiceName))
						ignored = true
					} else {
						si.OSSValidation.AddSource(string(r.CRNServiceName), ossvalidation.RMC)
						names = append(names, string(r.CRNServiceName))
					}
					si.AddValidationIssue(ossvalidation.INFO, "Found additional RMC entry", `RMC CRN service-name="%s"  name="%s"  Ignored=%v  Maturity:%v  Type=%v   ID=%v`, r.CRNServiceName, r.Name, ignored, r.Maturity, r.Type, r.ID).TagCRN().TagRunAction(ossrunactions.RMC)
				}
				if len(names) > 0 {
					si.AddValidationIssue(ossvalidation.CRITICAL, "Found more than one entry with similar names in RMC", `main_name="%s"  additional=%v  ignored=%v`, si.SourceRMC.CRNServiceName, names, namesIgnored).TagCRN().TagRunAction(ossrunactions.RMC)
				} else {
					si.AddValidationIssue(ossvalidation.INFO, "Found more than one entry with similar names in RMC (but all duplicates are ignored/retired)", `main_name="%s"  additional=%v  ignored=%v`, si.SourceRMC.CRNServiceName, names, namesIgnored).TagCRN().TagRunAction(ossrunactions.RMC)
				}
			}
		}
	} else {
		si.OSSValidation.CopyRunAction(si.GetPriorOSSValidation(), ossrunactions.RMC)
		for _, n := range rmcPriorNames {
			si.OSSValidation.AddSource(n, ossvalidation.RMC)
		}
	}

	// Note: if the entry was private in Catalog and not found anywhere else, we might end-up with no name at all

	return newName
}

// mergeReferenceDisplayName compares the DisplayName between three possible record sources (Catalog, ServiceNow and ScorecardV1)
// and generates ValidationIssues for any discrepancies or missing DisplayNames.
func (si *ServiceInfo) mergeReferenceDisplayName() string {
	// We cannot use the standard MergeValues function here because we want a different severity
	// if the DisplayName is missing from Catalog or ServiceNow than from ScorecardV1
	// result := si.MergeValues("DisplayName", SeverityIfMissing{ossvalidation.CRITICAL}, Catalog{si.SourceMainCatalog.OverviewUI.En.DisplayName}, ServiceNow{si.SourceServiceNow.DisplayName}, ScorecardV1{si.SourceScorecardV1Detail.DisplayName}, PriorOSS{si.PriorOSS.ReferenceDisplayName}).(string)

	var mergedVal string
	var mergedValSource string
	var numValidSources int

	if si.HasSourceMainCatalog() {
		numValidSources++
		if si.GetSourceMainCatalog().OverviewUI.En.DisplayName == "" {
			si.AddValidationIssue(ossvalidation.CRITICAL, "DisplayName is missing from Catalog", "").TagDataMissing()
		} else if mergedVal != "" {
			if mergedVal != si.GetSourceMainCatalog().OverviewUI.En.DisplayName {
				si.AddValidationIssue(ossvalidation.WARNING, "DisplayName has different value in Catalog than other sources (first source prevails)",
					`%s="%v"   %s="%v"`, mergedValSource, mergedVal, "Catalog", si.GetSourceMainCatalog().OverviewUI.En.DisplayName).TagDataMismatch()
			}
		} else {
			mergedVal = si.GetSourceMainCatalog().OverviewUI.En.DisplayName
			mergedValSource = "Catalog"
		}
	}

	if si.HasSourceRMC() {
		numValidSources++
		if si.GetSourceRMC().DisplayName == "" {
			t, err := si.GetSourceRMC().GetEntryType()
			if err != nil {
				si.AddValidationIssue(ossvalidation.SEVERE, "Cannot determine DisplayName from RMC", "%s", err.Error()).TagDataMismatch().TagRunAction(ossrunactions.RMC)
			} else if t == ossrecord.GAAS {
				si.AddValidationIssue(ossvalidation.INFO, "No DisplayName in RMC -- normal for a Operations_Only entry", "").TagDataMissing().TagRunAction(ossrunactions.RMC)
			} else {
				si.AddValidationIssue(ossvalidation.SEVERE, "DisplayName is missing from RMC", "").TagDataMissing().TagRunAction(ossrunactions.RMC)
			}
		} else if mergedVal != "" {
			if mergedVal != si.GetSourceRMC().DisplayName {
				si.AddValidationIssue(ossvalidation.WARNING, "DisplayName has different value in RMC than Catalog (Catalog prevails)",
					`%s="%v"   %s="%v"`, mergedValSource, mergedVal, "RMC", si.GetSourceRMC().DisplayName).TagDataMismatch()
			}
		} else {
			mergedVal = si.GetSourceRMC().DisplayName
			mergedValSource = "RMC"
		}
	}

	if si.HasSourceServiceNow() {
		numValidSources++
		if si.GetSourceServiceNow().DisplayName == "" {
			si.AddValidationIssue(ossvalidation.SEVERE, "DisplayName is missing from ServiceNow", "").TagDataMissing()
		} else if mergedVal != "" {
			if mergedVal != si.GetSourceServiceNow().DisplayName {
				si.AddValidationIssue(ossvalidation.WARNING, "DisplayName has different value in ServiceNow than other sources (first source prevails)",
					`%s="%v"   %s="%v"`, mergedValSource, mergedVal, "ServiceNow", si.GetSourceServiceNow().DisplayName).TagDataMismatch()
			}
		} else {
			mergedVal = si.GetSourceServiceNow().DisplayName
			mergedValSource = "ServiceNow"
		}
	}

	if si.HasSourceScorecardV1Detail() {
		numValidSources++
		if si.GetSourceScorecardV1Detail().DisplayName == "" {
			si.AddValidationIssue(ossvalidation.MINOR, "DisplayName is missing from ScorecardV1", "").TagDataMissing()
		} else if mergedVal != "" {
			if mergedVal != si.GetSourceScorecardV1Detail().DisplayName {
				si.AddValidationIssue(ossvalidation.WARNING, "DisplayName has different value in ScorecardV1 than other sources (first source prevails)",
					`%s="%v"   %s="%v"`, mergedValSource, mergedVal, "ScorecardV1", si.GetSourceScorecardV1Detail().DisplayName).TagDataMismatch()
			}
		} else {
			mergedVal = si.GetSourceScorecardV1Detail().DisplayName
			mergedValSource = "ScorecardV1"
		}
	}

	if si.HasSourceIAM() {
		numValidSources++
		if si.GetSourceIAM().DisplayName == "" {
			si.AddValidationIssue(ossvalidation.WARNING, "DisplayName is missing from IAM", "").TagDataMissing().TagIAM()
		} else if mergedVal != "" {
			if mergedVal != si.GetSourceIAM().DisplayName {
				si.AddValidationIssue(ossvalidation.WARNING, "DisplayName has different value in IAM than other sources (first source prevails)",
					`%s="%v"   %s="%v"`, mergedValSource, mergedVal, "IAM", si.GetSourceIAM().DisplayName).TagDataMismatch().TagIAM()
			}
		} else {
			mergedVal = si.GetSourceIAM().DisplayName
			mergedValSource = "IAM"
		}
	}

	if mergedVal == "" && si.HasPriorOSS() {
		mergedVal = si.GetPriorOSS().ReferenceDisplayName
		if mergedVal != "" {
			mergedValSource = "PriorOSS"
			si.AddValidationIssue(ossvalidation.MINOR, "DisplayName cannot be set in OSS record from any available source -- using value from PriorOSS record", mergedVal).TagDataMissing().TagPriorOSS()
		}
	}

	// TODO: Check for additional (duplicate) sources beyond the normal 3

	if mergedVal == "" {
		if numValidSources > 0 {
			si.AddValidationIssue(ossvalidation.SEVERE, "DisplayName cannot be set in OSS record from any available source", "").TagDataMissing()
		} else {
			si.AddValidationIssue(ossvalidation.IGNORE, "DisplayName cannot be set in OSS record because there are no source records containing this attribute", "").TagDataMissing()
		}
	}

	isTest := osstags.CheckOSSTestTag(&mergedVal, &si.OSSService.GeneralInfo.OSSTags)

	if !isTest {
		if m := displayNameAntiPattern.FindString(mergedVal); m != "" {
			if si.GeneralInfo.OSSTags.Contains(osstags.LenientDisplayName) {
				si.AddValidationIssue(ossvalidation.MINOR, "DisplayName contains suspicious word not expected in a client-facing DisplayName (allowed with lenient mode enabled for this entry)", `word="%s"   (pattern="%s")`, m, displayNameAntiPattern).TagCRN()
			} else {
				si.AddValidationIssue(ossvalidation.WARNING, "DisplayName contains suspicious word not expected in a client-facing DisplayName", `word="%s"   (pattern="%s")`, m, displayNameAntiPattern).TagCRN()
			}
		}
	}

	return mergedVal
}

var displayNameAntiPattern = regexp.MustCompile(`(GA|Beta|beta|Experimental|experimental|Boilerplate|boilerplate|Dedicated|dedicated|Closed|closed|\()`)

func (si *ServiceInfo) mergeGeneralInfo() {
	ossgi := &si.OSSService.GeneralInfo
	sngi := &si.SourceServiceNow.GeneralInfo
	sc := &si.SourceScorecardV1Detail
	prior := &si.PriorOSS.GeneralInfo

	// TODO: sopID is not a consistent JSON type in JSON output
	//	ossgi.RMCNumber = si.MergeValues("RMC/SOP number", SeverityIfMismatch{ossvalidation.SEVERE}, SeverityIfMissing{ossvalidation.MINOR}, ServiceNow{sngi.RMCNumber}, ScorecardV1{sc.SOPID}, PriorOSS{prior.RMCNumber}).(string)
	ossgi.RMCNumber = si.MergeValues("RMC/SOP number", SeverityIfMismatch{ossvalidation.SEVERE}, SeverityIfMissing{ossvalidation.MINOR}, ServiceNow{sngi.RMCNumber} /*XXX , ScorecardV1{sc.SOPID}*/, PriorOSS{prior.RMCNumber}).(string)
	// OSSTags have already been merged earlier (actually copied from the OSSMergeControl record)
	// FullCRN is deprecated
	if sc.CRN != "" {
		si.AddValidationIssue(ossvalidation.INFO, "Ignoring deprecated full CRN from ScorecardV1", sc.CRN).TagCRN()
	}
	if sngi.FullCRN != "" {
		si.AddValidationIssue(ossvalidation.INFO, "Ignoring deprecated full CRN from ServiceNow", sngi.FullCRN).TagCRN()
	}
	// TODO: Should we build a OSSDescription from other sources if not provided in ServiceNow
	ossgi.OSSDescription = si.MergeValues("OSSDescription", SeverityIfMissing{ossvalidation.IGNORE}, ServiceNow{sngi.OSSDescription}, PriorOSS{prior.OSSDescription}).(string)
	ossgi.ServiceNowSysid = si.MergeValues("ServiceNow Sysid", SeverityIfMissing{ossvalidation.INFO}, ServiceNow{ossrecord.ServiceNowSysid(si.SourceServiceNow.SysID)}, PriorOSS{prior.ServiceNowSysid}).(ossrecord.ServiceNowSysid)
	ossgi.ServiceNowCIURL = si.MergeValues("ServiceNowCIURL", SeverityIfMissing{ossvalidation.SEVERE}, ServiceNow{sngi.ServiceNowCIURL}, PriorOSS{prior.ServiceNowCIURL}).(string)
	// ParentResourceName must be merged in Phase Two, when we have access to all entries to check for consistency between parent and child
}

func (si *ServiceInfo) mergeOperationalStatus() ossrecord.OperationalStatus {
	item := "OperationalStatus"
	var result ossrecord.OperationalStatus
	sngi := &si.SourceServiceNow.GeneralInfo
	prior := &si.PriorOSS.GeneralInfo
	var ok bool
	var err error

	if si.OSSService.GeneralInfo.EntryType == "" {
		panic(fmt.Sprintf(`mergeOperationalStatus(): missing EntryType for "%s"`, si.String()))
	}

	var rmcOSSOperationalStatus ossrecord.OperationalStatus
	var rmcMainOperationalStatus ossrecord.OperationalStatus
	if si.HasSourceRMC() {
		st, err := si.GetSourceRMC().GetOperationalStatus()
		if err != nil {
			si.AddValidationIssue(ossvalidation.SEVERE, "Invalid RMC maturity attribute", err.Error()).TagDataMismatch().TagRunAction(ossrunactions.RMC)
			st = ossrecord.OperationalStatusUnknown
		}
		if si.HasPriorOSS() && prior.OSSOnboardingPhase != "" {
			rmcOSSOperationalStatus = prior.OperationalStatus
			if prior.FutureOperationalStatus != "" {
				switch st {
				case ossrecord.GA, ossrecord.BETA, ossrecord.EXPERIMENTAL:
					if st != prior.FutureOperationalStatus {
						si.AddValidationIssue(ossvalidation.WARNING, "Maturity attribute from main RMC entry does not match the FutureOperationalStatus in the OSS tab of that RMC entry (OSS tab prevails)", `main=%q  oss=%q`, st, prior.FutureOperationalStatus).TagDataMismatch().TagRunAction(ossrunactions.RMC)
					}
				case ossrecord.OperationalStatusUnknown:
					// We've already logged a validation issue for bad operational status
				default:
					// Just ignore the main RMC entry maturity - we have a more specific value in the OSS tab
				}
			} else {
				switch st {
				case ossrecord.GA, ossrecord.BETA, ossrecord.EXPERIMENTAL:
					if st != prior.OperationalStatus {
						si.AddValidationIssue(ossvalidation.WARNING, "Maturity attribute from main RMC entry does not match the current OperationalStatus in the OSS tab of that RMC entry (FutureOperationalStatus not specified; OSS tab prevails)", `main=%q  oss=%q`, st, prior.OperationalStatus).TagDataMismatch().TagRunAction(ossrunactions.RMC)
					}
				case ossrecord.OperationalStatusUnknown:
					// We've already logged a validation issue for bad operational status
				default:
					// Just ignore the main RMC entry maturity - we have a more specific value in the OSS tab
				}
			}
		} else {
			rmcMainOperationalStatus = st
		}
	}
	var forceRetired parameter
	if si.GeneralInfo.OSSOnboardingPhase == ossrecord.INVALID {
		if rmcOSSOperationalStatus != ossrecord.RETIRED {
			si.AddValidationIssue(ossvalidation.MINOR, "Forcing OperationalStatus to RETIRED for entry with OSSOnboardingPhase=INVALID", "prior RMC OperationalStatus=%s", rmcOSSOperationalStatus).TagCRN().TagRunAction(ossrunactions.RMC)
		}
		rmcOSSOperationalStatus = ossrecord.RETIRED
		forceRetired = Custom{N: "Force Retired for OSSOnboardingPhase=INVALID", V: ossrecord.RETIRED}
	}

	var catOperationalStatus ossrecord.OperationalStatus
	if si.HasSourceMainCatalog() {
		catOperationalStatus, err = si.SourceMainCatalog.GetOperationalStatus()
		if err != nil {
			if catOperationalStatus == ossrecord.OperationalStatusUnknown {
				si.AddValidationIssue(ossvalidation.CRITICAL, "Cannot extract a OSS OperationalStatus from the Catalog tags", err.Error()).TagCRN()
				catOperationalStatus = ""
			} else {
				si.AddValidationIssue(ossvalidation.INFO, "Extracting OSS OperationalStatus from multiple Catalog tags", err.Error()).TagCRN()
			}
		}
	}

	var scOperationalStatus ossrecord.OperationalStatus
	if si.HasSourceScorecardV1Detail() {
		scOperationalStatus, ok = si.SourceScorecardV1Detail.GetOperationalStatus()
		if !ok {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Cannot translate Status attribute from ScorecardV1 into a valid OSS OperationalStatus", "%s", si.SourceScorecardV1Detail.Status).TagDataMismatch().TagCRN()
			scOperationalStatus = ""
		}
	} else if !ossrunactions.ScorecardV1.IsEnabled() && si.HasPriorOSS() {
		scOperationalStatus = si.GetPriorOSS().GeneralInfo.OperationalStatus
	}

	var snOperationalStatus ossrecord.OperationalStatus
	if si.HasSourceServiceNow() {
		snOperationalStatus, err = ossrecord.ParseOperationalStatus(string(sngi.OperationalStatus))
		if err != nil {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Cannot translate Status attribute from ServiceNow into a valid OSS OperationalStatus", "%s", sngi.OperationalStatus).TagDataMismatch().TagCRN()
			snOperationalStatus = ""
		}
	}

	if rmcOSSOperationalStatus != "" {
		switch rmcOSSOperationalStatus {
		case ossrecord.GA, ossrecord.BETA, ossrecord.EXPERIMENTAL, ossrecord.THIRDPARTY, ossrecord.COMMUNITY, ossrecord.DEPRECATED:
			// Mismatch results in a warning, because all other sources are capable of reflectiing the accurate status
			result = si.MergeValues(item, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, forceRetired, RMCOSS{rmcOSSOperationalStatus}, Catalog{catOperationalStatus}, ServiceNow{sngi.OperationalStatus}, ScorecardV1{scOperationalStatus}, PriorOSS{prior.OperationalStatus}).(ossrecord.OperationalStatus)
		default:
			// Exclude Catalog from the warning because it cannot reflect this status
			if catOperationalStatus != "" {
				si.AddValidationIssue(ossvalidation.INFO, "Ignoring OperationalStatus from Catalog -- RMC status takes precedence and cannot be reflected in Catalog", "catalog=%s  rmc=%s", catOperationalStatus, rmcOSSOperationalStatus).TagDataMismatch().TagCRN().TagRunAction(ossrunactions.RMC)
			}
			result = si.MergeValues(item, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, forceRetired, RMCOSS{rmcOSSOperationalStatus} /* Catalog{catOperationalStatus}, */, ServiceNow{sngi.OperationalStatus}, ScorecardV1{scOperationalStatus}, PriorOSS{prior.OperationalStatus}).(ossrecord.OperationalStatus)
		}
		if tag := si.OSSMergeControl.OSSTags.GetTagByGroup(osstags.GroupOperationalStatus); tag != "" {
			si.AddValidationIssue(ossvalidation.SEVERE, "Ignoring OSSTag for OperationalStatus when getting OperationalStatus from RMC OSS tab", `tag=%s  rmc=%s`, tag, rmcOSSOperationalStatus).TagDataMismatch().TagRunAction(ossrunactions.RMC)
		}
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.NotReady) {
		result = si.MergeValues(item, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.INFO}, OverrideTag{N: osstags.NotReady, V: ossrecord.NOTREADY}, Catalog{catOperationalStatus}, RMC{rmcMainOperationalStatus}, ServiceNow{sngi.OperationalStatus}, ScorecardV1{scOperationalStatus}, PriorOSS{prior.OperationalStatus}).(ossrecord.OperationalStatus)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.SelectAvailability) {
		result = si.MergeValues(item, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.INFO}, OverrideTag{N: osstags.SelectAvailability, V: ossrecord.SELECTAVAILABILITY}, Catalog{catOperationalStatus}, RMC{rmcMainOperationalStatus}, ServiceNow{sngi.OperationalStatus}, ScorecardV1{scOperationalStatus}, PriorOSS{prior.OperationalStatus}).(ossrecord.OperationalStatus)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.Deprecated) {
		result = si.MergeValues(item, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, OverrideTag{N: osstags.Deprecated, V: ossrecord.DEPRECATED}, Catalog{catOperationalStatus}, RMC{rmcMainOperationalStatus}, ServiceNow{sngi.OperationalStatus}, ScorecardV1{scOperationalStatus}, PriorOSS{prior.OperationalStatus}).(ossrecord.OperationalStatus)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.Retired) {
		result = si.MergeValues(item, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, OverrideTag{N: osstags.Retired, V: ossrecord.RETIRED}, Catalog{catOperationalStatus}, RMC{rmcMainOperationalStatus}, ServiceNow{sngi.OperationalStatus}, ScorecardV1{scOperationalStatus}, PriorOSS{prior.OperationalStatus}).(ossrecord.OperationalStatus)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.Internal) {
		result = si.MergeValues(item, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, OverrideTag{N: osstags.Internal, V: ossrecord.INTERNAL}, Catalog{catOperationalStatus}, RMC{rmcMainOperationalStatus}, ServiceNow{sngi.OperationalStatus}, ScorecardV1{scOperationalStatus}, PriorOSS{prior.OperationalStatus}).(ossrecord.OperationalStatus)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.Invalid) {
		si.AddValidationIssue(ossvalidation.SEVERE, `Making OperationalStatus <unknown> based on "invalid" OSS Merge Control tag`, "Catalog=%v  ServiceNow=%v  ScorecardV1=%v", catOperationalStatus, sngi.OperationalStatus, scOperationalStatus).TagCRN()
		result = si.MergeValues(item, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, OverrideTag{N: osstags.Invalid, V: ossrecord.OperationalStatusUnknown}, Catalog{catOperationalStatus}, RMC{rmcMainOperationalStatus}, ServiceNow{sngi.OperationalStatus}, ScorecardV1{scOperationalStatus}, PriorOSS{prior.OperationalStatus}).(ossrecord.OperationalStatus)
	} else if scOperationalStatus == ossrecord.INTERNAL && si.HasSourceServiceNow() && snOperationalStatus != ossrecord.INTERNAL {
		// ServiceNow may not have a INTERNAL status, but the one from ScorecardV1 takes precedence
		si.AddValidationIssue(ossvalidation.WARNING, "Overriding Operational Status from ServiceNow with INTERNAL value from ScorecardV1", "ServiceNow=%v", snOperationalStatus).TagDataMismatch().TagCRN()
		result = si.MergeValues(item, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, Catalog{catOperationalStatus}, RMC{rmcMainOperationalStatus}, ScorecardV1{scOperationalStatus}, PriorOSS{prior.OperationalStatus}).(ossrecord.OperationalStatus)
	} else {
		result = si.MergeValues(item, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, Catalog{catOperationalStatus}, RMC{rmcMainOperationalStatus}, ServiceNow{sngi.OperationalStatus}, ScorecardV1{scOperationalStatus}, PriorOSS{prior.OperationalStatus}).(ossrecord.OperationalStatus)
	}
	if snOperationalStatus == ossrecord.RETIRED && result != ossrecord.RETIRED {
		si.AddValidationIssue(ossvalidation.SEVERE, "ServiceNow entry is RETIRED but there is an active entry in Catalog or ScorecardV1", "overall OperationalStatus is %v", result).TagCRN()
	}

	if result == "" {
		result = ossrecord.OperationalStatusUnknown
	}

	return result
}

func (si *ServiceInfo) mergeClientFacing() {
	catalogItem := "CatalogClientFacing flag"
	sngi := &si.SourceServiceNow.GeneralInfo
	prior := &si.PriorOSS.GeneralInfo

	if si.OSSService.GeneralInfo.EntryType == "" {
		panic(fmt.Sprintf(`mergeClientFacing(): missing EntryType for "%s"`, si.String()))
	}
	if si.OSSService.GeneralInfo.OperationalStatus == "" {
		panic(fmt.Sprintf(`mergeClientFacing(): missing OperationalStatus for "%s"`, si.String()))
	}

	// Sanity checks - do not allow client-facing if INVALID in RMC or RETIRED in ServiceNow
	var forceDisable parameter
	if prior.ClientFacing || sngi.ClientFacing {
		if si.GeneralInfo.OSSOnboardingPhase == ossrecord.INVALID {
			si.AddValidationIssue(ossvalidation.WARNING, "Forcing disable ClientFacing for entry with OSSOnboardingPhase=INVALID", "prior RMC ClientFacing=%v   ServiceNow ClientFacing=%v", prior.ClientFacing, sngi.ClientFacing).TagCRN().TagRunAction(ossrunactions.RMC)
			forceDisable = Custom{N: "Disable ClientFacing for OSSOnboardingPhase=INVALID", V: false}
		} else if sngi.OperationalStatus == ossrecord.RETIRED {
			si.AddValidationIssue(ossvalidation.WARNING, "Forcing disable ClientFacing for entry that is RETIRED in ServiceNow", "prior RMC ClientFacing=%v   ServiceNow ClientFacing=%v", prior.ClientFacing, sngi.ClientFacing).TagCRN().TagRunAction(ossrunactions.RMC)
			forceDisable = Custom{N: "Disable ClientFacing for OSSOnboardingPhase=INVALID", V: false}
		}
	}

	//
	// No need to worry about the Catalog for the main Client-facing flag -- it is controlled only by ServiceNow
	//
	si.OSSService.GeneralInfo.ClientFacing = si.MergeValues("Client-facing flag", forceDisable, RMCOSS{prior.ClientFacing}, ServiceNow{sngi.ClientFacing}, PriorOSS{prior.ClientFacing}).(bool)

	//
	// Now try to determine the CatalogClientFacing flags (independent of ServiceNow)
	//

	// Check for Override tag
	// Note that we cannot just add both OverrideTag{osstags.ClientFacing, true} and OverrideTag{osstags.NotClientFacing, false}
	// because that would cause a duplicate MergeValue option of type OverrideTag, which is not allowed
	var overrideTagName osstags.Tag
	var overrideTagValue bool
	if si.OSSMergeControl.OSSTags.Contains(osstags.ClientFacing) {
		overrideTagName = osstags.ClientFacing
		overrideTagValue = true
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.NotClientFacing) {
		overrideTagName = osstags.NotClientFacing
		overrideTagValue = false
	} else {
		overrideTagName = ""
	}

	if si.HasSourceMainCatalog() {
		switch si.OSSService.GeneralInfo.EntryType {
		case ossrecord.SERVICE, ossrecord.RUNTIME, ossrecord.IAAS, ossrecord.TEMPLATE, ossrecord.COMPOSITE:
			si.OSSService.CatalogInfo.CatalogClientFacing = si.MergeValues(catalogItem, OverrideTag{overrideTagName, overrideTagValue}, Catalog{si.SourceMainCatalog.IsPublicVisible()}, PriorOSS{prior.ClientFacing}).(bool)
		case ossrecord.PLATFORMCOMPONENT:
			// TODO: Should have a way to mark platform_component entries as Client-facing in Catalog
			si.AddValidationIssue(ossvalidation.INFO, "Cannot determine CatalogClientFacing flag for PLATFORMCOMPONENT entry in Catalog - assuming non-Client-facing unless Override", "EntryType=%s  ServiceNow-Client-facing=%v  Override=%s", si.OSSService.GeneralInfo.EntryType, si.OSSService.GeneralInfo.ClientFacing, overrideTagName).TagDataMismatch()
			si.OSSService.CatalogInfo.CatalogClientFacing = si.MergeValues(catalogItem, OverrideTag{overrideTagName, overrideTagValue}, Custom{N: "Non-public Catalog entry for PLATFORMCOMPONENT", V: false}, PriorOSS{prior.ClientFacing}).(bool)
		case ossrecord.VMWARE:
			if si.GetSourceMainCatalog().IsPublicVisibleHiddenOK() {
				si.AddValidationIssue(ossvalidation.INFO, "Cannot determine CatalogClientFacing/visibility for VMWARE entry in Catalog - assuming a hidden Catalog is actually visible in custom UI", "EntryType=%s  ServiceNow-Client-facing=%v  Override=%s", si.OSSService.GeneralInfo.EntryType, si.OSSService.GeneralInfo.ClientFacing, overrideTagName).TagDataMismatch()
				si.OSSService.CatalogInfo.CatalogClientFacing = si.MergeValues(catalogItem, OverrideTag{overrideTagName, overrideTagValue}, Custom{N: "Hidden Catalog entry for VMWARE", V: true}, PriorOSS{prior.ClientFacing}).(bool)
			} else {
				/*
					var sn string
					if si.HasSourceServiceNow() {
						sn = fmt.Sprintf("%v", sngi.ClientFacing)
					} else {
						sn = "<no-ServiceNow-entry>"
					}
				*/
				si.AddValidationIssue(ossvalidation.INFO, "Cannot determine CatalogClientFacing flag for VMWARE entry in Catalog - assuming non-Client-facing unless Override", "EntryType=%s  ServiceNow-Client-facing=%v  Override=%s", si.OSSService.GeneralInfo.EntryType /* sn */, si.OSSService.GeneralInfo.ClientFacing, overrideTagName).TagDataMismatch()
				si.OSSService.CatalogInfo.CatalogClientFacing = si.MergeValues(catalogItem, OverrideTag{overrideTagName, overrideTagValue}, Custom{N: "Non-public Catalog entry for VMWARE", V: false}, PriorOSS{prior.ClientFacing}).(bool)
			}
		default:
			si.AddValidationIssue(ossvalidation.WARNING, "Cannot determine CatalogClientFacing flag from Catalog - unexpected EntryType; assuming non-Client-facing unless Override", "EntryType=%s  ServiceNow-Client-facing=%v  Override=%s", si.OSSService.GeneralInfo.EntryType, si.OSSService.GeneralInfo.ClientFacing, overrideTagName).TagDataMismatch()
			si.OSSService.CatalogInfo.CatalogClientFacing = si.MergeValues(catalogItem, OverrideTag{overrideTagName, overrideTagValue}, Custom{N: "Non-public Catalog entry for unknown EntryType", V: false}, PriorOSS{prior.ClientFacing}).(bool)
		}
	} else {
		// We assume that if we have no Catalog entry, this service is not visible in the Client-facing Catalog either
		si.OSSService.CatalogInfo.CatalogClientFacing = si.MergeValues(catalogItem, OverrideTag{overrideTagName, overrideTagValue}, Custom{N: "Catalog(non-existent)", V: false}, PriorOSS{prior.ClientFacing}).(bool)
	}

	//
	// Warn about inconsistencies between Catalog and ServiceNow-based Client-facing flag
	//
	switch si.OSSService.GeneralInfo.EntryType {
	case ossrecord.SERVICE, ossrecord.RUNTIME, ossrecord.IAAS, ossrecord.TEMPLATE, ossrecord.VMWARE:
		switch si.OSSService.GeneralInfo.OperationalStatus {
		case ossrecord.SELECTAVAILABILITY, ossrecord.RETIRED, ossrecord.INTERNAL, ossrecord.NOTREADY, ossrecord.OperationalStatusUnknown:
			if si.HasSourceMainCatalog() &&
				(si.SourceMainCatalog.IsPublicVisible() ||
					(si.OSSService.GeneralInfo.EntryType == ossrecord.VMWARE && si.GetSourceMainCatalog().IsPublicVisibleHiddenOK())) {
				if si.OSSService.GeneralInfo.ClientFacing == true {
					if overrideTagName == "" {
						si.AddValidationIssue(ossvalidation.SEVERE, "Catalog entry is public and Client-facing is true for an OperationalStatus that would not be expected to be client-facing", "OperationalStatus=%s  EntryType=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType).TagConsistency().TagCRN()
					} else {
						si.AddValidationIssue(ossvalidation.MINOR, "Catalog entry is public and Client-facing is true for an OperationalStatus that would not be expected to be client-facing (overridden by OSSTag)", "OperationalStatus=%s  EntryType=%s  Override=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType, overrideTagName).TagControlOverride().TagCRN()
					}
				} else {
					if overrideTagName == "" {
						si.AddValidationIssue(ossvalidation.SEVERE, "Catalog entry is public (but Client-facing is false) for an OperationalStatus that would not be expected to be client-facing", "OperationalStatus=%s  EntryType=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType).TagConsistency().TagCRN()
					} else {
						si.AddValidationIssue(ossvalidation.SEVERE, "Catalog entry is public (but Client-facing is false) for an OperationalStatus that would not be expected to be client-facing (overridden by OSSTag)", "OperationalStatus=%s  EntryType=%s  Override=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType, overrideTagName).TagControlOverride().TagCRN()
					}
				}
			} else {
				if si.OSSService.GeneralInfo.ClientFacing == true {
					if overrideTagName == "" {
						si.AddValidationIssue(ossvalidation.SEVERE, "Client-facing is true (but no public Catalog entry) for an OperationalStatus that would not be expected to be client-facing", "OperationalStatus=%s  EntryType=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType).TagConsistency().TagCRN()
					} else {
						si.AddValidationIssue(ossvalidation.MINOR, "Client-facing is true (but no public Catalog entry) for an OperationalStatus that would not be expected to be client-facing (overridden by OSSTag)", "OperationalStatus=%s  EntryType=%s  Override=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType, overrideTagName).TagControlOverride().TagCRN()
					}
				}
			}
		default:
			if si.HasSourceMainCatalog() &&
				(si.SourceMainCatalog.IsPublicVisible() ||
					(si.OSSService.GeneralInfo.EntryType == ossrecord.VMWARE && si.GetSourceMainCatalog().IsPublicVisibleHiddenOK())) {
				if si.OSSService.GeneralInfo.ClientFacing == false {
					if overrideTagName == "" {
						si.AddValidationIssue(ossvalidation.WARNING, "Client-facing is false (but Catalog entry is public) for an OperationalStatus that would normally be expected to be client-facing", "OperationalStatus=%s  EntryType=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType).TagConsistency().TagCRN()
					} else {
						si.AddValidationIssue(ossvalidation.MINOR, "Client-facing is false (but Catalog entry is public) for an OperationalStatus that would normally be expected to be client-facing (overridden by OSSTag)", "OperationalStatus=%s  EntryType=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType).TagConsistency().TagCRN()
					}
				}
			} else {
				if si.OSSService.GeneralInfo.ClientFacing == false {
					if overrideTagName == "" {
						si.AddValidationIssue(ossvalidation.SEVERE, "No public Catalog entry and Client-facing is false for an OperationalStatus that would normally be expected to be client-facing", "OperationalStatus=%s  EntryType=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType).TagConsistency().TagCRN()
					} else {
						si.AddValidationIssue(ossvalidation.MINOR, "No public Catalog entry and Client-facing is false for an OperationalStatus that would normally be expected to be client-facing (overridden by OSSTag)", "OperationalStatus=%s  EntryType=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType).TagConsistency().TagCRN()
					}
				} else {
					if overrideTagName == "" {
						si.AddValidationIssue(ossvalidation.SEVERE, "No public Catalog entry (but Client-facing is true) for an OperationalStatus that would normally be expected to be client-facing", "OperationalStatus=%s  EntryType=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType).TagConsistency().TagCRN()
					} else {
						si.AddValidationIssue(ossvalidation.SEVERE, "No public Catalog entry (but Client-facing is true) for an OperationalStatus that would normally be expected to be client-facing (overridden by OSSTag)", "OperationalStatus=%s  EntryType=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType).TagConsistency().TagCRN()
					}
				}
			}
		}
	default:
		if si.OSSService.GeneralInfo.ClientFacing == true {
			if overrideTagName == "" {
				si.AddValidationIssue(ossvalidation.SEVERE, "Client-facing is true for an EntryType that would not be expected to be client-facing", "OperationalStatus=%s  EntryType=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType).TagConsistency().TagCRN()
			} else {
				si.AddValidationIssue(ossvalidation.MINOR, "Client-facing is true for an EntryType that would not be expected to be client-facing (overridden by OSSTag)", "OperationalStatus=%s  EntryType=%s  Override=%s", si.OSSService.GeneralInfo.OperationalStatus, si.OSSService.GeneralInfo.EntryType, overrideTagName).TagControlOverride().TagCRN()
			}
		}
	}

}

func (si *ServiceInfo) mergeEntryType() ossrecord.EntryType {
	itemMain := "EntryType"
	itemOverride := "EntryType (special override)"
	var result ossrecord.EntryType
	prior := &si.PriorOSS.GeneralInfo

	var rmcEntryType ossrecord.EntryType
	if si.HasSourceRMC() {
		t, err := si.GetSourceRMC().GetEntryType()
		if err != nil {
			si.AddValidationIssue(ossvalidation.SEVERE, "Invalid RMC entry type", err.Error()).TagDataMismatch().TagRunAction(ossrunactions.RMC)
			t = ossrecord.EntryTypeUnknown
		}
		if si.HasPriorOSS() && prior.OSSOnboardingPhase != "" {
			switch t {
			case ossrecord.EntryType(ossrecord.GAAS):
				rmcEntryType = prior.EntryType
			case ossrecord.SERVICE, ossrecord.PLATFORMCOMPONENT, ossrecord.COMPOSITE:
				if t != prior.EntryType {
					si.AddValidationIssue(ossvalidation.WARNING, "Entry type from main RMC entry does not match the type in the OSS tab of that RMC entry (OSS tab prevails)", `main_type=%q  oss_type=%q`, t, prior.EntryType).TagDataMismatch().TagRunAction(ossrunactions.RMC)
				}
				rmcEntryType = prior.EntryType
			case ossrecord.EntryTypeUnknown:
				// We've already logged a validation issue for bad entry type
				rmcEntryType = prior.EntryType
			default:
				panic(fmt.Sprintf(`Unexpected RMC entry type %q in entry %s`, t, si.String()))
			}
		} else {
			rmcEntryType = t
		}
	}

	var scEntryType ossrecord.EntryType
	var scEntryTypeMerge parameter // nil value by default
	if si.HasSourceScorecardV1Detail() {
		scOperationalStatus, ok := si.SourceScorecardV1Detail.GetOperationalStatus()
		if ok && (scOperationalStatus == ossrecord.INTERNAL) {
			// XXX We assume that INTERNAL ("NA - International Tool" in ScorecardV1) means this is a PLATFORMCOMPONENT
			scEntryType = ossrecord.PLATFORMCOMPONENT
			scEntryTypeMerge = ScorecardV1{scEntryType}
		} else {
			scEntryType = ossrecord.EntryTypeUnknown
		}
	} else if !ossrunactions.ScorecardV1.IsEnabled() && si.HasPriorOSS() {
		// We need a dummy ScorecardV1 merge parameter to trigger MergeValues() into copying from PriorOSS, even though it's actually not going to pick the value from ScorecardV1 itself
		//scEntryType = si.GetPriorOSS().GeneralInfo.EntryType
		scEntryType = ""
		scEntryTypeMerge = ScorecardV1{scEntryType}
	}

	var segmentType ossrecord.SegmentType
	var segmentName string
	if si.OSSService.Ownership.SegmentID != "" {
		if seg, found := LookupSegment(si.OSSService.Ownership.SegmentID, false); found {
			segmentType = seg.OSSSegment.SegmentType
			segmentName = seg.OSSSegment.String()
		} else {
			// Just move on if the segment could not be found -- a warning will already have been generated in mergeOwnership()
		}
	}

	var catEntryType ossrecord.EntryType
	var isCompositeName bool
	if si.HasSourceMainCatalog() {
		var ok bool
		catEntryType, ok = si.SourceMainCatalog.GetEntryType()
		if !ok {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Cannot translate Kind attribute from Catalog into a valid OSS EntryType", "%s", si.SourceMainCatalog.Kind).TagDataMismatch().TagCRN()
			catEntryType = ""
		}
		categoryTags := catalog.ScanCategoryTags(si.GetSourceMainCatalog())
		si.OSSService.CatalogInfo.CategoryTags = si.MergeValues("Catalog.CategoryTags", TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMissing{ossvalidation.MINOR}, Catalog{categoryTags}, PriorOSS{si.PriorOSS.CatalogInfo.CategoryTags}).(string)
		if catEntryType == ossrecord.SERVICE {
			if strings.Contains(strings.ToLower(categoryTags), "vmware") {
				si.AddValidationIssue(ossvalidation.INFO, fmt.Sprintf("Setting EntryType to %s based on Catalog category tags", ossrecord.VMWARE), "tags=%s", si.GetSourceMainCatalog().Tags).TagCRN()
				catEntryType = ossrecord.VMWARE
			}
		}
		if base, _ := ParseCompositeName(si.GetSourceMainCatalog().Name); base != "" {
			isCompositeName = true // TODO: Need a more rigorous check for a Composite child, e.g. does the parent actually exist
		}
	}

	var snEntryType ossrecord.EntryType
	if si.HasSourceServiceNow() {
		snEntryType = si.GetSourceServiceNow().GeneralInfo.EntryType
	}

	if si.HasSourceRMC() && si.HasPriorOSS() && prior.OSSOnboardingPhase != "" {
		result = si.MergeValues(itemMain, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, RMC{rmcEntryType}, Catalog{catEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
		if tag := si.OSSMergeControl.OSSTags.GetTagByGroup(osstags.GroupEntryType); tag != "" {
			si.AddValidationIssue(ossvalidation.SEVERE, "Ignoring OSSTag for EntryType when getting EntryType from RMC OSS tab", `tag=%s  rmc=%s`, tag, rmcEntryType).TagDataMismatch().TagRunAction(ossrunactions.RMC)
		}
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.TypeContent) {
		result = si.MergeValues(itemOverride, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.DEFERRED}, OverrideTag{N: osstags.TypeContent, V: ossrecord.CONTENT}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.TypeConsulting) {
		result = si.MergeValues(itemOverride, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.DEFERRED}, OverrideTag{N: osstags.TypeConsulting, V: ossrecord.CONSULTING}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.TypeInternal) {
		result = si.MergeValues(itemOverride, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.DEFERRED}, OverrideTag{N: osstags.TypeInternal, V: ossrecord.INTERNALSERVICE}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.TypeOtherOSS) {
		result = si.MergeValues(itemOverride, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.DEFERRED}, OverrideTag{N: osstags.TypeOtherOSS, V: ossrecord.OTHEROSS}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.TypeIAMOnly) {
		result = si.MergeValues(itemOverride, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.DEFERRED}, OverrideTag{N: osstags.TypeIAMOnly, V: ossrecord.IAMONLY}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.TypeGaaS) {
		result = si.MergeValues(itemOverride, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.DEFERRED}, OverrideTag{N: osstags.TypeGaaS, V: ossrecord.GAAS}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if segmentType == ossrecord.SegmentTypeGaaS {
		si.AddValidationIssue(ossvalidation.INFO, "Overriding Entry Type from all sources with GAAS, based on owning Segment type", "Catalog=%v  ServiceNow=%v  ScorecardV1=%v  RMC=%v", catEntryType, snEntryType, scEntryType, rmcEntryType).TagNewProperty().TagCRN()
		// TODO: Change severity for non-GAAS type in ServiceNow from DEFERRED to SEVERE
		result = si.MergeValues(itemOverride, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.DEFERRED}, OverrideTag{N: osstags.TypeGaaS, V: ossrecord.GAAS}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if !ossrunactions.Tribes.IsEnabled() && si.HasPriorOSS() && si.GetPriorOSS().GeneralInfo.EntryType == ossrecord.GAAS {
		// XXX If we already had GAAS type from a previous run and cannot reconfirm based on segment data in this run, go with it
		// TODO: Change severity for non-GAAS type in ServiceNow from DEFERRED to SEVERE
		result = si.MergeValues(itemOverride, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.DEFERRED}, OverrideTag{N: osstags.TypeGaaS, V: ossrecord.GAAS}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.TypeComponent) {
		result = si.MergeValues(itemMain, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, OverrideTag{N: osstags.TypeComponent, V: ossrecord.PLATFORMCOMPONENT}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.TypeSubcomponent) {
		si.AddValidationIssue(ossvalidation.INFO, "Overriding Entry Type from all sources with SUBCOMPONENT from OSSMergeControl tags", "Catalog=%v  ServiceNow=%v  ScorecardV1=%v  RMC=%v", catEntryType, snEntryType, scEntryType, rmcEntryType).TagNewProperty().TagCRN()
		result = ossrecord.SUBCOMPONENT
		// XXX need to account for RMC
		if catEntryType != "" && catEntryType != ossrecord.PLATFORMCOMPONENT {
			si.AddValidationIssue(ossvalidation.SEVERE, "Entry Type is SUBCOMPONENT (based on OSSMergeControl tags) but Catalog entry kind is not platform_component", "Catalog.kind=%s", catEntryType).TagDataMismatch().TagCRN()
		}
		if rmcEntryType != "" && rmcEntryType != ossrecord.PLATFORMCOMPONENT {
			si.AddValidationIssue(ossvalidation.SEVERE, "Entry Type is SUBCOMPONENT (based on OSSMergeControl tags) but RMC entry type is not platform_component", "RMC.type=%s", rmcEntryType).TagDataMismatch().TagCRN().TagRunAction(ossrunactions.RMC)
		}
	} else if si.OSSMergeControl.OSSTags.Contains(osstags.TypeSupercomponent) {
		si.AddValidationIssue(ossvalidation.INFO, "Overriding Entry Type from all sources with SUPERCOMPONENT from OSSMergeControl tags", "Catalog=%v  ServiceNow=%v  ScorecardV1=%v", catEntryType, snEntryType, scEntryType).TagNewProperty().TagCRN()
		result = ossrecord.SUPERCOMPONENT
		if catEntryType != "" && catEntryType != ossrecord.PLATFORMCOMPONENT {
			si.AddValidationIssue(ossvalidation.SEVERE, "Entry Type is SUPERCOMPONENT (based on OSSMergeControl tags) but Catalog entry kind is not platform_component", "Catalog.kind=%s", catEntryType).TagDataMismatch().TagCRN()
		}
		if rmcEntryType != "" && rmcEntryType != ossrecord.PLATFORMCOMPONENT {
			si.AddValidationIssue(ossvalidation.SEVERE, "Entry Type is SUPERCOMPONENT (based on OSSMergeControl tags) but RMC entry type is not platform_component", "RMC.type=%s", rmcEntryType).TagDataMismatch().TagCRN().TagRunAction(ossrunactions.RMC)
		}
	} else if catEntryType == ossrecord.VMWARE || si.OSSMergeControl.OSSTags.Contains(osstags.TypeVMware) {
		if scEntryType == ossrecord.SERVICE {
			// TODO: Need a native VMWARE type in Scorecard (obsolete)
			si.AddValidationIssue(ossvalidation.INFO, "Simulating EntryType=VMWARE in ScorecardV1 instead of SERVICE, because ScorecardV1 does not have a native VMWARE type", "").TagCRN()
			scEntryType = ossrecord.VMWARE
			if scEntryTypeMerge != nil {
				scEntryTypeMerge = ScorecardV1{scEntryType}
			}
		}
		if rmcEntryType == ossrecord.SERVICE {
			si.AddValidationIssue(ossvalidation.INFO, "Simulating EntryType=VMWARE in RMC instead of SERVICE, because RMC does not have a native VMWARE type", "").TagCRN().TagRunAction(ossrunactions.RMC)
			rmcEntryType = ossrecord.VMWARE
		}
		result = si.MergeValues(itemOverride, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.DEFERRED}, OverrideTag{N: osstags.TypeVMware, V: ossrecord.VMWARE}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if catEntryType == ossrecord.COMPOSITE {
		si.AddValidationIssue(ossvalidation.INFO, "Overriding Entry Type from all sources with COMPOSITE from Catalog", "Catalog=%v  ServiceNow=%v  ScorecardV1=%v  RMC=%v", catEntryType, snEntryType, scEntryType, rmcEntryType).TagNewProperty().TagCRN().TagCatalogComposite()
		result = si.MergeValues(itemMain, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, Catalog{catEntryType}, RMC{rmcEntryType}, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if catEntryType == ossrecord.IAAS && snEntryType == ossrecord.SERVICE {
		si.AddValidationIssue(ossvalidation.INFO, "Overriding Entry Type in ServiceNow to IAAS, to match type from Catalog (ServiceNow does not have an explicit type for IAAS)", "Catalog=%v  ServiceNow=%v  ScorecardV1=%v  RMC=%v", catEntryType, snEntryType, scEntryType, rmcEntryType).TagNewProperty().TagCRN()
		result = si.MergeValues(itemMain, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{ossrecord.IAAS}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if strings.HasPrefix(string(si.OSSService.ReferenceResourceName), `content-ibm-cp`) { // FIXME: remove temporary hard-coded prefix for CONTENT type
		si.AddValidationIssue(ossvalidation.INFO, "Overriding Entry Type from all sources with CONTENT -- temporarily hard-coded based on service-name", "Catalog=%v  ServiceNow=%v  ScorecardV1=%v  RMC=%v", catEntryType, snEntryType, scEntryType, rmcEntryType).TagNewProperty().TagCRN()
		si.OSSMergeControl.OSSTags.AddTag(osstags.TypeContent)
		result = si.MergeValues(itemOverride, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, SeverityIfMismatch{ossvalidation.DEFERRED}, OverrideTag{N: osstags.TypeContent, V: ossrecord.CONTENT}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	} else if si.HasSourceIAM() && catEntryType == "" && snEntryType == "" && scEntryType == "" {
		// TODO: need better way to detect IAM_ONLY entries in RMC
		si.AddValidationIssue(ossvalidation.INFO, "Entry found only in IAM (no sources in Catalog, ServiceNow, Scorecard) -- setting type to IAM_ONLY", `IAM_name="%s"`, si.SourceIAM.Name).TagCRN().TagIAM()
		result = ossrecord.IAMONLY
		if rmcEntryType != "" && rmcEntryType != ossrecord.PLATFORMCOMPONENT && rmcEntryType != ossrecord.SERVICE && rmcEntryType != ossrecord.COMPOSITE {
			si.AddValidationIssue(ossvalidation.SEVERE, "Entry Type is IAM_ONLY (based on IAM and no Catalog entry) but RMC entry type is unexpected", "RMC.type=%s", rmcEntryType).TagDataMismatch().TagCRN().TagRunAction(ossrunactions.RMC)
		}
	} else if si.HasSourceIAM() && catEntryType == ossrecord.PLATFORMCOMPONENT && snEntryType == "" && scEntryType == "" && isCompositeName {
		// TODO: need better way to detect IAM_ONLY entries in RMC
		si.AddValidationIssue(ossvalidation.INFO, "Entry found only in IAM and as a PLATFORM_COMPONENT child of a COMPOSITE in Catalog -- assuming this is a IAM_ONLY entry", `IAM_name="%s"`, si.SourceIAM.Name).TagCRN().TagIAM()
		result = ossrecord.IAMONLY
		if rmcEntryType != "" && rmcEntryType != ossrecord.PLATFORMCOMPONENT {
			si.AddValidationIssue(ossvalidation.SEVERE, "Entry Type is IAM_ONLY (based on IAM and Catalog=platform_component) but RMC entry type is not platform_component", "RMC.type=%s", rmcEntryType).TagDataMismatch().TagCRN().TagRunAction(ossrunactions.RMC)
		}
	} else {
		result = si.MergeValues(itemMain, TagIfMismatch{ossvalidation.TagCRN}, TagIfMissing{ossvalidation.TagCRN}, Catalog{catEntryType}, RMC{rmcEntryType}, ServiceNow{snEntryType}, scEntryTypeMerge, PriorOSS{prior.EntryType}).(ossrecord.EntryType)
	}

	switch segmentType {
	case "":
		// ignore
	case ossrecord.SegmentTypeGaaS:
		if result != ossrecord.GAAS {
			si.AddValidationIssue(ossvalidation.SEVERE, "Segment of type GaaS contains an entry of type other than GaaS", "%s   SegmentType=%s   EntryType=%s", segmentName, segmentType, result).TagCRN()
		}
	default:
		if result == ossrecord.GAAS {
			si.AddValidationIssue(ossvalidation.SEVERE, "Segment of type other than GaaS contains an entry of type GaaS", "%s   SegmentType=%s   EntryType=%s", segmentName, segmentType, result).TagCRN()
		}
	}

	if strings.HasPrefix(string(si.OSSService.ReferenceResourceName), `content-ibm-cp`) && result != ossrecord.CONTENT { // FIXME: remove temporary hard-coded prefix for CONTENT type
		si.AddValidationIssue(ossvalidation.SEVERE, "Entry has a service-name that suggests CONTENT (IBM CloudPak) but type is not CONTENT", "EntryType=%s", result).TagDataMismatch().TagCRN()
	}

	if result == "" {
		// TODO: consider some last-ditch assumptions for EntryType
		result = ossrecord.EntryTypeUnknown
	}

	return result
}

func (si *ServiceInfo) mergeOwnership() {
	// XXX This code cannot depend on the EntryType, because we may use the Segment info constructed here to determine the entry type (for GaaS entries)
	oss := &si.OSSService.Ownership
	//cat := &si.SourceMainCatalog
	sn := &si.SourceServiceNow.Ownership
	sc := &si.SourceScorecardV1Detail
	prior := &si.PriorOSS.Ownership

	var rmcOfferingManager parameter
	if si.HasSourceRMC() && si.GetSourceRMC().SOPGTM != nil {
		w3id := si.GetSourceRMC().SOPGTM.OwnerEmail
		name := si.GetSourceRMC().SOPGTM.OwnerName
		if (w3id == "N/A" || w3id == "") && (name == "N/A" || name == "") {
			// No valid OfferingManager info in RMC
			// See https://github.ibm.com/cloud-sre/osscatalog/issues/404
		} else {
			rmcOfferingManager = RMC{ossrecord.Person{W3ID: si.GetSourceRMC().SOPGTM.OwnerEmail, Name: si.GetSourceRMC().SOPGTM.OwnerName}}
		}
	}

	if !si.HasSourceServiceNow() {
		// Don't worry about Key Contacts information for entries that have neither ServiceNow nor RMCOSS entries (e.g. thirdparty-services, etc.)
		// (Note that we must omit RMCOSS here, otherwise the merge may pick-up the PriorOSS value even if the OSS entry is not onboarded in RMC)
		// TODO: Should use Severity=IGNORE? (but keep as WARNING for now, for backward compatibility)
		oss.OfferingManager = si.MergeValues("OfferingManager", ServiceNow{sn.OfferingManager}, PriorOSS{prior.OfferingManager}).(ossrecord.Person)
	} else {
		oss.OfferingManager = si.MergeValues("OfferingManager", RMCOSS{prior.OfferingManager}, ServiceNow{sn.OfferingManager}, rmcOfferingManager, PriorOSS{prior.OfferingManager}).(ossrecord.Person)
	}
	oss.DevelopmentManager = si.MergeValues("DevelopmentManager",
		ScorecardV1{ossrecord.Person{Name: sc.MgmtContact, W3ID: sc.MgmtContactEmail}}, PriorOSS{prior.DevelopmentManager}).(ossrecord.Person)
	oss.TechnicalContactDEPRECATED = si.MergeValues("TechnicalContact",
		ScorecardV1{ossrecord.Person{Name: sc.TechContact, W3ID: sc.TechContactEmail}}, PriorOSS{prior.TechnicalContactDEPRECATED}).(ossrecord.Person)
	if ossrunactions.Tribes.IsEnabled() {
		si.OSSValidation.RecordRunAction(ossrunactions.Tribes)
		//oss.SegmentName = si.MergeValues("SegmentName", ScorecardV1{sc.BusinessUnit}, ServiceNow{sn.SegmentName}, PriorOSS{prior.SegmentName}).(string)
		oss.SegmentName = si.MergeValues("SegmentName", RMCOSS{prior.SegmentName}, ServiceNow{sn.SegmentName}, ScorecardV1{sc.BusinessUnit}, PriorOSS{prior.SegmentName}).(string) // XXX
		if oss.SegmentName == "" {
			oss.SegmentName = "<none>"
		}
		// XXX This duplicates logic in ConstructOSSSegment, but we need it here to log validation issues for each service
		seg, found := LookupSegmentName(oss.SegmentName)
		if found {
			// Update segment name -- in case it was actually a test record
			oss.SegmentName = seg.OSSSegment.DisplayName

			if seg.HasSourceScorecardV1() {
				scSegment := seg.GetSourceScorecardV1()
				oss.SegmentOwner = si.MergeValues("SegmentOwner",
					ScorecardV1{ossrecord.Person{Name: scSegment.MgmtName, W3ID: scSegment.MgmtEmail}}, PriorOSS{prior.SegmentOwner}).(ossrecord.Person)
				if scSegment.TechEmail != "" && strings.ToLower(scSegment.TechEmail) != "none" {
					si.AddValidationIssue(ossvalidation.IGNORE, "Ignoring TechEmail attribute for Segment", `segment="%s"   techemail=%s   techname=%q`, oss.SegmentName, scSegment.TechEmail, scSegment.TechName).TagDataMismatch().TagRunAction(ossrunactions.Tribes)
				}
			} else {
				oss.SegmentOwner = seg.OSSSegment.Owner
			}
			oss.SegmentID = seg.OSSSegment.SegmentID
		} else {
			si.AddValidationIssue(ossvalidation.WARNING, "Cannot find Segment information", `segment="%s"`, oss.SegmentName).TagDataMissing().TagRunAction(ossrunactions.Tribes)
		}
		//oss.TribeName = si.MergeValues("TribeName", ScorecardV1{sc.Tribe}, ServiceNow{sn.TribeName}, PriorOSS{prior.TribeName}).(string)
		oss.TribeName = si.MergeValues("TribeName", RMCOSS{prior.TribeName}, ServiceNow{sn.TribeName}, ScorecardV1{sc.Tribe}, PriorOSS{prior.TribeName}).(string) // XXX
		if oss.TribeName == "" {
			oss.TribeName = "<none>"
		}
		if seg != nil {
			// XXX This duplicates logic in ConstructTribe, but we need it here to log validation issues for each service
			tribe, found := seg.LookupTribeName(oss.TribeName)
			if found {
				// Update tribe name -- in case it was actually a test record
				oss.TribeName = tribe.OSSTribe.DisplayName

				if tribe.HasSourceScorecardV1() {
					scTribe := tribe.GetSourceScorecardV1()
					oss.TribeOwner = si.MergeValues("TribeOwner",
						ScorecardV1{ossrecord.Person{Name: scTribe.OwnerContact, W3ID: scTribe.OwnerEmail}}, ServiceNow{sn.TribeOwner}, PriorOSS{prior.TribeOwner}).(ossrecord.Person)
				} else {
					oss.TribeOwner = tribe.OSSTribe.Owner
				}
				oss.TribeID = tribe.OSSTribe.TribeID
			} else {
				si.AddValidationIssue(ossvalidation.WARNING, "Cannot find Tribe information", `tribe="%s/%s"`, oss.SegmentName, oss.TribeName).TagDataMissing().TagRunAction(ossrunactions.Tribes)
				oss.TribeOwner = si.MergeValues("TribeOwner", ServiceNow{sn.TribeOwner}, PriorOSS{prior.TribeOwner}).(ossrecord.Person)
			}
		}
		if si.HasSourceServiceNow() {
			snTribeID := ossrecord.TribeID(si.GetSourceServiceNow().Ownership.TribeID)
			if snTribeID != "" {
				if !strings.Contains(string(snTribeID), "-") {
					originalSNTribeID := snTribeID
					snTribeID = ossrecord.TribeID(fmt.Sprintf("%s-%s", oss.SegmentID, snTribeID))
					si.AddValidationIssue(ossvalidation.DEFERRED, "TribeID obtained from ServiceNow does not appear to start with a SegmentID portion -- prepending SegmentID obtained from other sources", "SN_TribeID=%q   full_TribeID=%q", originalSNTribeID, snTribeID).TagDataMismatch().TagRunAction(ossrunactions.Tribes)
				}
				if oss.TribeID == "" {
					si.AddValidationIssue(ossvalidation.SEVERE, "No Tribe information found from ScorecardV1 but ServiceNow reports a TribeID", "ServiceNow.TribeID=%s", snTribeID).TagDataMismatch().TagRunAction(ossrunactions.Tribes)
				} else if oss.TribeID != snTribeID {
					si.AddValidationIssue(ossvalidation.SEVERE, "TribeID found from ScorecardV1 does not match the TribeID reported by ServiceNow (ScorecardV1 prevails)", "ScorecardV1.TribedID=%s  ServiceNow.TribeID=%s", oss.TribeID, snTribeID).TagDataMismatch().TagRunAction(ossrunactions.Tribes)
				}
			} else if oss.TribeID != "" {
				si.AddValidationIssue(ossvalidation.MINOR, "TribeID found from ScorecardV1 but not in ServiceNow", "ScorecardV1.TribedID=%s", oss.TribeID).TagDataMissing().TagRunAction(ossrunactions.Tribes)
			}
		}
	} else {
		si.OSSValidation.CopyRunAction(si.GetPriorOSSValidation(), ossrunactions.Tribes)
		if si.HasPriorOSS() && si.HasPriorOSSValidation() {
			prior := si.GetPriorOSS().Ownership
			oss.SegmentName = prior.SegmentName
			oss.SegmentOwner = prior.SegmentOwner
			oss.SegmentID = prior.SegmentID
			oss.TribeName = prior.TribeName
			oss.TribeOwner = prior.TribeOwner
			oss.TribeID = prior.TribeID
		}
	}
	oss.ServiceOffering = si.MergeValues("ServiceOffering", SeverityIfMissing{ossvalidation.MINOR}, ServiceNow{sn.ServiceOffering}, PriorOSS{prior.ServiceOffering}).(string)
	oss.MainRepository = ossrecord.GHRepo(si.MergeValues("MainRepository", ScorecardV1{sc.CentralRepositoryLocation}, PriorOSS{string(prior.MainRepository)}).(string))
}

func (si *ServiceInfo) mergeSupport() {
	oss := &si.OSSService.Support
	sn := &si.SourceServiceNow.Support
	sc := &si.SourceScorecardV1Detail
	prior := &si.PriorOSS.Support

	if !si.HasSourceServiceNow() {
		// Don't worry about Key Contacts information for entries that have neither ServiceNow nor RMCOSS entries (e.g. thirdparty-services, etc.)
		// (Note that we must omit RMCOSS here, otherwise the merge may pick-up the PriorOSS value even if the OSS entry is not onboarded in RMC)
		// TODO: Should use Severity=IGNORE? (but keep as WARNING for now, for backward compatibility)
		oss.Manager = si.MergeValues("Support Manager", ServiceNow{sn.Manager}, PriorOSS{prior.Manager}).(ossrecord.Person)
	} else if si.OSSService.ServiceNowInfo.SupportNotApplicable {
		si.AddValidationIssue(ossvalidation.INFO, "Ignore Support Manager info because SupportNotApplicable=true", "").TagSNEnrollment()
		oss.Manager = si.MergeValues("Support Manager", SeverityIfMissing{ossvalidation.IGNORE}, SeverityIfMismatch{ossvalidation.IGNORE}, RMCOSS{prior.Manager}, ServiceNow{sn.Manager}, PriorOSS{prior.Manager}).(ossrecord.Person)
	} else if si.OSSService.GeneralInfo.EntryType == ossrecord.GAAS {
		si.AddValidationIssue(ossvalidation.INFO, "Ignore Support Manager info for a GaaS offering", "").TagSNEnrollment()
		oss.Manager = si.MergeValues("Support Manager", SeverityIfMissing{ossvalidation.IGNORE}, SeverityIfMismatch{ossvalidation.IGNORE}, RMCOSS{prior.Manager}, ServiceNow{sn.Manager}, PriorOSS{prior.Manager}).(ossrecord.Person)
	} else {
		oss.Manager = si.MergeValues("Support Manager", RMCOSS{prior.Manager}, ServiceNow{sn.Manager}, PriorOSS{prior.Manager}).(ossrecord.Person)
	}
	oss.ClientExperience = si.MergeValues("Support ClientExperience", ServiceNow{sn.ClientExperience}, PriorOSS{prior.ClientExperience}).(ossrecord.ClientExperience)
	oss.SpecialInstructions = si.MergeValues("Support SpecialInstructions", SeverityIfMissing{ossvalidation.INFO}, ServiceNow{sn.SpecialInstructions}, PriorOSS{prior.SpecialInstructions}).(string)
	oss.Tier2EscalationType = si.MergeValues("Support Tier2EscalationType", ServiceNow{sn.Tier2EscalationType}, PriorOSS{prior.Tier2EscalationType}).(ossrecord.Tier2EscalationType)
	// Note: could ignore ScorecardV1 Slack channel ... it is copied from ServiceNow
	oss.Slack = si.MergeValues("Support Slack Channel", ServiceNow{sn.Slack}, ScorecardV1{ossrecord.SlackChannel(sc.SupportPublicSlackChannel)}, PriorOSS{prior.Slack}).(ossrecord.SlackChannel)
	if oss.Tier2EscalationType == ossrecord.GITHUB {
		oss.Tier2Repository = si.MergeValues("Support Tier2Repo", ServiceNow{sn.Tier2Repo}, PriorOSS{prior.Tier2Repository}).(ossrecord.GHRepo)
	} else {
		oss.Tier2Repository = si.MergeValues("Support Tier2Repo", SeverityIfMissing{ossvalidation.IGNORE}, ServiceNow{sn.Tier2Repo}, PriorOSS{prior.Tier2Repository}).(ossrecord.GHRepo)
	}
	if oss.Tier2EscalationType == ossrecord.RTC {
		oss.Tier2RTC = si.MergeValues("Support Tier2RTC", ServiceNow{sn.Tier2RTC}, PriorOSS{prior.Tier2RTC}).(ossrecord.RTCCategory)
	} else {
		oss.Tier2RTC = si.MergeValues("Support Tier2RTC", SeverityIfMissing{ossvalidation.IGNORE}, ServiceNow{sn.Tier2RTC}, PriorOSS{prior.Tier2RTC}).(ossrecord.RTCCategory)
	}
	oss.ThirdPartySupportURL = si.MergeValues("Third-party Support URL", SeverityIfMissing{ossvalidation.IGNORE}, PriorOSS{prior.ThirdPartySupportURL}).(string)
}

func (si *ServiceInfo) mergeOperations() {
	oss := &si.OSSService.Operations
	//cat := &si.SourceMainCatalog
	sn := &si.SourceServiceNow.Operations
	sc := &si.SourceScorecardV1Detail
	prior := &si.PriorOSS.Operations

	if !si.HasSourceServiceNow() {
		// Don't worry about Key Contacts information for entries that have neither ServiceNow nor RMCOSS entries (e.g. thirdparty-services, etc.)
		// (Note that we must omit RMCOSS here, otherwise the merge may pick-up the PriorOSS value even if the OSS entry is not onboarded in RMC)
		// TODO: Should use Severity=IGNORE? (but keep as WARNING for now, for backward compatibility)
		oss.Manager = si.MergeValues("Operations Manager", ServiceNow{sn.Manager}, PriorOSS{prior.Manager}).(ossrecord.Person)
	} else if si.OSSService.ServiceNowInfo.OperationsNotApplicable {
		si.AddValidationIssue(ossvalidation.INFO, "Ignore Operations Manager info because OperationsNotApplicable=true", "").TagSNEnrollment()
		oss.Manager = si.MergeValues("Operations Manager", SeverityIfMissing{ossvalidation.IGNORE}, SeverityIfMismatch{ossvalidation.IGNORE}, RMCOSS{prior.Manager}, ServiceNow{sn.Manager}, PriorOSS{prior.Manager}).(ossrecord.Person)
	} else {
		oss.Manager = si.MergeValues("Operations Manager", RMCOSS{prior.Manager}, ServiceNow{sn.Manager}, PriorOSS{prior.Manager}).(ossrecord.Person)
	}
	oss.SpecialInstructions = si.MergeValues("Operations SpecialInstructions", SeverityIfMissing{ossvalidation.INFO}, ServiceNow{sn.SpecialInstructions}, PriorOSS{prior.SpecialInstructions}).(string)
	oss.Tier2EscalationType = si.MergeValues("Operations Tier2EscalationType", SeverityIfMissing{ossvalidation.INFO}, ServiceNow{sn.Tier2EscalationType}, PriorOSS{prior.Tier2EscalationType}).(ossrecord.Tier2EscalationType)
	// Note: could ignore ScorecardV1 Slack channel ... it is copied from ServiceNow
	oss.Slack = si.MergeValues("Operations Slack Channel", ServiceNow{sn.Slack}, PriorOSS{prior.Slack}).(ossrecord.SlackChannel)
	if oss.Tier2EscalationType == ossrecord.GITHUB {
		oss.Tier2Repository = si.MergeValues("Operations Tier2Repo", ServiceNow{sn.Tier2Repo}, PriorOSS{prior.Tier2Repository}).(ossrecord.GHRepo)
	} else {
		oss.Tier2Repository = si.MergeValues("Operations Tier2Repo", SeverityIfMissing{ossvalidation.IGNORE}, ServiceNow{sn.Tier2Repo}, PriorOSS{prior.Tier2Repository}).(ossrecord.GHRepo)
	}
	if oss.Tier2EscalationType == ossrecord.RTC {
		oss.Tier2RTC = si.MergeValues("Operations Tier2RTC", ServiceNow{sn.Tier2RTC}, PriorOSS{prior.Tier2RTC}).(ossrecord.RTCCategory)
	} else {
		oss.Tier2RTC = si.MergeValues("Operations Tier2RTC", SeverityIfMissing{ossvalidation.IGNORE}, ServiceNow{sn.Tier2RTC}, PriorOSS{prior.Tier2RTC}).(ossrecord.RTCCategory)
	}
	oss.TIPOnboarded = si.MergeValues("TIPOnboarded", ScorecardV1{sc.TIPOnboarded}, PriorOSS{prior.TIPOnboarded}).(bool)
	// ScorecardV1.TOC -> ignore
	// ScorecardV1.ManuallyTOC -> ignore
	// ScorecardV1.TOCBypass (from external API) -> deprecated
	oss.BypassProductionReadiness = si.MergeValues("BypassProductionReadiness", ScorecardV1{ossrecord.BypassPassProductionReadiness(sc.BypassPRC)}, PriorOSS{prior.BypassProductionReadiness}).(ossrecord.BypassPassProductionReadiness)

	var scAVMEnabledValue bool
	var scRunbookEnabledValue bool
	if si.HasSourceScorecardV1Detail() {
		if status, success := sc.GetOperationalStatus(); status != ossrecord.GA || !success {
			scAVMEnabledValue = true
			scRunbookEnabledValue = true
			si.AddValidationIssue(ossvalidation.INFO, `Forcing AVMEnabled and RunbookEnabled values to "true" for non-GA entry in ScorecardV1`, `ScorecardV1.status=%s`, status).TagDataMissing()
		} else {
			scAVMEnabledValue = sc.AVMEnabled
			scRunbookEnabledValue = sc.RunbookEnabled
		}
	}
	oss.AVMEnabled = si.MergeValues("AVMEnabled", ScorecardV1{scAVMEnabledValue}, PriorOSS{prior.AVMEnabled}).(bool)
	oss.RunbookEnabled = si.MergeValues("RunbookEnabled", ScorecardV1{scRunbookEnabledValue}, PriorOSS{prior.RunbookEnabled}).(bool)

	// TODO: oss.TOCAVMFocal -> not in ScorecardV1; get from AVM cheat-sheet?
	// TODO: oss.CIEDistList -> not in ScorecardV1; get from AVM cheat-sheet?
	// ScorecardV1.BaileyID -> deprecated
	// ScorecardV1.BaileyURL -> deprecated
	// ScorecardV1.BaileyProject -> deprecated
	oss.EUAccessUSAMName = si.MergeValues("EUAccessUSAMName", ScorecardV1{sc.EUAccessEmergencyUSAMServiceName}, PriorOSS{prior.EUAccessUSAMName}).(string)
	// set AutomationIDs
	if len(oss.AutomationIDs) == 0 && si.HasPriorOSS() {
		oss.AutomationIDs = si.GetPriorOSS().Operations.AutomationIDs
	}
}

func (si *ServiceInfo) mergeCompliance() {
	oss := &si.OSSService.Compliance
	//cat := &si.SourceMainCatalog
	sc := &si.SourceScorecardV1Detail
	prior := &si.PriorOSS.Compliance

	if si.HasSourceServiceNow() && !si.GetSourceServiceNow().IsRetired() {
		oss.ServiceNowOnboarded = true
		if si.HasSourceScorecardV1Detail() && !sc.IsServiceNowOnboarded {
			si.AddValidationIssue(ossvalidation.MINOR, `Entry exists in ServiceNow but is not marked "ServiceNowOnboarded" in ScorecardV1`, "").TagCRN()
		}
	} else {
		if si.HasSourceScorecardV1Detail() && sc.IsServiceNowOnboarded {
			si.AddValidationIssue(ossvalidation.SEVERE, `Entry is marked "ServiceNowOnboarded" in ScorecardV1 but does not exist in ServiceNow`, "").TagCRN()
		}
		oss.ServiceNowOnboarded = false
	}
	// TODO: oss.ProvisionMonitors = si.MergeValues("ProvisionMonitors", ScorecardV1{sc.PMMonitor}, PriorOSS{prior.ProvisionMonitors}).(ossrecord.AvailabilityMonitoringInfo) -> need to define full data structure def
	// TODO: oss.ConsumptionMonitors = si.MergeValues("ConsumptionMonitors", ScorecardV1{sc.CMMonitor}, PriorOSS{prior.ConsumptionMonitors}).(ossrecord.AvailabilityMonitoringInfo)
	var rmcBCDRFocal ossrecord.Person
	if si.HasSourceRMC() && si.GetSourceRMC().SOPBCDR != nil {
		rmcBCDRFocal = ossrecord.ConstructPerson(si.GetSourceRMC().SOPBCDR.ContactEmail)
	}
	oss.BCDRFocal = si.MergeValues("BCDRFocal", RMCOSS{prior.BCDRFocal}, ScorecardV1{ossrecord.ConstructPerson(sc.BCDRFocal)}, RMC{rmcBCDRFocal}, PriorOSS{prior.BCDRFocal}).(ossrecord.Person)
	var rmcSecurityFocal ossrecord.Person
	if si.HasSourceRMC() && si.GetSourceRMC().SOPSecurity != nil {
		rmcSecurityFocal = ossrecord.ConstructPerson(si.GetSourceRMC().SOPSecurity.FocalPoint)
	}
	oss.SecurityFocal = si.MergeValues("SecurityFocal", RMCOSS{prior.SecurityFocal}, ScorecardV1{ossrecord.ConstructPerson(sc.SOPSecurityFocal)}, RMC{rmcSecurityFocal}, PriorOSS{prior.SecurityFocal}).(ossrecord.Person)
	var rmcArchitectureFocal ossrecord.Person
	if si.HasSourceRMC() && si.GetSourceRMC().SOPGTM != nil {
		rmcArchitectureFocal = ossrecord.ConstructPerson(si.GetSourceRMC().SOPArchitecture.FocalEmail)
	}
	oss.ArchitectureFocal = si.MergeValues("ArchitectureFocal", RMC{rmcArchitectureFocal}, PriorOSS{prior.SecurityFocal}).(ossrecord.Person)
	// ScorecardV1.CentralizedVersionControl -> deprecated
	// ScorecardV1.OnCallRotation -> deprecated
	// ScorecardV1.ReliabilityDesignReview -> deprecated
	// ScorecardV1.AutomatedBuild -> deprecated
	// TODO: ScorecardV1.PagerDuty -> keep as array or URLs
	// ScorecardV1.PagerDutyDetails -> derived data, do not capture
	oss.BypassSupportCompliances = si.MergeValues("BypassSupportCompliances", ScorecardV1{ossrecord.BypassSupportCompliances(sc.BypassSupportCompliances)}, PriorOSS{prior.BypassSupportCompliances}).(ossrecord.BypassSupportCompliances)
	oss.CertificateManagerCRNs = si.MergeValues("CertificateManagerCRNs", ScorecardV1{[]string(sc.CertificateManagerCRNs)}, PriorOSS{prior.CertificateManagerCRNs}).([]string)
	oss.CompletedSkillTransferAndEnablement = si.MergeValues("CompletedSkillTransferAndEnablement", ScorecardV1{sc.SupportCompliancesOSS007}, PriorOSS{prior.CompletedSkillTransferAndEnablement}).(bool)
}

func (si *ServiceInfo) mergeCatalogInfo() {
	if si.HasSourceMainCatalog() {
		// Attention: we should not copy any of these values from PriorOSS, if the entry is no longer in the Catalog
		oss := &si.OSSService.CatalogInfo
		cat := &si.SourceMainCatalog
		prior := &si.PriorOSS.CatalogInfo
		oss.Provider = si.MergeValues("Catalog.Provider", SeverityIfMissing{ossvalidation.MINOR}, Catalog{ossrecord.Person{Name: cat.Provider.Name, W3ID: cat.Provider.Email}}, PriorOSS{prior.Provider}).(ossrecord.Person)
		oss.ProviderContact = si.MergeValues("Catalog.ProviderContact", SeverityIfMissing{ossvalidation.MINOR}, Catalog{cat.Provider.Contact}, PriorOSS{prior.ProviderContact}).(string)
		oss.ProviderSupportEmail = si.MergeValues("Catalog.ProviderSupportEmail", SeverityIfMissing{ossvalidation.MINOR}, Catalog{cat.Provider.SupportEmail}, PriorOSS{prior.ProviderSupportEmail}).(string)
		oss.ProviderPhone = si.MergeValues("Catalog.ProviderPhone", SeverityIfMissing{ossvalidation.MINOR}, Catalog{cat.Provider.Phone}, PriorOSS{prior.ProviderPhone}).(string)
		if ossrunactions.Deployments.IsEnabled() {
			si.OSSValidation.RecordRunAction(ossrunactions.Deployments)
			if cex := si.GetCatalogExtra(false); cex != nil {
				oss.Locations = SortLocationsList(cex.Locations.Slice())
			} else {
				oss.Locations = nil
			}
		} else {
			si.OSSValidation.CopyRunAction(si.GetPriorOSSValidation(), ossrunactions.Deployments)
			if si.HasPriorOSS() && si.HasPriorOSSValidation() {
				oss.Locations = prior.Locations
			} else {
				oss.Locations = nil
			}
		}
	}
}

// checkOwnership checks if we have valid owners/contacts for this ServiceInfo
func (si *ServiceInfo) checkOwnership() {
	// TODO: Check that the W3IDs are valid in BluePages
	oss := si.OSSService
	if oss.Ownership.OfferingManager.IsValid() {
		return
	}
	if oss.Ownership.DevelopmentManager.IsValid() {
		return
	}
	//	if oss.Ownership.TechnicalContact.IsValid() {
	//		return
	//	}
	if oss.Support.Manager.IsValid() {
		return
	}
	if oss.Operations.Manager.IsValid() {
		return
	}
	if strings.Contains(oss.CatalogInfo.Provider.W3ID, "@") {
		return
	}
	if strings.Contains(oss.CatalogInfo.Provider.Name, "@") {
		return
	}
	if strings.Contains(oss.CatalogInfo.ProviderContact, "@") {
		return
	}
	if strings.Contains(oss.CatalogInfo.ProviderSupportEmail, "@") {
		return
	}
	/*
		if oss.Ownership.TribeOwner.IsValid() {
			return
		}
		if oss.Ownership.SegmentOwner.IsValid() {
			return
		}
	*/
	si.AddNamedValidationIssue(ossvalidation.NoValidOwnership, "")
}

// checkExpiredOSSTags determines if any OSS tags have expired
// and produces an validation for each expired tag
func (si *ServiceInfo) checkExpiredOSSTags() {
	expiredTags := si.OSSService.GeneralInfo.OSSTags.GetExpiredTags()
	for _, t := range expiredTags {
		si.AddValidationIssue(ossvalidation.CRITICAL, "OSS Tag is expired", "%s", t).TagExpired().TagCRN()
	}
}

// mergeParentResourceName merges the GeneralInfo.ParentResourceName attribute, which may link to an optional "parent" OSS entry
func (si *ServiceInfo) mergeParentResourceName() {
	si.checkMergePhase(mergePhaseServicesTwo)

	ossgi := &si.OSSService.GeneralInfo
	prior := &si.PriorOSS.GeneralInfo

	// TODO: Exploit parent information from the group relationships in Global Catalog itself

	var compositeParent parameter
	if si.mergeWorkArea.compositeParent != "" {
		compositeParent = Custom{N: "Composite.parent", V: si.mergeWorkArea.compositeParent}
		debug.Debug(debug.Composite, `Found composite parent, to be used as ParentResourceName if not overridden %s -> %s`, compositeParent, si.OSSService.ReferenceResourceName)
	}

	ossgi.ParentResourceName = si.MergeValues("Parent OSS entry (ParentResourceName)", SeverityIfMissing{ossvalidation.IGNORE}, SeverityIfMismatch{ossvalidation.IGNORE}, OverrideProperty{"GeneralInfo.ParentResourceName"}, RMCOSS{prior.ParentResourceName}, compositeParent, PriorOSS{prior.ParentResourceName}).(ossrecord.CRNServiceName)

	if ossgi.ParentResourceName == "" {
		return
	}

	if parent, found := LookupService(MakeComparableName(string(ossgi.ParentResourceName)), false); found {
		if ossgi.ParentResourceName != parent.OSSService.ReferenceResourceName {
			si.AddValidationIssue(ossvalidation.SEVERE, "Parent OSS entry (ParentResourceName) is found but does not match the canonical name", `ParentResourceName="%s"  CanonicalName="%s"`, ossgi.ParentResourceName, parent.OSSService.ReferenceResourceName).TagCRN()
		}
		if ossgi.OperationalStatus != parent.OSSService.GeneralInfo.OperationalStatus {
			if parent.OSSService.GeneralInfo.EntryType == ossrecord.COMPOSITE {
				si.AddValidationIssue(ossvalidation.MINOR, "Operational status of this entry does not match that of the parent COMPOSITE OSS entry (may be OK for a COMPOSITE)", "this entry->%s   parent(%s)->%s", ossgi.OperationalStatus, parent.OSSService.GeneralInfo.ParentResourceName, parent.OSSService.GeneralInfo.OperationalStatus).TagCRN()
			} else {
				si.AddValidationIssue(ossvalidation.WARNING, "Operational status of this entry does not match that of the parent OSS entry (ParentResourceName)", "this entry->%s   parent(%s)->%s", ossgi.OperationalStatus, parent.OSSService.GeneralInfo.ParentResourceName, parent.OSSService.GeneralInfo.OperationalStatus).TagCRN()
			}
		}

		// TODO: Check the EntryType of parent and child (not necessarily identical, but some combinations are obviously wrong)
	} else {
		si.AddValidationIssue(ossvalidation.SEVERE, "Parent OSS entry (ParentResourceName) not found", string(ossgi.ParentResourceName)).TagCRN()
	}
}

var kindIaaSOrService = collections.NewStringSet(catalogapi.KindIaaS, catalogapi.KindService)

// checkIgnoreCatalogGroup checks if this a Catalog group entry that should be ignored and returns true in that case
func (si *ServiceInfo) checkIgnoreCatalogGroup() (shouldIgnore bool) {
	if si.HasSourceMainCatalog() && si.GetSourceMainCatalog().Group {
		if si.OSSValidation.NumTrueSources() > 0 {
			si.AddValidationIssue(ossvalidation.INFO, "Catalog Group entry has other sources beside the Catalog itself -- cannot ignore it", "").TagCRN()
			return false
		}
		cex := si.GetCatalogExtra(true)
		if len(cex.Plans) > 0 {
			si.AddValidationIssue(ossvalidation.INFO, "Catalog Group entry contains some Plans as direct children -- cannot ignore it", "%s", si.ListPlans()).TagCRN()
			return false
		}
		switch si.GetSourceMainCatalog().Kind {
		case catalogapi.KindService:
			if cex.ChildrenKinds.Len() == 1 && cex.ChildrenKinds.Contains(catalogapi.KindService) {
				si.AddValidationIssue(ossvalidation.INFO, "Ignoring Catalog Group entry of kind=service that contains only other services as children", "original_name=%s", si.OSSService.ReferenceResourceName).TagCRN()
			} else {
				si.AddValidationIssue(ossvalidation.DEFERRED, "Catalog Group entry or type Service contains children of unexpected kinds -- cannot ignore it", "kind=%s  children=%v", si.GetSourceMainCatalog().Kind, cex.ChildrenKinds.Slice()).TagCRN()
				return false
			}
		case catalogapi.KindIaaS:
			if leftOnly, _ := cex.ChildrenKinds.Compare(kindIaaSOrService); leftOnly == nil {
				si.AddValidationIssue(ossvalidation.INFO, "Ignoring Catalog Group entry of kind=iaas that contains only other iaas as children", "original_name=%s", si.OSSService.ReferenceResourceName).TagCRN()
			} else {
				si.AddValidationIssue(ossvalidation.DEFERRED, "Catalog Group entry or type IaaS contains children of unexpected kinds -- cannot ignore it", "kind=%s  children=%v", si.GetSourceMainCatalog().Kind, cex.ChildrenKinds.Slice()).TagCRN()
				return false
			}
		default:
			si.AddValidationIssue(ossvalidation.DEFERRED, "Catalog Group entry contains children of unexpected kinds -- cannot ignore it", "kind=%s  children=%v", si.GetSourceMainCatalog().Kind, cex.ChildrenKinds.Slice()).TagCRN()
			return false
		}
		return true
	}
	return false
}
