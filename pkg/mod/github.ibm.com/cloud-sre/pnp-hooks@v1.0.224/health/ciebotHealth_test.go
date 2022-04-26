package health

import (
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestCheckCiebotHealth(t *testing.T) {
	httpmock.Activate()

	res := `{"href":"http://localhost/health","code":0,"description":"The API is available and operational."}`

	r := httpmock.NewStringResponder(http.StatusOK, res)

	httpmock.RegisterResponder("GET", "http://localhost/health", r)

	c := NewCiebotHealth("http://localhost/health", "http://localhost/health", "http://localhost/health", false)

	healthy, healthDescription := c.CheckCiebotServices()
	
	if !healthy {
		t.Fatal("One ciebot serivce is not healthy: " + healthDescription)
	}
}

func TestCheckCiebotHealthWithCheckDisabled(t *testing.T) {
	httpmock.Activate()

	res := `{"href":"http://localhost/health1","code":0,"description":"The API is available and operational."}`

	// Track the number of calls made to CIEBot healthz APIs:
	numberOfCallsMadeToBackendCiebotHealthChecks := 0
	httpmock.RegisterResponder("GET", "http://localhost/health1",
		func(req *http.Request) (*http.Response, error) {
			numberOfCallsMadeToBackendCiebotHealthChecks++
			return httpmock.NewStringResponse(http.StatusOK, res), nil
		},
	)

	// Passing in true to tell the CiebotHealth to not perform backend health checks:
	c := NewCiebotHealth("http://localhost/health1", "http://localhost/health1", "http://localhost/health1", true)

	// Sleep for a bit to give the NewCiebotHealth call a chance to execute:
	time.Sleep(time.Second * 5)

	// Ensure overall CIEBot health is true since we are not performing backend health checks:
	isHealthy, description := c.CheckCiebotHealth()
	if !isHealthy {
		t.Fatal("Expected call to CheckCiebotHealth to return healthy, but it returned unhealthy")
	}
	if description != "" {
		t.Fatal("Expected call to CheckCiebotHealth to return an empty description, but it returned the following description:", description)
	}

	// Given that we told CiebotHealth to not perform health checks, the number of calls should be 0:
	if numberOfCallsMadeToBackendCiebotHealthChecks != 0 {
		t.Fatal("Expected CIEBot backend health checks to be called 0 times, but they were called", numberOfCallsMadeToBackendCiebotHealthChecks, "times")
	}
}
