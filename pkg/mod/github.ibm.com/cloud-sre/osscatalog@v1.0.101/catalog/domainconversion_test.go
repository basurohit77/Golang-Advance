package catalog

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

// TestCreateNewOSSServiceExtendedWithService tests the createNewOSSServiceExtendedWithService function
func TestCreateNewOSSServiceExtendedWithService(t *testing.T) {
	// Create OSS service extended record that will be used for testing:
	ossrec := createTestOSSServiceExtendedObject(true)

	// Get the OSS service object for the US regulated domain:
	var entry ossrecord.OSSEntry
	entry = &ossrec.OSSService
	ossEntryForDomain, err := getOSSEntryForDomain(&entry, IncludeServicesDomainUSRegulated)
	updatedService := ossEntryForDomain.(*ossrecord.OSSService)
	testhelper.AssertError(t, err)

	// Call createNewOSSServiceExtendedWithService and verify result:
	serviceExtended := createNewOSSServiceExtendedWithService(&ossrec, updatedService)
	compareOutput := compare.Output{}
	compare.DeepCompare("original", ossrec, "new", serviceExtended, &compareOutput)
	testhelper.AssertEqual(t, "OSS Service Extended: orginal vs new diff", true, len(compareOutput.GetDiffs()) > 0) // different OSS service extended
	compareOutput = compare.Output{}
	compare.DeepCompare("original", ossrec.OSSService, "new", serviceExtended.OSSService, &compareOutput)
	testhelper.AssertEqual(t, "OSS Service: orginal vs new diff", true, len(compareOutput.GetDiffs()) > 0) // different OSS services
	testhelper.AssertEqual(t, "service", *updatedService, serviceExtended.OSSService)                      // but correct service under the new OSS service extended
}

