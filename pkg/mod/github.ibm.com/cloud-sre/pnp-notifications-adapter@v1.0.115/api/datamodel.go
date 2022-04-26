package api

// NotificationList is a list of normalized notifications. Intended to hide the cloudant specifics.
type NotificationList struct {
	Items []*Notification
}

// Notification is a normalized notification object. Intended to hide the cloudant specifics.
type Notification struct {
	RecordID               string // Set by database
	ShortDescription       string
	LongDescription        string
	CRNs                   []string
	EventTimeStart         string
	EventTimeEnd           string
	IncidentID             string
	CreationTime           string
	UpdateTime             string
	PNPCreationTime        string
	PNPUpdateTime          string
	Category               string
	CategoryNotificationID string
	Source                 string
	SourceID               string
	NotificationType       string
	DisplayName            []*TranslatedString
	Tags                   []string
	PnPRemoved             bool
}

// TranslatedString represents a list of strings translated for languages
type TranslatedString struct {
	Language string
	Text     string
}
