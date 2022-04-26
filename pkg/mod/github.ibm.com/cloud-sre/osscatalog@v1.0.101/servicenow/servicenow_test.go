package servicenow

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestReadServiceNowRecordBySysidWatson(t *testing.T) {
	testReadServiceNowRecordBySysidCommon(t, "ecf79f03db8543008799327e9d9619d5", false, WATSON, "appid")
}

func TestReadServiceNowRecordBySysidTestWatson(t *testing.T) {
	testReadServiceNowRecordBySysidCommon(t, "ecf79f03db8543008799327e9d9619d5", true, WATSON, "appid")
}

func TestReadServiceNowRecordBySysidCloudfed(t *testing.T) {
	testReadServiceNowRecordBySysidCommon(t, "ecf79f03db8543008799327e9d9619d5", false, CLOUDFED, "appid")
}

func TestReadServiceNowRecordBySysidTestCloudfed(t *testing.T) {
	testReadServiceNowRecordBySysidCommon(t, "ecf79f03db8543008799327e9d9619d5", true, CLOUDFED, "appid")
}

func testReadServiceNowRecordBySysidCommon(t *testing.T, sysid ossrecord.ServiceNowSysid, testMode bool, snDomain ServiceNowDomain, expectedCRNServiceName string) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test testReadServiceNowRecordBySysidCommon() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}
	rest.LoadDefaultKeyFile()
	result, err := ReadServiceNowRecordBySysid(sysid, testMode, snDomain)
	if err != nil {
		t.Errorf("ReadServiceNowRecordBySysid failed: %v", err)
	}
	if result.CRNServiceName != expectedCRNServiceName {
		t.Errorf("ReadServiceNowRecordBySysid did not return a record with the expected name \"%s\"", expectedCRNServiceName)
	}
}

func TestListServiceNowRecordsWatson(t *testing.T) {
	testListServiceNowRecordsCommon(t, false, WATSON)
}

func TestListServiceNowRecordsTestWatson(t *testing.T) {
	testListServiceNowRecordsCommon(t, true, WATSON)
}

func TestListServiceNowRecordsCloudfed(t *testing.T) {
	testListServiceNowRecordsCommon(t, false, CLOUDFED)
}

func TestListServiceNowRecordsTestCloudfed(t *testing.T) {
	testListServiceNowRecordsCommon(t, true, CLOUDFED)
}

func testListServiceNowRecordsCommon(t *testing.T, testMode bool, snDomain ServiceNowDomain) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test testListServiceNowRecordsCommon() in short mode")
	}
	/* *testhelper.VeryVerbose = true /* XXX */

	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	rest.LoadDefaultKeyFile()

	/*
		// Dummy up the sysid file
		sysidTable["bluemix-developer-experience"] = []ossrecord.ServiceNowSysid{"003e024cdb5d47008799327e9d9619f3"}
		sysidTable["cloudantnosqldb"] = []ossrecord.ServiceNowSysid{"01f7df03db8543008799327e9d96193a"}
		sysidTable["apiconnect"] = []ossrecord.ServiceNowSysid{"20f79f03db8543008799327e9d9619d4"}
		sysidTable["dashdb"] = []ossrecord.ServiceNowSysid{"34f79f03db8543008799327e9d9619e4"}
		sysidTable["watson-health-security"] = []ossrecord.ServiceNowSysid{"ac928a60db4863406001327e9d961909", "afe27a5edb72d7406001327e9d9619e6"}
	*/

	//	pattern := regexp.MustCompile(".*node.*")
	pattern := regexp.MustCompile(".*")

	countResults := 0
	countIssues := 0

	err := ListServiceNowRecords(pattern, testMode, snDomain, func(e *ConfigurationItem, issues []*ossvalidation.ValidationIssue) {
		countResults++
		countIssues += len(issues)
		if *testhelper.VeryVerbose {
			if len(issues) > 0 {
				fmt.Printf(" -> found entry %s   with validation issues=%v\n", e.CRNServiceName, issues)
			} else {
				fmt.Printf(" -> found entry %s\n", e.CRNServiceName)
			}
		}
	})

	testhelper.AssertError(t, err)
	const minResults = 200
	if countResults < minResults {
		t.Errorf("Expected at least %d results but got %d", minResults, countResults)
	}
	testhelper.AssertEqual(t, "Number of entries returned", len(getSysidTable(testMode, snDomain)), countResults)
	testhelper.AssertEqual(t, "Number of validation issues created", 1, countIssues)

	if *testhelper.VeryVerbose {
		fmt.Printf("Read %d entries from ServiceNow  with %d validation issues\n", countResults, countIssues)
	}

}

