// IBM Confidential OCO Source Materials
// (C) Copyright and Licensed by IBM Corp. 2019
//
// The source code for this program is not published or otherwise
// divested of its trade secrets  irrespective of what has
// been deposited with the U.S. Copyright Office.

package notification

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

const SLACK_TOKEN_KEY string = "slack"

func AuthenticationTest() (bool, error) {
	slackToken, err := rest.GetKey(SLACK_TOKEN_KEY)
	if err != nil {
		return false, err
	}
	endpoint := "https://slack.com/api/auth.test"

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/x-www-form-urlencoded"
	headers["Authorization"] = fmt.Sprintf("Bearer %s", slackToken)

	data, err := fireHTTPRequest(endpoint, headers, http.MethodPost, nil)
	if err != nil {
		return false, err
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(data, &response)
	if err != nil {
		return false, err
	}

	isOkey := response["ok"].(bool)
	if !isOkey {
		return false, fmt.Errorf("Authentication test failed with error: %v", response["error"])
	}

	return true, nil
}

// GetSlackUserByEmail - Get slack user by email
func getSlackUserByEmail(email string) (map[string]interface{}, error) {
	slackToken, err := rest.GetKey(SLACK_TOKEN_KEY)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("https://slack.com/api/users.lookupByEmail?email=%s", email)

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/x-www-form-urlencoded"
	headers["Authorization"] = fmt.Sprintf("Bearer %s", slackToken)

	data, err := fireHTTPRequest(endpoint, headers, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	isOkey := response["ok"].(bool)
	if !isOkey {
		return nil, fmt.Errorf("Unable to get user info because of %v", response["error"])
	}
	user := response["user"].(map[string]interface{})

	return user, err
}

func PostSimpleMessage(channel, title, body, color, email string) error {
	attachments := make([]map[string]interface{}, 1)
	attachment := make(map[string]interface{})
	attachment["pretext"] = title
	attachment["text"] = body
	attachment["color"] = color
	attachments[0] = attachment
	return PostMessage(channel, attachments, email)
}

// SendSlackNotificationToExecutor - Post a message to a slack channel, and notify user
func PostMessage(channel string, attachments []map[string]interface{}, email string) error {
	debug.Info("sending " + attachments[0]["pretext"].(string) + " to " + channel + ".")
	if channel == "" {
		errorMsg := "channel is required params."
		debug.PrintError(errorMsg)
		return errors.New(errorMsg)
	}
	slackToken, err := rest.GetKey(SLACK_TOKEN_KEY)
	if err != nil {
		debug.PrintError(err.Error())
		return err
	}
	endpoint := "https://slack.com/api/chat.postMessage"
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"
	headers["Authorization"] = fmt.Sprintf("Bearer %s", slackToken)

	// Generate text
	pretext := attachments[0]["pretext"].(string)
	if email != "" {
		debug.Info("SendSlackNotification email: %s. ", email)
		userInfo, err := getSlackUserByEmail(email)
		if err != nil {
			debug.PrintError(err.Error())
		} else {
			userID := userInfo["id"].(string)
			debug.Info(fmt.Sprintf("Got user ID: %s", userID))
			pretext = fmt.Sprintf("<@%s>, %s", userID, pretext)
			attachments[0]["pretext"] = pretext
		}
	}
	body := make(map[string]interface{})
	body["channel"] = channel
	body["attachments"] = attachments

	data, err := fireHTTPRequest(endpoint, headers, http.MethodPost, body)
	if err != nil {
		debug.PrintError(err.Error())
		return err
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(data, &response)
	if err != nil {
		debug.PrintError(err.Error())
		return err
	}

	isOkey := response["ok"].(bool)
	if !isOkey {
		err = fmt.Errorf("Unable to send slack message to user %v because of %v", email, response["error"])
		debug.PrintError(err.Error())
		return err
	}
	return nil
}

//FireHTTPRequest http wrapper
func fireHTTPRequest(url string, headers map[string]string, method string, body interface{}) ([]byte, error) {

	if url == "" {
		return nil, fmt.Errorf("url is required")
	}
	if method == "" {
		return nil, fmt.Errorf("method is required")
	}

	transport := &http.Transport{DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}} // #nosec

	var bodyReader io.Reader
	if body != nil {
		tp := reflect.TypeOf(body)
		tps := tp.String()
		if tps == "string" {
			bodyReader = strings.NewReader(body.(string))
		} else {
			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			bodyReader = strings.NewReader(string(bodyBytes))
		}
	}
	req, _ := http.NewRequest(strings.ToUpper(method), url, bodyReader)

	for k, v := range headers {
		if k == "BasicAuth" {
			usernamePwdPair := strings.Split(v, "::")
			if len(usernamePwdPair) == 2 {
				u := usernamePwdPair[0]
				p := usernamePwdPair[1]
				req.SetBasicAuth(u, p)
			}
			continue
		}
		req.Header.Set(k, v)
	}

	res, err := transport.RoundTrip(req)

	if err != nil {
		return nil, err
	}

	if transport != nil {
		defer transport.CloseIdleConnections()
	}
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}

	data, err := ioutil.ReadAll(res.Body)

	return data, err
}
