package snowauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/oss-secrets/secret"

	"github.ibm.com/cloud-sre/iam-authorize/monitoring"
	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	constants "github.ibm.com/cloud-sre/pnp-abstraction/constants"
)

// SnowAuth is the base struct that implements all methods in this lib
type SnowAuth struct {
	SNowBaseURL string
	SnowToken   string
	Cache       *cache
}

// snowAuthRequest is the SNow auth API request struct
type snowAuthRequest struct {
	Email string `json:"user_name"`
	CRN   string `json:"crn"`
}

// snowAuthResp is the SNow auth API response struct used for both error and successful responses
type snowAuthResp struct {
	snowAuthRespResult
	snowAuthRespError
}

// snowAuthResult is the SNow auth API successful response struct
type snowAuthRespResult struct {
	Result []result `json:"result,omitempty"`
}

type result struct {
	UserName    string     `json:"user_name,omitempty"`
	CRN         string     `json:"crn,omitempty"`
	UserType    string     `json:"userType,omitempty"`
	ServiceType string     `json:"serviceType,omitempty"`
	Authorized  authorized `json:"authorized,omitempty"`
}

type authorized struct {
	Valid   bool   `json:"valid,omitempty"`
	Message string `json:"message,omitempty"`
}

// snowAuthRespError is the SNow auth API error response struct
type snowAuthRespError struct {
	Err struct {
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`
}

// SvcsAuth holds a service as string and the authorization as a bool
type SvcsAuth map[string]bool

const (
	snowAuthAPI = "/api/ibmwc/v1/gaas/userAuthorization"

	// errors that can be returned to caller
	errNoCRN   = "no crns provided"
	errNoEmail = "no email provided"
)

var (
	// ByPassFlag tells whether the resource authorization should be skipped or not
	ByPassFlag = false
)

// NewSNowAuth sets up the base config and returns a struct with methods to check SNow authorization
// snowBaseURL should only contain the protocol and the SNow instance URL, without the path
func NewSNowAuth(snowBaseURL, snowToken string) *SnowAuth {
	const fct = "NewSNowAuth: "

	if snowBaseURL == "" || snowToken == "" {
		// log.Panic(fct + "no ServiceNow URL or token provided") // activate panic when snow auth is required
		log.Println(fct + "no ServiceNow URL or token provided")
	}

	snauth := &SnowAuth{
		SNowBaseURL: snowBaseURL,
		SnowToken:   snowToken,
		Cache:       newSNAuthCache(),
	}

	if flag := secret.Get("SNOW_BYPASS_FLAG"); flag == "true" {
		log.Println(fct + "BYPASS FLAG is active. No resource authorization is performed under this mode.")
		ByPassFlag = true
	}

	return snauth
}

// SNowAuthorization checks authorization in ServiceNow whether UserID(email) has access to each service
// a map of CRNs and a bool showing whether the user has access to the CRN is returned, together with an error
func (snauth *SnowAuth) SNowAuthorization(ctx context.Context, email string, crns []string) (SvcsAuth, error) {

	const fct = "SNowAuthorization: "

	if email == "" {
		return make(SvcsAuth), errors.New(errNoEmail)
	}
	if len(crns) == 0 {
		return make(SvcsAuth), errors.New(errNoCRN)
	}

	// count number of transactions in this lib
	monitoring.SetTag(ctx, "SNAuthTransaction", 1)

	// map to hold services and the access to the service; true=authorized, false=unauthorized
	var svcsAuth = make(SvcsAuth)

	var batchCRNs []string

	var checkAuth = make(map[string]bool)

	// remove CRN duplicates
	for _, crn := range crns {
		if !checkAuth[crn] {
			checkAuth[crn] = true
		}
	}

	for crn := range checkAuth {

		log.Printf("Checking ServiceNow authorization for email=%s; crn=%s\n", email, crn)

		// check cache
		auth, found := snauth.Cache.search(email, crn)
		if found {
			svcsAuth[crn] = auth
			continue
		}

		if ByPassFlag { // allow anyone access to any service when bypass is true

			log.Println(fct + "WARNING! BYPASS FLAG is true. No resource authorization is being performed.")
			svcsAuth[crn] = true
			snauth.Cache.add(email, crn, true)

		} else if api.IsPublicCRN(crn) { // allow anyone access to any public bmx service

			log.Println("Access allowed. CRN starts with " + constants.GenericCRNMaskShort + ", we allow anyone access to these CRNs.")
			svcsAuth[crn] = true
			snauth.Cache.add(email, crn, true)

		} else { // check access using ServiceNow Auth API

			// append all CRNs to make a single batch request to SNow
			batchCRNs = append(batchCRNs, crn)

		}
	}

	return snauth.snowBatchRequest(ctx, email, batchCRNs, svcsAuth)
}

func (snauth *SnowAuth) snowBatchRequest(ctx context.Context, email string, crns []string, svcsAuth SvcsAuth) (SvcsAuth, error) {

	if len(crns) > 0 { // only call SNow when there are CRNs to check authorization
		res, err := snauth.callSNowAuthAPI(email, crns)
		if err != nil {
			return svcsAuth, err
		}

		snowSvcsAuth, err := snauth.getAuthorization(res)
		if err != nil {
			return svcsAuth, err
		}

		for crn, isAuthorized := range snowSvcsAuth {
			svcsAuth[crn] = isAuthorized

			// add to cache the authorization result, either false or true
			snauth.Cache.add(email, crn, isAuthorized)

			if !isAuthorized {
				// send metrics to monitoring to show who has been denied access on and what service/resource
				monitoring.SetTagsKV(ctx, "SNAuthEmail", email,
					"SNAuthAuthorized", isAuthorized,
					"SNAuthCRN", crn)
				if svc, _ := api.GetServiceFromCRN(crn); svc != "" {
					monitoring.SetTag(ctx, "SNAuthServiceName", svc)
				}
			}
		}
	}

	return svcsAuth, nil
}

// callSNowAuthAPI call ServiceNow auth API to check whether an user ID has access to a service
func (snauth *SnowAuth) callSNowAuthAPI(email string, crns []string) (*http.Response, error) {

	var bulkReq []*snowAuthRequest

	for _, crn := range crns {
		bulkReq = append(bulkReq, &snowAuthRequest{Email: email, CRN: crn})
	}

	reqBody, err := json.Marshal(bulkReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, snauth.SNowBaseURL+snowAuthAPI, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	authToken := ""
	if strings.HasPrefix(snauth.SnowToken, "Bearer") {
		authToken = snauth.SnowToken
	} else {
		authToken = "Bearer " + snauth.SnowToken
	}

	req.Header.Set("Authorization", authToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	log.Println("Calling ServiceNow auth API. URL=", req.URL)
	log.Println("Request body=", string(reqBody))

	tTraceStart := time.Now()

	res, status, err := api.MakeHTTPCallWithRetry(client, req, http.StatusOK)
	if err != nil {
		message := fmt.Sprintf("ServiceNow returned unexpected status. Status=%d;", status)
		return res, errors.New(message)
	}
	tTraceEnd := time.Now()
	log.Printf("Request time: %s", tTraceEnd.Sub(tTraceStart).String())

	return res, nil

}

func (snauth *SnowAuth) getAuthorization(res *http.Response) (SvcsAuth, error) {
	defer closeBody(res.Body.Close)

	snResp, err := snauth.decodeResponse(res.Body)
	if err != nil {
		return make(SvcsAuth), err
	}

	var svcsAuth = make(SvcsAuth)

	for _, result := range snResp.Result {
		svcsAuth[result.CRN] = result.Authorized.Valid
		snauth.Cache.addToUsers(result)
		snauth.Cache.addToServices(result)

		// if a msg(error) is returned for a specific CRN that is invalid then print it
		if !result.Authorized.Valid && result.Authorized.Message != "" {
			log.Println(result.Authorized.Message)
		}
	}

	return svcsAuth, nil
}

// GetUserType returns the user type
func (snauth *SnowAuth) GetUserType(email string, crn []string) string {

	if userType := snauth.Cache.getUserType(email); userType != "" {
		return userType
	}

	// go out to SNow to check for this email as it was not found in cache
	res, err := snauth.callSNowAuthAPI(email, crn)
	if err != nil {
		log.Println(err)
		return ""
	}

	_, err = snauth.getAuthorization(res)
	if err != nil {
		log.Println(err)
		return ""
	}

	// at this point we should be able to find the entry
	if userType := snauth.Cache.getUserType(email); userType != "" {
		return userType
	}

	return ""

}

// GetServiceType gets the service type
func (snauth *SnowAuth) GetServiceType(service, email string) string {

	if svcType := snauth.Cache.getServiceType(service); svcType != "" {
		return svcType
	}

	// go out to SNow to check for this service as it was not found in cache
	res, err := snauth.callSNowAuthAPI(email, []string{fmt.Sprintf("crn:v1:::%s:::::", service)})
	if err != nil {
		log.Println(err)
		return ""
	}

	_, err = snauth.getAuthorization(res)
	if err != nil {
		log.Println(err)
		return ""
	}

	// at this point we should be able to find the entry
	if svcType := snauth.Cache.getServiceType(service); svcType != "" {
		return svcType
	}

	return ""

}

// decodeResponse decodes an HTTP response body to snowAuthResp type
func (snauth *SnowAuth) decodeResponse(body io.ReadCloser) (*snowAuthResp, error) {

	responseBody, err := ioutil.ReadAll(body)
	if err != nil {
		// failed to read SN response body, internal error
		message := "Failed to read ServiceNow response body. Error= " + err.Error()
		return &snowAuthResp{}, errors.New(message)
	}

	var snResp *snowAuthResp

	err = json.Unmarshal(responseBody, &snResp)
	if err != nil {
		// failed to unmarshal resp body, internal error
		message := "Failed to unmarshal ServiceNow response body: " + string(responseBody) + ". Error= " + err.Error()
		return &snowAuthResp{}, errors.New(message)
	}

	log.Printf("serviceNowResponse= %s", responseBody)

	return snResp, nil
}

// closeBody closes HTTP response body and if unsuccessful print out error
func closeBody(f func() error) {
	if err := f(); err != nil {
		log.Println(err)
	}
}
