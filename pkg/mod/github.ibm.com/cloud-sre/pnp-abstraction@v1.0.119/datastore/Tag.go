package datastore

type Tag struct{
	ID          string `json:"id,omitempty"`
}

type TagInsert struct {
	ID          string `json:"id,omitempty"`
}

type TagGet struct {
	RecordID    string `json:"record_id,omitempty"`
	ID          string `json:"id,omitempty"`
}

type TagJunctionInsert struct {
	ResourceID   string `json:"resource_id,omitempty"`
	TagID        string `json:"tag_id,omitempty"`
}

type TagJunctionGet struct {
	ResourceID   string `json:"resource_id,omitempty"`
	TagID        string `json:"tag_id,omitempty"`
}
