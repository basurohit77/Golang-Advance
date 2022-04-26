package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
	"github.ibm.com/cloud-sre/pnp-nq2ds/shared"
	"log"
	"os"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/ossmon"

	// "database/sql"
	"github.com/google/go-cmp/cmp"
	newrelic "github.com/newrelic/go-agent"
	"github.com/stretchr/testify/assert"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-nq2ds/testutils"
)

var (
	sampleMaintenanceMsg1 = `{
	"operation":"insert",
	"info":{
		"source_id":"123456",
		"disruptive":true,
		"short_description":"maintenance short description 1",
		"long_description":"maintenance long description 1",
		"planned_start_time":null,
		"planned_end_time":null,
		"source_creation_time":"2018-06-07T15:28:58Z",
		"source_update_time":"2018-08-14T15:34:31Z",
		"state":"new",
		"crnFull":["crn:v1:bluemix:public:service-name1:us-east::::"],
		"source":"servicenow",
		"regulatory_domain": null,
		"maintenance_duration": 480,
		"disruption_type": "Other (specify in Description)",
		"disruption_description": "test",
		"disruption_duration": 480,
		"source_state": "Pending Schedule",
		"notification_status": "<unassigned>",
		"notification_type": "<unassigned>",
		"notification_channels": ""
		},
	"tags":"%s"
	}`
	//
	//sampleMaintenanceMsg3 = `{
	//"operation":"update",
	//"info":{
	//	"source_id":"123456",
	//	"disruptive":true,
	//	"short_description":"%s",
	//	"long_description":"maintenance long description 1",
	//	"planned_start_time":null,
	//	"planned_end_time":null,
	//	"source_creation_time":"2018-06-07T15:28:58Z",
	//	"source_update_time":"%s",
	//	"state":"%s",
	//	"crnFull":["crn:v1:bluemix:public:service-name1:us-east::::"],
	//	"source":"servicenow",
	//	"regulatory_domain": null,
	//	"maintenance_duration": 480,
	//	"disruption_type": "Other (specify in Description)",
	//	"disruption_description": "test",
	//	"disruption_duration": 480,
	//	"source_state": "%s",
	//	"notification_status": "%s",
	//	"notification_type": "Maintenance",
	//	"notification_channels": "Public,Local",
	//	"completion_code": "%s"
	//	},
	//"tags":null
	//}`

	sampleMaintenanceMsg4 = `{
		"result_from_sn":[{
			"source_id":"%s",
			"disruptive":true,
			"short_description":"%s",
			"long_description":"SNmaintenance long description 1",
			"planned_start_time":null,
			"planned_end_time":null,
			"source_creation_time":"2018-06-07T15:28:58Z",
			"source_update_time":"2018-09-07T15:28:58Z",
			"state":"complete",
			"crnFull":["crn:v1:bluemix:public:service-name1:us-east::::"],
			"source":"servicenow",
			"regulatory_domain": null,
			"maintenance_duration": 480,
			"record_hash": "%s",
			"disruption_type": "Other (specify in Description)",
			"disruption_description": "test",
			"disruption_duration": 480,
			"u_outage_duration": "1 day 1 Hour 1 Minute 5 Seconds"
			}]
		}`

	sampleMaintenanceMsg4NotBulk = `{
				"source_id":"%s",
				"disruptive":true,
				"short_description":"%s",
				"long_description":"SNmaintenance long description 1",
				"planned_start_time":null,
				"planned_end_time":null,
				"source_creation_time":"2018-06-07T15:28:58Z",
				"source_update_time":"2018-09-07T15:28:58Z",
				"state":"complete",
				"crnFull":["crn:v1:bluemix:public:service-name1:us-east::::"],
				"source":"servicenow",
				"regulatory_domain": null,
				"maintenance_duration": 480,
				"record_hash": "%s",
				"disruption_type": "Other (specify in Description)",
				"disruption_description": "test",
				"disruption_duration": 480,
				"u_outage_duration": "1 Day 1 Hour 1 Minute 5 Seconds"
			}`

	sampleMaintenanceMsg5 = `{
	"record_id": "%s",
	"pnp_creationTime": "2018-10-30T15:55:13Z",
	"pnp_update_time": "2018-10-30T15:55:13Z",
	"source_creation_time": "2018-10-02T01:58:42Z",
	"source_update_time": "%s",
	"planned_start_time": "2018-10-22T11:00:00Z",
	"planned_end_time": "2018-10-22T13:00:00Z",
	"short_description": "from bspn loader",
	"long_description": "test",
	"text": "<b> Test text </b>",
	"state": "complete",
	"disruptive": true,
	"source_id": "882920",
	"source": "servicenow",
	"crnFull": ["crn:v1:d-wbmdn-28801:dedicated:conversation:us-south::::"],
	"maintenance_duration": 60,
	"disruption_type": "Other ",
	"disruption_description": "test",
	"disruption_duration": 20
}`

	sampleMaintenanceMsgWithCommunication = `{
	"operation": "update",
	"sys_id": "b28ab5e4db303f00fc0e389f9d9619e0",
	"number": "CHG0196584",
	"sys_created_on": "2019-04-10 19:47:03",
	"state": "scheduled",
	"priority": "critical",
	"short_description": "servicenow IBM Yellow Staging 0 (YS0) Standard 2019-04-11 21:37:49",
	"description": "desc",
	"u_severity": "impact7",
	"u_purpose_goal": "test",
	"backout_plan": "backout",
	"start_date": "2019-04-11 21:37:49",
	"end_date": "2019-04-12 19:39:53",
	"work_start": "",
	"work_end": "",
	"u_outage_duration": 61,
	"sys_created_by": "michael_lee@us.ibm.com",
	"sys_updated_by": "michael_lee@us.ibm.com",
	"sys_updated_on": "2019-04-17 02:47:54",
	"crn": [
	  "crn:v1:bluemix:public:cloudantnosqldb:us-south::::"
	],
	"communications": [
	  {
		"sys_id": "b3c5e24cdbd47fc08799327e9d961905",
		"number": "PUB0001020",
		"stage": "active",
		"short_description": "test by mike.little@us.ibm.com / update",
		"text": "<p><em><strong>What are we changing?</strong></em></p><p>test by mike.little&#64;us.ibm.com</p><br /><p><em><strong>Why are we making this change?</strong></em></p><p>test by mike.little&#64;us.ibm.com</p><br /><p><em><strong>How will it impact the environment?</strong></em></p><p>test by mike.little&#64;us.ibm.com</p>",
		"sys_created_by": "mike.little@us.ibm.com",
		"sys_updated_by": "michael_lee@us.ibm.com",
		"sys_created_on": "2019-03-15 13:07:25",
		"sys_updated_on": "2019-04-10 19:48:48"
	  }
	],
	"instance": "watsondev"
  }`

	sampleMaintenanceMsgWithCommunicationWithPublishDate = `{
	"operation": "update",
	"sys_id": "b28ab5e4db303f00fc0e389f9d9619e0",
	"number": "CHG0196584",
	"sys_created_on": "2019-04-10 19:47:03",
	"state": "scheduled",
	"priority": "critical",
	"short_description": "servicenow IBM Yellow Staging 0 (YS0) Standard 2019-04-11 21:37:49",
	"description": "desc",
	"u_severity": "impact7",
	"u_purpose_goal": "test",
	"backout_plan": "backout",
	"start_date": "2019-04-11 21:37:49",
	"end_date": "2019-04-12 19:39:53",
	"work_start": "",
	"work_end": "",
	"u_outage_duration": 61,
	"sys_created_by": "michael_lee@us.ibm.com",
	"sys_updated_by": "michael_lee@us.ibm.com",
	"sys_updated_on": "2019-04-17 02:47:54",
	"crn": [
	  "crn:v1:bluemix:public:cloudantnosqldb:us-south::::"
	],
	"communications": [
	  {
		"sys_id": "b3c5e24cdbd47fc08799327e9d961905",
		"number": "PUB0001020",
		"stage": "active",
		"short_description": "test by mike.little@us.ibm.com / update",
		"text": "<p><em><strong>What are we changing?</strong></em></p><p>test by mike.little&#64;us.ibm.com</p><br /><p><em><strong>Why are we making this change?</strong></em></p><p>test by mike.little&#64;us.ibm.com</p><br /><p><em><strong>How will it impact the environment?</strong></em></p><p>test by mike.little&#64;us.ibm.com</p>",
		"sys_created_by": "mike.little@us.ibm.com",
		"sys_updated_by": "michael_lee@us.ibm.com",
		"sys_created_on": "2019-03-15 13:07:25",
		"sys_updated_on": "2019-04-10 19:48:48",
        "publish_date": "2019-04-10 19:48:48"
	  }
	],
	"instance": "watsondev"
  }`

	sampleMaintenanceMsgWithCommunicationNonDisruptive = `{
	"operation": "update",
	"sys_id": "b28ab5e4db303f00fc0e389f9d9619e0",
	"number": "CHG0196584",
	"sys_created_on": "2019-04-10 19:47:03",
	"state": "scheduled",
	"priority": "critical",
	"short_description": "servicenow IBM Yellow Staging 0 (YS0) Standard 2019-04-11 21:37:49",
	"description": "desc",
	"u_severity": "impact7",
	"u_purpose_goal": "test",
	"backout_plan": "backout",
	"start_date": "2019-04-11 21:37:49",
	"end_date": "2019-04-12 19:39:53",
	"work_start": "",
	"work_end": "",
	"u_outage_duration": 0,
	"sys_created_by": "michael_lee@us.ibm.com",
	"sys_updated_by": "michael_lee@us.ibm.com",
	"sys_updated_on": "2019-04-17 02:47:54",
	"crn": [
	  "crn:v1:bluemix:public:cloudantnosqldb:us-south::::"
	],
	"communications": [
	  {
		"sys_id": "b3c5e24cdbd47fc08799327e9d961905",
		"number": "PUB0001020",
		"stage": "active",
		"short_description": "test by mike.little@us.ibm.com / update",
		"text": "<p><em><strong>What are we changing?</strong></em></p><p>test by mike.little&#64;us.ibm.com</p><br /><p><em><strong>Why are we making this change?</strong></em></p><p>test by mike.little&#64;us.ibm.com</p><br /><p><em><strong>How will it impact the environment?</strong></em></p><p>test by mike.little&#64;us.ibm.com</p>",
		"sys_created_by": "mike.little@us.ibm.com",
		"sys_updated_by": "michael_lee@us.ibm.com",
		"sys_created_on": "2019-03-15 13:07:25",
		"sys_updated_on": "2019-04-10 19:48:48"
	  }
	],
	"instance": "watsondev"
  }`

	sampleMaintenanceMsg6 = `{
	"source_creation_time": "2018-10-02T01:58:42Z",
	"source_update_time": "%s",
	"planned_start_time": "2018-10-22T11:00:00Z",
	"planned_end_time": "2018-10-22T13:00:00Z",
	"short_description": "from bspn loader",
	"long_description": "test",
	"state": "%s",
	"disruptive": true,
	"source_id": "%s",
	"source": "%s",
	"crnFull": ["%s"],
	"maintenance_duration": 60,
	"disruption_type": "Other ",
	"disruption_description": "test",
	"disruption_duration": 20
}`

	sampleMaintenanceMsg7 = `{
	"source_creation_time": "2018-10-02T01:58:42Z",
	"source_update_time": "2018-10-02T01:58:42Z",
	"planned_start_time": "%s",
	"planned_end_time": "%s",
	"short_description": "from bspn loader",
	"long_description": "test",
	"state": "%s",
	"disruptive": true,
	"source_id": "890000",
	"source": "servicenow",
	"crnFull": ["crn:v1:d-wbmdn-28801:dedicated:conversation:us-south::::"],
	"maintenance_duration": 60,
	"disruption_type": "Other ",
	"disruption_description": "test",
	"disruption_duration": 20
}`

	sampleChangePayloadCrIsNewer = `{
	"info": {
	  "cr_update_time": "2019-08-14T18:24:54Z",
	  "source_update_time": "2019-08-14T16:24:09Z"
	}
  }`

	sampleChangePayloadSourceIsNewer = `{
	"info": {
	  "cr_update_time": "2019-08-14T18:24:54Z",
	  "source_update_time": "2019-08-14T19:24:54Z"
	}
  }`

	sampleChangePayloadCrIsEmpty = `{
	"info": { 
	  "source_update_time": "2019-08-14T19:24:54Z"
	}
  }`
)

