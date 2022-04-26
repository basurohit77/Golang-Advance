package testDefs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	hlp "github.ibm.com/cloud-sre/pnp-nq2ds/helper"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
)

func GetSubscriptionList(server *rest.Server) (datastore.SubscriptionResponse, int, error) {
	const fct = "[GetSubscriptionList]"
	if server == nil {
		server = &rest.Server{}
		server.Token = os.Getenv("SERVER_KEY")
	}

	subscriptionsReturned := new(datastore.SubscriptionResponse)

	// Get subscription list
	err := server.GetAndDecode(fct, "SubscriptionReturn", SubscriptionUrl, subscriptionsReturned)
	if err != nil {
		log.Println(fct, err.Error())
		if strings.Contains(err.Error(), "404 Not Found") { // returns empty subscription list
			err = nil
		}
		return *subscriptionsReturned, 0, err
	}

	log.Println(fct, "Subscriptions:", len(subscriptionsReturned.Resources))
	return *subscriptionsReturned, len(subscriptionsReturned.Resources), nil
}

func GetSubscription(server *rest.Server, subscriptionUrl string) (*datastore.SubscriptionSingleResponse, error) {
	const fct = "[GetSubscription]"
	log.Println(fct, "starting")
	if server == nil {
		server = &rest.Server{}
		server.Token = os.Getenv("SERVER_KEY")
	}

	subscriptionReturned := new(datastore.SubscriptionSingleResponse)

	// Get subscription
	err := server.GetAndDecode(fct, "SubscriptionReturn", subscriptionUrl, subscriptionReturned)
	if err != nil {
		log.Println(fct, err.Error())
		return nil, err
	}

	return subscriptionReturned, nil
}

func CreateSubscription(name string) (SubscriptionResponse datastore.SubscriptionSingleResponse, err error) {
	const fct = "[CreateSubscription]"
	log.Println(fct, "starting")

	server := rest.Server{}
	server.Token = os.Getenv("SERVER_KEY")

	_, initialCount, err := GetSubscriptionList(&server)
	if err != nil {
		log.Println(fct, err.Error())
		return SubscriptionResponse, err
	}

	log.Println(fct, "webhookTarget:", webhookTarget)
	if webhookTarget == "" {
		return SubscriptionResponse, errors.New("webhookTarget is empty")
	}

	subscription := Subscription{}
	subscription.Name = name
	subscription.TargetAddress = webhookTarget

	subscriptionByte, err := json.Marshal(subscription)
	if err != nil {
		log.Println(fct, err.Error())
		return SubscriptionResponse, err
	}

	// Create a subscription
	err = server.PostAndDecode(fct, "SubscriptionSingleResponse", SubscriptionUrl, string(subscriptionByte), &SubscriptionResponse)
	if err != nil {
		log.Println(fct, err.Error())
		return SubscriptionResponse, err
	}

	log.Println(fct, "Subscription created:", hlp.GetPrettyJson(SubscriptionResponse))
	_, finalCount, err := GetSubscriptionList(&server)
	if err != nil {
		log.Println(fct, err.Error())
		return SubscriptionResponse, err
	}

	if finalCount != (initialCount + 1) {
		errMsg := fmt.Sprintf("Error: subscription count incorrect -> initial: %d final: %d", initialCount, finalCount)
		log.Println(fct, errMsg)
		return SubscriptionResponse, errors.New(errMsg)
	}

	return SubscriptionResponse, nil
}

func CreateBasicWatches(server rest.Server, sub1 *datastore.SubscriptionSingleResponse) ([]datastore.WatchReturn, error) {
	const fct = "[CreateBasicWatches]"

	incidentWatch := datastore.WatchReturn{}
	log.Println(fct, createIncidentWatchPostBody)
	if err := server.PostAndDecode(fct, "createIncidentWatchPostBody", sub1.IncidentWatch.URL, createIncidentWatchPostBody, &incidentWatch); err != nil {
		log.Println(fct, err.Error())
		return nil, err
	}
	log.Println(fct, "Incident watch created:", hlp.GetPrettyJson(incidentWatch))

	resourceWatch := datastore.WatchReturn{}
	if err := server.PostAndDecode(fct, "createStatusWatchPostBody", sub1.ResourceWatch.URL, createResourceWatchPostBody, &resourceWatch); err != nil {
		log.Println(fct, err.Error())
		return nil, err
	}
	log.Println(fct, "Status watch created:", hlp.GetPrettyJson(resourceWatch))

	retWatches := []datastore.WatchReturn{incidentWatch, resourceWatch}
	return retWatches, nil
}

