package convert

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/api"
)

func TestConvert(t *testing.T) {

	cnList := new(api.NotificationList)

	item1 := new(api.Notification)
	item1.ShortDescription = "This is notification #1"
	item1.LongDescription = "Long description for notification #1"
	item1.CRNs = []string{"crn:v1:bluemix:public:cloudantnosqldb:us-south::::", "crn:v1:bluemix:public:cloudantnosqldb:us-east::::"}
	item1.EventTimeStart = "2018-10-31T12:12:12.000Z"
	item1.EventTimeEnd = "2018-10-31T23:23:23.000Z"
	item1.IncidentID = "INC001001"
	item1.CreationTime = "2018-11-01T01:01:01.000Z"
	item1.UpdateTime = "2018-11-01T02:02:02.000Z"
	item1.CategoryNotificationID = "cloudoe.parature.l19"
	item1.Source = "cloudant"
	item1.SourceID = "123456789012345678901234567890"
	item1.NotificationType = "announcement"
	item1.DisplayName = append(item1.DisplayName, &api.TranslatedString{Language: "en", Text: "Cloudant No SQL Database"})

	cnList.Items = append(cnList.Items, item1)

	item2 := new(api.Notification)
	item2.ShortDescription = "This is notification #2"
	item2.LongDescription = "Long description for notification #2"
	item2.CRNs = []string{"crn:v1:bluemix:public:appid:eu-gb::::"}
	item2.EventTimeStart = "2018-08-31T12:12:12.000Z"
	item2.EventTimeEnd = "2018-08-31T23:23:23.000Z"
	item2.IncidentID = "INC001002"
	item2.CreationTime = "2018-09-01T01:01:01.000Z"
	item2.UpdateTime = "2018-09-01T02:02:02.000Z"
	item2.CategoryNotificationID = "cloudoe.parature.l20"
	item2.Source = "cloudant"
	item2.SourceID = "234567890123456789012345678901"
	item2.NotificationType = "announcement"
	item2.DisplayName = append(item2.DisplayName, &api.TranslatedString{Language: "en", Text: "Application ID"})

	cnList.Items = append(cnList.Items, item2)

	if len(cnList.Items) != 2 {
		t.Fatal("Did not get 2 items. Got", len(cnList.Items))
	}

	oList := CnToPGniList(cnList)

	if len(oList) != 3 {
		t.Fatal("Did not get 3 items. Got", len(oList))
	}

	var eugb, useast, ussouth bool
	for _, p := range oList {

		if p.CRNFull == "crn:v1:bluemix:public:cloudantnosqldb:us-south::::" {
			ussouth = true
			compareNi(t, item1, &p)
		}

		if p.CRNFull == "crn:v1:bluemix:public:cloudantnosqldb:us-east::::" {
			useast = true
			compareNi(t, item1, &p)
		}

		if p.CRNFull == "crn:v1:bluemix:public:appid:eu-gb::::" {
			eugb = true
			compareNi(t, item2, &p)
		}

	}

	if !eugb || !useast || !ussouth {
		fmt.Println("ussouth", ussouth, "useast", useast, "eugb", eugb)
		t.Fatal("Did get every record")
	}

	nrList := make([]datastore.NotificationReturn, 0)

	for _, o := range oList {

		nri := new(datastore.NotificationReturn)

		nri.SourceCreationTime = o.SourceCreationTime
		nri.SourceUpdateTime = o.SourceUpdateTime
		nri.EventTimeStart = o.EventTimeStart
		nri.EventTimeEnd = o.EventTimeEnd
		nri.Source = o.Source
		nri.SourceID = o.SourceID
		nri.Type = o.Type
		nri.Category = o.Category
		nri.IncidentID = o.IncidentID
		nri.CRNFull = o.CRNFull
		nri.ResourceDisplayNames = append(nri.ResourceDisplayNames, o.ResourceDisplayNames...)
		nri.ShortDescription = append(nri.ShortDescription, o.ShortDescription...)
		nri.LongDescription = append(nri.LongDescription, o.LongDescription...)

		//nri.RecordID           =  string         `json:"record_id,omitempty"`
		//nri.PnpCreationTime    =  string         `json:"pnp_creation_time,omitempty"`
		//nri.PnpUpdateTime      =  string         `json:"pnp_update_time,omitempty"`

		nrList = append(nrList, *nri)
	}

	nList := PGnrToCnList(nrList, "")

	if len(nList.Items) != 2 {
		t.Fatal("Got wrong count of nList")
	}

	var got2a, got2b, got1 bool
	for _, p := range nList.Items {
		if len(p.CRNs) == 1 {
			got1 = true
			if p.CRNs[0] != "crn:v1:bluemix:public:appid:eu-gb::::" {
				t.Fatal("Didn't get eu-gb crn")
			}
			compareNr(t, p, item2)
		}
		if len(p.CRNs) == 2 {
			if p.CRNs[0] == "crn:v1:bluemix:public:cloudantnosqldb:us-south::::" {
				got2a = true
			}
			if p.CRNs[1] == "crn:v1:bluemix:public:cloudantnosqldb:us-south::::" {
				got2a = true
			}
			if p.CRNs[0] == "crn:v1:bluemix:public:cloudantnosqldb:us-east::::" {
				got2b = true
			}
			if p.CRNs[1] == "crn:v1:bluemix:public:cloudantnosqldb:us-east::::" {
				got2b = true
			}
			compareNr(t, p, item1)
		}
	}

	if !got1 || !got2a || !got2b {
		t.Fatal("didn't get correct conversion items")
	}
}

