package iam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.ibm.com/IAM/pep/v3"
)

/*
Wrapper for the IAM Pep library. Allows IAM health check to be unit tested.

Example usage:

import (
    "log"
    "time"
    "github.ibm.com/cloud-sre/iam-authorize/iam"
    "github.ibm.com/IAM/pep/v3"
)

...
apiKey := "1234567890"

config := &pep.Config{
	Environment: pep.Staging,
	APIKey:      apiKey,
	LogLevel:    pep.LevelError,
	AuthzRetry:  true,
	PDPTimeout:  time.Duration(30) * time.Second, // not required to add; default is 15s
}

pep.Configure(config)

pepWrapper := iam.NewPepWrapper()
token, pepErr := pepWrapper.GetToken(apiKey)
if pepErr != nil {
  log.Println(pepErr.Error())
  return
}

isTokenValid, pepErr := pepWrapper.GetClaims(token)
if pepErr != nil {
  log.Println(pepErr.Error())
  return
}
*/

// PepWrapper is the interface implemented by the wrapper
type PepWrapper interface {
	GetToken(HealthCheckConfig) (string, error)
	GetClaims(token string) (bool, error)
}

// NewPepWrapper constructs a instance of the wrapper.
func NewPepWrapper() PepWrapper {
	return &pepWrapper{}
}

// pepWrapper is the implementation of the Pep wrapper interface. This pepWrapper implementation uses the IAM Pep
// library.
type pepWrapper struct {
}

type tokenStruct struct {
	AccessToken string `json:"access_token"`
}

// GetServiceToken returns an IAM token for the provided IAM API key.
func (pepWrapper *pepWrapper) GetToken(healthConfig HealthCheckConfig) (string, error) {
	return getToken(healthConfig)
}

// VerifyUserToken validates the provided token.
func (pepWrapper *pepWrapper) GetClaims(token string) (bool, error) {
	_, err := pep.GetClaims(token, false)
	if err != nil {
		return false, err
	}
	return true, err
}

// getToken returns an access token generated with the provided api key, and an error
func getToken(healthConfig HealthCheckConfig) (string, error) {
	postData := fmt.Sprintf("grant_type=urn:ibm:params:oauth:grant-type:apikey&apikey=%s", healthConfig.IAMApiKey)

	url := healthConfig.IAMUrl + "/identity/token"

	rsp, err := retrieveIAMToken(url, postData, healthConfig.IAMTimeoutInSecs)
	if err != nil {
		return "", err
	}

	defer closeBody(rsp.Body.Close)

	jsonBytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	// Decode the json result into a struct
	ts := new(tokenStruct)
	if err := json.NewDecoder(bytes.NewReader(jsonBytes)).Decode(ts); err != nil {
		return "", err
	}

	return ts.AccessToken, nil
}

// retrieveIAMToken calls IAM URL to retrieve an access token
func retrieveIAMToken(url, data string, timeout int) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth("bx", "bx")

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("Request to URL %s failed with status code %d", url, resp.StatusCode)
	}

	return resp, err
}

func closeBody(f func() error) {
	if err := f(); err != nil {
		log.Println(err)
	}
}
