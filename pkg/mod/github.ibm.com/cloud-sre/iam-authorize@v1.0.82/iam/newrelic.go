package iam

import (
	"context"
	instana "github.com/instana/go-sensor"
	newrelicPreModules "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.ibm.com/cloud-sre/iam-authorize/monitoring"
	"log"
)

/*
Wrapper for the NewRelic APM library. Allows IAM health check to be unit tested.

Example usage:

import (
    "log"
    "os"
    "time"
    "github.ibm.com/cloud-sre/iam-authorize/iam"
)

...

nrApp, err := iam.NewDefaultNRApp(iam.NRAppEnvDev, monitoringLicense)
if err != nil {
  log.Println(err.Error())
  return
}

nrWrapperConfig := iam.NRWrapperConfig{
	NRApp: nrApp,
	InstanaSensor: sensor,
	Ctx: content.Background(),
	Environment: "staging",
	Region: "us-east",
}

nrWrapper := iam.NewNRWrapper(nrWrapperConfig)
nrWrapper.SendTransaction("IAMHealthCheck", true, "")
*/

// NRAppEnv represents the environment of the NewRelic application.
type NRAppEnv string

const (
	// NRAppEnvProd represents the production environment
	NRAppEnvProd NRAppEnv = "prod"
	// NRAppEnvStage represents the staging environment
	NRAppEnvStage NRAppEnv = "staging"
	// NRAppEnvDev represents the development environment
	NRAppEnvDev NRAppEnv = "dev"
)

// NewDefaultNRApp creates a NewRelic application that will report metrics under the
// iam-authorize-iam-health-dev, iam-authorize-iam-health-stage, or iam-authorize-iam-health NewRelic
// application name depending on the env value. If there is a problem creating the NewRelic application,
// an error is returned.
func NewDefaultNRApp(env NRAppEnv, license string) (*newrelic.Application, error) {
	nrAppName := "iam-authorize-iam-health"
	if env == NRAppEnvProd {
		// no update needed
	} else if env == NRAppEnvStage {
		nrAppName = nrAppName + "-" + "stage"
	} else if env == NRAppEnvDev {
		nrAppName = nrAppName + "-" + "dev"
	} else {
		nrAppName = nrAppName + "-" + "test"
	}
	return NewNRApp(nrAppName, license)
}

// NewNRApp creates a new NewRelic application using the provided application name and license. If there is a
// problem creating the NewRelic application, an error is returned.
func NewNRApp(nrAppName string, nrLicense string) (*newrelic.Application, error) {
	nrApp, err := newrelic.NewApplication(
		newrelic.ConfigAppName(nrAppName),
		newrelic.ConfigLicense(nrLicense),
	)
	return nrApp, err
}

// NRWrapperConfig is the configuration needed to create a NewRelic wrapper.
type NRWrapperConfig struct {
	NRApp             newrelicPreModules.Application
	NRAppV3           *newrelic.Application
	InstanaSensor     *instana.Sensor
	Environment       string // environment to be reported to NewRelic in transaction
	Region            string // region to be reported to NewRelic in transaction
	URL               string // url to be reported to NewRelic in transaction
	SourceServiceName string // service name to be reported to NewRelic in transaction (is the source service)
}

// NRWrapper is the interface implemented by the NewRelic wrapper and represents the functions that are available.
type NRWrapper interface {
	SendTransaction(transactionName string, success bool, err string)
}

// NewNRWrapper constructs a instance of the wrapper given the provided configuration.
func NewNRWrapper(config NRWrapperConfig) NRWrapper {
	nrWrapper := &nrWrapper{}
	nrWrapper.config = config
	return nrWrapper
}

// nrWrapper is the implementation of the NewRelic wrapper interface.
type nrWrapper struct {
	config NRWrapperConfig
}

// SendTransaction creates a new APM transaction/span which is sent to the monitoring tool. The success parameter indicates
// what result should be reported, and the error parameter what error (if any).
func (nrWrapper *nrWrapper) SendTransaction(transactionName string, success bool, err string) {
	fct := transactionName + ":"

	// Instana
	var sensor *instana.Sensor
	if nrWrapper.config.InstanaSensor != nil {
		sensor = nrWrapper.config.InstanaSensor
	}
	healthCheckSpan, ctx := monitoring.NewSpan(context.Background(), sensor, transactionName)
	if healthCheckSpan == nil {
		log.Println(fct, "couldn't retrieve or create a span")
	} else {
		defer healthCheckSpan.Finish()
	}

	// New Relic
	var txnPreModules newrelicPreModules.Transaction
	if nrWrapper.config.NRApp != nil {
		txnPreModules = nrWrapper.config.NRApp.StartTransaction(transactionName, nil, nil)
		ctx = newrelicPreModules.NewContext(ctx, txnPreModules)
		defer txnPreModules.End()
	}

	var txn *newrelic.Transaction
	txn = nrWrapper.config.NRAppV3.StartTransaction(transactionName)
	ctx = newrelic.NewContext(ctx, txn)
	defer txn.End()

	// Send data to Instana and New Relic
	monitoring.SetTagsKV(ctx,
		"environment", nrWrapper.config.Environment,
		"region", nrWrapper.config.Region,
		"sourceServiceName", nrWrapper.config.SourceServiceName,
		"url", nrWrapper.config.URL,
		"result", success)
	if err != "" {
		monitoring.SetTag(ctx, "healthCheckError", err)
	}
}
