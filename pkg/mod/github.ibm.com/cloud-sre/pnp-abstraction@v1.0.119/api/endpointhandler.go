package api

import (
	"github.ibm.com/cloud-sre/pnp-abstraction/monitoring"
	"net/http"
)

// WrappedEndpointHandler - instance data
type WrappedEndpointHandler struct {
	handler      func(resp http.ResponseWriter, req *http.Request)
	validMethods map[string]string
}

// NewWrappedEndpointHandler - constructor to create a wrapped endpoint handler
func NewWrappedEndpointHandler(handler func(resp http.ResponseWriter, req *http.Request), validMethods map[string]string) *WrappedEndpointHandler {
	wrappedEndpointHandler := &WrappedEndpointHandler{
		handler:      handler,
		validMethods: validMethods,
	}
	return wrappedEndpointHandler
}

func (wrappedEndpointHandler WrappedEndpointHandler) wrappedHandler(resp http.ResponseWriter, req *http.Request) {

	resp = defaultHeaders(resp)

	monitoring.AddKubeAttributes(resp)

	// Verify that the method is supported:
	_, ok := wrappedEndpointHandler.validMethods[req.Method]
	if !ok {
		http.Error(resp, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	// Call real handler:
	wrappedEndpointHandler.handler(resp, req)
}

// defaultHeaders sets the required HTTP headers to comply with Zap scan
// rules.
func defaultHeaders(res http.ResponseWriter) http.ResponseWriter {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	res.Header().Set("Pragma", "no-cache")
	res.Header().Set("X-Content-Type-Options", "nosniff")
	res.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	return res
}
