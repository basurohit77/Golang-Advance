package mq

import (
	"testing"

	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

func TestSendNotes(t *testing.T) {

	ni := datastore.NotificationInsert{
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

	type IConnection interface {
		Produce(msg string) error
		SendNotifications(nList []datastore.NotificationInsert, msgType NotificationMsgType) (err error)
		SendCompareAndUpdateNotifications(existingList, newList []datastore.NotificationInsert) error
		Close() error
	}
	nList := []datastore.NotificationInsert{ni}
	err := internalSendNotifications(ctxt.Context{}, &utTestConn{}, nList, "")
	if err != nil {
		t.Fatal(err)
	}
}

type utTestConn struct {
	Name string
}

func (c *utTestConn) Produce(msg string) error {
	return nil
}
func (c *utTestConn) SendNotifications(ctx ctxt.Context, nList []datastore.NotificationInsert, msgType datastore.NotificationMsgType) (err error) {
	return nil
}
func (c *utTestConn) SendCompareAndUpdateNotifications(ctx ctxt.Context, existingList, newList []datastore.NotificationInsert) error {
	return nil
}
func (c *utTestConn) Close() error {
	return nil
}

func TestCompareNotes(t *testing.T) {

	ni := datastore.NotificationInsert{
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

	n2 := ni
	n2.ShortDescription = []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "ShortDescription123"}}

	type IConnection interface {
		Produce(msg string) error
		SendNotifications(nList []datastore.NotificationInsert, msgType NotificationMsgType) (err error)
		SendCompareAndUpdateNotifications(existingList, newList []datastore.NotificationInsert) error
		Close() error
	}
	nList := []datastore.NotificationInsert{ni}
	n2List := []datastore.NotificationInsert{n2}
	err := internalSendCompareAndUpdateNotifications(ctxt.Context{}, &utTestConn{}, nList, n2List)
	if err != nil {
		t.Fatal(err)
	}
}