// IT IS NOT TESTING ANY CUSTOM CODE FUNCTIONALITY
//func Test_timeparse(t *testing.T) {
//
//	loc, err := time.LoadLocation("EST")
//	if err != nil {
//		log.Fatal(err.Error())
//	}
//	time.Local = loc
//
//	dnew, _ := dateparse.ParseLocal("2019-04-24 15:53:26")
//	dexisting, _ := dateparse.ParseLocal("2019-04-24 15:48:57 -0500 -0500")
//	log.Println("New:", dnew, "\nExisting:", dexisting, "\nnewer?:", dnew.After(dexisting))
//}

func Test_validateMaintInsert_source(t *testing.T) {
	log.Print(tlog.Log())
	source := ""
	sourceID := "882920"
	crn := "crn:v1:bluemix:public:service-name1:us-east::::"
	var testData = fmt.Sprintf(sampleMaintenanceMsg6, "2018-10-24T11:25:36Z", "new", sourceID, source, crn)
	maintenanceInsert := datastore.MaintenanceInsert{}
	err := json.Unmarshal([]byte(testData), &maintenanceInsert)
	if err != nil {
		log.Print("Error occurred unmarshaling , err = ", err)
	}
	errMsg := validateMaintInsert(&maintenanceInsert)
	assert.NotNil(t, errMsg, tlog.Log()+"Bad message: source is empty")
}