// TestGetDomainSpecificOSSEntries tests the getDomainSpecificOSSEntries function
func TestGetDomainSpecificOSSEntries(t *testing.T) {
	// Create OSS service extended record that will be used for testing:
	ossrec := createTestOSSServiceExtendedObject(true)

	// Test passing in OSS service object:
	var entry ossrecord.OSSEntry
	entry = &ossrec.OSSService
	domainSpecificEntries := getDomainSpecificOSSEntries(&entry, IncludeServicesDomainCommercial|IncludeServicesDomainUSRegulated)
	testhelper.AssertEqual(t, "number of entries", 2, len(domainSpecificEntries))

	domainSpecificEntry1 := domainSpecificEntries[0]
	actualServiceUSRegulated, ok := domainSpecificEntry1.(*ossrecord.OSSService)
	testhelper.AssertEqual(t, "cast from OSSEntry to OSSService", true, ok)
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.USREGULATED, actualServiceUSRegulated.GeneralInfo.Domain)

	domainSpecificEntry2 := domainSpecificEntries[1]
	actualServiceCommercial, ok := domainSpecificEntry2.(*ossrecord.OSSService)
	testhelper.AssertEqual(t, "cast from OSSEntry to OSSService", true, ok)
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.COMMERCIAL, actualServiceCommercial.GeneralInfo.Domain)

	// Test passing in OSS extended service object:
	entry = &ossrec
	domainSpecificEntries = getDomainSpecificOSSEntries(&entry, IncludeServicesDomainCommercial|IncludeServicesDomainUSRegulated)
	testhelper.AssertEqual(t, "number of entries", 2, len(domainSpecificEntries))

	domainSpecificEntry1 = domainSpecificEntries[0]
	actualServiceExtUSRegulated, ok := domainSpecificEntry1.(*ossrecordextended.OSSServiceExtended)
	testhelper.AssertEqual(t, "cast from OSSEntry to OSSServiceExtended", true, ok)
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.USREGULATED, actualServiceExtUSRegulated.OSSService.GeneralInfo.Domain)

	domainSpecificEntry2 = domainSpecificEntries[1]
	actualServiceExtCommercial, ok := domainSpecificEntry2.(*ossrecordextended.OSSServiceExtended)
	testhelper.AssertEqual(t, "cast from OSSEntry to OSSServiceExtended", true, ok)
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.COMMERCIAL, actualServiceExtCommercial.OSSService.GeneralInfo.Domain)

	// Test that an OSS entry type other than OSSService or OSSServiceExtended just returns the provided object:
	segment := &ossrecord.OSSSegment{
		SegmentID:     "fakeSegmentID",
		DisplayName:   "fakeSegmentDisplayName",
		SchemaVersion: ossrecord.OSSCurrentSchema,
	}
	entry = segment
	domainSpecificEntries = getDomainSpecificOSSEntries(&entry, IncludeServicesDomainCommercial|IncludeServicesDomainUSRegulated)
	testhelper.AssertEqual(t, "number of entries", 1, len(domainSpecificEntries))
	testhelper.AssertEqual(t, "compare returned segment", segment, domainSpecificEntries[0])

	// Create OSS service extended record that will be used for testing:
	ossrec = createTestOSSServiceExtendedObject(false)

	// Test various include options when we have an OSS service but no overrides:
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeNone, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeNone|IncludeServicesDomainOverrides, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServices, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainOverrides, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServicesDomainCommercial, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServicesDomainCommercial|IncludeServicesDomainOverrides, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainCommercial, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainCommercial|IncludeServicesDomainOverrides, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServicesDomainOverrides, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated|IncludeServicesDomainCommercial, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated|IncludeServicesDomainCommercial|IncludeServicesDomainOverrides, false)
	callGetDomainSpecificOSSEntriesAndVerifyNoRecordsAreReturned(t, &ossrec.OSSService, IncludeServicesDomainUSRegulated)
	callGetDomainSpecificOSSEntriesAndVerifyNoRecordsAreReturned(t, &ossrec.OSSService, IncludeServicesDomainUSRegulated|IncludeServicesDomainOverrides)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServicesDomainUSRegulated|IncludeServicesDomainCommercial, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServicesDomainUSRegulated|IncludeServicesDomainCommercial|IncludeServicesDomainOverrides, false)
	callGetDomainSpecificOSSEntriesAndVerifyNoRecordsAreReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated)
	callGetDomainSpecificOSSEntriesAndVerifyNoRecordsAreReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated|IncludeServicesDomainOverrides)

	// Add US regulated override to the OSS service record:
	ossrec = createTestOSSServiceExtendedObject(true)

	// Test various include options when we have an OSS service with the US regulated override:
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeNone, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeNone|IncludeServicesDomainOverrides, true)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServices, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainOverrides, true)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServicesDomainCommercial, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServicesDomainCommercial|IncludeServicesDomainOverrides, true)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainCommercial, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainCommercial|IncludeServicesDomainOverrides, true)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t, &ossrec.OSSService, IncludeServicesDomainOverrides, true)
	callGetDomainSpecificOSSEntriesAndVerifyCommercialAndUSRegulatedAreReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated|IncludeServicesDomainCommercial, false)
	callGetDomainSpecificOSSEntriesAndVerifyCommercialAndUSRegulatedAreReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated|IncludeServicesDomainCommercial|IncludeServicesDomainOverrides, true)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyUSRegulatedIsReturned(t, &ossrec.OSSService, IncludeServicesDomainUSRegulated, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyUSRegulatedIsReturned(t, &ossrec.OSSService, IncludeServicesDomainUSRegulated|IncludeServicesDomainOverrides, true)
	callGetDomainSpecificOSSEntriesAndVerifyCommercialAndUSRegulatedAreReturned(t, &ossrec.OSSService, IncludeServicesDomainUSRegulated|IncludeServicesDomainCommercial, false)
	callGetDomainSpecificOSSEntriesAndVerifyCommercialAndUSRegulatedAreReturned(t, &ossrec.OSSService, IncludeServicesDomainUSRegulated|IncludeServicesDomainCommercial|IncludeServicesDomainOverrides, true)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyUSRegulatedIsReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated, false)
	callGetDomainSpecificOSSEntriesAndVerifyOnlyUSRegulatedIsReturned(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated|IncludeServicesDomainOverrides, true)

	// Test case where domain is not set on OSS service (i.e. service has not been saved since the domain attribute was added):
	ossrec.OSSService.GeneralInfo.Domain = ""
	entry = &ossrec.OSSService
	services := getDomainSpecificOSSEntries(&entry, IncludeNone)
	testhelper.AssertEqual(t, "number of services", 1, len(services))
	actualService := *(services[0].(*ossrecord.OSSService))
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.Domain(""), actualService.GeneralInfo.Domain)
	verifyOverridesOnOSSService(t, actualService, false)
	verifyCommercialOSSService(t, ossrec.OSSService, actualService, false)
}

