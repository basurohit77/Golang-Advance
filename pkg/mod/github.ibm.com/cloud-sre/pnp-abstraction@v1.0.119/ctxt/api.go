package ctxt

import "github.ibm.com/cloud-sre/pnp-abstraction/exmon"

// Context carries information between calls
type Context struct {
	LogID string
	NRMon *exmon.Monitor
}
