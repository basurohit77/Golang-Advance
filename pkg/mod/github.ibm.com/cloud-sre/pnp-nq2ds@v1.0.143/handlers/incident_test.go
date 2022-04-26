package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
	"github.ibm.com/cloud-sre/pnp-nq2ds/shared"
	"log"
	. "os"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	newrelic "github.com/newrelic/go-agent"
	"github.com/stretchr/testify/assert"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/ossmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-abstraction/exmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/cloudant"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/globalcatalog"
	"github.ibm.com/cloud-sre/pnp-nq2ds/testutils"
)

var (
	nrConfig   newrelic.Config
	nrApp      newrelic.Application
	updateTime = "2018-09-04 14:43:02"
	desc       = "test short desc"
	status     = "Confirmed CIE"
	priority   = "1"
	crn        = "[\"crn:v1:bluemix:public:testService:eu-gb::::\"]"
	audience   = "Public"

	incidentMsg = `{
		"sys_created_on":"2018-09-04 14:43:02",
		"sys_updated_on":"%s",
		"u_disruption_began":"2018-09-04 14:40:02",
		"u_disruption_ended":"2018-09-04 14:50:02",
		"short_description":"test short desc",
		"description":"This is test long description.",
		"u_current_status":"current status",
		"incident_state":"new",
		"u_affected_activity":"Login",
		"u_description_customer_impact":"test",
		"cmdb_ci":"testService",
		"u_status":"%s",
		"priority":"%s",
		"crn":%s,
		"short_description":"%s",
		"number":"%s",
		"u_audience":"%s"
		}`

	current        = time.Now()
	tnow           = current.UTC().Format("2006-01-02T15:04:05Z")
	state          = "new"
	crnFull        = []string{"crn:v1:internal:public:myservice1:us-east::::", "crn:v1:internal:public:myservice1:eu-gb::::"}
	classification = "confirmed-cie"
	reDomain       = "reg domain data 123"
	aa             = "Login"
	ciDes          = "test"
	//ciDes       = "Resolved BSPN Posting from NRE Mgmt:\r\n\r\nAt 07 December 06:48 UTC, Network Specialists identified multiple Network System generating over heating alarms in DC SAO01 Datacenter. Network specialists received confirmation from data center staff that the issue is with the Cooling System failing in the data center. This overheating lead to the network devices in the facility shutting down. The Facility Provider worked to correct the problems with the cooling system and Network Specialists then worked to recover the networking devices. At 7 December 2020 21:50 UTC Network Specialists had confirmed Frontend, Backend and the Services network devices had all been restored.\r\n\r\n- - -\r\n\r\nAt 7 December 2020 21:50 UTC Network Specialists have confirmed Frontend, Backend and the Services network devices have all been restored. Network Specialists continue to work with internal teams on any outstanding issues. At this time the networking portion of the incident is being considered resolved.\r\n\r\nThe core network and services network have been restored. Network Specialists are working to restore the backend network now. \r\n\r\nTemperature within the Sao Paulo 01 DC has stabilized. Data center recovery is progressing and network connectivity is being restored first and then rest of the services. As services are brought up, Customers will see intermittent connectivity issues until the operations team completes the recovery efforts.\r\n\r\nThe DC facilities team has confirmed temperatures are normalize and data center staff is now safe to enter the facility. Network specialists have begun working to restore network connectivity starting with the core network routers in the facility.\r\n\r\nIBM Cloud Datacenter engineers have reported a sharp drop in temperatures since the last update. Work is continuing, in conjunction with the facility provider, to mitigate any effects that the issue may have upon IBM Cloud customer services. The Facility Provider found a break in the main water supply feeding the cooling towers. The break has now been repaired and tanks are being re-filled to restore service of the cooling infrastructure. As tanks are undergoing the refill process, temperatures are continuing to decrease and normalize. Facility provider engineers are still reporting an ETA of 1 hours to stabilize temperatures at which point IBM Cloud service teams can begin any additional restore\\verification activities to services impacted. This is a current estimation and may change based on changes in situation.\r\n\r\nNetwork Specialists are currently engaging with all the hardware vendors for further troubleshooting once temperatures in the data center fully recover. We are waiting on data center facilities to confirm once the temperature inside the data center returns to normal. At this time, the estimate is 30 minutes before data center technicians will be able to enter the facility and begin a physical check of devices.\r\n\r\nThe DC Facilities Team has confirmed that due to the fire risk presented by the high temperatures at the data center, they will be shutting down the SAO01 DC network devices and customer servers until the cooling system issue is repaired. DC Facilities Team is currently checking for the approvals for this shutdown. More updates to follow.\r\n\r\nThe DC facility provider installed a secondary Cooling system to cool down the gears and the temperature is still round the 54 degree which is still high. At this time, there have been reports of impacts to storage (file, block, and VSI-SAN), Bare Metal, and VSI offerings. DC facilities team is still working on further Cooling the systems. More update to follow.\r\n\r\nAt 07 December 06:48 UTC, Network Specialists identified multiple System Over heating alarms on DC SAO 01 Datacenter. At 07:25  UTC, Issue was reported to DC-Ops Team. At 07 December 07:27 UTC, Network specialists received confirmation from DC-Ops that the issue is with the Cooling System failing in the Data center. Currently Datacenter Facility confirmed that they have added cold water to the Cooling system to mitigate the issue."
	serviceName = "myservice1"

	updateIncident = &incidentFromQueue{
		SourceCreationTime:        &tnow,
		SourceUpdateTime:          &tnow,
		OutageStartTime:           &tnow,
		OutageEndTime:             &tnow,
		ShortDescription:          &desc,
		LongDescription:           &desc,
		State:                     &state,
		Classification:            &classification,
		Severity:                  &priority,
		CRNFull:                   &crnFull,
		ServiceName:               &serviceName,
		SourceID:                  &testutils.SourceID,
		Source:                    &testutils.Source,
		RegulatoryDomain:          &reDomain,
		AffectedActivity:          &aa,
		CustomerImpactDescription: &ciDes,
		Audience:                  &audience,
	}

	msgExistingCase = `{
		"number":"123456",
		"sys_id":"000000",
		"operation":"",
		"Process":""
		}`
)

