package utils

import "net/http"

const CONTENT_TYPE_JSON = "application/json"

// Enables CORS for swagger
func EnableCORS(resp http.ResponseWriter, setJsonContentType bool) {
	//resp.Header()["Access-Control-Allow-Origin"] = []string{HOST_URI+":9090"}
	//resp.Header()["Access-Control-Allow-Origin"] = []string{"http://sretools.rtp.raleigh.ibm.com"}
	resp.Header()["Access-Control-Allow-Origin"] = []string{"*"}
	resp.Header()["Access-Control-Allow-Headers"] = []string{"Origin", "X-Requested-With", "Content-Type", "Accept"}

	if setJsonContentType {
		resp.Header()["Content-Type"] = []string{CONTENT_TYPE_JSON}
	}
}

// This method will either handle a CORS OPTIONS method, or enable
// CORS on the response so that
// the Swagger UI can work. It will return TRUE if the request was
// handled because it was an OPTIONS method, or it will return
// false if it just enabled CORS on the response.
// Set 'setJsonContentType' if you would also like the content type
// automatically set to be application/json
func HandleCORS(resp http.ResponseWriter, req *http.Request, setJsonContentType bool) bool {

	options := req.Method == http.MethodOptions

	if options {
		EnableCORS(resp, false)
		resp.Header()["Content-Type"] = []string{"text/plain"}
		resp.Header()["Access-Control-Allow-Methods"] = []string{http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodDelete, http.MethodPatch}
		resp.WriteHeader(http.StatusOK)
	} else {
		EnableCORS(resp, setJsonContentType)
	}

	return options
}