func compareNi(t *testing.T, item1 *api.Notification, p *datastore.NotificationInsert) {

	if p.SourceCreationTime != item1.CreationTime {
		t.Fatal("Did not match")
	}
	if p.SourceUpdateTime != item1.UpdateTime {
		t.Fatal("Did not match")
	}
	if p.EventTimeStart != item1.EventTimeStart {
		t.Fatal("Did not match")
	}
	if p.EventTimeEnd != item1.EventTimeEnd {
		t.Fatal("Did not match")
	}
	if p.Source != item1.Source {
		t.Fatal("Did not match")
	}
	if p.SourceID != item1.SourceID {
		t.Fatal("Did not match")
	}
	if p.Type != item1.NotificationType {
		t.Fatal("Did not match")
	}
	if p.Category != item1.Category {
		t.Fatal("Did not match")
	}
	if p.IncidentID != item1.IncidentID {
		t.Fatal("Did not match")
	}

	if p.ResourceDisplayNames[0].Name != item1.DisplayName[0].Text {
		t.Fatal("Did not match")
	}
	if p.ResourceDisplayNames[0].Language != item1.DisplayName[0].Language {
		t.Fatal("Did not match")
	}

	if p.ShortDescription[0].Name != item1.ShortDescription {
		t.Fatal("Did not match")
	}
	if p.ShortDescription[0].Language != "en" {
		t.Fatal("Did not match")
	}

	if p.LongDescription[0].Name != item1.LongDescription {
		t.Fatal("Did not match")
	}
	if p.LongDescription[0].Language != "en" {
		t.Fatal("Did not match")
	}
}

func compareNr(t *testing.T, item1 *api.Notification, p *api.Notification) {

	if p.CreationTime != item1.CreationTime {
		t.Fatal("Did not match")
	}
	if p.UpdateTime != item1.UpdateTime {
		t.Fatal("Did not match")
	}
	if p.EventTimeStart != item1.EventTimeStart {
		t.Fatal("Did not match")
	}
	if p.EventTimeEnd != item1.EventTimeEnd {
		t.Fatal("Did not match")
	}
	if p.Source != item1.Source {
		t.Fatal("Did not match")
	}
	if p.SourceID != item1.SourceID {
		t.Fatal("Did not match")
	}
	if p.NotificationType != item1.NotificationType {
		t.Fatal("Did not match")
	}
	// CategoryNotificationID not stored in postgres
	//if p.CategoryNotificationID != item1.CategoryNotificationID {
	//	t.Fatal("Did not match " + p.CategoryNotificationID + " != " + item1.CategoryNotificationID)
	//}
	if p.IncidentID != item1.IncidentID {
		t.Fatal("Did not match")
	}

	if p.DisplayName[0].Text != item1.DisplayName[0].Text {
		t.Fatal("Did not match")
	}
	if p.DisplayName[0].Language != item1.DisplayName[0].Language {
		t.Fatal("Did not match")
	}

	if p.ShortDescription != item1.ShortDescription {
		t.Fatal("Did not match")
	}

	if p.LongDescription != item1.LongDescription {
		t.Fatal("Did not match")
	}
}

