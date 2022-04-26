package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-nq2ds/shared"

	newrelic "github.com/newrelic/go-agent"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/ossmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/exmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
	hooks "github.ibm.com/cloud-sre/pnp-hooks/handlers"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/initadapter"
	"github.ibm.com/cloud-sre/pnp-nq2ds/helper"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
	rabbitmq "github.ibm.com/cloud-sre/pnp-rabbitmq-connector"
)

// NOTE: Once NewRelic get completed decommissioned we need to remove all instances
// ossmon.SetTag(ctx, DBFailedErr, bool), ossmon.SetTag(ctx, DecryptionErr, bool), ossmon.SetTag(ctx, EncryptionErr, bool)
// SetError calls replaces the NQRL monitors

const (
	// MQEncryptionEnabled controls whether or not encryption will execute on this consumer
	MQEncryptionEnabled = true
	// Update is a constant for the message type representing an update
	Update = datastore.NotificationMsgType("update")
)

var (
	msgCounter  = make(map[string][]int)
	ctxtContext ctxt.Context
	// RabbitmqURLs holds the rabbitMQ URLs
	RabbitmqURLs []string
	// RabbitmqTLSCert holds the rabbitMQ cert needed for Messages for RabbitMQ
	RabbitmqTLSCert string
	// RabbitmqEnableMessages indicates whether Messages for RabbitMQ is being used
	RabbitmqEnableMessages bool
	// NotificationRoutingKey is the routing key used to produce a msg to rabbitMQ
	NotificationRoutingKey string
	// MQExchangeName is the exchange used to produce a msg to rabbitMQ
	MQExchangeName string
)

func init() {
	loc, err := time.LoadLocation("")
	if err != nil {
		panic(err.Error())
	}
	time.Local = loc
}

// ProcessNotification receives the message that came from the notification adapter
func ProcessNotification(database *sql.DB, message []byte, mon *ossmon.OSSMon) (*datastore.NotificationMsg, error) {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	defer span.Finish()
	var notification *datastore.NotificationMsg
	var err error
	isRetryNeeded := true
	for isRetryNeeded {
		notification, isRetryNeeded, err = internalProcessNotification(ctx, mon, database, message)
		log.Println(tlog.Log(), err)
		if isRetryNeeded {
			log.Print(tlog.Log() + "sleep and retry")
			time.Sleep(time.Second * 5)
		}
	}

	return notification, err
}

func internalProcessNotification(ctxParent context.Context, mon *ossmon.OSSMon, database *sql.DB, message []byte) (notification *datastore.NotificationMsg, isRetryNeeded bool, err error) {
	txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRprocNotification, nil, nil)
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
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
		"pnp-StartTime", startTime.Unix(),
	)
	notification = new(datastore.NotificationMsg)
	decryptedMsg := message
	//var err error
	isRetryNeeded = false
	if MQEncryptionEnabled {
		decryptedMsg, err = encryption.Decrypt(message)
		if err != nil {
			errMsg := ": failed to decrypt message err=" + err.Error()
			log.Println(tlog.Log(), errMsg)
			ossmon.SetTag(ctxParent, DecryptionErr, err.Error())
			ossmon.SetError(ctxParent, DecryptionErr+"-"+errMsg)
			addCount(false)
			return nil, false, err
		}
	}
	log.Printf(tlog.Log()+"msg coming in : \n %s", string(decryptedMsg))
	if err := json.NewDecoder(bytes.NewReader(decryptedMsg)).Decode(notification); err != nil {
		log.Printf(tlog.Log()+"failed to parse message (decrypt=%t) msg=%s", MQEncryptionEnabled, string(decryptedMsg))
		addCount(false)
		return nil, false, err
	}
	// Check if the notification has a parent record that is already in the database
	if !shared.BypassLocalStorage || notification.Type == "announcement" || notification.Type == "security" || notification.Type == "release_note" {
		err = hasValidIncidentID(database, notification)
		if err != nil {
			err = errors.New(fmt.Sprintf("Notification(%s) does not have a record(%s) in PnP: ",
				notification.SourceID, notification.IncidentID) + err.Error())
			return nil, isRetryNeeded, err
		}
	}

	notification.CRNFull = strings.ToLower(notification.CRNFull)
	notification.CRNFull = api.NormalizeCRN(notification.CRNFull)
	if !api.IsCrnPnpValid(notification.CRNFull) {
		log.Printf(tlog.Log()+"CRN is not a public CRN, stopping processing of notification. CRN=%s", notification.CRNFull)
		return nil, false, nil
	}
	//  Ignoring language types for now
	if notification.Source == "servicenow" && notification.Type == "maintenance" {
		for index := range notification.LongDescription {
			notification.LongDescription[index].Name = notification.LongDescription[index].Name +
				"<br/><b>Maintenance Duration</b>: " + strconv.Itoa(notification.MaintenanceDuration) + " minutes" +
				"<br/><br/><b>Disruption Duration</b>: " + strconv.Itoa(notification.DisruptionDuration) + " minutes"
		}
	}
	log.Printf(tlog.Log()+"notification: %+v", notification.LongDescription)
	exists, sourceUpdateTimeOfExistingResource := noteExistsInDB(database, &notification.NotificationInsert)
	// Determine duration to report to NewRelic and Instana
	ossmon.SetTag(ctxParent, "pnp-Duration", time.Since(startTime))

	// Bypass local storage: https://github.ibm.com/cloud-sre/toolsplatform/issues/9491
	if !shared.BypassLocalStorage || notification.Type == "announcement" || notification.Type == "security" || notification.Type == "release_note" {
		err = sendNoteToDB(ctxParent, mon, database, &notification.NotificationInsert, exists, sourceUpdateTimeOfExistingResource)
		if err != nil {
			log.Printf("error trying to send notification to database error=[%s]  msg=%s", err.Error(), string(message))
			notification = nil
			isRetryNeeded = true
		}
	} else {
		log.Println("notice: the record was not added to the database since it is not an annoucement, security bulletin, or release note", notification.SourceID)
	}

	addCount(err == nil)
	return notification, isRetryNeeded, err
}

