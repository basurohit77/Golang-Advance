package ossvalidation

import (
	"fmt"

	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"

	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// Definitions for named ValidationIssue patterns, that can be referenced in multiple places.
var (
	NoValidOwnership                    = newNamedIssue(SEVERE, "No valid ownership or contact information found (no Offering Manager, etc.)", TagCRN)
	StatusCategoryParent                = newNamedIssue(INFO, "This entry is the Status Page Category Parent for its CategoryID", TagStatusPage)
	PnPEnabledWithoutCandidate          = newNamedIssue(INFO, fmt.Sprintf(`*Entry is "%s" -- meets all the criteria for PnP enablement `, osstags.PnPEnabled), TagStatusPage)
	PnPNotEnabledWithoutCandidate       = newNamedIssue(INFO, fmt.Sprintf(`*Entry is not "%s" -- does not meet one or more criteria for PnP enablement and has no special Pnp enablement control tags`, osstags.PnPEnabled), TagStatusPage)
	PnPNotEnabledMissBasicCriteria      = newNamedIssue(INFO, fmt.Sprintf(`*Entry is not "%s" -- does not meet basic criteria for PnP enablement (type, status)  and has no special Pnp enablement control tags`, osstags.PnPEnabled), TagStatusPage)
	PnPEnabledWithUnnecessaryCandidate  = newNamedIssue(MINOR, fmt.Sprintf(`*Entry is "%s" and has an unnecessary "%s" tag -- already meets all the criteria for PnP enablement`, osstags.PnPEnabled, osstags.PnPCandidate), TagStatusPage)
	PnPNotEnabledButCandidate           = newNamedIssue(WARNING, fmt.Sprintf(`*Entry cannot be "%s" even though it has the "%s" tag because it does not meet one or more criteria for PnP enablement`, osstags.PnPEnabled, osstags.PnPCandidate), TagStatusPage)
	PnPEnabledWithInclude               = newNamedIssue(WARNING, fmt.Sprintf(`*Entry is "%s" because of "%s" tag -- but does not meet one or more overridable criteria for PnP enablement`, osstags.PnPEnabled, osstags.PnPInclude), TagStatusPage)
	PnPEnabledWithUnnecessaryInclude    = newNamedIssue(MINOR, fmt.Sprintf(`*Entry is "%s" and has an unnecessary "%s" tag -- already meets all the criteria for PnP enablement`, osstags.PnPEnabled, osstags.PnPInclude), TagStatusPage)
	PnPNotEnabledButInclude             = newNamedIssue(SEVERE, fmt.Sprintf(`*Entry cannot be "%s" even though it has the "%s" tag because it does not meet one or more non-overridable criteria for PnP enablement`, osstags.PnPEnabled, osstags.PnPInclude), TagStatusPage)
	PnPNotEnabledWithExclude            = newNamedIssue(MINOR, fmt.Sprintf(`*Entry is not "%s" because it has the "%s" tag -- even though it meets all the criteria for PnP enablement`, osstags.PnPEnabled, osstags.PnPExclude), TagStatusPage)
	PnPNotEnabledWithUnnecessaryExclude = newNamedIssue(MINOR, fmt.Sprintf(`*Entry is not "%s" and has an unnecessary "%s" tag -- already does not meet one or more criteria for PnP enablement`, osstags.PnPEnabled, osstags.PnPExclude), TagStatusPage)
	PnPNotEnabledButEnabledIaaS         = newNamedIssue(WARNING, fmt.Sprintf(`Entry has the "%s" tag but cannot be "%s" because it does not meet one or more criteria for enablement`, osstags.PnPEnabledIaaS, osstags.PnPEnabled), TagStatusPage)
	PnPEnabledAdded                     = newNamedIssue(INFO, fmt.Sprintf(`Entry newly added to the %s list`, osstags.PnPEnabled), TagStatusPage)
	PnPEnabledRemoved                   = newNamedIssue(INFO, fmt.Sprintf(`Entry newly removed from the %s list`, osstags.PnPEnabled), TagStatusPage)
	CHMultipleOSS                       = newNamedIssue(SEVERE, "More than one OSS entry found with similar names (after merging all names including ClearingHouse names) -- see log for details", TagProductInfo /*, TagCRN*/)
	CHSomeFoundByPIDOnly                = newNamedIssue(MINOR, "Some ClearingHouse entries found by matching PIDs only, no other sources", TagProductInfo)
	CHSomeFoundByNameOnly               = newNamedIssue(MINOR, "Some ClearingHouse entries found by matching names only, no other sources", TagProductInfo)
	CHSomeFoundByCRNOnly                = newNamedIssue(MINOR, "Some ClearingHouse entries found by the CRN service-name attribute in ClearingHouse, no other sources", TagProductInfo)
	CHNotFoundByPID                     = newNamedIssue(MINOR, "Some ClearingHouse entries not found by matching PIDs, only found through other sources", TagProductInfo)
	CHNotFoundByName                    = newNamedIssue(MINOR, "Some ClearingHouse entries not found by matching names, only found through other sources", TagProductInfo)
	CHNotFoundByCRN                     = newNamedIssue(MINOR, "Some ClearingHouse entries not found by the CRN service-name attribute in ClearingHouse, only found through other sources", TagProductInfo)
	CHFoundByCRN                        = newNamedIssue(INFO, "Some (or all) ClearingHouse entries found by the CRN service-name attribute in ClearingHouse, and possibly by other sources", TagProductInfo)
	CHMissingCRN                        = newNamedIssue(MINOR, "ClearingHouse entry does not contain any CRN service-name information", TagProductInfo)
	CHBadCRN                            = newNamedIssue(SEVERE, "ClearingHouse entry contains CRN service-name information that does not reference a valid OSS record", TagProductInfo)
	CHNoNameGroup                       = newNamedIssue(WARNING, "ClearingHouse entry does not contain sufficient name information to validate a match by CRN or PID", TagProductInfo)
	CHMismatchNameGroup                 = newNamedIssue(WARNING, "ClearingHouse entry name information does not overlap the name of the OSS entry to which it is being associated (by CRN or PID)", TagProductInfo)
	CHDependencyInboundNotOSS           = newNamedIssue(WARNING, "Some inbound (Originators) dependencies in ClearingHouse do not map to a known OSS record", RunActionToTag(ossrunactions.DependenciesClearingHouse))
	CHDependencyInboundNotCloud         = newNamedIssue(INFO, "Some inbound (Originators) dependencies in ClearingHouse do not appear to be valid, current Cloud items", RunActionToTag(ossrunactions.DependenciesClearingHouse))
	CHDependencyOutboundNotOSS          = newNamedIssue(WARNING, "Some outbound (Providers) dependencies in ClearingHouse do not map to a known OSS record", RunActionToTag(ossrunactions.DependenciesClearingHouse))
	CHDependencyOutboundNotCloud        = newNamedIssue(INFO, "Some outbound (Providers) dependencies in ClearingHouse do not appear to be valid, current Cloud items", RunActionToTag(ossrunactions.DependenciesClearingHouse))
	EDBMissingButRequired               = newNamedIssue(SEVERE, fmt.Sprintf(`Found no Availability Monitoring info (EDB) - but explicitly required by the "%s" tag`, osstags.EDBInclude), RunActionToTag(ossrunactions.Monitoring))
	EDBMissingClientFacing              = newNamedIssue(WARNING, "Found no Availability Monitoring info (EDB) - Client-facing service or component", RunActionToTag(ossrunactions.Monitoring))
	EDBMissingClientFacingNonStandard   = newNamedIssue(MINOR, "Found no Availability Monitoring info (EDB) - Client-facing service or component (non-standard type or status)", RunActionToTag(ossrunactions.Monitoring))
	EDBMissingNotClientFacing           = newNamedIssue(MINOR, "Found no Availability Monitoring info (EDB) - non Client-facing service or component meeting other standard criteria", RunActionToTag(ossrunactions.Monitoring))
	EDBFoundButExcluded                 = newNamedIssue(SEVERE, fmt.Sprintf(`Found some Availability Monitoring info (EDB) - but explicitly excluded by the "%s" tag`, osstags.EDBExclude), RunActionToTag(ossrunactions.Monitoring))
	EDBExcludedOK                       = newNamedIssue(INFO, fmt.Sprintf(`Availability Monitoring info (EDB) is explicitly excluded by the "%s" tag (and none found)`, osstags.EDBExclude), RunActionToTag(ossrunactions.Monitoring))
	EDBIncludedOK                       = newNamedIssue(INFO, fmt.Sprintf(`Availability Monitoring info (EDB) is explicitly included by the "%s" tag (and found)`, osstags.EDBInclude), RunActionToTag(ossrunactions.Monitoring))
	InvalidatedByRMC                    = newNamedIssue(CRITICAL, "All existing Validation Issues may be invalid due to editing the entry in RMC -- waiting to rerun the validator")
)
