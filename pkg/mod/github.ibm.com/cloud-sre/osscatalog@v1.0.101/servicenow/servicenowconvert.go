package servicenow

import (
	"fmt"
	"strconv"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// ParseEntryType converts a ServiceNow EntryType string into the common ossrecord.EntryType type
func ParseEntryType(input string) (ossrecord.EntryType, error) {
	return ossrecord.ParseEntryType(input)
}

var validOperationalStatuses = map[string]ossrecord.OperationalStatus{
	"3rd Party":          ossrecord.THIRDPARTY,
	"3RD PARTY":          ossrecord.THIRDPARTY,
	"BETA IBM":           ossrecord.BETA,
	"Community":          ossrecord.COMMUNITY,
	"Deprecated":         ossrecord.DEPRECATED,
	"Experimental":       ossrecord.EXPERIMENTAL,
	"GA IBM":             ossrecord.GA,
	"Inactive":           ossrecord.NOTREADY,
	"INACTIVE":           ossrecord.NOTREADY,
	"Internal":           ossrecord.INTERNAL,
	"Retired":            ossrecord.RETIRED,
	"SelectAvailability": ossrecord.SELECTAVAILABILITY,
}

// ParseOperationalStatus converts a ServiceNow EntryType string into the common ossrecord.EntryType type
func ParseOperationalStatus(input string) (ossrecord.OperationalStatus, error) {
	if input == "" {
		return "", nil
	}
	if result, ok := validOperationalStatuses[input]; ok {
		return result, nil
	}
	return ossrecord.ParseOperationalStatus(input)
}

var validTier2EscalationTypes = map[string]ossrecord.Tier2EscalationType{
	"ServiceNow": ossrecord.SERVICENOW,
	"GitHub":     ossrecord.GITHUB,
	"RTC":        ossrecord.RTC,
	"Other":      ossrecord.OTHERESCALATION,
	"OTHER":      ossrecord.OTHERESCALATION,
}

// ParseTier2EscalationType converts a ServiceNow Tier2EscalationType string into the common ossrecord.Tier2EscalationType type
func ParseTier2EscalationType(input string) (ossrecord.Tier2EscalationType, error) {
	if input == "" {
		return "", nil
	}
	if result, ok := validTier2EscalationTypes[input]; ok {
		return result, nil
	}
	return ossrecord.ParseTier2EscalationType(input)
}

var validClientExperiences = map[string]ossrecord.ClientExperience{
	"Bluemix supported":      ossrecord.ACSSUPPORTED,
	"IBM Cloud Supported":    ossrecord.ACSSUPPORTED,
	"IBM CLOUD SUPPORTED":    ossrecord.ACSSUPPORTED,
	"DSET_SUPPORTED":         ossrecord.ACSSUPPORTED,
	"Service-team supported": ossrecord.TRIBESUPPORTED,
	"SERVICE-TEAM SUPPORTED": ossrecord.TRIBESUPPORTED,
}

// ParseClientExperience converts a ServiceNow ClientExperience string into the common ossrecord.ClientExperience type
func ParseClientExperience(input string) (ossrecord.ClientExperience, error) {
	if input == "" {
		return "", nil
	}
	if result, ok := validClientExperiences[input]; ok {
		return result, nil
	}
	return ossrecord.ParseClientExperience(input)
}

// IsRetired returns true if this ServiceNow info record is retired.
func (sn *ConfigurationItem) IsRetired() bool {
	return sn.GeneralInfo.OperationalStatus == ossrecord.RETIRED
}

// serviceNowDomainToDomain returns the OSS record domain for the provided ServiceNow domain
func serviceNowDomainToDomain(snDomain ServiceNowDomain) (domain ossrecord.Domain, err error) {
	if snDomain == WATSON {
		return ossrecord.COMMERCIAL, nil
	} else if snDomain == CLOUDFED {
		return ossrecord.USREGULATED, nil
	}
	return domain, fmt.Errorf("serviceNowDomainToDomain(%s): returned with unexpected domain %s", snDomain, snDomain)
}

// MergeCIToOSSService update OSSService with servicenow CI values
func MergeCIToOSSService(sn *ConfigurationItem, service *ossrecord.OSSService, snDomain ServiceNowDomain) (*ossrecord.OSSService, error) {
	// Convert ServiceNow domain to an OSS record domain:
	domain, err := serviceNowDomainToDomain(snDomain)
	if err != nil {
		return nil, err
	}

	// Determine if the ServiceNow CI values need to be applied to OSS service or to an override on the OSS service:
	overrideIndex, err := catalog.GetOrCreateOverride(service, domain)
	if err != nil {
		return nil, err
	}
	if overrideIndex == -1 {
		// Apply to the OSS service directly:
		service.GeneralInfo.ServiceNowSysid = ossrecord.ServiceNowSysid(sn.SysID)
		service.GeneralInfo.ServiceNowCIURL = sn.GeneralInfo.ServiceNowCIURL
		service.Compliance.ServiceNowOnboarded = sn.GeneralInfo.ServiceNowCIURL != "" && service.GeneralInfo.OperationalStatus != ossrecord.RETIRED

		// Store values that come from ServiceNow into the OSS Service record:
		service.Operations.Slack = sn.Operations.Slack
		service.Operations.SpecialInstructions = sn.Operations.SpecialInstructions
		service.Operations.Tier2EscalationType = sn.Operations.Tier2EscalationType
		service.Operations.Tier2RTC = sn.Operations.Tier2RTC
		service.Operations.Tier2Repository = sn.Operations.Tier2Repo
		service.Ownership.ServiceOffering = sn.Ownership.ServiceOffering
		service.Ownership.TribeOwner = sn.Ownership.TribeOwner
		service.Support.ClientExperience = sn.Support.ClientExperience
		service.Support.Slack = sn.Support.Slack
		service.Support.SpecialInstructions = sn.Support.SpecialInstructions
		service.Support.Tier2EscalationType = sn.Support.Tier2EscalationType
		service.Support.Tier2RTC = sn.Support.Tier2RTC
		service.Support.Tier2Repository = sn.Support.Tier2Repo
		service.ServiceNowInfo.CIEPageout = sn.ServiceNowInfo.CIEPageout
		service.ServiceNowInfo.ERCAApprovalGroup = sn.ServiceNowInfo.ERCAApprovalGroup
		service.ServiceNowInfo.OperationsTier1AG = sn.ServiceNowInfo.OperationsTier1AG
		service.ServiceNowInfo.OperationsTier2AG = sn.ServiceNowInfo.OperationsTier2AG
		service.ServiceNowInfo.RCAApprovalGroup = sn.ServiceNowInfo.RCAApprovalGroup
		service.ServiceNowInfo.SupportTier1AG = sn.ServiceNowInfo.SupportTier1AG
		service.ServiceNowInfo.SupportTier2AG = sn.ServiceNowInfo.SupportTier2AG
		service.ServiceNowInfo.TargetedCommunication = sn.ServiceNowInfo.TargetedCommunication
	} else {
		// Apply to appropriate override on the OSS service:
		service.Overrides[overrideIndex].GeneralInfo.ServiceNowSysid = ossrecord.ServiceNowSysid(sn.SysID)
		service.Overrides[overrideIndex].GeneralInfo.ServiceNowCIURL = sn.GeneralInfo.ServiceNowCIURL
		service.Overrides[overrideIndex].Compliance.ServiceNowOnboarded = sn.GeneralInfo.ServiceNowCIURL != "" && service.GeneralInfo.OperationalStatus != ossrecord.RETIRED
	}
	return service, nil
}

// GetCIFromOSSService return servicenow CI from given ossservice. Used to create or update a CI in servicenow and
// only include values for properties needed in ServiceNow whose source of truth is the OSS Service record. Note
// that this function does not take a domain because there are no overriden properties set in the returned servicenow
// CI.
func GetCIFromOSSService(service *ossrecord.OSSService) (sn *ConfigurationItem) {
	sn = new(ConfigurationItem)
	sn.CRNServiceName = string(service.ReferenceResourceName)
	sn.DisplayName = service.ReferenceDisplayName

	sn.GeneralInfo.RMCNumber = service.GeneralInfo.RMCNumber
	sn.GeneralInfo.EntryType = service.GeneralInfo.EntryType
	sn.GeneralInfo.ClientFacing = service.GeneralInfo.ClientFacing
	sn.GeneralInfo.OSSDescription = service.GeneralInfo.OSSDescription
	sn.GeneralInfo.OperationalStatus = service.GeneralInfo.OperationalStatus

	sn.Ownership.OfferingManager.W3ID = service.Ownership.OfferingManager.W3ID
	sn.Ownership.TribeID = string(service.Ownership.TribeID)

	sn.Support.Manager.W3ID = service.Support.Manager.W3ID

	sn.Operations.Manager.W3ID = service.Operations.Manager.W3ID

	sn.StatusPage.CategoryID = service.StatusPage.CategoryID
	sn.StatusPage.Group = service.StatusPage.Group

	sn.ServiceNowInfo.SupportNotApplicable = strconv.FormatBool(service.ServiceNowInfo.SupportNotApplicable)
	sn.ServiceNowInfo.OperationsNotApplicable = strconv.FormatBool(service.ServiceNowInfo.OperationsNotApplicable)

	return sn
}
