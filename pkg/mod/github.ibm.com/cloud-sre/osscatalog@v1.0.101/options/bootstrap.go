package options

import (
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// Special functions used to bootstrap the debugging/logging
// This is needed to avoid a circular dependency between packages,
// because we need the "rest" package to find the LogDNA key, but
// the "rest" package depends on the "debug" package

// BootstrapLogDNA initializes the logging to LogDNA for this instance of the library
//
// Note: the function in this package takes care of bootstrapping
func BootstrapLogDNA(app string, panicOnError bool, maxLines int) error {
	if GlobalOptions().NoLogDNA {
		debug.Warning("Disabling error/audit logging to LogDNA for this run")
		return nil
	}
	key, err := rest.GetKey(debug.LogDNAIngestionKeyName)
	if err != nil {
		return debug.WrapError(err, "Cannot find the ingestion key for LogDNA")
	}
	return debug.SetupLogDNA(key, app, panicOnError, maxLines)
}
