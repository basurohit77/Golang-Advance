package catalog

import "github.ibm.com/cloud-sre/pnp-rest-test/common"

// APIInfo represents the struct for the status API info
type APIInfo struct {
	ClientID             string      `json:"clientId"`
	Description          string      `json:"description"`
	Version              string      `json:"version"`
	Categories           []string    `json:"categories"`
	APIImplementations   common.Href `json:"apiImplementations"`
	CatalogSubscriptions common.Href `json:"catalogSubscriptions"`
	CatalogList          common.Href `json:"catalogList"`
	Healthz              common.Href `json:"healthz"`
}
