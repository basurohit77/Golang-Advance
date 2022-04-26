package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/pnp-nq2ds/helper"
	"github.ibm.com/cloud-sre/pnp-nq2ds/shared"

	"github.com/araddon/dateparse"
	newrelic "github.com/newrelic/go-agent"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/ossmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-abstraction/exmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/initadapter"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
)

const (
	longDescriptionPrefix          = "Description:\n"
	longDescriptionSeperator       = "\n\nCurrent Status and Next Steps:\n"
	longDescriptionImpactSeperator = "\n\nDescription of Customer Facing Impact:\n"
)

type incidentFromQueue struct {
	SourceCreationTime        *string
	SourceUpdateTime          *string
	OutageStartTime           *string
	OutageEndTime             *string
	ShortDescription          *string
	LongDescription           *string
	State                     *string
	Classification            *string
	Severity                  *string
	CRNFull                   *[]string
	ServiceName               *string
	SourceID                  *string
	Source                    *string
	RegulatoryDomain          *string
	AffectedActivity          *string
	CustomerImpactDescription *string
	TargetedURL               *string
	Audience                  *string
}

// NOTE: Once NewRelic get completed decommissioned we need to remove all instances
// ossmon.SetTag(ctx, DBFailedErr, bool), ossmon.SetTag(ctx, DecryptionErr, bool), ossmon.SetTag(ctx, EncryptionErr, bool)
// SetError calls replaces the NQRL monitors

// ProcessIncident - processes new or updated incidents
// https://github.ibm.com/cloud-sre/pnp-abstraction/blob/master/datastore/Incident.go
// Replaced NewRelic monitoring by the monitor at the ossmon library for Instana migration
func ProcessIncident(database *sql.DB, message []byte, mon *ossmon.OSSMon) (incident *datastore.IncidentReturn, isBulkLoad bool, isBadMessage bool) {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	defer span.Finish()
	isRetryNeeded := true
	for isRetryNeeded {
		incident, isBulkLoad, isBadMessage, isRetryNeeded = internalProcessIncident(ctx, mon, database, message)
		if isRetryNeeded {
			log.Print(tlog.Log() + "sleep and retry")
			time.Sleep(time.Second * 5)
		}
	}

	// Creates a new transaction in New Relic and Instana and add attributes
	if incident != nil && !isBulkLoad && !isBadMessage {
		txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRprocIncPublic, nil, nil)
		ctx = newrelic.NewContext(ctx, txn)
		defer func() {
			err := txn.End()
			if err != nil {
				log.Println(tlog.Log(), err)
			}
		}()
		source := "servicenow"
		sourceID := incident.SourceID
		sourceCreationTime, _ := parseStringToTime(incident.SourceCreationTime)
		sourceUpdateTime, _ := parseStringToTime(incident.SourceUpdateTime)
		pnpCreationTime, _ := parseStringToTime(incident.PnpCreationTime)
		pnpUpdateTime, _ := parseStringToTime(incident.PnpUpdateTime)

		wait := pnpUpdateTime - sourceUpdateTime

		ossmon.SetTagsKV(ctx,
			ossmon.TagRegion, monitor.REGION,
			ossmon.TagEnv, monitor.ENVIRONMENT,
			ossmon.TagZone, monitor.ZONE,
			"pnp-Source", source,
			"pnp-SourceID", sourceID,
			"pnp-Kind", "incident",
			"sourceCreationTimeString", incident.SourceCreationTime,
			"sourceUpdateTimeString", incident.SourceUpdateTime,
			"pnpCreationTimeString", incident.PnpCreationTime,
			"pnpUpdateTimeString", incident.PnpUpdateTime,
			"sourceCreationTime", sourceCreationTime,
			"sourceUpdateTime", sourceUpdateTime,
			"pnpCreationTime", pnpCreationTime,
			"pnpUpdateTime", pnpUpdateTime,
			"pnp-WaitTime", wait,
			"apiKubeClusterRegion", monitor.NRregion,
			"apiKubeAppDeployedEnv", monitor.NRenvironment,
		)
	}
	return incident, isBulkLoad, isBadMessage
}

