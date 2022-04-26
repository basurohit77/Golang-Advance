package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/pnp-hooks/shared"

	"github.com/stretchr/testify/assert"
)

func TestSnowCasesValidToken(t *testing.T) {

	if err := shared.ConnectProducerRabbitMQ(); err != nil {
		t.Error("unable to connect to rabbitmq. err=", err)
	}

	// defer shared.ProducerConn.Close()
	// defer shared.ProducerCh.Close()

	httpPost(t, "api/v1/snow/cases", "", http.StatusOK, ServeSnowCases)

}

// func TestSnowCasesValidTokenRabbitMQErr(t *testing.T) {
// 	shared.Producer = nil
// 	httpPost(t, "api/v1/snow/cases", "", http.StatusInternalServerError, ServeSnowCases, os.Getenv("SNOW_TOKEN"))
// }

func httpPost(t *testing.T, url, expectedResponse string, expectedStatus int, handlerfunc http.HandlerFunc) {

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		strings.NewReader("some test data"))
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
