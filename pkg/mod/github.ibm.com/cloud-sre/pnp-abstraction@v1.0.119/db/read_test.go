package db

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetMaintenanceByQuery(t *testing.T) {
	pdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer pdb.Close()

	rowsIns := sqlmock.NewRows([]string{"record_id", "pnp_creation_time", "pnp_update_time", "source_creation_time", "source_update_time", "start_time", "end_time", "short_description", "long_description", "state", "disruptive", "source_id", "source", "regulatory_domain", "record_hash", "maintenance_duration", "disruption_type", "disruption_description", "disruption_duration", "completion_code", "crn_full", "pnp_removed", "targeted_url", "audience", "total_count"}).
		AddRow("recordid1", "2019-04-30T20:04:37.313Z", "2019-04-30T20:04:37.313Z", "2019-04-30T20:04:37.313Z", "2019-04-30T20:04:37.313Z", "2019-05-30T20:04:37.313Z", "2019-05-30T21:04:37.313Z", "short desc", "long desc", "state1", true, "sourceId", "servicenow", "", "hash123", 60, "disruption type", "disruption description", 60, "complete success", "crn:v1:bluemix:public:cloudantnosqldb:us-south::::", false, "https://foo.com", "Public", 1)
	mock.ExpectQuery("WITH cte AS.*").WillReturnRows(rowsIns)

	mr, _, _, _ := GetMaintenanceByQuery(pdb, "pnp_creation_time_start=2019-04-30T16:28:07.929Z&pnp_update_time_start=2019-04-30T16:28:07.929Z", 0, 1)

	actualReturn := fmt.Sprintf("%+v", mr)
	expectedReturn := "&[{RecordID:recordid1 PnpCreationTime:2019-04-30T20:04:37.313Z PnpUpdateTime:2019-04-30T20:04:37.313Z SourceCreationTime:2019-04-30T20:04:37.313Z SourceUpdateTime:2019-04-30T20:04:37.313Z PlannedStartTime:2019-05-30T20:04:37.313Z PlannedEndTime:2019-05-30T21:04:37.313Z ShortDescription:short desc LongDescription:long desc State:state1 Disruptive:true SourceID:sourceId Source:servicenow RegulatoryDomain: CRNFull:[crn:v1:bluemix:public:cloudantnosqldb:us-south::::] RecordHash:hash123 MaintenanceDuration:60 DisruptionType:disruption type DisruptionDescription:disruption description DisruptionDuration:60 CompletionCode:complete success PnPRemoved:false TargetedURL:https://foo.com Audience:Public}]"
	assert.Equal(t, expectedReturn, actualReturn)

}

// Removed targeted_url is not longer part of the long description. Now it is passed from SN using the field name u_targeted_notification_url
// func TestURLFromLongDescriptionWithMaintenance(t *testing.T) {
// 	pdb, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
// 	}
// 	defer pdb.Close()

// 	longDesc := "Customers may experience issues pushing and pulling images as well as intermittent failures on image management calls such as listing images \n" +
// 		"Current Status and Next Steps:\n" +
// 		"[targeted notification](https://www.ibm.com)\n" +
// 		"The CIE has now been resolved\n" +
// 		"[targeted notification](http://foo.com)"

// 	// Check if there is a reference to `[targeted notification](URL)` in the incoming
// 	// itemToInsert.
// 	// If there is a URL defined, parse it and store it into TargetedURL.
// 	// If there isn't a URL defined, the function returns ErrLongDescNoMatch.
// 	// This error is logged as a warning.
// 	tURL, err := targurl.URLFromLongDescription(longDesc)
// 	if err != nil {
// 		log.Println(" Warn:", err)
// 	}

