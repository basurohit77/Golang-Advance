package ossrecord

import (
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// Data structures that describe a OSS record in the Global Catalog (along with underlying field types)

//go:generate easytags $GOFILE

// OSSService is the main record for OSS information for one service/component in the Global Catalog
type OSSService struct {
	SchemaVersion         string         `json:"schema_version"`          // The version of the schema for this OSS record entry
	ReferenceResourceName CRNServiceName `json:"reference_resource_name"` // CRNServiceName of the main Global Catalog entry (if any) that represents the service or component for which this entry contains the OSS information
	ReferenceDisplayName  string         `json:"reference_display_name"`  // DisplayName from the main Global Catalog entry (if any) -- should be the same as in ServiceNow
	ReferenceCatalogID    CatalogID      `json:"reference_catalog_id"`    // ID of the main Global Catalog entry (if any) that represents the service or component for which this entry contains the OSS information
	ReferenceCatalogPath  string         `json:"reference_catalog_path"`  // path of the main Global Catalog entry (if any) that represents the service or component for which this entry contains the OSS information
	// No CRNServiceName for the OSS record itself - comes from the name attribute of the enclosing GC entry
	// No DisplayName for the OSS record itself - comes from the enclosing GC entry

	GeneralInfo struct {
		RMCNumber                 string             `json:"rmc_number"` // historical - replace with CRN Service Name alone
		OperationalStatus         OperationalStatus  `json:"operational_status"`
		FutureOperationalStatus   OperationalStatus  `json:"future_operational_status"`
		OSSTags                   osstags.TagSet     `json:"oss_tags"`
		OSSOnboardingPhase        OSSOnboardingPhase `json:"oss_onboarding_phase"`
		OSSOnboardingApprover     Person             `json:"oss_onboarding_approver"`      // Person who approved this OSS record in RMC
		OSSOnboardingApprovalDate string             `json:"oss_onboarding_approval_date"` // Date that the person approved this OSS record in RMC
		ClientFacing              bool               `json:"client_facing"`
		EntryType                 EntryType          `json:"entry_type"`
		OSSDescription            string             `json:"oss_description"`
		ServiceNowSysid           ServiceNowSysid    `json:"servicenow_sys_id"`
		ServiceNowCIURL           string             `json:"service_now_ciurl"`
		ParentResourceName        CRNServiceName     `json:"parent_resource_name"` // CRNServiceName of the parent for this entry, i.e. some service for which this entry is a component or subcomponent
		Domain                    Domain             `json:"domain"`               // Domain of the OSS service record (for example, commercial or US regulated)
	} `json:"general_info"`

	Ownership struct {
		OfferingManager            Person    `json:"offering_manager"`
		DevelopmentManager         Person    `json:"development_manager"`
		TechnicalContactDEPRECATED Person    `json:"technical_contact"`
		SegmentName                string    `json:"segment_name"`
		SegmentID                  SegmentID `json:"segment_id"`
		SegmentOwner               Person    `json:"segment_owner"`
		TribeName                  string    `json:"tribe_name"`
		TribeID                    TribeID   `json:"tribe_id"`
		TribeOwner                 Person    `json:"tribe_owner"`
		ServiceOffering            string    `json:"service_offering"`
		MainRepository             GHRepo    `json:"main_repository"`
	} `json:"ownership"`

	Support struct {
		Manager                    Person              `json:"manager"`
		ClientExperience           ClientExperience    `json:"client_experience"`
		SpecialInstructions        string              `json:"special_instructions"`
		Tier2EscalationType        Tier2EscalationType `json:"tier2_escalation_type"`
		Slack                      SlackChannel        `json:"slack"`
		Tier2Repository            GHRepo              `json:"tier2_repository"`
		Tier2RTC                   RTCCategory         `json:"tier2rtc"`
		ThirdPartySupportURL       string              `json:"third_party_support_url"`
		ThirdPartyCaseProcess      string              `json:"third_party_case_process"`
		ThirdPartyContacts         string              `json:"third_party_contacts"`
		ThirdPartyCountryLocations string              `json:"third_party_country_locations"`
	} `json:"support"`

	Operations struct {
		Manager                   Person                        `json:"manager"`
		SpecialInstructions       string                        `json:"special_instructions"`
		TIPOnboarded              bool                          `json:"tip_onboarded"`
		AVMEnabled                bool                          `json:"avm_enabled"`
		BypassProductionReadiness BypassPassProductionReadiness `json:"bypass_production_readiness"`
		RunbookEnabled            bool                          `json:"runbook_enabled"`
		TOCAVMFocal               Person                        `json:"toc_avm_focal"`
		CIEDistList               string                        `json:"cie_dist_list"`
		Tier2EscalationType       Tier2EscalationType           `json:"tier2_escalation_type"`
		Slack                     SlackChannel                  `json:"slack"`
		Tier2Repository           GHRepo                        `json:"tier2_repository"`
		Tier2RTC                  RTCCategory                   `json:"tier2rtc"`
		EUAccessUSAMName          string                        `json:"euaccess_usam_name"` // "euaccess_emerg_usam_service_name: attribute from ScorecardV1"
		AutomationIDs             []*PersonListEntry            `json:"automation_ids"`     // list of functional IDs used for Operations
	} `json:"operations"`

	StatusPage struct {
		Group                string         `json:"group"`
		CategoryID           string         `json:"category_id"`
		CategoryParent       CRNServiceName `json:"category_parent"`
		CategoryIDMisspelled string         `json:"cateogry_id"` // XXX temporary
	} `json:"status_page"`

	Compliance struct {
		ServiceNowOnboarded                 bool                       `json:"servicenow_onboarded"`
		ArchitectureFocal                   Person                     `json:"architecture_focal"`
		BCDRFocal                           Person                     `json:"bcdr_focal"`
		SecurityFocal                       Person                     `json:"security_focal"`
		ProvisionMonitors                   AvailabilityMonitoringInfo `json:"availability_monitors"`
		ConsumptionMonitors                 AvailabilityMonitoringInfo `json:"consumption_monitors"`
		PagerDutyURLs                       []string                   `json:"pagerduty_urls"`
		OnboardingContact                   Person                     `json:"onboarding_contact"`           // Set and used by RMC UI
		OnboardingIssueTrackerURL           string                     `json:"onboarding_issue_tracker_url"` // Set and used by RMC UI
		BypassSupportCompliances            BypassSupportCompliances   `json:"bypass_support_compliances"`
		CertificateManagerCRNs              []string                   `json:"certificate_manager_crns"`
		CompletedSkillTransferAndEnablement bool                       `json:"completed_skill_transfer_and_enablement"`
	} `json:"compliance"`

	AdditionalContacts string `json:"additional_contacts"`

	ServiceNowInfo struct { // "ghosting" info from ServiceNow
		SupportTier1AG        string `json:"support_tier1ag"`
		SupportTier2AG        string `json:"support_tier2ag"`
		OperationsTier1AG     string `json:"operations_tier1ag"`
		OperationsTier2AG     string `json:"operations_tier2ag"`
		RCAApprovalGroup      string `json:"rca_approval_group"`
		ERCAApprovalGroup     string `json:"erca_approval_group"`
		TargetedCommunication string `json:"targeted_communication"`
		CIEPageout            string `json:"cie_pageout"`
		//BackupContacts          string `json:"backup_contacts"` // see AdditionalContacts section
		//GBT30                   string `json:"gbt30"`           // see ProductInfo section
		//District                string `json:"district"`        // see ProductInfo section
		SupportNotApplicable    bool `json:"support_not_applicable"`
		OperationsNotApplicable bool `json:"operations_not_applicable"`
	} `json:"service_now_info"`

	CatalogInfo struct { // "ghosting" info from the Main Catalog entry
		Provider             Person   `json:"provider"`
		ProviderContact      string   `json:"provider_contact"`
		ProviderSupportEmail string   `json:"provider_support_email"`
		ProviderPhone        string   `json:"provider_phone"`
		CategoryTags         string   `json:"category_tags"`
		CatalogClientFacing  bool     `json:"catalog_client_facing"`
		Locations            []string `json:"locations"`
	} `json:"catalog_info"`

	ProductInfo struct {
		PartNumbers             []string                `json:"part_numbers"`
		PartNumbersRefreshed    string                  `json:"part_numbers_refreshed"`
		ProductIDs              []string                `json:"product_ids"`
		ProductIDSource         ProductIDSource         `json:"product_id_source"`
		ClearingHouseReferences ClearingHouseReferences `json:"clearinghouse_references"`
		Taxonomy                Taxonomy                `json:"taxonomy"`
		Division                string                  `json:"division"`
		OSSUID                  string                  `json:"oss_uid"` // A unique ID for this OSS entry, preserved across all renames or changes of state. Can be used as a substitute for a PID if no PID is available
	} `json:"product_info"`

	DependencyInfo struct {
		OutboundDependencies Dependencies `json:"outbound_dependencies"`
		InboundDependencies  Dependencies `json:"inbound_dependencies"`
	} `json:"dependency_info"`

	MonitoringInfo struct {
		Metrics []*Metric `json:"metrics"`
	} `json:"monitoring_info"`

	Overrides []OSSServiceOverride `json:"overrides,omitempty"` // overridden values per domain
}

// OSSServiceOverride stores values that are overridden per domain
type OSSServiceOverride struct {
	GeneralInfo struct {
		OSSTags         osstags.TagSet  `json:"oss_tags"`
		ServiceNowSysid ServiceNowSysid `json:"servicenow_sys_id"`
		ServiceNowCIURL string          `json:"service_now_ciurl"`
		Domain          Domain          `json:"domain"`
	} `json:"general_info"`

	Compliance struct {
		ServiceNowOnboarded bool `json:"servicenow_onboarded"`
	} `json:"compliance"`
}

// MakeOSSServiceID creates a OSSEntryID for a OSS service/component record
func MakeOSSServiceID(name CRNServiceName) OSSEntryID {
	if name != "" {
		return OSSEntryID("oss." + string(name))
	}
	return ""
}

// GetOSSEntryID returns a unique identifier for this record
func (oss *OSSService) GetOSSEntryID() OSSEntryID {
	return MakeOSSServiceID(oss.ReferenceResourceName)
}

// CRNServiceName represents the CRN service-name for one service or component
// (corresponding to the "name" attribute in each Global Catalog entry, shown in the UI as "Programmatic name")
type CRNServiceName string

// CatalogID represents the "id" attribute in each Global Catalog entry
type CatalogID string

// ServiceNowSysid represents the "sysid" used to uniquely identify ServiceNow Configuration Items
type ServiceNowSysid string

// EntryType represents the type of a service or component represented by an OSS entry
type EntryType string

const (
	// SERVICE is the EntryType for a OSS entry representing a plain service in the IBM Cloud catalog
	SERVICE EntryType = "SERVICE"
	// VMWARE is the EntryType for a OSS entry representing a VMware service in the IBM Cloud catalog
	VMWARE EntryType = "SERVICE/VMWARE"
	// RUNTIME is the EntryType for a OSS entry representing a runtime in the IBM Cloud catalog
	RUNTIME EntryType = "RUNTIME"
	// TEMPLATE is the EntryType for a OSS entry representing a template (aka "boilerplate" or "starter") in the IBM Cloud catalog
	TEMPLATE EntryType = "TEMPLATE"
	// IAAS is the EntryType for a OSS entry representing an Infrastructure (IaaS) service in the IBM Cloud catalog
	IAAS EntryType = "IAAS"
	// PLATFORMCOMPONENT is the EntryType for a OSS entry representing an internal component or tool in the IBM Cloud
	// (not directly visible to Clients in the catalog)
	PLATFORMCOMPONENT EntryType = "PLATFORM_COMPONENT"
	// SUBCOMPONENT is the EntryType for a OSS entry representing a subcomponent of a service in the IBM Cloud catalog
	// (not itself directly visible to Clients in the catalog)
	SUBCOMPONENT EntryType = "SUBCOMPONENT"
	// SUPERCOMPONENT is the EntryType for a OSS entry representing a supercomponent shared by multiple services in the IBM Cloud catalog
	// (not itself directly visible to Clients in the catalog)
	// SUPERCOMPONENTs are related to COMPOSITEs (see below) but they may differ in two key aspects:
	//  - a COMPOSITE is a specific main entry type in the IBM Cloud catalog, with specifc semantics with respect to IAM handling and a specific internal structure
	//  - one service may reference multiple SUPERCOMPONENTS but only a single COMPOSITE (not yet implemented: )
	SUPERCOMPONENT EntryType = "SUPERCOMPONENT"
	// COMPOSITE is the EntryType for a OSS entry representing a "composite" in the IBM Public Cloud catalog, which groups multiple related services or components
	COMPOSITE EntryType = "COMPOSITE"
	// GAAS is the EntryType for a OSS entry representing a service or component that is not part of the IBM Public Cloud catalog proper but that is nonetheless tracked through IBM Cloud OSS,
	// usihg the Global Ops as a Service (GaaS) framework.
	GAAS EntryType = "GAAS"
	// OTHEROSS is the EntryType for a OSS entry representing a service or component that is not part of the IBM Cloud catalog proper but that is nonetheless tracked through IBM Cloud OSS
	// not associated with GaaS
	OTHEROSS EntryType = "OTHEROSS"
	// IAMONLY is a special EntryType for a OSS entry that is a placeholder for a Catalog+IAM entry that is present only for IAM integration and not actually required for any OSS purposes
	// (e.g. not in ServiceNow, etc.).
	// Note that not every IAM entry has a corresponding OSS entry of type IAMONLY. But we need this placeholder in some cases to remember the fact that a given Catalog+IAM entry
	// does not represent another type of incomplete OSS registration.
	IAMONLY EntryType = "IAM_ONLY"
	// CONTENT is the EntryType for a OSS entry that represent a Content (software) offering in the catalog (e.g. CloudPaks)
	CONTENT EntryType = "CONTENT"
	// CONSULTING is the EntryType for a OSS entry that represent a Content offering in the catalog (e.g. CloudPaks)
	CONSULTING EntryType = "CONSULTING"
	// INTERNALSERVICE is the EntryType for a OSS entry that represent an Internal Service of the IBM Cloud Platform, with no client-facing aspect
	INTERNALSERVICE EntryType = "INTERNAL"
	// EntryTypeUnknown indicates that we were unable to determine a valid EntryType
	EntryTypeUnknown EntryType = "<unknown>"
)

// ShortString returns a short string representation of this Entrytype
func (e EntryType) ShortString() string {
	if e == PLATFORMCOMPONENT {
		return "PLATF_COMP"
	}
	return string(e)
}

// OperationalStatus represents the maturity level or other handling type for a service or component in the IBM Cloud (e.g. GA, Beta, etc.)
type OperationalStatus string

const (
	// GA is the OperationalStatus for a service or component that is GA in the IBM Cloud catalog
	GA OperationalStatus = "GA"
	// BETA is the OperationalStatus for a service or component that is Betain the IBM Cloud catalog
	BETA OperationalStatus = "BETA"
	// EXPERIMENTAL is the OperationalStatus for a service or component that is Experimentalin the IBM Cloud catalog
	EXPERIMENTAL OperationalStatus = "EXPERIMENTAL"
	// THIRDPARTY is the OperationalStatus for a third-party service in the IBM Cloud catalog
	THIRDPARTY OperationalStatus = "THIRDPARTY"
	// COMMUNITY is the OperationalStatus for a service or runtime that is Community-supported in the IBM Cloud catalog
	COMMUNITY OperationalStatus = "COMMUNITY"
	// DEPRECATED is the OperationalStatus for a service that has announced its deprecation the IBM Cloud catalog
	DEPRECATED OperationalStatus = "DEPRECATED"
	// SELECTAVAILABILITY is the OperationalStatus for a service or compomnent that is available to a selected set of Clients in the IBM Cloud catalog, but not generally available
	SELECTAVAILABILITY OperationalStatus = "SELECTAVAILABILITY"
	// RETIRED is the OperationalStatus for a service that is no longer available (past deprecation)
	RETIRED OperationalStatus = "RETIRED"
	// INTERNAL is the OperationalStatus for an internal component in IBM Cloud (which does not really have an OperationalStatus itself)
	INTERNAL OperationalStatus = "INTERNAL"
	// NOTREADY is the OperationalStatus for a service that is not yet ready for release (registration/OSS status may be incomplete)
	NOTREADY OperationalStatus = "NOTREADY"
	// OperationalStatusUnknown indicates that we were unable to determine a valid OperationalStatus
	OperationalStatusUnknown OperationalStatus = "<unknown>"
)

// ShortString returns a short string representation of this OperationalStatus
func (os OperationalStatus) ShortString() string {
	if os == SELECTAVAILABILITY {
		return "SELECTAVAIL"
	}
	return string(os)
}

// GHRepo represents the URL to a GitHub repository referenced in some OSS entry
type GHRepo string

// SlackChannel represents the name of a Slack channel referencef in some OSS entry
type SlackChannel string

// Tier2EscalationType represents the type of Tier2 escalation defined in ServiceNow Configuration Item
type Tier2EscalationType string

const (
	// SERVICENOW is the Tier2EscalationType indicating that tickets should be escalated to another ServiceNow Assignment Group
	SERVICENOW Tier2EscalationType = "SERVICENOW"
	// GITHUB is the Tier2EscalationType indicating that tickets should be escalated through a GitHub issue
	GITHUB Tier2EscalationType = "GITHUB"
	// RTC is the Tier2EscalationType indicating that tickets should be escalated through a RTC work item
	RTC Tier2EscalationType = "RTC"
	// OTHERESCALATION is the Tier2EscalationType indicating that tickets should be escalated through some other means (specified in the comments)
	OTHERESCALATION Tier2EscalationType = "OTHERESCALATION"
)

// ClientExperience represents the type of Client experience for ServiceNow Support Cases
type ClientExperience string

const (
	// ACSSUPPORTED is the ClientExperience for ServiceNow Cases handled at Tier1 by the ACS (formerly DSET) team
	ACSSUPPORTED ClientExperience = "ACS_SUPPORTED"
	// TRIBESUPPORTED is the ClientExperience for ServiceNow Cases handled at Tier1 by a specific Tribe
	TRIBESUPPORTED ClientExperience = "TRIBE_SUPPORTED"
)

// AvailabilityMonitoringInfo represents the information captured in ScorecardV1 for availabiliy monitoring
type AvailabilityMonitoringInfo struct {
	// TODO: Define AvailabilityMonitoringInfo
}

// RTCCategory represents the RTC category for Tier2 escalation of ServiceNow tickets that use RTC as the Tier2EscalationType
type RTCCategory string // historical

// BypassPassProductionReadiness represents possible values for the BypassPassProductionReadiness Operations attribute
// XXX Need to define possible values for  BypassPassProductionReadiness
type BypassPassProductionReadiness string

// BypassSupportCompliances represents possible values for the BypassSupportCompliances Support attribute
type BypassSupportCompliances string

// Constant definitions for BypassSupportCompliances
const (
	BypassSupportComplianceFailure BypassSupportCompliances = "failure"
	BypassSupportComplianceForceOK BypassSupportCompliances = "force_ok"
	BypassSupportComplianceOK      BypassSupportCompliances = "ok"
)

// ProductInfoNone is a constant used to denote that one attribute in ProductInfo (e.g. PartNumbers or ProductIDs) has no valid value
// (as opposed to being uninitialized)
const ProductInfoNone = "(none)"

// CleanEntryForCompare cleans the content of an OSSEntry so that it can be compared with other entries
// without flagging irrelevant differences
func (oss *OSSService) CleanEntryForCompare() {
	// nothing to clean
	return
}

// IsUpdatable returns true if this OSS record can be updated in the Catalog (i.e. it is a real standalone OSS record, not a placeholder for a native Catalog entry)
func (oss *OSSService) IsUpdatable() bool {
	return true
}

// SetTimes sets the Created and Updated times of this OSSEntry
func (oss *OSSService) SetTimes(created string, updated string) {
	// No-op: Created and Updated times not tracked for non-extended versions of OSS entries
}

// GetOSSTags returns the OSS tags associated with this OSSEntry
func (oss *OSSService) GetOSSTags() *osstags.TagSet {
	return &oss.GeneralInfo.OSSTags
}

// ResetForRMC resets the validation info (status and issues) that might become obsolete when an entry is being edited in RMC
func (oss *OSSService) ResetForRMC() {
	oss.GeneralInfo.OSSTags = oss.GeneralInfo.OSSTags.WithoutPureStatus().Copy()
}

// GetOSSOnboardingPhase returns the current OSS onboarding phase associated with this OSSEntry
func (oss *OSSService) GetOSSOnboardingPhase() OSSOnboardingPhase {
	return oss.GeneralInfo.OSSOnboardingPhase
}

// CheckSchemaVersion checks if the SchemaVersion in this entry is compatible with the current library version.
// If it is compatible, it returns nil
// if it is not compatible, it return a descriptive error AND UPDATES the SchemaVersion in the entry to mark the fact
// that it is not compatible
func (oss *OSSService) CheckSchemaVersion() error {
	return checkSchemaVersion(oss, &oss.SchemaVersion)
}

// OSSOnboardingPhase represents the status/phase of onboarding of a record in the RMC tool
type OSSOnboardingPhase string

const (
	// EDIT represents a record whose OSS tab is currently being edited in RMC - record is "dirty"
	EDIT OSSOnboardingPhase = "EDIT"
	// REVIEW represents a record whose OSS tab has been submitted for review/approval in RMC
	REVIEW OSSOnboardingPhase = "REVIEW"
	// APPROVED represents a record whose OSS tab has been approved in RMC, waiting for the owner to push to Production when ready
	APPROVED OSSOnboardingPhase = "APPROVED"
	// PRODUCTION represents a record that has been pushed to Production, or should be pushed to Production on the next run of osscatpublisher
	PRODUCTION OSSOnboardingPhase = "PRODUCTION"
	// INVALID represents a record that is not really used - can be deleted
	INVALID OSSOnboardingPhase = "INVALID"
)

// ProductIDSource represents the possible sources for the ProductInfo.ProductIDs attribute in the OSS records
type ProductIDSource string

// Possible values for ProductIDSource
const (
	ProductIDSourcePartNumbers   ProductIDSource = "Part Numbers"
	ProductIDSourceParent        ProductIDSource = "Inherited from Parent"
	ProductIDSourceCloudPlatform ProductIDSource = "Implicit for IBM Cloud Platform"
	ProductIDSourceManual        ProductIDSource = "Manual"
	ProductIDSourceCHName        ProductIDSource = "ClearingHouse Entry Name"
	ProductIDSourceCHCRN         ProductIDSource = "ClearingHouse Entry CRN"
	ProductIDSourceHistorical    ProductIDSource = "Historical (unknown)"
)

var _ OSSEntry = &OSSService{} // verify

// Domain represents the domain of the OSS service record
type Domain string

const (
	// COMMERCIAL is the Domain for the commercial OSS service record and is the main domain
	COMMERCIAL Domain = "COMMERCIAL"
	// USREGULATED is the Domain for the US regulated OSS service record
	USREGULATED Domain = "USREGULATED"
)