func Test_validateMaintInsert_sourceId(t *testing.T) {
	log.Print(tlog.Log())
	source := "servicenow"
	sourceID := ""
	crn := "crn:v1:bluemix:public:service-name1:us-east::::"
	var testData = fmt.Sprintf(sampleMaintenanceMsg6, "2018-10-24T11:25:36Z", "new", sourceID, source, crn)
	maintenanceInsert := datastore.MaintenanceInsert{}
	err := json.Unmarshal([]byte(testData), &maintenanceInsert)
	if err != nil {
		log.Print("Error occurred unmarshaling , err = ", err)
	}
	errMsg := validateMaintInsert(&maintenanceInsert)
	assert.NotNil(t, errMsg, tlog.Log()+"Bad message: SourceID is empty")
}

func Test_validateMaintInsert_crn(t *testing.T) {
	log.Print(tlog.Log())
	source := "servicenow"
	sourceID := "882920"
	crn := ""
	var testData = fmt.Sprintf(sampleMaintenanceMsg6, "2018-10-24T11:25:36Z", "new", sourceID, source, crn)
	maintenanceInsert := datastore.MaintenanceInsert{}
	err := json.Unmarshal([]byte(testData), &maintenanceInsert)
	if err != nil {
		log.Print("Error occurred unmarshaling , err = ", err)
	}
	errMsg := validateMaintInsert(&maintenanceInsert)
	assert.NotNil(t, errMsg, tlog.Log()+"Bad message: CRN is empty")
}

