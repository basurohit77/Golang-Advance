package testDefs

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	hlp "github.ibm.com/cloud-sre/pnp-nq2ds/helper"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
)

const (
	SUCCESS_MSG = "Successfully completed test"
	FAIL_MSG    = "FAILED"
)

/* SAMPLE USAGE
// create sub
sub, err := testDefs.CreateSubscription("my first")
if err != nil {
	log.Println(err.Error())
}
log.Println(sub)

// create watch
watch := testDefs.ReturnWatch("", "", testDefs.CRNSample, false)
testDefs.CreateWatch(nil, sub.IncidentWatch.URL, watch)
sub1, err := testDefs.GetSubscription(nil, sub.SubscriptionURL)
log.Print(sub1.Expiration)

// clean up
err = testDefs.DeleteSubscription(nil, sub.SubscriptionURL)
if err != nil {
	log.Println(err.Error())
}
*/

func init() {
	if envType == "stage" {
		webhookTarget = "http://" + IngressIP + ":8000"
	}
}

// Create a subscription
// Create a watch on incident and resources
// Post an incident update, resource update and a status update
// Should generate 3 hook updates
func SimpleRun(w http.ResponseWriter) string {
	const fct = "[SimpleRun]"
	log.Println(fct, "starting")

	server := rest.Server{}
	server.Token = os.Getenv("SERVER_KEY")
	subscriptionsReturned := new(datastore.SubscriptionResponse)
	sub1 := new(datastore.SubscriptionSingleResponse)

	// Get subscriptions
	if err := server.GetAndDecode(fct, "SubscriptionReturn", SubscriptionUrl, subscriptionsReturned); err != nil {
		log.Println(fct, err.Error())
		if !strings.Contains(err.Error(), "status code 404") {
			return FAIL_MSG
		}
	}

	log.Println(fct, "Subscriptions:", len(subscriptionsReturned.Resources))

	// Create a subscription
	if err := server.PostAndDecode(fct, "SubscriptionSingleResponse", SubscriptionUrl, createSubPostBody, sub1); err != nil {
		log.Println(fct, err.Error())
		return FAIL_MSG
	}

	log.Println(fct, "Subscription created:", hlp.GetPrettyJson(sub1))

	// Create watches
	watchesReturned, err := CreateBasicWatches(server, sub1)
	if err != nil {
		log.Println(fct, err.Error())
		return FAIL_MSG
	}

	// Get watches
	watches := datastore.WatchResponse{}
	if err := server.GetAndDecode(fct, "WatchReturnArray", sub1.Watches.URL, &watches); err != nil {
		log.Println(fct, err.Error())
		return FAIL_MSG

	}

	log.Println(fct, "Returned watches:", hlp.GetPrettyJson(watches))

	for _, w := range watchesReturned {
		if !watchInArray(watches.Resources, w.RecordID) {
			log.Println(fct, w.RecordID, "was not created")
			return FAIL_MSG
		}
	}

	// Post a message to RMQ for case
	go postBasic2RMQ(w)
	/*
		err = dumpWatchMap(w)
		if err != nil {
			log.Println(fct, err.Error())
			return FAIL_MSG
		}
	*/
	retVal := listenForWebhook(basicValidationStrings, 2)

	log.Println(fct, "Deleting:", sub1.SubscriptionURL)
	err = cleanup(server, sub1)
	if err != nil {
		log.Println(fct, err.Error())
	}

	return retVal
}

