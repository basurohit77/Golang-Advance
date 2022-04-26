package api

import (
	"errors"
	"log"
	"os"

	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/globalcatalog"
)

// SourceConfig provides information about sources such as URLs and credentials
type SourceConfig struct {
	//Cloudant struct {
	//	AccountID        string
	//	Password         string
	//	NotificationsURL string
	//	ServiceNamesURL  string
	//	RuntimeNamesURL  string
	//	PlatformNamesURL string
	//}

	GlobalCatalog struct {
		ResourcesURL string
	}

	OSSCatalog osscatalog.Creds
}

// gavila remove dead code
// GetNotifications will return all notification records
//func GetNotifications(ctx ctxt.Context, creds *SourceConfig) (result *NotificationList, err error) {
//
//	err = SetupOSSCatalogCredentials(creds)
//	if err != nil {
//		return nil, err
//	}
//
//	cloudantRecords, err := cloudant.GetNotifications(ctx, creds.Cloudant.NotificationsURL, creds.Cloudant.AccountID, creds.Cloudant.Password)
//	if err != nil {
//		return nil, err
//	}
//
//	notifications, err := ConvertNotifications(ctx, cloudantRecords, creds)
//
//	return notifications, err
//}

// SetupOSSCatalogCredentials with setup the credentials for the OSS catalog library
// using only the oss credential part of the whole credential structure
func SetupOSSCatalogCredentials(creds *SourceConfig) error {

	if creds != nil && creds.OSSCatalog.CredentialsSet {
		return nil
	}

	// This is a bit of a hack right now.  Here we are loading credentials for use
	// by the OSS catalog by directly calling the load that it uses to load VCAP information.
	if creds == nil || creds.OSSCatalog.Credentials == nil {
		return errors.New("Missing OSSCatalog credentials")
	}

	return osscatalog.SetupOSSCatalogCredentials(&creds.OSSCatalog)

	//	creds.OSSCatalog.CredentialsSet = true
	//	return rest.LoadVCAPServices(creds.OSSCatalog.Credentials)
}

// AddOSSCatalogCredential is a convenience method to add oss credentials to the credential struct
func (creds *SourceConfig) AddOSSCatalogCredential(keyName, label, value string) {
	creds.OSSCatalog.AddOSSCatalogCredential(keyName, label, value)
}

// GetCredentials will retrieve the necessary credentials for notifications
func GetCredentials() (creds *SourceConfig, err error) {

	creds = new(SourceConfig)
	// gavila remove dead code
	//creds.Cloudant.AccountID = os.Getenv(cloudant.AccountID)
	//if creds.Cloudant.AccountID == "" {
	//	return nil, errors.New("Missing cloudant account ID")
	//}
	//
	//creds.Cloudant.Password = os.Getenv(cloudant.AccountPW)
	//if creds.Cloudant.Password == "" {
	//	return nil, errors.New("Missing cloudant account password")
	//}
	//
	//creds.Cloudant.NotificationsURL = os.Getenv(cloudant.NotificationsURLEnv)
	//if creds.Cloudant.NotificationsURL == "" {
	//	return nil, errors.New("Missing notifications URL")
	//}
	//
	//creds.Cloudant.ServiceNamesURL = os.Getenv(cloudant.ServicesURLEnv)
	//if creds.Cloudant.ServiceNamesURL == "" {
	//	return nil, errors.New("Missing cloudant service names URL")
	//}
	//
	//creds.Cloudant.RuntimeNamesURL = os.Getenv(cloudant.RuntimesURLEnv)
	//if creds.Cloudant.RuntimeNamesURL == "" {
	//	return nil, errors.New("Missing cloudant runtime names URL")
	//}
	//
	//creds.Cloudant.PlatformNamesURL = os.Getenv(cloudant.PlatformURLEnv)
	//if creds.Cloudant.PlatformNamesURL == "" {
	//	return nil, errors.New("Missing cloudant platform names URL")
	//}

	creds.GlobalCatalog.ResourcesURL = os.Getenv(globalcatalog.GCResourceURL)
	if creds.GlobalCatalog.ResourcesURL == "" {
		return nil, errors.New("Missing global catalog resource URL")
	}

	err = creds.OSSCatalog.GetOSSCatalogCredentials()
	if err != nil {
		log.Println("GetCredentials: FAIL to get OSS catalog credentials.")
		return nil, err
	}

	return creds, nil
}
