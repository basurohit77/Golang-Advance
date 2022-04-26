package datastore

import (
	"database/sql"
)

// ResourceInsert - Used for inserts and updates to the database
type ResourceInsert struct {
	SourceCreationTime string        `json:"source_creation_time,omitempty"`
	SourceUpdateTime   string        `json:"source_update_time,omitempty"`
	CRNFull            string        `json:"crn_full,omitempty"`
	State              string        `json:"state,omitempty"`
	OperationalStatus  string        `json:"operational_status,omitempty"`
	Source             string        `json:"source,omitempty"`
	SourceID           string        `json:"source_id,omitempty"`
	Status             string        `json:"status,omitempty"`
	StatusUpdateTime   string        `json:"status_update_time,omitempty"`
	RegulatoryDomain   string        `json:"regulatory_domain,omitempty"`
	CategoryID         string        `json:"category_id,omitempty"`
	CategoryParent     bool          `json:"category_parent,omitempty"`
	IsCatalogParent    bool          `json:"is_catalog_parent,omitempty"`
	CatalogParentID    string        `json:"catalog_parent_id,omitempty"`
	DisplayNames       []DisplayName `json:"displayName,omitempty"`
	Visibility         []string      `json:"visibility,omitempty"`
	Tags               []Tag         `json:"tags,omitempty"`
	RecordHash         string        `json:"record_hash,omitempty"`
}

// ResourceGet - Used to scan in results from the db and then be converted into a ResourceReturn
type ResourceGet struct {
	RecordID           string         `json:"record_id,omitempty"`
	PnpCreationTime    string         `json:"pnp_creation_time,omitempty"`
	PnpUpdateTime      string         `json:"pnp_update_time,omitempty"`
	SourceCreationTime sql.NullString `json:"source_creation_time,omitempty"`
	SourceUpdateTime   sql.NullString `json:"source_update_time,omitempty"`
	CRNFull            string         `json:"crn_full,omitempty"`
	State              sql.NullString `json:"state,omitempty"`
	OperationalStatus  sql.NullString `json:"operational_status,omitempty"`
	Source             string         `json:"source,omitempty"`
	SourceID           string         `json:"source_id,omitempty"`
	Status             sql.NullString `json:"status,omitempty"`
	StatusUpdateTime   sql.NullString `json:"status_update_time,omitempty"`
	RegulatoryDomain   sql.NullString `json:"regulatory_domain,omitempty"`
	CategoryID         sql.NullString `json:"category_id,omitempty"`
	CategoryParent     string         `json:"category_parent,omitempty"`
	IsCatalogParent    bool           `json:"is_catalog_parent,omitempty"`
	CatalogParentID    sql.NullString `json:"catalog_parent_id,omitempty"`
	RecordHash         sql.NullString `json:"record_hash,omitempty"`
}

// ResourceGetNull - Used by GetResourceByQuery in case it's a funky record
type ResourceGetNull struct {
	RecordID           sql.NullString `json:"record_id,omitempty"`
	PnpCreationTime    sql.NullString `json:"pnp_creation_time,omitempty"`
	PnpUpdateTime      sql.NullString `json:"pnp_update_time,omitempty"`
	SourceCreationTime sql.NullString `json:"source_creation_time,omitempty"`
	SourceUpdateTime   sql.NullString `json:"source_update_time,omitempty"`
	CRNFull            sql.NullString `json:"crn_full,omitempty"`
	State              sql.NullString `json:"state,omitempty"`
	OperationalStatus  sql.NullString `json:"operational_status,omitempty"`
	Source             sql.NullString `json:"source,omitempty"`
	SourceID           sql.NullString `json:"source_id,omitempty"`
	Status             sql.NullString `json:"status,omitempty"`
	StatusUpdateTime   sql.NullString `json:"status_update_time,omitempty"`
	RegulatoryDomain   sql.NullString `json:"regulatory_domain,omitempty"`
	CategoryID         sql.NullString `json:"category_id,omitempty"`
	CategoryParent     sql.NullString `json:"category_parent,omitempty"`
	// Catalog Parent concept more of a group concept since there is no real inheritance
	IsCatalogParent sql.NullString `json:"is_catalog_parent,omitempty"`
	CatalogParentID sql.NullString `json:"catalog_parent_id,omitempty"`
	RecordHash      sql.NullString `json:"record_hash,omitempty"`
}

