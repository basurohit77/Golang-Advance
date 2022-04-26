package status

import (
	"strings"

	"github.ibm.com/cloud-sre/pnp-rest-test/commontest"
	"github.ibm.com/cloud-sre/pnp-rest-test/lg"
	"github.ibm.com/cloud-sre/pnp-rest-test/testutils"
)

// NotificationTest runs the notification test
func (api *API) NotificationTest() {

	lg.Info("NotificationTest", "Executing notification tests")

	api.AllNotificationTest()

}

// AllNotificationTest gets all notifications
func (api *API) AllNotificationTest() {

	METHOD := "AllNotificationTest"
	lg.Info("AllNotificationTest", "Querying all notifications test")

	list := new(NotificationList)
	err := api.cat.Server.GetAndDecode(METHOD, "notification.NotificationList", api.apiInfo.Notifications.Href, list)
	if err != nil {
		return
	}

	err = commontest.CheckPagination(METHOD, "notification.NotificationList", api.cat.Server, api.apiInfo.Notifications.Href)
	if err != nil {
		return
	}

	api.checkNotificationList(METHOD, list)
}

func (api *API) checkNotificationList(fct string, list *NotificationList) {

	METHOD := fct + "->checkNotificationList"

	if len(list.Resources) == 0 {
		lg.Err(METHOD, nil, "No notifications returned in query.")
		return
	}

	expected := 200
	if list.Limit > expected {
		lg.Err(METHOD, nil, "Limit for notifications is greater than expected maximum. Is %d, should be %d.", list.Limit, expected)
	}

	checkList := new(NotificationList)
	err := api.cat.Server.GetAndDecode(METHOD, "notification.NotificationList", list.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Href on notification list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, err, "Did not get a match on notification list length using href")
	}

	checkList = new(NotificationList)
	err = api.cat.Server.GetAndDecode(METHOD, "notification.NotificationList", list.First.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get First on notification list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, err, "Did not get a match on notification list length using first")
	}

	checkList = new(NotificationList)
	err = api.cat.Server.GetAndDecode(METHOD, "notification.NotificationList", list.Last.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Last on notification list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, nil, "Did not get a match on notification list length using last original(%d) != last(%d)", list.Count, checkList.Count)
	}

	for _, item := range list.Resources {
		api.checkNotification(METHOD, item)
	}
}

func (api *API) checkNotification(fct string, inc NotificationGet) {
	METHOD := fct + "->checkNotification"

	if inc.RecordID == "" {
		lg.Err(METHOD, nil, "No record ID in the notification")
	}

	testutils.CheckTime(METHOD, "Notification.CreationTime", inc.CreationTime)
	testutils.CheckTime(METHOD, "Notification.UpdateTime", inc.UpdateTime)
	testutils.CheckTime(METHOD, "Notification.EventTimeStart", inc.EventTimeStart)

	if strings.TrimSpace(inc.EventTimeEnd) != "" {
		testutils.CheckTime(METHOD, "Notification.EventTimeEnd", inc.EventTimeEnd)
	}

	testutils.CheckEnum(METHOD, "Notification.Kind", inc.Kind, "notification")
	testutils.CheckEnum(METHOD, "Notification.Type", inc.Type, "announcement", "security")

	if len(inc.CRNMasks) == 0 {
		lg.Err(METHOD, nil, "CRNMasks is empty")
	}

	testutils.CheckEnum(METHOD, "Notification.Category", inc.Category, "runtimes", "services", "platform")

	if len(inc.ResourceDisplayName) == 0 {
		lg.Err(METHOD, nil, "Display name in notification is empty")
	}
	if len(inc.LongDescription) == 0 {
		lg.Err(METHOD, nil, "Long description in notification is empty")
	}
	if len(inc.ShortDescription) == 0 {
		lg.Err(METHOD, nil, "Short description in notification is empty")
	}

	checkNotification := new(NotificationGet)
	err := api.cat.Server.GetAndDecode(METHOD, "notification.Notification", inc.Href, checkNotification)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Href on notification.")
	}
	if checkNotification.RecordID != inc.RecordID {
		lg.Err(METHOD, err, "Href notification does not match original")
	}
}