/*
REFERENCE: NotificationReturn struct {
	RecordID             string         `json:"record_id,omitempty"`
	PnpCreationTime      string         `json:"pnp_creation_time,omitempty"`
	PnpUpdateTime        string         `json:"pnp_update_time,omitempty"`
	SourceCreationTime   string         `json:"source_creation_time,omitempty"`
	SourceUpdateTime     string         `json:"source_update_time,omitempty"`
	EventTimeStart       string         `json:"event_time_start,omitempty"`
	EventTimeEnd         string         `json:"event_time_end,omitempty"`
	Source               string         `json:"source,omitempty"`
	SourceID             string         `json:"source_id,omitempty"`
	Type                 string         `json:"type,omitempty"`
	Category             string         `json:"category,omitempty"`
	IncidentID           string         `json:"incident_id,omitempty"`
	CRNFull              string         `json:"crn_full,omitempty"`
	ResourceDisplayNames []DisplayName  `json:"resource_display_names,omitempty"`
	ShortDescription     []DisplayName  `json:"short_description,omitempty"`
	LongDescription      []DisplayName  `json:"long_description,omitempty"`
}
*/
func TestCollation(t *testing.T) {

	input := []datastore.NotificationReturn{
		datastore.NotificationReturn{
			RecordID:             "11111",
			PnpCreationTime:      "2018-10-31T12:12:12Z",
			PnpUpdateTime:        "2018-10-31T12:12:12Z",
			SourceCreationTime:   "2018-10-31T12:12:12Z",
			SourceUpdateTime:     "2018-10-31T12:12:12Z",
			EventTimeStart:       "2018-10-31T12:12:12Z",
			EventTimeEnd:         "2018-10-31T12:12:12Z",
			Source:               "test1",
			SourceID:             "ID1",
			Type:                 "security",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service1:::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			Tags:                 "",
		},
		datastore.NotificationReturn{
			RecordID:             "22222",
			PnpCreationTime:      "2018-10-31T12:12:12Z",
			PnpUpdateTime:        "2018-10-31T12:12:12Z",
			SourceCreationTime:   "2018-10-31T12:12:12Z",
			SourceUpdateTime:     "2018-10-31T12:12:12Z",
			EventTimeStart:       "2018-10-31T12:12:12Z",
			EventTimeEnd:         "2018-10-31T12:12:12Z",
			Source:               "test1",
			SourceID:             "ID1",
			Type:                 "security",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service2:::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			Tags:                 "",
		},
		datastore.NotificationReturn{
			RecordID:             "33333",
			PnpCreationTime:      "2018-10-31T12:12:12Z",
			PnpUpdateTime:        "2018-10-31T12:12:12Z",
			SourceCreationTime:   "2018-10-31T12:12:12Z",
			SourceUpdateTime:     "2018-10-31T12:12:12Z",
			EventTimeStart:       "2018-10-31T12:12:12Z",
			EventTimeEnd:         "2018-10-31T12:12:12Z",
			Source:               "test1",
			SourceID:             "ID1",
			Type:                 "security",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service3:::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			Tags:                 "",
		},
	}

	output := PGnrToCnList(input, "")

	if len(output.Items) != 1 {
		t.Error("Collation failed")
	}
}

