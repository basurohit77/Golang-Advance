package api

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	defaultInitialRetryDelay   = 2  // seconds
	defaultRetryDelayIncrement = 10 // seconds
	defaultMaxRetries          = 2
)

// MakeHTTPCallWithRetry will make an HTTP call using the provided client and request, and will retry
// the call if necessary.
// The status code received when making the call is compared to the provided expectedStatusCode.
// If there are no errors making the call, and the status code is as expected, the response is returned.
// If there is an error, or if the response code is not as expected, retries will be attempted and if
// still unsuccessful, an error will be returned.
// Note that if an HTTP 4xx response is received, the HTTP call is not reattempted as HTTP 4xx responses
// are client errors and not transient
func MakeHTTPCallWithRetry(client *http.Client, req *http.Request, expectedStatusCode int) (resp *http.Response, statusCode int, err error) {
	const FCT = "MakeHTTPCallWithRetry"

	retry := true // allow to run at least once
	retryDelay := defaultInitialRetryDelay
	maxRetries := defaultMaxRetries

	// read in request body to use for resetting the request body in retry loop
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			return resp, http.StatusInternalServerError, err
		}
	}

	for i := 0; retry; i++ {
		if req.Body != nil { // keep nil request body as is, otherwise reset request body
			req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, statusCode, err = makeHTTPCall(client, req, expectedStatusCode)
		if err != nil {
			// error, see if we can retry:
			retry = i < maxRetries && (statusCode < 400 || statusCode > 499)
			if retry {
				log.Print(FCT+": Error (", statusCode, ") received in HTTP call. Retry after ", retryDelay, " seconds")
				time.Sleep(time.Second * time.Duration(retryDelay))
				retryDelay = defaultRetryDelayIncrement
			}
		} else {
			// no error, no need to retry:
			retry = false
		}
	}

	return resp, statusCode, err
}

// makeHTTPCall makes an HTTP call using the provided client and request.
// The status code received when making the call is compared to the provided expectedStatusCode.
// If there are no errors making the call, and the status code is as expected, the response and
// status code are returned.
// If there is an error, or if the response code is not as expected, an error and status code is
// returned (HTTP 500 if the method not get to the point of actually making the call).
func makeHTTPCall(client *http.Client, req *http.Request, expectedStatusCode int) (*http.Response, int, error) {
	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			resp.Header.Set("X-ZAP", "OK")
		}
		return resp, http.StatusInternalServerError, err
	}

	if resp.StatusCode != expectedStatusCode {
		err = errors.New("Error: Request returned unexpected status=" + resp.Status)
		return resp, resp.StatusCode, err
	}

	return resp, resp.StatusCode, nil
}
