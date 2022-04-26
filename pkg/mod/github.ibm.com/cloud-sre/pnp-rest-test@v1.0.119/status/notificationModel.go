package status

import (
	"github.ibm.com/cloud-sre/pnp-rest-test/common"
)

// NotificationGet defines the resource returned on a GET
type NotificationGet struct {
	Href                string                    `json:"href"`
	RecordID            string                    `json:"recordID"`
	CreationTime        string                    `json:"creationTime"`
	UpdateTime          string                    `json:"updateTime"`
	EventTimeStart      string                    `json:"eventTimeStart"`
	EventTimeEnd        string                    `json:"eventTimeEnd"`
	Kind                string                    `json:"kind"`
	Type                string                    `json:"type"`
	Category            string                    `json:"category"`
	ShortDescription    []common.TranslatedString `json:"shortDescription"`
	LongDescription     []common.TranslatedString `json:"longDescription"`
	ResourceDisplayName []common.TranslatedString `json:"resourceDisplayName"`
	CRNMasks            []string                  `json:"crnMasks"`
	Source              string                    `json:"source"`
	SourceID            string                    `json:"sourceID"`
}

// NotificationList is returned on queries of multiple resources
type NotificationList struct {
	common.Pagination
	Resources []NotificationGet `json:"resources"`
}
