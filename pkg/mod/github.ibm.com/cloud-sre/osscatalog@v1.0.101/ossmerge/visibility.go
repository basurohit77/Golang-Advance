package ossmerge

import (
	"fmt"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// checkIgnoreMainCatalog determines if the Main Catalog entry should be ignored,
// based on visibility restrictions and various flags.
// If necessary, generate a Validation issue and remove the Main Catalog entry from the
// sources in the ServiceInfo record
func (si *ServiceInfo) checkIgnoreMainCatalog() {
	// TODO: need to clarify expected semantics of Catalog "Active" and "Disabled" flags

	if !si.HasSourceMainCatalog() {
		return
	}

	if !si.GetSourceMainCatalog().IsPublicVisible() {
		cat := si.GetSourceMainCatalog()
		si.OSSValidation.RecordCatalogVisibility(cat.EffectiveVisibility.Restrictions, cat.Visibility.Restrictions, cat.Active, cat.ObjectMetaData.UI.Hidden, cat.Disabled)
	}

	if si.isMainCatalogIgnored(si.GetSourceMainCatalog()) {
		debug.Debug(debug.Visibility, "checkIgnoreMainCatalog(%s) - ignoring Catalog entry %q", si.String(), si.GetSourceMainCatalog().Name)
		// Blank everything
		saved := *si.GetSourceMainCatalog() // Make a copy
		si.IgnoredMainCatalog = &saved
		si.SourceMainCatalog = catalogapi.Resource{}
	}
}

func (si *ServiceInfo) isMainCatalogIgnored(r *catalogapi.Resource) bool {
	var mustIgnore bool
	if r.EffectiveVisibility.Restrictions == string(catalogapi.VisibilityPrivate) {
		if si.GeneralInfo.OSSOnboardingPhase != "" && si.GeneralInfo.OperationalStatus == ossrecord.SELECTAVAILABILITY {
			si.AddValidationIssue(ossvalidation.INFO, "Found private Main Catalog entry -- for a LimitedAvailability service; cannot verify proper visibility (from RMC)", "%s", r.String()).TagCatalogVisibility().TagCRN()
		} else if si.OSSMergeControl.OSSTags.Contains(osstags.SelectAvailability) {
			si.AddValidationIssue(ossvalidation.INFO, "Found private Main Catalog entry -- for a LimitedAvailability service; cannot verify proper visibility (from OSSTag)", "%s", r.String()).TagCatalogVisibility().TagCRN()
		} else if r.Kind == "platform_service" {
			si.AddValidationIssue(ossvalidation.MINOR, "Found private Main Catalog entry -- for a Kind=platform_service; cannot verify proper visibility", "%s", r.String()).TagCatalogVisibility().TagCRN()
		} else if r.Kind == "composite" {
			si.AddValidationIssue(ossvalidation.MINOR, "Found private Main Catalog entry -- for a Kind=composite; cannot verify proper visibility", "%s", r.String()).TagCatalogVisibility().TagCRN()
		} else {
			mustIgnore = true
			si.AddValidationIssue(ossvalidation.MINOR, "Ignoring private Main Catalog entry", "%s", r.String()).TagCatalogVisibility().TagCRN()
		}
	}

	if r.Disabled {
		mustIgnore = true
		si.AddValidationIssue(ossvalidation.MINOR, "Ignoring Main Catalog entry with flag Disabled=true", "%s", r.String()).TagCatalogVisibility().TagCRN()
	}

	return mustIgnore
}

// checkCatalogVisibility checks for issues related to the visibility of the Main Catalog entry
// and generates appropriate ValidationIssues.
// This method also takes into account a number of exception cases.
//
// Note: we cannot do this before we completed the merges for EntryType and OperationalStatus,
// because we use that information to determine which entries might not need to be publicly visible.
//
// This method assumes that checkIgnoreMainCatalog() has already executed, which
// ensures that we do not take into account entries that truly should not be visible
func (si *ServiceInfo) checkCatalogVisibility() {
	et := si.OSSService.GeneralInfo.EntryType
	os := si.OSSService.GeneralInfo.OperationalStatus
	futos := si.OSSService.GeneralInfo.FutureOperationalStatus

	if et == "" {
		panic(fmt.Sprintf(`checkCatalogVisibility(): missing EntryType for "%s"`, si.String()))
	}
	if os == "" {
		panic(fmt.Sprintf(`checkCatalogVisibility(): missing OperationalStatus for "%s"`, si.String()))
	}

	if !si.HasSourceMainCatalog() {
		switch et {
		case ossrecord.SERVICE, ossrecord.RUNTIME, ossrecord.TEMPLATE, ossrecord.IAAS, ossrecord.COMPOSITE, ossrecord.VMWARE:
			switch os {
			case ossrecord.GA, ossrecord.BETA, ossrecord.EXPERIMENTAL, ossrecord.SELECTAVAILABILITY, ossrecord.THIRDPARTY, ossrecord.COMMUNITY:
				if si.HasSourceRMC() && si.OSSService.GeneralInfo.OSSOnboardingPhase != "" {
					si.AddValidationIssue(ossvalidation.SEVERE, "Missing or non-public Main Catalog entry for a GA/BETA/EXPERIMENTAL/THIRDPARTY/COMMUNITY/etc. service (with a RMC entry; RMC entry prevails)", "EntryType=%s  RMC.OperationalStatus=%s   RMC.FutureOperationalStatus=%v", et, os, futos).TagCatalogVisibility().TagCRN()
					/* TODO: fix OperationalStatus mismatch in RMC https://github.ibm.com/cloud-sre/osscatalog/issues/342
					if futos != "" && futos != os {
						si.AddValidationIssue(ossvalidation.SEVERE, "Missing or non-public Main Catalog entry for a GA/BETA/EXPERIMENTAL/THIRDPARTY/COMMUNITY/etc. service (with a RMC entry) -- assuming the current OperationalStatus is <unknown> and RMC has an explicit FutureOperationalStatus", "EntryType=%s  Original OperationalStatus=%s   FutureOperationalStatus=%v",
							et, os, futos).TagCatalogVisibility().TagCRN()
						si.OSSService.GeneralInfo.FutureOperationalStatus = si.OSSService.GeneralInfo.OperationalStatus
						// XXX si.OSSService.GeneralInfo.OperationalStatus = ossrecord.NOTREADY
						si.OSSService.GeneralInfo.OperationalStatus = ossrecord.OperationalStatusUnknown
					} else {
						si.AddValidationIssue(ossvalidation.SEVERE, "Missing or non-public Main Catalog entry for a GA/BETA/EXPERIMENTAL/THIRDPARTY/COMMUNITY/etc. service (with a RMC entry) -- assuming the current OperationalStatus is <unknown> and the current status from RMC is actually the FutureOperationalStatus", "EntryType=%s  Original/Future OperationalStatus=%s",
							et, os).TagCatalogVisibility().TagCRN()
						si.OSSService.GeneralInfo.FutureOperationalStatus = si.OSSService.GeneralInfo.OperationalStatus
						// XXX si.OSSService.GeneralInfo.OperationalStatus = ossrecord.NOTREADY
						si.OSSService.GeneralInfo.OperationalStatus = ossrecord.OperationalStatusUnknown
					}
					*/
				} else {
					si.AddValidationIssue(ossvalidation.SEVERE, "Missing or non-public Main Catalog entry for a GA/BETA/EXPERIMENTAL/THIRDPARTY/COMMUNITY/etc. service (and no RMC entry) -- cannot determine actual OperationalStatus - resetting", "EntryType=%s  Original OperationalStatus=%s",
						et, os).TagCatalogVisibility().TagCRN()
					si.OSSService.GeneralInfo.OperationalStatus = ossrecord.OperationalStatusUnknown
					// Leave the FutureOperationalStatus alone
				}
			}
		}
		return
	}

	cat := si.GetSourceMainCatalog()

	switch et {
	case ossrecord.SERVICE, ossrecord.RUNTIME, ossrecord.TEMPLATE, ossrecord.IAAS, ossrecord.COMPOSITE:
		switch os {
		case ossrecord.DEPRECATED:
			if cat.IsPublicVisible() {
				si.AddValidationIssue(ossvalidation.INFO, "Main Catalog for DEPRECATED entry is publicly visible -- assuming first phase of deprecation, announced but not yet in effect", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
					et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
			} else {
				if cat.Active {
					si.AddValidationIssue(ossvalidation.WARNING, "Main Catalog for DEPRECATED entry is not publicly visible but still Active=true", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
						et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
				}
				if cat.Visibility.Restrictions != cat.EffectiveVisibility.Restrictions {
					si.AddValidationIssue(ossvalidation.MINOR, "Main Catalog for DEPRECATED entry is hidden through its parent but has different local visibilty than its parent", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
						et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
				}
			}
		case ossrecord.SELECTAVAILABILITY:
			if cat.EffectiveVisibility.Restrictions == string(catalogapi.VisibilityPublic) {
				si.AddValidationIssue(ossvalidation.WARNING, "Main Catalog for LIMITEDAVAILABILITY entry has public visibility", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
					et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
			} else if cat.Visibility.Restrictions != cat.EffectiveVisibility.Restrictions {
				si.AddValidationIssue(ossvalidation.MINOR, "Main Catalog for LIMITEDAVAILABILITY entry is hidden through its parent but has different local visibilty than its parent", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
					et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
			}
			if !cat.Active {
				si.AddValidationIssue(ossvalidation.WARNING, "Main Catalog for LIMITEDAVAILABILITY entry has Active=false -- not really available even to limited users?", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
					et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()

			}
		case ossrecord.RETIRED, ossrecord.INTERNAL, ossrecord.NOTREADY, ossrecord.OperationalStatusUnknown:
			if cat.IsPublicVisible() {
				si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog for RETIRED/INTERNAL/NOTREADY entry is publicly visible", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
					et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
			} else if cat.Visibility.Restrictions != cat.EffectiveVisibility.Restrictions {
				si.AddValidationIssue(ossvalidation.MINOR, "Main Catalog for RETIRED/INTERNAL/NOTREADY entry is hidden through its parent but has different local visibilty than its parent", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
					et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
			}
		default:
			if !cat.IsPublicVisible() {
				if !cat.Active {
					si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog for a GA/BETA/EXPERIMENTAL/THIRDPARTY/COMMUNITY/etc. entry is Active=false -- forcing to DEPRECATED status", "EntryType=%s  Original OperationalStatus=%s  Visibility=%+v",
						et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
					si.OSSService.GeneralInfo.OperationalStatus = ossrecord.DEPRECATED
					// Leave the FutureOperationalStatus alone
				} else {
					if si.HasSourceRMC() && si.OSSService.GeneralInfo.OSSOnboardingPhase != "" {
						si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog entry for a GA/BETA/EXPERIMENTAL/THIRDPARTY/COMMUNITY/etc. service is not publicly visible but has Active=true (with a RMC entry; RMC entry prevails)", "EntryType=%s  RMC.OperationalStatus=%s   RMC.FutureOperationalStatus=%v", et, os, futos).TagCatalogVisibility().TagCRN()
						/* TODO: fix OperationalStatus mismatch in RMC https://github.ibm.com/cloud-sre/osscatalog/issues/342
						if futos != "" && futos != os {
							si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog for a GA/BETA/EXPERIMENTAL/THIRDPARTY/COMMUNITY/etc. entry is not publicly visible but has Active=true (with a RMC entry) -- assuming the current OperationalStatus is <unknown> and RMC has an explicit FutureOperationalStatus", "EntryType=%s  Original OperationalStatus=%s   FutureOperationalStatus=%v  Visibility=%+v",
								et, os, futos, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
							si.OSSService.GeneralInfo.FutureOperationalStatus = si.OSSService.GeneralInfo.OperationalStatus
							// XXX si.OSSService.GeneralInfo.OperationalStatus = ossrecord.NOTREADY
							si.OSSService.GeneralInfo.OperationalStatus = ossrecord.OperationalStatusUnknown
						} else {
							si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog for a GA/BETA/EXPERIMENTAL/THIRDPARTY/COMMUNITY/etc. entry is not publicly visible but has Active=true (with a RMC entry) -- assuming the current OperationalStatus is <unknown> and the current status from RMC is actually the FutureOperationalStatus", "EntryType=%s  Original/Future OperationalStatus=%s  Visibility=%+v",
								et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
							si.OSSService.GeneralInfo.FutureOperationalStatus = si.OSSService.GeneralInfo.OperationalStatus
							// XXX si.OSSService.GeneralInfo.OperationalStatus = ossrecord.NOTREADY
							si.OSSService.GeneralInfo.OperationalStatus = ossrecord.OperationalStatusUnknown
						}
						*/
					} else {
						si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog for a GA/BETA/EXPERIMENTAL/THIRDPARTY/COMMUNITY/etc. entry is not publicly visible but has Active=true (and no RMC entry) -- cannot determine actual OperationalStatus - resetting", "EntryType=%s  Original OperationalStatus=%s  Visibility=%+v",
							et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
						si.OSSService.GeneralInfo.OperationalStatus = ossrecord.OperationalStatusUnknown
						// Leave the FutureOperationalStatus alone
					}
				}
				if cat.Visibility.Restrictions != cat.EffectiveVisibility.Restrictions {
					si.AddValidationIssue(ossvalidation.MINOR, "Main Catalog for GA/BETA/EXPERIMENTAL/THIRDPARTY/COMMUNITY/etc. entry is hidden through its parent but has different local visibilty than its parent", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
						et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
				}
			}
		}
	case ossrecord.VMWARE:
		if cat.IsPublicVisible() {
			si.AddValidationIssue(ossvalidation.WARNING, "Main Catalog entry for a VMWARE service is publicly visible -- expected to be inside the main VMWARE tile", "").TagCatalogVisibility().TagCRN()
		}
	case ossrecord.OTHEROSS:
		si.AddValidationIssue(ossvalidation.CRITICAL, "Found a Catalog entry for type OTHEROSS -- not expected to have a main Catalog entry at all", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
			et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
	case ossrecord.GAAS:
		si.AddValidationIssue(ossvalidation.CRITICAL, "Found a Catalog entry for type GAAS -- not expected to have a main Catalog entry at all", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
			et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
	case ossrecord.PLATFORMCOMPONENT, ossrecord.SUBCOMPONENT, ossrecord.SUPERCOMPONENT:
		if cat.EffectiveVisibility.Restrictions != string(catalogapi.VisibilityIBMOnly) {
			si.AddValidationIssue(ossvalidation.MINOR, "Main Catalog entry for a PLATFORMCOMPONENT/SUBCOMPONENT/SUPERCOMPONENT has visibility other than IBM-only", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
				et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
		} else if cat.Visibility.Restrictions != cat.EffectiveVisibility.Restrictions {
			si.AddValidationIssue(ossvalidation.MINOR, "Main Catalog for PLATFORMCOMPONENT/SUBCOMPONENT/SUPERCOMPONENT has visibility IBM-only through its parent but has different local visibilty than its parent", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
				et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
		}
		fallthrough
	default:
		// Cannot draw useful conclusions for non-provisionable entry types, as they are never publicly visible anyway, by their nature
		if cat.Active {
			si.AddValidationIssue(ossvalidation.MINOR, "Main Catalog entry for a PLATFORMCOMPONENT or other non-provisionable entry is Active=true -- assuming this is irrelevant", "EntryType=%s  OperationalStatus=%s  Visibility=%+v",
				et, os, si.OSSValidation.CatalogVisibility).TagCatalogVisibility().TagCRN()
		}
		// TODO: More detailed checks for use of specific visibility scopes for non-provisionable entries
	}
}