// 	rowsIns := sqlmock.NewRows([]string{"record_id", "pnp_creation_time", "pnp_update_time", "source_creation_time", "source_update_time", "start_time", "end_time", "short_description", "long_description", "state", "disruptive", "source_id", "source", "regulatory_domain", "record_hash", "maintenance_duration", "disruption_type", "disruption_description", "disruption_duration", "completion_code", "crn_full", "pnp_removed", "targeted_url", "total_count"}).
// 		AddRow("recordid1", "2019-04-30T20:04:37.313Z", "2019-04-30T20:04:37.313Z", "2019-04-30T20:04:37.313Z", "2019-04-30T20:04:37.313Z", "2019-05-30T20:04:37.313Z", "2019-05-30T21:04:37.313Z", "short desc", longDesc, "state1", true, "sourceId", "servicenow", "", "hash123", 60, "disruption type", "disruption description", 60, "complete success", "crn:v1:bluemix:public:cloudantnosqldb:us-south::::", false, tURL, 1)
// 	mock.ExpectQuery("WITH cte AS.*").WillReturnRows(rowsIns)

// 	mr, _, _, _ := GetMaintenanceByQuery(pdb, "pnp_creation_time_start=2019-04-30T16:28:07.929Z&pnp_update_time_start=2019-04-30T16:28:07.929Z", 0, 1)

// 	actualReturn := fmt.Sprintf("%+v", mr)
// 	expectedReturn := "&[{RecordID:recordid1 PnpCreationTime:2019-04-30T20:04:37.313Z PnpUpdateTime:2019-04-30T20:04:37.313Z SourceCreationTime:2019-04-30T20:04:37.313Z SourceUpdateTime:2019-04-30T20:04:37.313Z PlannedStartTime:2019-05-30T20:04:37.313Z PlannedEndTime:2019-05-30T21:04:37.313Z ShortDescription:short desc LongDescription:Customers may experience issues pushing and pulling images as well as intermittent failures on image management calls such as listing images \nCurrent Status and Next Steps:\n[targeted notification](https://www.ibm.com)\nThe CIE has now been resolved\n[targeted notification](http://foo.com) State:state1 Disruptive:true SourceID:sourceId Source:servicenow RegulatoryDomain: CRNFull:[crn:v1:bluemix:public:cloudantnosqldb:us-south::::] RecordHash:hash123 MaintenanceDuration:60 DisruptionType:disruption type DisruptionDescription:disruption description DisruptionDuration:60 CompletionCode:complete success PnPRemoved:false TargetedURL:https://www.ibm.com}]"
// 	assert.Equal(t, expectedReturn, actualReturn)

// 	// Must return an empty string
// 	longDesc = "Does description does not have a targeted tag"
// 	tURL, err = targurl.URLFromLongDescription(longDesc)
// 	log.Println("URLFromLongDescription: ", tURL)
// 	log.Println("longDesc: ", longDesc)

// 	if err != nil {
// 		log.Println(" Warn:", err)
// 	}
// 	assert.Equal(t, "", tURL)

// 	// Since the URL is invalid will return a null sting
// 	longDesc = "Invalid URL [targeted notification](invalid url)"
// 	tURL, err = targurl.URLFromLongDescription(longDesc)
// 	log.Println("URLFromLongDescription: ", tURL)

// 	if err != nil {
// 		log.Println(" Warn:", err)
// 	}
// 	assert.Equal(t, "targurl: Supplied URL is invalid", err.Error()+tURL)

// }

