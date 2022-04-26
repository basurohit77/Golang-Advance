package test

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

// Test Resource for IBM Cloud, not Gaas
func TestResource(database *sql.DB) int {
	FCT := "TestResource "
	failed := 0

	log.Println("======== "+FCT+" Test 1 ========")

	log.Println("Test simple InsertResource and GetResourceByRecordID")

	resource := datastore.ResourceInsert {
		SourceID:			"source_id1",
		SourceCreationTime: "2018-06-07T22:01:01Z",
		CRNFull: 			"crn:v1:Bluemix:public:Service-Name1:location1::Service-Instance1:Resource-Type1:",
		State:				"ok",
		OperationalStatus: 	"none",
		Source:				"test_source1",
		Status:				"ok",
		RecordHash:			db.CreateRecordIDFromString("Test 1"),
	}
	

	record_id, err,_ := db.InsertResource(database, &resource)
	if err != nil {
		log.Println("Insert Resource failed: ", err.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id)
		log.Println("Insert Resource passed")
	}
	log.Println("Test get Resource")
	resourceReturn, err1,_ := db.GetResourceByRecordID(database, record_id)
	if err1 != nil {
		log.Println("Get Resource failed: ", err1.Error())
		failed++
	} else {
		log.Println("Get result: ", resourceReturn)
		log.Println("resourceReturn.SourceCreationTime: ",resourceReturn.SourceCreationTime)
		log.Println("resource.SourceCreationTime:       ", resource.SourceCreationTime)
		if  resourceReturn.CRNFull == strings.ToLower(resource.CRNFull) &&
			resourceReturn.State == resource.State &&
			resourceReturn.OperationalStatus == resource.OperationalStatus &&
			resourceReturn.SourceID == resource.SourceID &&
			resourceReturn.Source == resource.Source {

			log.Println("Get Resource passed")
		} else {
			log.Println("Get Resource failed")
			failed++
		}
	}

	log.Println("======== "+FCT+" Test 2 ========")

	log.Println("Test InsertResource with DisplayNames, Visibility, Member and GetResourceByRecordID")

	displayNames := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "my display name1",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "my display name2",
			Language: "fr",
		},
	}

	visibility :=[]string{
		"test_vis1",
		"test_vis2",
	}

	tag :=[]datastore.Tag{
		datastore.Tag {
			ID: "test_tag0001",
		},
		datastore.Tag {
			ID: "test_tag0001a",
		},
	}
	resource2 := datastore.ResourceInsert {
		SourceID:			"source_id2",
		SourceCreationTime: "2018-07-07T22:01:01Z",
		CRNFull: 			"crn:v1:bluemix:public:service-name2:location2:scope2:service-instance2:resource-type2:resource2",
		State:				"ok",
		OperationalStatus: 	"none",
		Source:				"test_source2",
		DisplayNames:		displayNames,
		Visibility:			visibility,
		Tags:				tag,
		Status:				"ok",
		RecordHash:			db.CreateRecordIDFromString("Test 2"),
	}
	log.Println("Test insert Resource")
	record_id2, err2,_ := db.InsertResource(database, &resource2)
	if err2 != nil {
		log.Println("Insert Resource failed: ", err2.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id2)
		log.Println("Insert Resource passed")

		log.Println("Test get Resource")
		var resourceReturn2 *datastore.ResourceReturn
		resourceReturn2, err3,_ := db.GetResourceByRecordID(database, record_id2)
		if err3 != nil {
			log.Println("Get Resource failed: ", err3.Error())
			failed++
		} else {
			log.Println("Get result: ", resourceReturn2)
			if  resourceReturn2.CRNFull == strings.ToLower(resource2.CRNFull) &&
				resourceReturn2.State == resource2.State &&
				resourceReturn2.OperationalStatus == resource2.OperationalStatus &&
				resourceReturn2.SourceID == resource2.SourceID &&
				resourceReturn2.Source == resource2.Source {

				log.Println("Get Resource passed")
			} else {
				log.Println("Get Resource failed")
				failed++
			}
		}
	}

	log.Println("======== "+FCT+" Test 3 ========")

	log.Println("Test GetResourceByQuery - query visibility")

	queryStr3 := db.RESOURCE_QUERY_VISIBILITY + "=test_vis2"
	log.Println("query: ", queryStr3)
	resourceReturn3, total_count3, err3,_ := db.GetResourceByQuery(database, queryStr3, 0, 0)
	if err3 != nil {
		log.Println("Get Resource failed: ", err3.Error())
		failed++
	} else if total_count3 == 1 {

		if len(*resourceReturn3) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*resourceReturn3))
			failed++
		} else {
			match := 0
			for idx,r := range *resourceReturn3 {
				if idx == 0 &&
					r.CRNFull == resource2.CRNFull &&
					r.State == strings.ToLower(resource2.State) &&
					r.OperationalStatus == resource2.OperationalStatus &&
					r.SourceID == resource2.SourceID &&
					r.Source == resource2.Source {

					log.Println("Row 0 matched")
					match++
				}
			}
			if (match < 1){
				log.Println("There are rows unmatched, failed")
				failed++
			}
			log.Println("Get Resource passed")
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count3)
		failed++
	}

	log.Println("======== "+FCT+" Test 4 ========")

	log.Println("Test GetResourceByQuery - query crn wildcard")

	queryStr4 := db.RESOURCE_QUERY_CRN + "=crn:v1:bluemix:Public:service-name2:Location2::::"
	log.Println("query: ", queryStr4)
	resourceReturn4, total_count4, err4,_ := db.GetResourceByQuery(database, queryStr4, 0, 0)

	if err4 != nil {
		log.Println("Get Resource failed: ", err4.Error())
		failed++
	} else if total_count4 == 1 {
		log.Println("total_count4: ", total_count4)
		log.Println("resourceReturn4: ", resourceReturn4)

		if len(*resourceReturn4) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*resourceReturn4))
			failed++
		} else {
			match := 0
			for idx,r := range *resourceReturn4 {
				if idx == 0 &&
					r.CRNFull == strings.ToLower(resource2.CRNFull) &&
					r.State == resource2.State &&
					r.OperationalStatus == resource2.OperationalStatus &&
					r.SourceID == resource2.SourceID &&
					r.Source == resource2.Source {

					log.Println("Row 0 matched")
					match++
				}
			}
			if (match < 1){
				log.Println("There are rows unmatched, failed")
				failed++
			} else {
				log.Println("Get Resource passed")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count4)
		failed++
	}

	log.Println("======== "+FCT+" Test 5 ========")

	log.Println("Test InsertResource with DisplayNames, Visibility, Member and GetResourceByQuery - query crn wildcard and visibility")

	displayNames5 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "my display name3",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "my display name4",
			Language: "fr",
		},
	}

	visibility5 :=[]string{
		"test_vis3",
	}

	tag5 :=[]datastore.Tag{
		datastore.Tag {
			ID: "test_tag0005",
		},
	}
	resource5 := datastore.ResourceInsert {
		SourceID:			"source_id3",
		SourceCreationTime: "2018-07-08T22:01:01Z",
		CRNFull: 			"crn:v1:aa:dedicated:service-Name2:location3:scope3:service-Instance3:resource-Type3:resource3",
		State:				"ok",
		OperationalStatus: 	"none",
		Source:				"test_source2",
		DisplayNames:		displayNames5,
		Visibility:			visibility5,
		Tags:				tag5,
		Status:				"ok",
		RecordHash:			db.CreateRecordIDFromString("Test 5"),
	}
	log.Println("Test insert Resource")
	source_id5, err5,_ := db.InsertResource(database, &resource5)
	if err5 != nil {
		log.Println("Insert Resource failed: ", err5.Error())
		failed++
	} else {
		log.Println("source_id: ", source_id5)
		log.Println("Insert Resource passed")

		log.Println("Test get Resource")
		queryStr5 := db.RESOURCE_QUERY_CRN + "=crn:v1:bluemix:public:SERVICE-name2:::::&" + db.RESOURCE_QUERY_VISIBILITY + "=test_vis2"
		log.Println("query: ", queryStr5)
		resourceReturn5, total_count5, e5,_ := db.GetResourceByQuery(database, queryStr5, 0, 0)
		if e5 != nil {
			log.Println("Get Resource failed: ", e5.Error())
			failed++
		} else if total_count5 == 1 {
			log.Println("total_count5: ", total_count5)
			log.Println("resourceReturn5: ", resourceReturn5)

			if len(*resourceReturn5) != 1 {
				log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*resourceReturn5))
				failed++
			} else {
				match := 0
				for idx,r := range *resourceReturn5 {
					if idx == 0 &&
						r.CRNFull == strings.ToLower(resource2.CRNFull) &&
						r.State == resource2.State &&
						r.OperationalStatus == resource2.OperationalStatus &&
						r.SourceID == resource2.SourceID &&
						r.Source == resource2.Source {

						log.Println("Row 0 matched")
						match++
					}
				}
				if (match < 1){
					log.Println("There are rows unmatched, failed")
					failed++
				} else {
					log.Println("Get Resource passed")
				}
			}
		} else {
			log.Printf("total_count is %d, expecting 1, failed", total_count5)
			failed++
		}
	}

	log.Println("======== "+FCT+" Test 5b ========")
	log.Println("Test get Resource")
	queryStr5b := db.RESOURCE_QUERY_CRN + "=crn:v1:aa:dedicated:SERVICE-name2:::::&" + db.RESOURCE_QUERY_VISIBILITY + "=test_vis3"
	log.Println("query: ", queryStr5b)
	resourceReturn5b, total_count5b, e5b,_ := db.GetResourceByQuery(database, queryStr5b, 0, 0)
	if e5b != nil {
		log.Println("Get Resource failed: ", e5b.Error())
		failed++
	} else if total_count5b == 1 {
		log.Println("total_count5b: ", total_count5b)
		log.Println("resourceReturn5b: ", resourceReturn5b)

		if len(*resourceReturn5b) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*resourceReturn5b))
			failed++
		} else {
			match := 0
			for idx,r := range *resourceReturn5b {
				if idx == 0 &&
					r.CRNFull == strings.ToLower(resource5.CRNFull) &&
					r.State == resource5.State &&
					r.OperationalStatus == resource5.OperationalStatus &&
					r.SourceID == resource5.SourceID &&
					r.Source == resource5.Source {

					log.Println("Row 0 matched")
					match++
				}
			}
			if (match < 1){
				log.Println("There are rows unmatched, failed")
				failed++
			} else {
				log.Println("Get Resource passed")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count5b)
		failed++
	}


	log.Println("======== "+FCT+" Test 6 ========")

	log.Println("Test GetResourceByQuery - query not found, should do second search, but still cannot find it")

	queryStr6 := db.RESOURCE_QUERY_CRN + "=crn:v1:aa:dedicated:service-NAME2:location2::::&" + db.RESOURCE_QUERY_VISIBILITY + "=test_vis3"
	log.Println("query: ", queryStr6)
	resourceReturn6, total_count6, err6,_ := db.GetResourceByQuery(database, queryStr6, 0, 0)
	if err6 != nil {
		log.Println("Get Resource failed: ", err6.Error())
		failed++
	} else if total_count6 !=0 {
		log.Printf("Error: total_count: %d, expecting: 0\n", total_count6)
		failed++
	} else if len(*resourceReturn6) != 0 {
		log.Printf("Error: Number of returned: %d, expecting: 0\n", len(*resourceReturn6))
		failed++
	} else {
		log.Println("Get Resource passed")
	}

	log.Println("======== "+FCT+" Test 7 ========")

	log.Println("Test delete archived Resource")

	resource7 := datastore.ResourceInsert {
		SourceCreationTime: "2018-07-07T22:01:01Z",
		CRNFull: 			"crn:v1:bluemix:public:service-name7:location7:scope7:service-instance7:resource-type7:resource7",
		State:				"archived",
		OperationalStatus: 	"none",
		SourceID:			"resource_id7",
		Source:				"test_source7",
		Status:				"ok",
		RegulatoryDomain:   "regulatory domain 1",
		RecordHash:			db.CreateRecordIDFromString("Test 7a"),
	}

	resource7a := datastore.ResourceInsert {
		SourceCreationTime: "2018-07-07T22:01:01Z",
		CRNFull: 			"crn:v1:aa:public:service-name7a:location7a:scope7a:service-instance7a:resource-type7a:resource7a",
		State:				"ok",
		OperationalStatus: 	"none",
		SourceID:			"resource_id7a",
		Source:				"test_source7a",
		Status:				"ok",
		RegulatoryDomain:   "regulatory domain 2",
		RecordHash:			db.CreateRecordIDFromString("Test 7b"),
	}

	record_id7, err,_ := db.InsertResource(database, &resource7)
	if err != nil {
		log.Println("Insert Resource failed: ", err.Error())
	} else {
		log.Println("record_id: ", record_id7)
		log.Println("Insert Resource7 passed")
	}

	record_id7a, err,_ := db.InsertResource(database, &resource7a)
	if err != nil {
		log.Println("Insert Resource failed: ", err.Error())
	} else {
		log.Println("record_id: ", record_id7a)
		log.Println("Insert Resource7a passed")
	}

	err,_ = db.DeleteArchivedResources(database)
	if err != nil {
		log.Println("Delete Resource failed: ", err.Error())
	}

	resourceGet7, _,_ := db.GetResourceByRecordID(database, record_id7)
	if resourceGet7 != nil {
		//record_id7 should have been deleted from database
		failed++
		log.Println("Delete archived Resource failed in resourceGet7")
	}

	resourceGet7a, _,_ := db.GetResourceByRecordID(database, record_id7a)
	if resourceGet7a == nil {
		//record_id7 should not have been deleted from database
		failed++
		log.Println("Delete archived Resource failed in resourceGet7a")
	}

	if resourceGet7 == nil && resourceGet7a != nil {
		log.Println("Delete archived Resource passed")
	}

	log.Println("======== "+FCT+" Test 8 ========")

	log.Println("Test GetResourceSourceIDsByVisibility")

	v := visibility5[0]
	log.Println("visibility: ", v)
	resourceRecordIds := []string{db.CreateRecordIDFromSourceSourceID("test_source1", "source_id1"), db.CreateRecordIDFromSourceSourceID("test_source2", "source_id2"), db.CreateRecordIDFromSourceSourceID("test_source2", "source_id3")}
	expectedRecordId := db.CreateRecordIDFromSourceSourceID("test_source2", "source_id3")

	resourceReturn8, err8, status8 := db.GetResourceRecordIDsByVisibility(database, resourceRecordIds, v)
	if status8 != http.StatusOK {
		log.Println("Get Resource failed: ", err8.Error())
		failed++
	} else if len(resourceReturn8) != 1 {
		log.Printf("Error: Number of returned: %d, expecting: 0\n", len(resourceReturn8))
		failed++
	} else if resourceReturn8[0] != expectedRecordId {
		log.Printf("Error: RecordID is %s. Expecting RecordID is %s\n",
			resourceReturn8[0], expectedRecordId)
		failed++
	} else {
		log.Println("Get Resource passed")
	}

	log.Println("======== "+FCT+" Test 9 ========")

	log.Println("Test Update Resource Simple")

	resource7a.State = "archived"

	err9,_ := db.UpdateResource(database, &resource7a)
	if err9 != nil {
		failed++
		log.Println("Update Resource failed")
	} else {
		resourceGet7a, err,_ = db.GetResourceByRecordID(database, record_id7a)
		if resourceGet7a != nil {
			if resourceGet7a.State == resource7a.State {
				log.Println("Update Resource passed")
			} else {
				failed++
				log.Println("Update Resource failed")
			}
		}
	}

	log.Println("======== "+FCT+" Test 10 ========")

	log.Println("Test Update Resource Full")

	displayNames10 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "my display name10a",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "my display name10b",
			Language: "fr",
		},
	}

	visibility10 :=[]string{
		"test_vis10",
	}

	tag10 :=[]datastore.Tag{
		datastore.Tag {
			ID: "test_tag0007a",
		},
	}
	
	resource10 := datastore.ResourceInsert {
		SourceID:			"source_id10",
		SourceCreationTime: "2018-07-07T22:01:01Z",
		CRNFull: 			"crn:v1:bluemix:public:service-name10:LOCATION10:scope10:service-instance10:resource-type10:resource10",
		State:				"ok",
		OperationalStatus: 	"none",
		Source:				"test_source10",
		DisplayNames:		displayNames10,
		Visibility:			visibility10,
		Tags:				tag10,
		Status:				"ok",
		RecordHash:			db.CreateRecordIDFromString("Test 10"),
	}

	record_id10, err,_ := db.InsertResource(database, &resource10)
	if err != nil {
		log.Println("Insert Resource failed: ", err.Error())
	} else {
		// Remove tags and verify that it is deleted.
		resource10.Tags = []datastore.Tag{}
		err10a,_ := db.UpdateResource(database, &resource10)
		if err10a != nil {
			failed++
			log.Println("Update Resource failed, err10a: ", err10a.Error())
		} else {
			resourceGet10, _,_ := db.GetResourceByRecordID(database, record_id10)
			if resourceGet10 != nil {
				if len(resourceGet10.Tags) == 0 {
					log.Println("Update Resource passed")
				} else {
					failed++
					log.Println("Update Resource failed")
				}
			}
		}
	}

	log.Println("======== "+FCT+" Test 11 ========")

	log.Println("Test Get Resource with query crn wildcard, visibility, offset and limit")

	queryStr11 := db.RESOURCE_QUERY_CRN + "=crn:v1:::service-name2:::::&" + db.RESOURCE_QUERY_VISIBILITY + "=test_vis2,test_vis3"
	log.Println("query: ", queryStr11)
	resourceReturn11, total_count11, e11,_ := db.GetResourceByQuery(database, queryStr11, 1, 1)
	if e11 != nil {
		log.Println("Get Resource failed: ", e11.Error())
		failed++
	} else if total_count11 == 2 {
		log.Println("total_count11: ", total_count11)
		log.Println("resourceReturn11: ", resourceReturn11)

		if len(*resourceReturn11) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*resourceReturn11))
			failed++
		} else {
			match := 0
			for _,r := range *resourceReturn11 {
				// have to use || in the comparison, because the query is returning from offset 1, and 
				// the result set is order by record_id, we cannot predict what is the hash value for 
				// record_id, and which record comes first.
				if  (r.CRNFull == strings.ToLower(resource5.CRNFull) || r.CRNFull == strings.ToLower(resource2.CRNFull)) &&
					(r.State == resource5.State || r.State == resource2.State) &&
					(r.OperationalStatus == resource5.OperationalStatus || r.OperationalStatus == resource2.OperationalStatus) &&
					(r.SourceID == resource5.SourceID || r.SourceID == resource2.SourceID) &&
					(r.Source == resource5.Source || r.Source == resource2.Source) {

					//log.Println("Row 0 matched")
					match++
				}
			}
			if (match < 1){
				log.Println("There are rows unmatched, failed")
				failed++
			} else {
				log.Println("Get Resource passed")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 2, failed", total_count11)
		failed++
	}

	log.Println("======== "+FCT+" Test 11a ========")

	log.Println("Test Get Resource with query crn wildcard, visibility, offset and limit")

	queryStr11a := db.RESOURCE_QUERY_CRN + "=crn:v1:::service-name2:::::&" + db.RESOURCE_QUERY_VISIBILITY + "=test_vis2,test_vis3"
	log.Println("query: ", queryStr11a)
	resourceReturn11a, total_count11a, e11a,_ := db.GetResourceByQuery(database, queryStr11a, 0,0 )
	if e11a != nil {
		log.Println("Get Resource failed: ", e11a.Error())
		failed++
	} else if total_count11a == 2 {
		log.Println("total_count11a: ", total_count11a)
		log.Println("resourceReturn11a: ", resourceReturn11a)

		if len(*resourceReturn11a) != 2 {
			log.Printf("Error: Number of returned: %d, expecting: 2\n", len(*resourceReturn11a))
			failed++
		} else {
			match := 0
			for _,r := range *resourceReturn11a {
				// have to use || in the comparison, because the query is returning from offset 1, and 
				// the result set is order by record_id, we cannot predict what is the hash value for 
				// record_id, and which record comes first.
				if  (r.CRNFull == strings.ToLower(resource5.CRNFull) || r.CRNFull == strings.ToLower(resource2.CRNFull)) &&
					(r.State == resource5.State || r.State == resource2.State) &&
					(r.OperationalStatus == resource5.OperationalStatus || r.OperationalStatus == resource2.OperationalStatus) &&
					(r.SourceID == resource5.SourceID || r.SourceID == resource2.SourceID) &&
					(r.Source == resource5.Source || r.Source == resource2.Source) {

					//log.Println("Row 0 matched")
					match++
				}
			}
			if (match < 2){
				log.Println("There are rows unmatched, failed")
				failed++
			} else {
				log.Println("Get Resource passed")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 2, failed", total_count11a)
		failed++
	}
	
	log.Println("======== "+FCT+" Test 11b ========")

	log.Println("Test Get Resource with query crn wildcard, visibility, offset and limit")

	queryStr11b := db.RESOURCE_QUERY_CRN + "=crn:v1:::service-name2:::::&" + db.RESOURCE_QUERY_VISIBILITY + "=test_vis2,test_vis3"
	log.Println("query: ", queryStr11b)
	resourceReturn11b, total_count11b, e11b,_ := db.GetResourceByQuery(database, queryStr11b, 0,2 ) // limit 0, offset 2
	if e11b != nil {
		log.Println("Get Resource failed: ", e11b.Error())
		failed++
	} else if total_count11b == 2 {
		log.Println("total_count11b: ", total_count11b)
		log.Println("resourceReturn11b: ", resourceReturn11b)

		if len(*resourceReturn11b) != 0 {
			log.Printf("Error: Number of returned: %d, expecting: 0\n", len(*resourceReturn11b))
			failed++
		}
	} else {
		log.Printf("total_count is %d, expecting 2, failed", total_count11b)
		failed++
	}
	
	log.Println("======== "+FCT+" Test 11c ========")

	log.Println("Test Get Resource with query crn wildcard, visibility, offset and limit")

	queryStr11c := db.RESOURCE_QUERY_CRN + "=crn:v1:::service-name2:::::&" + db.RESOURCE_QUERY_VISIBILITY + "=test_vis2,test_vis3"
	log.Println("query: ", queryStr11c)
	resourceReturn11c, total_count11c, e11c,_ := db.GetResourceByQuery(database, queryStr11c, 1, 0) //limit 1, offset 0
	if e11c != nil {
		log.Println("Get Resource failed: ", e11c.Error())
		failed++
	} else if total_count11c == 2 {
		log.Println("total_count11c: ", total_count11c)
		log.Println("resourceReturn11c: ", resourceReturn11c)

		if len(*resourceReturn11c) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*resourceReturn11c))
			failed++
		} else {
			match := 0
			for _,r := range *resourceReturn11c {
				// have to use || in the comparison, because the query is returning from offset 1, and 
				// the result set is order by record_id, we cannot predict what is the hash value for 
				// record_id, and which record comes first.
				if  (r.CRNFull == strings.ToLower(resource5.CRNFull) || r.CRNFull == strings.ToLower(resource2.CRNFull)) &&
					(r.State == resource5.State || r.State == resource2.State) &&
					(r.OperationalStatus == resource5.OperationalStatus || r.OperationalStatus == resource2.OperationalStatus) &&
					(r.SourceID == resource5.SourceID || r.SourceID == resource2.SourceID) &&
					(r.Source == resource5.Source || r.Source == resource2.Source) {

					//log.Println("Row 0 matched")
					match++
				}
			}
			if (match < 1){
				log.Println("There are rows unmatched, failed")
				failed++
			} else {
				log.Println("Get Resource passed")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 2, failed", total_count11c)
		failed++
	}

	log.Println("======== "+FCT+" Test 11d ========")

	log.Println("Test Get Resource with query crn wildcard, visibility, offset and limit")

	queryStr11d := db.RESOURCE_QUERY_CRN + "=crn:v1:aa:dedicated:service-name2:::::&" + db.RESOURCE_QUERY_VISIBILITY + "=test_vis2,test_vis3"
	log.Println("query: ", queryStr11d)
	resourceReturn11d, total_count11d, e11d,_ := db.GetResourceByQuery(database, queryStr11d, 0, 0) //limit 0, offset 0
	if e11d != nil {
		log.Println("Get Resource failed: ", e11d.Error())
		failed++
	} else if total_count11d == 1 {
		log.Println("total_count11c: ", total_count11d)
		log.Println("resourceReturn11c: ", resourceReturn11d)

		if len(*resourceReturn11d) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*resourceReturn11d))
			failed++
		} else {
			match := 0
			for _,r := range *resourceReturn11d {
				// have to use || in the comparison, because the query is returning from offset 1, and 
				// the result set is order by record_id, we cannot predict what is the hash value for 
				// record_id, and which record comes first.
				if  r.CRNFull == strings.ToLower(resource5.CRNFull) &&
					r.State == resource5.State &&
					r.OperationalStatus == resource5.OperationalStatus &&
					r.SourceID == resource5.SourceID &&
					r.Source == resource5.Source {

					match++
				}
			}
			if (match < 1){
				log.Println("There are rows unmatched, failed")
				failed++
			} else {
				log.Println("Get Resource passed")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count11d)
		failed++
	}
	log.Println("**** SUMMARY ****")
	if failed > 0 {
		log.Println("**** "+FCT+"FAILED ****")
	} else {
		log.Println("**** "+FCT+"PASSED ****")
	}
	return failed
}
