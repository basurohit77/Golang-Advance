package testDefs

import (
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.ibm.com/cloud-sre/pnp-nq2ds/helper"
)

const (
	// This allows the mock servers to work.  It's useful to turn off/set to false if you want to test against actual servers
	isLocal               = false
	subscriptionURLwSlash = "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/"

	subscriptionsBefore = `{
		"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions",
		"resources": [
		  {
			"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/2967b80d-6ddc-42fb-b761-cc4b7c618638",
			"recordId": "2967b80d-6ddc-42fb-b761-cc4b7c618638",
			"name": "test_gaas",
			"targetAddress": "https://api-pnp-rest-test:8000",
			"targetToken": "mytoken1",
			"watches": {
			  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/2967b80d-6ddc-42fb-b761-cc4b7c618638/watches"
			},
			"incidentWatch": {
			  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/2967b80d-6ddc-42fb-b761-cc4b7c618638/watches/incidents"
			},
			"maintenanceWatch": {
			  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/2967b80d-6ddc-42fb-b761-cc4b7c618638/watches/maintenance"
			},
			"resourceWatch": {
			  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/2967b80d-6ddc-42fb-b761-cc4b7c618638/watches/resources"
			},
			"caseWatch": {
			  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/2967b80d-6ddc-42fb-b761-cc4b7c618638/watches/case"
			},
			"notificationWatch": {
			  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/2967b80d-6ddc-42fb-b761-cc4b7c618638/watches/notifications"
			}
		  }
		]
	  }`

	subscriptionJSON = `{
		"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/bdc150ee-6205-46ae-b2cd-7e41b8ae0f55",
		"recordId": "bdc150ee-6205-46ae-b2cd-7e41b8ae0f55",
		"name": "my first",
		"targetAddress": "https://test.org/pnpintegrationtest/",
		"watches": {
		  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/bdc150ee-6205-46ae-b2cd-7e41b8ae0f55/watches"
		},
		"incidentWatch": {
		  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/bdc150ee-6205-46ae-b2cd-7e41b8ae0f55/watches/incidents"
		},
		"maintenanceWatch": {
		  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/bdc150ee-6205-46ae-b2cd-7e41b8ae0f55/watches/maintenance"
		},
		"resourceWatch": {
		  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/bdc150ee-6205-46ae-b2cd-7e41b8ae0f55/watches/resources"
		},
		"caseWatch": {
		  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/bdc150ee-6205-46ae-b2cd-7e41b8ae0f55/watches/case"
		},
		"notificationWatch": {
		  "href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/bdc150ee-6205-46ae-b2cd-7e41b8ae0f55/watches/notifications"
		}
	  }`

	subscriptionResponse = `{
		"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/3468cc09-4385-4e00-877d-2d8c11f167e7",
		"recordId": "3468cc09-4385-4e00-877d-2d8c11f167e7",
		"name": "my first",
		"targetAddress": "https://test.org/pnpintegrationtest/",
		"watches": {
			"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/3468cc09-4385-4e00-877d-2d8c11f167e7/watches"
		},
		"incidentWatch": {
			"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/3468cc09-4385-4e00-877d-2d8c11f167e7/watches/incidents"
		},
		"maintenanceWatch": {
			"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/3468cc09-4385-4e00-877d-2d8c11f167e7/watches/maintenance"
		},
		"resourceWatch": {
			"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/3468cc09-4385-4e00-877d-2d8c11f167e7/watches/resources"
		},
		"caseWatch": {
			"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/3468cc09-4385-4e00-877d-2d8c11f167e7/watches/case"
		},
		"notificationWatch": {
			"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/3468cc09-4385-4e00-877d-2d8c11f167e7/watches/notifications"
		}
	}`

	createWatchResponse = `{
		"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/3468cc09-4385-4e00-877d-2d8c11f167e7/watches/3cefabd5-60ea-41dd-aaaa-a4675d6607f9",
		"record_id": "3cefabd5-60ea-41dd-aaaa-a4675d6607f9",
		"subscription": {
			"href": "https://test.org/pnpsubscription/api/v1/pnp/subscriptions/3468cc09-4385-4e00-877d-2d8c11f167e7"
		},
		"kind": "incidentWatch",
		"crnMasks": [
			"crn:v1:bluemix:public:miketest:us-east::::"
		],
		"wildcards": "false",
		"subscriptionEmail": "michael_lee@us.ibm.com"
	}`

	incidentSNWriteResponse = `{
		"result": 
			{
				"sys_id": "sys_id",
				"sys_updated_on": "2019-12-03 13:51:56",
				"number": "INC1348439",
				"state": "1",
				"sys_created_by": "michael_lee@us.ibm.com",
				"delivery_plan": null,
				"cmdb_ci": "cmdb_ci",
				"u_crn": [
					"crn:v1:bluemix:public:cloud-object-storage:us-east::::"
				],
				"sys_created_on": "2019-12-03 13:42:20",
				"description": "This is a sample alert created from the API explorer application.  It is not a real incident.  This should only appear in the test environment.",
				"contact_type": "Manual",
				"incident_state": "1",
				"u_potential_cie": "0",
				"approval": "not requested",
				"u_environment": "u_environment",
				"u_tribe_name": "COS Storage"
			}
		
	}`

	incidentSNResponse = `{
		"results": [
			{
				"sys_id": "sysid",
				"sys_updated_on": "2019-12-03 13:51:56",
				"number": "INC1348439",
				"state": "1",
				"sys_created_by": "michael_lee@us.ibm.com",
				"delivery_plan": null,
				"cmdb_ci": "cmdb_ci",
				"u_crn": [
					"crn:v1:bluemix:public:cloud-object-storage:us-east::::"
				],
				"sys_created_on": "2019-12-03 13:42:20",
				"description": "This is a sample alert created from the API explorer application.  It is not a real incident.  This should only appear in the test environment.",
				"contact_type": "Manual",
				"incident_state": "1",
				"u_potential_cie": "0",
				"approval": "not requested",
				"u_environment": "u_environment",
				"u_tribe_name": "COS Storage"
			}
		]
	}`
)