func Test_IncidentWithDecryptError(t *testing.T) {
	log.Println(tlog.Log())
	var msgBadEncryption = `{
			"number":"123456"
	}`

	testutils.RecreateTable = true
	dbConn, _, mon := testutils.PrepareTestInc(t, msgBadEncryption)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	_, _, isBadMessage := ProcessIncident(dbConn, []byte(msgBadEncryption), &mon)
	assert.True(t, isBadMessage, tlog.Log()+"Should have decryption error.")
}

func Test_IncidentWithUnmashalError(t *testing.T) {
	log.Println(tlog.Log())
	var msgUnmarshError = `{
		"number":"123456",
		}`

	dbConn, msg, mon := testutils.PrepareTestInc(t, msgUnmarshError)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	_, _, isBadMessage := ProcessIncident(dbConn, msg, &mon)
	assert.True(t, isBadMessage, tlog.Log()+"Should have unmarshal error.")
}

func Test_IncidentWithEmptySourceId(t *testing.T) {
	log.Println(tlog.Log())
	var testData = fmt.Sprintf(incidentMsg, updateTime, status, priority, crn, desc, "", "Private")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	_, _, isBadMessage := ProcessIncident(dbConn, msg, &mon)
	assert.True(t, isBadMessage, tlog.Log()+"SourceId is empty, should return error.")
}

func Test_IncidentWithInvalidTimestamp(t *testing.T) {
	log.Println(tlog.Log())
	var testData = fmt.Sprintf(incidentMsg, "timeNow", status, priority, crn, desc, testutils.SourceID, "Public")
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	_, _, isBadMessage := ProcessIncident(dbConn, msg, &mon)
	assert.True(t, isBadMessage, tlog.Log()+"A timestamp is invalid, should return error.")
}

