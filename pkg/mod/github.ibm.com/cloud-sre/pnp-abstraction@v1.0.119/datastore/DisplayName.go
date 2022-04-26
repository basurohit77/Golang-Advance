package datastore

type DisplayName struct {
	Name     string `json:"name,omitempty"` // translated text
	Language string `json:"language,omitempty"`
}

type CatalogDisplayName []struct {
	Language string `json:"language,omitempty"`
	Text     string `json:"text,omitempty"`
}