func internalProcessIncident(ctxParent context.Context, mon *ossmon.OSSMon, database *sql.DB, message []byte) (incident *datastore.IncidentReturn, isBulkLoad bool, isBadMessage bool, isRetryNeeded bool) {
	txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRprocIncident, nil, nil)
	ctxParent = newrelic.NewContext(ctxParent, txn)
	defer func() {
		err := txn.End()
		if err != nil {
			log.Println(tlog.Log(), err)
		}
	}()
	// Record start time:
	startTime := time.Now()
	ossmon.SetTagsKV(ctxParent,
		"pnp-StartTime", startTime.Unix(),
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
	)
	decryptedMsg, err := encryption.Decrypt(message)
	if err != nil {
		log.Println(tlog.Log()+"Message could not be decrypted, err = ", err)
		isBadMessage = true
		isRetryNeeded = false
		ossmon.SetTag(ctxParent, DecryptionErr, err.Error())
		ossmon.SetError(ctxParent, DecryptionErr+"-"+err.Error()) //Replaces api-pnp-nq2ds_ProcIncDecryptionErr NQRL monitor
		return incident, isBulkLoad, isBadMessage, isRetryNeeded
	}
	log.Print(tlog.Log()+"INFO: Payload from queue: ", string(api.RedactAttributes(decryptedMsg)))
	// Unmarshal byte array:
	messageMap := make(map[string]interface{})
	err = json.Unmarshal(decryptedMsg, &messageMap)
	if err != nil {
		log.Print(err)
		isBadMessage = true
		isRetryNeeded = false
		return incident, isBulkLoad, isBadMessage, isRetryNeeded
	}
	source := "servicenow"
	sourceID := interfaceToString(messageMap["number"])
	isBulkLoad = interfaceToString(messageMap["Process"]) == "BULK"
	log.Print("isBulkLoad: ", isBulkLoad)
	// add attributes to New Relic and Instana
	ossmon.SetTagsKV(ctxParent, "pnp-Source", source, "pnp-SourceID", sourceID, "pnp-Kind", "incident")
	// Verify the sourceID before continuing:
	if sourceID == "" {
		log.Print(tlog.Log() + "Source id (number of Incident) is empty. Invalid message.")
		isBadMessage = true
		isRetryNeeded = false
		return incident, isBulkLoad, isBadMessage, isRetryNeeded
	}
	// Get record ID based on payload from queue:
	recordID := db.CreateRecordIDFromSourceSourceID(source, sourceID)
	log.Print(tlog.Log()+"RecordID = ", recordID)
	// Check if the incident already exists:
	// Read incident from queue:
	incidentFromQueue, isBadMessage, isRetryNeeded := messageMapToIncident(messageMap)
	if isBadMessage || isRetryNeeded {
		return incident, isBulkLoad, isBadMessage, isRetryNeeded
	}
	api.PrettyPrintJSON(tlog.Log()+"Incident from queue", incidentFromQueue)
	incidentFromQueue = removeInvalidCRNs(incidentFromQueue)
	// notification needs the environment list
	var incEnvList string
	envList, ok := messageMap["u_environment"]
	if ok {
		incEnvList = interfaceToString(envList)
	}
	// detach the notification check to not delay the incident processing
	go checkNotificationForIncident(database, incidentFromQueue, incEnvList)

	// Bypass local storage: https://github.ibm.com/cloud-sre/toolsplatform/issues/9491
	// Se to bypass in charts (BYPASS_LOCAL_STORAGE = "true")
	if !shared.BypassLocalStorage {
		existingIncident, doesIncidentAlreadyExist := getExistingIncident(ctxParent, database, recordID)
		if doesIncidentAlreadyExist {
			log.Printf(tlog.Log() + "incident already exists")
			log.Printf(tlog.Log()+"Existing incident: %#v", existingIncident)
		}

		// Does the incident already exist in the PnP data store?
		if doesIncidentAlreadyExist &&
			incidentFromQueue.SourceUpdateTime != nil &&
			!helper.IsNewTimeAfterExistingTime(*incidentFromQueue.SourceUpdateTime, "2006-01-02 15:04:05Z", existingIncident.SourceUpdateTime, time.RFC3339) {
			ossmon.SetTag(ctxParent, "pnp-Operation", "N/A")
			log.Print(tlog.Log() + "Not updating incident because source update time is before source update time already in PG")
			isBadMessage = false
			isRetryNeeded = false
			return nil, isBulkLoad, isBadMessage, isRetryNeeded
		} else if doesIncidentAlreadyExist {
			// just need to update it in the database:
			updateIncident := buildIncidentToUpdate(existingIncident, incidentFromQueue)
			api.PrettyPrintJSON(tlog.Log()+"Incident to update", updateIncident)
			// tombstone marker
			pnpRemoved := isPnPRemoved(incidentFromQueue)
			// Update the incident into the database:
			ossmon.SetTag(ctxParent, "pnp-Operation", "update")
			err, httpStatusCode := db.UpdateIncident(database, toIncidentInsertWithTombstone(updateIncident, pnpRemoved))
			if err != nil {
				log.Print(tlog.Log()+"Error updating incident, err = ", err, ", http status code = ", httpStatusCode)
				// Don't want to keep trying to update with invalid data - tossing out the update.
				if hasDBValidationError(err.Error()) {
					log.Println(tlog.Log()+"Validation error on message, treating as bad message. Not going to retry. err =", err.Error())
					isBadMessage = true
					isRetryNeeded = false
				} else {
					isBadMessage = false
					isRetryNeeded = true
				}
				ossmon.SetTag(ctxParent, DBFailedErr, true)
				ossmon.SetError(ctxParent, DBFailedErr+"-"+err.Error()) // Replaces api-pnp-nq2ds_ProcIncDecryptionErr NQRL monitor
				return incident, isBulkLoad, isBadMessage, isRetryNeeded
			}
			ossmon.SetTag(ctxParent, DBFailedErr, false)
			log.Print(tlog.Log()+"Incident is updated, record id=", recordID)
			// Notify when pnpRemoved changes
			if existingIncident.PnPRemoved != pnpRemoved && !pnpRemoved {
				isRetryNeeded = notificationForIncidentUpdate(database, incidentFromQueue, pnpRemoved, incEnvList)
				isBadMessage = false
			}
			existingIncident, doesIncidentAlreadyExist = getExistingIncident(ctxParent, database, recordID)
			log.Printf(tlog.Log()+"Updated incident %#v, exist: %v", existingIncident, doesIncidentAlreadyExist)
		} else if !doesIncidentAlreadyExist && !isHighSevCIE(incidentFromQueue) {
			ossmon.SetTag(ctxParent, "pnp-Operation", "N/A")
			log.Print(tlog.Log() + "Not inserting incident because severity is not 1 or status is not confirmed-cie")
		} else if !doesIncidentAlreadyExist && !hasAtleastOneCRN(incidentFromQueue) {
			ossmon.SetTag(ctxParent, "pnp-Operation", "N/A")
			log.Print(tlog.Log() + "Not inserting incident because does not contain at least one public CRN")
		} else if !doesIncidentAlreadyExist {
			// No, so need to insert it into the database:
			if incidentFromQueue.TargetedURL != nil && incidentFromQueue.SourceID != nil {
				var re = regexp.MustCompile(`\$SN_RECORD_ID`)
				tempString := re.ReplaceAllString(*incidentFromQueue.TargetedURL, *incidentFromQueue.SourceID)
				incidentFromQueue.TargetedURL = &tempString
			}
			ossmon.SetTag(ctxParent, "pnp-Operation", "insert")
			recordID, err, httpStatusCode := db.InsertIncident(database, toIncidentInsert(incidentFromQueue))
			if err != nil && httpStatusCode != 400 {
				log.Print(tlog.Log()+"Error creating incident, err = ", err, ", http status code = ", httpStatusCode)
				isBadMessage = false
				isRetryNeeded = true
				ossmon.SetTag(ctxParent, DBFailedErr, true)
				ossmon.SetError(ctxParent, DBFailedErr+"-"+err.Error()) // Replaces api-pnp-nq2ds_ProcIncDecryptionErr NQRL monitor
				return incident, isBulkLoad, isBadMessage, isRetryNeeded
			} else if err != nil && httpStatusCode == 400 {
				log.Print(tlog.Log()+"Error creating incident, err = ", err, ", http status code = ", httpStatusCode)
				isBadMessage = true
				isRetryNeeded = false
				ossmon.SetTag(ctxParent, DBFailedErr, false)
				return incident, isBulkLoad, isBadMessage, isRetryNeeded
			}
			log.Print(tlog.Log()+"Incident is inserted, record id=", recordID)
			existingIncident, doesIncidentAlreadyExist = getExistingIncident(ctxParent, database, recordID)
			log.Printf(tlog.Log()+"DEBUG: Inserted incident %#v, exist: %v", api.RedactAttributes(existingIncident), doesIncidentAlreadyExist)
		}
		incident = existingIncident
		isBadMessage = false
		isRetryNeeded = false
	}
	// Capture end time:
	ossmon.SetTag(ctxParent, "pnp-Duration", time.Since(startTime))
	return incident, isBulkLoad, isBadMessage, isRetryNeeded
}

