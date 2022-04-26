package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func Test_retry(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	//query category - not found
	testUrl := "http://test/test"
	httpmock.RegisterResponder("GET", testUrl, httpmock.NewStringResponder(http.StatusBadGateway, "test internal error"))
	req, err := http.NewRequest(http.MethodGet, testUrl, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{
		Timeout: 1000 * time.Second,
	}
	res, _, err := MakeHTTPCallWithRetry(client, req, http.StatusOK)
	defer res.Body.Close()
	assert.NotNil(t, err, "Should return error")
	assert.Equal(t, http.StatusBadGateway, res.StatusCode)
}

func Test_retryWithBody(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	testUrl := "http://test/test"
	testBody := "test body"
	httpmock.RegisterResponder(http.MethodPost, testUrl,
		func(req *http.Request) (*http.Response, error) {
			bodyBytes, _ := ioutil.ReadAll(req.Body)
			assert.Equal(t, testBody, string(bodyBytes))
			return httpmock.NewStringResponse(http.StatusBadGateway, "test internal error"), nil
		},
	)

	req, err := http.NewRequest(http.MethodPost, testUrl, bytes.NewReader([]byte(testBody)))
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{
		Timeout: 1000 * time.Second,
	}

	res, _, err := MakeHTTPCallWithRetry(client, req, http.StatusOK)
	defer res.Body.Close()
	assert.NotNil(t, err, "Should return error")
	assert.Equal(t, http.StatusBadGateway, res.StatusCode)
}
