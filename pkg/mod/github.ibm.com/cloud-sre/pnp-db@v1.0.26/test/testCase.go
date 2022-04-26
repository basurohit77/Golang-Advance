package test

import (
	"database/sql"
	"log"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)


func TestCase(database *sql.DB) int {
	FCT := "TestCase "
	failed := 0

	log.Println("======== "+FCT+" Test 1 ========")
	caseInsert := datastore.CaseInsert {
		Source:           "source1",
		SourceID:         "case_source_id1",
		SourceSysID:      "sys_id1",
	}

	record_id, err, _ := db.InsertCase(database, &caseInsert)
	if err != nil {
		log.Println("Insert Case failed: ", err.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id)
		log.Println("Insert Case passed")
	
		log.Println("Test get Case")
		caseRet, err1, _ := db.GetCaseBySourceID(database, caseInsert.Source, caseInsert.SourceID)
		if err1 != nil {
			log.Println("Get Case failed: ", err1.Error())
			failed++
		} else {
			log.Println("Get result: ", caseRet)
			if caseRet.RecordID == record_id &&
				caseRet.Source == caseInsert.Source &&
				caseRet.SourceID == caseInsert.SourceID &&
				caseRet.SourceSysID == caseInsert.SourceSysID {
		
				log.Println("Get Case passed")
			}
		}
	}

	log.Println("======== "+FCT+" Test 3 ========")
	log.Println("Test delete Case")
	err3,_ := db.DeleteCaseStatement(database, record_id)
	if err3 != nil {
		log.Println("Delete Case failed: ", err3.Error())
		log.Println("Test 3: Delete Case FAILED")
		failed++
	}
	caseReturn, err3a, _ := db.GetCaseByRecordID(database, record_id)
	if caseReturn != nil {
		log.Println("Delete Case failed, Case still exists: ", err3a.Error())
		log.Println("Test 3: Delete Case FAILED")
		failed++
	} else {
		log.Println("Test 3: Delete Case PASSED")
	}



	log.Println("**** SUMMARY ****")
	if failed > 0 {
		log.Println("**** "+FCT+"FAILED ****")
	} else {
		log.Println("**** "+FCT+"PASSED ****")
	}

	return failed
}

