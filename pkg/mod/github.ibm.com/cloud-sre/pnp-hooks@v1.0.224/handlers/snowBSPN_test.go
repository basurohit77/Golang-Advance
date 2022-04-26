package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	ossadapter "github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	adapter "github.ibm.com/cloud-sre/pnp-notifications-adapter/api"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/globalcatalog"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/postgres"
)

const (
	cloudantCategoryID   = "cloudoe.sop.enum.paratureCategory.literal.l119"
	dialogCategoryID     = "cloudoe.sop.enum.paratureCategory.literal.l200"
	containersCategoryID = "cloudoe.sop.enum.paratureCategory.literal.l185"
	cloudantnosqldb      = "cloudantnosqldb"
	containers           = "containers-kubernetes"
)

func TestTimestamp(t *testing.T) {

	tstamp := "2018-10-31T12:13:14.000Z"
	myTime, err := time.Parse(time.RFC3339, tstamp)
	if err != nil {
		t.Fatal(err)
	}

	ms := myTime.UnixNano() / 1000000

	out := msToNoteTime(ms)

	if out != tstamp {
		t.Fatalf("Times do not match. %s != %s", tstamp, out)
	}

}

func TestBuildCRN(t *testing.T) {
	inputCRN := "crn:v1:d-aa1:dedicated:velocity_test:us-east::::"
	serviceName := "platform"
	outputCRN := "crn:v1:d-aa1:dedicated:platform:us-east::::"

	checkString(t, "buildCRN", buildCRN(inputCRN, serviceName), outputCRN)
}

func TestParseBSPN(t *testing.T) {

	msg := `{"type":"notification","regions":["ibm:yp:eu-gb"],"start_date":1536748020000,"end_date":1536764700000,"title":"RESOLVED: Issues with App Connect service","description":"<p>SERVICES/COMPONENTS AFFECTED:<br />&nbsp; - App Connect</p>\n<p><br />IMPACT:<br />&nbsp; - Users may experience problems running event-driven flows. Users will notice that their flows are not running. The flows will show an error state in the UI.<br />&nbsp; - Pure API-driven flows are not impacted.</p>\n<p><br />STATUS:<br />&nbsp; - 2018-09-12 11:00 UTC - INVESTIGATING - The operations team is aware of the issues and is currently investigating.<br />&nbsp; - 2018-09-12 15:25 UTC - RESOLVED - The issues have been resolved as of 15:05 UTC.</p>\n<p>&nbsp;</p>","severity":"Sev - 1","components":["cloudoe.sop.enum.paratureCategory.literal.l377"],"id":"BSP0002135","parent_id":"INC0279855","status":"CIE In Progress","affected_activities":"Application Availability","modified":1536765943588,"crn":["crn:v1:d-aa2:dedicated:velocity_test:us-south::::","crn:v1:d-aa1:dedicated:velocity_test:us-east::::","crn:v1:d-alfaevolution:dedicated:velocity_test:mil01::::"]}`

	b, err := parseBSPN([]byte(msg))
	if err != nil {
		t.Fatal(err)
	}

	checkString(t, "notification.Type", b.Type, "notification")
	checkString(t, "notification.Title", b.Title, "RESOLVED: Issues with App Connect service")
	checkString(t, "notification.Description", b.Description, "<p>SERVICES/COMPONENTS AFFECTED:<br />&nbsp; - App Connect</p>\n<p><br />IMPACT:<br />&nbsp; - Users may experience problems running event-driven flows. Users will notice that their flows are not running. The flows will show an error state in the UI.<br />&nbsp; - Pure API-driven flows are not impacted.</p>\n<p><br />STATUS:<br />&nbsp; - 2018-09-12 11:00 UTC - INVESTIGATING - The operations team is aware of the issues and is currently investigating.<br />&nbsp; - 2018-09-12 15:25 UTC - RESOLVED - The issues have been resolved as of 15:05 UTC.</p>\n<p>&nbsp;</p>")
	checkString(t, "notification.Severity", b.Severity, "Sev - 1")
	checkString(t, "notification.ID", b.ID, "BSP0002135")
	checkString(t, "notification.ParentID", b.ParentID, "INC0279855")
	checkString(t, "notification.Status", b.Status, "CIE In Progress")
	checkString(t, "notification.AffectedActivities", b.AffectedActivities, "Application Availability")

	checkStringArray(t, "notification.Regions", b.Regions, []string{"ibm:yp:eu-gb"})
	checkStringArray(t, "notification.Components", b.Components, []string{"cloudoe.sop.enum.paratureCategory.literal.l377"})
	checkStringArray(t, "notification.CRN", b.CRN, []string{"crn:v1:d-aa2:dedicated:velocity_test:us-south::::", "crn:v1:d-aa1:dedicated:velocity_test:us-east::::", "crn:v1:d-alfaevolution:dedicated:velocity_test:mil01::::"})

	checkInt64(t, "notification.StartDate", b.StartDate, 1536748020000)
	checkInt64(t, "notification.EndDate", b.EndDate, 1536764700000)
	checkInt64(t, "notification.Modified", b.Modified, 1536765943588)

}