type ResourceReturn struct {
	RecordID           string `json:"record_id,omitempty"`
	PnpCreationTime    string `json:"pnp_creation_time,omitempty"`
	PnpUpdateTime      string `json:"pnp_update_time,omitempty"`
	SourceCreationTime string `json:"source_creation_time,omitempty"`
	SourceUpdateTime   string `json:"source_update_time,omitempty"`
	CRNFull            string `json:"crn_full,omitempty"`
	State              string `json:"state,omitempty"`
	OperationalStatus  string `json:"operational_status,omitempty"`
	Source             string `json:"source,omitempty"`
	SourceID           string `json:"source_id,omitempty"`
	Status             string `json:"status,omitempty"`
	StatusUpdateTime   string `json:"status_update_time,omitempty"`
	RegulatoryDomain   string `json:"regulatory_domain,omitempty"`
	CategoryID         string `json:"category_id,omitempty"`
	CategoryParent     bool   `json:"category_parent,omitempty"`
	// Catalog Parent concept more of a group concept since there is no real inheritance
	IsCatalogParent bool          `json:"is_catalog_parent,omitempty"`
	CatalogParentID string        `json:"catalog_parent_id,omitempty"`
	DisplayNames    []DisplayName `json:"displayName,omitempty"`
	Visibility      []string      `json:"visibility,omitempty"`
	Tags            []Tag         `json:"tags,omitempty"`
	RecordHash      string        `json:"record_hash,omitempty"`
}

type CatalogParentResource struct {
	Href     string `json:"href,omitempty"`
	RecordID string `json:"record_id,omitempty"`
}

type PnpStatusResource struct {
	Id         string `json:"id,omitempty"`
	Href       string `json:"href,omitempty"`
	RecordID   string `json:"recordID,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Mode       string `json:"mode,omitempty"`
	CategoryId string `json:"categoryId,omitempty"`
	// Catalog Parent concept more of a group concept since there is no real inheritance
	IsCatalogParent bool               `json:"is_catalog_parent,omitempty"`
	CatalogParentID string             `json:"catalog_parent_id,omitempty"`
	Crn             string             `json:"crn,omitempty"`
	Tags            []string           `json:"-"`
	DisplayName     CatalogDisplayName `json:"displayName,omitempty"`
	EntryType       string             `json:"entry_type,omitempty"`

	SNId              string           `json:"servicenow_sys_id,omitempty"`
	SnCIUrl           string           `json:"service_now_ciurl,omitempty"`
	State             string           `json:"state,omitempty"`
	Status            string           `json:"status,omitempty"`
	OperationalStatus string           `json:"operationalStatus,omitempty"`
	Visibility        []string         `json:"visibility,omitempty"`
	Source            string           `json:"source"`
	SourceID          string           `json:"sourceId"`
	CreationTime      string           `json:"creationTime,omitempty"`
	UpdateTime        string           `json:"updateTime,omitempty"`
	Deployments       []*PnpDeployment `json:"deployments,omitempty"`
	Parent            bool             `json:"-"`
}

type PnpDeployment struct {
	Id                string             `json:"id,omitempty"`
	Active            string             `json:"active,omitempty"`
	Disabled          string             `json:"disabled,omitempty"`
	Href              string             `json:"href,omitempty"`
	RecordID          string             `json:"recordID,omitempty"`
	Kind              string             `json:"kind,omitempty"`
	CategoryId        string             `json:"categoryId,omitempty"`
	Crn               string             `json:"crn,omitempty"`
	Tags              []string           `json:"tags,omitempty"`
	DisplayName       CatalogDisplayName `json:"displayName,omitempty"`
	EntryType         string             `json:"entry_type,omitempty"`
	State             string             `json:"state,omitempty"`
	Status            string             `json:"status,omitempty"`
	OperationalStatus string             `json:"operationalStatus,omitempty"`
	Visibility        []string           `json:"visibility,omitempty"`
	Source            string             `json:"source,omitempty"`
	SourceID          string             `json:"sourceId,omitempty"`
	CreationTime      string             `json:"creationTime,omitempty"`
	UpdateTime        string             `json:"updateTime,omitempty"`
	Parent            bool               `json:"parent,omitempty"`
	// Catalog Parent concept more of a group concept since there is no real inheritance
	IsCatalogParent bool   `json:"is_catalog_parent,omitempty"`
	CatalogParentID string `json:"catalog_parent_resource_id,omitempty"`
}
