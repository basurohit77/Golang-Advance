package servicenow

import (
	"strconv"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

// TestMergeCIToOSSServiceWatson tests the MergeCIToOSSService function for the Watson ServiceNow domain
func TestMergeCIToOSSServiceWatson(t *testing.T) {
	// Call the merge function with a populated SN CI object, empty OSS service object, and Watson ServiceNow domain:
	sn := createSNCIObject()
	service := new(ossrecord.OSSService)
	var err error
	service, err = MergeCIToOSSService(sn, service, WATSON)
	testhelper.AssertError(t, err)

	// Verify the OSS service values that are the same regardless of the domain passed into the merge function:
	checkCommonOSSServiceValuesAfterMerge(t, service, sn)

	// Verify that the OSS service values that are specific to the domain, or shared across multiple domains, are set correctly:
	testhelper.AssertEqual(t, "GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid(sn.SysID), service.GeneralInfo.ServiceNowSysid)
	testhelper.AssertEqual(t, "GeneralInfo.ServiceNowCIURL", sn.GeneralInfo.ServiceNowCIURL, service.GeneralInfo.ServiceNowCIURL)
	testhelper.AssertEqual(t, "Compliance.ServiceNowOnboarded", true, service.Compliance.ServiceNowOnboarded)
	testhelper.AssertEqual(t, "Operations.Slack", sn.Operations.Slack, service.Operations.Slack)
	testhelper.AssertEqual(t, "Operations.SpecialInstructions", sn.Operations.SpecialInstructions, service.Operations.SpecialInstructions)
	testhelper.AssertEqual(t, "Operations.Tier2EscalationType", sn.Operations.Tier2EscalationType, service.Operations.Tier2EscalationType)
	testhelper.AssertEqual(t, "Operations.Tier2RTC", sn.Operations.Tier2RTC, service.Operations.Tier2RTC)
	testhelper.AssertEqual(t, "Operations.Tier2Repository", sn.Operations.Tier2Repo, service.Operations.Tier2Repository)
	testhelper.AssertEqual(t, "Ownership.ServiceOffering", sn.Ownership.ServiceOffering, service.Ownership.ServiceOffering)
	testhelper.AssertEqual(t, "Ownership.TribeOwner", ossrecord.Person{W3ID: sn.Ownership.TribeOwner.W3ID, Name: sn.Ownership.TribeOwner.Name}, service.Ownership.TribeOwner)
	testhelper.AssertEqual(t, "Support.ClientExperience", sn.Support.ClientExperience, service.Support.ClientExperience)
	testhelper.AssertEqual(t, "Support.Slack", sn.Support.Slack, service.Support.Slack)
	testhelper.AssertEqual(t, "Support.SpecialInstructions", sn.Support.SpecialInstructions, service.Support.SpecialInstructions)
	testhelper.AssertEqual(t, "Support.Tier2EscalationType", sn.Support.Tier2EscalationType, service.Support.Tier2EscalationType)
	testhelper.AssertEqual(t, "Support.Tier2RTC", sn.Support.Tier2RTC, service.Support.Tier2RTC)
	testhelper.AssertEqual(t, "Support.Tier2Repository", sn.Support.Tier2Repo, service.Support.Tier2Repository)
	testhelper.AssertEqual(t, "ServiceNowInfo.CIEPageout", sn.ServiceNowInfo.CIEPageout, service.ServiceNowInfo.CIEPageout)
	testhelper.AssertEqual(t, "ServiceNowInfo.ERCAApprovalGroup", sn.ServiceNowInfo.ERCAApprovalGroup, service.ServiceNowInfo.ERCAApprovalGroup)
	testhelper.AssertEqual(t, "ServiceNowInfo.OperationsTier1AG", sn.ServiceNowInfo.OperationsTier1AG, service.ServiceNowInfo.OperationsTier1AG)
	testhelper.AssertEqual(t, "ServiceNowInfo.OperationsTier2AG", sn.ServiceNowInfo.OperationsTier2AG, service.ServiceNowInfo.OperationsTier2AG)
	testhelper.AssertEqual(t, "ServiceNowInfo.RCAApprovalGroup", sn.ServiceNowInfo.RCAApprovalGroup, service.ServiceNowInfo.RCAApprovalGroup)
	testhelper.AssertEqual(t, "ServiceNowInfo.SupportTier1AG", sn.ServiceNowInfo.SupportTier1AG, service.ServiceNowInfo.SupportTier1AG)
	testhelper.AssertEqual(t, "ServiceNowInfo.SupportTier2AG", sn.ServiceNowInfo.SupportTier2AG, service.ServiceNowInfo.SupportTier2AG)
	testhelper.AssertEqual(t, "ServiceNowInfo.TargetedCommunication", sn.ServiceNowInfo.TargetedCommunication, service.ServiceNowInfo.TargetedCommunication)

	// Verify that there are no overrides set since we called merge for the Watson ServiceNow domain:
	testhelper.AssertEqual(t, "Overrides", 0, len(service.Overrides))
}

// TestMergeCIToOSSServiceCloudfed tests the MergeCIToOSSService function for the Cloudfed ServiceNow domain
func TestMergeCIToOSSServiceCloudfed(t *testing.T) {
	// Call the merge function with a populated SN CI object, empty OSS service object, and Cloudfed ServiceNow domain:
	sn := createSNCIObject()
	service := new(ossrecord.OSSService)
	var err error
	service, err = MergeCIToOSSService(sn, service, CLOUDFED)
	testhelper.AssertError(t, err)

	// Verify the OSS service values that are the same regardless of the domain passed into the merge function:
	checkCommonOSSServiceValuesAfterMerge(t, service, sn)

	// The following OSS service values should all be set to empty values since these are commercial properties or common properties:
	testhelper.AssertEqual(t, "GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid(""), service.GeneralInfo.ServiceNowSysid)
	testhelper.AssertEqual(t, "GeneralInfo.ServiceNowCIURL", "", service.GeneralInfo.ServiceNowCIURL)
	testhelper.AssertEqual(t, "Operations.Slack", ossrecord.SlackChannel(""), service.Operations.Slack)
	testhelper.AssertEqual(t, "Operations.SpecialInstructions", "", service.Operations.SpecialInstructions)
	testhelper.AssertEqual(t, "Operations.Tier2EscalationType", ossrecord.Tier2EscalationType(""), service.Operations.Tier2EscalationType)
	testhelper.AssertEqual(t, "Operations.Tier2RTC", ossrecord.RTCCategory(""), service.Operations.Tier2RTC)
	testhelper.AssertEqual(t, "Operations.Tier2Repository", ossrecord.GHRepo(""), service.Operations.Tier2Repository)
	testhelper.AssertEqual(t, "Ownership.ServiceOffering", "", service.Ownership.ServiceOffering)
	testhelper.AssertEqual(t, "Ownership.TribeOwner", ossrecord.Person{W3ID: "", Name: ""}, service.Ownership.TribeOwner)
	testhelper.AssertEqual(t, "Support.ClientExperience", ossrecord.ClientExperience(""), service.Support.ClientExperience)
	testhelper.AssertEqual(t, "Support.Slack", ossrecord.SlackChannel(""), service.Support.Slack)
	testhelper.AssertEqual(t, "Support.SpecialInstructions", "", service.Support.SpecialInstructions)
	testhelper.AssertEqual(t, "Support.Tier2EscalationType", ossrecord.Tier2EscalationType(""), service.Support.Tier2EscalationType)
	testhelper.AssertEqual(t, "Support.Tier2RTC", ossrecord.RTCCategory(""), service.Support.Tier2RTC)
	testhelper.AssertEqual(t, "Support.Tier2Repository", ossrecord.GHRepo(""), service.Support.Tier2Repository)
	testhelper.AssertEqual(t, "ServiceNowInfo.CIEPageout", "", service.ServiceNowInfo.CIEPageout)
	testhelper.AssertEqual(t, "ServiceNowInfo.ERCAApprovalGroup", "", service.ServiceNowInfo.ERCAApprovalGroup)
	testhelper.AssertEqual(t, "ServiceNowInfo.OperationsTier1AG", "", service.ServiceNowInfo.OperationsTier1AG)
	testhelper.AssertEqual(t, "ServiceNowInfo.OperationsTier2AG", "", service.ServiceNowInfo.OperationsTier2AG)
	testhelper.AssertEqual(t, "ServiceNowInfo.RCAApprovalGroup", "", service.ServiceNowInfo.RCAApprovalGroup)
	testhelper.AssertEqual(t, "ServiceNowInfo.SupportTier1AG", "", service.ServiceNowInfo.SupportTier1AG)
	testhelper.AssertEqual(t, "ServiceNowInfo.SupportTier2AG", "", service.ServiceNowInfo.SupportTier2AG)
	testhelper.AssertEqual(t, "ServiceNowInfo.TargetedCommunication", "", service.ServiceNowInfo.TargetedCommunication)

	// Verify that the following OSS service values are set since they are US regulated properties:
	testhelper.AssertEqual(t, "Overrides", 1, len(service.Overrides))
	testhelper.AssertEqual(t, "Overrides.GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid(sn.SysID), service.Overrides[0].GeneralInfo.ServiceNowSysid)
	testhelper.AssertEqual(t, "Overrides.GeneralInfo.ServiceNowCIURL", sn.GeneralInfo.ServiceNowCIURL, service.Overrides[0].GeneralInfo.ServiceNowCIURL)
	testhelper.AssertEqual(t, "Overrides.GeneralInfo.Domain", ossrecord.USREGULATED, service.Overrides[0].GeneralInfo.Domain)
	testhelper.AssertEqual(t, "Overrides.Compliance.ServiceNowOnboarded", true, service.Overrides[0].Compliance.ServiceNowOnboarded)
}

// TestMergeCIToOSSServiceInvalidDomain tests the MergeCIToOSSService function for the an invalid ServiceNow domain
func TestMergeCIToOSSServiceInvalidDomain(t *testing.T) {
	sn := createSNCIObject()
	service := new(ossrecord.OSSService)
	snDomain := ServiceNowDomain("fakeDomain")
	service, err := MergeCIToOSSService(sn, service, snDomain)
	testhelper.AssertEqual(t, string(snDomain)+": error check", false, err == nil)
}

// TestGetCIFromOSSService tests the GetCIFromOSSService function
func TestGetCIFromOSSService(t *testing.T) {
	// Create a dummy OSS service:
	service := new(ossrecord.OSSService)
	service.ReferenceResourceName = ossrecord.CRNServiceName("crnServiceName1")
	service.ReferenceDisplayName = "displayName1"
	service.GeneralInfo.RMCNumber = "rmcNumber1"
	service.GeneralInfo.EntryType = ossrecord.EntryType(ossrecord.SERVICE)
	service.GeneralInfo.ClientFacing = true
	service.GeneralInfo.OSSDescription = "ossDescription1"
	service.GeneralInfo.OperationalStatus = ossrecord.OperationalStatus(ossrecord.BETA)
	service.Ownership.OfferingManager = ossrecord.Person{Name: "offeringManagerName1", W3ID: "offeringManagerW3ID1"}
	service.Ownership.TribeID = "tribeID1"
	service.Support.Manager = ossrecord.Person{Name: "supportManagerName1", W3ID: "supportManagerW3ID1"}
	service.Operations.Manager = ossrecord.Person{Name: "operationsManagerName1", W3ID: "operationsManagerW3ID1"}
	service.StatusPage.CategoryID = "statusPageCatagoryID1"
	service.StatusPage.Group = "statusPageGroup1"
	service.ServiceNowInfo.SupportNotApplicable = true
	service.ServiceNowInfo.OperationsNotApplicable = true

	// Call to get ServiceNow CI from the dummy OSS service:
	sn := GetCIFromOSSService(service)

	// Verify that the ServiceNow CI values are set correctly:
	testhelper.AssertEqual(t, "CRNServiceName", string(service.ReferenceResourceName), sn.CRNServiceName)
	testhelper.AssertEqual(t, "DisplayName", service.ReferenceDisplayName, sn.DisplayName)
	testhelper.AssertEqual(t, "GeneralInfo.RMCNumber", service.GeneralInfo.RMCNumber, sn.GeneralInfo.RMCNumber)
	testhelper.AssertEqual(t, "GeneralInfo.EntryType", service.GeneralInfo.EntryType, sn.GeneralInfo.EntryType)
	testhelper.AssertEqual(t, "GeneralInfo.ClientFacing", service.GeneralInfo.ClientFacing, sn.GeneralInfo.ClientFacing)
	testhelper.AssertEqual(t, "GeneralInfo.OSSDescription", service.GeneralInfo.OSSDescription, sn.GeneralInfo.OSSDescription)
	testhelper.AssertEqual(t, "GeneralInfo.OperationalStatus", service.GeneralInfo.OperationalStatus, sn.GeneralInfo.OperationalStatus)
	testhelper.AssertEqual(t, "Ownership.OfferingManager", service.Ownership.OfferingManager.W3ID, sn.Ownership.OfferingManager.W3ID)
	testhelper.AssertEqual(t, "Ownership.TribeID", string(service.Ownership.TribeID), sn.Ownership.TribeID)
	testhelper.AssertEqual(t, "Support.Manager", service.Support.Manager.W3ID, sn.Support.Manager.W3ID)
	testhelper.AssertEqual(t, "Operations.Manager", service.Operations.Manager.W3ID, sn.Operations.Manager.W3ID)
	testhelper.AssertEqual(t, "StatusPage.CategoryID", service.StatusPage.CategoryID, sn.StatusPage.CategoryID)
	testhelper.AssertEqual(t, "StatusPage.Group", service.StatusPage.Group, sn.StatusPage.Group)
	testhelper.AssertEqual(t, "ServiceNowInfo.SupportNotApplicable", strconv.FormatBool(service.ServiceNowInfo.SupportNotApplicable), sn.ServiceNowInfo.SupportNotApplicable)
	testhelper.AssertEqual(t, "ServiceNowInfo.OperationsNotApplicable", strconv.FormatBool(service.ServiceNowInfo.OperationsNotApplicable), sn.ServiceNowInfo.OperationsNotApplicable)
}

// createSNCIObject returns a ServiceNow configuration item object fully populated
func createSNCIObject() *ConfigurationItem {
	sn := new(ConfigurationItem)
	sn.SysID = "sysID1"
	sn.CRNServiceName = "crnServiceName1"
	sn.DisplayName = "displayName1"
	sn.GeneralInfo.RMCNumber = "rmcNumber1"
	sn.GeneralInfo.SOPNumber = "sopNumber1"
	sn.GeneralInfo.OperationalStatus = ossrecord.OperationalStatus(ossrecord.BETA)
	sn.GeneralInfo.ClientFacing = true
	sn.GeneralInfo.EntryType = ossrecord.EntryType(ossrecord.SERVICE)
	sn.GeneralInfo.FullCRN = "fullCRN1"
	sn.GeneralInfo.OSSDescription = "ossDescription1"
	sn.GeneralInfo.ServiceNowCIURL = "serviceNowInfo1"
	sn.Ownership.OfferingManager = ossrecord.Person{Name: "offeringManagerName1", W3ID: "offeringManagerW3ID1"}
	sn.Ownership.SegmentName = "segmentName1"
	sn.Ownership.TribeName = "tribeName1"
	sn.Ownership.TribeOwner = ossrecord.Person{Name: "tribeOwnerName1", W3ID: "tribeOwnerW3ID1"}
	sn.Ownership.TribeID = "tribeID1"
	sn.Support.Manager = ossrecord.Person{Name: "supportManagerName1", W3ID: "supportManagerW3ID1"}
	sn.Support.ClientExperience = "supportClientExperience1"
	sn.Support.SpecialInstructions = "supportSpecialInstructions1"
	sn.Support.Tier2EscalationType = "supportTier2EscalationType1"
	sn.Support.Slack = "supportSlack1"
	sn.Support.Tier2Repo = "supportTier2Rep"
	sn.Support.Tier2RTC = "supportTier2RTC"
	sn.Operations.Manager = ossrecord.Person{Name: "operationsManagerName1", W3ID: "operationsManagerW3ID1"}
	sn.Operations.SpecialInstructions = "operationsSpecialInstructions1"
	sn.Operations.Tier2EscalationType = "operationsTier2EscalationType1"
	sn.Operations.Slack = "operationsSlack1"
	sn.Operations.Tier2Repo = "operationsTier2Rep"
	sn.Operations.Tier2RTC = "operationsTier2RTC"
	sn.StatusPage.Group = "statusPageGroup1"
	sn.StatusPage.CategoryID = "statusPageCategoryID1"
	sn.StatusPage.CategoryIDMisspelled = "statusPageCategoryIDMisspelled1"
	sn.Compliance.SNSupportVerified = true
	sn.Compliance.SNOperationsVerified = true
	sn.ServiceNowInfo.SupportTier1AG = "supportTier1AG1"
	sn.ServiceNowInfo.SupportTier2AG = "supportTier2AG1"
	sn.ServiceNowInfo.OperationsTier1AG = "operationsTier1AG1"
	sn.ServiceNowInfo.OperationsTier2AG = "operationsTier2AG1"
	sn.ServiceNowInfo.RCAApprovalGroup = "rcaApprovalGroup1"
	sn.ServiceNowInfo.ERCAApprovalGroup = "ercaApprovalGroup1"
	sn.ServiceNowInfo.TargetedCommunication = "targetedCommunication1"
	sn.ServiceNowInfo.CIEPageout = "ciePageout1"
	sn.ServiceNowInfo.BackupContacts = "backupContacts1"
	sn.ServiceNowInfo.GBT30 = "gbt301"
	sn.ServiceNowInfo.District = "distract1"
	sn.ServiceNowInfo.SupportNotApplicable = "supportNotApplicable1"
	sn.ServiceNowInfo.OperationsNotApplicable = "operationsNotApplication1"
	return sn
}

// checkCommonOSSServiceValuesAfterMerge checks whether a subset of the service values are set correctly. The subset
// is the service values that are the same regardless of the domain. Note that all the values checked in here are
// empty or defaults because the merge operation only updates values that come from ServiceNow.
func checkCommonOSSServiceValuesAfterMerge(t *testing.T, service *ossrecord.OSSService, sn *ConfigurationItem) {
	var emptyArrayOfStrings []string

	testhelper.AssertEqual(t, "ReferenceResourceName", ossrecord.CRNServiceName(""), service.ReferenceResourceName)
	testhelper.AssertEqual(t, "ReferenceDisplayName", "", service.ReferenceDisplayName)
	testhelper.AssertEqual(t, "ReferenceCatalogID", ossrecord.CatalogID(""), service.ReferenceCatalogID)
	testhelper.AssertEqual(t, "ReferenceCatalogPath", "", service.ReferenceCatalogPath)

	testhelper.AssertEqual(t, "GeneralInfo.RMCNumber", "", service.GeneralInfo.RMCNumber)
	testhelper.AssertEqual(t, "GeneralInfo.OperationalStatus", ossrecord.OperationalStatus(""), service.GeneralInfo.OperationalStatus)
	testhelper.AssertEqual(t, "GeneralInfo.FutureOperationalStatus", ossrecord.OperationalStatus(""), service.GeneralInfo.FutureOperationalStatus)
	testhelper.AssertEqual(t, "GeneralInfo.OSSTags", osstags.TagSet(nil), service.GeneralInfo.OSSTags)
	testhelper.AssertEqual(t, "GeneralInfo.OSSOnboardingPhase", ossrecord.OSSOnboardingPhase(""), service.GeneralInfo.OSSOnboardingPhase)
	testhelper.AssertEqual(t, "GeneralInfo.OSSOnboardingApprover", ossrecord.Person{}, service.GeneralInfo.OSSOnboardingApprover)
	testhelper.AssertEqual(t, "GeneralInfo.OSSOnboardingApprovalDate", "", service.GeneralInfo.OSSOnboardingApprovalDate)
	testhelper.AssertEqual(t, "GeneralInfo.ClientFacing", false, service.GeneralInfo.ClientFacing)
	testhelper.AssertEqual(t, "GeneralInfo.EntryType", ossrecord.EntryType(""), service.GeneralInfo.EntryType)
	testhelper.AssertEqual(t, "GeneralInfo.OSSDescription", "", service.GeneralInfo.OSSDescription)
	testhelper.AssertEqual(t, "GeneralInfo.ParentResourceName", ossrecord.CRNServiceName(""), service.GeneralInfo.ParentResourceName)

	testhelper.AssertEqual(t, "Ownership.OfferingManager", ossrecord.Person{W3ID: "", Name: ""}, service.Ownership.OfferingManager)
	testhelper.AssertEqual(t, "Ownership.DevelopmentManager", ossrecord.Person{}, service.Ownership.DevelopmentManager)
	testhelper.AssertEqual(t, "Ownership.TechnicalContactDEPRECATED", ossrecord.Person{}, service.Ownership.TechnicalContactDEPRECATED)
	testhelper.AssertEqual(t, "Ownership.SegmentName", "", service.Ownership.SegmentName)
	testhelper.AssertEqual(t, "Ownership.SegmentID", ossrecord.SegmentID(""), service.Ownership.SegmentID)
	testhelper.AssertEqual(t, "Ownership.SegmentOwner", ossrecord.Person{}, service.Ownership.SegmentOwner)
	testhelper.AssertEqual(t, "Ownership.TribeName", "", service.Ownership.TribeName)
	testhelper.AssertEqual(t, "Ownership.TribeID", ossrecord.TribeID(""), service.Ownership.TribeID)
	testhelper.AssertEqual(t, "Ownership.MainRepository", ossrecord.GHRepo(""), service.Ownership.MainRepository)

	testhelper.AssertEqual(t, "Support.Manager", ossrecord.Person{W3ID: "", Name: ""}, service.Support.Manager)
	testhelper.AssertEqual(t, "Support.ThirdPartySupportURL", "", service.Support.ThirdPartySupportURL)
	testhelper.AssertEqual(t, "Support.ThirdPartyCaseProcess", "", service.Support.ThirdPartyCaseProcess)
	testhelper.AssertEqual(t, "Support.ThirdPartyContacts", "", service.Support.ThirdPartyContacts)
	testhelper.AssertEqual(t, "Support.ThirdPartyCountryLocations", "", service.Support.ThirdPartyCountryLocations)

	testhelper.AssertEqual(t, "Operations.Manager", ossrecord.Person{W3ID: "", Name: ""}, service.Operations.Manager)
	testhelper.AssertEqual(t, "Operations.TIPOnboarded", false, service.Operations.TIPOnboarded)
	testhelper.AssertEqual(t, "Operations.AVMEnabled", false, service.Operations.AVMEnabled)
	testhelper.AssertEqual(t, "Operations.BypassProductionReadiness", ossrecord.BypassPassProductionReadiness(""), service.Operations.BypassProductionReadiness)
	testhelper.AssertEqual(t, "Operations.RunbookEnabled", false, service.Operations.RunbookEnabled)
	testhelper.AssertEqual(t, "Operations.TOCAVMFocal", ossrecord.Person{}, service.Operations.TOCAVMFocal)
	testhelper.AssertEqual(t, "Operations.CIEDistList", "", service.Operations.CIEDistList)
	testhelper.AssertEqual(t, "Operations.EUAccessUSAMName", "", service.Operations.EUAccessUSAMName)
	var personListEntry []*ossrecord.PersonListEntry
	testhelper.AssertEqual(t, "Operations.AutomationIDs", personListEntry, service.Operations.AutomationIDs)

	testhelper.AssertEqual(t, "StatusPage.Group", "", service.StatusPage.Group)
	testhelper.AssertEqual(t, "StatusPage.CategoryID", "", service.StatusPage.CategoryID)
	testhelper.AssertEqual(t, "StatusPage.CategoryParent", ossrecord.CRNServiceName(""), service.StatusPage.CategoryParent)
	testhelper.AssertEqual(t, "StatusPage.CategoryIDMisspelled", "", service.StatusPage.CategoryIDMisspelled)

	testhelper.AssertEqual(t, "Compliance.ArchitectureFocal", ossrecord.Person{}, service.Compliance.ArchitectureFocal)
	testhelper.AssertEqual(t, "Compliance.BCDRFocal", ossrecord.Person{}, service.Compliance.BCDRFocal)
	testhelper.AssertEqual(t, "Compliance.SecurityFocal", ossrecord.Person{}, service.Compliance.SecurityFocal)
	testhelper.AssertEqual(t, "Compliance.ProvisionMonitors", ossrecord.AvailabilityMonitoringInfo{}, service.Compliance.ProvisionMonitors)
	testhelper.AssertEqual(t, "Compliance.ConsumptionMonitors", ossrecord.AvailabilityMonitoringInfo{}, service.Compliance.ConsumptionMonitors)
	testhelper.AssertEqual(t, "Compliance.PagerDutyURLs", emptyArrayOfStrings, service.Compliance.PagerDutyURLs)
	testhelper.AssertEqual(t, "Compliance.OnboardingContact", ossrecord.Person{}, service.Compliance.OnboardingContact)
	testhelper.AssertEqual(t, "Compliance.OnboardingIssueTrackerURL", "", service.Compliance.OnboardingIssueTrackerURL)
	testhelper.AssertEqual(t, "Compliance.BypassSupportCompliances", ossrecord.BypassSupportCompliances(""), service.Compliance.BypassSupportCompliances)
	testhelper.AssertEqual(t, "Compliance.CertificateManagerCRNs", emptyArrayOfStrings, service.Compliance.CertificateManagerCRNs)
	testhelper.AssertEqual(t, "Compliance.CompletedSkillTransferAndEnablement", false, service.Compliance.CompletedSkillTransferAndEnablement)

	testhelper.AssertEqual(t, "AdditionalContacts", "", service.AdditionalContacts)

	testhelper.AssertEqual(t, "CatalogInfo.Provider", ossrecord.Person{}, service.CatalogInfo.Provider)
	testhelper.AssertEqual(t, "CatalogInfo.ProviderContact", "", service.CatalogInfo.ProviderContact)
	testhelper.AssertEqual(t, "CatalogInfo.ProviderSupportEmail", "", service.CatalogInfo.ProviderSupportEmail)
	testhelper.AssertEqual(t, "CatalogInfo.ProviderPhone", "", service.CatalogInfo.ProviderPhone)
	testhelper.AssertEqual(t, "CatalogInfo.CategoryTags", "", service.CatalogInfo.CategoryTags)
	testhelper.AssertEqual(t, "CatalogInfo.CatalogClientFacing", false, service.CatalogInfo.CatalogClientFacing)
	testhelper.AssertEqual(t, "CatalogInfo.Locations", emptyArrayOfStrings, service.CatalogInfo.Locations)

	testhelper.AssertEqual(t, "ProductInfo.PartNumbers", emptyArrayOfStrings, service.ProductInfo.PartNumbers)
	testhelper.AssertEqual(t, "ProductInfo.PartNumbersRefreshed", "", service.ProductInfo.PartNumbersRefreshed)
	testhelper.AssertEqual(t, "ProductInfo.ProductIDs", emptyArrayOfStrings, service.ProductInfo.ProductIDs)
	testhelper.AssertEqual(t, "ProductInfo.ProductIDSource", ossrecord.ProductIDSource(""), service.ProductInfo.ProductIDSource)
	var clearingHouseReference []ossrecord.ClearingHouseReference
	testhelper.AssertEqual(t, "ProductInfo.ClearingHouseReferences", ossrecord.ClearingHouseReferences(clearingHouseReference), service.ProductInfo.ClearingHouseReferences)
	testhelper.AssertEqual(t, "ProductInfo.Taxonomy", ossrecord.Taxonomy{}, service.ProductInfo.Taxonomy)
	testhelper.AssertEqual(t, "ProductInfo.Division", "", service.ProductInfo.Division)
	testhelper.AssertEqual(t, "ProductInfo.OSSUID", "", service.ProductInfo.OSSUID)

	var emptyDependencyArray []*ossrecord.Dependency
	testhelper.AssertEqual(t, "DependencyInfo.OutboundDependencies", ossrecord.Dependencies(emptyDependencyArray), service.DependencyInfo.OutboundDependencies)
	testhelper.AssertEqual(t, "DependencyInfo.InboundDependencies", ossrecord.Dependencies(emptyDependencyArray), service.DependencyInfo.InboundDependencies)

	var emptyMetricsArray []*ossrecord.Metric
	testhelper.AssertEqual(t, "MonitoringInfo.Metrics", emptyMetricsArray, service.MonitoringInfo.Metrics)
}

// TestServiceNowDomainToDomain tests the serviceNowDomainToDomain function
func TestServiceNowDomainToDomain(t *testing.T) {
	snDomain := WATSON
	domain, err := serviceNowDomainToDomain(snDomain)
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, string(snDomain)+": domain check", ossrecord.COMMERCIAL, domain)

	snDomain = CLOUDFED
	domain, err = serviceNowDomainToDomain(snDomain)
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, string(snDomain)+": domain check", ossrecord.USREGULATED, domain)

	snDomain = ServiceNowDomain("fakeDomain")
	domain, err = serviceNowDomainToDomain(snDomain)
	testhelper.AssertEqual(t, string(snDomain)+": error check", false, err == nil)
}