func checkInt64(t *testing.T, label string, is, shouldbe int64) {
	if is != shouldbe {
		t.Errorf("%s mismatch. Should be [%s] is [%s]", label, strconv.FormatInt(shouldbe, 10), strconv.FormatInt(is, 10))
	}
}

func checkStringArray(t *testing.T, label string, is, shouldbe []string) {
	for i := 0; i < len(shouldbe); i++ {
		if is[i] != shouldbe[i] {
			t.Errorf("%s mismatch. Item %d should be [%s] is [%s]", label, i, shouldbe[i], is[i])
		}
	}
}

func checkString(t *testing.T, label, is, shouldbe string) {
	if is != shouldbe {
		t.Errorf("%s mismatch. Should be [%s] is [%s]", label, shouldbe, is)
	}
}

func setupCache(t *testing.T) {
	ossadapter.NewCache(ctxt.Context{}, myListOSSFunction) //(*OSSRecordCache, error)
}

func myListOSSFunction(rp *regexp.Regexp, include catalog.IncludeOptions, sendFunc func(r ossrecord.OSSEntry)) error {

	// Create Cloudant *******************************
	r := new(ossrecord.OSSService)
	r.ReferenceResourceName = cloudantnosqldb
	r.ReferenceDisplayName = "Cloudant"
	r.GeneralInfo.EntryType = ossrecord.SERVICE
	r.GeneralInfo.OSSTags = append(r.GeneralInfo.OSSTags, osstags.PnPEnabled)
	r.StatusPage.CategoryID = cloudantCategoryID
	sendFunc(r)

	// Create Containers *******************************
	r = new(ossrecord.OSSService)
	r.ReferenceResourceName = containers
	r.ReferenceDisplayName = "Kubernetes Service"
	r.GeneralInfo.EntryType = ossrecord.IAAS
	r.GeneralInfo.OSSTags = append(r.GeneralInfo.OSSTags, osstags.PnPEnabled)
	r.StatusPage.CategoryID = containersCategoryID
	sendFunc(r)

	// Create Non Compliant Service Watson Dialog *******************************
	r = new(ossrecord.OSSService)
	r.ReferenceResourceName = "dialog"
	r.ReferenceDisplayName = "Watson Dialog"
	r.GeneralInfo.EntryType = ossrecord.SERVICE
	r.GeneralInfo.OSSTags = append(r.GeneralInfo.OSSTags, osstags.StatusYellow)
	r.StatusPage.CategoryID = dialogCategoryID
	sendFunc(r)

	return nil
}

