package producer

import (
	"context"
	"encoding/json"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
	"log"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/ossmon"

	newrelic "github.com/newrelic/go-agent"
	"github.com/stretchr/testify/assert"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

var (
	testProducer *Producer
	nrApp        newrelic.Application
	current      = time.Now()
	tnow         = current.UTC().Format("2006-01-02T15:04:05Z")
)

func InitNRApp() ossmon.OSSMon {
	nrConfig := newrelic.NewConfig("", "")
	nrConfig.Enabled = false
	nrApp, _ = newrelic.NewApplication(nrConfig)
	svceName := monitor.SrvPrfx + "testing"
	mon := ossmon.OSSMon{
		NewRelicApp: nrApp,
		Sensor:      instana.NewSensor(svceName),
	}
	return mon
}

func TestNewProducer(t *testing.T) {
	var (
		urls                = []string{"amqp://guest:guest@localhost:5672"}
		exchangeName        = "pnp.direct"
		caseOutQKey         = "nq2ds.case"
		incidentOutQKey     = "nq2ds.incident"
		incidentBulkOutQKey = "nq2ds.incident"
		maintenanceOutQKey  = "nq2ds.maintenance"
		resourceOutQKey     = "nq2ds.resource"
	)
	mon:=InitNRApp()
	testProducer, _ = NewProducer(urls, exchangeName, "direct", caseOutQKey, &mon)
	testProducer.CaseRoutingKey = caseOutQKey
	testProducer.IncidentRoutingKey = incidentOutQKey
	testProducer.IncidentBulkRoutingKey = incidentBulkOutQKey
	testProducer.MaintenanceRoutingKey = maintenanceOutQKey
	testProducer.ResourceRoutingKey = resourceOutQKey
}

//func TestPostCase(t *testing.T) {
//	mon = InitNRApp()
//	const FCT = "Test process case: "
//	msgCase := `{"record_id":"be0c18b22450bf0c9edde821ec86f5163287586abaa0db6d52f5561307cfab4a","source":"servicenow","source_id":"123457","source_sys_id":"000000"}`
//	caseReturn := datastore.CaseReturn{}
//	err := json.Unmarshal([]byte(msgCase), &caseReturn)
//	assert.Nil(t, err, "should return nil.")
//	err = testProducer.PostCase(&caseReturn, &mon)
//	assert.Nil(t, err, "should return nil.")
//}

func TestPostIncident(t *testing.T) {
	mon := InitNRApp()
	incidentReturn := &datastore.IncidentReturn{
		RecordID:           "e082aa9ddb316784c06b58b8dc9619d0",
		PnpCreationTime:    tnow,
		PnpUpdateTime:      tnow,
		SourceCreationTime: tnow,
		SourceUpdateTime:   tnow,
		OutageStartTime:    tnow,
		OutageEndTime:      tnow,
		ShortDescription:   "This is a test data",
		LongDescription:    "some test data for long description of the incident",
		State:              "new",
		Classification:     "confirmed-cie",
		Severity:           "1",
		CRNFull:            []string{"crn:v1:internal:public:tip-oss-flow:eu-gb::::"},
		SourceID:           "demoIN00003",
		Source:             "SN",
		RegulatoryDomain:   "reg domain data 123",
		AffectedActivity:   "Service / Network Access",
	}
	err := testProducer.PostIncident(incidentReturn, false, &mon)
	assert.Nil(t, err, tlog.Log()+"should return nil.")
}

func TestPostMaintenance(t *testing.T) {
	var msg = `{
	"record_id": "efc11c93781555c7e5bb181629e6e97240799dac03cf151f00c60221cf1fb868",
	"pnp_creationTime": "2018-11-21T15:22:46+08:00",
	"pnp_update_time": "2018-12-01T10:19:05+08:00",
	"source_creation_time": "2018-11-12T23:10:43+08:00",
	"source_update_time": "2018-11-15T05:32:38+08:00",
	"planned_start_time": "2018-11-15T11:00:00+08:00",
	"planned_end_time": "2018-11-15T18:00:00+08:00",
	"short_description": "Increase capacity and security - Window 2 of 2",
	"long_description": "test",
	"state": "complete",
	"disruptive": true,
	"source_id": "895908",
	"source": "ServiceNow",
	"crnFull": [
		"crn:v1:bluemix:public:ace:us-south::::"
	],
	"maintenance_duration": 480,
	"disruption_type": "Other (specify in Description)",
	"disruption_description": "test",
	"disruption_duration": 480,
	"completion_code":"successful",
	"should_have_notification": false
}
`
	maintenanceMap := datastore.MaintenanceMap{}
	err := json.Unmarshal([]byte(msg), &maintenanceMap)
	if err != nil {
		log.Print(tlog.Log()+"Error occurred unmarshaling , err = ", err)
	}
	mon := InitNRApp()
	err = testProducer.PostMaintenance(&maintenanceMap, &mon)
	assert.Nil(t, err, tlog.Log()+"should return nil.")
}

func TestPostResources(t *testing.T) {
	var msg = `{
    "category_id":"test",
    "record_id":"efc11c93781555c7e5bb181629e6e97240799dac03cf151f00c60221cf1fb868",
    "source":"ServiceNow",
    "source_id":"123456",
    "status":"complete",
    "crn_full":"crn:v1:bluemix:public:ace:us-south::::",
    "pnp_creation_time":"",
    "pnp_update_time":"",
    "source_creation_time":"",
    "source_update_time":"",
    "state":"",
    "operational_status":"",
    "status_update_time":"",
    "regulatory_domain":"",
    "category_id":"",
    "category_parent":false,
    "record_hash":"",
    "displayName":[{"name":"test status","language":"en"}],
    "visibility": ["test"],
    "tags": [{"id":"tag1"}]
    }`

	resReturn := new(datastore.ResourceReturn)
	resReturns := []*datastore.ResourceReturn{}
	err := json.Unmarshal([]byte(msg), &resReturn)
	if err != nil {
		log.Print(tlog.Log()+"Error occurred unmarshaling , err = ", err)
	}
	mon := InitNRApp()
	resReturns = append(resReturns, resReturn)
	err = testProducer.PostResources(resReturns, &mon)
	assert.Nil(t, err, tlog.Log()+"should return nil.")
}

func Test_postMessageWithRetry(t *testing.T) {
	routingKey := "testKey"
	messageToPost := "postMsg"
	mon := InitNRApp()
	span, ctx := ossmon.StartParentSpan(context.Background(), mon, monitor.SrvPrfx+"testing")
	defer span.Finish()
	testProducer.postMessageWithRetry(ctx, routingKey, messageToPost)
}