// TestUpdatedBSPN is all about ensuring that updated BSPNs are reflected correctly.  Recall that a
// incident can have multiple BSPNs. A BSPN is never updated.  A new one is created.  So we want to be
// nice to our users and only return the latest (currently relevant) BSPN.
func TestUpdatedBSPN(t *testing.T) {

	input := []datastore.NotificationReturn{
		// First record is when there is a single environment associated with the incident.
		datastore.NotificationReturn{
			RecordID:             "11111",
			PnpCreationTime:      "2018-10-31T12:12:12Z",
			PnpUpdateTime:        "2018-10-31T12:12:12Z",
			SourceCreationTime:   "2018-10-31T12:12:12Z",
			SourceUpdateTime:     "2018-10-31T12:12:12Z",
			EventTimeStart:       "2018-10-31T12:12:12Z",
			EventTimeEnd:         "2018-10-31T12:12:12Z",
			Source:               "test1",
			SourceID:             "BSP001",
			Type:                 "incident",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service:us-south::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service1"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service1"}},
			Tags:                 "",
		},
		// The next two records are after the user added an additional environment (us-east) to the incident and fired a new BSPN.
		datastore.NotificationReturn{
			RecordID:             "22222",
			PnpCreationTime:      "2018-11-01T12:12:12Z",
			PnpUpdateTime:        "2018-11-01T12:12:12Z",
			SourceCreationTime:   "2018-11-01T12:12:12Z",
			SourceUpdateTime:     "2018-11-01T12:12:12Z",
			EventTimeStart:       "2018-11-01T12:12:12Z",
			EventTimeEnd:         "2018-11-01T12:12:12Z",
			Source:               "test1",
			SourceID:             "BSP002",
			Type:                 "incident",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service:us-south::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			Tags:                 "",
		},
		datastore.NotificationReturn{
			RecordID:             "33333",
			PnpCreationTime:      "2018-11-01T12:12:12Z",
			PnpUpdateTime:        "2018-11-01T12:12:12Z",
			SourceCreationTime:   "2018-11-01T12:12:12Z",
			SourceUpdateTime:     "2018-11-01T12:12:12Z",
			EventTimeStart:       "2018-11-01T12:12:12Z",
			EventTimeEnd:         "2018-11-01T12:12:12Z",
			Source:               "test1",
			SourceID:             "BSP002",
			Type:                 "incident",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service:us-east::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			Tags:                 "",
		},
	}

	output := PGnrToCnList(input, "")

	if len(output.Items) != 1 {
		t.Error(fmt.Sprintf("Collation failed for updated BSPN; got record count %d", len(output.Items)))
	}

	if output.Items[0].ShortDescription != "service2" {
		t.Error(fmt.Sprintf("Collation failed for updated BSPN; got wrong short description %s", output.Items[0].ShortDescription))
	}
}

// TestUpdatedBSPNWithSingleTag is the same as the previously similarly named function, but this one adds
// test cases with a single tag
func TestUpdatedBSPNWithSingleTag(t *testing.T) {

	input := []datastore.NotificationReturn{
		// First record is when there is a single environment associated with the incident.
		datastore.NotificationReturn{
			RecordID:             "11111",
			PnpCreationTime:      "2018-10-31T12:12:12Z",
			PnpUpdateTime:        "2018-10-31T12:12:12Z",
			SourceCreationTime:   "2018-10-31T12:12:12Z",
			SourceUpdateTime:     "2018-10-31T12:12:12Z",
			EventTimeStart:       "2018-10-31T12:12:12Z",
			EventTimeEnd:         "2018-10-31T12:12:12Z",
			Source:               "test1",
			SourceID:             "BSP001",
			Type:                 "incident",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service:us-south::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service1"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service1"}},
			Tags:                 "retract-1",
		},
		// The next two records are after the user added an additional environment (us-east) to the incident and fired a new BSPN.
		datastore.NotificationReturn{
			RecordID:             "22222",
			PnpCreationTime:      "2018-11-01T12:12:12Z",
			PnpUpdateTime:        "2018-11-01T12:12:12Z",
			SourceCreationTime:   "2018-11-01T12:12:12Z",
			SourceUpdateTime:     "2018-11-01T12:12:12Z",
			EventTimeStart:       "2018-11-01T12:12:12Z",
			EventTimeEnd:         "2018-11-01T12:12:12Z",
			Source:               "test1",
			SourceID:             "BSP002",
			Type:                 "incident",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service:eu-gb::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			Tags:                 "",
		},
		datastore.NotificationReturn{
			RecordID:             "33333",
			PnpCreationTime:      "2018-11-01T12:12:12Z",
			PnpUpdateTime:        "2018-11-01T12:12:12Z",
			SourceCreationTime:   "2018-11-01T12:12:12Z",
			SourceUpdateTime:     "2018-11-01T12:12:12Z",
			EventTimeStart:       "2018-11-01T12:12:12Z",
			EventTimeEnd:         "2018-11-01T12:12:12Z",
			Source:               "test1",
			SourceID:             "BSP002",
			Type:                 "incident",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service:us-east::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			Tags:                 "",
		},
	}

	output := PGnrToCnList(input, "retract-1")

	if len(output.Items) != 1 {
		t.Error(fmt.Sprintf("Collation failed for updated BSPN; got record count %d", len(output.Items)))
	}

	if output.Items[0].ShortDescription != "service1" {
		t.Error(fmt.Sprintf("Collation failed for updated BSPN; got wrong short description %s", output.Items[0].ShortDescription))
	}
}

