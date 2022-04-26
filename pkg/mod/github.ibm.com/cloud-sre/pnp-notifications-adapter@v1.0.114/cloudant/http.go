package cloudant

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

// GetFromCloudant provides a standardized call method for GET requests to cloudant
func GetFromCloudant(callingFunction, httpMethod, url, user, password string) (resp *http.Response, err error) {
	METHOD := "utils/GetFromCloudant"

	req, err := MakeCloudantRequest(httpMethod, url, user, password, nil)

	if err != nil {
		msg := fmt.Sprintf("ERROR (%s)->(%s): Failed to build cloudant request to query notifications: %s", callingFunction, METHOD, err.Error())
		log.Println(msg)
		return nil, errors.New(msg)
	}

	client := &http.Client{}
	resp, err = client.Do(req)

	if err != nil {
		msg := fmt.Sprintf("ERROR (%s)->(%s): Failed to get data from cloudant: %s", callingFunction, METHOD, err.Error())
		log.Println(msg)
		return nil, errors.New(msg)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := fmt.Sprintf("ERROR (%s)->(%s): Unexpected status code (%s) returned when querying cloudant: %s", callingFunction, METHOD, resp.Status, err.Error())
		log.Println(msg)
		return nil, errors.New(msg)
	}

	return resp, err
}

// MakeCloudantRequest creates an HTTP request for cloudant
func MakeCloudantRequest(method, myurl, id, pw string, body io.Reader) (req *http.Request, err error) {

	req, err = http.NewRequest(method, myurl, body)
	if err == nil {

		req.Header.Add("Content-Type", "application/json")

		req.SetBasicAuth(id, pw)

	}
	return req, err
}