func Test_getExistingIncident(t *testing.T) {
	log.Println(tlog.Log())
	setEnv()

	cache, err := osscatalog.NewCache(ctxt.Context{}, tagNameListingServer)
	if err != nil || cache == nil {
		t.Fatal("Unable to get a cache", err)
	}

	var updateTimeNow = time.Now().UTC().Format("2006-01-02 15:04:05")
	var updatedDesc = "updated description"
	var testData = fmt.Sprintf(incidentMsg, updateTimeNow, status, priority, crn, updatedDesc, testutils.SourceID, audience)
	dbConn, _, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	recordID := db.CreateRecordIDFromSourceSourceID("servicenow", testutils.SourceID)
	span, ctx := ossmon.StartParentSpan(context.Background(), mon, monitor.SrvPrfx+"testing")
	defer span.Finish()
	existingIncident, doesIncidentAlreadyExist := getExistingIncident(ctx, dbConn, recordID)
	assert.True(t, cmp.Equal(doesIncidentAlreadyExist, false))
	assert.Nil(t, existingIncident, tlog.Log()+"No incident created")
	existingIncident, doesIncidentAlreadyExist = getExistingIncident(ctx, dbConn, "")
	assert.True(t, cmp.Equal(doesIncidentAlreadyExist, false))
	assert.Nil(t, existingIncident, tlog.Log()+"No incident created")

}

func Test_messageMapToIncident(t *testing.T) {
	log.Println(tlog.Log())
	setEnv()
	mon := new(exmon.Monitor)
	mon.NRConfig = nrConfig
	mon.NRApp = nrApp
	gContext.NRMon = mon
	gContext.LogID = "nq2ds.incident"
	ctxAdapter, _ := getAdapterContext()

	cache, err := osscatalog.NewCache(ctxAdapter, tagNameListingServer)
	if err != nil || cache == nil {
		t.Fatal("Unable to get a cache", err)
	}

	desc = "================ create new incident ==========="
	var testData = fmt.Sprintf(incidentMsg, updateTime, status, priority, crn, desc, testutils.SourceID, "Private")
	messageMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(testData), &messageMap)
	if err != nil {
		log.Print(tlog.Log()+"unmarshaling error", err)
	}
	incidentFromQueue, isBadMessage, isRetryNeeded := messageMapToIncident(messageMap)
	log.Print(tlog.Log(), incidentFromQueue)
	log.Print(tlog.Log(), *incidentFromQueue.Audience)
	assert.False(t, isBadMessage, tlog.Log()+"isBadMessage should be false")
	assert.False(t, isRetryNeeded, tlog.Log()+"isRetryNeeded should be false")
	assert.NotNil(t, incidentFromQueue, tlog.Log()+"Should have returned a valid incidentFromQueue object")

}

func setEnv() {
	log.Println(tlog.Log())
	err := Setenv(cloudant.AccountID, "accountID")
	if err != nil {
		log.Fatalln(tlog.Log(), "Failing to set an environment variable", err)
	}
	err = Setenv(cloudant.AccountPW, "accountPW")
	if err != nil {
		log.Fatalln(tlog.Log(), "Failing to set an environment variable", err)
	}
	err = Setenv(cloudant.NotificationsURLEnv, "NotificationsURL")
	if err != nil {
		log.Fatalln(tlog.Log(), "Failing to set an environment variable", err)
	}
	err = Setenv(cloudant.ServicesURLEnv, "ServicesURL")
	if err != nil {
		log.Fatalln(tlog.Log(), "Failing to set an environment variable", err)
	}
	err = Setenv(cloudant.RuntimesURLEnv, "RuntimesURL")
	if err != nil {
		log.Fatalln(tlog.Log(), "Failing to set an environment variable", err)
	}
	err = Setenv(cloudant.PlatformURLEnv, "PlatformURL")
	if err != nil {
		log.Fatalln(tlog.Log(), "Failing to set an environment variable", err)
	}
	err = Setenv(globalcatalog.GCResourceURL, "GCURL")
	if err != nil {
		log.Fatalln(tlog.Log(), "Failing to set an environment variable", err)
	}
	err = Setenv(osscatalog.OSSCatalogCatYPKeyLabel, "catkeylabel")
	if err != nil {
		log.Fatalln(tlog.Log(), "Failing to set an environment variable", err)
	}
	err = Setenv(osscatalog.OSSCatalogCatYPKeyValue, "catkeyvalue")
	if err != nil {
		log.Fatalln(tlog.Log(), "Failing to set an environment variable", err)
	}
	nrConfig = newrelic.NewConfig("", "")
	nrConfig.Enabled = false
	nrApp, _ = newrelic.NewApplication(nrConfig)
}

func Test_getAdapterContext(t *testing.T) {
	log.Println(tlog.Log())
	setEnv()
	gContext.LogID = ""
	log.Print(gContext.LogID)
	ctxAdapter, _ := getAdapterContext()
	t.Log(ctxAdapter.LogID)

}