// GetWatches - returns the watches in a subscription. Must provide either the url or the subscription object
func GetWatches(serv *rest.Server, subscription *datastore.SubscriptionSingleResponse, url string) (watches *datastore.WatchResponse, err error) {
	const fct = "GetWatches"
	if serv == nil {
		serv = &rest.Server{}
		serv.Token = os.Getenv("SERVER_KEY")
	}

	if subscription != nil {
		url = subscription.Watches.URL
	} else if url == "" {
		log.Println(fct, "No url or subscription provided")
		return nil, errors.New("No url or subscription provided")
	}
	if err := serv.GetAndDecode(fct, "WatchReturnArray", subscription.Watches.URL, &watches); err != nil {
		log.Println(fct, err.Error())
		return nil, err

	}

	log.Println(fct, "watches: \n"+hlp.GetPrettyJson(watches))
	return watches, err

}

func CreateWatch(server *rest.Server, url string, watch interface{}) (watchReturn *datastore.WatchReturn, err error) {
	const fct = "[CreateWatch]"
	if server == nil {
		server = &rest.Server{}
		server.Token = os.Getenv("SERVER_KEY")
	}

	_, isWatch := watch.(Watch)
	_, isNotificationWatch := watch.(NotificationWatch)
	if !isWatch && !isNotificationWatch {
		errMsg := fmt.Sprintf("Error: invalid watch type %T", watch)
		log.Println(fct, errMsg)
		return nil, errors.New(errMsg)
	}

	watchByte, err := json.Marshal(watch)
	if err != nil {
		log.Println(fct, err.Error())
		return nil, err
	}

	log.Println(fct, "PostBody:", string(watchByte))
	if err := server.PostAndDecode(fct, "CreateWatch", url, string(watchByte), &watchReturn); err != nil {
		log.Println(fct, err.Error())
		return nil, err
	}

	log.Println(fct, "Watch created:", hlp.GetPrettyJson(watchReturn))
	return watchReturn, nil
}

// ReturnWatch simplifies the creation of a Watch structure.
// Useful for creating a watch through the API via the CreateWatch() function.
// Records and crns are comma separated strings that will become string arrays
func ReturnWatch(path, records, crns string, wildcard bool) Watch {
	return Watch{
		Path:            path,
		RecordIDToWatch: strings.Split(records, ","),
		CrnMasks:        strings.Split(crns, ","),
		Wildcards:       strconv.FormatBool(wildcard),
	}
}

func ReturnNotificationWatch(path, crns string, wildcard bool) NotificationWatch {
	return NotificationWatch{
		Path:      path,
		CrnMasks:  strings.Split(crns, ","),
		Wildcards: strconv.FormatBool(wildcard),
	}
}

func DeleteSubscription(server *rest.Server, subscriptionURL string) error {
	const fct = "[DeleteSubscription]"
	if server == nil {
		server = &rest.Server{}
		server.Token = os.Getenv("SERVER_KEY")
	}
	_, err := server.Delete(fct, subscriptionURL)
	if err != nil {
		log.Println(fct, err.Error())
		return err
	}
	return nil
}

func createWatches(serv rest.Server, sub1 *datastore.SubscriptionSingleResponse) ([]datastore.WatchReturn, error) {
	const fct = "createWatches"

	/* casewatch1 := datastore.WatchReturn{}
	if err := serv.PostAndDecode(fct, "createCaseWatchPostBody", sub1.CaseWatch.URL, createCaseWatchPostBody, &casewatch1); err != nil {
		log.Println(fct, "Error = "+err.Error())
		return nil, err

	}
	log.Println(fct, "create case watch return: \n"+hlp.GetPrettyJson(casewatch1))
	*/
	incidentwatch1 := datastore.WatchReturn{}
	log.Println(fct, createIncidentWatchPostBody)
	if err := serv.PostAndDecode(fct, "createIncidentWatchPostBody", sub1.IncidentWatch.URL, createIncidentWatchPostBody, &incidentwatch1); err != nil {
		log.Println(fct, "Error = "+err.Error())
		return nil, err

	}
	log.Println(fct, "create incident watch return: \n"+hlp.GetPrettyJson(incidentwatch1))

	resourcewatch1 := datastore.WatchReturn{}
	if err := serv.PostAndDecode(fct, "createStatusWatchPostBody", sub1.ResourceWatch.URL, createResourceWatchPostBody, &resourcewatch1); err != nil {
		log.Println(fct, "Error = "+err.Error())
		return nil, err

	}
	log.Println(fct, "create status watch return: \n"+hlp.GetPrettyJson(resourcewatch1))

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

	//retWatches := []datastore.WatchReturn{casewatch1, incidentwatch1, resourcewatch1, maintenancewatch1}
	retWatches := []datastore.WatchReturn{incidentwatch1, resourcewatch1, maintenancewatch1}
	return retWatches, nil
}
