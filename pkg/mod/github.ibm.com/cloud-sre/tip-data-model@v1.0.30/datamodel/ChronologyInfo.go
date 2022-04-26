package datamodel

type ChronologyBody struct {
	Content		string		`json:"content"`
}

//
// ServiceNow Payload for patch an incident's chronology
//
type ChronologyPayload struct {
	// SN: u_chronology_incident <- Concern: none
	UChronologyIncident        string `json:"u_chronology_incident"`
	// SN: work_notes <- Concern: journal
	WorkNotes                  string `json:"work_notes,omitempty"`
}