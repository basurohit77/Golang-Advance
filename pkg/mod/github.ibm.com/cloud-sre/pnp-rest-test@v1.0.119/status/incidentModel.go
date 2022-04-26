package status

import (
	"github.ibm.com/cloud-sre/pnp-rest-test/common"
)

// IncidentGet defines the resource returned on a GET
type IncidentGet struct {
	Href             string   `json:"href"`
	RecordID         string   `json:"recordID"`
	CreationTime     string   `json:"creationTime"`
	UpdateTime       string   `json:"updateTime"`
	Kind             string   `json:"kind"`
	ShortDescription string   `json:"shortDescription"`
	LongDescription  string   `json:"longDescription"`
	CRNMasks         []string `json:"crnMasks"`
	State            string   `json:"state"`
	Classification   string   `json:"classification"`
	Severity         int      `json:"severity"`
	OutageStart      string   `json:"outageStart"`
	OutageEnd        string   `json:"outageEnd"`
	Source           string   `json:"source"`
	SourceID         string   `json:"sourceID"`
}

// IncidentList is returned on queries of multiple resources
type IncidentList struct {
	common.Pagination
	Resources []IncidentGet `json:"resources"`
}