func TestDisplayNames(t *testing.T) {
	setupCache(t)
	postgres.SetupDBFunctionsForUT(utConnect, utDisconnect, utIsActive, utGetNotificationByQuery, utGetNotificationByRecordID, utInsertNotification, utUpdateNotification)
	ts := setupCreds(t)
	defer ts.Close()

	names, err := getDisplayNames(ctxt.Context{}, cloudantCategoryID, cloudantnosqldb) //([]datastore.DisplayName, error) {
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 2 {
		t.Error("did not get expected number of display names", len(names))
		for _, n := range names {
			t.Error(n.Name)
		}
	}
	for _, name := range names {

		switch name.Language {
		case "en":
			if name.Name != "CloudantGC" {
				t.Fatal("got unexpected en value", name.Name)
			}
		case "de":
			if name.Name != "CloudantGC_de" {
				t.Fatal("got unexpected de value", name.Name)
			}
		default:
			t.Fatal("got unexpected language", name.Language)
		}

	}

	names, err = getDisplayNames(ctxt.Context{}, containersCategoryID, containers)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 1 {
		t.Error("did not get expected number of display names for containers", len(names))
		for _, n := range names {
			t.Error(n.Name)
		}
	}

	if names[0].Name != "Kubernetes Service" {
		t.Fatal("Did not get correct service name", names[0])
	}
}

func setupCreds(t *testing.T) *httptest.Server {
	creds = new(adapter.SourceConfig)
	ts := setupGlobalCatalog(t)
	creds.GlobalCatalog.ResourcesURL = ts.URL
	return ts
}

// mocks db connection
func utConnect(host string, port int, databaseName string, user string, password string, sslmode string) (database *sql.DB, err error) {
	return nil, nil
}

// mocks disconnect
func utDisconnect(database *sql.DB) {
}

// mocks check of active db
func utIsActive(database *sql.DB) (err error) {
	return nil
}

// mocks getNotificationsbyQuery
func utGetNotificationByQuery(database *sql.DB, query string, limit int, offset int) (*[]datastore.NotificationReturn, int, error, int) {
	log.Println("utGetNotificationByQuery", query)
	return nil, 0, nil, 200
}

// utGetNotificationByRecordID mocks query by ID
func utGetNotificationByRecordID(database *sql.DB, recordID string) (*datastore.NotificationReturn, error, int) {
	log.Println("utGetNotificationByRecordID", recordID)
	return nil, nil, 200
}

// DBInsertNotificationFunc is mockable for unit test purposes
func utInsertNotification(database *sql.DB, itemToInsert *datastore.NotificationInsert) (string, error, int) {
	return "", nil, 200
}

// DBUpdateNotificationFunc is mockable for unit test purposes
func utUpdateNotification(database *sql.DB, itemToInsert *datastore.NotificationInsert) (error, int) {
	return nil, 200
}

// createNotificationReturn will create a mocked up notification return
func createNotificationReturn(source, sourceID, nType, nCategory, incidentID, crn string, cPnpTime, uPnpTime, cSrcTime, uSrcTime, sEvtTime, eEvtTime time.Duration, resourceName, shortDesc, longDesc string) datastore.NotificationReturn {

	result := &datastore.NotificationReturn{Source: source, SourceID: sourceID, Type: nType, Category: nCategory, IncidentID: incidentID, CRNFull: crn}

	result.RecordID = fmt.Sprintf("%s+%s+%s", source, sourceID, crn)

	result.PnpCreationTime = time.Now().Add(cPnpTime * -1).Format("2006-01-02T15:04:05:000Z")
	result.PnpUpdateTime = time.Now().Add(uPnpTime * -1).Format("2006-01-02T15:04:05:000Z")
	result.SourceCreationTime = time.Now().Add(cSrcTime * -1).Format("2006-01-02T15:04:05:000Z")
	result.SourceUpdateTime = time.Now().Add(uSrcTime * -1).Format("2006-01-02T15:04:05:000Z")
	result.EventTimeStart = time.Now().Add(sEvtTime * -1).Format("2006-01-02T15:04:05:000Z")
	result.EventTimeEnd = time.Now().Add(eEvtTime * -1).Format("2006-01-02T15:04:05:000Z")

	result.ResourceDisplayNames = append(result.ResourceDisplayNames, datastore.DisplayName{Name: resourceName, Language: "en"})
	result.ShortDescription = append(result.ShortDescription, datastore.DisplayName{Name: shortDesc, Language: "en"})
	result.LongDescription = append(result.LongDescription, datastore.DisplayName{Name: longDesc, Language: "en"})

	return *result
}

