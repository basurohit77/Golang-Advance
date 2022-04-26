package servicenow

import (
	"strconv"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

// TestSysidTableCache tests the sysdig table cache including the LookupServiceNowSysids and RecordSysid functions
func TestSysidTableCache(t *testing.T) {
	testSysidTableCacheCommon(t, true, WATSON)
	testSysidTableCacheCommon(t, true, CLOUDFED)
	testSysidTableCacheCommon(t, false, WATSON)
	testSysidTableCacheCommon(t, false, CLOUDFED)
}

func testSysidTableCacheCommon(t *testing.T, testMode bool, snDomain ServiceNowDomain) {
	crnServiceName := ossrecord.CRNServiceName("fakeCRNSN-" + strconv.FormatBool(testMode) + "-" + string(snDomain))
	sysID := ossrecord.ServiceNowSysid("fakeSysId-" + strconv.FormatBool(testMode) + "-" + string(snDomain))

	// Verify that the cache for the provided test mode and domain is empty to start:
	sysIdTable := getSysidTable(testMode, snDomain)
	testhelper.AssertEqual(t, string(crnServiceName)+": sysIdTable initial size check", 0, len(sysIdTable))

	// Verify that the sysid array is not present in the cache using the LookupServiceNowSysids function:
	sysIDArrayActual, ok := LookupServiceNowSysids(crnServiceName, testMode, snDomain)
	testhelper.AssertEqual(t, string(crnServiceName)+": Initial call to LookupServiceNowSysids ok check", false, ok)
	testhelper.AssertEqual(t, string(crnServiceName)+": Initial call to LookupServiceNowSysids array check", 0, len(sysIDArrayActual))

	// Add a sysid array to the cache using the RecordSysid function:
	RecordSysid(crnServiceName, sysID, testMode, snDomain)

	// Verify that the sysid array is present in the cache:
	sysIdTable = getSysidTable(testMode, snDomain)
	testhelper.AssertEqual(t, string(crnServiceName)+": sysIdTable size check after addition", 1, len(sysIdTable))
	var sysIDArray []ossrecord.ServiceNowSysid
	sysIDArray = append(sysIDArray, sysID)
	testhelper.AssertEqual(t, string(crnServiceName)+": sysIdTable record check", sysIDArray, sysIdTable[crnServiceName])

	// Verify that the sysid array is present in the cache using the LookupServiceNowSysids function:
	sysIDArrayActual, ok = LookupServiceNowSysids(crnServiceName, testMode, snDomain)
	testhelper.AssertEqual(t, string(crnServiceName)+": After update call to LookupServiceNowSysids ok check", true, ok)
	testhelper.AssertEqual(t, string(crnServiceName)+": After update call to LookupServiceNowSysids array check", 1, len(sysIDArrayActual))
}