var (
	testServiceName     string
	testSysid           = "testsysid"
	testCIUrl           = "testCIUrl"
	existingServiceName = "existingService"
	newServiceName      = "newService"
)

func TestCreateServicenowRecord(t *testing.T) {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	defer httpmock.DeactivateAndReset()

	mockHTTPRequest()

	ossrec := new(ossrecord.OSSService)
	testServiceName = newServiceName
	ossrec.ReferenceResourceName = ossrecord.CRNServiceName(testServiceName)

	serv, err := CreateServiceNowRecordFromOSSService("token", ossrec, true, WATSON)
	testhelper.AssertEqual(t, "Error returned from creation", nil, err)
	testhelper.AssertEqual(t, "SN sys id", testSysid, string(serv.GeneralInfo.ServiceNowSysid))
	testhelper.AssertEqual(t, "SN CI url", testCIUrl, serv.GeneralInfo.ServiceNowCIURL)

}

func TestCreateServicenowRecordCloudfed(t *testing.T) {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	defer httpmock.DeactivateAndReset()

	mockHTTPRequest()

	ossrec := new(ossrecord.OSSService)
	testServiceName = newServiceName
	ossrec.ReferenceResourceName = ossrecord.CRNServiceName(testServiceName)

	serv, err := CreateServiceNowRecordFromOSSService("token", ossrec, true, CLOUDFED)
	testhelper.AssertEqual(t, "Error returned from creation", nil, err)
	testhelper.AssertEqual(t, "SN sys id", testSysid, string(serv.Overrides[0].GeneralInfo.ServiceNowSysid))
	testhelper.AssertEqual(t, "SN CI url", testCIUrl, serv.Overrides[0].GeneralInfo.ServiceNowCIURL)

}

func TestCreateServicenowRecordWithConflictName(t *testing.T) {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	defer httpmock.DeactivateAndReset()

	mockHTTPRequest()

	ossrec := new(ossrecord.OSSService)
	testServiceName = existingServiceName
	ossrec.ReferenceResourceName = ossrecord.CRNServiceName(testServiceName)

	_, err := CreateServiceNowRecordFromOSSService("token", ossrec, true, WATSON)
	testhelper.AssertNotEqual(t, "Error returned from creation", nil, err)

}

func TestUpdateServicenowRecord(t *testing.T) {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	defer httpmock.DeactivateAndReset()

	testServiceName = existingServiceName
	mockHTTPRequest()

	ossrec := new(ossrecord.OSSService)
	ossrec.ReferenceResourceName = ossrecord.CRNServiceName(testServiceName)

	serv, err := UpdateServiceNowRecordFromOSSService("token", ossrec, true, WATSON)
	testhelper.AssertEqual(t, "Error returned from update", nil, err)
	testhelper.AssertEqual(t, "service offering", "updated", serv.Ownership.ServiceOffering)

}

func TestUpdateServicenowRecordCloudfed(t *testing.T) {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	defer httpmock.DeactivateAndReset()

	testServiceName = existingServiceName
	mockHTTPRequest()

	ossrec := new(ossrecord.OSSService)
	ossrec.ReferenceResourceName = ossrecord.CRNServiceName(testServiceName)

	serv, err := UpdateServiceNowRecordFromOSSService("token", ossrec, true, CLOUDFED)
	testhelper.AssertEqual(t, "Error returned from update", nil, err)
	testhelper.AssertEqual(t, "sn sysid", ossrec.GeneralInfo.ServiceNowSysid, serv.GeneralInfo.ServiceNowSysid)

}

func TestUpdateServicenowRecordWithNotExistName(t *testing.T) {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	defer httpmock.DeactivateAndReset()

	testServiceName = newServiceName
	mockHTTPRequest()

	ossrec := new(ossrecord.OSSService)
	ossrec.ReferenceResourceName = ossrecord.CRNServiceName(testServiceName)

	_, err := UpdateServiceNowRecordFromOSSService("token", ossrec, true, WATSON)
	testhelper.AssertNotEqual(t, "Error returned from update", nil, err)

}