// parseStringToTime parses a returned time string back to time
func parseStringToTime(stringToParse string) (int64, error) {
	loc, err := time.LoadLocation("")
	if err != nil {
		log.Println(tlog.Log()+"Location cannot be recognized, err = ", err)
	}
	time.Local = loc
	newParse, newParseErr := dateparse.ParseLocal(stringToParse)
	if newParseErr != nil {
		log.Println(tlog.Log()+"Unable to parse returned incident string back to time, err = ", newParseErr)
		return 0, newParseErr
	}
	return newParse.Unix(), nil
}

// hasDBValidationError returns true when an error matches a specific pre-defined DB error, and false otherwise
func hasDBValidationError(err string) bool {
	switch err {
	case db.ERR_NO_SERVICE,
		db.ERR_NO_CNAME,
		db.ERR_BAD_CLASSIFICATION,
		db.ERR_BAD_CRN_FORMAT,
		db.ERR_BAD_STATE,
		db.ERR_NO_CRN,
		db.ERR_NO_CRN_VERSION,
		db.ERR_NO_CTYPE,
		db.ERR_NO_LOCATION,
		db.ERR_NO_SOURCE,
		db.ERR_NO_SOURCEID:
		return true
	}
	return false
}

func getExistingIncident(ctxParent context.Context, database *sql.DB, recordID string) (*datastore.IncidentReturn, bool) {
	doesIncidentAlreadyExist := true
	existingIncident, err, httpStatusCode := db.GetIncidentByRecordIDStatement(database, recordID)
	if err != nil && httpStatusCode == 200 {
		doesIncidentAlreadyExist = false
	} else if err != nil {
		doesIncidentAlreadyExist = false
		log.Printf(tlog.Log()+"Error getting incident from db, err = %s , status: %d", err, httpStatusCode)
		ossmon.SetTag(ctxParent, DBFailedErr, true)
		ossmon.SetError(ctxParent, DBFailedErr+"-"+err.Error()) // Replaces api-pnp-nq2ds NQRL monitor
	}
	return existingIncident, doesIncidentAlreadyExist
}

