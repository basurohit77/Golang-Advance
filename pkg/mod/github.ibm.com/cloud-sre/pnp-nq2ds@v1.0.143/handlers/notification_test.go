package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	_ "net/http"
	"os"
	"strconv"
	"testing"
	_ "time"

	"github.ibm.com/cloud-sre/oss-globals/tlog"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	"github.ibm.com/cloud-sre/pnp-nq2ds/testutils"
)

var (
	//context    ctxt.Context

	notification = `{
		"source_creation_time":"2018-09-04 14:43:02Z",
		"source_update_time":"2018-09-04 14:43:02Z",
		"event_time_start":"2018-09-04 14:43:02Z",
		"event_time_end":"2018-09-04 14:43:02Z",
		"source":"%s",
		"source_id":"%s",
		"type":"maintenance",
		"category":"testCategory",
		"incident_id":"1234",
		"crn_full":"crn:v1:bluemix:public:testService:eu-gb::::",
		"resource_display_names":[{"name":"resourceName","language":"en"}],
		"short_description":[{"name":"%s","language":"en"}],
		"long_description":[{"name":"long desc","language":"en"}],
		"maintenance_duration": 60,
		"disruption_duration": 50,
		"disruption_type": "disruptive type",
		"disruption_description" : "disruption description"
		}`

	notificationReturnMsg = `{
		"record_id":"36177e26a754999875131ecbc818da01977dd87103825e30789418551848aecd",
		"pnp_creation_time":"2018-09-04 14:43:02",
		"pnp_update_time":"2018-09-04 14:43:02",
		"source_creation_time":"2018-09-04 14:43:02",
		"source_update_time":"2018-09-04 14:43:02",
		"event_time_start":"2018-09-04 14:43:02",
		"event_time_end":"2018-09-04 14:43:02",
		"source":"servicenow",
		"source_id":"123456",
		"type":"testType",
		"category":"testCategory",
		"incident_id":"1234",
		"crn_full":"crn:v1:bluemix:public:testService:eu-gb::::",
		"resource_display_names":[{"name":"resourceName","language":"en"}],
		"short_description":[{"name":"updated short desc","language":"en"}],
		"long_description":[{"name":"long desc","language":"en"}],
		"maintenance_duration": 60,
		"disruption_duration": 50,
		"disruption_type": "disruptive type",
		"disruption_description" : "disruption description"
		}`

	notificationReturnMsg2 = `{
		"record_id":"36177e26a754999875131ecbc818da01977dd87103825e30789418551848aecd",
		"pnp_creation_time":"2018-09-04 14:43:02",
		"pnp_update_time":"2018-09-04 14:43:02",
		"source_creation_time":"2018-09-04 14:43:02",
		"source_update_time":"2018-09-04 14:43:02",
		"event_time_start":"%s",
		"event_time_end":"%s",
		"source":"servicenow",
		"source_id":"123456",
		"type":"testType",
		"category":"testCategory",
		"incident_id":"1234",
		"crn_full":"crn:v1:bluemix:public:myservice1:eu-gb::::",
		"resource_display_names":[{"name":"","language":"en"}],
		"short_description":[{"name":"updated short desc","language":"en"}],
		"long_description":[{"name":"test","language":"en"}]
		}`

	bspnNotification = `{
		"msgtype":"update",
		"source_update_time":"2018-09-04 14:43:02Z",
		"event_time_start":"2018-09-04 14:43:02Z",
		"event_time_end":"2018-09-04 14:43:02Z",
		"source":"servicenow",
		"source_id":"BSP00001",
		"type":"testType",
		"category":"incident",
		"incident_id":"12345",
		"crn_full":"crn:v1:bluemix:public:testService:eu-gb::::",
		"resource_display_names":[{"name":"resourceName","language":"en"}],
		"short_description":[{"name":"short desc","language":"en"}],
		"long_description":[{"name":"long desc","language":"en"}]
		}`

	releaseNoteNotification = `{
		"source_creation_time":"2018-09-04 14:43:02Z",
		"source_update_time":"2018-09-04 14:43:02Z",
		"event_time_start":"2018-09-04 14:43:02Z",
		"event_time_end":"2018-09-04 14:43:02Z",
		"source":"ghost",
		"source_id":"appid-Jan2122",
		"type":"release-note",
		"category":"services",
		"incident_id":"",
		"crn_full":"crn:v1:bluemix:public:appid:us-east::::",
		"resource_display_names":[{"name":"resourceName","language":"en"}],
		"short_description":[{"name":"New region availability","language":"en"}],
		"long_description":[{"name":"\n <dt>New region availability</dt>\n <dd>As of 27 September 2021, App ID is now available in the Sao Paulo region. For a detailed list of the regions in which the service is available, see <a href=\"/docs/appid?topic=appid-regions-endpoints\">Regions and endpoints</a>.</dd>\n","language":"en"}],
        "release_note_url": "/docs/appid?topic=appid-release-notes"
		}`
)