func TestGetNotificationByQuery(t *testing.T) {
	pdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer pdb.Close()
	rowsIns := sqlmock.NewRows([]string{"RecordID", "PnpCreationTime", "PnpUpdateTime", "SourceCreationTime", "SourceUpdateTime", "EventTimeStart", "EventTimeEnd", "Source", "SourceID", "Type", "Category", "IncidentID", "CRNFull", "ResourceDisplayNames", "ShortDescription", "Tags", "RecordRetractionTime", "PnPRemoved", "ReleaseNoteUrl", "notifDescriptionLongDescription", "notifDescriptionLanguage", "totalCount"}).
		AddRow("recordid1", "2019-04-30T20:04:37.313Z", "2019-04-30T20:04:37.313Z", "2019-04-30T20:04:37.313Z", "2019-04-30T20:04:37.313Z", "2019-05-30T20:04:37.313Z", "2019-05-30T21:04:37.313Z", "CHG0083830", "servicenow", "maintenance", "", "5136b6abf6ea05e67e8b2f6bd50fc70083e96cf96305c7f0185d9a119714efba", "crn:v1:bluemix:public:ibm-blockchain-5-prod:eu-de::::", "{en                  Cloud Object Storage}", "s{en                  blockchain has a Provisioning issue in eu-de.}", "tag1, tag2", "retract, retract-1", false, "http//ghost.com", "notification desc", "en", 1) // pragma: whitelist secret
	mock.ExpectQuery("WITH cte AS.*").WillReturnRows(rowsIns)

	nr, _, _, _ := GetNotificationByQuery(pdb, "pnp_creation_time_start=2019-04-30T16:28:07.929Z&pnp_update_time_start=2019-04-30T16:28:07.929Z", 0, 1)

	log.Printf("%+v", nr)
	actualReturn := fmt.Sprintf("%+v", nr)
	expectedReturn := "&[{RecordID:recordid1 PnpCreationTime:2019-04-30T20:04:37.313Z PnpUpdateTime:2019-04-30T20:04:37.313Z SourceCreationTime:2019-04-30T20:04:37.313Z SourceUpdateTime:2019-04-30T20:04:37.313Z EventTimeStart:2019-05-30T20:04:37.313Z EventTimeEnd:2019-05-30T21:04:37.313Z Source:CHG0083830 SourceID:servicenow Type:maintenance Category: IncidentID:5136b6abf6ea05e67e8b2f6bd50fc70083e96cf96305c7f0185d9a119714efba CRNFull:crn:v1:bluemix:public:ibm-blockchain-5-prod:eu-de:::: ResourceDisplayNames:[{Name:Cloud Object Storage Language:en}] ShortDescription:[{Name:  blockchain has a Provisioning issue in eu-de. Language:s{en}] LongDescription:[{Name:notification desc Language:en}] Tags:tag1, tag2 RecordRetractionTime:retract, retract-1 PnPRemoved:false ReleaseNoteUrl:http//ghost.com}]"
	assert.Equal(t, expectedReturn, actualReturn)

}

