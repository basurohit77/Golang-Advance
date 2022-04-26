package testDefs

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	hlp "github.ibm.com/cloud-sre/pnp-nq2ds/helper"
	"github.ibm.com/cloud-sre/pnp-rest-test/lg"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
)

// Create a subscription
// Create a watch on maintenance and notification
// Post a maintenance update
// Should generate 2 hook updates
/* func Notifications(w http.ResponseWriter) string {
	const fct = "Notifications"
	lg.Info(fct, "starting")

	serv1 := rest.Server{}
	serv1.Token = os.Getenv("SERVER_KEY")
	subs1 := new(datastore.SubscriptionResponse)
	sub1 := new(datastore.SubscriptionSingleResponse)

	//Get subscriptions
	if err := serv1.GetAndDecode(fct, "SubscriptionReturn", subscriptionUrl, subs1); err != nil {
		log.Println(fct, "Error = "+err.Error())
		if !strings.Contains(err.Error(), "status code 404") {
			return "FAIL"
		}

	}
	lg.Info(fct, "Subscriptions: "+strconv.Itoa(len(subs1.Resources)))

	//
	//  Create a subscription
	if err := serv1.PostAndDecode(fct, "SubscriptionSingleResponse", subscriptionUrl, createSubPostBody, sub1); err != nil {
		log.Println(fct, "Error = "+err.Error())
		return "FAIL"

	}

	log.Println(fct, "create subscription created: return \n"+hlp.GetPrettyJson(sub1))

	//
	// create watches
	watchesReturned, err := createNotificationWatches(w, serv1, sub1)
	if err != nil {
		log.Println(fct, err.Error())
		return "FAIL"
	}

	//
	// get watches
	watches := datastore.WatchResponse{}

	if err := serv1.GetAndDecode(fct, "WatchReturnArray", sub1.Watches.URL, &watches); err != nil {
		log.Println(fct, "Error = "+err.Error())
		return "FAIL"

	}
	log.Println(fct, "\npost getting watches: \n"+hlp.GetPrettyJson(watches))

	// Make sure that all the watches we created are in the list we got back.  Kind of useless
	for _, v := range watchesReturned {
		if watchInArray(watches.Resources, v.RecordID) {
			continue
		} else {

			log.Println(v.RecordID + "was not created")
			return "FAIL"
		}

	}
	err = dumpWatchMap(w)
	if err != nil {
		log.Println(fct, "Error = "+err.Error())
		return "FAIL"

	}
	// Post a message to RMQ for case
	go post2RMQForNotifications(w)

	retVal := listenForWebhook(w, notificationChangeValidationStrings, 1)

	log.Println(fct, "Deleting : "+sub1.SubscriptionURL)
	cleanup(w, serv1, sub1)

	return retVal
}
*/
func createNotificationWatches(serv rest.Server, sub1 *datastore.SubscriptionSingleResponse) ([]datastore.WatchReturn, error) {
	const fct = "createWatches"

	maintenancewatch1 := datastore.WatchReturn{}
	if err := serv.PostAndDecode(fct, "createMaintenanceWatchPostBody", sub1.MaintenanceWatch.URL, createMaintenanceWatchPostBody, &maintenancewatch1); err != nil {
		log.Println(fct, "Error = "+err.Error())
		return nil, err

	}
	log.Println(fct, "create maintenance watch return: \n"+hlp.GetPrettyJson(maintenancewatch1))

	notificationwatch1 := datastore.WatchReturn{}
	if err := serv.PostAndDecode(fct, "createNotificationWatchPostBody", sub1.NotificationWatch.URL, createNotificationWatchPostBody, &notificationwatch1); err != nil {
		log.Println(fct, "Error = "+err.Error())
		return nil, err

	}
	log.Println(fct, "create notification watch return: \n"+hlp.GetPrettyJson(notificationwatch1))
	retWatches := []datastore.WatchReturn{maintenancewatch1, notificationwatch1}
	return retWatches, nil
}

func post2RMQForNotifications() error {
	const fct = "post2RMQForNotifications"
	lg.Info(fct, "********** Starting **********")
	time.Sleep(1 * time.Second)
	serv := rest.Server{}
	serv.Token = os.Getenv("HOOK_KEY")
	/* _, err := serv.Post(fct, RMQHooksCase, createCaseMessageInRMQ)
	if err != nil {
		log.Println(fct, "Error = "+err.Error())
		return err
	}
	log.Println(fct, "return from hooks for Case") */

	time.Sleep(1 * time.Second)
	changeUpdateTime := time.Now().Format("2006-01-02 15:04:05")
	lg.Info(fct, "Posting Snow Change", RMQHooksChange, createSNChangeMessageInRMQ[:50])
	_, err := serv.Post(fct, RMQHooksChange, fmt.Sprintf(createSNChangeMessageInRMQ, changeUpdateTime))
	if err != nil {
		log.Println(fct, "Error = "+err.Error())
		return err
	}

	return nil
}
