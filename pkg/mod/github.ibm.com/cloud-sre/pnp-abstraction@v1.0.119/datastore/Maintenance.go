package datastore

import (
	"database/sql"
)

// MaintenanceInsert metadata for maintenance table of insert commmand
type MaintenanceInsert struct {
	SourceCreationTime    string   `json:"source_creation_time,omitempty"`
	SourceUpdateTime      string   `json:"source_update_time,omitempty"`
	PlannedStartTime      string   `json:"planned_start_time,omitempty"`
	PlannedEndTime        string   `json:"planned_end_time,omitempty"`
	ShortDescription      string   `json:"short_description,omitempty"`
	LongDescription       string   `json:"long_description,omitempty"`
	CRNFull               []string `json:"crnFull,omitempty"`
	State                 string   `json:"state,omitempty"`
	Disruptive            bool     `json:"disruptive,omitempty"`
	SourceID              string   `json:"source_id,omitempty"`
	Source                string   `json:"source,omitempty"`
	RegulatoryDomain      string   `json:"regulatory_domain,omitempty"`
	RecordHash            string   `json:"record_hash,omitempty"`
	MaintenanceDuration   int      `json:"maintenance_duration,omitempty"`
	DisruptionType        string   `json:"disruption_type,omitempty"`
	DisruptionDescription string   `json:"disruption_description,omitempty"`
	DisruptionDuration    int      `json:"disruption_duration,omitempty"`
	CompletionCode        string   `json:"completion_code,omitempty"`
	PnPRemoved            bool     `json:"pnp_removed,omitempty"`
	TargetedURL           string   `json:"targeted_url,omitempty"`
	Audience              string   `json:"audience,omitempty"`
}

// MaintenanceGet metadata for maintenance select
type MaintenanceGet struct {
	RecordID              string         `json:"record_id,omitempty"`
	PnpCreationTime       string         `json:"pnp_creationTime,omitempty"`
	PnpUpdateTime         string         `json:"pnp_update_time,omitempty"`
	SourceCreationTime    string         `json:"source_creation_time,omitempty"`
	SourceUpdateTime      sql.NullString `json:"source_update_time,omitempty"`
	PlannedStartTime      sql.NullString `json:"planned_start_time,omitempty"`
	PlannedEndTime        sql.NullString `json:"planned_end_time,omitempty"`
	ShortDescription      sql.NullString `json:"short_description,omitempty"`
	LongDescription       sql.NullString `json:"long_description,omitempty"`
	State                 string         `json:"state,omitempty"`
	Disruptive            string         `json:"disruptive,omitempty"`
	SourceID              string         `json:"source_id,omitempty"`
	Source                string         `json:"source,omitempty"`
	RegulatoryDomain      sql.NullString `json:"regulatory_domain,omitempty"`
	CRNFull               sql.NullString `json:"crnFull,omitempty"`
	RecordHash            sql.NullString `json:"record_hash,omitempty"`
	MaintenanceDuration   string         `json:"maintenance_duration,omitempty"`
	DisruptionType        sql.NullString `json:"disruption_type,omitempty"`
	DisruptionDescription sql.NullString `json:"disruption_description,omitempty"`
	DisruptionDuration    string         `json:"disruption_duration,omitempty"`
	CompletionCode        sql.NullString `json:"completion_code,omitempty"`
	PnPRemoved            string         `json:"pnp_removed,omitempty"`
	TargetedURL           sql.NullString `json:"targeted_url,omitempty"`
	Audience              sql.NullString `json:"audience,omitempty"`
}

// MaintenanceGetNull This is for the case when offset is beyond the result set, and ends up with all Null in all columns in the returned row, except total_count
type MaintenanceGetNull struct {
	RecordID              sql.NullString `json:"record_id,omitempty"`
	PnpCreationTime       sql.NullString `json:"pnp_creationTime,omitempty"`
	PnpUpdateTime         sql.NullString `json:"pnp_update_time,omitempty"`
	SourceCreationTime    sql.NullString `json:"source_creation_time,omitempty"`
	SourceUpdateTime      sql.NullString `json:"source_update_time,omitempty"`
	PlannedStartTime      sql.NullString `json:"planned_start_time,omitempty"`
	PlannedEndTime        sql.NullString `json:"planned_end_time,omitempty"`
	ShortDescription      sql.NullString `json:"short_description,omitempty"`
	LongDescription       sql.NullString `json:"long_description,omitempty"`
	State                 sql.NullString `json:"state,omitempty"`
	Disruptive            sql.NullString `json:"disruptive,omitempty"`
	SourceID              sql.NullString `json:"source_id,omitempty"`
	Source                sql.NullString `json:"source,omitempty"`
	RegulatoryDomain      sql.NullString `json:"regulatory_domain,omitempty"`
	CRNFull               sql.NullString `json:"crnFull,omitempty"`
	RecordHash            sql.NullString `json:"record_hash,omitempty"`
	MaintenanceDuration   sql.NullString `json:"maintenance_duration,omitempty"`
	DisruptionType        sql.NullString `json:"disruption_type,omitempty"`
	DisruptionDescription sql.NullString `json:"disruption_description,omitempty"`
	DisruptionDuration    sql.NullString `json:"disruption_duration,omitempty"`
	CompletionCode        sql.NullString `json:"completion_code,omitempty"`
	PnPRemoved            sql.NullString `json:"pnp_removed,omitempty"`
	TargetedURL           sql.NullString `json:"targeted_url,omitempty"`
	Audience              sql.NullString `json:"audience,omitempty"`
}

