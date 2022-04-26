package mq

import (
	"bytes"
	"encoding/json"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

const (
	// BulkLoad is a constant for the bulk load message type
	BulkLoad = NotificationMsgType("bulkload")
	// Update is a constant for the message type representing an update
	Update = NotificationMsgType("update")
)

// NotificationMsg wraps a notification insert with meta data needed for traveling on MQ
type NotificationMsg struct {
	MsgType               NotificationMsgType `json:"msgtype"`
	MaintenanceDuration   int                 `json:"maintenance_duration,omitempty"`
	DisruptionDuration    int                 `json:"disruption_duration,omitempty"`
	DisruptionType        string              `json:"disruption_type,omitempty"`
	DisruptionDescription string              `json:"disruption_description,omitempty"`
	datastore.NotificationInsert
}

// NotificationMsgType represents the types that are valid for notification messages
type NotificationMsgType string

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
func WrapNotification(ni *datastore.NotificationInsert, msgType datastore.NotificationMsgType) (*datastore.NotificationMsg, error) {
	buffer := new(bytes.Buffer)

	json.NewEncoder(buffer).Encode(ni)

	wrap := new(datastore.NotificationMsg)
	json.NewDecoder(bytes.NewReader(buffer.Bytes())).Decode(wrap)

	wrap.MsgType = msgType

	return wrap, nil
}

// UnwrapNotification will take as input a wrapped NotificationInsert and return just the NotificationInsert
func UnwrapNotification(nm *datastore.NotificationMsg) (*datastore.NotificationInsert, error) {

	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(nm)

	ni := new(datastore.NotificationInsert)
	json.NewDecoder(bytes.NewReader(buffer.Bytes())).Decode(ni)

	return ni, nil
}