func processTimestamp(value interface{}) (string, error) {
	valueAsString := interfaceToString(value)
	if valueAsString != "" {
		// Time from servicenow is in UTC but does not have the separator (T) and the timezone indicator (Z)
		_, err := time.Parse("2006-01-02 15:04:05", valueAsString)
		if err != nil {
			log.Print(tlog.Log() + fmt.Sprintf("Timestamp format %s is not valid. Invalid message.", valueAsString))
			return "", err
		}
		// Add UTC timezone as expected by the downstream code
		valueAsString += "Z"
	}
	return valueAsString, nil
}

func messageMapToIncident(messageMap map[string]interface{}) (incident *incidentFromQueue, isBadMessage, isRetryNeeded bool) {
	incidentToInsert := new(incidentFromQueue)
	value, ok := messageMap["sys_created_on"]
	if ok {
		valueAsString, err := processTimestamp(value)
		if err != nil {
			return nil, true, false
		}
		incidentToInsert.SourceCreationTime = &valueAsString
	}
	value, ok = messageMap["sys_updated_on"]
	if ok {
		valueAsString, err := processTimestamp(value)
		if err != nil {
			return nil, true, false
		}
		incidentToInsert.SourceUpdateTime = &valueAsString
	}
	value, ok = messageMap["u_disruption_began"]
	if ok {
		valueAsString, err := processTimestamp(value)
		if err != nil {
			return nil, true, false
		}
		incidentToInsert.OutageStartTime = &valueAsString
	}
	value, ok = messageMap["u_disruption_ended"]
	if ok {
		valueAsString, err := processTimestamp(value)
		if err != nil {
			return nil, true, false
		}
		incidentToInsert.OutageEndTime = &valueAsString
	}
	value, ok = messageMap["short_description"]
	if ok {
		valueAsString := interfaceToString(value)
		incidentToInsert.ShortDescription = &valueAsString
	}
	longDescription := buildLongDescription(interfaceToString(messageMap["u_current_status"]),
		interfaceToString(messageMap["u_description_customer_impact"]))
	incidentToInsert.LongDescription = &longDescription
	value, ok = messageMap["incident_state"]
	if ok {
		valueAsString := snStateToText(interfaceToString(value))
		incidentToInsert.State = &valueAsString
	}
	value, ok = messageMap["u_status"]
	if ok {
		valueAsString := snUStatusToText(interfaceToString(value))
		incidentToInsert.Classification = &valueAsString
	}
	value, ok = messageMap["priority"]
	if ok {
		valueAsString := snPriorityToSeverity(interfaceToString(value))
		incidentToInsert.Severity = &valueAsString
	}
	// Add CRNs:
	value, ok = messageMap["crn"]
	if ok {
		crns := interfaceToStringArray(value)
		incidentToInsert.CRNFull = &crns
	} else {
		incidentToInsert.CRNFull = nil
		log.Println(tlog.Log() + "crn is not provided. Rejecting message.")
		return nil, true, false
	}
	log.Print(tlog.Log()+"crns = ", incidentToInsert.CRNFull)
	updatedList, err := normalizeCRNServiceNames(*incidentToInsert.CRNFull)
	if err != nil {
		return nil, false, true
	}
	incidentToInsert.CRNFull = &updatedList
	// Set service name based on CRNs (note all CRNs have the same service name):
	if incidentToInsert.CRNFull != nil && len(*incidentToInsert.CRNFull) > 0 {
		crn := (*incidentToInsert.CRNFull)[0]
		serviceName, err := api.GetServiceFromCRN(crn)
		if err != nil {
			log.Print(tlog.Log()+"Error occurred getting service name from the following CRN: ", crn)
		} else {
			incidentToInsert.ServiceName = &serviceName
		}
	}
	value, ok = messageMap["number"]
	if ok {
		valueAsString := interfaceToString(value)
		incidentToInsert.SourceID = &valueAsString
	}
	source := "servicenow"
	incidentToInsert.Source = &source
	value, ok = messageMap["u_affected_activity"]
	if ok {
		valueAsString := interfaceToString(value)
		incidentToInsert.AffectedActivity = &valueAsString
	}
	value, ok = messageMap["u_description_customer_impact"]
	if ok {
		valueAsString := interfaceToString(value)
		incidentToInsert.CustomerImpactDescription = &valueAsString
	}
	value, ok = messageMap["u_targeted_notification_url"]
	if ok {
		valueAsString := interfaceToString(value)
		incidentToInsert.TargetedURL = &valueAsString
	}
	value, ok = messageMap["u_audience"]
	if ok {
		valueAsString := interfaceToString(value)
		log.Println(tlog.Log() + "u_audience:" + valueAsString)
		if valueAsString == "" || len(valueAsString) == 0 {
			valueAsString = db.SNnill2PnP
		}
		incidentToInsert.Audience = &valueAsString
	}
	// Can not be set regulatory domain - see https://github.ibm.com/cloud-sre/pnp-status/issues/74 for details:
	//incidentToInsert.RegulatoryDomain = ""
	return incidentToInsert, false, false
}