func setupGlobalCatalog(t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		url := strings.TrimSpace(r.URL.String())
		log.Println("INFO: Global Catalog request", url)

		w.Header().Set("Content-Type", "application/json")

		var data *globalcatalog.CloudResourcesList
		if url == "/" {
			data = getGlobalCatalogPage1()
		} else {
			w.WriteHeader(400)
			return
		}

		buffer := new(bytes.Buffer)
		if err := json.NewEncoder(buffer).Encode(data); err != nil {
			t.Fatalf("cannot encode notification [%s]", err.Error())
		}

		_, err := w.Write(buffer.Bytes())
		if err != nil {
			t.Fatal("Error reading file:", err)
		}
	}))

	return ts
}

func getGlobalCatalogPage1() *globalcatalog.CloudResourcesList {

	list := make([]*globalcatalog.CloudResource, 0, 2)

	//--------------------------------------
	r := new(globalcatalog.CloudResource)
	r.Active = true
	r.Name = cloudantnosqldb
	r.Tags = append(r.Tags, "unittest")
	r.OverviewUI = json.RawMessage(`{"en":{"description":"Cloudant desc","display_name":"CloudantGC","long_description":"Cloudant long desc"},"de":{"description":"Cloudant desc_de","display_name":"CloudantGC_de","long_description":"Cloudant long desc_de"}}`)
	list = append(list, r)

	result := new(globalcatalog.CloudResourcesList)
	result.Limit = 200
	result.ResourceCount = len(list)
	result.Count = len(list)
	result.Resources = list

	return result
}

func TestBSPNToNoteInsert(t *testing.T) {
	setupCache(t)
	postgres.SetupDBFunctionsForUT(utConnect, utDisconnect, utIsActive, utGetNotificationByQuery, utGetNotificationByRecordID, utInsertNotification, utUpdateNotification)
	ts := setupCreds(t)
	defer ts.Close()

	b := &BSPN{
		Type:               "notification",
		Regions:            []string{"ibm:yp:us-south", "ibm:yp:us-east"},
		StartDate:          time.Now().Add(time.Hour*-1).UnixNano() / 1000000,
		EndDate:            time.Now().UnixNano() / 1000000,
		Title:              "Fake News",
		Description:        "Fake Description",
		Severity:           "Sev 1",
		Components:         []string{cloudantCategoryID},
		ID:                 "BSPN12345",
		ParentID:           "INC12345",
		Status:             "CIE In Progress",
		AffectedActivities: "Sleep",
		Modified:           time.Now().UnixNano() / 1000000,
		CRN:                []string{"crn:v1:bluemix:public:" + cloudantnosqldb + ":us-south::::", "crn:v1:bluemix:public:" + cloudantnosqldb + ":us-east::::"},
	}

	list, err := BSPNToNoteInsert(ctxt.Context{}, b)
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 2 {
		t.Fatal("Did not get the right conversion from input BSPN. Expected 2 items, got", len(list))
	}

	south := false
	east := false

	for _, n := range list {

		checkString(t, "n.Source", n.Source, "servicenow")
		checkString(t, "n.SourceID", n.SourceID, "BSPN12345")
		checkString(t, "n.Type", n.Type, "incident")
		checkString(t, "n.Category", n.Category, "services")

		if n.CRNFull == "crn:v1:bluemix:public:"+cloudantnosqldb+":us-south::::" {
			south = true
		}
		if n.CRNFull == "crn:v1:bluemix:public:"+cloudantnosqldb+":us-east::::" {
			east = true
		}

		enCheck := false
		deCheck := false

		for _, rd := range n.ResourceDisplayNames {
			switch rd.Language {
			case "en":
				if rd.Name == "CloudantGC" {
					enCheck = true
				}
			case "de":
				if rd.Name == "CloudantGC_de" {
					deCheck = true
				}
			default:
				t.Fatal("Got unexpected language")
			}
		}
		if !enCheck || !deCheck {
			t.Fatal("Wrong display names", n.ResourceDisplayNames)
		}

	}

	if !south || !east {
		t.Fatal("didn't find all CRNs", south, east)
	}

}