func Test_NotificationWithDecryptError(t *testing.T) {
	log.Println(tlog.Log())
	msg := `{"test":"test"}`
	dbConn, _, mon := testutils.PrepareTestInc(t, msg)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	notification, err := ProcessNotification(dbConn, []byte(msg), &mon)
	assert.NotNil(t, err, tlog.Log()+"Should have decryption error.")
	assert.Nil(t, notification, tlog.Log()+"Notification should be nil")
}

func Test_NotificationWithDecoderError(t *testing.T) {
	log.Println(tlog.Log())
	notificationMsg := `{"test":"test", "crn_full":"crn:v1:bluemix:public:testService:eu-gb::::"`
	dbConn, msg, mon := testutils.PrepareTestInc(t, notificationMsg)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	notification, err := ProcessNotification(dbConn, msg, &mon)
	assert.NotNil(t, err, tlog.Log()+"Should have decoder error.")
	assert.Nil(t, notification, tlog.Log()+"Notification should be nil")
}

func Test_ProcessNotification(t *testing.T) {
	log.Println(tlog.Log())
	var testData = fmt.Sprintf(notification, testutils.Source, testutils.SourceID, "short desc")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	notification, err := ProcessNotification(dbConn, msg, &mon)
	assert.Nil(t, err, tlog.Log()+"Should not have any error.")
	assert.NotNil(t, notification, tlog.Log()+"Notification should not be nil")
}

func Test_ProcessSNNotification(t *testing.T) {
	log.Println(tlog.Log())
	var testData = fmt.Sprintf(notification, testutils.Source, testutils.SourceID, "short desc")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	var b bytes.Buffer
	log.SetOutput(&b)
	notification, err := ProcessNotification(dbConn, msg, &mon)
	log.SetOutput(os.Stderr)
	expected := `long desc<br/><b>Maintenance Duration</b>: 60 minutes<br/><br/><b>Disruption Duration</b>: 50 minutes`
	assert.Contains(t, b.String(), expected)
	assert.Nil(t, err, tlog.Log()+"Should not have any error.")
	assert.NotNil(t, notification, tlog.Log()+"Notification should not be nil")
}

// REMOVED RTC-DOCTOR TESTS
//func Test_ProcessRTCNotification(t *testing.T) {
//	log.Println(tlog.Log())
//	var testData = fmt.Sprintf(notification, "Doctor-RTC", testutils.SourceID, "short desc")
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//	defer db.Disconnect(dbConn)
//	t.Logf(tlog.Log())
//	var b bytes.Buffer
//	log.SetOutput(&b)
//	notification, err := ProcessNotification(dbConn, msg, &mon)
//	log.SetOutput(os.Stderr)
//	expected := `<b>Update Description</b>: long desc<br/><br/><b>Maintenance Duration</b>: 60 minutes<br/><br/><b>Type of Disruption</b>: disruptive type<br/><br/><b>Disruption Description</b>: disruption description<br/><br/><b>Disruption Duration</b>: 50 minutes`
//	assert.Contains(t, b.String(), expected)
//	assert.Nil(t, err, tlog.Log()+"Should not have any error.")
//	assert.NotNil(t, notification, tlog.Log()+"Notification should not be nil")
//}

//func Test_ProcessRTCNotification(t *testing.T) {
//	const FCT = "Test process notification from servicenow: "
//	var testData = fmt.Sprintf(notification, "Doctor-RTC", testutils.SourceID, "short desc")
//	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
//
//	defer db.Disconnect(dbConn)
//
//	t.Logf(FCT)
//	var b bytes.Buffer
//	log.SetOutput(&b)
//
//	notification, err := ProcessNotification(dbConn, msg, &mon)
//	log.SetOutput(os.Stderr)
//
//	expected := `<b>Update Description</b>: long desc<br/><br/><b>Maintenance Duration</b>: 60 minutes<br/><br/><b>Type of Disruption</b>: disruptive type<br/><br/><b>Disruption Description</b>: disruption description<br/><br/><b>Disruption Duration</b>: 50 minutes`
//	assert.Contains(t, b.String(), expected)
//	assert.Nil(t, err, "Should not have any error.")
//	assert.NotNil(t, notification, "Notification should not be nil")
//}

