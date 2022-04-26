package resourceAdapterTest

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-rest-test/config"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
	"github.ibm.com/cloud-sre/pnp-rest-test/testDefs"
)

var running = false

func ResourceAdapterIntegrationTest() string {
	const fct = "[ResourceAdapterIntegrationTest]"
	if running == true {
		log.Println(fct, "is already running")
		return "508 - Is already running"
	}
	running = true

	preReturns, numPreResources, err, retCode := db.GetResourceByQuery(config.Pdb, "crn=crn:v1::::::::", 0, 0)
	if err != nil {
		log.Println(fct, "Returned code:", retCode, "error:", err.Error())
	}
	log.Println(fct, "Resources:", len(*preReturns))
	log.Println(fct, "Number of records prior to deleteResources(numPreResources):", numPreResources)

	// Get and delete random resources
	recordID1, sourceID1, err, retCode := getRandomResource(config.Pdb)
	if err != nil || recordID1 == "" {
		log.Println(fct, "Did not get random resource 1")
		log.Println(fct, "Returned code:", retCode, "error:", err.Error())
		running = false
		return "500 - Did not get random resource 1"
	}
	log.Println(fct, "Got random resource 1:", recordID1, "-", sourceID1)

	recordID2, sourceID2, err, retCode := getRandomResource(config.Pdb)
	if err != nil || recordID1 == "" {
		log.Println(fct, "Did not get random resource 2")
		log.Println(fct, "Returned code:", retCode, "error:", err.Error())
		running = false
		return "500 - Did not get random resource 2"
	}
	log.Println(fct, "Got random resource 2:", recordID2, "-", sourceID2)

	err, retCode = deleteResource(config.Pdb, recordID1)
	if err != nil {
		log.Println(fct, "Did not delete resource 1")
		log.Println(fct, "Returned code:", retCode, "error:", err.Error())
		running = false
		return "500 - Did not delete resource 1"
	}
	log.Println(fct, "Deleted random resource 1:", recordID1)

	err, retCode = deleteResource(config.Pdb, recordID2)
	if err != nil {
		log.Println(fct, "Did not delete resource 2")
		log.Println(fct, "Returned code:", retCode, "error:", err.Error())
		running = false
		return "500 - Did not delete resource 2"
	}
	log.Println(fct, "Deleted random resource 2:", recordID2)

	checkReturns, _, err, retCode := db.GetResourceByQuery(config.Pdb, "crn=crn:v1::::::::", 0, 0)
	if err != nil {
		log.Println(fct, "Returned code:", retCode, "error:", err.Error())
	}
	log.Println(fct, "Resources:", len(*checkReturns))
	log.Println(fct, "numPreResources:", numPreResources)

	err = runImportResources()
	if err != nil {
		log.Println(fct, "->runImportResources:", err.Error())
	}

	// Check that resource adapter updated correctly
	found1 := false
	found2 := false
	maxRetry := 10
	for numRetry := 0; numRetry < maxRetry && (!found1 || !found2); numRetry++ {
		// Run import again as hashes might had been cached at nq2ds
		if numRetry == 8 {
			runImportResources()
		}

		if !found1 {
			return1, err, _ := db.GetResourceByRecordID(config.Pdb, recordID1)
			if err != nil {
				log.Println(fct, "Error getting recordID1:", err.Error())
			}
			if return1 != nil {
				found1 = true
			}
		}
		if !found2 {
			return2, err, _ := db.GetResourceByRecordID(config.Pdb, recordID2)
			if err != nil {
				log.Println(fct, "Error getting recordID2:", err.Error())
			}
			if return2 != nil {
				found2 = true
			}
		}

		if numRetry < maxRetry-1 && (!found1 || !found2) {
			time.Sleep(30 * time.Second)
			log.Println(fct, "sleeping for 30 seconds")
		}

		log.Println(fct, "found record 1:", found1, "\tfound record 2:", found2, "\tretry number:", numRetry, "/", maxRetry)
	}

	if found1 && found2 {
		log.Println(fct, testDefs.SUCCESS_MSG)
		running = false
		return testDefs.SUCCESS_MSG
	}

	log.Println(fct, "One or more of the deleted resources not found:", sourceID1, "->", found1, "\t", sourceID2, "->", found2)
	running = false
	return "500 - One or more of the deleted resources not found"
}

func runImportResources() error {
	const fct = "[runImportResources]"

	runCount := 0
	maxRun := 3
	resourceServer := rest.Server{}
	for runCount < maxRun {
		importResponse, err := resourceServer.Get(fct, "http://api-pnp-resource-adapter/importResources")
		if err != nil {
			log.Println(fct, "Error importResources:", err.Error())
			return err
		}

		log.Println(fct, "Returned code:", importResponse.StatusCode)
		if importResponse.StatusCode < 299 {
			return nil
		}

		log.Println(fct, "Sleeping for 60 seconds")
		time.Sleep(60 * time.Second)
		runCount++
		log.Println(fct, "Rerunning:", runCount, "/", maxRun)
	}

	return nil
}

func deleteResources(dbConnection *sql.DB) (error, int) {
	DefaultMaxRetries := 3
	DefaultRetryDelay := 2
	FCT := "DeleteResources: "

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		deleteQuery := "DELETE FROM resource_table"
		// Delete from Resource Table
		result, err := dbConnection.Exec(deleteQuery)
		log.Printf(FCT+"Result: %+v\n", result)
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(FCT+" Error Deleting Resource: ", err)
			if nretry < DefaultMaxRetries-1 {
				time.Sleep(time.Second * time.Duration(DefaultRetryDelay))
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

func deleteResource(dbConnection *sql.DB, recordId string) (error, int) {
	DefaultMaxRetries := 3
	DefaultRetryDelay := 2
	FCT := "deleteResource: "

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		deleteQuery := "DELETE FROM resource_table WHERE record_id = $1"
		// Delete from Resource Table
		result, err := dbConnection.Exec(deleteQuery, recordId)
		log.Printf(FCT+"Result: %+v\n", result)
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(FCT+" Error Deleting Resource: ", err)
			if nretry < DefaultMaxRetries-1 {
				time.Sleep(time.Second * time.Duration(DefaultRetryDelay))
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

func GetRandomResource(dbConnection *sql.DB) (string, string, error, int) {
	return getRandomResource(dbConnection)
}

func getRandomResource(dbConnection *sql.DB) (string, string, error, int) {
	FCT := "getRandomResource: "

	selectStmt := "SELECT record_id,source_id FROM resource_table where state IS NULL OR state != 'archived' order  by random() limit 1"
	rows, err := dbConnection.Query(selectStmt)
	rowsReturned := 0
	var recordid, source_id string

	switch {
	case err == sql.ErrNoRows:
		log.Println(FCT + "No Resource found.")
		return "", "", err, http.StatusOK
	case err != nil:
		log.Println(FCT+"Error : ", err)
		return "", "", err, http.StatusInternalServerError
	default:
		defer rows.Close()
		for rows.Next() {
			rowsReturned++
			err = rows.Scan(&recordid, &source_id)

			if err != nil {
				log.Printf(FCT+"Row Scan Error: %v", err)
				return "", "", err, http.StatusInternalServerError
			}
		}

		if rowsReturned == 0 {
			return "", "", sql.ErrNoRows, http.StatusOK
		}
	}

	//log.Println(FCT + "resourceToReturn: ", resourceToReturn)
	return recordid, source_id, nil, http.StatusOK
}
