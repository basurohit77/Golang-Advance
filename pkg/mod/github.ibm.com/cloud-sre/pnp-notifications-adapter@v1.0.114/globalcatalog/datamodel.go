package globalcatalog

import (
	"encoding/json"
	"time"
)

// CloudResourcesList is a list of resource information from the global catalog
type CloudResourcesList struct {
	Offset        int              `json:"offset"`
	Limit         int              `json:"limit"`
	Count         int              `json:"count"`
	ResourceCount int              `json:"resource_count"`
	First         string           `json:"first"`
	Next          string           `json:"next"`
	Resources     []*CloudResource `json:"resources"`
}

// CloudResource is a single record
type CloudResource struct {
	Active          bool            `json:"active"`
	Name            string          `json:"name"`
	Tags            []string        `json:"tags"`
	OverviewUI      json.RawMessage `json:"overview_ui"`
	LanguageStrings []*ResourceLang
}

// CloudResourceCache is a cache keyed by the service name
type CloudResourceCache struct {
	CacheURL    string
	CacheExpire time.Time
	Resources   map[string]*CloudResource
}

// ResourceLang is a struct providing language information
type ResourceLang struct {
	Language        string
	Description     string `json:"description"`
	DisplayName     string `json:"display_name"`
	LongDescription string `json:"long_description"`
}