func Test_snStateToText(t *testing.T) {
	log.Println(tlog.Log())
	state := "1"
	result := snStateToText(state)
	assert.True(t, cmp.Equal(result, "new"))

	state = "Resolved"
	result = snStateToText(state)
	assert.True(t, cmp.Equal(result, "resolved"))

	state = ""
	result = snStateToText(state)
	assert.True(t, cmp.Equal(result, ""))
}

func Test_snUStatusToText(t *testing.T) {
	log.Println(tlog.Log())
	snUStatus := "20"
	result := snUStatusToText(snUStatus)
	assert.True(t, cmp.Equal(result, "potential-cie"))

	snUStatus = "Confirmed CIE"
	result = snUStatusToText(snUStatus)
	assert.True(t, cmp.Equal(result, "confirmed-cie"))

	snUStatus = ""
	result = snUStatusToText(snUStatus)
	assert.True(t, cmp.Equal(result, ""))
}

func Test_snPriorityToSeverity(t *testing.T) {
	log.Println(tlog.Log())
	snPriority := "1"
	result := snPriorityToSeverity(snPriority)
	assert.True(t, cmp.Equal(result, "1"))

	snPriority = "Sev - 2"
	result = snPriorityToSeverity(snPriority)
	assert.True(t, cmp.Equal(result, "2"))

	snPriority = "Sev - 3"
	result = snPriorityToSeverity(snPriority)
	assert.True(t, cmp.Equal(result, "3"))

	snPriority = "4"
	result = snPriorityToSeverity(snPriority)
	assert.True(t, cmp.Equal(result, "4"))
}

func Test_interfaceToStringArray(t *testing.T) {
	log.Println(tlog.Log())
	crnFull := `["crn:v1:bluemix:public:spark:eu-gb::::","crn:v1:staging:public::::::"]`
	strArr := interfaceToStringArray(crnFull)
	t.Log("Test_interfaceToStringArray", strArr)
}

func Test_buildIncidentToUpdate(t *testing.T) {
	log.Println(tlog.Log())

	currentIncident := &datastore.IncidentReturn{
		RecordID:           "e082aa9ddb316784c06b58b8dc9619d0",
		PnpCreationTime:    tnow,
		PnpUpdateTime:      tnow,
		SourceCreationTime: tnow,
		SourceUpdateTime:   tnow,
		OutageStartTime:    tnow,
		OutageEndTime:      tnow,
		ShortDescription:   desc,
		LongDescription:    desc,
		State:              "new",
		Classification:     "confirmed-cie",
		Severity:           "1",
		CRNFull:            []string{"crn:v1:internal:public:tip-oss-flow:us-east::::", "crn:v1:internal:public:tip-oss-flow:eu-gb::::"},
		SourceID:           testutils.SourceID,
		Source:             testutils.Source,
		RegulatoryDomain:   "reg domain data 123",
		AffectedActivity:   "Service / Network Access",
		Audience:           "Private",
	}

	updateIncident = buildIncidentToUpdate(currentIncident, updateIncident)
	log.Print("Test_buildIncidentToUpdate :", updateIncident)
	log.Print(*updateIncident.CRNFull)
	log.Println(*updateIncident.Audience)
	assert.Equal(t, audience, *updateIncident.Audience)

	updateIncident2 := &incidentFromQueue{}
	updateIncident2 = buildIncidentToUpdate(currentIncident, updateIncident2)
	log.Print("Test_buildIncidentToUpdate :", updateIncident2)
	log.Print(*updateIncident2.CRNFull)
	log.Println(*updateIncident2.Audience)
	assert.Equal(t, currentIncident.Audience, *updateIncident2.Audience)

	updateIncident3 := &incidentFromQueue{}
	strNone := ""
	updateIncident3.Audience = &strNone
	updateIncident3 = buildIncidentToUpdate(currentIncident, updateIncident3)
	log.Print("Test_buildIncidentToUpdate :", updateIncident3)
	log.Print(*updateIncident3.CRNFull)
	log.Println(*updateIncident3.Audience)
	assert.Equal(t, db.SNnill2PnP, *updateIncident3.Audience)
}

func Test_isSev1CIE(t *testing.T) {
	log.Println(tlog.Log())
	isSev1 := isSev1CIE(updateIncident)
	assert.True(t, cmp.Equal(isSev1, true))
}

