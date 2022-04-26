package datastore

type CaseInsert struct {
	Source             string `json:"source,omitempty"`
	SourceID           string `json:"source_id,omitempty"`
	SourceSysID        string `json:"source_sys_id,omitempty"`
}

type CaseReturn struct {
	RecordID           string `json:"record_id,omitempty"`
	Source             string `json:"source,omitempty"`
	SourceID           string `json:"source_id,omitempty"`
	SourceSysID        string `json:"source_sys_id,omitempty"`}

