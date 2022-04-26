package datamodel

type JournalQueryResponse struct {
	// The list of journal entries
	Resources []*JournalQueryEntry `json:"resources,omitempty"`
}

type JournalQueryEntry struct {

	// The time this journal entry was created.
	CreateTime string `json:"create_time,omitempty"`

	// The ID of the entity that creted this entry
	CreatedBy string `json:"created_by,omitempty"`

	// The value of this journal entry
	Value string `json:"value,omitempty"`
}