func TestGetResourceByQuery(t *testing.T) {
	pdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer pdb.Close()

	// Test pnp_creation_times
	rowsIns := sqlmock.NewRows([]string{"RecordID", "PnpCreationTime", "PnpUpdateTime", "SourceCreationTime", "SourceUpdateTime", "State", "OperationalStatus", "Source", "SourceID", "Status", "StatusUpdateTime", "RegulatoryDomain", "CategoryID", "CategoryParent", "CRNFull", "IsCatalogParent", "CatalogParentID", "RecordHash", "visibilityName", "tagID", "displayNameName", "displayNameLanguage", "totalCount"}).
		AddRow("f65851b86b3ef8039a5a683896a9719217678f8f2204f302d3bdf560864ac836", "43594.5125462963", "43594.5125462963", "43501.7621180556", "43501.7621180556", "ok", "GA", "globalCatalog", "crn:v1:bluemix:public:iam-ui:dal10::::", "ok", "43501.7621180556", "", "cloudoe.sop.enum.paratureCategory.literal.l133", false, "crn:v1:bluemix:public:iam-ui:dal10::::", true, "", "iam-ui", "hasStatus", "iaas,lux", "Test Display Name", "en", 1) // pragma: whitelist secret
	mock.ExpectQuery("WITH cte AS.*").WillReturnRows(rowsIns)

	var buf bytes.Buffer
	log.SetOutput(&buf)

	nr, _, _, _ := GetResourceByQuery(pdb, "pnp_creation_time_start=2019-04-30T16:28:07.929Z&pnp_creation_time_end=2019-05-10T16:28:07.929Z", 0, 1)
	log.SetOutput(os.Stderr)
	log.Print("QUERY:", buf.String())
	log.Printf("%+v", nr)
	actualReturn := fmt.Sprintf("%+v", nr)
	expectedQuery := "WITH cte AS (SELECT DISTINCT r.record_id FROM resource_table r inner join visibility_junction_table j on r.record_id=j.resource_id inner join visibility_table v on j.visibility_id=v.record_id WHERE  (r.pnp_creation_time>='2019-04-30T16:28:07.929Z') AND (r.pnp_creation_time<='2019-05-10T16:28:07.929Z')) SELECT DISTINCT * FROM ( SELECT DISTINCT r.record_id,r.pnp_creation_time,r.pnp_update_time,r.source_creation_time,r.source_update_time,r.state,r.operational_status,r.source,r.source_id,r.status,r.status_update_time,r.regulatory_domain,r.category_id,r.category_parent,r.crn_full,r.is_catalog_parent,r.catalog_parent_id,r.record_hash,v.name,t.id,d.name,d.language FROM resource_table r full join visibility_junction_table j on r.record_id=j.resource_id full join visibility_table v on j.visibility_id=v.record_id full join tag_junction_table tj on r.record_id=tj.resource_id full join tag_table t on tj.tag_id=t.record_id full join display_names_table d on d.resource_id=r.record_id WHERE r.record_id= ANY (SELECT * FROM cte  ORDER BY record_id OFFSET 1) ) AS r2 RIGHT JOIN  (SELECT count(*) FROM cte) c(total_count) ON true ORDER BY record_id;"
	expectedReturn := "&[{RecordID:f65851b86b3ef8039a5a683896a9719217678f8f2204f302d3bdf560864ac836 PnpCreationTime:43594.5125462963 PnpUpdateTime:43594.5125462963 SourceCreationTime:43501.7621180556 SourceUpdateTime:43501.7621180556 CRNFull:crn:v1:bluemix:public:iam-ui:dal10:::: State:ok OperationalStatus:GA Source:globalCatalog SourceID:crn:v1:bluemix:public:iam-ui:dal10:::: Status:ok StatusUpdateTime:43501.7621180556 RegulatoryDomain: CategoryID:cloudoe.sop.enum.paratureCategory.literal.l133 CategoryParent:false IsCatalogParent:true CatalogParentID: DisplayNames:[{Name:Test Display Name Language:en}] Visibility:[hasStatus] Tags:[{ID:iaas,lux}] RecordHash:}]"
	assert.Equal(t, expectedReturn, actualReturn)
	assert.Contains(t, buf.String(), expectedQuery, "Incorrect query returned:\n "+buf.String())

	// Test pnp_update_times
	buf.Reset()

	rowsIns = sqlmock.NewRows([]string{"RecordID", "PnpCreationTime", "PnpUpdateTime", "SourceCreationTime", "SourceUpdateTime", "State", "OperationalStatus", "Source", "SourceID", "Status", "StatusUpdateTime", "RegulatoryDomain", "CategoryID", "CategoryParent", "CRNFull", "IsCatalogParent", "CatalogParentID", "RecordHash", "visibilityName", "tagID", "displayNameName", "displayNameLanguage", "totalCount"}).
		AddRow("f65851b86b3ef8039a5a683896a9719217678f8f2204f302d3bdf560864ac836", "43594.5125462963", "43594.5125462963", "43501.7621180556", "43501.7621180556", "ok", "GA", "globalCatalog", "crn:v1:bluemix:public:iam-ui:dal10::::", "ok", "43501.7621180556", "", "cloudoe.sop.enum.paratureCategory.literal.l133", false, "crn:v1:bluemix:public:iam-ui:dal10::::", true, "", "iam-ui", "hasStatus", "iaas,lux", "Test Display Name", "en", 1) // pragma: whitelist secret

	mock.ExpectQuery("WITH cte AS.*").WillReturnRows(rowsIns)
	log.SetOutput(&buf)
	nr, _, _, _ = GetResourceByQuery(pdb, "pnp_update_time_start=2019-04-30T16:28:07.929Z&pnp_update_time_end=2019-05-10T16:28:07.929Z", 0, 1)
	log.SetOutput(os.Stderr)
	log.Print("QUERY:", buf.String())
	log.Printf("%+v", nr)
	actualReturn = fmt.Sprintf("%+v", nr)
	assert.Equal(t, expectedReturn, actualReturn)
	expectedQuery = "WITH cte AS (SELECT DISTINCT r.record_id FROM resource_table r inner join visibility_junction_table j on r.record_id=j.resource_id inner join visibility_table v on j.visibility_id=v.record_id WHERE  (r.pnp_update_time>='2019-04-30T16:28:07.929Z') AND (r.pnp_update_time<='2019-05-10T16:28:07.929Z')) SELECT DISTINCT * FROM ( SELECT DISTINCT r.record_id,r.pnp_creation_time,r.pnp_update_time,r.source_creation_time,r.source_update_time,r.state,r.operational_status,r.source,r.source_id,r.status,r.status_update_time,r.regulatory_domain,r.category_id,r.category_parent,r.crn_full,r.is_catalog_parent,r.catalog_parent_id,r.record_hash,v.name,t.id,d.name,d.language FROM resource_table r full join visibility_junction_table j on r.record_id=j.resource_id full join visibility_table v on j.visibility_id=v.record_id full join tag_junction_table tj on r.record_id=tj.resource_id full join tag_table t on tj.tag_id=t.record_id full join display_names_table d on d.resource_id=r.record_id WHERE r.record_id= ANY (SELECT * FROM cte  ORDER BY record_id OFFSET 1) ) AS r2 RIGHT JOIN  (SELECT count(*) FROM cte) c(total_count) ON true ORDER BY record_id;"
	assert.Contains(t, buf.String(), expectedQuery, "Incorrect query returned:\n "+buf.String())

}

