package testDefs

import (
	"os"
	"sync"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
)

type Watch struct {
	Path            string   `json:"path,omitempty"`
	RecordIDToWatch []string `json:"recordIDToWatch,omitempty"`
	CrnMasks        []string `json:"crnMasks,omitempty"`
	Wildcards       string   `json:"wildcards,omitempty"`
}

type NotificationWatch struct {
	Path      string   `json:"path,omitempty"`
	CrnMasks  []string `json:"crnMasks,omitempty"`
	Wildcards string   `json:"wildcards,omitempty"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Subscription struct {
	TargetAddress string   `json:"targetAddress"`
	TargetToken   string   `json:"targetToken"`
	Headers       []Header `json:"headers"`
	Name          string   `json:"name"`
	Expiration    string   `json:"expiration"`
}

type validation struct {
	Value string `json:"name"`
	Desc  string `json:"desc"`
}

var (
	Messages = make(chan string, 30)
	RMQUrl   = os.Getenv("NQ_URL")
	//For using Messages for RabbitMQ
	RMQEnableMessages = os.Getenv("RABBITMQ_ENABLE_MESSAGES")
	RMQAMQPSEndpoint  = os.Getenv("RABBITMQ_AMQPS_ENDPOINT")
	RMQTLSCert        = os.Getenv("RABBITMQ_TLS_CERT")
	APIKey            = os.Getenv("DRApiKey")
	RMQHooksCase      = os.Getenv("RMQHooksCase")
	RMQHooksIncident  = os.Getenv("RMQHooksIncident")
	RMQHooksChange    = os.Getenv("RMQHooksChange")
	IngressIP         = os.Getenv("ingressIP")
	envType           = os.Getenv("envType")
	SubscriptionUrl   = os.Getenv("subscriptionURL")

	RMQRoutingKey            = "nq2ds.case"
	RMQRoutingKeyMaintenance = "nq2ds.maintenance"
	RMQRoutingKeyStatus      = "status"
	DefaultCRN               = "crn:v1:bluemix:public:cloud-object-storage:us-east::::"
	CRNSample                = "crn:v1:bluemix:public:cloud-object-storage:us-east::::"
	CRNSample2               = "crn:v1:bluemix:public:cloud-object-storage:::::"
	exchangeName             = "pnp.direct"
	exchangeType             = "direct"
	webhookTarget            = "https://pnp-api-oss.dev.cloud.ibm.com/pnpintegrationtest/" // us-east nginx 10.190.24.45
	numWatches               = 0
	dumpWatchMapUrl          = os.Getenv("dumpWatchMapUrl") // "http://api-pnp-subscription-consumer/dump"

	createSubPostBody = `{
		"targetAddress": "` + webhookTarget + `",
		"targetToken": "tokenstring",
		"name": "mjl rest test"
		}`

	createCaseWatchPostBody     = `{"recordIDToWatch":["CS12345"]}`
	createIncidentWatchPostBody = `{
		"crnMasks": [
			"crn:v1:bluemix:public:miketest:us-east::::"
		],
		"wildcards": "false"
	  }`
	createNotificationWatchPostBody = `{
		"crnMasks": [
			"crn:v1:bluemix:public:miketest:::::"
		],
		"wildcards": "true"
	  }`
	createResourceWatchPostBody = `{
		"crnMasks": [
			"crn:v1:bluemix:public:miketest:::::"
		],
		"wildcards": "true"
	  }`
	createMaintenanceWatchPostBody = `{
		"crnMasks": [
			"crn:v1:bluemix:public:maintenancetest:::::", "crn:v1:bluemix:public:miketest:us-east::::"
		],
		"wildcards": "false"
	  }`

	preCreateMessageInRMQ      = `{"properties":{},"routing_key":"case","payload_encoding":"string", "payload":"`
	createCaseMessageInRMQ     = `{"operation":"update","sys_id":"f57123454","number":"CS12345","sys_created_on":"2018-09-08 18:02:32","comments":"2018-09-25 20:35:37 - System (Additional comments)\nCase was in a Resolved state for longer than 14 days, Case was automatically closed\n\n2018-09-11 19:38:39 - Mike Carrillo (Additional comments)\nClose notes: Hello,\r\n\r\nPlease reach out to our Sales team at sales@softlayer.com for further assistance.\r\n\r\nBest Regards,\r\n\r\nIBM Cloud\r\nRevenue Support Enablement Team (RSET)\n\n","sys_updated_on":"2018-09-25 20:35:37","u_status":"Closed","instance":"watsondev","new_comments":"2018-09-25 20:35:37 - System (Additional comments)\nCase was in a Resolved state for longer than 14 days, Case was automatically closed\n\n"}`
	incidentUpdateTime         = time.Now().Format("2006-01-02 15:04:05")
	createIncidentMessageInRMQ = `{
		"operation": "insert",
		"sys_id": "f6176a1ddb74e7808799327e9d961922",
		"number": "INC12345",
		"sys_created_on": "2018-09-26 10:46:33",
		"incident_state": "New",
		"u_disruption_ended": "",
		"u_disruption_began": "",
		"priority": "Sev - 1",
		"u_status": "Confirmed CIE",
		"short_description": "test 123",
		"description": "",
		"cmdb_ci": "servicenow",
		"comments": "",
		"sys_updated_on": "%s",
		"u_environment": "IBM Public US-EAST (WDC)",
		"u_description_customer_impact": "",
		"u_current_status": "",
		"crn": [
		  "crn:v1:bluemix:public:miketest:us-east::::"
		],
		"instance": "watsondev",
		"new_work_notes": "2018-09-26 10:46:33 - Michael Lee (Work notes)\nCIE Paging disabled in Development environment.\n\n"
		}`

	createDrMessageInRMQ   = `{"operation": "insert","info": {"notification_type": "Maintenance","notification_channels": "testPublic","source_creation_time": "%d-%02d-%02dT23:05:23Z","source_update_time": "%d-%02d-%02dT02:50:28Z","planned_start_time": "%d-%02d-%02dT01:00:01Z","planned_end_time": null,"short_description": "Fake DR6-%s created for testing purposes only","long_description": "Fake DR-%s long desc","crnFull": ["crn:v1:bluemix:public:cloud-object-storage:us-south::::","crn:v1:bluemix:public:db2-warehouse-on-cloud:us-south::::"],"state": "in-progress","disruptive": true,"source_id": "2222243-%s","source": "Doctor-RTC","regulatory_domain": null,"maintenance_duration": 180,"notification_status": "Publish","disruption_type": "Console Access","disruption_description": "Test2: Console not accessible during update","disruption_duration": 180}}`
	mutexValidation        = &sync.Mutex{}
	basicValidationStrings = []validation{
		validation{"/api/v1/pnp/status/maintenances/" + maintenanceRecordId, "maintenance record"},
		validation{"api/v1/pnp/status/resources/1147af4a3172a2844925f2b8124d7acfb1131bad6d419a0c58b2518be5886134", "resource record"},
		validation{"api/v1/pnp/status/incidents/3bdab210d926464269c7eb2df5ff41d96258cddef1964b12f3aca207f5c21db6", "incident id"},
	}

	notificationChangeValidationStrings = []string{
		"api/v1/pnp/status/maintenances/508c3fc25f0415bca12e73ffb7d82997850283d012ac753311541c5ff8a86476",
		"api/v1/pnp/status/notifications/2d501ff093edea60c9dd9d92ff20be9718a1754055ccf818005eddfb4e75e543"}

	//snChangeNumber = "CHG0" + strconv.Itoa(rand.Intn(100000)+800000)
	snChangeNumber = "CHG0898081"
	// createIncidentMessageInRMQ2 = `{"operation":"insert","sys_id":"f6176a1ddb74e7808799327e9d961922","number":"INC12345","sys_created_on":"2018-09-26 10:46:33","incident_state":"New","u_disruption_ended":"","u_disruption_began":"","priority":"Sev - 1","u_status":"Confirmed CIE","short_description":"test 123","description":"","cmdb_ci":"servicenow","comments":"","sys_updated_on":"2018-09-26 10:46:33","u_environment":"IBM Public US-EAST (WDC)","u_description_customer_impact":"","u_current_status":"","crn":["crn:v1:bluemix:public:miketest:us-east::::"],"instance":"watsondev","new_work_notes":"2018-09-26 10:46:33 - Michael Lee (Work notes)\nCIE Paging disabled in Development environment.\n\n"}`
	snChangeRecordId           = db.CreateRecordIDFromSourceSourceID("ServiceNow", snChangeNumber)
	createSNChangeMessageInRMQ = `{
		"operation": "insert",
		"sys_id": "44ef5342db19e3c48799327e9d961970",
		"number": "` + snChangeNumber + `",
		"sys_created_on": "2018-10-23 19:21:20",
		"start_date": "2018-10-26 19:21:11",
		"end_date": "2018-11-25 13:14:00",
		"short_description": "testing: short desc",
		"description": "test from rest test",
		"cmdb_ci": "cloudantnosqldb",
		"sys_updated_on": "%s",
		"u_environment": "American Airlines 1",
		"phase_state": "Open",
		"instance": "watsondev",
		"new_work_notes": "2018-10-23 19:21:20 - Rahamathulla Nalband (Work notes)\nAttachment \"Please request approval manually\" added by workflow \"Change Request - IBM\".\n\n",
		"crn": [
			"crn:v1:bluemix:public:miketest:us-east::::"
		]
	  }`

	maintenanceRecordId           = db.CreateRecordIDFromSourceSourceID("Doctor-RTC", "12345")
	CreateMaintenanceMessageinRMQ = `{
		"record_id": "` + maintenanceRecordId + `",
		"pnp_creationTime": "2018-09-25T18:33:52Z",
		"pnp_update_time": "2018-09-25T21:04:50Z",
		"source_creation_time": "2018-09-18T23:21:03Z",
		"source_update_time": "2018-09-25T21:00:23Z",
		"short_description": "rest test maintenance",
		"long_description": "rest test maintenance long desc",
		"state": "in-progress",
		"disruptive": true,
		"source_id": "12345",
		"source": "Doctor-RTC",
		"crnFull": [
		  "crn:v1:bluemix:public:cloud-object-storage:us-east::::"
		]
	  }
		`
	createBSPNViaHooks       = `{"type":"notification","regions":["aa1:prod:us-ne"],"start_date":%s,"end_date":0,"title":"INVESTIGATING: Testing BSPN - PnP123567","description":"<p>Testing BSPN - PnP - working</p>","severity":"Sev - 1","components":["cloudoe.sop.enum.paratureCategory.literal.l128"],"id":"BSPN%s","parent_id":"INC0311030","status":"CIE In Progress","affected_activities":"Application Availability","modified":%s,"crn":["crn:v1:bluemix:public:cloud-object-storage:us-east::::"]}`
	statusRecordId           = db.CreateRecordIDFromSourceSourceID("globalCatalog", "crn:v1:bluemix:public:sql-query:us-south::::")
	createStatusMessageInRMQ = `{
		"record_id": "` + statusRecordId + `",
		"pnp_creation_time": "2018-09-14T11:50:22Z",
		"pnp_update_time": "2018-09-14T11:50:22Z",
		"source_creation_time": "2018-07-10T10:26:33.701Z",
		"source_update_time": "%s",
		"crn_full": "crn:v1:bluemix:public:sql-query:us-south::::",
		"state": "ok",
		"operational_status": "GA",
		"source": "globalCatalog",
		"source_id": "crn:v1:bluemix:public:sql-query:us-south::::",
		"status": "ok",
		"category_id": "cloudant",
		"displayName": [
		  {
			"name": "Cloudant",
			"language": "en"
		  }
		],
		"visibility": [
		  "clientFacing"
		]
	  }`

	resource_routingKey            = "resource"
	sample_resourceAdapterToNq2ds  = `{"id":"accesstrail","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"accesstrail","creationTime":"2015-12-11T20:01:39Z","updateTime":"2018-08-22T20:46:03.8Z","deployments":[{"id":"free-au-syd","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:au-syd::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:au-syd::::","creationTime":"2018-01-18T05:53:47Z","updateTime":"2018-03-09T06:32:48.007Z","parent":true},{"id":"free-eu-de","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:eu-de::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:eu-de::::","creationTime":"2017-03-21T03:25:34Z","updateTime":"2018-03-09T06:32:48.261Z","parent":true},{"id":"lite-eu-gb","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:eu-gb::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:eu-gb::::","creationTime":"2018-02-01T06:55:49Z","updateTime":"2018-03-09T06:32:48.209Z","parent":true},{"id":"free-us-south","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l247","crn":"crn:v1:bluemix:public:accesstrail:us-south::::","displayName":[{"language":"en","text":"Activity Tracker"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:accesstrail:us-south::::","creationTime":"2015-12-11T20:01:39Z","updateTime":"2018-03-09T06:32:47.929Z","parent":true}]}`
	sample_resourceAdapterToNq2ds2 = `{"id":"cloudantnosqldb","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l119","crn":"crn:v1:bluemix:public:cloudantnosqldb:::::","displayName":[{"language":"en","text":"Cloudant"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"cloudantnosqldb","creationTime":"2014-06-03T07:04:12Z","updateTime":"2018-10-01T16:51:19.275Z","deployments":[{"id":"dedicated-hardware-au-syd-rc","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l119","crn":"crn:v1:bluemix:public:cloudantnosqldb:au-syd::::","displayName":[{"language":"en","text":"Cloudant"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:cloudantnosqldb:au-syd::::","creationTime":"2018-07-10T10:25:15.801Z","updateTime":"2018-07-16T15:41:43.138Z","parent":true},{"id":"dedicated-hardware-eu-de-rc","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l119","crn":"crn:v1:bluemix:public:cloudantnosqldb:eu-de::::","displayName":[{"language":"en","text":"Cloudant"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:cloudantnosqldb:eu-de::::","creationTime":"2018-07-10T10:25:38.772Z","updateTime":"2018-07-16T15:41:47.099Z","parent":true},{"id":"dedicated-hardware-eu-gb-rc","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l119","crn":"crn:v1:bluemix:public:cloudantnosqldb:eu-gb::::","displayName":[{"language":"en","text":"Cloudant"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:cloudantnosqldb:eu-gb::::","creationTime":"2018-07-10T10:26:00.752Z","updateTime":"2018-07-16T15:41:50.744Z","parent":true},{"id":"dedicated-hardware-jp-tok-rc","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l119","crn":"crn:v1:bluemix:public:cloudantnosqldb:jp-tok::::","displayName":[{"language":"en","text":"Cloudant"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:cloudantnosqldb:jp-tok::::","creationTime":"2018-09-20T11:42:58.054Z","updateTime":"2018-09-25T13:20:31.648Z","parent":true},{"id":"dedicated-hardware-us-east-rc","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l119","crn":"crn:v1:bluemix:public:cloudantnosqldb:us-east::::","displayName":[{"language":"en","text":"Cloudant"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:cloudantnosqldb:us-east::::","creationTime":"2018-07-10T10:26:17.56Z","updateTime":"2018-07-16T15:41:54.679Z","parent":true},{"id":"dedicated-hardware-us-south-rc","active":"true","disabled":"false","kind":"deployment","categoryId":"cloudoe.sop.enum.paratureCategory.literal.l119","crn":"crn:v1:bluemix:public:cloudantnosqldb:us-south::::","displayName":[{"language":"en","text":"Cloudant"}],"state":"ok","operationalStatus":"GA","visibility":["hasStatus","clientFacing"],"source":"globalCatalog","sourceId":"crn:v1:bluemix:public:cloudantnosqldb:us-south::::","creationTime":"2018-07-10T10:26:33.701Z","updateTime":"2018-07-16T15:42:54.174Z","parent":true}]}`
)
