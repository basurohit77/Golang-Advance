package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/ossmon"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
	"log"
	"testing"

	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-nq2ds/testutils"
)

var (
	resource = `{
    "id": "samplewdeployments",
  	"href": "",
  	"recordID": "",
  	"kind": "resource",
  	"mode": "",
  	"categoryId": "categoryid",
  	"crn": "crn:v1:staging:public:samplewdeployments:::::",
    "tags":["tag1","tag2"],
    "displayName":[{"text":"displayName","language":"en"}],
    "entry_type":"",
    "servicenow_sys_id":"",
    "service_now_ciurl":"",
    "state":"ok",
    "status":"ok",
    "operationalStatus":"GA",
    "visibility":["clientFacing"],
    "source":"%s",
		"sourceId":"%s",
    "creationTime": "2017-09-21T15:02:48.691Z",
  	"updateTime": "%s",
    "deployments":[%s],
    "parent":true
    }`

	deployment = `{
    "id": "sample-ys1-eu-gb",
    "active": "true",
    "disabled": "false",
    "href": "",
    "recordID": "",
    "kind": "deployment",
    "categoryId": "categoryid",
    "crn": "crn:v1:staging:public:samplewdeployments:eu-gb::::",
      "tags":["tag1"],
      "displayName":[{"text":"Sample Deployment in EU","language": "en"}],
      "entry_type": "",
			"state": "ok",
			"status": "",
			"operationalStatus": "GA",
      "visibility":["clientFacing"],
      "source":"%s",
      "sourceId":"%s",
      "creationTime": "2017-09-21T15:04:12.917Z",
      "updateTime":"%s"
      }`


 sampleWithDeploymentJSON = `{
	"id": "samplewdeployments",
	"href": "",
	"recordID": "",
	"kind": "resource",
	"mode": "",
	"categoryId": "categoryidSampleWDeployments",
	"crn": "crn:v1:staging:public:samplewdeployments:::::",
	"tags": [
		"apidocs_enabled",
		"dev_ops",
		"ibm_created",
		"lite",
		"security"
	],
	"displayName": [
		{
			"language": "en",
			"text": "Sample With Deployments"
		}
	],
	"entry_type": "",
	"servicenow_sys_id": "",
	"service_now_ciurl": "",
	"state": "ok",
	"status": "",
	"operationalStatus": "GA",
	"visibility": [
		"clientFacing"
	],
	"source": "globalCatalog",
	"sourceId": "samplewdeployments",
	"creationTime": "2017-09-21T15:02:48.691Z",
	"updateTime": "2018-08-02T10:18:45.167Z",
	"deployments": [
		{
			"id": "sample-ys1-eu-gb",
			"active": "true",
			"disabled": "false",
			"href": "",
			"recordID": "",
			"kind": "deployment",
			"categoryId": "categoryidSampleWDeployments",
			"crn": "crn:v1:staging:public:samplewdeployments:eu-gb::::",
			"tags": null,
			"displayName": [
				{
					"language": "en",
					"text": "Sample Deployment in EU"
				}
			],
			"entry_type": "",
			"state": "ok",
			"status": "",
			"operationalStatus": "GA",
			"visibility": [
				"clientFacing"
			],
			"source": "globalCatalog",
			"sourceId": "crn:v1:staging:public:samplewdeployments:eu-gb::::",
			"creationTime": "2017-09-21T15:04:12.917Z",
			"updateTime": "2018-01-18T19:28:40.318Z"
		},
		{
			"id": "sample-ys1-us-south",
			"active": "true",
			"disabled": "false",
			"href": "",
			"recordID": "",
			"kind": "deployment",
			"categoryId": "categoryidSampleWDeployments",
			"crn": "crn:v1:staging:public:samplewdeployments:us-south::::",
			"tags": null,
			"displayName": [
				{
					"language": "en",
					"text": "Sample Deploy in US South"
				}
			],
			"entry_type": "",
			"state": "ok",
			"status": "",
			"operationalStatus": "GA",
			"visibility": [
				"clientFacing"
			],
			"source": "globalCatalog",
			"sourceId": "crn:v1:staging:public:samplewdeployments:us-south::::",
			"creationTime": "2017-09-21T15:04:33.97Z",
			"updateTime": "2018-01-18T19:28:40.4Z"
		}
	],
	"parent": true
}`

 sampleWithoutDeploymentJSON = `{
	"id": "sample without deployments",
	"href": "",
	"recordID": "",
	"kind": "resource",
	"mode": "",
	"categoryId": "CategoryidNoDeployments",
	"crn": "crn:v1:staging:public:samplenodeployments:::::",
	"tags": [
		"apidocs_enabled",
		"dev_ops",
		"ibm_created",
		"lite",
		"security"
	],
	"displayName": [
		{
			"language": "en",
			"text": "Sample No Deploys"
		}
	],
	"entry_type": "",
	"servicenow_sys_id": "",
	"service_now_ciurl": "",
	"state": "ok",
	"status": "",
	"operationalStatus": "GA",
	"visibility": [
		"clientFacing"
	],
	"source": "globalCatalog",
	"sourceId": "samplenodeployments",
	"creationTime": "2017-09-21T15:02:48.691Z",
	"updateTime": "2018-08-02T10:18:45.167Z",
	"deployments": [
	],
	"parent": true
}`

 sampleBadDataJSON = `{
	"id": "sampleBadData_JSON",
	"href": "",
	"recordID": "",
	"kind": "hope",
	"mode": "",
	"categoryId": "xxx",
	"crn": "crn:v1:staging:public:sampleBadData_JSON:::::",
	"tags": [
		"apidocs_enabled",
		"dev_ops",
		"ibm_created",
		"lite",
		"security"
	],
	"displayName": [
		{
			"language": "en",
			"text": "sampleBadData_JSON"
		}
	],
	"entry_type": "",
	"servicenow_sys_id": "",
	"service_now_ciurl": "",
	"state": "ok",
	"status": "",
	"operationalStatus": "GA",
	"visibility": [
		"clientFacing"
	],
	"source": "globalCatalog",
	"sourceId": "sampleBadData_JSON",
	"creationTime": "2xxxx",

	"deployments": [
	],
	"parent": true
}`

 returnedResourceWithDeploymentsJSON = `[
	{
		"source_creation_time": "2017-09-21T15:04:12.917Z",
		"source_update_time": "2018-01-18T19:28:40.318Z",
		"crn_full": "crn:v1:staging:public:samplewdeployments:eu-gb::::",
		"state": "ok",
		"operational_status": "GA",
		"source": "globalCatalog",
		"source_id": "crn:v1:staging:public:samplewdeployments:eu-gb::::",
		"category_id": "categoryidSampleWDeployments",
		"displayName": [
		  {
			"name": "Sample Deployment in EU",
			"language": "en"
		  }
		],
		"visibility": [
		  "clientFacing"
		],
		"record_hash": "bc58b19a20b55621c91499046373d06f4054e14002f7b8735118acb0f0cd4e1e"
	  },
	  {
		"source_creation_time": "2017-09-21T15:04:33.97Z",
		"source_update_time": "2018-01-18T19:28:40.4Z",
		"crn_full": "crn:v1:staging:public:samplewdeployments:us-south::::",
		"state": "ok",
		"operational_status": "GA",
		"source": "globalCatalog",
		"source_id": "crn:v1:staging:public:samplewdeployments:us-south::::",
		"category_id": "categoryidSampleWDeployments",
		"displayName": [
		  {
			"name": "Sample Deploy in US South",
			"language": "en"
		  }
		],
		"visibility": [
		  "clientFacing"
		],
		"record_hash": "737ef24159401e6a7275648290f42279b537ba49f54fd3714a37b924221c6af2"
	  }
]`

 returnedResourceWithoutDeploymentsJSON = `[]`
)

