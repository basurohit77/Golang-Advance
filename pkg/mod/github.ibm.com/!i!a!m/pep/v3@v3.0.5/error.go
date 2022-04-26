package pep

import (
	"fmt"
)

// APIError is the data structure with relevant information when the error occurs
type APIError struct {
	EndpointURI     string
	RequestHeaders  string
	ResponseHeaders string
	Message         string
	StatusCode      int
	Trace           string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error code: %d calling %v for txid: %v Details: %v", e.StatusCode, e.EndpointURI, e.Trace, e.Message)
}

// InternalError has relevant information related to internal errors.
type InternalError struct {
	Message string
	Trace   string
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("Internal error %v while processing txid: %v", e.Message, e.Trace)
}