var gContext ctxt.Context

func getAdapterContext() (ctxt.Context, error) {
	if gContext.LogID == "" { // Determine if we have initialized
		_, err := initadapter.Initialize()
		if err != nil {
			log.Println(tlog.Log(), err)
			return gContext, err
		}
		monitorEx, err := exmon.CreateMonitor()
		if err != nil {
			log.Println(tlog.Log(), err)
			return gContext, err
		}
		gContext.NRMon = monitorEx
		gContext.LogID = "nq2ds.incident"

	}
	return gContext, nil
}

// normalizeServiceName will ensure that a service name reflects its categoryID parent if it exists
func normalizeServiceName(servicename string, context ctxt.Context) (normalizedServiceName string, err error) {
	normalizedServiceName = servicename
	record, err := osscatalog.ServiceNameToOSSRecord(context, servicename)
	if err != nil {
		log.Println(tlog.Log(), "Error getting category ID", err)
		return "", err
	}
	parent := string(record.StatusPage.CategoryParent)
	if parent != "" {
		normalizedServiceName = parent
	}
	return normalizedServiceName, nil
}

func normalizeCRNServiceNames(inputList []string) (outList []string, err error) {
	// CRN form = crn:v1:bluemix:public:service:location::::
	outList = make([]string, 0)
	ctxAdapter, err := getAdapterContext()
	if err != nil {
		return nil, err
	}
	for _, crn := range inputList {
		parts := strings.Split(crn, ":")
		if len(parts) < 10 {
			// ignoring crns that are malformed
			log.Print(tlog.Log()+" found invalid crn, ignoring", crn)
			continue
		}
		originalServiceName := parts[4]
		parts[4], err = normalizeServiceName(parts[4], ctxAdapter)
		if err != nil {
			log.Print(tlog.Log()+" Error occurred normalizing service name, so using original service name: ", originalServiceName)
			parts[4] = originalServiceName
			err = nil
		}
		//ignore duplicates after normalization - ToolsPlatform#8239
		newCrn := strings.Join(parts, ":")
		dup := false
		for _, c := range outList {
			if c == newCrn {
				dup = true
			}
		}
		if !dup {
			outList = append(outList, newCrn)
		}
	}
	return outList, err
}