func setup() {
	SubscriptionUrl = subscriptionURLwSlash
	SnURL = "https://test.org"
	SnAPIGetURL = SnURL + "/api/ibmwc/v2/incident/getIncidents"
	SnAPIURL = SnURL + "/api/ibmwc/v2/incident"
	os.Setenv("disableSkipTransport", "true")

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if isLocal {
		os.Setenv("NQ_URL", "amqp://guest:guest@localhost:5672") // # pragma: whitelist secret
	}
}

func TestSample(t *testing.T) {
	setup()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	r1 := httpmock.NewStringResponder(http.StatusOK, subscriptionsBefore)
	r1_createSub := httpmock.NewStringResponder(http.StatusOK, subscriptionJSON)
	httpmock.RegisterResponder("GET", subscriptionURLwSlash, r1)
	httpmock.RegisterResponder("POST", subscriptionURLwSlash, r1_createSub)

	// create sub
	sub, err := CreateSubscription("my first")
	if err != nil {
		if strings.Contains(err.Error(), "subscription count incorrect") {
			log.Print("This is expected due to complications in handling the mock requests")
		} else {
			log.Fatal(err.Error())
		}
	}

	watchUrlRegex := `=~^https://test.org/pnpsubscription/api/v1/pnp/subscriptions/(\w(-)?)+/watches`
	r2 := httpmock.NewStringResponder(http.StatusOK, createWatchResponse)
	httpmock.RegisterResponder("POST", watchUrlRegex, r2)

	// create watch
	watch := ReturnWatch("", "", CRNSample, false)
	CreateWatch(nil, sub.IncidentWatch.URL, watch)

	// Get subscription

	subscriptionUrlRegex := `=~^https://test\.org/pnpsubscription/api/v1/pnp/subscriptions/(\w(-)?)+$`
	r3 := httpmock.NewStringResponder(http.StatusOK, subscriptionResponse)
	httpmock.RegisterResponder("GET", subscriptionUrlRegex, r3)

	sub1, err := GetSubscription(nil, sub.SubscriptionURL)
	if err != nil {
		log.Fatalf("Error trying to get subscription  - %s \n\t %s", sub.SubscriptionURL, err.Error())
	}
	log.Print("Obtained subscription: ", sub1.SubscriptionURL)

	r4 := httpmock.NewStringResponder(http.StatusOK, "")
	httpmock.RegisterResponder("DELETE", sub.SubscriptionURL, r4)

	// clean up
	err = DeleteSubscription(nil, sub.SubscriptionURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Print("Properly cleaned up subscription subscription: ", sub1.SubscriptionURL)
}

func TestGetSubscriptionList(t *testing.T) {
	setup()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	r1 := httpmock.NewStringResponder(http.StatusOK, subscriptionsBefore)
	httpmock.RegisterResponder("GET", subscriptionURLwSlash, r1)

	subs, count, err := GetSubscriptionList(nil)
	if err != nil {
		log.Println(err.Error())
	}

	log.Print(count)
	log.Print(helper.GetPrettyJson(subs))
}

func TestGetIncident(t *testing.T) {
	setup()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	r1 := httpmock.NewStringResponder(http.StatusOK, incidentSNResponse)
	httpmock.RegisterResponder("GET", `https://test.org/api/ibmwc/v2/incident/getIncidents`, r1)

	sysid, err := GetIncidentFromSN(nil, "INC1348439")
	if err != nil {
		log.Fatal("Error getting incident:", err)
	}
	log.Print("Sysid: ", sysid)
}

func TestCreateIncident(t *testing.T) {
	setup()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	r1 := httpmock.NewStringResponder(http.StatusOK, incidentSNWriteResponse)

	createRegex := `=~https://test.org/api/ibmwc/v2/incident`
	patchRegex := `=~https://test.org/api/ibmwc/v2/incident/INC`
	httpmock.RegisterResponder("POST", createRegex, r1)
	httpmock.RegisterResponder("PATCH", patchRegex, r1)
	incidentID, _, err := CreateIncident()
	if err != nil || incidentID == "" {
		log.Fatal(err)
	}

	_, err = UpdateIncidentToCIE(incidentID)
	if err != nil {
		log.Fatal(err)
	}
}

func TestCreateWatch(t *testing.T) {
	setup()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	r1 := httpmock.NewStringResponder(http.StatusOK, createWatchResponse)
	httpmock.RegisterResponder(http.MethodPost, `https://test.org/`, r1)

	badWatch := []int{9}
	_, err := CreateWatch(nil, "https://test.org/", badWatch)
	assert.NotNil(t, err, "Should return error")
	assert.Contains(t, err.Error(), "invalid watch type", "Should return type error")

	watch := ReturnNotificationWatch("", CRNSample, false)
	watchReturn, err := CreateWatch(nil, "https://test.org/", watch)
	assert.Nil(t, err, "Should not return error")
	assert.NotNil(t, watchReturn, "Should return watch")
}

func TestPostMsg2RMQ(t *testing.T) {
	if isLocal {
		log.Print("This should not run if not local: ", isLocal)
		setup()
		PostMsg2RMQ("hi to the resource queue from rest test.....................................", "resource")
	}
}