// TestGetOSSEntryForDomain tests the getOSSEntryForDomain function
func TestGetOSSEntryForDomain(t *testing.T) {
	// Create OSS service extended record that will be used for testing:
	ossrec := createTestOSSServiceExtendedObject(true)

	// Test passing in OSS service object:
	var entry ossrecord.OSSEntry
	entry = &ossrec.OSSService
	domainSpecificEnty, err := getOSSEntryForDomain(&entry, IncludeServicesDomainCommercial|IncludeServicesDomainOverrides)
	testhelper.AssertError(t, err)
	actualService, ok := domainSpecificEnty.(*ossrecord.OSSService)
	testhelper.AssertEqual(t, "cast from OSSEntry to OSSService", true, ok)
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.COMMERCIAL, actualService.GeneralInfo.Domain)

	// Test passing in OSS extended service object:
	entry = &ossrec
	domainSpecificEnty, err = getOSSEntryForDomain(&entry, IncludeServicesDomainCommercial|IncludeServicesDomainOverrides)
	testhelper.AssertError(t, err)
	actualServiceExt, ok := domainSpecificEnty.(*ossrecordextended.OSSServiceExtended)
	testhelper.AssertEqual(t, "cast from OSSEntry to OSSServiceExtended", true, ok)
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.COMMERCIAL, actualServiceExt.GeneralInfo.Domain)

	// Test that an OSS entry type other than OSSService or OSSServiceExtended just returns the provided object:
	segment := &ossrecord.OSSSegment{
		SegmentID:     "fakeSegmentID",
		DisplayName:   "fakeSegmentDisplayName",
		SchemaVersion: ossrecord.OSSCurrentSchema,
	}
	entry = segment
	domainSpecificEnty, err = getOSSEntryForDomain(&entry, IncludeServicesDomainCommercial|IncludeServicesDomainUSRegulated)
	testhelper.AssertEqual(t, "compare returned segment", segment, domainSpecificEnty)

	// Create OSS service extended record that will be used for testing:
	ossrec = createTestOSSServiceExtendedObject(false)

	// Want commercial OSS service record (no conversation necessary as commercial):
	convertedService := callGetOSSEntryForDomainAndReturnService(t, &ossrec.OSSService, IncludeServices, false)
	testhelper.AssertEqual(t, "OSSService record", ossrec.OSSService, convertedService)
	testhelper.AssertEqual(t, "ossdata.Overrides", true, convertedService.Overrides == nil)
	verifyCommercialOSSService(t, ossrec.OSSService, convertedService, false)

	// Want commercial OSS service record with overrides (no conversation necessary as commercial and there are no overrides):
	convertedService = callGetOSSEntryForDomainAndReturnService(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainOverrides, false)
	testhelper.AssertEqual(t, "OSSService record", ossrec.OSSService, convertedService)
	testhelper.AssertEqual(t, "ossdata.Overrides", true, convertedService.Overrides == nil)
	verifyCommercialOSSService(t, ossrec.OSSService, convertedService, false)

	// Want US regulated OSS service record (conversation fails because there are no overrides in the main OSS service record):
	convertedService = callGetOSSEntryForDomainAndReturnService(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated, true)

	// Want US regulated OSS service record with overrides (conversation fails because there are no overrides in the main OSS service record):
	convertedService = callGetOSSEntryForDomainAndReturnService(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated|IncludeServicesDomainOverrides, true)

	// Add US regulated overrides to the main OSS service record:
	ossrec = createTestOSSServiceExtendedObject(true)

	// Want commercial entry with overrides (no converation necessary):
	convertedService = callGetOSSEntryForDomainAndReturnService(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainOverrides, false)
	testhelper.AssertEqual(t, "OSSService record", ossrec.OSSService, convertedService)
	testhelper.AssertEqual(t, "ossdata.Overrides", false, convertedService.Overrides == nil)
	verifyCommercialOSSService(t, ossrec.OSSService, convertedService, true)

	// Want commercial entry without overrides (converstion necessary to remove overrides):
	convertedService = callGetOSSEntryForDomainAndReturnService(t, &ossrec.OSSService, IncludeServices, false)
	testhelper.AssertEqual(t, "ossdata.Overrides", true, convertedService.Overrides == nil)
	verifyCommercialOSSService(t, ossrec.OSSService, convertedService, false)

	// Want US regulated OSS service record with overrides (conversation is successful this time because there are overrides in the main OSS service record):
	convertedService = callGetOSSEntryForDomainAndReturnService(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated|IncludeServicesDomainOverrides, false)
	testhelper.AssertEqual(t, "ossdata.Overrides", false, convertedService.Overrides == nil)
	verifyUSRegulatedOSSService(t, ossrec.OSSService, convertedService)

	// Want US regulated OSS service record without overrides (conversation is successful this time because there are overrides in the main OSS service record):
	convertedService = callGetOSSEntryForDomainAndReturnService(t, &ossrec.OSSService, IncludeServices|IncludeServicesDomainUSRegulated, false)
	testhelper.AssertEqual(t, "ossdata.Overrides", true, convertedService.Overrides == nil)
	verifyUSRegulatedOSSService(t, ossrec.OSSService, convertedService)

	// Test case where domain is not set on OSS service (i.e. service has not been saved since the domain attribute was added):
	ossrec.OSSService.GeneralInfo.Domain = ""
	convertedService = callGetOSSEntryForDomainAndReturnService(t, &ossrec.OSSService, IncludeServices, false)
	testhelper.AssertEqual(t, "ossdata.Overrides", true, convertedService.Overrides == nil)
	verifyCommercialOSSService(t, ossrec.OSSService, convertedService, false)
}