func prepareResourceData(database *sql.DB) {
	tnow := time.Now().AddDate(0, 0, -1).Format("2006-01-02T15:04:05Z")
	visibility := []string{
		"hasStatus",
	}
	dn := datastore.DisplayName{Name: "name", Language: "en"}
	tag := datastore.Tag{ID: "tag1"}
	resource1 := datastore.ResourceInsert{
		SourceID:           testutils.SourceID,
		SourceCreationTime: "2018-06-07T22:01:01Z",
		SourceUpdateTime:   "2018-09-07T22:01:01Z",
		CRNFull:            "crn:v1:staging:public:samplewdeployments:eu-gb::::",
		State:              "ok",
		DisplayNames:       []datastore.DisplayName{dn},
		OperationalStatus:  "none",
		Tags:               []datastore.Tag{tag},
		Source:             testutils.Source,
		Status:             "",
		Visibility:         visibility,
		RecordHash:         "some hash",
	}

	recordID, err, _ := db.InsertResource(database, &resource1)
	if err != nil {
		log.Println("Insert Resource failed: ", err.Error())
	} else {
		log.Println("record_id: ", recordID)
		log.Println("Insert Resource passed")
		rr, _, _ := db.GetResourceByRecordID(database, recordID)
		log.Println(rr.PnpUpdateTime)
		log.Println(rr.Status)
	}

	maint2 := datastore.MaintenanceInsert{
		SourceCreationTime:    tnow,
		SourceUpdateTime:      tnow,
		PlannedStartTime:      tnow,
		PlannedEndTime:        tnow,
		ShortDescription:      "testing maintenance dummy data",
		LongDescription:       "this is just a long description of a maintenance record",
		CRNFull:               []string{"crn:v1:staging:public:samplewdeployments:eu-gb::::"},
		State:                 "scheduled",
		Disruptive:            true,
		SourceID:              "dummyMaintenance0002",
		Source:                "SN",
		RegulatoryDomain:      "some reg dom data",
		DisruptionType:        "Other (specify in Description)",
		DisruptionDuration:    30,
		MaintenanceDuration:   30,
		DisruptionDescription: "test disruption description",
	}
	maintenance002, errM, _ := db.InsertMaintenance(database, &maint2)
	if errM != nil {
		log.Print(errM)
	}
	log.Print("maintenance002=", maintenance002)

	in3 := datastore.IncidentInsert{
		SourceCreationTime: tnow,
		SourceUpdateTime:   tnow,
		OutageStartTime:    tnow,
		OutageEndTime:      tnow,
		ShortDescription:   "This is a test data",
		LongDescription:    "some test data for long description of the incident",
		State:              "new",
		Classification:     "confirmed-cie",
		Severity:           "1",
		CRNFull:            []string{"crn:v1:staging:public:samplewdeployments:eu-gb::::"},
		SourceID:           "demoIN00003",
		Source:             "SN",
		RegulatoryDomain:   "reg domain data 123",
		AffectedActivity:   "Service / Network Access",
	}
	incident003, errN, _ := db.InsertIncident(database, &in3)
	if errN != nil {
		log.Print(errN)
	}
	log.Print("incident003=", incident003)
}