func addCount(success bool) {
	if len(msgCounter) > 40 {
		msgCounter = make(map[string][]int)
	}
	t := time.Now().Format("02T15")
	m := msgCounter[t]
	if m == nil {
		m = make([]int, 2)
		msgCounter[t] = m
	}
	if success {
		m[0]++
	} else {
		m[1]++
	}
	msg := ""
	for k, v := range msgCounter {
		msg += fmt.Sprintf("[%s,%d,%d]", k, v[0], v[1])
	}
	log.Printf(tlog.Log()+"INFO: Notification totals: %s", msg)
}

// Bypass local storage: https://github.ibm.com/cloud-sre/toolsplatform/issues/9491
func sendNoteToDB(ctxParent context.Context, mon *ossmon.OSSMon, database *sql.DB, notification *datastore.NotificationInsert, update bool, sourceUpdateTimeOfExistingResource string) error {
	txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRsendNoteToDB, nil, nil)
	ctxParent = newrelic.NewContext(ctxParent, txn)
	defer func() {
		err := txn.End()
		if err != nil {
			log.Println(tlog.Log(), err)
		}
	}()
	// Record start time:
	startTime := time.Now()
	var recordID string
	var err error
	var statusCode int
	ossmon.SetTagsKV(ctxParent,
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
		"pnp-StartTime", startTime.Unix(),
		"pnp-Source", notification.Source,
		"pnp-SourceID", notification.SourceID,
		"pnp-Kind", "notification",
		"pnp-Type", notification.Type,
	)
	var re = regexp.MustCompile(`([0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]) ([0-9][0-9]:[0-9][0-9]:[0-9][0-9]Z)`)
	if re.Match([]byte(notification.SourceUpdateTime)) {
		notification.SourceUpdateTime = strings.Replace(notification.SourceUpdateTime, " ", "T", 1)
		log.Print(tlog.Log()+"Matched a different format( '<date> <time>Z' ).  Changing to ", notification.SourceUpdateTime)
	}
	layout := time.RFC3339
	if !strings.Contains(notification.SourceUpdateTime, "T") {
		layout = "2006-01-02 15:04:05"
	}
	ct, err := helper.CompareTime(notification.SourceUpdateTime, layout, sourceUpdateTimeOfExistingResource, time.RFC3339)
	log.Print(tlog.Log()+"sendNoteToDB: Compare time error: ", err)
	if update && ct <= 0 {
		// Exists, but resource in database is newer
		log.Printf(tlog.Log() + "Not updating resource because source update time is before the source update time already in PG")
		log.Printf(tlog.Log()+"\n%s", notification, sourceUpdateTimeOfExistingResource)
		statusCode = http.StatusOK
	} else if update { // Exists, so update it
		ossmon.SetTag(ctxParent, "pnp-Operation", "update")
		recordID := db.CreateNotificationRecordID(notification.Source, notification.SourceID, notification.CRNFull, notification.IncidentID, notification.Type)
		log.Println(tlog.Log()+"Updating Notification: ", recordID, notification.Source, notification.SourceID, notification.CRNFull, notification.IncidentID)
		err, statusCode = db.UpdateNotification(database, notification)
		log.Println(tlog.Log()+"Notification Updated: ", statusCode, err)
	} else { // Does not exist, so add it
		ossmon.SetTag(ctxParent, "pnp-Operation", "insert")
		if notification.SourceCreationTime == "" {
			notification.SourceCreationTime = notification.SourceUpdateTime
		}
		log.Print(tlog.Log() + "Inserting Notification")
		recordID, err, statusCode = db.InsertNotification(database, notification)
		log.Println(tlog.Log()+"Notification inserted, ", statusCode, ", ", err)
	}
	if statusCode != http.StatusOK {
		ossmon.SetTag(ctxParent, DBFailedErr, true)
		var errMsg string
		if update {
			errMsg = fmt.Sprintf("): Received unexpected http status code (%d) when trying to update notification. error=[%s]", statusCode, err.Error())
		} else {
			errMsg = fmt.Sprintf(": Received unexpected http status code (%d) when trying to add notification. error=[%s]", statusCode, err.Error())
		}
		ossmon.SetError(ctxParent, DBFailedErr+"-"+errMsg)
		return errors.New(errMsg)
	}
	ossmon.SetTag(ctxParent, DBFailedErr, false)
	if recordID != "" {
		log.Printf(tlog.Log()+"Added notification with recordID %s", recordID)
	}
	pnpUpdateTime := time.Now()
	sourceCreationTime, _ := parseStringToTime(notification.SourceCreationTime)
	sourceUpdateTime, _ := parseStringToTime(notification.SourceUpdateTime)
	wait := pnpUpdateTime.Unix() - sourceUpdateTime
	ossmon.SetTagsKV(ctxParent,
		"pnp-Duration", time.Since(startTime),
		"sourceCreationTimeString", notification.SourceCreationTime,
		"sourceUpdateTimeString", notification.SourceUpdateTime,
		"pnpUpdateTimeString", pnpUpdateTime.Format("2006-01-02T15:04:05Z"),
		"sourceCreationTime", sourceCreationTime,
		"sourceUpdateTime", sourceUpdateTime,
		"pnpUpdateTime", pnpUpdateTime.Unix(),
		"shortDescription", notification.ShortDescription[0].Name,
		"recordID", recordID,
		"pnp-WaitTime", wait,
	)
	return err
}