// TestGetOSSServiceFromOSSEntry tests the getOSSServiceFromOSSEntry function
func TestGetOSSServiceFromOSSEntry(t *testing.T) {
	// Create OSS service extended record that will be used for testing:
	ossrec := createTestOSSServiceExtendedObject(false)

	// Test that an OSSServiceExtended causes it's OSSService to be returned:
	var entry ossrecord.OSSEntry
	entry = &ossrec
	serviceExtended, service := getOSSServiceFromOSSEntry(&entry)
	testhelper.AssertEqual(t, "OSSService pointer", true, service != nil)
	testhelper.AssertEqual(t, "OSSServiceExtended record", &ossrec, serviceExtended)
	testhelper.AssertEqual(t, "OSSService record", ossrec.OSSService, *service)

	// Test that the same OSSService is returned:
	entry = &ossrec.OSSService
	serviceExtended, service = getOSSServiceFromOSSEntry(&entry)
	testhelper.AssertEqual(t, "OSSService pointer", true, service != nil)
	testhelper.AssertEqual(t, "OSSServiceExtended record", (*ossrecordextended.OSSServiceExtended)(nil), serviceExtended)
	testhelper.AssertEqual(t, "OSSService record", ossrec.OSSService, *service)

	// Test that an OSS entry type other than OSSService or OSSServiceExtended does not return an OSSService object:
	segment := &ossrecord.OSSSegment{
		SegmentID:     "fakeSegmentID",
		DisplayName:   "fakeSegmentDisplayName",
		SchemaVersion: ossrecord.OSSCurrentSchema,
	}
	entry = segment
	_, service = getOSSServiceFromOSSEntry(&entry)
	testhelper.AssertEqual(t, "OSSService pointer", true, service == nil)
}