func Test_validateMaintInsert_state(t *testing.T) {
	log.Print(tlog.Log())
	source := "servicenow"
	sourceID := "882920"
	crn := "crn:v1:bluemix:public:service-name1:us-east::::"
	var testData = fmt.Sprintf(sampleMaintenanceMsg6, "2018-10-24T11:25:36Z", "", sourceID, source, crn)
	maintenanceInsert := datastore.MaintenanceInsert{}
	err := json.Unmarshal([]byte(testData), &maintenanceInsert)
	if err != nil {
		log.Print("Error occurred unmarshaling , err = ", err)
	}
	errMsg := validateMaintInsert(&maintenanceInsert)
	assert.NotNil(t, errMsg, tlog.Log()+"Bad message: state is empty")
	sampleMaint := `{"source_creation_time":"","source_update_time":"","planned_start_time":"2018-10-22T11:00:00Z","planned_end_time":"2018-10-22T13:00:00Z","short_description":"from bspn loader","long_description":"test","state":"new","disruptive":true,"source_id":"123456","source":"servicenow","crnFull":["crn:v1:bluemix:public:service-name1:us-east::::"],"maintenance_duration":60,"disruption_type":"Other ","disruption_description":"test","disruption_duration":20}`
	err = json.Unmarshal([]byte(sampleMaint), &maintenanceInsert)
	if err != nil {
		log.Print("Error occurred unmarshaling , err = ", err)
	}
	errMsg = validateMaintInsert(&maintenanceInsert)
	assert.NotNil(t, errMsg, tlog.Log()+"Bad message: SourceCreationTime is empty")
}

func Test_DecodeAndMapMaintenceMessage_decryptionError(t *testing.T) {
	var testData = `{"test":"test"}`
	log.Print(tlog.Log())
	msgMap, decodedMsg := DecodeAndMapMaintenceMessage([]byte(testData), setupMonitor(t))
	assert.Nil(t, msgMap, tlog.Log()+"Error, no messageMap return")
	assert.Nil(t, decodedMsg, tlog.Log()+"Error, no decodedMsg return")
}

func Test_DecodeAndMapMaintenceMessage_unmarshalError(t *testing.T) {
	log.Print(tlog.Log())
	var testData = `{"test":"test"`
	_, msg, _ := testutils.PrepareTestInc(t, testData)
	msgMap, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
	assert.Nil(t, msgMap, "Unmarshal error, no messageMap return")
	assert.NotNil(t, decodedMsg, tlog.Log()+"Unmarshal error, return decryted message")
}

func Test_getExistingMaintenance(t *testing.T) {
	os.Setenv("BYPASS_LOCAL_STORAGE", "true")
	defer os.Unsetenv("BYPASS_LOCAL_STORAGE")
	log.Print(tlog.Log())
	if !shared.BypassLocalStorage {
		var testData = fmt.Sprintf(sampleMaintenanceMsg1, "")
		dbConn, _, mon := testutils.PrepareTestInc(t, testData)
		defer db.Disconnect(dbConn)
		span, ctx := ossmon.StartParentSpan(context.Background(), mon, monitor.SrvPrfx+"testing")
		defer span.Finish()
		existingMaintenance, doesMaintenanceAlreadyExist := getExistingMaintenance(ctx, dbConn, "")
		assert.Nil(t, existingMaintenance, "No return")
		assert.Equal(t, doesMaintenanceAlreadyExist, false)
	}
}

// NO LONGER IN USE, IT IS RELATED TO RTC-DOCTOR
//func Test_processMaintenanceInsert_badTag(t *testing.T) {
//	log.Print(tlog.Log())
//	var testData = fmt.Sprintf(sampleMaintenanceMsg1, "km_communication_not_needed")
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	messageMap, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
//	maintenanceMap := ProcessMaintenance(dbConn, decodedMsg, messageMap, &mon)
//	assert.Nil(t, maintenanceMap, tlog.Log()+"should return nil.")
//}

// NO LONGER IN USE, IT IS RELATED TO RTC-DOCTOR
//func Test_processMaintenanceInsert_moreTags(t *testing.T) {
//	log.Print(tlog.Log())
//	var testData = fmt.Sprintf(sampleMaintenanceMsg1, "yp,km_communication_not_needed")
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	messageMap, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
//	maintenanceMap := ProcessMaintenance(dbConn, decodedMsg, messageMap, &mon)
//	assert.Nil(t, maintenanceMap, tlog.Log()+"should return nil")
//}

// NO LONGER IN USE, IT IS RELATED TO RTC-DOCTOR
//func Test_processMaintenanceInsert_noTag(t *testing.T) {
//	log.Print(tlog.Log())
//	var testData = fmt.Sprintf(sampleMaintenanceMsg1, "")
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	messageMap, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
//	maintenanceMap := ProcessMaintenance(dbConn, decodedMsg, messageMap, &mon)
//	assert.NotNil(t, maintenanceMap, tlog.Log()+"should return nil")
//}

