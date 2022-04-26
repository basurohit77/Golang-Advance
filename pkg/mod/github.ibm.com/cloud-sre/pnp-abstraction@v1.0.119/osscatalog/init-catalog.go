package osscatalog

import (
	"errors"
	"os"

	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
)

// This contains functions used to initialize the oss catalog cache functions
// The basic use is simply to call the InitializeCatalogFromScratch() function.
//
// There are some environment variables that need to be set.  These variables
// are dictated by the env.go file.  The basic variables to be set are:
//
//	   "PNP_CATALOG_CATYP_NAME"
//	   "PNP_CATALOG_CATYS1_NAME"
//	   "PNP_CATALOG_CATYP_VALUE"
//	   "PNP_CATALOG_CATYS1_VALUE"
//
// Set these in the oss-charts for your project. The following shows the info for each file
//
// *values.yaml*
//   PNP_CATALOG_CATYS1_NAME: osscatPlatformKey1
//   PNP_CATALOG_CATYP_NAME: osscatYPPlatformKey1
//
// *staging-values.yaml*
//   PNP_CATALOG_CATYP_VALUE: /generic/crn/v1/staging/local/tip-oss-flow/global/apiplatform/bluemix_api_tokens/catalog-yp
//   PNP_CATALOG_CATYS1_VALUE: /generic/crn/v1/staging/local/tip-oss-flow/global/apiplatform/bluemix_api_tokens/ys1Token
//
// *development-values.yaml*
//   PNP_CATALOG_CATYP_VALUE: /generic/crn/v1/dev/local/tip-oss-flow/global/apiplatform/bluemix_api_tokens/catalog-yp
//   PNP_CATALOG_CATYS1_VALUE: /generic/crn/v1/dev/local/tip-oss-flow/global/apiplatform/bluemix_api_tokens/ys1Token
//
// *production-values.yaml*
//   PNP_CATALOG_CATYP_VALUE: /generic/crn/v1/internal/local/tip-oss-flow/global/apiplatform/bluemix_api_tokens/catalog-yp
//   PNP_CATALOG_CATYS1_VALUE: /generic/crn/v1/internal/local/tip-oss-flow/global/apiplatform/bluemix_api_tokens/ys1Token
//
// *templates/deployment.yaml*
//   - name: PNP_CATALOG_CATYS1_NAME
//     value: "{{ .Values.PNP_CATALOG_CATYS1_NAME }}"
//   - name: PNP_CATALOG_CATYP_NAME
//     value: "{{ .Values.PNP_CATALOG_CATYP_NAME }}"
//   - name: PNP_CATALOG_CATYP_VALUE
//     valueFrom:
//       secretKeyRef:
//         name: {{ .Values.kdep.secrets.PNP_CATALOG_CATYP_VALUE }}
//         key: value
//   - name: PNP_CATALOG_CATYS1_VALUE
//     valueFrom:
//       secretKeyRef:
//         name: {{ .Values.kdep.secrets.PNP_CATALOG_CATYS1_VALUE }}
//         key: value

// Creds is the basic structure containing credentials used for the
// oss catalog library
type Creds struct {
	CredentialsSet bool
	Credentials    cfenv.Services
}

// InitializeCatalogFromScratch can be called if you only need to initialize the
// oss catalog cache functions.  If you plan to also use the pnp-notification-adapter
// you should use its initialization functions because it includes the oss catalog
// You may not have any reason to save the returned credentials for later use.
func InitializeCatalogFromScratch() (*Creds, error) {
	creds := &Creds{}
	err := InitializeCatalog(creds)
	return creds, err
}

var osscatLibInitialized = false
var ListingFunc ListOSSFunction

// InitializeCatalog is called from the pnp-notifications-adapter when initializing
// its structures.  May be useful for others.  It is expected that the input credentials
// are actually empty
func InitializeCatalog(creds *Creds) error {

	if osscatLibInitialized {
		return nil // Prevent repeated initializations
	}

	// We should always load using lenient mode
	options.GlobalOptions().Lenient = true

	err := creds.GetOSSCatalogCredentials()
	if err != nil {
		return err
	}

	err = SetupOSSCatalogCredentials(creds)

	if err != nil {
		return err
	}

	ListingFunc = GetDefaultListingFunction()

	cache, err := NewCache(ctxt.Context{}, ListingFunc)
	if err != nil || cache == nil {
		return errors.New("Unable to get a cache " + err.Error())
	}

	osscatLibInitialized = true

	return nil
}

// SetupOSSCatalogCredentials actually sets the credentials within the osscatalog library.
func SetupOSSCatalogCredentials(creds *Creds) error {
	if creds != nil && creds.CredentialsSet {
		return nil
	}

	// This is a bit of a hack right now.  Here we are loading credentials for use
	// by the OSS catalog by directly calling the load that it uses to load VCAP information.
	if creds == nil || creds.Credentials == nil {
		return errors.New("Missing OSSCatalog credentials")
	}
	creds.CredentialsSet = true
	return rest.LoadVCAPServices(creds.Credentials)
}

// GetOSSCatalogCredentials is used to pull the credential information from environment
// variables and setup the credential structure
func (creds *Creds) GetOSSCatalogCredentials() error {

	countOSSCreds := 0
	for i := 0; i < len(OSSCatalogCredentialLookup); i = i + 3 {

		keyLabel := os.Getenv(OSSCatalogCredentialLookup[i+1])
		keyValue := os.Getenv(OSSCatalogCredentialLookup[i+2])

		if keyLabel != "" && keyValue != "" {
			creds.AddOSSCatalogCredential(OSSCatalogCredentialLookup[i], keyLabel, keyValue)
			countOSSCreds++
		}
	}

	if countOSSCreds == 0 {
		return errors.New("We need at least one OSS catalog credential")
	}

	return nil
}

// AddOSSCatalogCredential is a convenience method to add oss credentials to the credential struct
func (creds *Creds) AddOSSCatalogCredential(keyName, label, value string) {

	if creds.Credentials == nil {
		creds.Credentials = make(cfenv.Services)
	}

	s := make([]cfenv.Service, 1)

	s[0].Name = keyName
	s[0].Credentials = make(map[string]interface{})
	s[0].Credentials["name"] = label
	s[0].Credentials["key"] = value

	creds.Credentials[keyName] = s
}