func noteExistsInDB(database *sql.DB, note *datastore.NotificationInsert) (result bool, sourceUpdateTime string) {
	recordID := db.CreateNotificationRecordID(note.Source, note.SourceID, note.CRNFull, note.IncidentID, note.Type)
	nr, err, code := db.GetNotificationByRecordID(database, recordID)
	if err != nil {
		log.Printf(tlog.Log()+"Could not lookup notification by record ID (%s) therefore will try to add. error=[%s]", recordID, err.Error())
		return false, ""
	}
	if code != http.StatusOK {
		log.Printf(tlog.Log()+"Received bad http status code when trying to lookup notification by record ID (%s) therefore will try to add. error=[%s]", recordID, err)
		return false, ""
	}
	if nr != nil {
		return true, nr.SourceUpdateTime
	}
	return false, ""
}

// Bypass local storage: https://github.ibm.com/cloud-sre/toolsplatform/issues/9491
// checkNotificationForIncident checks if a new notification is needed to be created based on an incident
// if no existing notifications with this incident exists in the database,
// or the incident and notification have different data then a new notification is created
func checkNotificationForIncident(database *sql.DB, incident *incidentFromQueue, incEnvList string) {
	log.Print(tlog.Log() + "Starting...")
	// https://github.ibm.com/cloud-sre/pnp-nq2ds/issues/126
	// When an IaaS incident comes in, need to check the notification related fields
	// Check if the notification already exists and compare the related fields to see if an update to the notification is necessary
	// If the notification doesn't exist, then it needs to be created.
	if incident.ServiceName == nil {
		log.Print(tlog.Log() + "incident.ServiceName is nil, returning without processing further")
		return
	}
	isIaaSIncident, err := osscatalog.ServiceNameHasTag(ctxtContext, *incident.ServiceName, osstags.PnPEnabledIaaS)
	if err != nil {
		log.Println(tlog.Log(), err)
	}
	// Only perform the special logic for IaaS services if we are before the BSPN cutover time:
	if isIaaSIncident && isIncidentCreationTimeBeforeIaaSBSPNCutover(incident) {
		for _, crn := range *incident.CRNFull {

			// Check to make sure its a parent or not
			crnSegments := strings.Split(crn, ":")
			if len(crnSegments) != 10 || crnSegments[4] == "" {
				log.Print(tlog.Log(), errors.New("Invalid CRN: "+crn))
				return
			}
			//Transform crn to url query form
			var crnFullQuery string
			crnFullQueryParam, ok := db.CRNStringToQueryParms(crn)
			if ok {
				crnFullQuery = crnFullQueryParam.Encode()
			}
			// Checking to see if its a catalog parent or not.
			// If so, make a loop of crn queries to process
			// else just process the crn as is
			resources, _, errRes, rc := db.GetResourceByQuery(database, crnFullQuery, 0, 0)
			if errRes != nil {
				log.Print(tlog.Log(), "ERROR ALERT: ", errRes)
			}
			log.Printf(tlog.Log()+"DEBUG: resources from query %s:  %+v", crnFullQuery, resources)
			if len(*resources) > 1 {
				log.Print(tlog.Log(), rc, ":ERROR ALERT: length of resources should only be 1 or 0: ", len(*resources), "\n\t", *resources)
			}
			if rc == 200 && len(*resources) == 1 {
				catParentID := (*resources)[0].RecordID
				resourcesToProcess, _, err, _ := db.GetResourceByQuery(database, "catalog_parent_resource_id="+catParentID, 0, 0)
				log.Println(tlog.Log(), "\nGetResourceByQuery: Got child resources \n\terror: ", err, "\n\tnum results: ", len(*resourcesToProcess))
				crnsToProcess := []string{crn}
				if (*resources)[0].IsCatalogParent {
					for _, v := range *resourcesToProcess {
						crnsToProcess = append(crnsToProcess, v.CRNFull)
					}
				}
				log.Print(tlog.Log(), "crnsToProcess ", crnsToProcess)
				var isPrimaryNotification = false
				for _, crnToProcess := range crnsToProcess {
					builtCRN := ""
					if crn == crnToProcess {
						builtCRN = hooks.BuildCRN(crnToProcess, *incident.ServiceName)
					} else {
						builtCRN = crnToProcess
					}
					// check notification by record_id
					recordID := db.CreateNotificationRecordID(*incident.Source, *incident.SourceID, builtCRN, "", "incident")
					log.Println(tlog.Log(), "searching for notification with recordID="+recordID)
					notification, err, status := db.GetNotificationByRecordID(database, recordID)
					if err != nil && status != http.StatusOK {
						log.Println(tlog.Log(), err)
					}
					// create notification
					n := incidentToNotification(incident, incEnvList, builtCRN)
					notificationMsg := datastore.NotificationMsg{MsgType: Update, NotificationInsert: *n}
					if !isPrimaryNotification {
						isPrimaryNotification = true
						notificationMsg.IsPrimary = true
					}
					if notification != nil {
						log.Println(tlog.Log(), "notification found")
						log.Println(tlog.Log(), "DEBUG: Found the following notification:", *notification)
						if isNewNotificationNeeded(notification, incident) {
							produceNotificationToMQ(&notificationMsg)
						}
					} else {
						log.Println(tlog.Log(), "Notification not found in PG, creating one for", *incident.SourceID, "crn="+builtCRN)
						produceNotificationToMQ(&notificationMsg)
					}
				}
			}
		}
	}
	log.Println(tlog.Log(), "Done")
}

