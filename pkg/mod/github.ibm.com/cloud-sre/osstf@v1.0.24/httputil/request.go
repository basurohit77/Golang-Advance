package httputil

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"github.ibm.com/cloud-sre/osstf/lg"
	"net/http"
	"time"
	"strconv"
)

//###########################################################################
//                  Global Variables
//###########################################################################

//###########################################################################
//                     Structures
//###########################################################################
type RequestContext struct {
	ClientName     string            // string to log this client
	Url            string            // Full url to what is being requested
	Headers        map[string]string // Additional headers to add to the request
	LastStatusCode int               // The status call from the last REST request
	Timeout        int               // Number of seconds that a call will wait before timing out. Default value 0 means no timeout
}

func (rc *RequestContext) CreateClient() (client *http.Client) {
	client = &http.Client{
		Timeout: time.Duration(rc.Timeout) * time.Second,
	}
	return client
}

//// Takes a URL base and an URI path and creates one full URL
//func MakeFullUrl(url_base string, uri string) (*url.URL, error) {
//	u, err := url.Parse(uri)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	base, err := url.Parse(url_base)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	return base.ResolveReference(u), nil
//}

//--------------------------------------------------------------
// Send a reader of json data
func (rc *RequestContext) PostReader(data io.Reader, contentType string) (responseBody []byte, err error) {
	client := rc.CreateClient()
	req, err := http.NewRequest(http.MethodPost, rc.Url, data)

	if len(contentType) > 0 {
		req.Header.Add("Content-Type", contentType)
	}

	if err != nil {
		return nil, err
	}

	return rc.doRequest(client, req)
}

//// Send a PATCH reader of json data
//func (rc *RequestContext) PatchReader(data io.Reader, contentType string) (responseBody []byte, err error) {
//
//	client := &http.Client{}
//	req, err := http.NewRequest(http.MethodPatch, rc.Url, data)
//
//	if len(contentType) > 0 {
//		req.Header.Add("Content-Type", contentType)
//	}
//
//	if err != nil {
//		return nil, err
//	}
//
//	return rc.doRequest(client, req)
//}

//// Put a reader of json data
//func (rc *RequestContext) PutReader(data io.Reader, contentType string) (responseBody []byte, err error) {
//
//	client := &http.Client{}
//	req, err := http.NewRequest(http.MethodPut, rc.Url, data)
//
//	if len(contentType) > 0 {
//		req.Header.Add("Content-Type", contentType)
//	}
//
//	if err != nil {
//		return nil, err
//	}
//
//	return rc.doRequest(client, req)
//}

//--------------------------------------------------------------
func (rc *RequestContext) Get() (responseBody []byte, err error) {

	client := rc.CreateClient()
	req, err := http.NewRequest(http.MethodGet, rc.Url, nil)

	if err != nil {
		return nil, err
	}

	return rc.doRequest(client, req)
}

func (rc *RequestContext) Delete() (responseBody []byte, err error) {

	client := rc.CreateClient()
	req, err := http.NewRequest(http.MethodDelete, rc.Url, nil)

	if err != nil {
		return nil, err
	}

	return rc.doRequest(client, req)
}

func (rc *RequestContext) doRequest(client *http.Client, req *http.Request) (responseBody []byte, err error) {

	for header, value := range rc.Headers {
		req.Header.Add(header, value)
	}

	num := lg.SendRequest(rc.ClientName, req)

	// 2017-01-06: Found the sockets where remaining open for a while which could cause the system to
	// run out of file descriptors.  Adding this close connection header to make the server close out
	// the connection since REST interfaces don't need to keep them open
	req.Header.Set("Connection", "close")

	result := make([]byte, 0)
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err == nil {
		result, err = ioutil.ReadAll(resp.Body)

		rc.LastStatusCode = resp.StatusCode
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			err = errors.New(fmt.Sprintf("%d", resp.StatusCode))
		}
	}
	
	lg.ReceiveResponse(num, rc.ClientName, resp, result)

	return result, err
}

func SetDefaultResponseExpires(resp http.ResponseWriter, minutes int) {
	// set to expire one hour from now
	resp.Header().Set("Expires", time.Now().Add(time.Minute * time.Duration(minutes)).Format(http.TimeFormat))
}

// This is a helper function specifically for the error returned after calling one of the requests in this package.
// It returns true for successful HTTP request even if there was an error return from  HTTP, as long as the HTTP status code was 2xx;
// It returns false for HTTP status code other then 200-level.
func IsHttpStatusSuccessful(rc *RequestContext, err error) bool {
	// See code in "doRequest(client *http.Client, req *http.Request) (responseBody []byte, err error)"
	//	resp, err := client.Do(req)
	//	if err == nil {
	//		if resp.StatusCode < 200 || resp.StatusCode > 299 {
	//			err = errors.New(fmt.Sprintf("%d", resp.StatusCode))
	//		}
	//	}
	// According to GO documentation https://golang.org/pkg/net/http/#Client.Do
	// 	An error is returned if caused by client policy (such as CheckRedirect), 
	//	or failure to speak HTTP (such as a network connectivity problem). 
	//	A non-2xx status code doesn't cause an error.
	// Therefore if the error message can be parsed into an integer less then 200 or greater than 299
	//	then that's the HTTP status code for an unsuccessful request
	if err == nil {
		// no error from HTTP from doRequest
		return true
	} else {
		// there was an error from doRequest
		i, e := strconv.Atoi(err.Error())
		if e == nil {
			// return the parsed HTTP status code
			return i > 199 && i < 300
		} else {
			// cannot parse the error message into an integer. The error was from HTTP, not made up by doRequest method
			return rc.LastStatusCode > 199 && rc.LastStatusCode < 300
		}
	}
}
