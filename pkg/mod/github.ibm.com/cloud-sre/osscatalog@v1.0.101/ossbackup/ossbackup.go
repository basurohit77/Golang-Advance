package ossbackup

import (
	"encoding/json"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// Functions for backing-up the key information from OSS records outside of Global Catalog

// KeyBackup represents all the backup information that cannot be re-created from a merge
type KeyBackup struct {
	CanonicalName      ossrecord.CRNServiceName         `json:"canonical_name"`
	OSSTags            osstags.TagSet                   `json:"oss_tags"`
	ServiceNowSysid    ossrecord.ServiceNowSysid        `json:"servicenow_sys_id"`
	ParentResourceName ossrecord.CRNServiceName         `json:"parent_resource_name"`
	OSSMergeControl    *ossmergecontrol.OSSMergeControl `json:"oss_merge_control"`
}

// NewKeyBackup creates a new KeyBackup record containing all the key information from a given OSSServiceExtended
func NewKeyBackup(ossrec *ossrecordextended.OSSServiceExtended) *KeyBackup {
	return &KeyBackup{
		CanonicalName:      ossrec.OSSService.ReferenceResourceName,
		OSSTags:            ossrec.OSSService.GeneralInfo.OSSTags,
		ServiceNowSysid:    ossrec.OSSService.GeneralInfo.ServiceNowSysid,
		ParentResourceName: ossrec.OSSService.GeneralInfo.ParentResourceName,
		OSSMergeControl:    ossrec.OSSMergeControl,
	}
}

// String returns a string representation of this OSS KeyBackup record
func (k *KeyBackup) String() string {
	var result strings.Builder
	json, _ := json.MarshalIndent(k, "    ", "    ")
	_, err := result.Write(json)
	if err != nil {
		panic(err)
	}
	return result.String()
}