func Test_BSPNNotification(t *testing.T) {
	log.Println(tlog.Log())
	dbConn, msg, mon := testutils.PrepareTestInc(t, bspnNotification)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	notification, err := ProcessNotification(dbConn, msg, &mon)
	assert.Nil(t, err, tlog.Log()+"Should not have any error.")
	assert.NotNil(t, notification, tlog.Log()+"Notification should not be nil")
}

func Test_hasValidIncidentID(t *testing.T) {
	log.Println(tlog.Log())
	dbConn, _, _ := testutils.PrepareTestInc(t, "")
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	notification := new(datastore.NotificationMsg)
	if err := json.NewDecoder(bytes.NewReader([]byte(bspnNotification))).Decode(notification); err != nil {
		assert.Fail(t, "Failed to decode message ", err)
	}
	notification.IncidentID = "INC12345"
	err := hasValidIncidentID(dbConn, notification)
	if err != nil {
		log.Print(tlog.Log()+"Should have an error. :", err)
	}
	// insert an incident
	crnFull := []string{"crn:v1:watson:public:testservice:location1::::"}
	incident := datastore.IncidentInsert{
		SourceCreationTime:        "2018-07-07T22:01:01Z",
		SourceUpdateTime:          "2018-07-07T22:01:01Z",
		OutageStartTime:           "2018-07-07T21:55:30Z",
		ShortDescription:          "incident short description 1",
		LongDescription:           "[targeted notification](https://url.to.parse) \nincident long description 1",
		State:                     "new",
		Classification:            "confirmed-cie",
		Severity:                  "1",
		CRNFull:                   crnFull,
		SourceID:                  "INC12345",
		Source:                    "servicenow",
		RegulatoryDomain:          "regulatory domain 1",
		AffectedActivity:          "affected activity 1",
		CustomerImpactDescription: "customer impact description 1",
	}
	_, err, _ = db.InsertIncident(dbConn, &incident)
	if err != nil {
		log.Fatal(tlog.Log()+"Should not have an error. :", err)
	}
	err = hasValidIncidentID(dbConn, notification)
	if err != nil {
		log.Fatal(tlog.Log()+"Should not have an error. :", err)
	}
	// Now test changes
	notification.IncidentID = "CHG12345"
	maintenance := datastore.MaintenanceInsert{
		SourceCreationTime:    "2018-07-07T22:01:01Z",
		SourceUpdateTime:      "2018-07-07T22:01:01Z",
		PlannedStartTime:      "2018-07-07T21:55:30Z",
		PlannedEndTime:        "2018-07-21T21:55:30Z",
		ShortDescription:      "maintenance short description 1",
		LongDescription:       "[targeted notification](https://url.to.parse) \nmaintenance long description 1",
		State:                 "new",
		Disruptive:            true,
		CRNFull:               crnFull,
		SourceID:              "CHG12345",
		Source:                "servicenow",
		MaintenanceDuration:   240,
		DisruptionType:        "Running Applications,Application Management (start/stop/stage/etc.)",
		DisruptionDescription: "During this change, there might be occasional intermittent",
		DisruptionDuration:    20,
		RecordHash:            db.CreateRecordIDFromString("Test 1"),
		CompletionCode:        "completion code 1",
	}
	_, err, _ = db.InsertMaintenance(dbConn, &maintenance)
	if err != nil {
		log.Fatal(tlog.Log()+"Should not have an error. :", err)
	}
	err = hasValidIncidentID(dbConn, notification)
	if err != nil {
		log.Fatal(tlog.Log()+"Should not have an error. :", err)
	}
}

func Test_ExistingNotification(t *testing.T) {
	log.Println(tlog.Log())
	var testData = fmt.Sprintf(notification, testutils.Source, testutils.SourceID, "updated short desc")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	notification, err := ProcessNotification(dbConn, msg, &mon)
	assert.Nil(t, err, tlog.Log()+"Should not have any error.")
	assert.NotNil(t, notification, tlog.Log()+"Notification should not be nil")
}

