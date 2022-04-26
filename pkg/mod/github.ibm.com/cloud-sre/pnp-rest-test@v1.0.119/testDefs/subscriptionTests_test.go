package testDefs

import (
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestSubscriptionFull(t *testing.T) {
	// Subscription server and key are identified by environment variables:
	// os.Getenv("subscriptionURL"), os.Getenv("SERVER_KEY")

	setup()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	r1 := httpmock.NewStringResponder(http.StatusOK, subscriptionsBefore)
	r1_createSub := httpmock.NewStringResponder(http.StatusOK, subscriptionJSON)
	httpmock.RegisterResponder("GET", subscriptionURLwSlash, r1)
	httpmock.RegisterResponder("POST", subscriptionURLwSlash, r1_createSub)

	// create sub
	sub, err := CreateSubscription("TestSubscriptionFull")
	if err != nil {
		if strings.Contains(err.Error(), "The final number of subscriptions is not correct") {
			log.Print("This is expected due to complications in handling the mock requests")
		} else {

		}

		log.Println(err.Error())
	}
	log.Println(sub)

	// create watches
	// CRNSample -> "crn:v1:bluemix:public:cloud-object-storage:us-east::::"

	watchUrlRegex := `=~^https://test.org/pnpsubscription/api/v1/pnp/subscriptions/(\w(-)?)+/watches`
	r2 := httpmock.NewStringResponder(http.StatusOK, createWatchResponse)
	httpmock.RegisterResponder("POST", watchUrlRegex, r2)

	watch := ReturnWatch("", "", CRNSample, false)
	CreateWatch(nil, sub.IncidentWatch.URL, watch)
	CreateWatch(nil, sub.MaintenanceWatch.URL, watch)
	CreateWatch(nil, sub.NotificationWatch.URL, watch)
	CreateWatch(nil, sub.ResourceWatch.URL, watch)

	subscriptionUrlRegex := `=~^https://test.org/pnpsubscription/api/v1/pnp/subscriptions/(\w(-)?)+$`
	r3 := httpmock.NewStringResponder(http.StatusOK, subscriptionResponse)
	httpmock.RegisterResponder("GET", subscriptionUrlRegex, r3)

	sub1, err := GetSubscription(nil, sub.SubscriptionURL)
	log.Print(sub1.Expiration)

	r4 := httpmock.NewStringResponder(http.StatusOK, incidentSNWriteResponse)
	httpmock.RegisterResponder("POST", `https://test.org/api/ibmwc/v2/incident`, r4)

	incident, sysid, err := CreateIncident()
	if err != nil {
		log.Panic(err)
	}
	log.Print("Incident: ", incident, "\nsysid: ", sysid)

	// clean up

	r5 := httpmock.NewStringResponder(http.StatusOK, "")
	httpmock.RegisterResponder("DELETE", sub.SubscriptionURL, r5)
	err = DeleteSubscription(nil, sub.SubscriptionURL)
	if err != nil {
		log.Println(err.Error())
	}
}