// NO LONGER IN USE, IT IS RELATED TO RTC-DOCTOR
//func Test_processMaintenanceUpdate_DR(t *testing.T) {
//	log.Print(tlog.Log())
//	des := "maintenance short description 1"
//	var testData = fmt.Sprintf(sampleMaintenanceMsg3, des, "2018-09-14T15:34:31Z", "scheduled", "Ready to schedule", "Publish", "")
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	messageMap, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
//	maintenanceMap := ProcessMaintenance(dbConn, decodedMsg, messageMap, &mon)
//	bytes, _ := json.Marshal(maintenanceMap)
//	messageToPost := string(bytes)
//	log.Print(tlog.Log()+" messageToPost: ", messageToPost)
//	assert.NotNil(t, maintenanceMap, tlog.Log()+"Should have maintenanceMap.")
//}

// NO LONGER IN USE, IT IS RELATED TO RTC-DOCTOR
//func Test_processMaintenanceUpdate_CR(t *testing.T) {
//	log.Print(tlog.Log())
//	des := "maintenance short description cr"
//	var testData = fmt.Sprintf(sampleMaintenanceMsg3, des, "2018-09-14T15:34:31Z", "complete", "Deployment Success", "Published - All Updates Completed", "successful")
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	messageMap, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
//	maintenanceMap := ProcessMaintenance(dbConn, decodedMsg, messageMap, &mon)
//	assert.NotNil(t, maintenanceMap, tlog.Log()+"Should have maintenanceMap.")
//}

// NO LONGER IN USE, IT IS RELATED TO RTC-DOCTOR
//func Test_processMaintenanceUpdate_noupdate(t *testing.T) {
//	log.Print(tlog.Log())
//	des := "maintenance short description cr"
//	var testData = fmt.Sprintf(sampleMaintenanceMsg3, des, "2018-09-14T15:34:31Z")
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	messageMap, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
//	maintenanceMap := ProcessMaintenance(dbConn, decodedMsg, messageMap, &mon)
//	assert.Nil(t, maintenanceMap, tlog.Log()+"should return nil.")
//}

// NO LONGER IN USE, IT IS RELATED TO RTC-DOCTOR
//func Test_processMaintenanceUpdate_noupdate_DR(t *testing.T) {
//	log.Print(tlog.Log())
//	des := "maintenance short description cr"
//	var testData = fmt.Sprintf(sampleMaintenanceMsg3, des, "2018-09-10T15:34:31Z")
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	messageMap, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
//	maintenanceMap := ProcessMaintenance(dbConn, decodedMsg, messageMap, &mon)
//	assert.Nil(t, maintenanceMap, tlog.Log()+"should return nil.")
//}

func Test_processSNMaintenance_insert(t *testing.T) {
	os.Setenv("BYPASS_LOCAL_STORAGE", "true")
	defer os.Unsetenv("BYPASS_LOCAL_STORAGE")
	log.Print(tlog.Log())
	if !shared.BypassLocalStorage {
		var testData = fmt.Sprintf(sampleMaintenanceMsg4, "CH000001", "short desc", "aaa")
		dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
		defer db.Disconnect(dbConn)
		_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
		maintenanceMap, notifications := ProcessSNMaintenance(dbConn, decodedMsg, true, &mon)
		assert.NotNil(t, maintenanceMap, tlog.Log()+"Should have maintenanceMap.")
		assert.Equal(t, len(notifications), 0)
	}
}

func Test_processSNMaintenance_insert_oldMap(t *testing.T) {
	os.Setenv("BYPASS_LOCAL_STORAGE", "true")
	defer os.Unsetenv("BYPASS_LOCAL_STORAGE")
	log.Print(tlog.Log())
	if !shared.BypassLocalStorage {
		var testData = fmt.Sprintf(sampleMaintenanceMsg4, "CH000001", "short desc", "aaa")
		dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
		defer db.Disconnect(dbConn)
		resourceMapCache.updateTime = resourceMapCache.updateTime.AddDate(0, -1, 0)
		_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
		maintenanceMap, notifications := ProcessSNMaintenance(dbConn, decodedMsg, true, &mon)
		assert.Equal(t, len(notifications), 0)
		assert.Equal(t, len(maintenanceMap), 0)
	}
}

// Add the following lines in once the ServiceNow team adds publish date to the incomming communications.
//func Test_processSNMaintenance_withCommunication(t *testing.T) {
//	log.Print(tlog.Log())
//	var testData = fmt.Sprintf(sampleMaintenanceMsgWithCommunication)
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
//	maintenanceMap, notifications := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
//	assert.Nil(t, maintenanceMap, tlog.Log()+"Should have maintenanceMap.")
//	assert.Equal(t, 0, len(notifications))
//}

// Add the following lines in once the ServiceNow team adds publish date to the incomming communications.
//func Test_processSNMaintenance_withCommunicationWithPublishDate(t *testing.T) {
//	log.Print(tlog.Log())
//	var testData = fmt.Sprintf(sampleMaintenanceMsgWithCommunicationWithPublishDate)
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
//	maintenanceMap, notifications := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
//	assert.NotNil(t, maintenanceMap, tlog.Log()+"Should have maintenanceMap.")
//	assert.Equal(t, 1, len(notifications))
//}

