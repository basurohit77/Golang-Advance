package servicenowdatamodel

// ServiceNowIncident
type ServiceNowWorkNote struct {
	Element      string `json:"element,omitempty"`
	ElementId    string `json:"element_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Number       string `json:"number,omitempty"`
	SysId        string `json:"sys_id,omitempty"`
	SysCreatedOn string `json:"sys_created_on,omitempty"`
	SysCreatedBy string `json:"sys_created_by,omitempty"`
	Value        string `json:"Value,omitempty"`
}

type ServiceNowWorkNotes []ServiceNowWorkNote

type ServiceNowWorkNotesResult struct {
	Result ServiceNowWorkNotes `json:"result"`
}