// TestGetOSSServiceRecordForDomain tests the getOSSServiceRecordForDomain function
func TestGetOSSServiceRecordForDomain(t *testing.T) {
	// Create OSS service extended record that will be used for testing:
	ossrec := createTestOSSServiceExtendedObject(true)

	service, err := getOSSServiceRecordForDomain(ossrec.OSSService, ossrecord.COMMERCIAL)
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "OSSService record", ossrec.OSSService, service)

	service, err = getOSSServiceRecordForDomain(ossrec.OSSService, ossrecord.USREGULATED)
	testhelper.AssertError(t, err)
	verifyUSRegulatedOSSService(t, ossrec.OSSService, service)

	service, err = getOSSServiceRecordForDomain(ossrec.OSSService, ossrecord.Domain("fakeDomain"))
	testhelper.AssertEqual(t, "expect error due to fake domain", false, err == nil)
}

// TestGetOrCreateOverride tests the GetOrCreateOverride function
func TestGetOrCreateOverride(t *testing.T) {
	// Create OSS service extended record that will be used for testing:
	ossrec := createTestOSSServiceExtendedObject(false)
	service := ossrec.OSSService

	// Service does not have any existing overrides, expet one to be added:
	overrideIndex, err := GetOrCreateOverride(&service, ossrecord.USREGULATED)
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "overrideIndex", 0, overrideIndex)
	testhelper.AssertEqual(t, "override.GeneralInfo.OSSTags", 0, len(service.Overrides[overrideIndex].GeneralInfo.OSSTags))
	testhelper.AssertEqual(t, "override.GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid(""), service.Overrides[overrideIndex].GeneralInfo.ServiceNowSysid)
	testhelper.AssertEqual(t, "override.GeneralInfo.ServiceNowCIURL", "", service.Overrides[overrideIndex].GeneralInfo.ServiceNowCIURL)
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.USREGULATED, service.Overrides[overrideIndex].GeneralInfo.Domain)
	testhelper.AssertEqual(t, "override.Compliance.ServiceNowOnboarded", false, service.Overrides[overrideIndex].Compliance.ServiceNowOnboarded)

	// Update added override:
	ossrec = createTestOSSServiceExtendedObject(true)
	service = ossrec.OSSService

	// Service now has an existing override, check that the values of the override are as expected:
	overrideIndex, err = GetOrCreateOverride(&service, ossrecord.USREGULATED)
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "overrideIndex", 0, overrideIndex)
	testhelper.AssertEqual(t, "override.GeneralInfo.OSSTags", 1, len(service.Overrides[overrideIndex].GeneralInfo.OSSTags))
	testhelper.AssertEqual(t, "override.GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid("fakeSNSysId"), service.Overrides[overrideIndex].GeneralInfo.ServiceNowSysid)
	testhelper.AssertEqual(t, "override.GeneralInfo.ServiceNowCIURL", "fakeSNCIURL", service.Overrides[overrideIndex].GeneralInfo.ServiceNowCIURL)
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.USREGULATED, service.Overrides[overrideIndex].GeneralInfo.Domain)
	testhelper.AssertEqual(t, "override.Compliance.ServiceNowOnboarded", true, service.Overrides[overrideIndex].Compliance.ServiceNowOnboarded)

	// Commercial domain - expect -1 return:
	overrideIndex, err = GetOrCreateOverride(&service, ossrecord.COMMERCIAL)
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "overrideIndex", -1, overrideIndex)

	// Bad domain - expect error:
	overrideIndex, err = GetOrCreateOverride(&service, ossrecord.Domain("fakeDomain"))
	testhelper.AssertEqual(t, "error", false, err == nil)
	testhelper.AssertEqual(t, "overrideIndex", -2, overrideIndex)
}

// TestIsValidDomain tests the isValidDomain function
func TestIsValidDomain(t *testing.T) {
	domain := ossrecord.COMMERCIAL
	isValid := isValidDomain(domain)
	testhelper.AssertEqual(t, string(domain), true, isValid)

	domain = ossrecord.USREGULATED
	isValid = isValidDomain(domain)
	testhelper.AssertEqual(t, string(domain), true, isValid)

	domain = ossrecord.Domain("fakeDomain")
	isValid = isValidDomain(domain)
	testhelper.AssertEqual(t, string(domain), false, isValid)
}