func Test_processResource_insert(t *testing.T) {
	log.Println(tlog.Log())
	updateTime := "2018-10-15T19:28:40.4Z"
	deploymentData := fmt.Sprintf(deployment, testutils.Source, testutils.SourceID, updateTime)
	testData := fmt.Sprintf(resource, testutils.Source, testutils.SourceID, updateTime, deploymentData)
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	prepareResourceData(dbConn)
	t.Logf(tlog.Log())
	updateResources, err := ProcessResourceMsg(dbConn, msg, &mon)
	assert.Nil(t, err, tlog.Log()+"Should have no error.")
	assert.NotEqual(t, len(updateResources), 0)
}

func Test_resourceWithDecryptionError(t *testing.T) {
	log.Println(tlog.Log())
	msg := `{"test":"test"}`
	dbConn, _, mon := testutils.PrepareTestInc(t, msg)
	//prepareResourceData(dbConn)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	_, err := ProcessResourceMsg(dbConn, []byte(msg), &mon)
	assert.NotNil(t, err, tlog.Log()+"Should have decryption error.")
}

func Test_resourceWithUnmarshalError(t *testing.T) {
	log.Println(tlog.Log())
	testData := `{
	"id": "samplewdeployments",
	"href": "",
	"recordID": "",
	"kind": "resource",
	"mode": "",
	"categoryId": "categoryidSampleWDeployments",
	"crn": "crn:v1:staging:public:samplewdeployments:::::",`
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	_, err := ProcessResourceMsg(dbConn, msg, &mon)
	assert.NotNil(t, err, tlog.Log()+"Should have unmarshal error.")
}

