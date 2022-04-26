package servicenow

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestUpdateServiceNowCI(t *testing.T) {
	if !testhelper.IsReleaseTest() {
		t.Skip("Skipping test TestUpdateServiceNowCI()")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}

	rest.LoadDefaultKeyFile()
	type ciPatch struct {
		CRNServiceName string `json:"crn_service_name"`
		DisplayName    string `json:"display_name"`
	}
	var testname = "emma test name"
	ossrec := ciPatch{
		CRNServiceName: "emmatest5",
		DisplayName:    testname,
	}

	result, err := UpdateServiceNowRecordByName("", ossrec, true, WATSON)
	testhelper.AssertEqual(t, "Should be no error returned from update", nil, err)
	testhelper.AssertEqual(t, "Updated DisplayName", testname, result.DisplayName)
}

func TestCRUDServicenowRecord(t *testing.T) {
	if !testhelper.IsReleaseTest() {
		t.Skip("Skipping test TestCRUDServicenowRecord()")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ServiceNow | debug.Fine /* XXX */)
	}
	rest.LoadDefaultKeyFile()

	sample := []byte(`{
		"schema_version": "1.0.11",
		"general_info": {
		  "operational_status": "BETA",
		  "entry_type": "RUNTIME"
		},
		"ownership": {
		  "offering_manager": {
			"w3id": "hongling@ca.ibm.com",
			"name": "offmger"
		  },
		  "tribe_id": "TBD"
		},
		"support": {
		  "manager": {
			"w3id": "hongling@ca.ibm.com",
			"name": "support"
		  },
		  "client_experience": "ACS_SUPPORTED"
		},
		"operations": {
		  "manager": {
			"w3id": "hongling@ca.ibm.com",
			"name": "opmgr"
		  }
		}
	  }`)
	ossrec := new(ossrecord.OSSService)
	json.Unmarshal(sample, ossrec)
	testname := fmt.Sprintf("osscatalog-testing%v", time.Now().Unix())
	debug.Info("Create servicenow CI with name: %v", testname)
	ossrec.ReferenceResourceName = ossrecord.CRNServiceName(testname)

	//sleepTime := 30 * time.Second
	sleepTime := time.Duration(1)

	// dt := debugT{t}
	dt := t
	dt.Run("create", func(t *testing.T) {
		serv, err := CreateServiceNowRecordFromOSSService("", ossrec, true, WATSON)
		testhelper.AssertEqual(t, "Error returned from creation", nil, err)
		testhelper.AssertNotEqual(t, "SN sys id", "", serv.GeneralInfo.ServiceNowSysid)
		testhelper.AssertNotEqual(t, "SN CI url", "", serv.GeneralInfo.ServiceNowCIURL)
		testhelper.AssertEqual(t, "CI status", ossrecord.BETA, serv.GeneralInfo.OperationalStatus)
		testhelper.AssertEqual(t, "CI type", ossrecord.RUNTIME, serv.GeneralInfo.EntryType)
		time.Sleep(sleepTime)
	})

	dt.Run("sync", func(t *testing.T) {
		serv, err := SyncServiceNowRecord("", ossrec, true, WATSON)
		testhelper.AssertEqual(t, "Error returned from sync", nil, err)
		testhelper.AssertEqual(t, "SN sys id", ossrec.GeneralInfo.ServiceNowSysid, serv.GeneralInfo.ServiceNowSysid)
		testhelper.AssertEqual(t, "SN CI url", ossrec.GeneralInfo.ServiceNowCIURL, serv.GeneralInfo.ServiceNowCIURL)
		testhelper.AssertEqual(t, "CI status", ossrec.GeneralInfo.OperationalStatus, serv.GeneralInfo.OperationalStatus)
		testhelper.AssertEqual(t, "CI type", ossrec.GeneralInfo.EntryType, serv.GeneralInfo.EntryType)
		time.Sleep(sleepTime)
	})

	dt.Run("update", func(t *testing.T) {
		ossrec.GeneralInfo.OperationalStatus = ossrecord.GA
		ossrec.GeneralInfo.EntryType = ossrecord.SERVICE
		result, err := UpdateServiceNowRecordFromOSSService("", ossrec, true, WATSON)
		testhelper.AssertEqual(t, "Error returned from update", nil, err)
		testhelper.AssertEqual(t, "CI status", ossrecord.GA, result.GeneralInfo.OperationalStatus)
		testhelper.AssertEqual(t, "CI type", ossrecord.SERVICE, result.GeneralInfo.EntryType)
	})
}