func Test_isHighSevCIE(t *testing.T) {
	log.Println(tlog.Log())
	// Test sev 1
	isHighSev := isHighSevCIE(updateIncident)
	assert.True(t, cmp.Equal(isHighSev, true))
	// Test sev 2
	*updateIncident.Severity = "2"
	isHighSev = isHighSevCIE(updateIncident)
	assert.True(t, cmp.Equal(isHighSev, true))
}

func Test_hasAtleastOneCRN(t *testing.T) {
	log.Println(tlog.Log())
	hasAtleastOneCRN := hasAtleastOneCRN(updateIncident)
	assert.True(t, cmp.Equal(hasAtleastOneCRN, true))
}

func Test_isPnPRemoved(t *testing.T) {
	log.Println(tlog.Log())
	isPnPRemoved := isPnPRemoved(updateIncident)
	assert.True(t, cmp.Equal(isPnPRemoved, false))
}

func Test_removeInvalidCRNs(t *testing.T) {
	log.Println(tlog.Log())
	incidentFromQueue := removeInvalidCRNs(updateIncident)
	assert.True(t, cmp.Equal(incidentFromQueue.CRNFull, updateIncident.CRNFull))
	log.Print(*updateIncident.CRNFull)
}

func Test_toIncidentInsert(t *testing.T) {
	log.Println(tlog.Log())
	incidentInset := toIncidentInsert(updateIncident)
	log.Print("Test_toIncidentInsert:", incidentInset)
}

func Test_ProcessIncident_Not_HighSev(t *testing.T) {
	log.Println(tlog.Log())
	gContext.LogID = "nq2ds.incident"
	var testData = fmt.Sprintf(incidentMsg, updateTime, status, "3", crn, desc, testutils.SourceID)
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	log.Println("DEBuG:", testData)
	incident, _, isBadMessage := ProcessIncident(dbConn, msg, &mon)
	log.Println("DEBuG:", incident)
	log.Println("DEBuG:", isBadMessage)
	assert.True(t, cmp.Equal(isBadMessage, false))
	assert.Nil(t, incident, tlog.Log()+"Low severity, no incident created.")
}

// Bypass local storage: https://github.ibm.com/cloud-sre/toolsplatform/issues/9491
func Test_ProcessIncident(t *testing.T) {
	// override to ensure test is run agains the DB
	shared.BypassLocalStorage = false
	if !shared.BypassLocalStorage {
		log.Println(tlog.Log())
		var testData = fmt.Sprintf(incidentMsg, updateTime, status, priority, crn, desc, testutils.SourceID, audience)
		dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
		defer db.Disconnect(dbConn)
		t.Logf(tlog.Log())
		incident, _, isBadMessage := ProcessIncident(dbConn, msg, &mon)
		assert.True(t, cmp.Equal(isBadMessage, false))
		log.Print(tlog.Log(), incident)
		assert.NotNil(t, incident, tlog.Log()+"incident updated.")
	}
}

// Bypass local storage: https://github.ibm.com/cloud-sre/toolsplatform/issues/9491
func Test_ProcessExistingIncident(t *testing.T) {
	// override to ensure test is run agains the DB
	shared.BypassLocalStorage = false
	if !shared.BypassLocalStorage {
		log.Println(tlog.Log())
		var updateTimeNow = time.Now().UTC().Format("2006-01-02 15:04:05")
		var updatedDesc = "updated description"
		var testData = fmt.Sprintf(incidentMsg, updateTimeNow, status, priority, crn, updatedDesc, testutils.SourceID, "Private")
		dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
		defer db.Disconnect(dbConn)
		t.Logf(tlog.Log())
		incident, _, isBadMessage := ProcessIncident(dbConn, msg, &mon)
		assert.True(t, cmp.Equal(isBadMessage, false))
		log.Print(incident)
		assert.NotNil(t, incident, tlog.Log()+"incident updated.")
	}
}

