package datastore

type NotificationDescriptionGet struct {
	RecordID          string `json:"record_id,omitempty"`
	LongDescription   string `json:"long_description,omitempty"`
	Language          string `json:"language,omitempty"`
}