func Test_processResource_update(t *testing.T) {
	log.Println(tlog.Log())
	updateTime := "2018-10-20T15:04:05Z"
	deploymentData := fmt.Sprintf(deployment, testutils.Source, testutils.SourceID, updateTime)
	testData := fmt.Sprintf(resource, testutils.Source, testutils.SourceID, updateTime, deploymentData)
	dbConn, msg, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	t.Logf(tlog.Log())
	updateResources, err := ProcessResourceMsg(dbConn, msg, &mon)
	assert.Nil(t, err, tlog.Log()+"Should have no error.")
	assert.NotEqual(t, len(updateResources), 0)
}

func Test_insertUpdateResourceInDB(t *testing.T) {
	log.Println(tlog.Log())
	testData := ""
	dbConn, _, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	span, ctx := ossmon.StartParentSpan(context.Background(), mon, monitor.SrvPrfx+"testing")
	defer span.Finish()
	resources := []datastore.ResourceInsert{}
	resources = append(resources, datastore.ResourceInsert{})
	resources[0].Visibility = append(resources[0].Visibility, "hasStatus")
	resources[0].CRNFull = "crn:v1:bluemix:public:accesstrail:au-syd::::"
	resources[0].Source = "servicenow"
	resources[0].SourceID = "Test123"
	updateResources, err := insertUpdateResourceInDB(ctx, &mon,dbConn, resources)
	assert.Nil(t, err, tlog.Log()+"Should have no error.")
	assert.Equal(t, len(updateResources), 1)
}

func Test_insertUpdateResourceInDB_EmptyResources(t *testing.T) {
	log.Println(tlog.Log())
	testData := ""
	dbConn, _, mon := testutils.PrepareTestInc(t, testData)
	defer db.Disconnect(dbConn)
	resources := []datastore.ResourceInsert{}
	span, ctx := ossmon.StartParentSpan(context.Background(), mon, monitor.SrvPrfx+"testing")
	defer span.Finish()
	updateResources, err := insertUpdateResourceInDB(ctx, &mon,dbConn, resources)
	assert.Nil(t, err, tlog.Log()+"Should have no error.")
	assert.Equal(t, len(updateResources), 0)
}