func Test_processSNMaintenance_withCommunication_nonDisuptive(t *testing.T) {
	log.Print(tlog.Log())
	var testData = fmt.Sprintf(sampleMaintenanceMsgWithCommunicationNonDisruptive)
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
	maintenanceMap, notifications := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
	assert.Equal(t, 0, len(maintenanceMap))
	assert.Equal(t, 0, len(notifications))
}

func Test_processSNMaintenance_update(t *testing.T) {
	log.Print(tlog.Log())
	var testData = fmt.Sprintf(sampleMaintenanceMsg4, "CH000001", "updated short desc", "bbb")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	log.Print(tlog.Log())
	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
	maintenanceMap, notifications := ProcessSNMaintenance(dbConn, decodedMsg, true, &mon)
	assert.NotNil(t, len(maintenanceMap), "should have return")
	assert.Equal(t, len(notifications), 0)
}

func Test_processSNMaintenance_noupdate(t *testing.T) {
	log.Print(tlog.Log())
	var testData = fmt.Sprintf(sampleMaintenanceMsg4, "CH000001", "updated short desc", "bbb")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	log.Print(tlog.Log())
	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
	maintenanceMap, notifications := ProcessSNMaintenance(dbConn, decodedMsg, true, &mon)
	assert.True(t, cmp.Equal(len(maintenanceMap), 0))
	assert.Equal(t, len(notifications), 0)
}

func Test_processSNMaintenance_notBulk(t *testing.T) {
	log.Print(tlog.Log())
	var testData = fmt.Sprintf(sampleMaintenanceMsg4NotBulk, "234568", "updated short desc", "bbb")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	log.Print(tlog.Log())
	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
	maintenanceMap, notifications := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
	assert.True(t, cmp.Equal(len(maintenanceMap), 0))
	assert.Equal(t, len(notifications), 0)

}

// Add the following lines in once the ServiceNow team adds publish date to the incomming communications.
//func Test_processSNMaintenance_notBulk_inResourceMapCache(t *testing.T) {
//	log.Print(tlog.Log())
//	testReturnJSON := `{RecordID: "id",
//	PnpCreationTime:    "create",
//	SourceCreationTime: "sctime",
//	SourceUpdateTime:   "sutime",
//	CRNFull:            "crn:v1:ys0-dallas:public:servicenow:us-south::::",
//	State:              "state",
//	OperationalStatus:  "ostatus",
//	Source:             "servicenow",
//	SourceID:           "sourceID",
//	Status:             "status",
//	StatusUpdateTime:   "stutime",
//	RegulatoryDomain:   "rdom",
//	CategoryID:         "catid",
//	CategoryParent:     true,
//	DisplayNames:       []datastore.DisplayName{datastore.DisplayName{Name: "dname", Language: "lang"}},
//	Visibility:         []string{"vis"},
//	Tags:               []datastore.Tag{datastore.Tag{ID: "tag"}},
//	CatalogParentID:    "catParentID",
//	RecordHash:         "123"}`
//	resourceReturn := new(datastore.ResourceReturn)
//	err := json.Unmarshal([]byte(testReturnJSON), &resourceReturn)
//	if err != nil {
//		log.Print("Error occurred unmarshaling , err = ", err)
//	}
//	tmpMap := make(map[string]*datastore.ResourceReturn)
//	tmpMap["crn:v1:ys0-dallas:public:servicenow:us-south::::"] = resourceReturn
//	resourceMapCache.swap(tmpMap)
//	var testData = fmt.Sprintf(sampleMaintenanceMsgWithCommunication)
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	log.Print(tlog.Log())
//	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
//	maintenanceMap, notifications := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
//	assert.True(t, cmp.Equal(len(maintenanceMap), 0))
//	assert.Equal(t, len(notifications), 0)
//}

//func Test_processSNMaintenance_notBulk_inResourceMapCacheWithPublishDate(t *testing.T) {
//	log.Print(tlog.Log())
//	testReturnJSON := `{RecordID: "id",
//	PnpCreationTime:    "create",
//	SourceCreationTime: "sctime",
//	SourceUpdateTime:   "sutime",
//	CRNFull:            "crn:v1:ys0-dallas:public:servicenow:us-south::::",
//	State:              "state",
//	OperationalStatus:  "ostatus",
//	Source:             "servicenow",
//	SourceID:           "sourceID",
//	Status:             "status",
//	StatusUpdateTime:   "stutime",
//	RegulatoryDomain:   "rdom",
//	CategoryID:         "catid",
//	CategoryParent:     true,
//	DisplayNames:       []datastore.DisplayName{datastore.DisplayName{Name: "dname", Language: "lang"}},
//	Visibility:         []string{"vis"},
//	Tags:               []datastore.Tag{datastore.Tag{ID: "tag"}},
//	CatalogParentID:    "catParentID",
//	RecordHash:         "123"}`
//	resourceReturn := new(datastore.ResourceReturn)
//	err := json.Unmarshal([]byte(testReturnJSON), &resourceReturn)
//	if err != nil {
//		log.Print("Error occurred unmarshaling , err = ", err)
//	}
//	tmpMap := make(map[string]*datastore.ResourceReturn)
//	tmpMap["crn:v1:ys0-dallas:public:servicenow:us-south::::"] = resourceReturn
//	resourceMapCache.swap(tmpMap)
//	var testData = fmt.Sprintf(sampleMaintenanceMsgWithCommunicationWithPublishDate)
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	log.Print(tlog.Log())
//	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
//	maintenanceMap, notifications := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
//	assert.True(t, cmp.Equal(len(maintenanceMap), 0))
//	assert.Equal(t, len(notifications), 1)
//}

