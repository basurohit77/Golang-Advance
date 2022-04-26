package pep

// Eventually this should have its own package

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// RequestResponseFunc is a function that return *http.Response or an error given *http.Request
// Eventually this shoould be moved this to its own module to be shared with the attribute server.
type RequestResponseFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface
func (f RequestResponseFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

//NewHTTPMocker returns *http.Client that uses the user-provided transport
func NewHTTPMocker(fn RequestResponseFunc) *http.Client {
	return &http.Client{
		Transport: RequestResponseFunc(fn),
	}
}

// mockedResponse defines a json and its http status code.
// If the error is not nil, it will be returned instead.
type mockedResponse struct {
	header     map[string]string
	json       string
	statusCode int
	error      error // if not nil, this error will be returned of the mocked data
}

// JSONMockerFunc is a function that return a mockedResponse given *http.Request
type JSONMockerFunc func(req *http.Request) mockedResponse

//mockedData contains
// - an array of JSONMockerFunc to be returned
// - an indication if this data is for the bulk or authz api
type mockedData struct {
	responders []JSONMockerFunc
	bulk       bool //True if bulk api, False if authz
}

// PDPMocker mimicks pdp by returning mocked json responses.
type PDPMocker struct {
	httpClient *http.Client // an http client that returns a fake json response
	counter    func() int   // a func that returns the number of calls to the fake pdp
}

// newMocker return *pdpMocker
// If bulk is true, it is expected that the http.Request is for the bulk API, otherwise it is for the authz API
// Important notes:
// - when initially invoked, this mocker returns the first response in mockedData.responses
// - subsequent call will return the next response in mockedData.responses.
// - the last response in mockedData.responses is returned when there are more calls than the number of responses
func newMocker(t *testing.T, option mockedData) *PDPMocker {

	if len(option.responders) < 1 {
		t.Errorf("require at least 1 response")
	}

	i := 0
	pdpCallCount := 0

	mocker := func(req *http.Request) (*http.Response, error) {
		suf := "v2/authz"
		msg := "Expected v2/authz but found " + req.URL.String()
		if option.bulk {
			suf = suf + "/bulk"
			msg = "Expected v2/authz/bulk but found " + req.URL.String()
		}
		require.True(t, strings.HasSuffix(req.URL.String(), suf), msg)
		header := http.Header{}
		header.Set("Content-Type", "application/json")

		responder := option.responders[i]

		i = i + 1
		if i >= len(option.responders) {
			i = len(option.responders) - 1
		}

		pdpCallCount = pdpCallCount + 1

		res := responder(req)
		if res.error != nil {
			return nil, res.error
		}

		for k, v := range res.header {
			header.Add(k, v)
		}

		return &http.Response{
			StatusCode: res.statusCode,
			Body:       ioutil.NopCloser(bytes.NewBufferString(res.json)),
			Header:     header,
		}, nil
	}

	pdpCounter := func() int {
		return pdpCallCount
	}

	return &PDPMocker{
		NewHTTPMocker(mocker),
		pdpCounter,
	}
}
