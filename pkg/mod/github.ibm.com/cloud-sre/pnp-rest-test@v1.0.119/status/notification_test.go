package status

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/pnp-rest-test/catalog"
	"github.ibm.com/cloud-sre/pnp-rest-test/common"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
	"github.ibm.com/cloud-sre/pnp-rest-test/testutils"
)

var notification = NotificationGet{
	//Href:             tsSingle,
	RecordID:            "RECORDID123",
	CreationTime:        testutils.GetDate(time.Hour),
	UpdateTime:          testutils.GetDate(time.Hour),
	EventTimeStart:      testutils.GetDate(time.Hour),
	EventTimeEnd:        testutils.GetDate(time.Hour),
	Kind:                "notification",
	Type:                "security",
	Category:            "services",
	ShortDescription:    []common.TranslatedString{common.TranslatedString{Language: "en", Text: "Short Description PDQ"}},
	LongDescription:     []common.TranslatedString{common.TranslatedString{Language: "en", Text: "Long Description PDQ"}},
	ResourceDisplayName: []common.TranslatedString{common.TranslatedString{Language: "en", Text: "Cloudant PDQ"}},
	CRNMasks:            []string{"crn:v1:bluemix:public:cloudant:::::"},
}

var notificationList *NotificationList

func TestNotifications(t *testing.T) {

	tsNotification := testutils.NewDataServer(serveNotification)
	notification.Href = tsNotification.URL
	defer tsNotification.Close()

	tsList := testutils.NewDataServer(serveNoteList)
	defer tsList.Close()

	list := new(NotificationList)
	list.Offset = 0
	list.Limit = 1
	list.Count = 1
	list.Href = tsList.URL
	list.First.Href = tsList.URL
	list.Last.Href = tsList.URL

	list.Resources = append(list.Resources, notification)
	notificationList = list

	srvr := testutils.NewJSONServer(list)
	defer srvr.Close()

	apiInfo := new(APIInfo)
	apiInfo.Notifications.Href = srvr.URL

	api := &API{apiInfo: apiInfo, cat: &catalog.Catalog{Server: &rest.Server{}}}

	api.AllNotificationTest()

}

func serveNoteList(url *url.URL) ([]byte, int) {
	var err error
	buffer := new(bytes.Buffer)
	if err = json.NewEncoder(buffer).Encode(notificationList); err != nil {
		return nil, http.StatusInternalServerError
	}
	return buffer.Bytes(), http.StatusOK
}

func serveNotification(url *url.URL) ([]byte, int) {

	var err error
	buffer := new(bytes.Buffer)
	if err = json.NewEncoder(buffer).Encode(notification); err != nil {
		return nil, http.StatusInternalServerError
	}
	return buffer.Bytes(), http.StatusOK
}