func TestNotificationByQuery(t *testing.T) {
	pdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer pdb.Close()
	rowsIns := sqlmock.NewRows([]string{"record_id", "pnp_creation_time", "pnp_update_time", "source_creation_time", "source_update_time", "event_time_start", "event_time_end", "source", "source_id", "type", "category", "incident_id", "crn_full", "resource_display_name", "short_description", "tags", "record_retraction_time", "PnPRemoved", "ReleaseNoteUrl", "long_description", "language", "total_count"}).
		AddRow("a14e9fc78a67dcdb283144f95e7c1a85d3ae7a6e20dd33a819cddc186e3cdc93", "43585.8754166667", "43594.9221990741", "43571.7765509259", "43582.8333680556", "43571.714849537", "43571.9369675926", "servicenow", "INC0739328", "incident", "services", "INC0739328", "crn:v1:bluemix:public:virtual-server:sjc04::::", "{en                  Virtual Servers}", "{en                  Virtual-Server has a Provisioning issue in sjc04.}", "", "", false, "http//ghost.com", "Customers may experience provisioning delays", "en", 1) // pragma: whitelist secret
	mock.ExpectQuery("WITH cte AS.*").WillReturnRows(rowsIns)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	// Creation time
	nr, _, _, _ := GetNotificationByQuery(pdb, "pnp_creation_time_start=2019-04-30T16:28:07.929Z&pnp_creation_time_end=2019-05-04T16:28:07.929Z", 0, 1)
	log.SetOutput(os.Stderr)
	log.Print("QUERY:", buf.String())
	log.Printf("%+v", nr)
	actualReturn := fmt.Sprintf("%+v", nr)
	expectedReturn := "&[{RecordID:a14e9fc78a67dcdb283144f95e7c1a85d3ae7a6e20dd33a819cddc186e3cdc93 PnpCreationTime:43585.8754166667 PnpUpdateTime:43594.9221990741 SourceCreationTime:43571.7765509259 SourceUpdateTime:43582.8333680556 EventTimeStart:43571.714849537 EventTimeEnd:43571.9369675926 Source:servicenow SourceID:INC0739328 Type:incident Category:services IncidentID:INC0739328 CRNFull:crn:v1:bluemix:public:virtual-server:sjc04:::: ResourceDisplayNames:[{Name:Virtual Servers Language:en}] ShortDescription:[{Name:Virtual-Server has a Provisioning issue in sjc04. Language:en}] LongDescription:[{Name:Customers may experience provisioning delays Language:en}] Tags: RecordRetractionTime: PnPRemoved:false ReleaseNoteUrl:http//ghost.com}]"
	assert.Equal(t, expectedReturn, actualReturn)
	expectedQuery := "WITH cte AS (SELECT n.record_id FROM notification_table n WHERE  (n.pnp_creation_time>='2019-04-30T16:28:07.929Z') AND (n.pnp_creation_time<='2019-05-04T16:28:07.929Z') AND (n.pnp_removed='false')) SELECT DISTINCT * FROM ( SELECT DISTINCT n.record_id,n.pnp_creation_time,n.pnp_update_time,n.source_creation_time,n.source_update_time,n.event_time_start,n.event_time_end,n.source,n.source_id,n.type,n.category,n.incident_id,n.crn_full,n.resource_display_name,n.short_description,n.tags,n.record_retraction_time,n.pnp_removed,n.release_note_url,nd.long_description,nd.language FROM notification_table n full join notification_description_table nd on nd.notification_id=n.record_id WHERE n.record_id= ANY (SELECT * FROM cte  ORDER BY record_id OFFSET 1) ) AS n2 RIGHT JOIN  (SELECT count(*) FROM cte) c(total_count) ON true ORDER BY record_id"
	assert.Contains(t, buf.String(), expectedQuery, "Incorrect query returned:\n "+buf.String())

	// Update times
	rowsIns = sqlmock.NewRows([]string{"record_id", "pnp_creation_time", "pnp_update_time", "source_creation_time", "source_update_time", "event_time_start", "event_time_end", "source", "source_id", "type", "category", "incident_id", "crn_full", "resource_display_name", "short_description", "tags", "record_retraction_time", "pnp_removed", "ReleaseNoteUrl", "long_description", "language", "total_count"}).
		AddRow("a14e9fc78a67dcdb283144f95e7c1a85d3ae7a6e20dd33a819cddc186e3cdc93", "43585.8754166667", "43594.9221990741", "43571.7765509259", "43582.8333680556", "43571.714849537", "43571.9369675926", "servicenow", "INC0739328", "incident", "services", "INC0739328", "crn:v1:bluemix:public:virtual-server:sjc04::::", "{en                  Virtual Servers}", "{en                  Virtual-Server has a Provisioning issue in sjc04.}", "", "", false, "http//ghost.com", "Customers may experience provisioning delays", "en", 1) // pragma: whitelist secret
	mock.ExpectQuery("WITH cte AS.*").WillReturnRows(rowsIns)

	buf.Reset()

	log.SetOutput(&buf)
	nr, _, _, _ = GetNotificationByQuery(pdb, "pnp_update_time_start=2019-04-30T16:28:07.929Z&pnp_update_time_end=2019-05-04T16:28:07.929Z", 0, 1)
	log.SetOutput(os.Stderr)
	log.Print("QUERY:", buf.String())
	log.Printf("%+v", nr)
	actualReturn = fmt.Sprintf("%+v", nr)
	assert.Equal(t, expectedReturn, actualReturn)
	expectedQuery = "WITH cte AS (SELECT n.record_id FROM notification_table n WHERE  (n.pnp_update_time>='2019-04-30T16:28:07.929Z') AND (n.pnp_update_time<='2019-05-04T16:28:07.929Z') AND (n.pnp_removed='false')) SELECT DISTINCT * FROM ( SELECT DISTINCT n.record_id,n.pnp_creation_time,n.pnp_update_time,n.source_creation_time,n.source_update_time,n.event_time_start,n.event_time_end,n.source,n.source_id,n.type,n.category,n.incident_id,n.crn_full,n.resource_display_name,n.short_description,n.tags,n.record_retraction_time,n.pnp_removed,n.release_note_url,nd.long_description,nd.language FROM notification_table n full join notification_description_table nd on nd.notification_id=n.record_id WHERE n.record_id= ANY (SELECT * FROM cte  ORDER BY record_id OFFSET 1) ) AS n2 RIGHT JOIN  (SELECT count(*) FROM cte) c(total_count) ON true ORDER BY record_id"
	assert.Contains(t, buf.String(), expectedQuery, "Incorrect query returned:\n "+buf.String())

}