func Test_processMaintenanceUpdate_bspn(t *testing.T) {
	log.Print(tlog.Log())
	source := "servicenow"
	sourceID := "882920"
	recordID := db.CreateRecordIDFromSourceSourceID(source, sourceID)
	var testData = fmt.Sprintf(sampleMaintenanceMsg5, recordID, "2018-10-24T11:25:36Z")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	crnFull := []string{"crn:v1:d-wbmdn-28801:dedicated:conversation:us-south::::"}
	sampleMaintenanceInsert := datastore.MaintenanceInsert{
		SourceCreationTime:    "2018-10-02T01:58:42Z",
		SourceUpdateTime:      "2018-10-02T01:58:42Z",
		PlannedStartTime:      "2018-10-22T11:00:00Z",
		PlannedEndTime:        "2018-10-22T13:00:00Z",
		ShortDescription:      "from bspn loader",
		LongDescription:       "test",
		CRNFull:               crnFull,
		State:                 "in-progress",
		Disruptive:            true,
		SourceID:              "882920",
		Source:                "servicenow",
		DisruptionType:        "Other (specify in Description)",
		RegulatoryDomain:      "",
		DisruptionDuration:    60,
		DisruptionDescription: "test",
		MaintenanceDuration:   20,
	}
	recordID, err, statusOk := db.InsertMaintenance(dbConn, &sampleMaintenanceInsert)
	if err != nil {
		log.Print(tlog.Log(), err)
	}
	log.Print(tlog.Log()+"- InsertMaintenance recordID:"+recordID+",status code:", statusOk)
	messageMap, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
	log.Println("DEBUB msg:", msg)
	log.Println("DEBUB messageMap:", messageMap, "decodedMsg:", decodedMsg)
	maintenanceMap, _ := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
	log.Println("maintenanceMap:", maintenanceMap)
	assert.Nil(t, maintenanceMap, tlog.Log()+"Should not have maintenanceMap is not disruptive")
}

func Test_processMaintUpdate_bspn_RecordErr(t *testing.T) {
	log.Print(tlog.Log())
	source := "servicenow"
	sourceID := "882920"
	recordID := db.CreateRecordIDFromSourceSourceID(source, sourceID)
	var testData = fmt.Sprintf(sampleMaintenanceMsg5, recordID+"0", "2018-10-24T11:25:36Z")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
	maintenanceMap, _ := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
	//maintenanceMap := ProcessMaintenance(dbConn, decodedMsg, messageMap, &mon)
	assert.Nil(t, maintenanceMap, tlog.Log()+"Error, no  maintenanceMap")
}

func Test_processMaintUpdate_bspn_noRecord(t *testing.T) {
	log.Print(tlog.Log())
	source := "servicenow"
	sourceID := "882920"
	crn := "crn:v1:bluemix:public:service-name1:us-east::::"
	var testData = fmt.Sprintf(sampleMaintenanceMsg6, "2018-10-24T11:25:36Z", "new", sourceID, source, crn)
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
	maintenanceMap, _ := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
	assert.Nil(t, maintenanceMap, tlog.Log()+"should not have maintenanceMap, ignoring because in new state")
}

func Test_shouldChangeState_inprogress(t *testing.T) {
	log.Print(tlog.Log())
	plannedStartTime := "2018-10-24T11:25:36Z"
	plannedEndTime := "2018-11-11T11:25:36Z"
	var testData = fmt.Sprintf(sampleMaintenanceMsg7, plannedStartTime, plannedEndTime, "new")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
	maintenanceMap, _ := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
	assert.Nil(t, maintenanceMap, tlog.Log()+"should not have maintenanceMap, ignoring because in new state")
}

func Test_shouldChangeState_complete(t *testing.T) {
	log.Print(tlog.Log())
	plannedStartTime := "2018-10-24T11:25:36Z"
	plannedEndTime := "2018-11-08T11:25:36Z"
	var testData = fmt.Sprintf(sampleMaintenanceMsg7, plannedStartTime, plannedEndTime, "new")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
	maintenanceMap, _ := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
	assert.Nil(t, maintenanceMap, tlog.Log()+"should not have maintenanceMap, ignoring because in new state")
}