func Test_noteExistsInDB(t *testing.T) {
	log.Println(tlog.Log())
	notificationMsg := `{"test":"test", "crn_full":"crn:v1:bluemix:public:testService:eu-gb::::"}`
	dbConn, _, _ := testutils.PrepareTestInc(t, notificationMsg)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	notifInsert := new(datastore.NotificationInsert)
	err := json.Unmarshal([]byte(notificationMsg), &notifInsert)
	if err != nil {
		log.Print(tlog.Log()+"Test_produceNotificationToMQ Unmarshal failed", err)
	}
	exist, sourceUpdateTime := noteExistsInDB(dbConn, notifInsert)
	assert.Equal(t, exist, false)
	assert.Equal(t, sourceUpdateTime, "")

}

func Test_addCount(t *testing.T) {
	log.Println(tlog.Log())
	addCount(false)
	addCount(true)
	t.Log("Test_addCount finished..")
}

func Test_produceNotificationToMQ(t *testing.T) {
	log.Println(tlog.Log())
	var testData = fmt.Sprintf(notification, testutils.Source, testutils.SourceID, "short desc")
	notifInsert := new(datastore.NotificationInsert)
	err := json.Unmarshal([]byte(testData), &notifInsert)
	if err != nil {
		log.Print(tlog.Log()+"Test_produceNotificationToMQ Unmarshal failed", err)
	}
	RabbitmqURLs = []string{"amqp://guest:guest@localhost:5672"}
	NotificationRoutingKey = "nq2ds.notification"
	MQExchangeName = "pnp.direct"
	notificationMsg := datastore.NotificationMsg{NotificationInsert: *notifInsert}
	produceNotificationToMQ(&notificationMsg)
}

func Test_isNewNotificationNeeded(t *testing.T) {
	log.Println(tlog.Log())
	notificationReturn := new(datastore.NotificationReturn)
	err := json.Unmarshal([]byte(notificationReturnMsg), &notificationReturn)
	if err != nil {
		log.Print(tlog.Log()+"Test_isNewNotificationNeeded Unmarshal failed", err)
	}
	isNew := isNewNotificationNeeded(notificationReturn, updateIncident)
	assert.True(t, cmp.Equal(isNew, true))
	var testData = fmt.Sprintf(notificationReturnMsg2, tnow, tnow)
	err = json.Unmarshal([]byte(testData), &notificationReturn)
	if err != nil {
		log.Print(tlog.Log()+"Test_isNewNotificationNeeded Unmarshal failed", err)
	}
	isNew = isNewNotificationNeeded(notificationReturn, updateIncident)
	assert.True(t, cmp.Equal(isNew, false))
}

func Test_getCategoryID(t *testing.T) {
	log.Println(tlog.Log())
	catID := getCategoryID(ctxt.Context{}, "")
	t.Log(catID)
	InitNotificationsAdapter()
	catID = getCategoryID(ctxtContext, "cloud-object-storage")
	assert.True(t, cmp.Equal(catID, ""))
}

func Test_getDisplayName(t *testing.T) {
	log.Println(tlog.Log())
	disPlayName := getDisplayName("cloud-object-storage")
	assert.True(t, cmp.Equal(disPlayName, ""))
}

func Test_incidentToNotification(t *testing.T) {
	log.Println(tlog.Log())
	notification := new(datastore.NotificationInsert)
	notification = incidentToNotification(updateIncident, "", "crn:v1:internal:public:myservice1:us-east::::")
	log.Print("Test_incidentToNotification:", notification)
}