// func TestComputeResourceRecordHashUsingReturnWithInteger(t *testing.T) {

// 	record := datastore.ResourceInsert{}
// 	record.SourceCreationTime = "sctime2"
// 	record.SourceUpdateTime = "sutime"
// 	record.CRNFull = "crn"
// 	record.State = "state"
// 	record.OperationalStatus = "ostatus"
// 	record.Source = ""
// 	record.SourceID = "sid2"
// 	record.Status = "status"
// 	record.StatusUpdateTime = "stutime"
// 	record.RegulatoryDomain = "rdom2"
// 	record.CategoryID = "catid"
// 	record.CategoryParent = true
// 	record.DisplayNames = []datastore.DisplayName{datastore.DisplayName{Name: "dname", Language: "lang"}}
// 	record.Visibility = []string{"vis"}
// 	record.Tags = []datastore.Tag{datastore.Tag{ID: "tag2"}}
// 	record.RecordHash = "123"
// 	tempint2 := 234
// 	record.IntValue2 = &tempint2

// 	bout, _ := json.Marshal(record)
// 	fmt.Println("bout: ", string(bout))

// 	recordFromDb := datastore.ResourceReturn{}
// 	recordFromDb.RecordID = "id"
// 	recordFromDb.PnpCreationTime = "create"
// 	recordFromDb.SourceCreationTime = "sctime"
// 	recordFromDb.SourceUpdateTime = "sutime"
// 	recordFromDb.CRNFull = "crn"
// 	recordFromDb.State = "state"
// 	recordFromDb.OperationalStatus = "ostatus"
// 	recordFromDb.Source = "src"
// 	recordFromDb.SourceID = "sid"
// 	recordFromDb.Status = "status"
// 	recordFromDb.StatusUpdateTime = "stutime"
// 	recordFromDb.RegulatoryDomain = "rdom"
// 	recordFromDb.CategoryID = "catid"
// 	recordFromDb.CategoryParent = true
// 	recordFromDb.DisplayNames = []datastore.DisplayName{datastore.DisplayName{Name: "dname", Language: "lang"}}
// 	recordFromDb.Visibility = []string{"vis"}
// 	recordFromDb.Tags = []datastore.Tag{datastore.Tag{ID: "tag"}}
// 	recordFromDb.RecordHash = "123"
// 	tempint := 123
// 	recordFromDb.IntValue = &tempint