func TestConvertToResourceInsertArray(t *testing.T) {
	log.Println(tlog.Log())
	var returnedResourceWithoutDeployments []datastore.ResourceInsert
	var returnedResourceWithDeployments []datastore.ResourceInsert
	var sampleWithoutDeployment = new(datastore.PnpStatusResource)
	var sampleWithDeployments = new(datastore.PnpStatusResource)

	err := json.Unmarshal([]byte(sampleWithDeploymentJSON), &sampleWithDeployments)
	if err != nil {
		log.Println(tlog.Log()+"Error in unmarshal :", err.Error())
		assert.Fail(t, "Unmarshal error")
	}
	err = json.Unmarshal([]byte(sampleWithoutDeploymentJSON), &sampleWithoutDeployment)
	if err != nil {
		log.Println(tlog.Log()+"Error in unmarshal :", err.Error())
		assert.Fail(t, "Unmarshal error")
	}
	err = json.Unmarshal([]byte(returnedResourceWithoutDeploymentsJSON), &returnedResourceWithoutDeployments)
	if err != nil {
		log.Println(tlog.Log()+"Error in unmarshal :", err.Error())
		assert.Fail(t, "Unmarshal error")
	}
	err = json.Unmarshal([]byte(returnedResourceWithDeploymentsJSON), &returnedResourceWithDeployments)
	if err != nil {
		log.Println(tlog.Log()+"Error in unmarshal :", err.Error())
		assert.Fail(t, "Unmarshal error")
	}
	returned := ConvertToResourceInsertArray(sampleWithDeployments)
	log.Printf("Returned:\n%+v", returned)
	log.Printf("returnedResourceWithDeployments: \n%+v", returnedResourceWithDeployments)
	log.Println("Assert returned:", assert.True(t, cmp.Equal(returned, returnedResourceWithDeployments)))
	returned2 := ConvertToResourceInsertArray(sampleWithoutDeployment)
	log.Printf("Returned2:\n%+v", returned2)
	log.Printf("returnedResourceWithoutDeployments: \n%+v", returnedResourceWithoutDeployments)
	log.Println("Assert returned2:", assert.True(t, cmp.Equal(len(returned2), len(returnedResourceWithoutDeployments))))

}
func TestBadData(t *testing.T) {
	log.Println(tlog.Log())
	var sampleBadData = new(datastore.PnpStatusResource)
	err := json.Unmarshal([]byte(sampleBadDataJSON), &sampleBadData)
	if err != nil {
		log.Println(tlog.Log()+"Error in unmarshal :", err.Error())
		assert.Fail(t, "Unmarshal error")
	}
	returned := ConvertToResourceInsertArray(sampleBadData)
	if len(returned) != 0 {
		// Should be zero since the date would have forced a failure
		assert.Fail(t, "Should not have returned a resource to insert")
	}
}

// The first deployment should be a match and not need processing.   The second one should.
func Test_shouldProcess(t *testing.T) {
	log.Println(tlog.Log())
	var pnpRec = new(datastore.PnpStatusResource)
	err := json.Unmarshal([]byte(sampleWithDeploymentJSON), &pnpRec)
	if err != nil {
		log.Println(tlog.Log()+"Error in unmarshal :", err.Error())
		assert.Fail(t, "Unmarshal error")
	}
	counter := 0
	for _, v := range pnpRec.Deployments {
		var exist bool
		var hashString string
		resInsert := db.ConvertPnpDeploymentToResourceInsert(v, pnpRec.DisplayName)
		if counter == 1 {
			hashString = db.ComputeResourceRecordHash(resInsert)
			resourceMap.Resources[v.SourceID] = &(datastore.ResourceReturn{})
			resourceMap.Resources[v.SourceID].RecordHash = hashString
			resInsert.DisplayNames[0].Name = resInsert.DisplayNames[0].Name + "test"
			hashString = db.ComputeResourceRecordHash(resInsert)
			// We expect we will need to process now
			exist = shouldProcess(v, hashString)
			assert.True(t, exist)
		} else {
			// No changes so we don't need to process
			hashString = db.ComputeResourceRecordHash(resInsert)
			resourceMap.Resources[v.SourceID] = &(datastore.ResourceReturn{})
			resourceMap.Resources[v.SourceID].RecordHash = hashString
			exist = shouldProcess(v, hashString)
			assert.False(t, exist)
		}
		counter++
	}
}

func Test_getExistingResource(t *testing.T) {
	log.Println(tlog.Log())
	dbConn, _, mon := testutils.PrepareTestInc(t, msgExistingCase)
	defer db.Disconnect(dbConn)
	ctx := context.Background()
	span, ctx := ossmon.StartParentSpan(ctx, mon, monitor.SrvPrfx+"testing")
	defer span.Finish()
	t.Logf(tlog.Log())
	_, doesResourceAlreadyExist := getExistingResource(ctx, dbConn, "badResource", "")
	assert.Equal(t, doesResourceAlreadyExist, false)
}

