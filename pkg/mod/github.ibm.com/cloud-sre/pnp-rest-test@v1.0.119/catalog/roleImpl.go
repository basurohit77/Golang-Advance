package catalog

import "github.ibm.com/cloud-sre/pnp-rest-test/common"

// RoleImpl is a implementation
type RoleImpl struct {
	Href       string      `json:"href"`
	ID         string      `json:"id"`
	ClientID   string      `json:"clientId"`
	Categories []string    `json:"categories"`
	SourceInfo common.Href `json:"sourceInfo"`
}

// RoleImplList is a list of RoleImple
type RoleImplList struct {
	Href  string     `json:"href"`
	Impls []RoleImpl `json:"impls"`
}