// 	expectedHash := "3f2fe38bd532698894378121a03378140e46f0b7ee8c6b39a774122276a10e3e" // # pragma: whitelist secret
// 	hash := ComputeResourceRecordHashUsingReturn(&record, &recordFromDb)
// 	if expectedHash != hash {
// 		t.Fatalf("Error running ComputeResourceRecordHashUsingReturn: Did not get the expected hash: %s != %s", expectedHash, hash)
// 	}

// }

func TestGetIncidentByQuery(t *testing.T) {
	pdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer pdb.Close()

	rowsIns := sqlmock.NewRows([]string{"record_id", "pnp_creation_time", "pnp_update_time", "source_creation_time", "source_update_time", "start_time", "end_time", "short_description", "long_description", "state", "classification", "severity", "source", "source_id", "regulatory_domain", "crn_full", "affected_activity", "customer_impact_description", "pnp_removed", "targeted_url", "audience", "total_count"}).
		AddRow("a14e9fc78a67dcdb283144f95e7c1a85d3ae7a6e20dd33a819cddc186e3cdc93", "43585.8754166667", "43594.9221990741", "43571.7765509259", "43582.8333680556", "43571.714849537", "43571.9369675926", "short description here", "long description here", "state1", "normal", "1", "servicenow", "INC0739328", "", "crn:v1:bluemix:public:virtual-server:sjc04::::", "Service Availability", "Customers may experience provisioning delay", false, "https://foo.com", "Public", 1) // pragma: whitelist secret
	mock.ExpectQuery("WITH cte AS.*").WillReturnRows(rowsIns)

	incidentRec, _, _, _ := GetIncidentByQuery(pdb, "pnp_creation_time_start=2019-04-30T16:28:07.929Z&pnp_creation_time_end=2019-05-04T16:28:07.929Z", 0, 1)
	actualReturn := fmt.Sprintf("%+v", incidentRec)
	expectedReturn := "&[{RecordID:a14e9fc78a67dcdb283144f95e7c1a85d3ae7a6e20dd33a819cddc186e3cdc93 PnpCreationTime:43585.8754166667 PnpUpdateTime:43594.9221990741 SourceCreationTime:43571.7765509259 SourceUpdateTime:43582.8333680556 OutageStartTime:43571.714849537 OutageEndTime:43571.9369675926 ShortDescription:short description here LongDescription:long description here State:state1 Classification:normal Severity:1 CRNFull:[crn:v1:bluemix:public:virtual-server:sjc04::::] SourceID:servicenow Source:INC0739328 RegulatoryDomain: AffectedActivity:Service Availability CustomerImpactDescription:Customers may experience provisioning delay PnPRemoved:false TargetedURL:https://foo.com Audience:Public}]"
	assert.Equal(t, expectedReturn, actualReturn)
}

