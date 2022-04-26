package auth

import (
	"github.com/newrelic/go-agent"
	iam "github.ibm.com/cloud-sre/iam-authorize"
	iamp "github.ibm.com/cloud-sre/iam-authorize/iam"
	"net/http"
	"os"
)

// IAMAuth defines the functions supported
type IAMAuth interface {
	IsIAMAuthorized(req *http.Request, resp http.ResponseWriter) (*iam.Authorization, error)
}

// NewIAMAuth creates a new IAM authorization instance that can be used to authorize a request
func NewIAMAuth(nrApp newrelic.Application) IAMAuth {
	// Create NewRelic wrapper containing NewRelic app so IAM monitoring can report IAM health status to NewRelic:
	nrWrapperConfig := &iamp.NRWrapperConfig{
		NRApp:             nrApp,
		Environment:       os.Getenv("KUBE_APP_DEPLOYED_ENV"),
		Region:            os.Getenv("KUBE_CLUSTER_REGION"),
		SourceServiceName: "api-pnp-hooks",
	}

	// Create IAM Auth config and IAM auth with monitor:
	newIAMAuthConfig := iam.NewIAMAuthConfig{
		SvcName:                 os.Getenv("IAM_SERVICE_NAME"),
		IAMURL:                  os.Getenv("IAM_PUBLIC_URL"),
		IAMAPIKey:               os.Getenv("API_PLATFORM_API_KEY"),
		NRConfig:                nrWrapperConfig,
		EnableAutoIAMBreakGlass: os.Getenv("ENABLE_AUTO_IAM_BREAK_GLASS") == "true",
	}

	return NewIAMAuthFromConfig(newIAMAuthConfig)
}

// NewIAMAuthFromConfig creates a new IAM authorization instance that can be used to authorize a request
func NewIAMAuthFromConfig(newIAMAuthConfig iam.NewIAMAuthConfig) IAMAuth {
	iamAuth := &defaultIAMAuth{}
	iamAuth.iamAuth = iam.NewIAMAuthWithIAMMonitor(newIAMAuthConfig)
	iamAuth.iamAuth.SetIAMResourceActionPlugin(IAMResourceActionPluginType{})
	return iamAuth
}

// defaultIAMAuth is the default implementation of the IAMAuth interface that uses iam-authorize
type defaultIAMAuth struct {
	iamAuth *iam.IAMAuth
}

// IsIAMAuthorized checks auth of the request. It returns the email address of the user only if
// the user is authorized, or an error if not.
func (iamAuth *defaultIAMAuth) IsIAMAuthorized(req *http.Request, resp http.ResponseWriter) (*iam.Authorization, error) {
	return iamAuth.iamAuth.IsIAMAuthorized(req, resp)
}