func isIncidentCreationTimeBeforeIaaSBSPNCutover(incident *incidentFromQueue) bool {
	result := true // assume true because this is the behaviour before the cutover
	// IaaS services will start using BSPN May 28, 2019 at 9 AM central time (which is May 28, 2019 2 PM UTC):
	iaasCutoverToBSPNTime := time.Date(2019, time.May, 28, 14, 0, 0, 0, time.UTC)
	if incident != nil && incident.SourceCreationTime != nil && *incident.SourceCreationTime != "" && *incident.SourceCreationTime != "Z" {
		// Check whether the incident creation time is before the IaaS cutover time to BSPN:
		result = helper.IsNewTimeBeforeExistingTime(*incident.SourceCreationTime, "2006-01-02 15:04:05Z", iaasCutoverToBSPNTime.Format("2006-01-02T15:04:05Z"), time.RFC3339)
	} else {
		// Do not have sufficient information from the incident to make a determination, so check if the current time is before the IaaS cutover time to BSPN:
		result = time.Now().UTC().Before(iaasCutoverToBSPNTime)
	}
	log.Println(tlog.Log(), result)
	return result
}

// produceNotificationToMQ posts notification to MQ
func produceNotificationToMQ(notification *datastore.NotificationMsg) {
	//wrap in mq.NotificationMsg struct for https://github.ibm.com/cloud-sre/pnp-nq2ds/issues/172
	noteMsg := datastore.NotificationMsg{MsgType: Update, NotificationInsert: notification.NotificationInsert}
	n, err := json.Marshal(noteMsg)
	if err != nil {
		log.Println(tlog.Log() + err.Error())
	}
	var nProducer *rabbitmq.AMQPProducer
	if RabbitmqEnableMessages {
		nProducer = rabbitmq.NewSSLProducer(RabbitmqURLs, RabbitmqTLSCert, NotificationRoutingKey, MQExchangeName, "direct")
	} else {
		nProducer = rabbitmq.NewProducer(RabbitmqURLs, NotificationRoutingKey, MQExchangeName, "direct")
	}
	log.Println(tlog.Log()+"the following notification will be produced to MQ with routing key="+NotificationRoutingKey+":", string(n))
	// Encrypt message
	encryptedData, err := encryption.Encrypt(string(n))
	if err != nil {
		log.Print(tlog.Log()+"Error occurred trying to encrypt data, err = ", err)
	}
	err = nProducer.ProduceOnce(string(encryptedData))
	if err != nil {
		log.Printf(tlog.Log()+"Could not post the following notification to rabbitmq= "+string(n)+", err :", err.Error())
		log.Println(tlog.Log(), "Could not post the following notification to rabbitmq=", string(n))
	}
}

