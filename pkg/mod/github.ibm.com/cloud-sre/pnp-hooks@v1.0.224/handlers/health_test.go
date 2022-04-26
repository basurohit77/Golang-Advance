package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/cloud-sre/pnp-hooks/shared"
)

/*
	All tests expect a sucessful response body & response code
*/

// testHealthz tests /healthz endpoint
func TestHealthz(t *testing.T) {

	healthzHandler := NewHealthzHandler()
	healthzHandler.KongURL = "https://api-oss.cloud.ibm.com"
	healthzHandler.APIRequestPathPrefix = "pnphooks"
	healthzHandler.SubPath = shared.APIHealthzPath

	expectedResponse := `{"href":"https://api-oss.cloud.ibm.com/pnphooks/api/v1/pnp/hooks/healthz","code":0,"description":"The API is available and operational."}`

	httpGet(t, "/pnphooks/api/v1/pnp/hooks/healthz", expectedResponse, http.StatusOK, healthzHandler.ServeHealthz)
}

func httpGet(t *testing.T, url, expectedResponse string, expectedStatus int, handlerfunc http.HandlerFunc) {

	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlerfunc)

	handler.ServeHTTP(rr, req)

	// ensure expected status and returned status match
	assert.Equal(t, expectedStatus, rr.Code)

	// ensure expected response and returned response match
	assert.Equal(t, expectedResponse, strings.TrimSuffix(rr.Body.String(), "\n"))
}