// MaintenanceReturn metadata of select statement
type MaintenanceReturn struct {
	RecordID              string   `json:"record_id,omitempty"`
	PnpCreationTime       string   `json:"pnp_creationTime,omitempty"`
	PnpUpdateTime         string   `json:"pnp_update_time,omitempty"`
	SourceCreationTime    string   `json:"source_creation_time,omitempty"`
	SourceUpdateTime      string   `json:"source_update_time,omitempty"`
	PlannedStartTime      string   `json:"planned_start_time,omitempty"`
	PlannedEndTime        string   `json:"planned_end_time,omitempty"`
	ShortDescription      string   `json:"short_description,omitempty"`
	LongDescription       string   `json:"long_description,omitempty"`
	State                 string   `json:"state,omitempty"`
	Disruptive            bool     `json:"disruptive,omitempty"`
	SourceID              string   `json:"source_id,omitempty"`
	Source                string   `json:"source,omitempty"`
	RegulatoryDomain      string   `json:"regulatory_domain,omitempty"`
	CRNFull               []string `json:"crnFull,omitempty"`
	RecordHash            string   `json:"record_hash,omitempty"`
	MaintenanceDuration   int      `json:"maintenance_duration,omitempty"`
	DisruptionType        string   `json:"disruption_type,omitempty"`
	DisruptionDescription string   `json:"disruption_description,omitempty"`
	DisruptionDuration    int      `json:"disruption_duration,omitempty"`
	CompletionCode        string   `json:"completion_code,omitempty"`
	PnPRemoved            bool     `json:"pnp_removed,omitempty"`
	TargetedURL           string   `json:"targeted_url,omitempty"`
	Audience              string   `json:"audience,omitempty"`
}

// MaintenanceJunctionGet metadata of maintenance_juntion get statment
type MaintenanceJunctionGet struct {
	RecordID      string `json:"record_id,omitempty"`
	ResourceID    string `json:"resource_id,omitempty"`
	MaintenanceID string `json:"maintenance_id,omitempty"`
}

//MaintenanceMap use to upload SN paased values
type MaintenanceMap struct {
	RecordID               string   `json:"record_id,omitempty"`
	PnpCreationTime        string   `json:"pnp_creationTime,omitempty"`
	PnpUpdateTime          string   `json:"pnp_update_time,omitempty"`
	SourceCreationTime     string   `json:"source_creation_time,omitempty"`
	SourceUpdateTime       string   `json:"source_update_time,omitempty"`
	PlannedStartTime       string   `json:"planned_start_time,omitempty"`
	PlannedEndTime         string   `json:"planned_end_time,omitempty"`
	ShortDescription       string   `json:"short_description,omitempty"`
	LongDescription        string   `json:"long_description,omitempty"`
	State                  string   `json:"state,omitempty"`
	Disruptive             bool     `json:"disruptive,omitempty"`
	SourceID               string   `json:"source_id,omitempty"`
	Source                 string   `json:"source,omitempty"`
	RegulatoryDomain       string   `json:"regulatory_domain,omitempty"`
	CRNFull                []string `json:"crnFull,omitempty"`
	RecordHash             string   `json:"record_hash,omitempty"`
	MaintenanceDuration    int      `json:"maintenance_duration,omitempty"`
	DisruptionType         string   `json:"disruption_type,omitempty"`
	DisruptionDescription  string   `json:"disruption_description,omitempty"`
	DisruptionDuration     int      `json:"disruption_duration,omitempty"`
	CompletionCode         string   `json:"completion_code,omitempty"`
	ShouldHaveNotification bool     `json:"should_have_notification"`
	PnPRemoved             bool     `json:"pnp_removed,omitempty"`
	TargetedURL            string   `json:"targeted_url,omitempty"`
	Audience               string   `json:"audience,omitempty"`
}
