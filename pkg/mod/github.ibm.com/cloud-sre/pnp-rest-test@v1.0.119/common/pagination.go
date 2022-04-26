package common

// Pagination is the common pagination struct items
type Pagination struct {
	Href     string `json:"href"`
	Offset   int    `json:"offset"`
	Limit    int    `json:"limit"`
	Count    int    `json:"count"`
	First    Href   `json:"first"`
	Last     Href   `json:"last"`
	Previous Href   `json:"previous"`
	Next     Href   `json:"next"`
}