// isNewNotificationNeeded checks if the notification found in the database needs to be updated
// the functions check if some fields are the same, if they are not then we create a new notification
func isNewNotificationNeeded(notification *datastore.NotificationReturn, incident *incidentFromQueue) bool {
	if incident.OutageStartTime != nil && notification.EventTimeStart == *incident.OutageStartTime {
		if incident.OutageEndTime != nil && notification.EventTimeEnd == *incident.OutageEndTime {
			if incident.Source != nil && notification.Source == *incident.Source {
				if incident.CustomerImpactDescription != nil && notification.LongDescription[0].Name == *incident.CustomerImpactDescription {
					if incident.SourceID != nil && notification.SourceID == *incident.SourceID {
						if incident.ServiceName != nil && notification.ResourceDisplayNames[0].Name == getDisplayName(*incident.ServiceName) {
							if isPnPRemoved(incident) != notification.PnPRemoved {
								// if they are all the same then no need to create a new notification
								// it means there already is a notification for this incident and they are the same
								return false
							}
						}
					}
				}
			}
		}
	}
	return true
}

// getCategoryID returns the categoryID for a service name
func getCategoryID(ctx ctxt.Context, serviceName string) string {
	catID, err := osscatalog.ServiceNameToCategoryID(ctx, serviceName)
	if err != nil {
		log.Println(tlog.Log(), err)
		return ""
	}
	return catID
}

// getDisplayName returns the display name from Global Catalog for a specific service name
func getDisplayName(serviceName string) string {
	displayName, err := osscatalog.CategoryIDToDisplayName(ctxtContext, getCategoryID(ctxtContext, serviceName))
	if err != nil {
		log.Println(tlog.Log(), err)
		return ""
	}
	return displayName
}