func buildLongDescription(currentAndNextSteps string, customerImpact string) (longDescription string) {
	// 2020-08-13: Do not provide long description per Hamilton security restrictions. Just remove here for now so as not to break backward compatibility
	longDescription = longDescriptionPrefix + "" + longDescriptionSeperator + currentAndNextSteps + longDescriptionImpactSeperator + customerImpact
	return longDescription
}

func splitLongDescription(longDescription string) (description string, currentAndNextSteps string, customerImpact string) {
	if longDescription != "" {
		splitLongDescription := strings.Split(longDescription, longDescriptionSeperator)
		if len(splitLongDescription) == 2 {
			description = strings.TrimPrefix(splitLongDescription[0], longDescriptionPrefix)
			currentAndNextSteps = splitLongDescription[1]
		} else if len(splitLongDescription) == 1 {
			description = strings.TrimPrefix(splitLongDescription[0], longDescriptionPrefix)
		}
	}
	if currentAndNextSteps != "" {
		splitCurrentAndNextSteps := strings.Split(currentAndNextSteps, longDescriptionImpactSeperator)
		if len(splitCurrentAndNextSteps) == 2 {
			currentAndNextSteps = splitCurrentAndNextSteps[0]
			customerImpact = splitCurrentAndNextSteps[1]
		}
	}
	return description, currentAndNextSteps, customerImpact
}

func snStateToText(snState string) string {
	result := "in-progress"
	if snState == "1" || snState == "New" {
		result = "new"
	} else if snState == "6" || snState == "Resolved" {
		result = "resolved"
	} else if snState == "7" || snState == "Closed" {
		result = "resolved"
	} else if snState == "" {
		result = ""
	}
	return result
}

func snUStatusToText(snUStatus string) string {
	result := "normal"
	if snUStatus == "20" || snUStatus == "Potential CIE" {
		result = "potential-cie"
	} else if snUStatus == "21" || snUStatus == "Confirmed CIE" {
		result = "confirmed-cie"
	} else if snUStatus == "" {
		result = ""
	}
	return result
}

func snPriorityToSeverity(snPriority string) string {
	result := ""
	if snPriority == "1" || snPriority == "Sev - 1" {
		result = "1"
	} else if snPriority == "2" || snPriority == "Sev - 2" {
		result = "2"
	} else if snPriority == "3" || snPriority == "Sev - 3" {
		result = "3"
	} else if snPriority == "4" || snPriority == "Sev - 4" {
		result = "4"
	}
	return result
}

func interfaceToString(anInterface interface{}) string {
	asString, ok := anInterface.(string)
	if ok {
		return asString
	}
	return ""
}

