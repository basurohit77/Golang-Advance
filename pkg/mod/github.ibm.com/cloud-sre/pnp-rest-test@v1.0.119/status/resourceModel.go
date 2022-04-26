package status

import (
	"github.ibm.com/cloud-sre/pnp-rest-test/common"
)

// ResourceGet defines the resource returned on a GET
type ResourceGet struct {
	Href             string                    `json:"href"`
	RecordID         string                    `json:"recordID"`
	CreationTime     string                    `json:"creationTime"`
	UpdateTime       string                    `json:"updateTime"`
	Kind             string                    `json:"kind"`
	CRNMask          string                    `json:"crnMask"`
	DisplayName      []common.TranslatedString `json:"displayName"`
	State            string                    `json:"state"`
	Status           string                    `json:"status"`
	StatusUpdateTime string                    `json:"statusUpdateTime"`
	Visibility       []string                  `json:"visibility"`
	Tags             []common.SimpleTag        `json:"tags"`
}

// ResourceList is returned on queries of multiple resources
type ResourceList struct {
	common.Pagination
	Resources []ResourceGet `json:"resources"`
}
