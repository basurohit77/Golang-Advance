package datastore

type VisibilityInsert struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type VisibilityGet struct {
	RecordID    string `json:"record_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type VisibilityJunctionInsert struct {
	ResourceID   string `json:"resource_id,omitempty"`
	VisibilityID string `json:"visibility_id,omitempty"`
}

type VisibilityJunctionGet struct {
	ResourceID   string `json:"resource_id,omitempty"`
	VisibilityID string `json:"visibility_id,omitempty"`
}