func callGetOSSEntryForDomainAndReturnService(t *testing.T, service *ossrecord.OSSService, incl IncludeOptions, expectError bool) ossrecord.OSSService {
	var entry ossrecord.OSSEntry
	entry = service
	ossEntryForDomain, err := getOSSEntryForDomain(&entry, incl)
	if expectError {
		testhelper.AssertEqual(t, "error", true, err != nil)
		return *service
	}
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "OSS entry for domain is nil", false, ossEntryForDomain == nil)
	return *(ossEntryForDomain.(*ossrecord.OSSService))
}

func createTestOSSServiceExtendedObject(includeOverrides bool) ossrecordextended.OSSServiceExtended {
	// Create OSS service extended record that will be used for testing:
	name := ossrecord.CRNServiceName("osscatalog-testing-convert")
	ossrec := ossrecordextended.NewOSSServiceExtended(name)
	ossrec.OSSService.GeneralInfo.EntryType = ossrecord.SERVICE
	ossrec.OSSService.GeneralInfo.Domain = ossrecord.COMMERCIAL
	ossrec.OSSService.GeneralInfo.OperationalStatus = ossrecord.GA
	ossrec.OSSService.GeneralInfo.OSSTags.AddTag(osstags.OSSTest)

	if includeOverrides {
		// Add a US regulated override to the OSS service:
		override := ossrecord.OSSServiceOverride{}
		tags := osstags.TagSet{}
		tags.AddTag(osstags.ServiceNowApproved)
		override.GeneralInfo.OSSTags = tags
		override.GeneralInfo.ServiceNowSysid = "fakeSNSysId"
		override.GeneralInfo.ServiceNowCIURL = "fakeSNCIURL"
		override.GeneralInfo.Domain = ossrecord.USREGULATED
		override.Compliance.ServiceNowOnboarded = true
		var overrides []ossrecord.OSSServiceOverride
		overrides = append(overrides, override)
		ossrec.OSSService.Overrides = overrides
	}

	return *ossrec
}

func callGetDomainSpecificOSSEntriesAndVerifyNoRecordsAreReturned(t *testing.T, ossService *ossrecord.OSSService, incl IncludeOptions) {
	var entry ossrecord.OSSEntry
	entry = ossService
	services := getDomainSpecificOSSEntries(&entry, incl)
	testhelper.AssertEqual(t, "number of services", 0, len(services))
}

func callGetDomainSpecificOSSEntriesAndVerifyOnlyCommercialIsReturned(t *testing.T, ossService *ossrecord.OSSService, incl IncludeOptions, isOverrideExpected bool) {
	var entry ossrecord.OSSEntry
	entry = ossService
	services := getDomainSpecificOSSEntries(&entry, incl)
	testhelper.AssertEqual(t, "number of services", 1, len(services))
	actualService := *(services[0].(*ossrecord.OSSService))
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.COMMERCIAL, actualService.GeneralInfo.Domain)
	verifyOverridesOnOSSService(t, actualService, isOverrideExpected)
	verifyCommercialOSSService(t, *ossService, actualService, isOverrideExpected)
}

func callGetDomainSpecificOSSEntriesAndVerifyCommercialAndUSRegulatedAreReturned(t *testing.T, ossService *ossrecord.OSSService, incl IncludeOptions, isOverrideExpected bool) {
	var entry ossrecord.OSSEntry
	entry = ossService
	services := getDomainSpecificOSSEntries(&entry, incl)
	testhelper.AssertEqual(t, "number of services", 2, len(services))

	actualServiceUSRegulated := *(services[0].(*ossrecord.OSSService))
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.USREGULATED, actualServiceUSRegulated.GeneralInfo.Domain)
	verifyOverridesOnOSSService(t, actualServiceUSRegulated, isOverrideExpected)
	verifyUSRegulatedOSSService(t, *ossService, actualServiceUSRegulated)

	actualServiceCommercial := *(services[1].(*ossrecord.OSSService))
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.COMMERCIAL, actualServiceCommercial.GeneralInfo.Domain)
	verifyOverridesOnOSSService(t, actualServiceCommercial, isOverrideExpected)
	verifyCommercialOSSService(t, *ossService, actualServiceCommercial, isOverrideExpected)
}