func interfaceToStringArray(anInterface interface{}) (stringArray []string) {
	interfaceArray, ok := anInterface.([]interface{})
	if ok {
		for _, anInterface := range interfaceArray {
			asString := interfaceToString(anInterface)
			if asString != "" {
				stringArray = append(stringArray, asString)
			}
		}
	}
	return
}

func buildIncidentToUpdate(currentIncident *datastore.IncidentReturn, updateIncident *incidentFromQueue) *incidentFromQueue {
	if updateIncident.SourceCreationTime == nil {
		updateIncident.SourceCreationTime = &currentIncident.SourceCreationTime
	}
	if updateIncident.SourceUpdateTime == nil {
		updateIncident.SourceUpdateTime = &currentIncident.SourceUpdateTime
	}
	if updateIncident.OutageStartTime == nil {
		updateIncident.OutageStartTime = &currentIncident.OutageStartTime
	}
	if updateIncident.OutageEndTime == nil {
		updateIncident.OutageEndTime = &currentIncident.OutageEndTime
	}
	if updateIncident.ShortDescription == nil {
		updateIncident.ShortDescription = &currentIncident.ShortDescription
	}

	// Build the long description using the current incident and the update incident:
	currentIncidentDescription, currentIncidentCurrentAndNextSteps, currentIncidentCustomerImpact := splitLongDescription(currentIncident.LongDescription)

	updateIncidentDescription, updateIncidentCurrentAndNextSteps, updateIncidentCustomerImpact := "", "", ""
	if updateIncident.LongDescription != nil {
		updateIncidentDescription, updateIncidentCurrentAndNextSteps, updateIncidentCustomerImpact = splitLongDescription(*updateIncident.LongDescription)
	}
	description := updateIncidentDescription
	if description == "" {
		description = currentIncidentDescription
	}
	currentAndNextSteps := updateIncidentCurrentAndNextSteps
	if currentAndNextSteps == "" {
		currentAndNextSteps = currentIncidentCurrentAndNextSteps
	}
	customerImpact := updateIncidentCustomerImpact
	if customerImpact == "" {
		customerImpact = currentIncidentCustomerImpact
	}

	longDescription := buildLongDescription(currentAndNextSteps, customerImpact)

	if longDescription != "" {
		updateIncident.LongDescription = &longDescription
	}

	if updateIncident.State == nil {
		updateIncident.State = &currentIncident.State
	}
	if updateIncident.Classification == nil {
		updateIncident.Classification = &currentIncident.Classification
	}

	// Update CRNs:
	if updateIncident.CRNFull == nil {
		// The update incident has not CRNs, therefore use the current incident CRNs:
		updateIncident.CRNFull = &currentIncident.CRNFull
	} else if len(*updateIncident.CRNFull) == 0 {
		// All CRNs have been removed, so set to nil:
		updateIncident.CRNFull = nil
	}

	if updateIncident.Severity == nil {
		updateIncident.Severity = &currentIncident.Severity
	}
	if updateIncident.SourceID == nil {
		updateIncident.SourceID = &currentIncident.SourceID
	}
	if updateIncident.Source == nil {
		updateIncident.Source = &currentIncident.Source
	}
	if updateIncident.AffectedActivity == nil {
		updateIncident.AffectedActivity = &currentIncident.AffectedActivity
	}
	if updateIncident.CustomerImpactDescription == nil {
		updateIncident.CustomerImpactDescription = &currentIncident.CustomerImpactDescription
	}

	if updateIncident.TargetedURL != nil && updateIncident.SourceID != nil {
		var re = regexp.MustCompile(`\$SN_RECORD_ID`)
		tempString := re.ReplaceAllString(*updateIncident.TargetedURL, *updateIncident.SourceID)
		updateIncident.TargetedURL = &tempString
	}

	log.Println(tlog.Log() + "Current: " + currentIncident.Audience)
	if updateIncident.Audience == nil {
		// updateIncident is a complete new record
		log.Println(tlog.Log() + "updateIncident.Audience is null")
		updateIncident.Audience = &currentIncident.Audience
	} else {
		log.Println(tlog.Log() + "Current: " + currentIncident.Audience + " New: " + *updateIncident.Audience)
		if len(*updateIncident.Audience) == 0 {
			log.Println(tlog.Log() + "updateIncident.Audience len == 0")
			strNone := db.SNnill2PnP
			updateIncident.Audience = &strNone
		}

	}

	return updateIncident
}