func Test_shouldChangeState_nochange(t *testing.T) {
	log.Print(tlog.Log())
	plannedStartTime := "2018-10-24T11:25:36Z"
	plannedEndTime := "2018-11-10T11:25:36Z"
	var testData = fmt.Sprintf(sampleMaintenanceMsg7, plannedStartTime, plannedEndTime, "complete")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	_, decodedMsg := DecodeAndMapMaintenceMessage(msg, setupMonitor(t))
	maintenanceMap, _ := ProcessSNMaintenance(dbConn, decodedMsg, false, &mon)
	assert.Nil(t, maintenanceMap, tlog.Log()+"should not have maintenanceMap, is not disruptive")
}

func Test_shouldProcessMaintenance(t *testing.T) {
	log.Print(tlog.Log())
	newSNMaintenanceMap := map[string]*datastore.MaintenanceReturn{}
	testData := `{"record_id":"db4271b7b373e5bceb7c99ee87076db52e2b0c957ca851d307efc931fb7fef21","pnp_creationTime":"2018-11-29T16:00:30-0600","pnp_update_time":"2018-11-29T16:29:52-0600","source_creation_time":"2018-11-26T09:31:14-0600","source_update_time":"2018-11-29T12:10:50-0600","planned_start_time":"2018-11-29T10:00:00-0600","planned_end_time":"2018-11-29T14:00:00-0600","short_description":"des","long_description":"test","record_hash":"11123","state":"complete","disruptive":true,"source_id":"899550","source":"servicenow","crnFull":["crn:v1:bluemix:public:dashdb-for-transactions:au-syd::::"],"maintenance_duration":240,"disruption_type":"Console Access,Existing Service Instances","disruption_description":"test.","disruption_duration":30}`
	maintenanceReturn := new(datastore.MaintenanceReturn)
	err := json.Unmarshal([]byte(testData), &maintenanceReturn)
	if err != nil {
		log.Print("Error occurred unmarshaling , err = ", err)
	}
	newSNMaintenanceMap[maintenanceReturn.Source+maintenanceReturn.SourceID] = maintenanceReturn
	snMaintenanceMap.SNMaintenances = newSNMaintenanceMap
	log.Printf("snMaintenanceMap Length: %d ", len(snMaintenanceMap.SNMaintenances))
	maintenanceInsert := maintReturnToMaintInsert(maintenanceReturn)
	maintenanceInsert.SourceUpdateTime = "2018-11-29 12:10:51-0600"
	shouldProcess := shouldProcessMaintenance(&maintenanceInsert, "23456")
	assert.Equal(t, shouldProcess, true)

	shouldProcess = shouldProcessMaintenance(&maintenanceInsert, "11123")
	assert.Equal(t, shouldProcess, false)

	testData2 := `{"record_id":"db4271b7b373e5bceb7c99ee87076db52e2b0c957ca851d307efc931fb7fef21","pnp_creationTime":"2018-11-29T16:00:30-06:00","pnp_update_time":"2018-11-29T16:29:52-06:00","source_creation_time":"2018-11-26T09:31:14-06:00","source_update_time":"2018-11-29T12:10:50-06:00","planned_start_time":"2018-11-29T10:00:00-06:00","planned_end_time":"2018-11-29T14:00:00-06:00","short_description":"des","long_description":"test","record_hash":"11123","state":"complete","disruptive":true,"source_id":"899549","source":"servicenow","crnFull":["crn:v1:bluemix:public:dashdb-for-transactions:au-syd::::"],"maintenance_duration":240,"disruption_type":"Console Access,Existing Service Instances","disruption_description":"test.","disruption_duration":30}`
	err = json.Unmarshal([]byte(testData2), &maintenanceReturn)
	if err != nil {
		log.Print("Error occurred unmarshaling , err = ", err)
	}
	maintenanceInsert = maintReturnToMaintInsert(maintenanceReturn)
	shouldProcess = shouldProcessMaintenance(&maintenanceInsert, "")
	assert.Equal(t, shouldProcess, true)
}

func Test_getCrUpdateTime(t *testing.T) {
	log.Print(tlog.Log())
	// Getting back the cr_update_time
	retTime, err := getCrUpdateTime([]byte(sampleChangePayloadCrIsNewer))
	if err != nil {
		log.Fatal("Test_getCrUpdateTime: ", err)
	}
	assert.True(t, retTime == "2019-08-14T18:24:54Z")

	retTime, err = getCrUpdateTime([]byte(sampleChangePayloadSourceIsNewer))
	if err != nil {
		log.Fatal("Test_getCrUpdateTime: ", err)
	}
	assert.True(t, retTime == "")

	retTime, err = getCrUpdateTime([]byte(sampleChangePayloadCrIsEmpty))
	if err != nil {
		log.Fatal("Test_getCrUpdateTime: ", err)
	}
	assert.True(t, retTime == "")
}

func setupMonitor(t *testing.T) *ossmon.OSSMon {
	nrConfig := newrelic.NewConfig("", "")
	nrConfig.Enabled = false
	nrApp, err := newrelic.NewApplication(nrConfig)
	if err != nil {
		t.Log(err)
	}
	mon := ossmon.OSSMon{
		NewRelicApp: nrApp,
		Sensor:      instana.NewSensor("testing-pnp-nq2ds"),
	}
	return &mon
}