// incidentToNotification converts an incident to a notification
func incidentToNotification(incident *incidentFromQueue, incidentRegionList, builtCRN string) *datastore.NotificationInsert {
	var err error
	if incident.SourceID != nil {
		log.Println(tlog.Log()+"Creating notification for incident", *incident.SourceID)
	} else {
		log.Println(tlog.Log() + "Creating notification for incident")
	}
	notification := new(datastore.NotificationInsert)
	if incident.SourceUpdateTime != nil {
		notification.SourceUpdateTime = *incident.SourceUpdateTime
	}
	if incident.OutageStartTime != nil {
		notification.EventTimeStart = *incident.OutageStartTime
	}
	if incident.OutageEndTime != nil {
		notification.EventTimeEnd = *incident.OutageEndTime
	}
	notification.Source = "servicenow"
	if incident.SourceID != nil {
		notification.SourceID = *incident.SourceID
	}
	notification.Type = "incident"
	if incident.SourceID != nil {
		notification.IncidentID = *incident.SourceID
	}
	aSep := "a"
	servicename := strings.Title(strings.Split(builtCRN, ":")[4])
	shortdesc := ""

	if incident.AffectedActivity == nil || *incident.AffectedActivity == "" {
		shortdesc = servicename + " has an issue in " + strings.Split(builtCRN, ":")[5] + "."
	} else {
		value := *incident.AffectedActivity
		switch value[0] {
		case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
			aSep = "an"
		}
		shortdesc = servicename + " has " + aSep + " " + *incident.AffectedActivity + " issue in " + strings.Split(builtCRN, ":")[5] + "."
	}
	notification.ShortDescription = []datastore.DisplayName{{Language: "en", Name: shortdesc}}
	if incident.CustomerImpactDescription == nil || *incident.CustomerImpactDescription == "" {
		notification.LongDescription = []datastore.DisplayName{{Language: "en", Name: "  "}}
	} else {
		notification.LongDescription = []datastore.DisplayName{{Language: "en", Name: *incident.CustomerImpactDescription}}
	}
	if incident.SourceCreationTime != nil {
		notification.SourceCreationTime = *incident.SourceCreationTime
	}
	notification.CRNFull = builtCRN
	if incident.ServiceName != nil {
		catID := getCategoryID(ctxtContext, *incident.ServiceName)
		notification.Category, err = hooks.GetEntryType(ctxtContext, catID)
		if err != nil {
			log.Println(tlog.Log(), err)
			notification.Category = ""
		}
		notification.ResourceDisplayNames, err = hooks.GetDisplayNames(ctxtContext, catID, *incident.ServiceName)
		if err != nil {
			log.Println(tlog.Log(), err)
			notification.ResourceDisplayNames = []datastore.DisplayName{{}}
		}
	} else {
		notification.Category = ""
		notification.ResourceDisplayNames = []datastore.DisplayName{{}}
	}
	notification.PnPRemoved = isPnPRemoved(incident)
	return notification
}

// InitNotificationsAdapter sets up notifications-adapter
func InitNotificationsAdapter() {

	//Don't need *api.SourceConfig
	_, err := initadapter.Initialize()
	if err != nil {
		log.Println(tlog.Log(), err)
	}

	monitorEx, err := exmon.CreateMonitor()
	if err != nil {
		log.Println(tlog.Log(), err)
	}
	ctxtContext.NRMon = monitorEx
	ctxtContext.LogID = "nq2ds"

}