func isHighSevCIE(incident *incidentFromQueue) bool {
	return incident != nil && incident.Severity != nil && (*incident.Severity == "1" || *incident.Severity == "2") && incident.Classification != nil && (*incident.Classification == "confirmed-cie" || *incident.Classification == "potential-cie")
}

func isSev1CIE(incident *incidentFromQueue) bool {
	return incident != nil && incident.Severity != nil && *incident.Severity == "1" && incident.Classification != nil && *incident.Classification == "confirmed-cie"
}

func hasAtleastOneCRN(incidentFromQueue *incidentFromQueue) (shouldProcess bool) {
	return incidentFromQueue != nil && incidentFromQueue.CRNFull != nil && len(*incidentFromQueue.CRNFull) > 0
}

func isPnPRemoved(incident *incidentFromQueue) bool {
	// tombstone marker: true when any of these is true: not a Sev=1, not a CIE, doesn't have CRN
	return !(isHighSevCIE(incident) && hasAtleastOneCRN(incident))
}

func removeInvalidCRNs(incidentFromQueue *incidentFromQueue) *incidentFromQueue {
	if incidentFromQueue != nil && incidentFromQueue.CRNFull != nil && len(*incidentFromQueue.CRNFull) > 0 {
		var validCRNs []string
		for _, crn := range *incidentFromQueue.CRNFull {
			crn = api.NormalizeCRN(crn)

			// special case for PnP Hearbeat, which uses service pnp-api-oss to test the PnP flow
			// see https://github.ibm.com/cloud-sre/toolsplatform/issues/8352
			svc, err := api.GetServiceFromCRN(crn)
			if err != nil {
				log.Println(err)
			}

			if api.IsCrnPnpValid(crn) || svc == "pnp-api-oss" {
				validCRNs = append(validCRNs, crn)
			}
		}
		incidentFromQueue.CRNFull = &validCRNs
	}
	return incidentFromQueue
}

func toIncidentInsert(incident *incidentFromQueue) (incidentInsert *datastore.IncidentInsert) {
	return toIncidentInsertWithTombstone(incident, false)
}

func toIncidentInsertWithTombstone(incident *incidentFromQueue, pnpRemoved bool) (incidentInsert *datastore.IncidentInsert) {
	incidentInsert = new(datastore.IncidentInsert)

	incidentInsert.PnPRemoved = pnpRemoved

	if incident.SourceCreationTime != nil {
		incidentInsert.SourceCreationTime = *incident.SourceCreationTime
	}
	if incident.SourceUpdateTime != nil {
		incidentInsert.SourceUpdateTime = *incident.SourceUpdateTime
	}
	if incident.OutageStartTime != nil {
		incidentInsert.OutageStartTime = *incident.OutageStartTime
	}
	if incident.OutageEndTime != nil {
		incidentInsert.OutageEndTime = *incident.OutageEndTime
	}
	if incident.ShortDescription != nil {
		incidentInsert.ShortDescription = *incident.ShortDescription
	}
	if incident.LongDescription != nil {
		incidentInsert.LongDescription = *incident.LongDescription
	}
	if incident.State != nil {
		incidentInsert.State = *incident.State
	}
	if incident.Classification != nil {
		incidentInsert.Classification = *incident.Classification
	}
	if incident.Severity != nil {
		incidentInsert.Severity = *incident.Severity
	}
	if incident.CRNFull != nil {
		incidentInsert.CRNFull = *incident.CRNFull
	}
	if incident.SourceID != nil {
		incidentInsert.SourceID = *incident.SourceID
	}
	if incident.Source != nil {
		incidentInsert.Source = *incident.Source
	}
	if incident.RegulatoryDomain != nil {
		incidentInsert.RegulatoryDomain = *incident.RegulatoryDomain
	}
	if incident.AffectedActivity != nil {
		incidentInsert.AffectedActivity = *incident.AffectedActivity
	}
	if incident.CustomerImpactDescription != nil {
		incidentInsert.CustomerImpactDescription = *incident.CustomerImpactDescription
	}
	if incident.TargetedURL != nil {
		incidentInsert.TargetedURL = *incident.TargetedURL
	}

	if incident.Audience != nil {
		incidentInsert.Audience = *incident.Audience
	}
	return incidentInsert
}