// Bypass local storage: https://github.ibm.com/cloud-sre/toolsplatform/issues/9491
func Test_ProcessExistingIncident_noSev1(t *testing.T) {
	// override to ensure test is run agains the DB
	shared.BypassLocalStorage = false
	if !shared.BypassLocalStorage {
		log.Println(tlog.Log())
		var updateTimeNow = time.Now().AddDate(0, 0, 1).UTC().Format("2006-01-02 15:04:05")
		var updatedDesc = "updated description"
		var testData = fmt.Sprintf(incidentMsg, updateTimeNow, status, "2", crn, updatedDesc, testutils.SourceID, "")
		dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
		defer db.Disconnect(dbConn)
		t.Logf(tlog.Log())
		existingIncident, err, httpStatusCode := db.GetIncidentBySourceID(dbConn, testutils.Source, testutils.SourceID)
		assert.True(t, existingIncident != nil && err == nil && httpStatusCode == 200, true)
		incident, _, isBadMessage := ProcessIncident(dbConn, msg, &mon)
		assert.True(t, cmp.Equal(isBadMessage, false))
		assert.NotNil(t, incident, tlog.Log()+"Sev 2 tomb-stoned with PnPRemoved.")
	}
}

func Test_normalizeServiceName(t *testing.T) {
	log.Println(tlog.Log())
	setEnv()
	ctxAdapter, _ := getAdapterContext()
	cache, err := osscatalog.NewCache(ctxAdapter, tagNameListingServer)
	if err != nil || cache == nil {
		t.Fatal("Unable to get a cache", err)
	}
	normalizeSN, err := normalizeServiceName("MyService3", ctxAdapter)
	if err != nil {
		log.Print(nil, err)
	}
	assert.NotNil(t, normalizeSN, "should not be nil")
	normalizeSN, err = normalizeServiceName("badservice", ctxAdapter)
	if err != nil {
		log.Print(tlog.Log(), err)
	}
	assert.Equal(t, normalizeSN, "")
}

func Test_normalizeCRNServiceNames(t *testing.T) {
	log.Println(tlog.Log())
	setEnv()
	// leaving Newrelic config here as this test relies on a call to oss catalog.
	cfg := newrelic.NewConfig("pnp-nq2ds-testing", "0000000000000000000000000000000000000000")
	nrApp, _ := newrelic.NewApplication(cfg)
	mon := new(exmon.Monitor)
	mon.NRConfig = nrConfig
	mon.NRApp = nrApp
	gContext.NRMon = mon
	gContext.LogID = "nq2ds.incident"
	cache, err := osscatalog.NewCache(ctxtContext, tagNameListingServer)
	if err != nil || cache == nil {
		t.Fatal("Unable to get a cache", err)
	}
	crnFull := []string{"crn:v1:bluemix:public:MyService3:eu-gb::::", "crn:v1:bluemix:public:MyService3:eu-gb::::", "crn:v1:bluemix:public:MyService3:eu-gb::::"}
	updatedCRN, err := normalizeCRNServiceNames(crnFull)
	if err != nil {
		log.Print("normalizeCRNServiceNames failed", err)
	}
	res := []string{"crn:v1:bluemix:public:MyServiceParent:eu-gb::::"}
	assert.Equal(t, res, updatedCRN)
	badCrn := []string{"crn:v1:bluemix:public:badservice:eu-gb::::"}
	updatedCRN, err = normalizeCRNServiceNames(badCrn)
	if err != nil {
		log.Print("normalizeCRNServiceNames failed", err)
	}
	assert.Equal(t, updatedCRN, badCrn)
}

func tagNameListingServer(r *regexp.Regexp, cio catalog.IncludeOptions, myFunc func(r ossrecord.OSSEntry)) error {
	log.Println(tlog.Log())
	o := new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyServiceParent" //CRNServiceName
	o.StatusPage.CategoryID = "MyServiceParentCategoryID"
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.PnPEnabled)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "myservice1" //CRNServiceName
	o.StatusPage.CategoryID = "cloudoe.sop.enum.paratureCategory.literal.l247"
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.PnPEnabledIaaS)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyService2" //CRNServiceName
	o.StatusPage.CategoryID = "MyService2_CategoryID"
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.OneCloud)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyService3" //CRNServiceName
	o.StatusPage.CategoryID = "MyServiceParentCategoryID"
	o.StatusPage.CategoryParent = "MyServiceParent"
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.PnPEnabled)
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.OneCloud)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "testservice" //CRNServiceName
	o.StatusPage.CategoryID = "testservice"
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.PnPEnabled)
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.OneCloud)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "testService" //CRNServiceName
	o.StatusPage.CategoryID = "MyServiceParentCategoryID"
	o.StatusPage.CategoryParent = "MyServiceParent"
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.PnPEnabledIaaS)
	myFunc(o)
	return nil
}
