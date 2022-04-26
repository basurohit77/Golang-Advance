package testDefs

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
)

func Test_watchInArray(t *testing.T) {
	watches := []datastore.WatchReturn{}
	watch1 := datastore.WatchReturn{}
	watch1.Href = "http://test1"
	watch1.RecordID = "123"
	watch2 := datastore.WatchReturn{}
	watch2.Href = "http://test2"
	watch2.RecordID = "234"
	watches = append(watches, watch1)
	watches = append(watches, watch2)

	inArray := watchInArray(watches, "abc")
	assert.False(t, inArray)
	inArray = watchInArray(watches, "123")
	assert.True(t, inArray)
}

func Test_cleanup(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	r1 := httpmock.NewStringResponder(http.StatusOK, "")
	httpmock.RegisterResponder("DELETE", "http://test.org", r1)
	server := rest.Server{}
	sub := datastore.SubscriptionSingleResponse{}
	sub.SubscriptionURL = "http://test.org"
	err := cleanup(server, &sub)
	assert.NoError(t, err)
}

func Test_isValidBody(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	r1 := httpmock.NewStringResponder(http.StatusInternalServerError, "internal server error")
	r2 := httpmock.NewStringResponder(http.StatusOK, `{"recordID": "123", "href": "valid"}`)
	httpmock.RegisterResponder("GET", "http://test.org", r1)
	httpmock.RegisterResponder("GET", "http://test2.org", r2)

	msg := "Body : test"

	isValid := isValidBody(msg, "")
	assert.False(t, isValid, "Should not be valid")

	// -----------------------------------
	msg = `Body : { "href": "http://test.org" }`

	isValid = isValidBody(msg, "")
	assert.False(t, isValid, "Should not be valid")

	// -----------------------------------
	msg = `Body : { "href": "http://test2.org" }`

	isValid = isValidBody(msg, "invalid")
	assert.False(t, isValid, "Should not be valid")

	isValid = isValidBody(msg, "valid")
	assert.True(t, isValid, "Should be valid")
}

func Test_listenForWebhook(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	r1 := httpmock.NewStringResponder(http.StatusOK, `{"recordID": "123", "href": "http://AdditionalValid"}`)
	httpmock.RegisterResponder("GET", "http://AdditionalValid", r1)
	r2 := httpmock.NewStringResponder(http.StatusOK, `{"recordID": "123", "href": "http://initial"}`)
	httpmock.RegisterResponder("GET", "http://initial", r2)

	validations := []validation{{"initial", "Description of validation"}}

	go func() {
		time.Sleep(1 * time.Second)
		Messages <- "Validation update|-1|AdditionalValid|Description of validation add"
	}()
	go func() {
		time.Sleep(1 * time.Second)
		Messages <- "Validation update|-1|AdditionalValid|Description of validation add 2"
	}()
	go func() {
		time.Sleep(2 * time.Second)
		Messages <- `Body : {"href": "http://AdditionalValid"}`
	}()
	go func() {
		time.Sleep(3 * time.Second)
		Messages <- `Body : {"href": "http://initial"}`
	}()

	returned := listenForWebhook(validations, 2)
	assert.Equal(t, SUCCESS_MSG, returned, "Should have gotten a SUCCESS message: " + returned)
}
