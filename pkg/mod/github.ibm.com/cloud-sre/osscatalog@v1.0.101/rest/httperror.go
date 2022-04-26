package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.ibm.com/cloud-sre/osscatalog/debug"
)

// HTTPError represents an error returned from a REST call.
// It might be a plain low-level error, or as non-200 status code (together with some error message)
type HTTPError interface {
	error
	IsEntryNotFound() bool
	GetHTTPStatusCode() int
	GetHTTPDetails() string
}

type httpErrorData struct {
	err           error
	innerError    error
	code          int
	status        string
	details       string
	entryNotFound bool
}

// MakeHTTPError checks the supplied error and HTTP response, and creates a HTTPError record if either of them reflects a failure.
// It returns nil if no failure was detected.
func MakeHTTPError(innerError error, response *http.Response, entryNotFound bool, message string, parms ...interface{}) HTTPError {
	if innerError == nil && response == nil && !entryNotFound {
		panic("HTTPError() must specify at least one of innerError, resp or entryNotFound")
	}
	var haveInnerError = "nil"
	var haveResponse = "nil"
	if innerError != nil {
		haveInnerError = "err"
		prefix := fmt.Sprintf(message, parms...)
		result := httpErrorData{}
		result.err = fmt.Errorf("%s: %v", prefix, innerError)
		result.innerError = innerError
		result.entryNotFound = entryNotFound
		debug.Debug(debug.Main, "MakeHTTPError(%s,%s,%t): %v", haveInnerError, haveResponse, entryNotFound, result.err)
		return &result
	}
	if response != nil {
		haveResponse = "response"
		code := response.StatusCode
		if code < 200 || code >= 300 {
			var details string
			buf, readErr := ioutil.ReadAll(response.Body)
			if readErr == nil {
				if len(buf) > 300 {
					details = string(buf[:300]) + "..."
				} else {
					details = string(buf)
				}
			} else {
				details = fmt.Sprintf("***Error reading response body for error message: %v***", readErr)
			}
			prefix := fmt.Sprintf(message, parms...)
			result := httpErrorData{}
			result.err = fmt.Errorf("%s: HTTPError code=%d %s : %s", prefix, code, http.StatusText(code), details)
			result.code = code
			result.status = response.Status
			result.details = details
			result.entryNotFound = entryNotFound
			debug.Debug(debug.Main, "MakeHTTPError(%s,%s,%t): %v", haveInnerError, haveResponse, entryNotFound, result.err)
			return &result
		}
	}
	if entryNotFound {
		result := httpErrorData{}
		result.err = fmt.Errorf(message, parms...)
		result.entryNotFound = entryNotFound
		debug.Debug(debug.Main, "MakeHTTPError(%s,%s,%t): %v", haveInnerError, haveResponse, entryNotFound, result.err)
		return &result
	}
	return nil
}

// Error returns a string representation of this HTTPError, to satisfy the error interface
func (he httpErrorData) Error() string {
	return he.err.Error()
}

// IsEntryNotFound returns true if this HTTPError represents the fact that a given entry was not found during a lookup
// (maybe because a 404 or because of other causes)
func (he httpErrorData) IsEntryNotFound() bool {
	return he.entryNotFound
}

// GetHTTPStatusCode returns the HTTP status code associated with this error, or 0 if this error is not linked to a HTTP status
func (he httpErrorData) GetHTTPStatusCode() int {
	return he.code
}

// GetHTTPDetails returns the HTTP details associated with this error, or empty string if this error is not linked to a HTTP status
func (he httpErrorData) GetHTTPDetails() string {
	return he.details
}

// IsEntryNotFound returns true if this generic error represents the fact that a given entry was not found during a lookup
// (maybe because a 404 or because of other causes)
func IsEntryNotFound(err error) bool {
	if he, ok := err.(HTTPError); ok {
		return he.IsEntryNotFound()
	}
	return false
}

// GetHTTPStatusCode returns the HTTP error code associated with this generic error if it is a HTTP error; 0 otherwise
func GetHTTPStatusCode(err error) int {
	if he, ok := err.(HTTPError); ok {
		return he.GetHTTPStatusCode()
	}
	return 0
}
