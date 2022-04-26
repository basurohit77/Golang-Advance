package common

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	h "github.ibm.com/cloud-sre/commontools/http"
)

type token struct {
	AccessToken            string `json:"access_token"`
	RefreshToken           string `json:"refresh_token"`
	ImsUserID              int    `json:"ims_user_id"`
	TokenType              string `json:"token_type"`
	ExpiresIn              int    `json:"expires_in"`
	Expiration             int    `json:"expiration"`
	RefreshTokenExpiration int    `json:"refresh_token_expiration"`
	Scope                  string `json:"scope"`
}

type groupData struct {
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	TotalCount int `json:"total_count"`
	First      struct {
		Href string `json:"href"`
	} `json:"first"`
	Last struct {
		Href string `json:"href"`
	} `json:"last"`
	Groups []Group `json:"groups"`
}

type Group struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	CreatedAt        string `json:"created_at"`
	CreatedByID      string `json:"created_by_id"`
	LastModifiedAt   string `json:"last_modified_at"`
	LastModifiedByID string `json:"last_modified_by_id"`
	Href             string `json:"href"`
}

var iam_server string

func init() {
	iam_server = os.Getenv("IAM_URL")
	if iam_server == "" {
		iam_server = "https://iam.cloud.ibm.com"
	}
}

func VerifyUser(token string, accountID string, allowedAccessGroups string) (bool, error) {

	bypass := os.Getenv("APIKEY_BYPASS")
	if bypass == "true" {
		return true, nil
	}

	//check token is apikey or accesstoken
	if token == "" {
		return false, fmt.Errorf("apikey or access token is required")
	}
	accessToken := token
	var err error
	if !strings.HasPrefix(token, "Bearer") { //if token is apikey, exchange it to accesstoken
		accessToken, err = apiKeyToToken(token)
		if err != nil {
			return false, err
		}
	}
	//read allowed access groups
	if strings.TrimSpace(allowedAccessGroups) == "" {
		return false, fmt.Errorf("can not read ALLOWED_ACCESS_GROUPS from environment variables")
	}

	//get user access groups
	groups, err := getUserAccessGroups(accessToken, accountID)
	if err != nil {
		return false, err
	}
	//check whether the allowed access groups are in user access groups or not
	acls := strings.Split(allowedAccessGroups, ",")
	for _, v := range acls {
		v = strings.TrimSpace(v)
		valid := checkGroups(v, groups)
		if valid {
			return valid, nil
		} else {
			continue
		}
	}
	return false, nil
}

func apiKeyToToken(apikey string) (string, error) {
	const funcName = "apiKeyToToken"
	var apiKeyToTokenURL = iam_server + "/identity/token"

	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	headers["Accept"] = "application/json"

	body := url.Values{}
	body.Set("grant_type", "urn:ibm:params:oauth:grant-type:apikey")
	body.Set("response_type", "cloud_iam")
	body.Set("apikey", apikey)

	md := &h.Metadata{
		URL:     apiKeyToTokenURL,
		Headers: headers,
		Method:  http.MethodPost,
		Body:    body.Encode(),
		Cookies: nil,
		Timeout: 0,
	}
	result := md.FireHTTPRequest()

	if result.Err != nil {
		log.Println(funcName, " exchange access token from apikey failed: ", result.Err)
		return "", result.Err
	}
	var t token
	err := json.Unmarshal(result.Data, &t)
	if err != nil {
		log.Println(funcName, " unmarshal data failed: ", err)
		return "", err
	}
	return t.AccessToken, nil
}

func getUserAccessGroups(accessToken string, accountID string) (*groupData, error) {
	const funcName = "getUserAccessGroups"
	if accountID == "" {
		errMsg := funcName + "can not find ACCOUNT_ID from environment variable"
		log.Println(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	var groupURL = iam_server + "/v2/groups?account_id="
	url := groupURL + accountID

	headers := make(map[string]string)
	if !strings.HasPrefix(accessToken, "Bearer") {
		accessToken = "Bearer " + accessToken
	}
	headers["Authorization"] = accessToken
	headers["Content-Type"] = "application/json"

	md := &h.Metadata{
		URL:     url,
		Headers: headers,
		Method:  http.MethodGet,
		Body:    nil,
		Cookies: nil,
		Timeout: 0,
	}
	result := md.FireHTTPRequest()
	if result.Err != nil {
		log.Println(funcName, " get user access groups failed, ", result.Err)
		return nil, result.Err
	}

	var groups *groupData
	err := json.Unmarshal(result.Data, &groups)
	if err != nil {
		log.Println(funcName, " unmarshal groups data failed", err)
		return nil, err
	}

	return groups, nil
}

func checkGroups(allowGroup string, gd *groupData) bool {
	groups := gd.Groups
	for _, v := range groups {
		name := v.ID
		if name == allowGroup {
			return true
		}
	}
	return false
}
