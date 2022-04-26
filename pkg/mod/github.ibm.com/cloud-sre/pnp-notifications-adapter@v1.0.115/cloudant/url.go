package cloudant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

const (
	//servicesIDsURL = "https://%s.cloudant.com/dashboard.html#database/notification-control/Category_SERVICES"
	servicesIDsURL = "https://%s.cloudant.com/notification-control/Category_SERVICES"

	//runtimesIDsURL = "https://%s.cloudant.com/dashboard.html#database/notification-control/Category_RUNTIMES"
	runtimesIDsURL = "https://%s.cloudant.com/notification-control/Category_RUNTIMES"

	//platformIDsURL = "https://%s.cloudant.com/dashboard.html#database/notification-control/Category_PLATFORM"
	platformIDsURL = "https://%s.cloudant.com/notification-control/Category_PLATFORM"

	//notificationsURL = "https://%s.cloudant.com/dashboard.html#database/notifications/86377f50a5a47df632923f35f125623d"
	notificationsURL = "https://%s.cloudant.com/notifications/_all_docs?include_docs=true"
)

// cloudantData provides information such as cloudant credentials.  This will need to get updated with Vault.
type cloudantData struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

var cloudantCredentials *cloudantData

// GetCloudantID returns the ID to use with Cloudant requests
func GetCloudantID() (id string, err error) {

	creds, err := pullCloudantCredentials()
	if err != nil {
		return "", err
	}

	return creds.ID, err
}

// GetCloudantPassword returns the password to use with Cloudant requests
func GetCloudantPassword() (pw string, err error) {

	creds, err := pullCloudantCredentials()
	if err != nil {
		return "", err
	}

	return creds.Password, err
}

// GetServicesIDsURL is the URL to retrieve the IDs related to Services
func GetServicesIDsURL() (url string, err error) {

	id, err := GetCloudantID()
	if err != nil {
		return "", err
	}

	// First check to see if a URL override is in the environment.
	url = os.Getenv(ServicesURLEnv)
	if url != "" {
		return url, err
	}

	return fmt.Sprintf(servicesIDsURL, id), nil
}

// GetRuntimesIDsURL is the URL to retrieve the IDs related to runtimes
func GetRuntimesIDsURL() (url string, err error) {

	id, err := GetCloudantID()
	if err != nil {
		return "", err
	}

	// First check to see if a URL override is in the environment.
	url = os.Getenv(RuntimesURLEnv)
	if url != "" {
		return url, err
	}

	return fmt.Sprintf(runtimesIDsURL, id), nil
}

// GetPlatformIDsURL is the URL to retrieve the IDs related to platform
func GetPlatformIDsURL() (url string, err error) {

	id, err := GetCloudantID()
	if err != nil {
		return "", err
	}

	// First check to see if a URL override is in the environment.
	url = os.Getenv(PlatformURLEnv)
	if url != "" {
		return url, err
	}

	return fmt.Sprintf(platformIDsURL, id), nil
}

// GetNotificationsURL is the URL to retrieve the IDs related to notifications
func GetNotificationsURL() (url string, err error) {

	id, err := GetCloudantID()
	if err != nil {
		return "", err
	}

	// First check to see if a URL override is in the environment.
	url = os.Getenv(NotificationsURLEnv)
	if url != "" {
		return url, err
	}

	return fmt.Sprintf(notificationsURL, id), nil
}

func pullCloudantCredentials() (data *cloudantData, err error) {

	// First look to see if an env variable is set
	if cloudantCredentials == nil {
		id := os.Getenv(AccountID)
		pw := os.Getenv(AccountPW)

		if id != "" && pw != "" {
			cloudantCredentials = new(cloudantData)
			cloudantCredentials.ID = id
			cloudantCredentials.Password = pw
		}
	}

	// Second, if not in env, see if there is a file in the home directory
	if cloudantCredentials == nil {
		newInfo, err := cloudantCredentialsFromUserHome()

		if err != nil {
			return nil, err
		}

		cloudantCredentials = newInfo
	}
	return cloudantCredentials, err
}

func cloudantCredentialsFromUserHome() (data *cloudantData, err error) {
	METHOD := "cloudantCredentialsFromUserHome"

	usr, err := user.Current()
	if err != nil {
		log.Println(fmt.Sprintf("ERROR (%s): Cannot get curent user - %s", METHOD, err.Error()))
		return nil, err
	}

	fileData, err := ioutil.ReadFile(filepath.Join(usr.HomeDir, "/.keys/cloudant.keys"))
	if err != nil {
		log.Println(fmt.Sprintf("ERROR (%s): Cannot read file in user dir - %s", METHOD, err.Error()))
		return nil, err
	}

	newInfo := new(cloudantData)
	if err := json.NewDecoder(bytes.NewReader(fileData)).Decode(newInfo); err != nil {
		log.Println(fmt.Sprintf("ERROR (%s): Cannot decode the credential file - %s", METHOD, err.Error()))
		return nil, err
	}

	return newInfo, err

}
