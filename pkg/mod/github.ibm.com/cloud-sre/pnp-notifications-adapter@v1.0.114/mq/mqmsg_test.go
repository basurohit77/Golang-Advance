package mq

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

func TestWrap(t *testing.T) {

	ni := &datastore.NotificationInsert{
		SourceCreationTime:   "2018-10-31T12:13:14Z",
		SourceUpdateTime:     "2018-10-31T12:13:14Z",
		EventTimeStart:       "2018-10-31T13:13:14Z",
		EventTimeEnd:         "2018-10-31T14:13:14Z",
		Source:               "test",
		SourceID:             "123",
		Type:                 "security",
		Category:             "services",
		IncidentID:           "INC0002",
		CRNFull:              "crn:v1:bluemix:public:service1:us-south::::",
		ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "DisplayName"}},
		ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "ShortDescription"}},
		LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "LongDescription"}},
	}

	msg, err := datastore.WrapNotification(ni, datastore.BulkLoad)
	if err != nil {
		t.Fatal(err)
	}

	if msg.MsgType != datastore.BulkLoad {
		t.Fatal("Wrong msg type")
	}

	if msg.SourceCreationTime != ni.SourceCreationTime {
		t.Fatal("Wrong SourceCreationTime")
	}

	if msg.SourceUpdateTime != ni.SourceUpdateTime {
		t.Fatal("Wrong SourceUpdateTime")
	}

	if msg.SourceID != ni.SourceID {
		t.Fatal("Wrong SourceID")
	}

	if !msg.IsBulkLoad() {
		t.Fatal("Should be bulk load")
	}
	msg.MsgType = ""
	if !msg.IsBulkLoad() {
		t.Fatal("Empty string should be bulk load")
	}

	msg.MsgType = datastore.Update
	if !msg.IsUpdate() {
		t.Fatal("Should be an update")
	}
	if msg.IsBulkLoad() {
		t.Fatal("Should not be bulk load")
	}

	msg2, err := UnwrapNotification(msg)
	if err != nil {
		t.Fatal(err)
	}

	if msg2.SourceCreationTime != ni.SourceCreationTime {
		t.Fatal("Wrong SourceCreationTime")
	}

	if msg2.SourceUpdateTime != ni.SourceUpdateTime {
		t.Fatal("Wrong SourceUpdateTime")
	}

	if msg2.SourceID != ni.SourceID {
		t.Fatal("Wrong SourceID")
	}

}

func TestUnwrap(t *testing.T) {

	ni := &datastore.NotificationInsert{
		SourceCreationTime:   "2018-10-31T12:13:14Z",
		SourceUpdateTime:     "2018-10-31T12:13:14Z",
		EventTimeStart:       "2018-10-31T13:13:14Z",
		EventTimeEnd:         "2018-10-31T14:13:14Z",
		Source:               "test",
		SourceID:             "123",
		Type:                 "security",
		Category:             "services",
		IncidentID:           "INC0002",
		CRNFull:              "crn:v1:bluemix:public:service1:us-south::::",
		ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "DisplayName"}},
		ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "ShortDescription"}},
		LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "LongDescription"}},
	}

	msg, err := WrapNotification(ni, datastore.BulkLoad)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure that we can decode as just a NotificationInsert
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(msg); err != nil {
		t.Fatal(err)
	}

	nm := new(datastore.NotificationInsert)
	if err := json.NewDecoder(bytes.NewReader(buffer.Bytes())).Decode(nm); err != nil {
		t.Fatal(err)
	}

	if nm.EventTimeStart != ni.EventTimeStart {
		t.Fatal("conversion failed")
	}

}

func TestWrapError(t *testing.T) {
	WrapNotification(nil, datastore.BulkLoad)
	UnwrapNotification(nil)
}
