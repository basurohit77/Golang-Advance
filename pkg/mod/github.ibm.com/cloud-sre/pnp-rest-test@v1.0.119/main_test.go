package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.ibm.com/cloud-sre/pnp-rest-test/testDefs"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func Test_SendWebhook(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	os.Setenv("disableSkipTransport", "true")
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	r1 := httpmock.NewStringResponder(http.StatusOK, "")
	httpmock.RegisterResponder("POST", "http://test.org", r1)

	sendWebhook(nil, "http://test.org", "This is a body")

	output := buf.String()
	assert.NotContains(t, output, "Error:", "Output contains an error: "+output)
}

func Test_Help(t *testing.T) {
	w := httptest.NewRecorder()
	help(w, nil)
	outputBytes, err := ioutil.ReadAll(w.Body)
	output := string(outputBytes)
	assert.NoError(t, err, "Error should be nil: ")
	assert.Contains(t, output, "curl localhost:8000/run/?basic", "Output should contain curl commands : \n"+string(output))
}

func Test_RunHandler(t *testing.T) {
	hooksHandler = func(w http.ResponseWriter) string {
		fmt.Fprint(w, "successful run")
		return ""
	}
	r := httptest.NewRequest("GET", "/run?basic", nil)
	w := httptest.NewRecorder()
	runHandler(w, r)
	outputBytes, err := ioutil.ReadAll(w.Body)
	output := string(outputBytes)
	assert.NoError(t, err, "error should be nil")
	assert.Contains(t, output, "successful run", "Output should indicate a successful run: "+output)

	/* ------------------- */

	httpmock.Activate()
	httpmock.Reset()
	r1 := httpmock.NewStringResponder(http.StatusOK, `{"status":"ok"}`)
	httpmock.RegisterResponder("POST", "http://test.org", r1)

	runFullSubscription = func() string {
		return testDefs.SUCCESS_MSG
	}
	r = httptest.NewRequest("GET", "/run?RunSubscriptionFull&address=test.org", nil)
	w = httptest.NewRecorder()
	runHandler(w, r)
	returnCode := w.Result().StatusCode
	assert.Equal(t, http.StatusOK, returnCode, "Expecting return code 200")

	/* ------------------- */

	resourceAdapterIntegrationTest = func() string {
		return testDefs.SUCCESS_MSG
	}
	r = httptest.NewRequest("GET", "/run?ResourceAdapterIntegrationTest&address=test.org", nil)
	w = httptest.NewRecorder()
	runHandler(w, r)
	returnCode = w.Result().StatusCode
	assert.Equal(t, http.StatusOK, returnCode, "Expecting return code 200")

	/* ------------------- */

	// caseAPIIntegrationTest = func() string {
	// 	return testDefs.FAIL_MSG
	// }
	// r = httptest.NewRequest("GET", "/run?CaseAPIIntegrationTest&address=test.org", nil)
	// w = httptest.NewRecorder()
	// runHandler(w, r)
	// returnCode = w.Result().StatusCode
	// assert.Equal(t, http.StatusInternalServerError, returnCode, "Expecting return code 500")
}
