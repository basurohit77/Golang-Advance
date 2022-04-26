package test

import (
	"database/sql"
	"log"
	"strings"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
)

func TestIncidentForGaas(database *sql.DB, gaasServices []string) int {
	FCT := "TestIncidentForGaas "
	failed := 0

	log.Println("======== " + FCT + " Test 1 Insert Incident ========")
	crnFull := []string{"crn:v1:watson:public:" + gaasServices[0] + ":location1::::"}

	incident := datastore.IncidentInsert{
		SourceCreationTime:        "2018-07-07T22:01:01Z",
		SourceUpdateTime:          "2018-07-07T22:01:01Z",
		OutageStartTime:           "2018-07-07T21:55:30Z",
		ShortDescription:          "incident short description 1",
		LongDescription:           "incident long description 1",
		State:                     "new",
		Classification:            "potential-cie",
		Severity:                  "1",
		CRNFull:                   crnFull,
		SourceID:                  "INC00001",
		Source:                    "gaas_test_source_1",
		RegulatoryDomain:          "regulatory domain 1",
		AffectedActivity:          "affected activity 1",
		CustomerImpactDescription: "customer impact description 1",
	}

	record_id, err, _ := db.InsertIncident(database, &incident)
	if err != nil {
		log.Println("Insert Incident failed: ", err.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id)
		log.Println("Insert Incident passed")
	}

	log.Println("Test get Incident")
	incidentReturn, err1, _ := db.GetIncidentByRecordIDStatement(database, record_id)
	if err1 != nil {
		log.Println("Get Incident failed: ", err1.Error())
		failed++
	} else {

		log.Println("Get result: ", incidentReturn)

		if incidentReturn.ShortDescription == incident.ShortDescription &&
			incidentReturn.LongDescription == incident.LongDescription &&
			incidentReturn.State == incident.State &&
			incidentReturn.Classification == incident.Classification &&
			incidentReturn.Severity == incident.Severity &&
			len(incidentReturn.CRNFull) == len(incident.CRNFull) &&
			incidentReturn.CRNFull[0] == strings.ToLower(incident.CRNFull[0]) &&
			incidentReturn.SourceID == incident.SourceID &&
			incidentReturn.Source == incident.Source &&
			incidentReturn.RegulatoryDomain == incident.RegulatoryDomain &&
			incidentReturn.AffectedActivity == incident.AffectedActivity &&
			incidentReturn.CustomerImpactDescription == incident.CustomerImpactDescription &&
			incidentReturn.PnPRemoved == false {

			log.Println("Test 1: Get Incident PASSED")
		} else {
			log.Println("Test 1: Get Incident FAILED")
			failed++
		}

	}

	log.Println("======== " + FCT + " Test 2 GetIncidentByRecordID ========")
	crnFull2 := []string{"crn:v1::aa:" + gaasServices[1] + ":location2::::",
		"crn:v1:watson:aa:" + gaasServices[1] + ":location3::::",
	}

	incident2 := datastore.IncidentInsert{
		SourceCreationTime:        "2018-07-08T22:01:00Z",
		OutageStartTime:           "2018-07-08T21:55:00Z",
		State:                     "new",
		Classification:            "confirmed-cie",
		Severity:                  "2",
		CRNFull:                   crnFull2,
		SourceID:                  "gaas_test_source_id2",
		Source:                    "SN",
		AffectedActivity:          "affected activity 2",
		CustomerImpactDescription: "customer impact description 2",
	}

	record_id2, err2, _ := db.InsertIncident(database, &incident2)
	if err2 != nil {
		log.Println("Insert Incident failed: ", err2.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id2)
		log.Println("Insert Incident passed")
	}

	log.Println("Test get Incident")
	incidentReturn2, err2a, _ := db.GetIncidentByRecordIDStatement(database, record_id2)
	if err2a != nil {
		log.Println("Get Incident failed: ", err2a.Error())
		failed++
	} else {
		log.Println("Get result: ", incidentReturn2)

		if incidentReturn2.Classification == incident2.Classification &&
			incidentReturn2.Severity == incident2.Severity &&
			len(incidentReturn2.CRNFull) == len(incident2.CRNFull) &&
			incidentReturn2.CRNFull[0] == strings.ToLower(incident2.CRNFull[0]) &&
			incidentReturn2.CRNFull[1] == strings.ToLower(incident2.CRNFull[1]) &&
			incidentReturn2.SourceID == incident2.SourceID &&
			incidentReturn2.Source == incident2.Source &&
			incidentReturn2.AffectedActivity == incident2.AffectedActivity &&
			incidentReturn2.CustomerImpactDescription == incident2.CustomerImpactDescription {

			log.Println("Test 2: Get Incident PASSED")
		} else {
			log.Println("Test 2: Get Incident FAILED")
			failed++
		}
	}

	log.Println("======== " + FCT + " Test 3 GetIncidentByQuery creation_time_start & creation_time_end ========")
	creationTimeStart := "2018-07-07T22:01:00Z"
	creationTimeEnd := "2018-07-07T22:02:00Z"
	queryStr3 := db.INCIDENT_QUERY_CREATION_TIME_START + "=" + creationTimeStart + "&" + db.INCIDENT_QUERY_CREATION_TIME_END + "=" + creationTimeEnd
	log.Println("query: ", queryStr3)
	incidentReturn3, total_count3, err3, _ := db.GetIncidentByQuery(database, queryStr3, 0, 0)
	if err3 != nil {
		log.Println("Get Incident failed: ", err3.Error())
		log.Println("Test 3: Get Incident FAILED")
		failed++
	} else if total_count3 == 1 {
		log.Println("Get result: ", incidentReturn3)
		if len(*incidentReturn3) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*incidentReturn3))
			log.Println("Test 3: Get Incident FAILED")
			failed++
		} else {
			match := 0
			for _, r := range *incidentReturn3 {

				if r.ShortDescription == incident.ShortDescription &&
					r.LongDescription == incident.LongDescription &&
					r.State == incident.State &&
					r.Classification == incident.Classification &&
					r.Severity == incident.Severity &&
					len(r.CRNFull) == len(incident.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(incident.CRNFull[0]) &&
					r.SourceID == incident.SourceID &&
					r.Source == incident.Source &&
					r.RegulatoryDomain == incident.RegulatoryDomain &&
					r.AffectedActivity == incident.AffectedActivity &&
					r.CustomerImpactDescription == incident.CustomerImpactDescription &&
					r.PnPRemoved == false {

					match++
				}
			}

			if match <= 0 {
				log.Println("There are rows unmatched, failed")
				log.Println("Test 3: Get Incident FAILED")
				failed++
			} else {
				log.Println("Test 3: Get Incident PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count3)
		log.Println("Test 3: Get Incident FAILED")
		failed++
	}

	log.Println("======== " + FCT + " Test 4 GetIncidentByQuery crn ========")
	queryStr4 := db.RESOURCE_QUERY_CRN + "=crn:v1:watson:public:" + gaasServices[1] + ":location2::::"
	log.Println("query: ", queryStr4)
	incidentReturn4, total_count4, err4, _ := db.GetIncidentByQuery(database, queryStr4, 0, 0)

	if err4 != nil {
		log.Println("Get Incident failed: ", err4.Error())
		log.Println("Test 4: Get Incident FAILED")
		failed++
	} else if total_count4 == 1 {
		log.Println("total_count4: ", total_count4)
		log.Println("incidentReturn4: ", incidentReturn4)

		if len(*incidentReturn4) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*incidentReturn4))
			log.Println("Test 4: Get Incident FAILED")
			failed++
		} else {
			match := 0
			for _, r := range *incidentReturn4 {

				if r.State == incident2.State &&
					r.Classification == incident2.Classification &&
					r.Severity == incident2.Severity &&
					len(r.CRNFull) == len(incident2.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(incident2.CRNFull[0]) &&
					r.SourceID == incident2.SourceID &&
					r.Source == incident2.Source {

					match++
				}
			}

			if match <= 0 {
				log.Println("There are rows unmatched, failed")
				log.Println("Test 4: Get Incident FAILED")
				failed++
			} else {
				log.Println("Test 4: Get Incident PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count4)
		log.Println("Test 4: Get Incident FAILED")
		failed++
	}

	log.Println("======== " + FCT + " Test 5 GetIncidentByQuery crn & creation_time_start & creation_time_end ========")

	creationTimeStart5 := "2018-07-08T22:00:00Z"
	creationTimeEnd5 := "2018-07-08T22:02:00Z"
	queryStr5 := db.RESOURCE_QUERY_CRN + "=crn:v1:::LIBERTY-for-java:::::" + "&" + db.INCIDENT_QUERY_CREATION_TIME_START + "=" + creationTimeStart5 + "&" + db.INCIDENT_QUERY_CREATION_TIME_END + "=" + creationTimeEnd5
	log.Println("query: ", queryStr5)
	incidentReturn5, total_count5, err5, _ := db.GetIncidentByQuery(database, queryStr5, 0, 0)
	log.Println("total_count5: ", total_count5)
	log.Println("incidentReturn5: ", incidentReturn5)

	if err5 != nil {
		log.Println("Get Incident failed: ", err5.Error())
		log.Println("Test 5: Get Incident FAILED")
		failed++
	} else if total_count5 == 1 {
		log.Println("total_count5: ", total_count5)
		log.Println("incidentReturn5: ", incidentReturn5)

		if len(*incidentReturn5) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*incidentReturn5))
			log.Println("Test 5: Get Incident FAILED")
			failed++
		} else {
			match := 0
			for _, r := range *incidentReturn5 {

				if r.State == incident2.State &&
					r.Classification == incident2.Classification &&
					r.Severity == incident2.Severity &&
					len(r.CRNFull) == len(incident2.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(incident2.CRNFull[0]) &&
					r.SourceID == incident2.SourceID &&
					r.Source == incident2.Source {

					match++
				}
			}

			if match <= 0 {
				log.Println("There are rows unmatched, failed")
				log.Println("Test 5: Get Incident FAILED")
				failed++
			} else {
				log.Println("Test 5: Get Incident PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count5)
		log.Println("Test 5: Get Incident FAILED")
		failed++
	}

	log.Println("======== " + FCT + " Test 6 GetIncidentByQuery cname & ctype & service-name & creation_time_start & creation_time_end ========")

	log.Println("Test get Incident")
	queryStr6 := db.RESOURCE_QUERY_CNAME + "=&" + db.RESOURCE_QUERY_CNAME + "=&" + db.RESOURCE_QUERY_SERVICE_NAME + "=" + gaasServices[0] + "&" + db.INCIDENT_QUERY_CREATION_TIME_START + "=" + creationTimeStart5 + "&" + db.INCIDENT_QUERY_CREATION_TIME_END + "=" + creationTimeEnd5
	log.Println("query: ", queryStr6)
	incidentReturn6, total_count6, err6, _ := db.GetIncidentByQuery(database, queryStr6, 0, 0)
	log.Println("total_count6: ", total_count6)
	log.Println("incidentReturn6: ", incidentReturn6)

	if err6 != nil {
		log.Println("Get Incident failed: ", err6.Error())
		log.Println("Test 6: Get Incident FAILED")
		failed++
	} else if total_count6 != 0 {
		log.Printf("Error: total_count: %d, expecting: 0\n", total_count6)
		log.Println("Test 6: Get Incident FAILED")
		failed++
	} else if len(*incidentReturn6) != 0 {
		log.Printf("Error: Number of returned: %d, expecting: 0\n", len(*incidentReturn6))
		log.Println("Test 6: Get Incident FAILED")
		failed++
	} else {
		log.Println("Test 6: Get Incident PASSED")
	}

	log.Println("======== " + FCT + " Test 7 DeleteIncidentStatement ========")

	log.Println("Test delete Incident")
	err7, _ := db.DeleteIncidentStatement(database, record_id)
	if err7 != nil {
		log.Println("Delete Incident failed: ", err7.Error())
		log.Println("Test 7: Delete Incident FAILED")
		failed++
	}
	incidentReturn, err7, _ = db.GetIncidentByRecordIDStatement(database, record_id)
	if incidentReturn != nil {
		log.Println("Delete Incident failed, Incident still exists: ", err7.Error())
		log.Println("Test 7: Delete Incident FAILED")
		failed++
	} else {
		log.Println("Test 7: Delete Incident PASSED")
	}

	log.Println("======== " + FCT + " Test 8 DeleteOldResolvedIncidents ========")

	log.Println("Test insert old Incident then delete")
	crnFull8 := []string{"crn:v1:watson:aa:" + gaasServices[1] + ":LOCATION2::::"}

	incident8 := datastore.IncidentInsert{
		SourceCreationTime: "2018-06-07T22:01:01Z",
		SourceUpdateTime:   "2018-06-07T22:01:01Z",
		OutageStartTime:    "2018-05-07T21:55:30Z",
		ShortDescription:   "incident short description 8",
		LongDescription:    "incident long description 8",
		State:              "resolved",
		Classification:     "potential-cie",
		Severity:           "1",
		CRNFull:            crnFull8,
		SourceID:           "INC00008",
		Source:             "gaas_test_source_8",
		RegulatoryDomain:   "regulatory domain 8",
	}

	record_id8, err8a, _ := db.InsertIncident(database, &incident8)
	if err8a != nil {
		log.Println("Insert Incident failed: ", err8a.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id8)
		log.Println("Insert Incident passed")
	}

	incident8b := datastore.IncidentInsert{
		SourceCreationTime: "2018-06-07T22:01:01Z",
		SourceUpdateTime:   "2018-06-07T22:01:01Z",
		OutageStartTime:    "2018-05-07T21:55:30Z",
		ShortDescription:   "incident short description 8",
		LongDescription:    "incident long description 8",
		State:              "new",
		Classification:     "potential-cie",
		Severity:           "1",
		CRNFull:            crnFull8,
		SourceID:           "INC00008b",
		Source:             "gaas_test_source_8b",
		RegulatoryDomain:   "regulatory domain 8",
		PnPRemoved:         true,
	}

	record_id8b, err8b, _ := db.InsertIncident(database, &incident8b)
	if err8b != nil {
		log.Println("Insert Incident failed: ", err8b.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id8b)
		log.Println("Insert Incident passed")
	}

	log.Println("Testing delete incidents older than 30 days")
	err8c, _ := db.DeleteOldResolvedIncidents(database, 30)
	if err8c != nil {
		log.Println("Delete old incidents failed: ", err8c.Error())
		failed++
	}
	incidentGet8d, _, _ := db.GetIncidentByRecordIDStatement(database, record_id8)
	if incidentGet8d != nil {
		log.Println("Delete Incident failed, Incident still exists")
		log.Println("Test 8d: Delete Old Incident FAILED")
		failed++
	} else {
		log.Println("Test 8d: Delete Old Incident PASSED")
	}

	incidentGet8e, _, _ := db.GetIncidentByRecordIDStatement(database, record_id8b)
	if incidentGet8e != nil {
		log.Println("Delete Incident failed, Incident still exists")
		log.Println("Test 8e: Delete Old Incident FAILED")
		failed++
	} else {
		log.Println("Test 8e: Delete Old Incident PASSED")
	}

	log.Println("======== " + FCT + " Test 9 GetIncidentJunctionByIncidentID, UpdateIncident========")
	// Insert 2 resources, then insert 1 incident with 1 crn. check incidentjunction has only 1 entry. Update incident with 2 crn. check that incidentjunction has 2 entries
	log.Println("Test Update Incident")
	resource9a := datastore.ResourceInsert{
		CRNFull:           "crn:v1:::" + gaasServices[2] + ":::::",
		State:             "ok",
		OperationalStatus: "none",
		SourceID:          "resource_id9a",
		Source:            "gaas_test_source9a",
		Status:            "ok",
		RegulatoryDomain:  "regulatory domain 1",
		RecordHash:        db.CreateRecordIDFromString("Incident Test 9a"),
	}

	resource9b := datastore.ResourceInsert{
		CRNFull:           "crn:v1:::" + gaasServices[3] + ":::::",
		State:             "ok",
		OperationalStatus: "none",
		SourceID:          "resource_id9b",
		Source:            "gaas_test_source9b",
		Status:            "ok",
		RegulatoryDomain:  "regulatory domain 9b",
		RecordHash:        db.CreateRecordIDFromString("Incident Test 9b"),
	}

	record_id9a, err, _ := db.InsertResource(database, &resource9a)
	if err != nil {
		log.Println("Insert Resource failed: ", err.Error())
	} else {
		log.Println("record_id: ", record_id9a)
		log.Println("Insert Resource9a passed")
	}

	record_id9b, err, _ := db.InsertResource(database, &resource9b)
	if err != nil {
		log.Println("Insert Resource failed: ", err.Error())
	} else {
		log.Println("record_id: ", record_id9b)
		log.Println("Insert Resource9b passed")
	}

	incidentCRN9a := []string{"crn:v1::aa:" + gaasServices[2] + ":location9a::::"}
	incidentCRN9b := []string{"crn:v1:watson::" + gaasServices[2] + ":location9a::::", "crn:v1:watson::" + gaasServices[3] + ":location9b:scope9b:::"}

	incident9 := datastore.IncidentInsert{
		SourceCreationTime:        "2018-06-07T22:01:01Z",
		SourceUpdateTime:          "2018-06-07T22:01:01Z",
		OutageStartTime:           "2018-06-07T22:01:01Z",
		ShortDescription:          "",
		LongDescription:           "",
		State:                     "resolved",
		Classification:            "normal",
		Severity:                  "1",
		CRNFull:                   incidentCRN9a,
		SourceID:                  "source_id_incident9",
		Source:                    "gaas_test_source",
		AffectedActivity:          "affected activity 9a",
		CustomerImpactDescription: "customer impact description 9a",
	}

	record_idinc9, err, _ := db.InsertIncident(database, &incident9)

	incidentJuncGet, errA, _ := db.GetIncidentJunctionByIncidentID(database, record_idinc9)
	if errA != nil {
		log.Println("Failed getting IncidentJunction")
		failed++
	}
	if len(*incidentJuncGet) != 1 {
		log.Println("Failed: Should only have 1 entry in IncidentJunction, len(*incidentJuncGet): ", len(*incidentJuncGet))
		failed++
	}

	// modify incident9 data
	incident9.CRNFull = incidentCRN9b
	incident9.AffectedActivity = "affected activity 9b"
	incident9.CustomerImpactDescription = "customer impact description 9b"

	err, _ = db.UpdateIncident(database, &incident9)
	if err != nil {
		log.Println("Failed in UpdateIncident")
		failed++
	}

	incidentReturn9b, err9b, _ := db.GetIncidentByRecordIDStatement(database, record_idinc9)
	if err9b != nil {
		log.Println("Get Incident failed: ", err9b.Error())
		failed++
	} else {

		log.Println("Get result: ", incidentReturn9b)

		if incidentReturn9b.ShortDescription == incident9.ShortDescription &&
			incidentReturn9b.LongDescription == incident9.LongDescription &&
			incidentReturn9b.State == incident9.State &&
			incidentReturn9b.Classification == incident9.Classification &&
			incidentReturn9b.Severity == incident9.Severity &&
			len(incidentReturn9b.CRNFull) == len(incident9.CRNFull) &&
			incidentReturn9b.CRNFull[0] == strings.ToLower(incident9.CRNFull[0]) &&
			incidentReturn9b.CRNFull[1] == strings.ToLower(incident9.CRNFull[1]) &&
			incidentReturn9b.SourceID == incident9.SourceID &&
			incidentReturn9b.Source == incident9.Source &&
			incidentReturn9b.AffectedActivity == incident9.AffectedActivity &&
			incidentReturn9b.CustomerImpactDescription == incident9.CustomerImpactDescription {

			log.Println("Update Incident passed")
		} else {
			log.Println("Update Incident failed")
			failed++
		}
	}

	incidentJuncGet, errA, _ = db.GetIncidentJunctionByIncidentID(database, record_idinc9)
	if errA != nil {
		log.Println("Failed getting IncidentJunction")
		failed++
	}
	if len(*incidentJuncGet) != 2 {
		log.Println("Failed: Should have 2 entries in IncidentJunction, len(*incidentJuncGet): ", len(*incidentJuncGet))
		failed++
	} else {
		log.Println("Update Incident Add CRN passed")
	}

	log.Println("======== " + FCT + " Test 10 UpdateIncident with pnp_removed true, UpdateIncident, GetIncidentByQuery ========")
	creationTimeStart10 := "2019-06-08T22:00:00Z"
	creationTimeEnd10 := "2019-06-08T22:02:00Z"

	incident10 := datastore.IncidentInsert{
		SourceCreationTime:        "2019-06-08T22:01:00Z",
		OutageStartTime:           "2019-06-08T21:55:00Z",
		State:                     "new",
		Classification:            "confirmed-cie",
		Severity:                  "2",
		CRNFull:                   crnFull2,
		SourceID:                  "source_id10",
		Source:                    "gaas_test_sourceSN",
		AffectedActivity:          "affected activity 10",
		CustomerImpactDescription: "customer impact description 10",
		PnPRemoved:                false,
	}
	record_id10, err10, _ := db.InsertIncident(database, &incident10)
	if err10 != nil {
		log.Println("Insert Incident failed: ", err10.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id10)
		log.Println("Insert Incident passed")
	}

	queryStr10a := db.RESOURCE_QUERY_CRN + "=crn:v1::public:" + gaasServices[1] + ":::::" + "&" + db.INCIDENT_QUERY_CREATION_TIME_START + "=" + creationTimeStart10 + "&" + db.INCIDENT_QUERY_CREATION_TIME_END + "=" + creationTimeEnd10
	log.Println("query: ", queryStr10a)
	incidentReturn10a, total_count10a, err10a, _ := db.GetIncidentByQuery(database, queryStr10a, 0, 0)
	log.Println("total_count10a: ", total_count10a)
	log.Println("incidentReturn10a: ", incidentReturn10a)

	if err10a != nil {
		log.Println("Get Incident failed: ", err10a.Error())
		log.Println("Test 10a: Get Incident FAILED")
		failed++
	} else if total_count10a != 1 {

		log.Printf("total_count is %d, expecting 1, failed", total_count10a)
		log.Println("Test 10a: Get Incident FAILED")
		failed++
	} else {
		log.Println("Get result: ", incidentReturn10a)

		match := 0
		for _, r := range *incidentReturn10a {

			if r.State == incident10.State &&
				r.Classification == incident10.Classification &&
				r.Severity == incident10.Severity &&
				len(r.CRNFull) == len(incident10.CRNFull) &&
				r.CRNFull[0] == strings.ToLower(incident10.CRNFull[0]) &&
				r.CRNFull[1] == strings.ToLower(incident10.CRNFull[1]) &&
				r.SourceID == incident10.SourceID &&
				r.Source == incident10.Source &&
				r.AffectedActivity == incident10.AffectedActivity &&
				r.CustomerImpactDescription == incident10.CustomerImpactDescription &&
				r.PnPRemoved == incident10.PnPRemoved {

				match++
			}
		}

		if match <= 0 {
			log.Println("There are rows unmatched, failed")
			log.Println("Test 10a: Get Incident FAILED")
			failed++
		} else {
			log.Println("Test 10a: Get Incident PASSED")
		}
	}

	// Update PnPRemoved to true
	incident10.PnPRemoved = true
	err10a, _ = db.UpdateIncident(database, &incident10)
	if err10a != nil {
		log.Println("Failed in UpdateIncident")
		failed++
	}

	incidentReturn10, err10b, _ := db.GetIncidentByRecordIDStatement(database, record_id10)
	if err10b != nil {
		log.Println("Get Incident failed: ", err10b.Error())
		failed++
	} else {

		log.Println("Get result: ", incidentReturn10)

		if incidentReturn10.State == incident10.State &&
			incidentReturn10.Classification == incident10.Classification &&
			incidentReturn10.Severity == incident10.Severity &&
			len(incidentReturn10.CRNFull) == len(incident10.CRNFull) &&
			incidentReturn10.CRNFull[0] == strings.ToLower(incident10.CRNFull[0]) &&
			incidentReturn10.CRNFull[1] == strings.ToLower(incident10.CRNFull[1]) &&
			incidentReturn10.SourceID == incident10.SourceID &&
			incidentReturn10.Source == incident10.Source &&
			incidentReturn10.AffectedActivity == incident10.AffectedActivity &&
			incidentReturn10.CustomerImpactDescription == incident10.CustomerImpactDescription &&
			incidentReturn10.PnPRemoved == incident10.PnPRemoved {

			log.Println("Test10: Update Incident passed")
		} else {
			log.Println("Test10: Update Incident failed")
			failed++
		}
	}

	// GetIncidentByQuery pnp_removed=true
	queryStr10b := db.RESOURCE_QUERY_CRN + "=crn:v1:::" + gaasServices[1] + ":::::" + "&" + db.INCIDENT_QUERY_CREATION_TIME_START + "=" + creationTimeStart10 + "&" + db.INCIDENT_QUERY_CREATION_TIME_END + "=" + creationTimeEnd10 + "&" + db.INCIDENT_COLUMN_PNP_REMOVED + "=true"
	log.Println("query: ", queryStr10b)
	incidentReturn10b, total_count10b, err10b, _ := db.GetIncidentByQuery(database, queryStr10b, 0, 0)
	log.Println("total_count10b: ", total_count10b)
	log.Println("incidentReturn10b: ", incidentReturn10b)

	if err10b != nil {
		log.Println("Get Incident failed: ", err10b.Error())
		log.Println("Test 10b: Get Incident FAILED")
		failed++
	} else if total_count10b == 1 {
		log.Println("total_count10b: ", total_count10b)
		log.Println("incidentReturn10b: ", incidentReturn10b)

		if len(*incidentReturn10b) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*incidentReturn10b))
			log.Println("Test 10b: Get Incident FAILED")
			failed++
		} else {
			match := 0
			for _, r := range *incidentReturn10b {

				if r.State == incident10.State &&
					r.Classification == incident10.Classification &&
					r.Severity == incident10.Severity &&
					len(r.CRNFull) == len(incident10.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(incident10.CRNFull[0]) &&
					r.SourceID == incident10.SourceID &&
					r.Source == incident10.Source &&
					r.PnPRemoved == incident10.PnPRemoved {

					match++
				}
			}

			if match <= 0 {
				log.Println("There are rows unmatched, failed")
				log.Println("Test 10b: Get Incident FAILED")
				failed++
			} else {
				log.Println("Test 10b: Get Incident PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count10b)
		log.Println("Test 10b: Get Incident FAILED")
		failed++
	}

	queryStr10c := db.RESOURCE_QUERY_CRN + "=crn:v1:watson:aa:" + gaasServices[1] + ":::::" + "&" + db.INCIDENT_QUERY_CREATION_TIME_START + "=" + creationTimeStart10 + "&" + db.INCIDENT_QUERY_CREATION_TIME_END + "=" + creationTimeEnd10
	log.Println("query: ", queryStr10c)
	incidentReturn10c, total_count10c, err10c, _ := db.GetIncidentByQuery(database, queryStr10c, 0, 0)
	log.Println("total_count10c: ", total_count10c)
	log.Println("incidentReturn10c: ", incidentReturn10c)

	if err10c != nil {
		log.Println("Get Incident failed: ", err10c.Error())
		log.Println("Test 10c: Get Incident FAILED")
		failed++
	} else if total_count10c != 0 {

		log.Printf("total_count is %d, expecting 0, failed", total_count10c)
		log.Println("Test 10c: Get Incident FAILED")
		failed++
	} else {
		log.Println("Test 10c: Get Incident PASSED")
	}

	// GetIncidentByQuery pnp_removed=false
	queryStr10d := db.RESOURCE_QUERY_CRN + "=crn:v1:watson:aa:" + gaasServices[1] + ":::::" + "&" + db.INCIDENT_COLUMN_PNP_REMOVED + "=false"
	log.Println("query: ", queryStr10d)
	incidentReturn10d, total_count10d, err10d, _ := db.GetIncidentByQuery(database, queryStr10d, 0, 0)
	log.Println("total_count10d: ", total_count10d)
	log.Println("incidentReturn10d: ", incidentReturn10d)

	if err10d != nil {
		log.Println("Get Incident failed: ", err10d.Error())
		log.Println("Test 10d: Get Incident FAILED")
		failed++
	}

	if len(*incidentReturn10d) != 1 {
		log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*incidentReturn10d))
		log.Println("Test 10d: Get Incident FAILED")
		failed++
	} else {
		log.Println("Test 10d: Get Incident PASSED")
	}

	// GetIncidentByQuery pnp_removed=true,false
	queryStr10e := db.RESOURCE_QUERY_CRN + "=crn:v1:watson:aa:" + gaasServices[1] + ":::::" + "&" + db.INCIDENT_COLUMN_PNP_REMOVED + "=true,false"
	log.Println("query: ", queryStr10e)
	incidentReturn10e, total_count10e, err10e, _ := db.GetIncidentByQuery(database, queryStr10e, 0, 0)
	log.Println("total_count10e: ", total_count10e)
	log.Println("incidentReturn10e: ", incidentReturn10e)

	if err10e != nil {
		log.Println("Get Incident failed: ", err10e.Error())
		log.Println("Test 10e: Get Incident FAILED")
		failed++
	}

	if len(*incidentReturn10e) != 2 {
		log.Printf("Error: Number of returned: %d, expecting: 2\n", len(*incidentReturn10e))
		log.Println("Test 10e: Get Incident FAILED")
		failed++
	} else {
		log.Println("Test 10e: Get Incident PASSED")
	}

	log.Println("**** SUMMARY ****")
	if failed > 0 {
		log.Println("**** " + FCT + "FAILED ****")
	} else {
		log.Println("**** " + FCT + "PASSED ****")
	}
	return failed
}