func TestUpdateServicenowRecordWithoutName(t *testing.T) {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	defer httpmock.DeactivateAndReset()

	testServiceName = existingServiceName
	mockHTTPRequest()

	type ciPatch struct {
		CRNServiceName string `json:"crn_service_name"`
		OtherField     string `json:"other_field"`
	}
	ossrec := ciPatch{
		CRNServiceName: testServiceName,
	}
	_, err := UpdateServiceNowRecordByName("token", ossrec, true, WATSON)
	testhelper.AssertEqual(t, "Should be no error returned from update", nil, err)

	_, err = UpdateServiceNowRecordByName("token", &ossrec, true, WATSON)
	testhelper.AssertEqual(t, "Should be no error returned from update", nil, err)

	ossrec = ciPatch{
		OtherField: "abc",
	}
	_, err = UpdateServiceNowRecordByName("token", ossrec, true, WATSON)
	testhelper.AssertNotEqual(t, "Should return name error", nil, err)
}

func TestSyncServicenowRecord(t *testing.T) {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	defer httpmock.DeactivateAndReset()

	testServiceName = existingServiceName
	mockHTTPRequest()

	ossrec := new(ossrecord.OSSService)
	ossrec.ReferenceResourceName = ossrecord.CRNServiceName(testServiceName)
	ossrec.GeneralInfo.ServiceNowSysid = ossrecord.ServiceNowSysid(testServiceName)

	serv, err := SyncServiceNowRecord("token", ossrec, true, WATSON)
	testhelper.AssertEqual(t, "Error returned from update", nil, err)
	testhelper.AssertEqual(t, "oss desc", "synced", serv.Ownership.ServiceOffering)

}

func TestSyncServicenowRecordCloudfed(t *testing.T) {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	defer httpmock.DeactivateAndReset()

	testServiceName = existingServiceName
	mockHTTPRequest()

	ossrec := new(ossrecord.OSSService)
	ossrec.ReferenceResourceName = ossrecord.CRNServiceName(testServiceName)
	ossrec.GeneralInfo.ServiceNowSysid = ossrecord.ServiceNowSysid(testServiceName)

	serv, err := SyncServiceNowRecord("token", ossrec, true, CLOUDFED)
	testhelper.AssertEqual(t, "Error returned from update", nil, err)
	testhelper.AssertEqual(t, "sn sysid", ossrec.GeneralInfo.ServiceNowSysid, serv.GeneralInfo.ServiceNowSysid)

}

func TestSyncServicenowRecordWithNotExistName(t *testing.T) {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	defer httpmock.DeactivateAndReset()

	testServiceName = newServiceName
	mockHTTPRequest()

	ossrec := new(ossrecord.OSSService)
	ossrec.ReferenceResourceName = ossrecord.CRNServiceName(testServiceName)

	_, err := SyncServiceNowRecord("token", ossrec, true, WATSON)
	testhelper.AssertNotEqual(t, "Error returned from update", nil, err)

}

