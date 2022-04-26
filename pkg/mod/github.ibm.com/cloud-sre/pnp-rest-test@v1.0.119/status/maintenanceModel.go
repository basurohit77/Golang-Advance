package status

import (
	"github.ibm.com/cloud-sre/pnp-rest-test/common"
)

// MaintenanceGet defines the resource returned on a GET
type MaintenanceGet struct {
	Href             string   `json:"href"`
	RecordID         string   `json:"recordID"`
	CreationTime     string   `json:"creationTime"`
	UpdateTime       string   `json:"updateTime"`
	Kind             string   `json:"kind"`
	ShortDescription string   `json:"shortDescription"`
	LongDescription  string   `json:"longDescription"`
	Disruptive       string   `json:"disruptive"`
	CRNMasks         []string `json:"crnMasks"`
	State            string   `json:"state"`
	PlannedStart     string   `json:"plannedStart"`
	PlannedEnd       string   `json:"plannedEnd"`
	Source           string   `json:"source"`
	SourceID         string   `json:"sourceID"`
}

// MaintenanceList is returned on queries of multiple resources
type MaintenanceList struct {
	common.Pagination
	Resources []MaintenanceGet `json:"resources"`
}
