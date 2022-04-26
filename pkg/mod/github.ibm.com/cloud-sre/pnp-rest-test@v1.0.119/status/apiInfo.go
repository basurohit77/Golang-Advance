package status

import "github.ibm.com/cloud-sre/pnp-rest-test/common"

// APIInfo represents the struct for the status API info
type APIInfo struct {
	ClientID      string      `json:"clientId"`
	Description   string      `json:"description"`
	Version       string      `json:"version"`
	Categories    []string    `json:"categories"`
	Resources     common.Href `json:"resources"`
	Incidents     common.Href `json:"incidents"`
	Maintenances  common.Href `json:"maintenances"`
	Notifications common.Href `json:"notifications"`
	Healthz       common.Href `json:"healthz"`
}