func mockHTTPRequest() {
	httpmock.Activate()
	url := serviceNowWatsonTestURLv2
	var resultContainer struct {
		Result ConfigurationItem `json:"result"`
	}
	httpmock.RegisterResponder("GET", url+"/"+testServiceName, func(req *http.Request) (*http.Response, error) {
		if testServiceName == newServiceName {
			resp := httpmock.NewStringResponse(http.StatusNotFound, "no record found")
			err := errors.New("not found")
			return resp, err
		}
		ci := new(ConfigurationItem)
		ci.CRNServiceName = testServiceName
		ci.Ownership.ServiceOffering = "synced"
		resultContainer.Result = *ci
		resp, _ := httpmock.NewJsonResponse(200, resultContainer)
		return resp, nil
	})

	httpmock.RegisterResponder("POST", url, func(req *http.Request) (*http.Response, error) {
		body, _ := ioutil.ReadAll(req.Body)
		var ci ConfigurationItem
		json.Unmarshal(body, &ci)
		if ci.CRNServiceName == existingServiceName {
			resp := httpmock.NewStringResponse(http.StatusConflict, "conflict")
			err := errors.New("conflict")
			return resp, err
		}

		ci.SysID = testSysid
		ci.GeneralInfo.ServiceNowCIURL = testCIUrl
		resultContainer.Result = ci
		resp, _ := httpmock.NewJsonResponse(200, resultContainer)
		return resp, nil
	})

	httpmock.RegisterResponder("PATCH", url+"/"+testServiceName, func(req *http.Request) (*http.Response, error) {
		if testServiceName == newServiceName {
			resp := httpmock.NewStringResponse(http.StatusBadRequest, "no record found")
			err := errors.New("not found")
			return resp, err
		}
		var ci ConfigurationItem
		ci.Ownership.ServiceOffering = "updated"
		resultContainer.Result = ci
		resp, _ := httpmock.NewJsonResponse(200, resultContainer)
		return resp, nil
	})

	url = serviceNowCloudfedTestURLv2

	httpmock.RegisterResponder("GET", url+"/"+testServiceName, func(req *http.Request) (*http.Response, error) {
		if testServiceName == newServiceName {
			resp := httpmock.NewStringResponse(http.StatusNotFound, "no record found")
			err := errors.New("not found")
			return resp, err
		}
		ci := new(ConfigurationItem)
		ci.CRNServiceName = testServiceName
		ci.Ownership.ServiceOffering = "synced-cloudfed"
		resultContainer.Result = *ci
		resp, _ := httpmock.NewJsonResponse(200, resultContainer)
		return resp, nil
	})

	httpmock.RegisterResponder("PATCH", url+"/"+testServiceName, func(req *http.Request) (*http.Response, error) {
		if testServiceName == newServiceName {
			resp := httpmock.NewStringResponse(http.StatusBadRequest, "no record found")
			err := errors.New("not found")
			return resp, err
		}
		var ci ConfigurationItem
		ci.Ownership.ServiceOffering = "updated-cloudfed"
		resultContainer.Result = ci
		resp, _ := httpmock.NewJsonResponse(200, resultContainer)
		return resp, nil
	})

	httpmock.RegisterResponder("POST", url, func(req *http.Request) (*http.Response, error) {
		body, _ := ioutil.ReadAll(req.Body)
		var ci ConfigurationItem
		json.Unmarshal(body, &ci)
		if ci.CRNServiceName == existingServiceName {
			resp := httpmock.NewStringResponse(http.StatusConflict, "conflict")
			err := errors.New("conflict")
			return resp, err
		}

		ci.SysID = testSysid
		ci.GeneralInfo.ServiceNowCIURL = testCIUrl
		resultContainer.Result = ci
		resp, _ := httpmock.NewJsonResponse(200, resultContainer)
		return resp, nil
	})
}

// TestGetCachedSNRecords tests the GetCachedSNRecords function
func TestGetCachedSNRecords(t *testing.T) {
	testGetCachedSNRecordsCommon(t, true, WATSON)
	testGetCachedSNRecordsCommon(t, true, CLOUDFED)
	testGetCachedSNRecordsCommon(t, false, WATSON)
	testGetCachedSNRecordsCommon(t, false, CLOUDFED)
}

func testGetCachedSNRecordsCommon(t *testing.T, testMode bool, snDomain ServiceNowDomain) {
	sysID := ossrecord.ServiceNowSysid("fakeSysId-" + strconv.FormatBool(testMode) + "-" + string(snDomain))

	// Verify that the cache for the provided test mode and domain is empty to start:
	cachedRecords := getCachedSNRecords(testMode, snDomain)
	testhelper.AssertEqual(t, string(sysID)+": cachedRecords initial size check", 0, len(cachedRecords))

	// Add a configuration item to the cache:
	configItem := ConfigurationItem{SysID: string(sysID)}
	getCachedSNRecords(testMode, snDomain)[sysID] = &configItem

	// Verify that the configuration item is present in the cache:
	cachedRecords = getCachedSNRecords(testMode, snDomain)
	testhelper.AssertEqual(t, string(sysID)+": cachedRecords size check after addition", 1, len(cachedRecords))
	testhelper.AssertEqual(t, string(sysID)+": cached record check", &configItem, cachedRecords[sysID])
}