// Produce notifications for an updated incident
// Returns true if retry is needed
func notificationForIncidentUpdate(database *sql.DB, incident *incidentFromQueue, pnpRemoved bool, incEnvList string) bool {

	var (
		sourceCreatedTime           time.Time
		primaryNotificationSourceID string
		primaryNotificationRecordID string
	)

	log.Print(tlog.Log() + "Starting...")

	if incident.ServiceName == nil {
		log.Print(tlog.Log(), "incident.ServiceName is nil, returning without processing further")
		return false // No service name, will not retry
	}

	if incident.SourceID == nil {
		log.Print(tlog.Log(), "incident.SourceID is nil, returning without processing further")
		return false // No service ID, will not retry
	}

	log.Println(tlog.Log(), "Getting for notifications from IncidentID="+*incident.SourceID)
	notificationReturns, err, status := db.GetNotificationsByIncidentID(database, *incident.SourceID)
	if err != nil && status != http.StatusOK {
		log.Println(tlog.Log(), err)
		log.Println(tlog.Log() + "Will restart...")
		return true // Failed in retrieving existing notifications, will retry
	}

	// Check for the primary notification.  Check for the last notification with a different id
	for _, notificationReturn := range *notificationReturns {
		if notificationReturn.PnPRemoved {
			continue
		}

		if primaryNotificationSourceID != notificationReturn.SourceID {
			notificationReturnsCreateTime, err := dateparse.ParseLocal(notificationReturn.SourceCreationTime)
			if err != nil {
				log.Fatalln(tlog.Log(), "Unable to parse ", notificationReturn.SourceCreationTime, "\n", err)
			}

			if sourceCreatedTime.IsZero() ||
				notificationReturnsCreateTime.After(sourceCreatedTime) {

				primaryNotificationSourceID = notificationReturn.SourceID
				primaryNotificationRecordID = notificationReturn.RecordID
				sourceCreatedTime = notificationReturnsCreateTime
			}
		}

	}

	log.Println(tlog.Log(), "Got notifications from IncidentID="+*incident.SourceID)
	for i, notificationReturn := range *notificationReturns {
		log.Println(tlog.Log(), "Found notification: ", i, notificationReturn)

		// If you've never seen a hack before, here is one:
		// need to convert the pointer to a slice of datastore.NotificationResult to a
		// regular slice of datastore.NotificationReturn to avoid triggering
		// gosec scan rule:
		//
		// G601 (CWE-118): Implicit memory aliasing in for loop. (Confidence: MEDIUM, Severity: MEDIUM)
		//
		// After each call to isNewNotificationNeeded we release the slice.
		var toCheck []datastore.NotificationReturn
		toCheck = append(toCheck, notificationReturn)
		log.Println(tlog.Log(), "toCheck length: ", len(toCheck), " index at: ", i)
		if len(toCheck) > 0 {
			// We always use index 0, the slice is set to nil right after the if that follows
			// very inefficient but this code should not be used when shared.BypassLocalStorage is set to true.
			if isNewNotificationNeeded(&toCheck[0], incident) {
				notification := incidentToNotification(incident, incEnvList, notificationReturn.CRNFull)
				notificationMsg := &datastore.NotificationMsg{}

				if primaryNotificationRecordID == notificationReturn.RecordID {
					notificationMsg.IsPrimary = true
				}

				notification = &datastore.NotificationInsert{}
				notification.CRNFull = notificationReturn.CRNFull
				notification.Source = notificationReturn.Source
				notification.SourceID = notificationReturn.SourceID
				notification.IncidentID = notificationReturn.IncidentID
				notification.LongDescription = notificationReturn.LongDescription
				notification.RecordRetractionTime = notificationReturn.RecordRetractionTime
				notification.ShortDescription = notificationReturn.ShortDescription
				notification.SourceCreationTime = notificationReturn.SourceCreationTime
				notification.Tags = notificationReturn.Tags
				notification.Category = notificationReturn.Category
				notification.EventTimeEnd = notificationReturn.EventTimeEnd
				notification.EventTimeStart = notificationReturn.EventTimeStart
				notification.ResourceDisplayNames = notificationReturn.ResourceDisplayNames
				notification.Type = notificationReturn.Type

				// Note that the BSPN won't have the same update time since this was driven by the incident
				notification.SourceUpdateTime = *incident.SourceUpdateTime
				notification.PnPRemoved = pnpRemoved
				notificationMsg.NotificationInsert = *notification
				produceNotificationToMQ(notificationMsg)
			}
			// finished with the slice release it
			toCheck = nil
		}
	}
	log.Println(tlog.Log() + "Done")
	return false // All done (no retry)
}

func hasValidIncidentID(database *sql.DB, notification *datastore.NotificationMsg) (err error) {

	if notification.Source != "servicenow" {
		return
	}
	if strings.HasPrefix(notification.IncidentID, "INC") {
		_, err, _ = db.GetIncidentBySourceID(database, notification.Source, notification.IncidentID)
		log.Print("")
		return err
	}
	if strings.HasPrefix(notification.IncidentID, "CHG") {
		_, err, _ = db.GetMaintenanceBySourceID(database, notification.Source, notification.IncidentID)
		return err
	}
	return
}
