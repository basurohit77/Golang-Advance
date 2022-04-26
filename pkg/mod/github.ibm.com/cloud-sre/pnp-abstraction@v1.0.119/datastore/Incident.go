package datastore

import (
	"database/sql"
)

// IncidentInsert metadata for incident table of insert commmand
type IncidentInsert struct {
	SourceCreationTime        string   `json:"source_creation_time,omitempty"`
	SourceUpdateTime          string   `json:"source_update_time,omitempty"`
	OutageStartTime           string   `json:"outage_start_time,omitempty"`
	OutageEndTime             string   `json:"outage_end_time,omitempty"`
	ShortDescription          string   `json:"short_description,omitempty"`
	LongDescription           string   `json:"long_description,omitempty"`
	State                     string   `json:"state,omitempty"`
	Classification            string   `json:"classification,omitempty"`
	Severity                  string   `json:"severity,omitempty"`
	CRNFull                   []string `json:"crnFull,omitempty"`
	SourceID                  string   `json:"source_id,omitempty"`
	Source                    string   `json:"source,omitempty"`
	RegulatoryDomain          string   `json:"regulatory_domain,omitempty"`
	AffectedActivity          string   `json:"affected_activity,omitempty"`
	CustomerImpactDescription string   `json:"customer_impact_description,omitempty"`
	PnPRemoved                bool     `json:"pnp_removed,omitempty"`
	TargetedURL               string   `json:"targeted_url,omitempty"`
	Audience                  string   `json:"audience,omitempty"`
}

// IncidentGet metadata for incident select
type IncidentGet struct {
	RecordID                  string         `json:"record_id,omitempty"`
	PnpCreationTime           string         `json:"pnp_creation_time,omitempty"`
	PnpUpdateTime             string         `json:"pnp_update_time,omitempty"`
	SourceCreationTime        string         `json:"source_creation_time,omitempty"`
	SourceUpdateTime          sql.NullString `json:"source_update_time,omitempty"`
	OutageStartTime           sql.NullString `json:"outage_start_time,omitempty"`
	OutageEndTime             sql.NullString `json:"outage_end_time,omitempty"`
	ShortDescription          sql.NullString `json:"short_description,omitempty"`
	LongDescription           sql.NullString `json:"long_description,omitempty"`
	State                     string         `json:"state,omitempty"`
	Classification            string         `json:"classification,omitempty"`
	Severity                  string         `json:"severity,omitempty"`
	CRNFull                   sql.NullString `json:"crnFull,omitempty"`
	SourceID                  string         `json:"source_id,omitempty"`
	Source                    string         `json:"source,omitempty"`
	RegulatoryDomain          sql.NullString `json:"regulatory_domain,omitempty"`
	AffectedActivity          sql.NullString `json:"affected_activity,omitempty"`
	CustomerImpactDescription sql.NullString `json:"customer_impact_description,omitempty"`
	PnPRemoved                string         `json:"pnp_removed,omitempty"`
	TargetedURL               sql.NullString `json:"targeted_url,omitempty"`
	Audience                  sql.NullString `json:"audience,omitempty"`
}

// IncidentGetNull This is for the case when offset is beyond the result set, and ends up with all Null in all columns in the returned row, except total_count
type IncidentGetNull struct {
	RecordID                  sql.NullString `json:"record_id,omitempty"`
	PnpCreationTime           sql.NullString `json:"pnp_creation_time,omitempty"`
	PnpUpdateTime             sql.NullString `json:"pnp_update_time,omitempty"`
	SourceCreationTime        sql.NullString `json:"source_creation_time,omitempty"`
	SourceUpdateTime          sql.NullString `json:"source_update_time,omitempty"`
	OutageStartTime           sql.NullString `json:"outage_start_time,omitempty"`
	OutageEndTime             sql.NullString `json:"outage_end_time,omitempty"`
	ShortDescription          sql.NullString `json:"short_description,omitempty"`
	LongDescription           sql.NullString `json:"long_description,omitempty"`
	State                     sql.NullString `json:"state,omitempty"`
	Classification            sql.NullString `json:"classification,omitempty"`
	Severity                  sql.NullString `json:"severity,omitempty"`
	CRNFull                   sql.NullString `json:"crnFull,omitempty"`
	SourceID                  sql.NullString `json:"source_id,omitempty"`
	Source                    sql.NullString `json:"source,omitempty"`
	RegulatoryDomain          sql.NullString `json:"regulatory_domain,omitempty"`
	AffectedActivity          sql.NullString `json:"affected_activity,omitempty"`
	CustomerImpactDescription sql.NullString `json:"customer_impact_description,omitempty"`
	PnPRemoved                sql.NullString `json:"pnp_removed,omitempty"`
	TargetedURL               sql.NullString `json:"targeted_url,omitempty"`
	Audience                  sql.NullString `json:"audience,omitempty"`
}

// IncidentReturn metadata of select statement
type IncidentReturn struct {
	RecordID                  string   `json:"record_id,omitempty"`
	PnpCreationTime           string   `json:"pnp_creation_time,omitempty"`
	PnpUpdateTime             string   `json:"pnp_update_time,omitempty"`
	SourceCreationTime        string   `json:"source_creation_time,omitempty"`
	SourceUpdateTime          string   `json:"source_update_time,omitempty"`
	OutageStartTime           string   `json:"outage_start_time,omitempty"`
	OutageEndTime             string   `json:"outage_end_time,omitempty"`
	ShortDescription          string   `json:"short_description,omitempty"`
	LongDescription           string   `json:"long_description,omitempty"`
	State                     string   `json:"state,omitempty"`
	Classification            string   `json:"classification,omitempty"`
	Severity                  string   `json:"severity,omitempty"`
	CRNFull                   []string `json:"crnFull,omitempty"`
	SourceID                  string   `json:"source_id,omitempty"`
	Source                    string   `json:"source,omitempty"`
	RegulatoryDomain          string   `json:"regulatory_domain,omitempty"`
	AffectedActivity          string   `json:"affected_activity,omitempty"`
	CustomerImpactDescription string   `json:"customer_impact_description,omitempty"`
	PnPRemoved                bool     `json:"pnp_removed,omitempty"`
	TargetedURL               string   `json:"targeted_url,omitempty"`
	Audience                  string   `json:"audience,omitempty"`
}

// IncidentJunctionGet metadata of incident_juntion get statment
type IncidentJunctionGet struct {
	RecordID   string `json:"record_id,omitempty"`
	ResourceID string `json:"resource_id,omitempty"`
	IncidentID string `json:"incident_id,omitempty"`
}
