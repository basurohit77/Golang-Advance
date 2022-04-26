package test

import (
	"database/sql"
	"log"
	"strings"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)


// Test Maintenance for IBM Public Cloud
func TestMaintenance(database *sql.DB) int {
	FCT := "TestMaintenance "
	failed := 0

	log.Println("======== "+FCT+" Test 1 InsertMaintenance ========")
	crnFull := []string{"crn:v1:bluemix:public:service-name1:location1::::",
				}

	maintenance := datastore.MaintenanceInsert {
		SourceCreationTime:    "2018-07-07T22:01:01Z",
		SourceUpdateTime:      "2018-07-07T22:01:01Z",
		PlannedStartTime:      "2018-07-07T21:55:30Z",
		PlannedEndTime:        "2018-07-21T21:55:30Z",
		ShortDescription:      "maintenance short description 1",
		LongDescription:       "maintenance long description 1",
		State:                 "new",
		Disruptive:            true,
		CRNFull:               crnFull,
		SourceID:              "INC00011",
		Source:                "test_source_1",
		MaintenanceDuration:   240,
		DisruptionType:        "Running Applications,Application Management (start/stop/stage/etc.)",
		DisruptionDescription: "During this change, there might be occasional intermittent",
		DisruptionDuration:    20,
		RecordHash:            db.CreateRecordIDFromString("Test 1"),
		CompletionCode:        "completion code 1",
	}

	record_id, err,_ := db.InsertMaintenance(database, &maintenance)
	if err != nil {
		log.Println("Insert Maintenance failed: ", err.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id)
		log.Println("Insert Maintenance passed")
	}


	log.Println("Test get Maintenance")
	maintenanceReturn, err1,_ := db.GetMaintenanceByRecordIDStatement(database, record_id)
	if err1 != nil {
		log.Println("Get Maintenance failed: ", err1.Error())
		failed++
	} else {

		log.Println("Get result: ", maintenanceReturn)

		if maintenanceReturn.ShortDescription == maintenance.ShortDescription &&
			maintenanceReturn.LongDescription == maintenance.LongDescription &&
			maintenanceReturn.State == maintenance.State &&
			maintenanceReturn.Disruptive == maintenance.Disruptive &&
			len(maintenanceReturn.CRNFull) == len(maintenance.CRNFull) &&
			maintenanceReturn.CRNFull[0] == strings.ToLower(maintenance.CRNFull[0]) &&
			maintenanceReturn.SourceID == maintenance.SourceID &&
			maintenanceReturn.Source == maintenance.Source &&
			maintenanceReturn.MaintenanceDuration == maintenance.MaintenanceDuration &&
			maintenanceReturn.DisruptionType == maintenance.DisruptionType &&
			maintenanceReturn.DisruptionDescription == maintenance.DisruptionDescription &&
			maintenanceReturn.DisruptionDuration == maintenance.DisruptionDuration &&
			maintenanceReturn.RecordHash == maintenance.RecordHash &&
			maintenanceReturn.CompletionCode == maintenance.CompletionCode {

			log.Println("Test 1a: Get Maintenance PASSED")
		} else {
			log.Println("Test 1a: Get Maintenance FAILED")
			failed++
		}
	}

	log.Println("======== "+FCT+" Test 2 GetMaintenanceByRecordIDStatement ========")
	crnFull2 := []string{"crn:v1:bluemix:public:service-name2:location2::::",
						"crn:v1:aa:dedicated:service-NAME2:location3::::",
				}
	maintenance2 := datastore.MaintenanceInsert {
		SourceCreationTime: "2018-07-08T22:01:00Z",
		PlannedStartTime:   "2018-07-08T21:55:00Z",
		State:              "new",
		Disruptive:         false,
		CRNFull:            crnFull2,
		SourceID:           "source_id2",
		Source:             "test_sourceSN",
		RecordHash:         db.CreateRecordIDFromString("Test 2"),
		CompletionCode:     "completion code 2",
	}

	record_id2, err2,_ := db.InsertMaintenance(database, &maintenance2)
	if err2 != nil {
		log.Println("Insert Maintenance failed: ", err2.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id2)
		log.Println("Insert Maintenance passed")
	}

	log.Println("Test get Maintenance")
	maintenanceReturn2, err2a,_ := db.GetMaintenanceByRecordIDStatement(database, record_id2)
	if err2a != nil {
		log.Println("Get Maintenance failed: ", err2a.Error())
		failed++
	} else {
		log.Println("Get result: ", maintenanceReturn2)

		if maintenanceReturn2.State == maintenance2.State &&
			maintenanceReturn2.Disruptive == maintenance2.Disruptive &&
			len(maintenanceReturn2.CRNFull) == len(maintenance2.CRNFull) &&
			maintenanceReturn2.CRNFull[0] == strings.ToLower(maintenance2.CRNFull[0]) &&
			maintenanceReturn2.CRNFull[1] == strings.ToLower(maintenance2.CRNFull[1]) &&
			maintenanceReturn2.SourceID == maintenance2.SourceID &&
			maintenanceReturn2.Source == maintenance2.Source &&
			maintenanceReturn2.RecordHash == maintenance2.RecordHash &&
			maintenanceReturn2.CompletionCode == maintenance2.CompletionCode &&
			maintenanceReturn2.PnPRemoved == false {

			log.Println("Test 2: Get Maintenance PASSED")
		} else {
			log.Println("Test 2: Get Maintenance FAILED")
			failed++
		}
	}

	log.Println("======== "+FCT+" Test 3 GetMaintenanceByQuery planned_start_start & planned_start_end ========")
	plannedStartStart := "2018-07-07T21:00:00Z"
	plannedStartEnd := "2018-07-07T22:00:00Z"
	queryStr3 := db.MAINTENANCE_QUERY_PLANNED_START_START + "=" + plannedStartStart +"&" +db.MAINTENANCE_QUERY_PLANNED_START_END+"=" + plannedStartEnd
	log.Println("query: ", queryStr3)
	maintenanceReturn3, total_count3, err3,_ := db.GetMaintenanceByQuery(database, queryStr3, 0, 0)
	if err3 != nil {
		log.Println("Get Maintenance failed: ", err3.Error())
		log.Println("Test 3: Get Maintenance FAILED")
		failed++
	} else if total_count3 == 1 {
		log.Println("Get result: ", maintenanceReturn3)
		if len(*maintenanceReturn3) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*maintenanceReturn3))
			log.Println("Test 3: Get Maintenance FAILED")
			failed++
		} else {
			match := 0
			for _,r := range *maintenanceReturn3 {

			if  r.ShortDescription == maintenance.ShortDescription &&
				r.LongDescription == maintenance.LongDescription &&
				r.State == maintenance.State &&
				r.Disruptive == maintenance.Disruptive &&
				len(r.CRNFull) == len(maintenance.CRNFull) &&
				r.CRNFull[0] == strings.ToLower(maintenance.CRNFull[0]) &&
				r.SourceID == maintenance.SourceID &&
				r.Source == maintenance.Source &&
				r.RecordHash == maintenance.RecordHash &&
				r.CompletionCode == maintenance.CompletionCode &&
				r.PnPRemoved == false {

					match++
				}
			}

			if (match <= 0){
				log.Println("There are rows unmatched, failed")
				log.Println("Test 3: Get Maintenance FAILED")
			failed++
			} else {
				log.Println("Test 3: Get Maintenance PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count3)
		log.Println("Test 3: Get Maintenance FAILED")
		failed++
	}

	log.Println("======== "+FCT+" Test 3A GetMaintenanceByQuery planned_start_start & planned_start_end ========")
	plannedEndStart := "2018-07-20T21:00:00Z"
	plannedEndEnd := "2018-07-22T22:00:00Z"
	queryStr3a := db.MAINTENANCE_QUERY_PLANNED_END_START + "=" + plannedEndStart +"&" +db.MAINTENANCE_QUERY_PLANNED_END_END+"=" + plannedEndEnd
	log.Println("query: ", queryStr3a)
	maintenanceReturn3a, total_count3a, err3a,_ := db.GetMaintenanceByQuery(database, queryStr3a, 0, 0)
	if err3a != nil {
		log.Println("Get Maintenance failed: ", err3a.Error())
		log.Println("Test 3a: Get Maintenance FAILED")
		failed++
	} else if total_count3a == 1 {
		log.Println("Get result: ", maintenanceReturn3a)
		if len(*maintenanceReturn3a) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*maintenanceReturn3a))
			log.Println("Test 3a: Get Maintenance FAILED")
			failed++
		} else {
			match := 0
			for _,r := range *maintenanceReturn3a {

			if r.ShortDescription == maintenance.ShortDescription &&
				r.LongDescription == maintenance.LongDescription &&
				r.State == maintenance.State &&
				r.Disruptive == maintenance.Disruptive &&
				len(r.CRNFull) == len(maintenance.CRNFull) &&
				r.CRNFull[0] == strings.ToLower(maintenance.CRNFull[0]) &&
				r.SourceID == maintenance.SourceID &&
				r.Source == maintenance.Source &&
				r.RecordHash == maintenance.RecordHash &&
				r.CompletionCode == maintenance.CompletionCode {

					match++
				}
			}

			if (match <= 0){
				log.Println("There are rows unmatched, failed")
				log.Println("Test 3a: Get Maintenance FAILED")
			failed++
			} else {
				log.Println("Test 3a: Get Maintenance PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count3a)
		log.Println("Test 3a: Get Maintenance FAILED")
		failed++
	}

	log.Println("======== "+FCT+" Test 4 GetMaintenanceByQuery crn ========")
	queryStr4 := db.RESOURCE_QUERY_CRN + "=crn:v1:bluemix:public:service-name2:Location2::::"
	log.Println("query: ", queryStr4)
	maintenanceReturn4, total_count4, err4,_ := db.GetMaintenanceByQuery(database, queryStr4, 0, 0)

	if err4 != nil {
		log.Println("Get Maintenance failed: ", err4.Error())
		log.Println("Test 4: Get Maintenance FAILED")
		failed++
	} else if total_count4 == 1 {
		log.Println("total_count4: ", total_count4)
		log.Println("maintenanceReturn4: ", maintenanceReturn4)

		if len(*maintenanceReturn4) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*maintenanceReturn4))
			log.Println("Test 4: Get Maintenance FAILED")
			failed++
		} else {
			match := 0
			for _,r := range *maintenanceReturn4 {

				if r.State == maintenance2.State &&
					r.Disruptive == maintenance2.Disruptive &&
					len(r.CRNFull) == len(maintenance2.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(maintenance2.CRNFull[0]) &&
					r.CRNFull[1] == strings.ToLower(maintenance2.CRNFull[1]) &&
					r.SourceID == maintenance2.SourceID &&
					r.Source == maintenance2.Source &&
					r.RecordHash == maintenance2.RecordHash &&
					r.CompletionCode == maintenance2.CompletionCode {

					match++
				}
			}

			if (match <= 0){
				log.Println("There are rows unmatched, failed")
				log.Println("Test 4: Get Maintenance FAILED")
				failed++
			} else {
				log.Println("Test 4: Get Maintenance PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count4)
		log.Println("Test 4: Get Maintenance FAILED")
		failed++
	}

	log.Println("======== "+FCT+" Test 5 GetMaintenanceByQuery crn & creation_time_start & creation_time_end ========")

	creationTimeStart5 := "2018-07-08T22:00:00Z"
	creationTimeEnd5 := "2018-07-08T22:02:00Z"
	queryStr5 := db.RESOURCE_QUERY_CRN + "=crn:v1:bluemix:PUBLIC:service-name2:::::" + "&" + db.MAINTENANCE_QUERY_CREATION_TIME_START + "=" + creationTimeStart5 +"&" +db.MAINTENANCE_QUERY_CREATION_TIME_END+"=" + creationTimeEnd5
	log.Println("query: ", queryStr5)
	maintenanceReturn5, total_count5, err5,_ := db.GetMaintenanceByQuery(database, queryStr5, 0, 0)
	log.Println("total_count5: ", total_count5)
	log.Println("maintenanceReturn5: ", maintenanceReturn5)

	if err5 != nil {
		log.Println("Get Maintenance failed: ", err5.Error())
		log.Println("Test 5: Get Maintenance FAILED")
		failed++
	} else if total_count5 == 1 {
		log.Println("total_count5: ", total_count5)
		log.Println("maintenanceReturn5: ", maintenanceReturn5)

		if len(*maintenanceReturn5) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*maintenanceReturn5))
			log.Println("Test 5: Get Maintenance FAILED")
			failed++
		} else {
			match := 0
			for _,r := range *maintenanceReturn5 {

				if r.State == maintenance2.State &&
					r.Disruptive == maintenance2.Disruptive &&
					len(r.CRNFull) == len(maintenance2.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(maintenance2.CRNFull[0]) &&
					r.CRNFull[1] == strings.ToLower(maintenance2.CRNFull[1]) &&
					r.SourceID == maintenance2.SourceID &&
					r.Source == maintenance2.Source &&
					r.RecordHash == maintenance2.RecordHash &&
					r.CompletionCode == maintenance2.CompletionCode {

					match++
				}
			}

			if (match <= 0){
				log.Println("There are rows unmatched, failed")
				log.Println("Test 5: Get Maintenance FAILED")
				failed++
			} else {
				log.Println("Test 5: Get Maintenance PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count5)
		log.Println("Test 5: Get Maintenance FAILED")
		failed++
	}

	log.Println("======== "+FCT+" Test 6 GetMaintenanceByQuery cname & ctype & creation_time_start & creation_time_end  ========")

	log.Println("Test get Maintenance")
	queryStr6 := db.RESOURCE_QUERY_CNAME + "=Bluemix&"+db.RESOURCE_QUERY_CNAME + "=Public&"+db.RESOURCE_QUERY_SERVICE_NAME + "=service-Name1&"+ db.MAINTENANCE_QUERY_CREATION_TIME_START + "=" + creationTimeStart5 +"&" +db.MAINTENANCE_QUERY_CREATION_TIME_END+"=" + creationTimeEnd5
	log.Println("query: ", queryStr6)
	maintenanceReturn6, total_count6, err6,_ := db.GetMaintenanceByQuery(database, queryStr6, 0, 0)
	log.Println("total_count6: ", total_count6)
	log.Println("maintenanceReturn6: ", maintenanceReturn6)

	if err6 != nil {
		log.Println("Get Maintenance failed: ", err6.Error())
		log.Println("Test 6: Get Maintenance FAILED")
		failed++
	} else if total_count6 !=0 {
		log.Printf("Error: total_count: %d, expecting: 0\n", total_count6)
		log.Println("Test 6: Get Maintenance FAILED")
		failed++
	} else if len(*maintenanceReturn6) != 0 {
		log.Printf("Error: Number of returned: %d, expecting: 0\n", len(*maintenanceReturn6))
		log.Println("Test 6: Get Maintenance FAILED")
		failed++
	} else {
		log.Println("Test 6: Get Maintenance PASSED")
	}

	log.Println("======== "+FCT+" Test 6a GetMaintenanceByQuery cname & ctype & creation_time_start & creation_time_end  ========")

	log.Println("Test get Maintenance")
	queryStr6a := db.RESOURCE_QUERY_CNAME + "=Bluemix&"+db.RESOURCE_QUERY_CNAME + "=Public&"+db.RESOURCE_QUERY_SERVICE_NAME + "=service-Name1&"+ db.MAINTENANCE_QUERY_CREATION_TIME_START + "=" + "2018-07-07T22:00:01Z" +"&" +db.MAINTENANCE_QUERY_CREATION_TIME_END+"=" + "2018-07-07T22:02:01Z"
	log.Println("query: ", queryStr6a)
	maintenanceReturn6a, total_count6a, err6a,_ := db.GetMaintenanceByQuery(database, queryStr6a, 0, 0)
	log.Println("total_count6a: ", total_count6)
	log.Println("maintenanceReturn6a: ", maintenanceReturn6a)

	if err6 != nil {
		log.Println("Get Maintenance failed: ", err6a.Error())
		log.Println("Test 6a: Get Maintenance FAILED")
		failed++
	} else if total_count6a !=1 {
		log.Printf("Error: total_count: %d, expecting: 1\n", total_count6a)
		log.Println("Test 6a: Get Maintenance FAILED")
		failed++
	} else if len(*maintenanceReturn6a) != 1 {
		log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*maintenanceReturn6a))
		log.Println("Test 6a: Get Maintenance FAILED")
		failed++
	} else {
		log.Println("Test 6a: Get Maintenance PASSED")
	}

	log.Println("======== "+FCT+" Test 7 Delete Maintenance ========")

	log.Println("Test delete Maintenance")
	err7,_ := db.DeleteMaintenanceStatement(database, record_id)
	if err7 != nil {
		log.Println("Delete Maintenance failed: ", err7.Error())
		failed++
	}
	maintenanceReturn, err7,_ = db.GetMaintenanceByRecordIDStatement(database, record_id)
	if maintenanceReturn != nil {
		log.Println("Delete Maintenance failed, Maintenance still exists: ", err7.Error())
		log.Println("Test 7: Delete Maintenance FAILED")
		failed++
	} else {
		log.Println("Test 7: Delete Maintenance PASSED")
	}


	log.Println("======== "+FCT+" Test 8 DeleteOldCompleteMaintenance ========")

	log.Println("Test insert old Maintenance then delete")
	crnFull8 := []string{"crn:v1:bluemix:public:SERvice-name2:location2::::",
				}

	maintenance8 := datastore.MaintenanceInsert {
		SourceCreationTime: "2018-06-07T22:01:01Z",
		SourceUpdateTime:   "2018-06-07T22:01:01Z",
		PlannedStartTime:   "2018-06-07T21:55:30Z",
		ShortDescription:   "maintenance short description 8",
		LongDescription:    "maintenance long description 8",
		State:              "complete",
		Disruptive:         true,
		CRNFull:            crnFull8,
		SourceID:           "INC00008",
		Source:             "test_source_8",
		RecordHash:         db.CreateRecordIDFromString("Test 8"),
		CompletionCode:     "completion code 8",
	}

	record_id8, err8a,_ := db.InsertMaintenance(database, &maintenance8)
	if err8a != nil {
		log.Println("Insert Maintenance failed: ", err8a.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id8)
		log.Println("Insert Maintenance passed")
	}

	maintenance8b := datastore.MaintenanceInsert {
		SourceCreationTime: "2018-06-07T22:01:01Z",
		SourceUpdateTime:   "2018-06-07T22:01:01Z",
		PlannedStartTime:   "2018-06-07T21:55:30Z",
		ShortDescription:   "maintenance short description 8b",
		LongDescription:    "maintenance long description 8b",
		State:              "new",
		Disruptive:         true,
		CRNFull:            crnFull8,
		SourceID:           "INC00008b",
		Source:             "test_source_8b",
		RecordHash:         db.CreateRecordIDFromString("Test 8b"),
		CompletionCode:     "completion code 8b",
		PnPRemoved:         true,
	}

	record_id8b, err8b,_ := db.InsertMaintenance(database, &maintenance8b)
	if err8b != nil {
		log.Println("Insert Maintenance 8b failed: ", err8b.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id8b)
		log.Println("Insert Maintenance 8b passed")
	}

	log.Println("Testing delete Maintenance older than 30 days")
	err8c,_ := db.DeleteOldCompleteMaintenance(database,30)
	if err8c != nil {
		log.Println("Delete old Maintenance failed: ", err8c.Error())
		failed++
	}

	maintenanceGet8d, _, _ := db.GetMaintenanceByRecordIDStatement(database, record_id8)
	if maintenanceGet8d != nil {
		log.Println("Delete Maintenance failed, Maintenance still exists")
		log.Println("Test 8d: Delete Old Maintenance FAILED")
		failed++
	} else {
		log.Println("Test 8d: Delete Old Maintenance PASSED")
	}

	maintenanceGet8e, _, _ := db.GetMaintenanceByRecordIDStatement(database, record_id8b)
	if maintenanceGet8e != nil {
		log.Println("Delete Maintenance failed, Maintenance still exists")
		log.Println("Test 8e: Delete Old Maintenance FAILED")
		failed++
	} else {
		log.Println("Test 8e: Delete Old Maintenance PASSED")
	}


	log.Println("======== "+FCT+" Test 9 Update Maintenance ========")
	// Insert 2 resources, then insert 1 maintenance with 1 crn. check maintenancejunction has only 1 entry. Update maintenance with 2 crn. check that maintenancejunction has 2 entries
	log.Println("Test Update Maintenance")
	resource9a := datastore.ResourceInsert {
		SourceCreationTime: "2018-07-07T22:01:01Z",
		CRNFull: 			"crn:v1:bluemix:public:service-namemain9a:locationmain9a:scopemain9a:service-instancemain9a:resource-typemain9a:resourcemain9a",
		State:				"ok",
		OperationalStatus: 	"none",
		SourceID:			"resource_idmain9a",
		Source:				"test_sourcemain9a",
		Status:				"ok",
		RegulatoryDomain:   "regulatory domain 1",
		RecordHash:			db.CreateRecordIDFromString("Maintenance Test 9a"),
	}

	resource9b := datastore.ResourceInsert {
		SourceCreationTime: "2018-07-07T22:01:01Z",
		CRNFull: 			"crn:v1:aa:dedicated:service-namemain9b:locationmain9b:scopemain9b:service-instancemain9b:resource-typemain9b:resourcemain9b",
		State:				"ok",
		OperationalStatus: 	"none",
		SourceID:			"resource_idmain9b",
		Source:				"test_sourcemain9b",
		Status:				"ok",
		RegulatoryDomain:   "regulatory domain 2",
		RecordHash:			db.CreateRecordIDFromString("Maintenance Test 9b"),
	}

	record_id9a, err,_ := db.InsertResource(database, &resource9a)
	if err != nil {
	    log.Println("Insert Resource failed: ", err.Error())
	} else {
	    log.Println("record_id: ", record_id9a)
	    log.Println("Insert Resource9a passed")
	}

	record_id9b, err,_ := db.InsertResource(database, &resource9b)
	if err != nil {
	    log.Println("Insert Resource failed: ", err.Error())
	} else {
	    log.Println("record_id: ", record_id9b)
	    log.Println("Insert Resource9b passed")
	}

	maintCRN9a := []string{"crn:v1:bluemix:public:service-namemain9a:locationmain9a::::"}
	maintCRN9b := []string{"crn:v1:bluemix:public:service-namemain9a:locationmain9a::::","crn:v1:aa:dedicated:service-namemain9b:locationmain9b:scopeX:::"}

	maint9 := datastore.MaintenanceInsert {
		SourceCreationTime:    "2018-06-07T22:01:01Z",
		SourceUpdateTime:      "2018-06-07T22:01:01Z",
		PlannedStartTime:      "2018-06-07T22:01:01Z",
		CRNFull:               maintCRN9a,
		State:                 "complete",
		Disruptive:            true,
		SourceID:              "source_id1",
		Source:                "test_source",
		MaintenanceDuration:   300,
		DisruptionType:        "disruptive type",
		DisruptionDescription: "Disruptive Desc",
		DisruptionDuration:    30,
		RecordHash:            db.CreateRecordIDFromString("Maintenance Test"),
		CompletionCode:        "completion code",
	}

	record_idmain9, err,_ := db.InsertMaintenance(database, &maint9)
	if err != nil {
		log.Println("Failed inserting Maintenance")
		failed++
	}

	maintJuncGet, errA,_ := db.GetMaintenanceJunctionByMaintenanceID(database, record_idmain9)
	if errA != nil {
		log.Println("Failed getting MaintenanceJunction")
		failed++
	}
	if len(*maintJuncGet) != 1  {
		log.Println("Should only have 1 entry in MaintenanceJunction")
		failed++
	}

	// update with one good and one bad crn
	maint9.CRNFull = maintCRN9b
	maint9.MaintenanceDuration = 390
	maint9.DisruptionType = "disruptive type 9"
	maint9.DisruptionDescription ="Disruptive Desc 9"
	maint9.DisruptionDuration = 39
	maint9.RecordHash = db.CreateRecordIDFromString("Maintenance Test 9")
	maint9.CompletionCode = "completion code 9"
	maint9.PnPRemoved = true

	err,_ = db.UpdateMaintenance(database, &maint9)
	maintenanceReturn9, err9,_ := db.GetMaintenanceByRecordIDStatement(database, record_idmain9)
	if err9 != nil {
		log.Println("Get Maintenance failed: ", err9.Error())
		failed++
	} else {

		log.Println("Get result: ", maintenanceReturn9)

		if maintenanceReturn9.State == maint9.State &&
			maintenanceReturn9.Disruptive == maint9.Disruptive &&
			len(maintenanceReturn9.CRNFull) == len(maint9.CRNFull) &&
			maintenanceReturn9.CRNFull[0] == strings.ToLower(maint9.CRNFull[0]) &&
			maintenanceReturn9.SourceID == maint9.SourceID &&
			maintenanceReturn9.Source == maint9.Source &&
			maintenanceReturn9.MaintenanceDuration == maint9.MaintenanceDuration &&
			maintenanceReturn9.DisruptionType == maint9.DisruptionType &&
			maintenanceReturn9.DisruptionDescription == maint9.DisruptionDescription &&
			maintenanceReturn9.RecordHash == maint9.RecordHash &&
			maintenanceReturn9.CompletionCode == maint9.CompletionCode &&
			maintenanceReturn9.PnPRemoved == true {

			log.Println("Test 9: Get Maintenance PASSED")
		} else {
			log.Println("Test 9: Get Maintenance FAILED")
			failed++
		}
	}

	maintJuncGet, errA,_ = db.GetMaintenanceJunctionByMaintenanceID(database, record_idmain9)
	if errA != nil {
		log.Println("Failed getting MaintenanceJunction")
		failed++
	}
	if len(*maintJuncGet) != 1  {
		log.Println("Failed: Should have 1 entries in MaintenanceJunction")
		failed++
	} else {
		log.Println("Test 9: Update Maintenance passed")
	}

	// update with added crn
	maintCRN9c := []string{"crn:v1:bluemix:public:service-namemain9a:locationmain9a::::","crn:v1:aa:dedicated:service-namemain9b:locationmain9b::::"}
	maint9.CRNFull = maintCRN9c

	err,_ = db.UpdateMaintenance(database, &maint9)
	maintenanceReturn9b, err9b,_ := db.GetMaintenanceByRecordIDStatement(database, record_idmain9)
	if err9b != nil {
		log.Println("Get Maintenance failed: ", err9b.Error())
		failed++
	} else {

		log.Println("Get result: ", maintenanceReturn9)

		if maintenanceReturn9b.State == maint9.State &&
			maintenanceReturn9b.Disruptive == maint9.Disruptive &&
			len(maintenanceReturn9b.CRNFull) == len(maint9.CRNFull) &&
			maintenanceReturn9b.CRNFull[0] == strings.ToLower(maint9.CRNFull[0]) &&
			maintenanceReturn9b.SourceID == maint9.SourceID &&
			maintenanceReturn9b.Source == maint9.Source &&
			maintenanceReturn9b.MaintenanceDuration == maint9.MaintenanceDuration &&
			maintenanceReturn9b.DisruptionType == maint9.DisruptionType &&
			maintenanceReturn9b.DisruptionDescription == maint9.DisruptionDescription &&
			maintenanceReturn9b.RecordHash == maint9.RecordHash &&
			maintenanceReturn9b.CompletionCode == maint9.CompletionCode &&
			maintenanceReturn9b.PnPRemoved == true {

			log.Println("Test 9b: Get Maintenance PASSED")
		} else {
			log.Println("Test 9b: Get Maintenance FAILED")
			failed++
		}
	}

	maintJuncGet, errA,_ = db.GetMaintenanceJunctionByMaintenanceID(database, record_idmain9)
	if errA != nil {
		log.Println("Failed getting MaintenanceJunction")
		failed++
	}
	if len(*maintJuncGet) != 2  {
		log.Println("Failed: Should have 2 entries in MaintenanceJunction, len(*maintJuncGet)=", len(*maintJuncGet))
		failed++
	} else {
		log.Println("Test 9b: Update Maintenance passed")
	}

	log.Println("======== "+FCT+" Test 10 UpdateMaintenance with pnp_removed true, UpdateMaintenance, GetMaintenanceByQuery, GetAllSNMaintenances ========")

	crnFull10 := []string{"crn:v1:bluemix:public:service-name2:location2::::",
						"crn:v1:aa:dedicated:service-NAME2:location3::::",
				}
	maint10 := datastore.MaintenanceInsert {
		SourceCreationTime: "2019-01-08T22:01:00Z",
		PlannedStartTime:   "2019-01-08T21:55:00Z",
		State:              "new",
		Disruptive:         false,
		CRNFull:            crnFull10,
		SourceID:           "source_id10",
		Source:             "servicenow",
		RecordHash:         db.CreateRecordIDFromString("Test 10"),
		CompletionCode:     "completion code 10",
		PnPRemoved:         false,
	}

	record_id10, err10,_ := db.InsertMaintenance(database, &maint10)
	if err10 != nil {
		log.Println("Insert Maintenance failed: ", err10.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id10)
		log.Println("Insert Maintenance passed")
	}

	creationTimeStart10 := "2019-01-08T22:00:00Z"
	creationTimeEnd10 := "2019-01-08T22:02:00Z"
	queryStr10 := db.RESOURCE_QUERY_CRN + "=crn:v1:bluemix:public:service-name2:::::" + "&" + db.MAINTENANCE_QUERY_CREATION_TIME_START + "=" + creationTimeStart10 +"&" +db.MAINTENANCE_QUERY_CREATION_TIME_END+"=" + creationTimeEnd10
	log.Println("query: ", queryStr10)
	maintenanceReturn10, total_count10, err10,_ := db.GetMaintenanceByQuery(database, queryStr10, 0, 0)
	log.Println("total_count10: ", total_count10)
	log.Println("maintenanceReturn10: ", maintenanceReturn10)

	if err10 != nil {
		log.Println("Get Maintenance failed: ", err10.Error())
		log.Println("Test 10: Get Maintenance FAILED")
		failed++
	} else if total_count10 == 1 {

		if len(*maintenanceReturn10) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*maintenanceReturn10))
			log.Println("Test 10: Get Maintenance FAILED")
			failed++
		} else {
			match := 0
			for _,r := range *maintenanceReturn10 {

				if r.State == maint10.State &&
					r.Disruptive == maint10.Disruptive &&
					len(r.CRNFull) == len(maint10.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(maint10.CRNFull[0]) &&
					r.CRNFull[1] == strings.ToLower(maint10.CRNFull[1]) &&
					r.SourceID == maint10.SourceID &&
					r.Source == maint10.Source &&
					r.RecordHash == maint10.RecordHash &&
					r.CompletionCode == maint10.CompletionCode &&
					r.PnPRemoved == false {

					match++
				}
			}

			if (match <= 0){
				log.Println("There are rows unmatched, failed")
				log.Println("Test 10: Get Maintenance FAILED")
				failed++
			} else {
				log.Println("Test 10: Get Maintenance PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count10)
		log.Println("Test 10: Get Maintenance FAILED")
		failed++
	}
	

	mReturn10b, error10b, _ := db.GetAllSNMaintenances(database)
	if error10b != nil {
		log.Println("Get Maintenance failed: ", error10b.Error())
		failed++
	} else {

		log.Println("Get result: ", mReturn10b)
		if len(mReturn10b) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(mReturn10b))
			log.Println("Test 10b: Get Maintenance FAILED")
			failed++
		} else {
			match := 0
			for _, r := range mReturn10b {

				if r.RecordID == record_id10 &&
					r.SourceID == maint10.SourceID &&
					r.Source == maint10.Source &&
					r.RecordHash == maint10.RecordHash {

					match++
				}
			}

			if (match <= 0){
				log.Println("There are rows unmatched, failed")
				log.Println("Test 10b: Get Maintenance FAILED")
				failed++
			} else {
				log.Println("Test 10b: Get Maintenance PASSED")
			}
		}
	}

	// Update PnPRemoved to true
	maint10.PnPRemoved = true
	err10a,_ := db.UpdateMaintenance(database, &maint10)
	if err10a != nil {
		log.Println("Failed in UpdateMaintenance")
		failed++
	}

	maintReturn10, err10b,_ := db.GetMaintenanceByRecordIDStatement(database, record_id10)
	if err10b != nil {
		log.Println("Get Maintenance failed: ", err10b.Error())
		failed++
	} else {

		log.Println("Get result: ", maintReturn10)

		if maintReturn10.State == maint10.State &&
			maintReturn10.Disruptive == maint10.Disruptive &&
			len(maintReturn10.CRNFull) == len(maint10.CRNFull) &&
			maintReturn10.CRNFull[0] == strings.ToLower(maint10.CRNFull[0]) &&
			maintReturn10.SourceID == maint10.SourceID &&
			maintReturn10.Source == maint10.Source &&
			maintReturn10.MaintenanceDuration == maint10.MaintenanceDuration &&
			maintReturn10.DisruptionType == maint10.DisruptionType &&
			maintReturn10.DisruptionDescription == maint10.DisruptionDescription &&
			maintReturn10.RecordHash == maint10.RecordHash &&
			maintReturn10.CompletionCode == maint10.CompletionCode &&
			maintReturn10.PnPRemoved == true {

			log.Println("Test10c: Update Maintenance passed")
		} else {
			log.Println("Test10c: Update Maintenance failed")
			failed++
		}
	}

	// query PnPRemoved=true
	queryStr10d := db.RESOURCE_QUERY_CRN + "=crn:v1:bluemix:public:service-name2:::::" + "&" + db.MAINTENANCE_QUERY_CREATION_TIME_START + "=" + creationTimeStart10 +"&" +db.MAINTENANCE_QUERY_CREATION_TIME_END+"=" + creationTimeEnd10 +"&"+db.MAINTENANCE_COLUMN_PNP_REMOVED+"=true"
	log.Println("query: ", queryStr10d)
	maintReturn10d, total_count10d, err10d,_ := db.GetMaintenanceByQuery(database, queryStr10d, 0, 0)
	log.Println("total_count10d: ", total_count10d)
	log.Println("maintReturn10d: ", maintReturn10d)

	if err10d != nil {
		log.Println("Get Maintenance failed: ", err10d.Error())
		log.Println("Test 10d: Get Maintenance FAILED")
		failed++
	} else if total_count10d == 1 {

		if len(*maintReturn10d) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*maintReturn10d))
			log.Println("Test 10d: Get Maintenance FAILED")
			failed++
		} else {
			match := 0
			for _,r := range *maintReturn10d {

				if r.State == maint10.State &&
					r.Disruptive == maint10.Disruptive &&
					len(r.CRNFull) == len(maint10.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(maint10.CRNFull[0]) &&
					r.SourceID == maint10.SourceID &&
					r.Source == maint10.Source &&
					r.MaintenanceDuration == maint10.MaintenanceDuration &&
					r.DisruptionType == maint10.DisruptionType &&
					r.DisruptionDescription == maint10.DisruptionDescription &&
					r.RecordHash == maint10.RecordHash &&
					r.CompletionCode == maint10.CompletionCode &&
					r.PnPRemoved == true {

					match++
				}
			}

			if (match <= 0){
				log.Println("There are rows unmatched, failed")
				log.Println("Test 10d: Get Maintenance FAILED")
				failed++
			} else {
				log.Println("Test 10d: Get Maintenance PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count10d)
		log.Println("Test 10d: Get Maintenance FAILED")
		failed++
	}

	queryStr10e := db.RESOURCE_QUERY_CRN + "=crn:v1:bluemix:public:service-name2:::::" + "&" + db.MAINTENANCE_QUERY_CREATION_TIME_START + "=" + creationTimeStart10 +"&" +db.MAINTENANCE_QUERY_CREATION_TIME_END+"=" + creationTimeEnd10
	log.Println("query: ", queryStr10e)
	maintReturn10e, total_count10e, err10e,_ := db.GetMaintenanceByQuery(database, queryStr10e, 0, 0)
	log.Println("total_count10e: ", total_count10e)
	log.Println("maintReturn10e: ", maintReturn10e)

	if err10e != nil {
		log.Println("Get Maintenance failed: ", err10e.Error())
		log.Println("Test 10e: Get Maintenance FAILED")
		failed++
	} else if total_count10e != 0 {

		log.Printf("total_count is %d, expecting 0, failed", total_count10e)
		log.Println("Test 10e: Get Maintenance FAILED")
		failed++
	} else {
		log.Println("Test 10e: Get Maintenance PASSED")
	}





	log.Println("**** SUMMARY ****")
	if failed > 0 {
		log.Println("**** "+FCT+"FAILED ****")
	} else {
		log.Println("**** "+FCT+"PASSED ****")
	}
	return failed
}
