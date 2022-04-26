package common

// HealthZ is a common struct returned by all healthz APIs across the API categories
type HealthZ struct {
	Href        string `json:"href"`
	Code        int    `json:"code"`
	Description string `json:"description"`
}