// Removed targeted_url is not longer part of the long description. Now it is passed from SN using the field name u_targeted_notification_url
// func TestURLFromLongDescriptionWithIncident(t *testing.T) {
// 	pdb, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
// 	}
// 	defer pdb.Close()
// 	//longDesc := "long desc"

// 	longDesc := "Customers may experience issues pushing and pulling images as well as intermittent failures on image management calls such as listing images \n" +
// 		"Current Status and Next Steps:\n" +
// 		"[targeted notification](https://www.ibm.com)\n" +
// 		"The CIE has now been resolved\n" +
// 		"[targeted notification](http://foo.com)"

// 	// Check if there is a reference to `[targeted notification](URL)` in the incoming
// 	// itemToInsert.
// 	// If there is a URL defined, parse it and store it into TargetedURL.
// 	// If there isn't a URL defined, the function returns ErrLongDescNoMatch.
// 	// This error is logged as a warning.
// 	tURL, err := targurl.URLFromLongDescription(longDesc)
// 	if err != nil {
// 		log.Println(" Warn:", err)
// 	}

// 	rowsIns := sqlmock.NewRows([]string{"record_id", "pnp_creation_time", "pnp_update_time", "source_creation_time", "source_update_time", "start_time", "end_time", "short_description", "long_description", "state", "classification", "severity", "source", "source_id", "regulatory_domain", "crn_full", "affected_activity", "customer_impact_description", "pnp_removed", "targeted_url", "total_count"}).
// 		AddRow("a14e9fc78a67dcdb283144f95e7c1a85d3ae7a6e20dd33a819cddc186e3cdc93", "43585.8754166667", "43594.9221990741", "43571.7765509259", "43582.8333680556", "43571.714849537", "43571.9369675926", "short description here", longDesc, "state1", "normal", "1", "servicenow", "INC0739328", "", "crn:v1:bluemix:public:virtual-server:sjc04::::", "Service Availability", "Customers may experience provisioning delay", false, tURL, 1) // pragma: whitelist secret
// 	mock.ExpectQuery("WITH cte AS.*").WillReturnRows(rowsIns)

// 	mr, _, _, _ := GetIncidentByQuery(pdb, "pnp_creation_time_start=2019-04-30T16:28:07.929Z&pnp_update_time_start=2019-04-30T16:28:07.929Z", 0, 1)

// 	actualReturn := fmt.Sprintf("%+v", mr)
// 	expectedReturn := "&[{RecordID:a14e9fc78a67dcdb283144f95e7c1a85d3ae7a6e20dd33a819cddc186e3cdc93 PnpCreationTime:43585.8754166667 PnpUpdateTime:43594.9221990741 SourceCreationTime:43571.7765509259 SourceUpdateTime:43582.8333680556 OutageStartTime:43571.714849537 OutageEndTime:43571.9369675926 ShortDescription:short description here LongDescription:Customers may experience issues pushing and pulling images as well as intermittent failures on image management calls such as listing images \nCurrent Status and Next Steps:\n[targeted notification](https://www.ibm.com)\nThe CIE has now been resolved\n[targeted notification](http://foo.com) State:state1 Classification:normal Severity:1 CRNFull:[crn:v1:bluemix:public:virtual-server:sjc04::::] SourceID:servicenow Source:INC0739328 RegulatoryDomain: AffectedActivity:Service Availability CustomerImpactDescription:Customers may experience provisioning delay PnPRemoved:false TargetedURL:https://www.ibm.com}]"
// 	assert.Equal(t, expectedReturn, actualReturn)
// }