func Test_checkNotificationForIncident(t *testing.T) {
	log.Println(tlog.Log())
	var (
		pgHost   = os.Getenv("PG_DB_IP")
		pgDB     = os.Getenv("PG_DB")
		pgDBUser = os.Getenv("PG_DB_USER")
		pgPass   = os.Getenv("PG_DB_PASS")
		pgPort   = os.Getenv("PG_DB_PORT")
	)
	pgPortInt, _ := strconv.Atoi(pgPort)
	dbConn, err := db.Connect(pgHost, pgPortInt, pgDB, pgDBUser, pgPass, "disable")
	if err != nil {
		log.Print(err)
	}
	defer db.Disconnect(dbConn)
	setEnv()
	InitNotificationsAdapter()
	cache, err := osscatalog.NewCache(ctxtContext, tagNameListingServer)
	if err != nil || cache == nil {
		t.Fatal(tlog.Log()+"Unable to get a cache", err)
	}
	log.Print(tlog.Log(), *updateIncident.CRNFull)
	testCrn := []string{"crn:v1:bluemix:public:testService:eu-gb::::"}
	sourceId := "testNo001"

	inFromQueue := &incidentFromQueue{
		SourceCreationTime:        &tnow,
		SourceUpdateTime:          &tnow,
		OutageStartTime:           &tnow,
		OutageEndTime:             &tnow,
		ShortDescription:          &desc,
		LongDescription:           &desc,
		State:                     &state,
		Classification:            &classification,
		Severity:                  &priority,
		CRNFull:                   &testCrn,
		ServiceName:               &serviceName,
		SourceID:                  &sourceId,
		Source:                    &testutils.Source,
		RegulatoryDomain:          &reDomain,
		AffectedActivity:          &aa,
		CustomerImpactDescription: &ciDes,
	}
	resourceToInsert := datastore.ResourceInsert{}
	resourceToInsert.CRNFull = "crn:v1:bluemix:public:testService:eu-gb::::"
	resourceToInsert.Source = "servicenow"
	resourceToInsert.SourceID = "testNo001"
	resRecordId, err, _ := db.InsertResource(dbConn, &resourceToInsert)
	log.Print("resRecordId: ", resRecordId)
	if err != nil {
		log.Panic("Error inserting resource: ", err)
	}
	checkNotificationForIncident(dbConn, inFromQueue, "")

	notifInsertMsg := `{
		"source_creation_time":"2018-09-04 14:43:02Z",
		"source_update_time":"2018-09-04 14:43:02Z",
		"event_time_start":"2018-09-04 14:43:02Z",
		"event_time_end":"2018-09-04 14:43:02Z",
		"source":"servicenow",
		"source_id":"testNo001",
		"type":"testType",
		"category":"testCategory",
		"incident_id":"1234",
		"crn_full":"crn:v1:bluemix:public:MyService1:eu-gb::::",
		"resource_display_names":[{"name":"resourceName","language":"en"}],
		"short_description":[{"name":"%s","language":"en"}],
		"long_description":[{"name":"long desc","language":"en"}]
		}`

	var testData = fmt.Sprintf(notifInsertMsg, "short desc")
	notificationInsrt := new(datastore.NotificationInsert)
	err = json.Unmarshal([]byte(testData), &notificationInsrt)
	if err != nil {
		log.Print(tlog.Log()+"Error occurred unmarshaling , err = ", err)
	}
	recordID, _, _ := db.InsertNotification(dbConn, notificationInsrt)
	log.Print(tlog.Log(), recordID)
	notifcationRetun, err, httpcode := db.GetNotificationByRecordID(dbConn, recordID)
	log.Print(tlog.Log(), notifcationRetun.CRNFull)
	log.Print(tlog.Log(), err)
	log.Print(tlog.Log(), httpcode)
	checkNotificationForIncident(dbConn, inFromQueue, "")
}

func TestIsIncidentCreationTimeBeforeIaaSBSPNCutover(t *testing.T) {
	log.Println(tlog.Log())
	// nil incident (expect true until after the cutover):
	result := isIncidentCreationTimeBeforeIaaSBSPNCutover(nil)
	assert.True(t, cmp.Equal(result, false))
	// nil source creation time (expect true until after the cutover):
	incident := new(incidentFromQueue)
	result = isIncidentCreationTimeBeforeIaaSBSPNCutover(incident)
	assert.True(t, cmp.Equal(result, false))
	// empty source creation time (expect true until after the cutover):
	creationTime := ""
	incident = new(incidentFromQueue)
	incident.SourceCreationTime = &creationTime
	result = isIncidentCreationTimeBeforeIaaSBSPNCutover(incident)
	assert.True(t, cmp.Equal(result, false))
	// "Z" source creation time (expect true until after the cutover):
	creationTime = "Z" // note that "Z" is added at end by nq2ds code
	incident = new(incidentFromQueue)
	incident.SourceCreationTime = &creationTime
	result = isIncidentCreationTimeBeforeIaaSBSPNCutover(incident)
	assert.True(t, cmp.Equal(result, false))
	// Check right before cutover:
	creationTime = "2019-05-28 13:59:59" + "Z"
	incident = new(incidentFromQueue)
	incident.SourceCreationTime = &creationTime
	result = isIncidentCreationTimeBeforeIaaSBSPNCutover(incident)
	assert.True(t, cmp.Equal(result, true))
	// Check right at cutover:
	creationTime = "2019-05-28 14:00:00" + "Z"
	incident = new(incidentFromQueue)
	incident.SourceCreationTime = &creationTime
	result = isIncidentCreationTimeBeforeIaaSBSPNCutover(incident)
	assert.True(t, cmp.Equal(result, false))
	// Check right after cutover:
	creationTime = "2019-05-28 14:00:01" + "Z"
	incident = new(incidentFromQueue)
	incident.SourceCreationTime = &creationTime
	result = isIncidentCreationTimeBeforeIaaSBSPNCutover(incident)
	assert.True(t, cmp.Equal(result, false))
}

