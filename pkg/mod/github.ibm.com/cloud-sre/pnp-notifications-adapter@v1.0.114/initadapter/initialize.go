package initadapter

import (
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/api"
	"github.ibm.com/cloud-sre/pnp-abstraction/exmon"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/mq"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/postgres"
)

// InitializationFunc is the type for the intialization function. Can be overriden for UT.
type InitializationFunc func() (*api.SourceConfig, error)

// Initialize must be called before any notification adapter methods can be called.
var Initialize = internalInitialize

func internalInitialize() (*api.SourceConfig, error) {
	postgres.SetupDBFunctions()
	mq.SetupMQFunctions()
	exmon.SetupMonitorFunctions()

	creds, err := api.GetCredentials()
	if err != nil {
		return nil, err
	}

	err = api.SetupOSSCatalogCredentials(creds)

	return creds, err
}
