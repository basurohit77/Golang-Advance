package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

// InfoHandler - info handler data
type InfoHandler struct {
	ClientID             string
	CategoryID           string
	Description          string
	KongURL              string
	APIRequestPathPrefix string
	Hrefs                []InfoHandlerHref
}

// InfoHandlerHref - info handler href data
type InfoHandlerHref struct {
	ID      string
	SubPath string
}

// HrefPointer is the struct used in APIInfo to reference other paths
type HrefPointer struct {
	Href string `json:"href"`
}

// ServeInfo - handle GET api/info requests
func (infoHandler InfoHandler) ServeInfo(resp http.ResponseWriter, req *http.Request) {
	const FCT = "ServeInfo: "

	// intercept panics: print error and stacktrace
	defer HandlePanics(resp)

	responseBody := make(map[string]interface{})
	responseBody["clientId"] = infoHandler.ClientID
	responseBody["description"] = infoHandler.Description
	responseBody["categories"] = []string{infoHandler.CategoryID}

	for _, href := range infoHandler.Hrefs {
		responseBody[href.ID] = HrefPointer{Href: infoHandler.KongURL + "/" + infoHandler.APIRequestPathPrefix + href.SubPath}
	}

	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(responseBody); err != nil {
		// Encoding of json response failed, this is an internal / programming error
		log.Println(FCT + "Error: failed to encode api info")
		resp.Header().Set("X-ZAP", "OK")
		resp.WriteHeader(http.StatusInternalServerError)
	} else {
		// return successful response
		resp.Header().Set("Content-Type", "application/json")
		resp.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		resp.Header().Set("Pragma", "no-cache")
		resp.Header().Set("X-Content-Type-Options", "nosniff")
		resp.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		resp.WriteHeader(http.StatusOK)
		if _, err := resp.Write(buffer.Bytes()); err != nil {
			log.Println(err)
		}
	}
}