func Test_notificationForIncidentUpdate(t *testing.T) {
	log.Println(tlog.Log())
	var (
		pgHost   = os.Getenv("PG_DB_IP")
		pgDB     = os.Getenv("PG_DB")
		pgDBUser = os.Getenv("PG_DB_USER")
		pgPass   = os.Getenv("PG_DB_PASS")
		pgPort   = os.Getenv("PG_DB_PORT")
	)
	pgPortInt, _ := strconv.Atoi(pgPort)
	dbConn, err := db.Connect(pgHost, pgPortInt, pgDB, pgDBUser, pgPass, "disable")
	if err != nil {
		log.Print(err)
	}
	defer db.Disconnect(dbConn)
	setEnv()
	InitNotificationsAdapter()
	cache, err := osscatalog.NewCache(ctxtContext, tagNameListingServer)
	if err != nil || cache == nil {
		t.Fatal("Unable to get a cache", err)
	}
	// create a notification to test with
	notifInsertMsg := `{
		"source_creation_time":"2018-09-04 14:43:02Z",
		"source_update_time":"2018-09-04 14:43:02Z",
		"event_time_start":"2018-09-04 14:43:02Z",
		"event_time_end":"2018-09-04 14:43:02Z",
		"source":"servicenow",
		"source_id":"testNo002",
		"type":"testType",
		"category":"testCategory",
		"incident_id":"1234",
		"crn_full":"crn:v1:bluemix:public:MyService1:eu-gb::::",
		"resource_display_names":[{"name":"resourceName","language":"en"}],
		"short_description":[{"name":"%s","language":"en"}],
		"long_description":[{"name":"long desc","language":"en"}]
		}`

	var testData = fmt.Sprintf(notifInsertMsg, "short desc")
	notificationInsrt := new(datastore.NotificationInsert)
	err = json.Unmarshal([]byte(testData), &notificationInsrt)
	if err != nil {
		log.Print(tlog.Log()+"Error occurred unmarshaling , err = ", err)
	}
	recordID, _, _ := db.InsertNotification(dbConn, notificationInsrt)
	log.Print(tlog.Log()+"Record ID of inserted notification = ", recordID)

	testCrn := []string{"crn:v1:bluemix:public:testService:eu-gb::::"}
	sourceId := "1234"
	priority := "2"
	// update the notification for incident in queue
	inFromQueue := &incidentFromQueue{
		SourceCreationTime:        &tnow,
		SourceUpdateTime:          &tnow,
		OutageStartTime:           &tnow,
		OutageEndTime:             &tnow,
		ShortDescription:          &desc,
		LongDescription:           &desc,
		State:                     &state,
		Classification:            &classification,
		Severity:                  &priority,
		CRNFull:                   &testCrn,
		ServiceName:               &serviceName,
		SourceID:                  &sourceId,
		Source:                    &testutils.Source,
		RegulatoryDomain:          &reDomain,
		AffectedActivity:          &aa,
		CustomerImpactDescription: &ciDes,
	}
	notificationForIncidentUpdate(dbConn, inFromQueue, true, "")
}

func Test_releaseNoteNotification(t *testing.T) {
	log.Println(tlog.Log())
	dbConn, msg, mon := testutils.PrepareTestInc(t, releaseNoteNotification)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	notification, err := ProcessNotification(dbConn, msg, &mon)
	assert.Nil(t, err, tlog.Log()+"Should not have any error.")
	assert.NotNil(t, notification, tlog.Log()+"Notification should not be nil")
}