func RunSubscriptionFull() string {
	const fct = "[RunSubscriptionFull]"
	//var wg sync.WaitGroup
	// Subscription server and key are identified by environment variables:
	// os.Getenv("subscriptionURL"), os.Getenv("SERVER_KEY")

	// create sub
	// runSubscriptionFullValidationStrings_bk := []validation{
	// 	{"/api/v1/pnp/status/incidents/", "for incidents"},
	// 	{"/api/v1/pnp/status/resources/", "for resources/status"},
	// 	{"/api/v1/pnp/status/notifications/", "for notifications"},
	// }

	var runSubscriptionFullValidationStrings []validation
	log.Println(fct, "Resetting validation strings:", runSubscriptionFullValidationStrings)

	timestamp := time.Now().Format("Jan2-1504")

	cleanSubs, _, err := GetSubscriptionList(nil)
	if err != nil {
		log.Println(fct, "->GetSubscriptionList:", err.Error())
	}

	for _, v := range cleanSubs.Resources {
		if strings.Contains(v.Name, "TestSubscription") {
			err = DeleteSubscription(nil, v.SubscriptionURL)
			if err != nil {
				log.Println(fct, "->DeleteSubscription:", err.Error())
			}
		} else {
			log.Println(fct, "->skipping delete:", v.Name)
		}
	}

	// Create subscription
	sub, err := CreateSubscription("TestSubscriptionFull_" + timestamp)
	if err != nil {
		log.Println(fct, "->CreateSubscription:", err.Error())
		return FAIL_MSG
	}
	log.Println(fct, "->CreateSubscription:", sub)

	// Create watches
	watch := ReturnWatch("", "", CRNSample, false)
	_, err = CreateWatch(nil, sub.IncidentWatch.URL, watch)
	if err != nil {
		log.Println(fct, "->CreateWatch:", err.Error())
		return FAIL_MSG
	}
	_, err = CreateWatch(nil, sub.MaintenanceWatch.URL, watch)
	if err != nil {
		log.Println(fct, "->CreateWatch:", err.Error())
		return FAIL_MSG
	}

	notificationWatch := ReturnNotificationWatch("", CRNSample2, true)
	_, err = CreateWatch(nil, sub.NotificationWatch.URL, notificationWatch)
	if err != nil {
		log.Println(fct, "->CreateWatch:", err.Error())
		return FAIL_MSG
	}
	_, err = CreateWatch(nil, sub.ResourceWatch.URL, notificationWatch)
	if err != nil {
		log.Println(fct, "->CreateWatch:", err.Error())
		return FAIL_MSG
	}
	log.Println(fct, "Successfully created watches")

	sub1, err := GetSubscription(nil, sub.SubscriptionURL)
	if err != nil {
		log.Println(fct, "->GetSubscription:", err.Error())
		return FAIL_MSG
	}
	log.Println(fct, "->GetSubscription:", sub1)

	retVal := SUCCESS_MSG
	// go func() {
	// 	wg.Add(1)
	// 	retVal = listenForWebhook(runSubscriptionFullValidationStrings, 4)
	// 	wg.Done()
	// }()

	////////////////////////////////////////////

	// log.Print("\n\nNow we send the data...\n")

	// log.Print("**************************** Posting Doctor to hooks server ...\n")
	// // Dr maintenances will get a resource and notification update (for open)
	// drRecId, err := PostDr2Hooks("crn:v1:bluemix:public:cloud-object-storage:us-south::::")
	// if err != nil {
	// 	log.Print("\n\nFailure posting Dr maintenance: ", err)
	// }
	// log.Print("Adding doctor maintenance to validation strings : drRecId: " + drRecId)
	// Messages <- "Validation update|-1|" + runSubscriptionFullValidationStrings_bk[3].Value + drRecId + "| doctor maintenance record id"

	// resourceRecordID := db.CreateRecordIDFromSourceSourceID("globalCatalog", "crn:v1:bluemix:public:cloud-object-storage:us-south::::")
	// log.Print("Adding resource to validation strings: record id : " + resourceRecordID + "\t globalCatalog id : " + "globalCatalog" + "\t crn : " + "crn:v1:bluemix:public:cloud-object-storage:us-south::::")

	// Messages <- "Validation update|-1|" + runSubscriptionFullValidationStrings_bk[2].Value + resourceRecordID + "| resource record id for the doctor maintenance"

	// time.Sleep(3 * time.Second)
	// log.Print("**************************** Posting BSPN to hooks server ...\n")
	// bspnRecId, err := PostBspn2Hooks(DefaultCRN)
	// if err != nil {
	// 	log.Print("\n\nFailure posting BSPN notification: ", err)
	// }

	// runSubscriptionFullValidationStrings = append(runSubscriptionFullValidationStrings, validation{"api/v1/pnp/status/notifications/" + bspnRecId, "Notification for BSPN Line 226"})

	//	log.Println(fct, "**************************** Creating test incident...")
	// each incident gets 2 of each type(incident, status, notification) for in-progress and resolved
	// incidentID, sysid, err := CreateIncident()
	// if err != nil {
	// 	if strings.Contains(err.Error(), "(error:json: cannot unmarshal bool into Go struct field Result.result of type snAdapter.SNRecord)") {
	// 		log.Println(fct, "Returning SUCCESS for now until watsondev is fixed for GAAS")
	// 		return SUCCESS_MSG
	// 	}

	// 	log.Println(fct, "->CreateIncident:", err.Error())
	// 	return FAIL_MSG
	// }

	// log.Println(fct, "--- Test incident successfully created :", incidentID)

	// incidentRecordID := db.CreateRecordIDFromSourceSourceID("servicenow", incidentID)
	// log.Println(fct, "Sending validation updates for incident -> recordID:", incidentRecordID)
	// Messages <- "Validation update|-1|" + runSubscriptionFullValidationStrings_bk[0].Value + incidentRecordID + "|incident is in progress"
	// Messages <- "Validation update|-1|" + runSubscriptionFullValidationStrings_bk[0].Value + incidentRecordID + "|incident is resolved"

	// resourceRecordID := db.CreateRecordIDFromSourceSourceID("globalCatalog", DefaultCRN)
	// log.Println(fct, "Sending validation updates for resource -> recordID:", resourceRecordID, "- globalCatalog DefaultCRN: ", DefaultCRN)
	// Messages <- "Validation update|-1|" + runSubscriptionFullValidationStrings_bk[1].Value + resourceRecordID + "|resource is in progress"
	// Messages <- "Validation update|-1|" + runSubscriptionFullValidationStrings_bk[1].Value + resourceRecordID + "|resource is resolved"

	// No longer creating notification on incidents?
	// //  notification open and close
	// notificationId := db.CreateNotificationRecordID("servicenow",  incident, DefaultCRN)
	// log.Print("Adding notificationId to validation strings (for open and close) : " + notificationId + "\t incident id : " + incident + "\t DefaultCRN id : " + DefaultCRN)
	// Messages <- "Validation update|-1|" + runSubscriptionFullValidationStrings_bk[3] + notificationId
	// Messages <- "Validation update|-1|" + runSubscriptionFullValidationStrings_bk[3] + notificationId

	// log.Println(fct, "Updating incident to CIE - sysid:", sysid)
	// _, err = UpdateIncidentToCIE(sysid)
	// if err != nil {
	// 	log.Println(fct, "->UpdateIncidentToCIE:", err.Error())
	// 	return FAIL_MSG
	// }
	//time.Sleep(3 * time.Second)
	// log.Print("\nWill resolve : ", sysid)
	// _, err = UpdateIncidentToResolved(nil, sysid)
	// if err != nil {
	// 	log.Panic(err)
	// }

	// wg.Wait()

	// time.Sleep(3 * time.Second)
	// clean up
	// err = DeleteSubscription(nil, sub.SubscriptionURL)
	// if err != nil {
	// 	log.Println("DeleteSubscription: ", err.Error())
	// }

	return retVal
}
