package datastore

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"

	"github.ibm.com/cloud-sre/oss-globals/consts"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
)

// NotificationInsert - Used for inserts and updates notifications to the database
type NotificationInsert struct {
	SourceCreationTime   string        `json:"source_creation_time,omitempty"`
	SourceUpdateTime     string        `json:"source_update_time,omitempty"`
	EventTimeStart       string        `json:"event_time_start,omitempty"`
	EventTimeEnd         string        `json:"event_time_end,omitempty"`
	Source               string        `json:"source,omitempty"`
	SourceID             string        `json:"source_id,omitempty"`
	Type                 string        `json:"type,omitempty"`
	Category             string        `json:"category,omitempty"`
	IncidentID           string        `json:"incident_id,omitempty"`
	CRNFull              string        `json:"crn_full,omitempty"`
	ResourceDisplayNames []DisplayName `json:"resource_display_names,omitempty"`
	ShortDescription     []DisplayName `json:"short_description,omitempty"`
	LongDescription      []DisplayName `json:"long_description,omitempty"`
	Tags                 string        `json:"tags,omitempty"`
	RecordRetractionTime string        `json:"record_retraction_time,omitempty"`
	PnPRemoved           bool          `json:"pnp_removed,omitempty"`
	ReleaseNoteUrl       string        `json:"release_note_url,omitempty"`
}

// NotificationGet - Used to scan in results from the db and then be converted into a NotificationReturn
type NotificationGet struct {
	RecordID             string         `json:"record_id,omitempty"`
	PnpCreationTime      string         `json:"pnp_creation_time,omitempty"`
	PnpUpdateTime        string         `json:"pnp_update_time,omitempty"`
	SourceCreationTime   sql.NullString `json:"source_creation_time,omitempty"`
	SourceUpdateTime     sql.NullString `json:"source_update_time,omitempty"`
	EventTimeStart       sql.NullString `json:"event_time_start,omitempty"`
	EventTimeEnd         sql.NullString `json:"event_time_end,omitempty"`
	Source               string         `json:"source,omitempty"`
	SourceID             string         `json:"source_id,omitempty"`
	Type                 string         `json:"type,omitempty"`
	Category             sql.NullString `json:"category,omitempty"`
	IncidentID           sql.NullString `json:"incident_id,omitempty"`
	CRNFull              string         `json:"crn_full,omitempty"`
	ResourceDisplayNames sql.NullString `json:"resource_display_names,omitempty"`
	ShortDescription     sql.NullString `json:"short_description,omitempty"`
	Tags                 sql.NullString `json:"tags,omitempty"`
	RecordRetractionTime sql.NullString `json:"record_retraction_time,omitempty"`
	PnPRemoved           string         `json:"pnp_removed,omitempty"`
	ReleaseNoteUrl       sql.NullString `json:"release_note_url,omitempty"`
}

// NotificationGetNull - Used by GetNotificationByQuery in case it's a funky record
type NotificationGetNull struct {
	RecordID             sql.NullString `json:"record_id,omitempty"`
	PnpCreationTime      sql.NullString `json:"pnp_creation_time,omitempty"`
	PnpUpdateTime        sql.NullString `json:"pnp_update_time,omitempty"`
	SourceCreationTime   sql.NullString `json:"source_creation_time,omitempty"`
	SourceUpdateTime     sql.NullString `json:"source_update_time,omitempty"`
	EventTimeStart       sql.NullString `json:"event_time_start,omitempty"`
	EventTimeEnd         sql.NullString `json:"event_time_end,omitempty"`
	Source               sql.NullString `json:"source,omitempty"`
	SourceID             sql.NullString `json:"source_id,omitempty"`
	Type                 sql.NullString `json:"type,omitempty"`
	Category             sql.NullString `json:"category,omitempty"`
	IncidentID           sql.NullString `json:"incident_id,omitempty"`
	CRNFull              sql.NullString `json:"crn_full,omitempty"`
	ResourceDisplayNames sql.NullString `json:"resource_display_names,omitempty"`
	ShortDescription     sql.NullString `json:"short_description,omitempty"`
	Tags                 sql.NullString `json:"tags,omitempty"`
	RecordRetractionTime sql.NullString `json:"record_retraction_time,omitempty"`
	PnPRemoved           sql.NullString `json:"pnp_removed,omitempty"`
	ReleaseNoteUrl       sql.NullString `json:"release_note_url,omitempty"`
}