// TestUpdatedBSPNWithMultipleTags is the same as the previously similarly named function, but this one adds
// test cases with multiple tags
func TestUpdatedBSPNWithMultipleTags(t *testing.T) {

	input := []datastore.NotificationReturn{
		// First record is when there is a single environment associated with the incident.
		datastore.NotificationReturn{
			RecordID:             "11111",
			PnpCreationTime:      "2018-10-31T12:12:12Z",
			PnpUpdateTime:        "2018-10-31T12:12:12Z",
			SourceCreationTime:   "2018-10-31T12:12:12Z",
			SourceUpdateTime:     "2018-10-31T12:12:12Z",
			EventTimeStart:       "2018-10-31T12:12:12Z",
			EventTimeEnd:         "2018-10-31T12:12:12Z",
			Source:               "test1",
			SourceID:             "BSP001",
			Type:                 "incident",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service:us-south::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service1"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service1"}},
			Tags:                 "dogs,retract-1",
		},
		// The next two records are after the user added an additional environment (us-east) to the incident and fired a new BSPN.
		datastore.NotificationReturn{
			RecordID:             "22222",
			PnpCreationTime:      "2018-11-01T12:12:12Z",
			PnpUpdateTime:        "2018-11-01T12:12:12Z",
			SourceCreationTime:   "2018-11-01T12:12:12Z",
			SourceUpdateTime:     "2018-11-01T12:12:12Z",
			EventTimeStart:       "2018-11-01T12:12:12Z",
			EventTimeEnd:         "2018-11-01T12:12:12Z",
			Source:               "test1",
			SourceID:             "BSP002",
			Type:                 "incident",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service:eu-gb::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			Tags:                 "",
		},
		datastore.NotificationReturn{
			RecordID:             "33333",
			PnpCreationTime:      "2018-11-01T12:12:12Z",
			PnpUpdateTime:        "2018-11-01T12:12:12Z",
			SourceCreationTime:   "2018-11-01T12:12:12Z",
			SourceUpdateTime:     "2018-11-01T12:12:12Z",
			EventTimeStart:       "2018-11-01T12:12:12Z",
			EventTimeEnd:         "2018-11-01T12:12:12Z",
			Source:               "test1",
			SourceID:             "BSP002",
			Type:                 "incident",
			Category:             "runtimes",
			IncidentID:           "INC001",
			CRNFull:              "crn:v1:bluemix:public:service:us-east::::",
			ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service"}},
			ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "service2"}},
			Tags:                 "",
		},
	}

	output := PGnrToCnList(input, "retract-1,dogs")

	if len(output.Items) != 1 {
		t.Error(fmt.Sprintf("Collation failed for updated BSPN; got record count %d", len(output.Items)))
	}

	if output.Items[0].ShortDescription != "service1" {
		t.Error(fmt.Sprintf("Collation failed for updated BSPN; got wrong short description %s", output.Items[0].ShortDescription))
	}
}

func TestNiToReturn(t *testing.T) {
	NInsertToNReturn(&datastore.NotificationInsert{})
}