func callGetDomainSpecificOSSEntriesAndVerifyOnlyUSRegulatedIsReturned(t *testing.T, ossService *ossrecord.OSSService, incl IncludeOptions, isOverrideExpected bool) {
	var entry ossrecord.OSSEntry
	entry = ossService
	services := getDomainSpecificOSSEntries(&entry, incl)
	testhelper.AssertEqual(t, "number of services", 1, len(services))
	actualService := *(services[0].(*ossrecord.OSSService))
	testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.USREGULATED, actualService.GeneralInfo.Domain)
	verifyOverridesOnOSSService(t, actualService, isOverrideExpected)
	verifyUSRegulatedOSSService(t, *ossService, actualService)
}

func verifyOverridesOnOSSService(t *testing.T, ossService ossrecord.OSSService, isOverrideExpected bool) {
	if isOverrideExpected {
		testhelper.AssertEqual(t, "ossdata.Overrides", 1, len(ossService.Overrides))
		testhelper.AssertEqual(t, "overrides.GeneralInfo.OSSTags", 1, len(ossService.Overrides[0].GeneralInfo.OSSTags))
		testhelper.AssertEqual(t, "overrides.GeneralInfo.OSSTags.ServiceNowApproved", true, ossService.Overrides[0].GeneralInfo.OSSTags.Contains(osstags.ServiceNowApproved))
		testhelper.AssertEqual(t, "override.GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid("fakeSNSysId"), ossService.Overrides[0].GeneralInfo.ServiceNowSysid)
		testhelper.AssertEqual(t, "override.GeneralInfo.ServiceNowCIURL", "fakeSNCIURL", ossService.Overrides[0].GeneralInfo.ServiceNowCIURL)
		testhelper.AssertEqual(t, "override.GeneralInfo.Domain", ossrecord.USREGULATED, ossService.Overrides[0].GeneralInfo.Domain)
		testhelper.AssertEqual(t, "override.Compliance.ServiceNowOnboarded", true, ossService.Overrides[0].Compliance.ServiceNowOnboarded)
	} else {
		testhelper.AssertEqual(t, "ossdata.Overrides", 0, len(ossService.Overrides))
	}
}

func verifyCommercialOSSService(t *testing.T, expected ossrecord.OSSService, actual ossrecord.OSSService, isOverrideExpected bool) {
	if !isOverrideExpected {
		// remove overrides from the expected service if there so we can compare to the actual service:
		expected.Overrides = nil
	}
	compareOutput := compare.Output{}
	compare.DeepCompare("expected", expected, "actual", actual, &compareOutput)
	testhelper.AssertEqual(t, "ossdata diff", 0, len(compareOutput.GetDiffs()))
}

func verifyUSRegulatedOSSService(t *testing.T, expected ossrecord.OSSService, actual ossrecord.OSSService) {
	testhelper.AssertEqual(t, "ossdata.ReferenceResourceName", expected.ReferenceResourceName, actual.ReferenceResourceName)
	testhelper.AssertEqual(t, "ossdata.GeneralInfo.EntryType", expected.GeneralInfo.EntryType, actual.GeneralInfo.EntryType)
	testhelper.AssertEqual(t, "ossdata.GeneralInfo.OperationalStatus", ossrecord.GA, actual.GeneralInfo.OperationalStatus)
	testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags", 2, len(actual.GeneralInfo.OSSTags))
	testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags.OSSTest", true, actual.GeneralInfo.OSSTags.Contains(osstags.OSSTest))
	testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags.ServiceNowApproved", true, actual.GeneralInfo.OSSTags.Contains(osstags.ServiceNowApproved))
	testhelper.AssertEqual(t, "ossdata.GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid("fakeSNSysId"), actual.GeneralInfo.ServiceNowSysid)
	testhelper.AssertEqual(t, "ossdata.GeneralInfo.ServiceNowCIURL", "fakeSNCIURL", actual.GeneralInfo.ServiceNowCIURL)
	testhelper.AssertEqual(t, "ossdata.GeneralInfo.Domain", ossrecord.USREGULATED, actual.GeneralInfo.Domain)
	testhelper.AssertEqual(t, "ossdata.Compliance.ServiceNowOnboarded", true, actual.Compliance.ServiceNowOnboarded)
}