// NotificationReturn - the vales once retieved into NotificationGet will be converted into NotificationReturn to display
type NotificationReturn struct {
	RecordID             string        `json:"record_id,omitempty"`
	PnpCreationTime      string        `json:"pnp_creation_time,omitempty"`
	PnpUpdateTime        string        `json:"pnp_update_time,omitempty"`
	SourceCreationTime   string        `json:"source_creation_time,omitempty"`
	SourceUpdateTime     string        `json:"source_update_time,omitempty"`
	EventTimeStart       string        `json:"event_time_start,omitempty"`
	EventTimeEnd         string        `json:"event_time_end,omitempty"`
	Source               string        `json:"source,omitempty"`
	SourceID             string        `json:"source_id,omitempty"`
	Type                 string        `json:"type,omitempty"`
	Category             string        `json:"category,omitempty"`
	IncidentID           string        `json:"incident_id,omitempty"`
	CRNFull              string        `json:"crn_full,omitempty"`
	ResourceDisplayNames []DisplayName `json:"resource_display_names,omitempty"`
	ShortDescription     []DisplayName `json:"short_description,omitempty"`
	LongDescription      []DisplayName `json:"long_description,omitempty"`
	Tags                 string        `json:"tags,omitempty"`
	RecordRetractionTime string        `json:"record_retraction_time,omitempty"`
	PnPRemoved           bool          `json:"pnp_removed,omitempty"`
	ReleaseNoteUrl       string        `json:"release_note_url,omitempty"`
}

const (
	// BulkLoad is a constant for the bulk load message type
	BulkLoad = NotificationMsgType("bulkload")
	// Update is a constant for the message type representing an update
	Update = NotificationMsgType("update")
)

// NotificationMsgType represents the types that are valid for notification messages
type NotificationMsgType string

// NotificationMsg Used to pass messages along MQ
type NotificationMsg struct {
	MsgType               NotificationMsgType `json:"msgtype"`
	IsPrimary             bool                `json:"is_primary,omitempty"`
	MaintenanceDuration   int                 `json:"maintenance_duration,omitempty"`
	DisruptionDuration    int                 `json:"disruption_duration,omitempty"`
	DisruptionType        string              `json:"disruption_type,omitempty"`
	DisruptionDescription string              `json:"disruption_description,omitempty"`
	NotificationInsert
}

// IsBulkLoad will return true if the message represents a bulk load
func (msg *NotificationMsg) IsBulkLoad() bool {
	// Important note that bulk load MsgType can also be the empty string
	// therefore, anything that is not an update is assumed to be bulk load
	return msg.MsgType != Update
}

// IsUpdate will return true if the message represents an update
func (msg *NotificationMsg) IsUpdate() bool {
	return msg.MsgType == Update
}

// WrapNotification will take as input a NotificationInsert and wrap it in a message suitable for sending on MQ
func WrapNotification(ni *NotificationInsert, msgType NotificationMsgType) (*NotificationMsg, error) {
	buffer := new(bytes.Buffer)

	err := json.NewEncoder(buffer).Encode(ni)
	if err != nil {
		log.Println(tlog.Log()+consts.ErrJSONEncode, ni, "\n", err)
		return nil, err

	}
	wrap := new(NotificationMsg)
	err = json.NewDecoder(bytes.NewReader(buffer.Bytes())).Decode(wrap)
	if err != nil {
		log.Println(tlog.Log()+consts.ErrJSONDecode, ni, "\n", err)
	}

	wrap.MsgType = msgType

	return wrap, err
}

// UnwrapNotification will take as input a wrapped NotificationInsert and return just the NotificationInsert
func UnwrapNotification(nm *NotificationMsg) (*NotificationInsert, error) {

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(nm)
	if err != nil {
		log.Println(tlog.Log()+consts.ErrJSONEncode, nm, "\n", err)
		return nil, err
	}

	ni := new(NotificationInsert)
	err = json.NewDecoder(bytes.NewReader(buffer.Bytes())).Decode(ni)
	if err != nil {
		log.Println(tlog.Log()+consts.ErrJSONDecode, ni, "\n", err)
	}

	return ni, err
}
