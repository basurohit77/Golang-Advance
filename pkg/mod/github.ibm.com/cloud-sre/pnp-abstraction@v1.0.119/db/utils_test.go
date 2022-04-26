package db

import (
	"testing"

	datastore "github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

func TestComputeResourceRecordHash(t *testing.T) {
	record := datastore.ResourceInsert{
		SourceCreationTime: "2017-06-05 19:02:11-05",
		SourceUpdateTime:   "2017-06-20T23:38:37+0230",
		CRNFull:            "crn",
		State:              "state",
		OperationalStatus:  "ostatus",
		Source:             "",
		SourceID:           "sid2",
		Status:             "status",
		StatusUpdateTime:   "2019-06-06T00:02:11Z",
		RegulatoryDomain:   "rdom2",
		CategoryID:         "catid",
		CategoryParent:     true,
		DisplayNames:       []datastore.DisplayName{datastore.DisplayName{Name: "dname", Language: "lang"}},
		Visibility:         []string{"vis"},
		Tags:               []datastore.Tag{datastore.Tag{ID: "tag2"}},
		CatalogParentID:    "catParentID",
		RecordHash:         "123"}

	expectedHash := "4789d7997a52e05c72b70742e2d9ea2cf8747a06529673b8f526ff2ecc52b0d5" // # pragma: whitelist secret
	hash := ComputeResourceRecordHash(&record)
	if expectedHash != hash {
		t.Fatalf("Error running ComputeResourceRecordHash: Did not get the expected hash: %s != %s", expectedHash, hash)
	}
}

func TestComputeMaintenanceRecordHash(t *testing.T) {
	record := datastore.MaintenanceInsert{
		SourceCreationTime: "2017-06-05 19:02:11-05",
		SourceUpdateTime:   "2017-06-20T23:38:37+0230",
		PlannedStartTime:	"2019-06-06T00:02:11Z",
		PlannedEndTime:		"2019-06-07 04:02:11+04",
		ShortDescription:	"Upgrade API Connect/API Management to Version 5.0.7.1",
		LongDescription:	"This change will upgrade the environment to the latest version of API Connect, which is Version 5.0.7.1. This version will provide enhanced stability and reliability. We do not expect any interruptions to API calls. To know more about the fixes in this version, see the \u0026lt;u\u0026gt;\u0026lt;a class=\u0026quot;link\u0026quot; href=\"https://www.ibm.com/support/knowledgecenter/SSMNED_5.0.0/com.ibm.apic.overview.doc/overview_whatsnew.html#ic-homepage__tab-8\" target=\u0026quot;_blank\u0026quot;\u0026gt;What's new for this release\u0026lt;/a\u0026gt;\u0026lt;/u\u0026gt; information.",
		CRNFull:            []string{"crn:v1:d-lloyds:dedicated:apiconnect:eu-gb::::"},
		State:              "new",
		Disruptive:			true,
		SourceID:           "477277",
		Source:             "Doctor-RTC",
		RegulatoryDomain:   "rdom2",
		RecordHash:         "123",
		MaintenanceDuration:360,
		DisruptionType:		"Other (specify in Description)",
		DisruptionDescription:	"There will be minor service disruptions to the user interface of the service while the underlying hosts are restarted in a rolling manner. API calls will not be disrupted during this upgrade.",
		DisruptionDuration:		30,
		CompletionCode:			"compcode",
		PnPRemoved:				false}
				
	expectedHash := "6d3417090fa85596c01464859d7d4fb8e9f7bcfae6b87f416f0766a02d236c90" // # pragma: whitelist secret
	hash := ComputeMaintenanceRecordHash(&record)
	if expectedHash != hash {
		t.Fatalf("Error running ComputeMaintenanceRecordHash: Did not get the expected hash: %s != %s", expectedHash, hash)
	}
}

func TestComputeResourceRecordHashUsingReturn(t *testing.T) {
	record := datastore.ResourceInsert{
		SourceCreationTime: "",
		SourceUpdateTime:   "2017-06-05 19:02:11-05",
		CRNFull:            "crn",
		State:              "state",
		OperationalStatus:  "ostatus",
		Source:             "",
		SourceID:           "sid2",
		Status:             "status",
		StatusUpdateTime:   "",
		RegulatoryDomain:   "rdom2",
		CategoryID:         "catid",
		CategoryParent:     true,
		// IsCatalogParent: true,
		CatalogParentID:    "catParentID",
		DisplayNames:       []datastore.DisplayName{datastore.DisplayName{Name: "dname", Language: "lang"}},
		Visibility:         []string{"vis"},
		Tags:               []datastore.Tag{datastore.Tag{ID: "tag2"}},
		RecordHash:         "123"}

	recordFromDb := datastore.ResourceReturn{RecordID: "id",
		// RecordID:		""
		PnpCreationTime:    "create",
		// PnpUpdateTime:	""
		SourceCreationTime: "2017-06-04 15:01:33-05",
		SourceUpdateTime:   "2017-06-05 19:02:11-05",
		CRNFull:            "crn",
		State:              "state",
		OperationalStatus:  "ostatus",
		Source:             "src",
		SourceID:           "sid",
		Status:             "status",
		StatusUpdateTime:   "2017-06-20T23:38:37+0230",
		RegulatoryDomain:   "rdom",
		CategoryID:         "catid",
		CategoryParent:     true,
		// IsCatalogParent:	true,
		CatalogParentID:    "catParentID",
		DisplayNames:       []datastore.DisplayName{datastore.DisplayName{Name: "dname", Language: "lang"}},
		Visibility:         []string{"vis"},
		Tags:               []datastore.Tag{datastore.Tag{ID: "tag"}},
		RecordHash:         "123"}

	expectedHash := "a673783074742923bd37a2eed0700dbd41aec536c752f676e13b28dc55f2bf27" // # pragma: whitelist secret
	_, hash := ComputeResourceRecordHashUsingReturn(&record, &recordFromDb)
	if expectedHash != hash {
		t.Fatalf("Error running ComputeResourceRecordHashUsingReturn: Did not get the expected hash: %s != %s", expectedHash, hash)
	}

}

func TestComputeMaintenanceRecordHashUsingReturn(t *testing.T) {
	record := datastore.MaintenanceInsert{
		SourceCreationTime: "",
		SourceUpdateTime:    "2019-06-08T00:02:11Z",
		PlannedStartTime:   "2017-06-20T23:38:37-0500",
		PlannedEndTime:     "",
		ShortDescription:   "short desc2",
		LongDescription:    "long desc",
		CRNFull:            []string{"crn"},
		State:              "state",
		Disruptive:         false,
		SourceID:           "sid",
		Source:             "src",
		RegulatoryDomain:   "rDom",
		// RecordHash:		"abcd1234",
		MaintenanceDuration:   123,
		DisruptionType:        "dtype",
		DisruptionDescription: "ddesc",
		// DisruptionDuration:    678,
		CompletionCode:        "ccode",
		// PnPRemoved:		false,
	}

	recordFromDb := datastore.MaintenanceReturn{RecordID: "id",
		// RecordID:		   "",
		PnpCreationTime:       "pctime",
		PnpUpdateTime:         "putime",
		SourceUpdateTime:      "2019-06-08T00:02:11Z",
		SourceCreationTime:    "2019-06-06 00:02:11+03",
		PlannedStartTime:      "2017-06-20T23:38:37-0500",
		PlannedEndTime:        "2017-06-21T09:30:00+0230",
		ShortDescription:      "short desc",
		LongDescription:       "long desc",
		State:                 "state",
		Disruptive:            true,
		SourceID:              "sid",
		Source:                "src",
		RegulatoryDomain:      "rDom",
		CRNFull:               []string{"crn"},
		RecordHash:            "123",
		MaintenanceDuration:   123,
		DisruptionType:        "dtype",
		DisruptionDescription: "ddesc",
		DisruptionDuration:    234,
		CompletionCode:        "ccode",
		/*PnPRemoved:			   false*/}
	expectedHash := "abbbd8f51365c3b48431bb2c901f7df7d33ac8bef43eff967b51819feeb31b40" // # pragma: whitelist secret
	_, hash := ComputeMaintenanceRecordHashUsingReturn(&record, &recordFromDb)
	if expectedHash != hash {
		t.Fatalf("Error running ComputeMaintenanceRecordHashUsingReturn: Did not get the expected hash: %s != %s", expectedHash, hash)
	}

}

func TestCreateNotificationRecordID(t *testing.T) {
	record := datastore.NotificationInsert{
		SourceCreationTime:   "sutime2",
		SourceUpdateTime:     "sutime2",
		Source:               "servicenow",
		SourceID:             "srcId",
		Type:                 "maintenance",
		IncidentID:           "CHG101",
		CRNFull:              "crn:v1:test:public:test:::::",
		ResourceDisplayNames: []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "resource1"}},
		ShortDescription:     []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "short desc2"}},
		LongDescription:      []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: "long desc2"}},
	}

	recordID := CreateNotificationRecordID(record.Source, record.SourceID, record.CRNFull, record.IncidentID, record.Type)
	expectedRecordID := "a8a2f43afbb77c23e37bda3b33c707990411f432d733786ed531a7ba31489a97" // # pragma: whitelist secret

	if expectedRecordID != recordID {
		t.Fatalf("Error running TestCreateNotificationRecordID: Did not get the expected recordID: %s != %s", expectedRecordID, recordID)
	}

	record.Type = "incident"
	recordID = CreateNotificationRecordID(record.Source, record.SourceID, record.CRNFull, record.IncidentID, record.Type)
	expectedRecordID = "619b6203f7d8aaafbf725a0801f0a51f7e7004dd7397a967724f238690a145bd" // # pragma: whitelist secret

	if expectedRecordID != recordID {
		t.Fatalf("Error running TestCreateNotificationRecordID: Did not get the expected recordID: %s != %s", expectedRecordID, recordID)
	}

}
