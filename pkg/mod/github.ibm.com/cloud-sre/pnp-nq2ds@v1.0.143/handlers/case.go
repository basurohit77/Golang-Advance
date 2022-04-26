package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	newrelic "github.com/newrelic/go-agent"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
)

// ProcessCase - processes new or updated servicenow Cases
// Returns the inserted or updated case, and whether or not
// the message is a valid message
func ProcessCase(database *sql.DB, message []byte, nrApp newrelic.Application) (caseReturn *datastore.CaseReturn, isBadMessage bool) {

	isRetryNeeded := true
	for isRetryNeeded {
		caseReturn, isBadMessage, isRetryNeeded = internalProcessCase(database, message, nrApp)
		if isRetryNeeded {
			log.Print(tlog.Log() + "sleep and retry")
			time.Sleep(time.Second * 5)
		}
	}
	return caseReturn, isBadMessage
}

func internalProcessCase(database *sql.DB, message []byte, nrApp newrelic.Application) (caseReturn *datastore.CaseReturn, isBadMessage bool, isRetryNeeded bool) {

	//	log.Print("- Starting NewRelic transaction")
	txn := nrApp.StartTransaction("pnp-nq2ds-process-case", nil, nil)
	defer txn.End()

	// Record start time:
	//startTime := time.Now()
	//	log.Print("- Start time: " + strconv.FormatInt(int64(startTime.UnixNano()), 10))
	//log.Print("- nq2ds-process-case-StartTime attribute [", startTime.Unix(), "]")
	//monitor.AddInt64CustomAttribute(txn, "pnp-StartTime", startTime.Unix())
	//monitor.AddCustomAttribute(txn, "apiKubeClusterRegion", monitor.REGION)
	//monitor.AddCustomAttribute(txn, "apiKubeAppDeployedEnv", monitor.ENVIRONMENT)

	decryptedMsg, err := encryption.Decrypt(message)
	if err != nil {
		log.Println(tlog.Log()+"Message could not be decrypted, err = ", err)
		isBadMessage = true
		isRetryNeeded = false
		//monitor.AddCustomAttribute(txn, "pnp-nq2ds-decryption-err", err.Error())
		return caseReturn, isBadMessage, isRetryNeeded
	}

	log.Print(tlog.Log()+"INFO: Payload from queue: ", string(api.RedactAttributes(decryptedMsg)))

	// Unmarshal byte array:
	messageMap := make(map[string]interface{})
	err = json.Unmarshal(decryptedMsg, &messageMap)
	if err != nil {
		log.Print(tlog.Log(), "Bad message")
		isBadMessage = true
		isRetryNeeded = false
		return caseReturn, isBadMessage, isRetryNeeded
	}

	source := "servicenow"
	sourceID := interfaceToString(messageMap["number"])
	sourceSysID := interfaceToString(messageMap["sys_id"])
	operation := interfaceToString(messageMap["operation"])
	isBulkLoad := interfaceToString(messageMap["Process"]) == "BULK"

	// add attributes to New Relic transaction
	//monitor.AddCustomAttribute(txn, "pnp-Source", source)
	//monitor.AddCustomAttribute(txn, "pnp-SourceID", sourceID)
	//monitor.AddCustomAttribute(txn, "pnp-Kind", "case")

	// Verify the sourceID before continuing:
	if sourceID == "" {
		log.Print(tlog.Log() + "Source id (number of Case) is empty. Invalid message.")
		isBadMessage = true
		isRetryNeeded = false
		return caseReturn, isBadMessage, isRetryNeeded
	}

	log.Print(tlog.Log() + "Source: " + source + "Source ID: " + sourceID + " | operation: " + operation)

	// Check if Case already exists in the PnP data store:
	existingCase, doesCaseAlreadyExist, err := getExistingCase(database, source, sourceID)
	if err != nil {
		log.Print(tlog.Log(), err)
		isBadMessage = false
		isRetryNeeded = true
		return caseReturn, isBadMessage, isRetryNeeded
	}

	// Does the Case already exist in the PnP data store?
	if doesCaseAlreadyExist {
		log.Printf(tlog.Log()+"No update is needed. Existing case: %#v", existingCase)
		caseReturn = existingCase
		//monitor.AddCustomAttribute(txn, "pnp-Operation", "N/A")
	} else if !doesCaseAlreadyExist {
		log.Print(tlog.Log() + "Case does not exist. Need to insert it.")

		// Build object to insert into PnP datastore:
		caseInsert := &datastore.CaseInsert{
			Source:      source,
			SourceID:    sourceID,
			SourceSysID: sourceSysID,
		}
		//monitor.AddCustomAttribute(txn, "pnp-Operation", "insert")

		// Insert Case in the PnP datastore:
		_, err, _ := db.InsertCase(database, caseInsert)
		if err != nil {
			log.Print(tlog.Log(), err)
			isBadMessage = false
			isRetryNeeded = true
			//monitor.AddBoolCustomAttribute(txn, "pnp-db-failed", true)
			return caseReturn, isBadMessage, isRetryNeeded
		}
		//monitor.AddBoolCustomAttribute(txn, "pnp-db-failed", false)

		// Get the inserted Case from the PnP datastore:
		caseReturn, _, err = getExistingCase(database, source, sourceID)
		if err != nil {
			log.Print(tlog.Log(), err)
			isBadMessage = false
			isRetryNeeded = true
			return caseReturn, isBadMessage, isRetryNeeded
		}

		log.Printf(tlog.Log()+"DEBUG: Case inserted: %#v", api.RedactAttributes(caseReturn))
	}

	isBadMessage = false
	isRetryNeeded = false
	if isBulkLoad {
		// If bulk loading, do not want to publish case:
		return nil, isBadMessage, isRetryNeeded
	}

	// Capture end time:
	//endTime := time.Now()

	// Determine duration to report to NewRelic:
	//duration := endTime.Unix() - startTime.Unix()
	//	log.Print("- pnp-Duration attribute [", duration, "]")
	//monitor.AddInt64CustomAttribute(txn, "pnp-Duration", duration)

	return caseReturn, isBadMessage, isRetryNeeded
}

func getExistingCase(database *sql.DB, source string, sourceID string) (*datastore.CaseReturn, bool, error) {

	doesCaseAlreadyExist := true
	existingCase, err, httpStatusCode := db.GetCaseBySourceID(database, source, sourceID)
	if existingCase == nil && err != nil && httpStatusCode == 200 {
		// No row found, but no real error processing request
		doesCaseAlreadyExist = false
		err = nil
	} else if err != nil {
		// Error processing request
		doesCaseAlreadyExist = false
		log.Printf(tlog.Log()+"Error getting case from database, err = %s , status code: %d", err, httpStatusCode)
		//log.Print(tlog.Log()+"Http status code returned while getting incident from database = ", httpStatusCode)
	}
	return existingCase, doesCaseAlreadyExist, err
}
