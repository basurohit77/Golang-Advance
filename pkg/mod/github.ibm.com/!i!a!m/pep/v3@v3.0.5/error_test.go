package pep

import "testing"
import "github.com/stretchr/testify/assert"

func TestAPIErrorMessage(t *testing.T) {

	sampleAPIErr := APIError{EndpointURI: "api.test.com/v2/test",
		RequestHeaders:  "reqheader1",
		ResponseHeaders: "respheader2",
		Message:         "testmessage",
		StatusCode:      500,
		Trace:           "1234",
	}
	assert.EqualError(t, &sampleAPIErr, "API error code: 500 calling api.test.com/v2/test for txid: 1234 Details: testmessage")
}
func TestInternalErrorMessage(t *testing.T) {

	sampleInternalErr := InternalError{Message: "testmessage",
		Trace: "1234",
	}
	assert.EqualError(t, &sampleInternalErr, "Internal error testmessage while processing txid: 1234")
}