func TestTransformation(t *testing.T) {
	setupCache(t)
	osscatalog.CatalogCheckBypass = true // disable the catalog check for now
	postgres.SetupDBFunctionsForUT(utConnect, utDisconnect, utIsActive, utGetNotificationByQuery, utGetNotificationByRecordID, utInsertNotification, utUpdateNotification)
	ts := setupCreds(t)
	defer ts.Close()

	b := &BSPN{
		Type:               "notification",
		Regions:            []string{"ibm:yp:us-south", "ibm:yp:us-east"},
		StartDate:          time.Now().Add(time.Hour*-1).UnixNano() / 1000000,
		EndDate:            time.Now().UnixNano() / 1000000,
		Title:              "Fake News",
		Description:        "Fake Description",
		Severity:           "Sev 1",
		Components:         []string{cloudantCategoryID},
		ID:                 "BSPN12345",
		ParentID:           "INC12345",
		Status:             "CIE In Progress",
		AffectedActivities: "Sleep",
		Modified:           time.Now().UnixNano() / 1000000,
		CRN:                []string{"crn:v1:bluemix:public:" + cloudantnosqldb + ":us-south::::", "crn:v1:bluemix:public:" + cloudantnosqldb + ":us-east::::"},
	}

	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(b); err != nil {
		t.Fatalf("cannot encode notification [%s]", err.Error())
	}

	req := httptest.NewRequest("POST", "/notification", bytes.NewReader(buffer.Bytes()))

	bResult, err := transformMessage(ctxt.Context{}, req) //(result [][]byte, err error) {
	if err != nil {
		t.Fatal(err)
	}

	if len(bResult) != 2 {
		t.Fatal("Did not get the right conversion from input BSPN. Expected 2 items, got", len(bResult))
	}

	list := make([]*datastore.NotificationInsert, 0, 2)
	for _, p := range bResult {

		ni := new(datastore.NotificationInsert)
		if err := json.NewDecoder(bytes.NewReader(p)).Decode(ni); err != nil {
			t.Fatalf("ERROR: Failed to decode the NotificationInsert: %s", err.Error())

		}

		list = append(list, ni)
	}

	south := false
	east := false

	for _, n := range list {

		checkString(t, "n.Source", n.Source, "servicenow")
		checkString(t, "n.SourceID", n.SourceID, "BSPN12345")
		checkString(t, "n.Type", n.Type, "incident")
		checkString(t, "n.Category", n.Category, "services")

		if n.CRNFull == "crn:v1:bluemix:public:"+cloudantnosqldb+":us-south::::" {
			south = true
		}
		if n.CRNFull == "crn:v1:bluemix:public:"+cloudantnosqldb+":us-east::::" {
			east = true
		}

		enCheck := false
		deCheck := false

		for _, rd := range n.ResourceDisplayNames {
			switch rd.Language {
			case "en":
				if rd.Name == "CloudantGC" {
					enCheck = true
				}
			case "de":
				if rd.Name == "CloudantGC_de" {
					deCheck = true
				}
			default:
				t.Fatal("Got unexpected language")
			}
		}
		if !enCheck || !deCheck {
			t.Fatal("Wrong display names", n.ResourceDisplayNames)
		}
	}

	if !south || !east {
		t.Fatal("didn't find all CRNs", south, east)
	}

}
