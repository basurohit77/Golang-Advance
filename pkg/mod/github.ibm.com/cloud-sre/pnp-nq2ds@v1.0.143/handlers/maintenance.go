package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.ibm.com/cloud-sre/pnp-nq2ds/shared"
	"log"
	"strings"
	"sync"
	"time"

	newrelic "github.com/newrelic/go-agent"

	"github.com/araddon/dateparse"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/ossmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-abstraction/exmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/initadapter"
	"github.ibm.com/cloud-sre/pnp-nq2ds/helper"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
)

var completionCodes = map[string]string{"successful": "successful", "successful_issues": "successful", "unsuccessful": "failed", "cancelled": "cancelled", "missed-window": "missed-window"}

type ResourceMapCache struct {
	Resources map[string]*datastore.ResourceReturn
	sync.RWMutex
	updateTime time.Time
}

func (rm *ResourceMapCache) get(key string) (value *datastore.ResourceReturn, ok bool) {
	rm.RLock()
	result, ok := rm.Resources[key]
	rm.RUnlock()
	return result, ok
}

func (rm *ResourceMapCache) swap(value map[string]*datastore.ResourceReturn) {
	rm.Lock()
	rm.Resources = value
	rm.updateTime = time.Now()
	rm.Unlock()
}

type SNMaintenanceMap struct {
	SNMaintenances map[string]*datastore.MaintenanceReturn
	sync.RWMutex
}

func (rm *SNMaintenanceMap) get(key string) (value *datastore.MaintenanceReturn, ok bool) {
	rm.RLock()
	result, ok := rm.SNMaintenances[key]
	rm.RUnlock()
	return result, ok
}

func (rm *SNMaintenanceMap) swap(value map[string]*datastore.MaintenanceReturn) {
	rm.Lock()
	rm.SNMaintenances = value
	rm.Unlock()
}

func (rm *SNMaintenanceMap) put(key string, value *datastore.MaintenanceReturn) {
	rm.Lock()
	rm.SNMaintenances[key] = value
	rm.Unlock()
}

type Communication struct {
	SysID              string `json:"sys_id"`
	Number             string `json:"number"`
	Stage              string `json:"stage"`
	ShortDescription   string `json:"short_description"`
	Text               string `json:"text"`
	UChangeDescription string `json:"u_change_description"`
	UChangeReason      string `json:"u_change_reason"`
	UChangeImpact      string `json:"u_change_impact"`
	SysCreatedBy       string `json:"sys_created_by"`
	SysUpdatedBy       string `json:"sys_updated_by"`
	SysCreatedOn       string `json:"sys_created_on"`
	SysUpdatedOn       string `json:"sys_updated_on"`
	PublishDate        string `json:"publish_date"`
}

// SNChange servicenow Change Request record mapped to PG maintenance_table
type SNChange struct {
	Operation        string          `json:"operation"`
	SysID            string          `json:"sys_id"`
	Number           string          `json:"number"`
	SysCreatedOn     string          `json:"sys_created_on"`
	State            string          `json:"state"`
	Priority         string          `json:"priority"`
	ShortDescription string          `json:"short_description"`
	Description      string          `json:"description"`
	USeverity        string          `json:"u_severity"`
	UPurposeGoal     string          `json:"u_purpose_goal"`
	BackoutPlan      string          `json:"backout_plan"`
	StartDate        string          `json:"start_date"`
	EndDate          string          `json:"end_date"`
	WorkStart        string          `json:"work_start"`
	WorkEnd          string          `json:"work_end"`
	UOutageDuration  int             `json:"u_outage_duration"`
	SysCreatedBy     string          `json:"sys_created_by"`
	SysUpdatedBy     string          `json:"sys_updated_by"`
	SysUpdatedOn     string          `json:"sys_updated_on"`
	Crn              []string        `json:"crn"`
	Communications   []Communication `json:"communications"`
	Instance         string          `json:"instance"`
	CloseCode        string          `json:"close_code"`
	TargetedURL      string          `json:"u_targeted_notification_url"`
	Audience         string          `json:"u_audience"`
}

// Maintenance - maintenance record copied from the pnp-change-adapter
// repository as the pnp-change-adapter repository is no longer used.
type Maintenance struct {
	SourceCreationTime    string   `json:"source_creation_time,omitempty"`
	SourceUpdateTime      string   `json:"source_update_time,omitempty"`
	PlannedStartTime      string   `json:"planned_start_time,omitempty"`
	PlannedEndTime        string   `json:"planned_end_time,omitempty"`
	ShortDescription      string   `json:"short_description,omitempty"`
	LongDescription       string   `json:"long_description,omitempty"`
	CRNFull               []string `json:"crnFull,omitempty"`
	State                 string   `json:"state,omitempty"`
	Disruptive            bool     `json:"disruptive,omitempty"`
	SourceID              string   `json:"source_id,omitempty"`
	Source                string   `json:"source,omitempty"`
	RegulatoryDomain      string   `json:"regulatory_domain,omitempty"`
	RecordHash            string   `json:"record_hash,omitempty"`
	MaintenanceDuration   int      `json:"maintenance_duration,omitempty"`
	DisruptionType        string   `json:"disruption_type,omitempty"`
	DisruptionDescription string   `json:"disruption_description,omitempty"`
	DisruptionDuration    int      `json:"disruption_duration,omitempty"`
	SourceState           string   `json:"source_state"`
	NotificationStatus    string   `json:"notification_status"`
	NotificationType      string   `json:"notification_type"`
	NotificationChannels  string   `json:"notification_channels"`
	CompletionCode        string   `json:"completion_code,omitempty"`
	TargetedURL           string   `json:"u_targeted_notification_url,omitempty"`
	Audience              string   `json:"u_audience,omitempty"`
}

// MaintenanceMsg - maintenance message record copied from the pnp-change-adapter
// repository as the pnp-change-adapter repository is no longer used.
type MaintenanceMsg struct {
	Info      Maintenance `json:"info"`
	Tags      string      `json:"tags"`
	Operation string      `json:"operation"`
}

var (
	snMaintenanceMap = SNMaintenanceMap{SNMaintenances: make(map[string]*datastore.MaintenanceReturn)}
	resourceMapCache = ResourceMapCache{Resources: make(map[string]*datastore.ResourceReturn), updateTime: time.Now()}
)

const utcSuffix = "-00"

// checkCommunicationsPublishDate validate that at least one communication in that list has a publish_date that is not the empty string. That is, it has some value
// https://github.ibm.com/cloud-sre/ToolsPlatform/issues/13229
func checkCommunicationsPublishDate(communications []Communication) bool {
	for _, c := range communications {
		if c.PublishDate != "" {
			return true
		}
	}
	return false
}

// ProcessSNMaintenance - processes new or updated SN maintenance records, returns
func ProcessSNMaintenance(database *sql.DB, decryptedMsg []byte, isBulk bool, mon *ossmon.OSSMon) (maintenanceMaps []datastore.MaintenanceMap, notifications []datastore.NotificationMsg) {
	log.Print(tlog.Log()+"Starting...:", string(api.RedactAttributes(decryptedMsg)))
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	defer span.Finish()
	ossmon.SetTagsKV(ctx,
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
	)
	now := time.Now()
	maintenanceReturns := []datastore.MaintenanceReturn{}
	notifications = []datastore.NotificationMsg{}
	maintenanceLongDescription := ""

	inserted := []string{}
	updated := []string{}
	readyToPost := []string{}
	// SNRecordsOut - array of SNRecordOut to NQ2DS
	type SNRecordsOut struct {
		Results []*datastore.MaintenanceInsert `json:"result_from_sn"`
	}
	// Create an array of MaintenanceInserts out of the payload
	snRecords := SNRecordsOut{}
	if isBulk {
		err := json.Unmarshal(decryptedMsg, &snRecords)
		if err != nil {
			errMsg := "Error occurred unmarshaling maintenance message to SNRecordsOut, err = " + err.Error()
			log.Print(tlog.Log() + errMsg)
			ossmon.SetTag(ctx, DecryptionErr, errMsg)
			ossmon.SetError(ctx, DecryptionErr+"-"+errMsg)
			return nil, nil
		}
	} else {
		snRecord := datastore.MaintenanceInsert{}
		snRecord.Source = "servicenow"
		snChangeRecord := SNChange{}
		err := json.Unmarshal(decryptedMsg, &snChangeRecord)
		if err != nil {
			errMsg := "Error occurred unmarshaling maintenance message to snChangeRecord, err = " + err.Error()
			ossmon.SetTag(ctx, DecryptionErr, errMsg)
			ossmon.SetError(ctx, DecryptionErr+"-"+errMsg)
			log.Print(tlog.Log() + errMsg)
			return nil, nil
		}

		if strings.ToLower(snChangeRecord.State) == "new" {
			ossmon.SetTag(ctx, "snChangeRecord", "Ignoring because in new state")
			log.Print(tlog.Log() + "Ignoring because in new state")
			return nil, nil
		}

		if &snChangeRecord.UOutageDuration != nil && snChangeRecord.UOutageDuration > 0 {
			snRecord.Disruptive = true
		} else {
			errMsg := "Ignoring because the Change(" + snChangeRecord.Number + ") is not disruptive"
			ossmon.SetTag(ctx, "snChangeRecord", errMsg)
			log.Print(tlog.Log(), errMsg)
			return nil, nil
		}

		var maintenance *datastore.MaintenanceReturn
		// Get the maintenance from the db.   Checking to see if we have an existing record.
		// Otherwise it never made it through the scheduled or implement state and we don't care about it.
		if !shared.BypassLocalStorage {
			recordID := db.CreateRecordIDFromSourceSourceID(snRecord.Source, snChangeRecord.Number)
			log.Println(tlog.Log()+"db.CreateRecordIDFromSourceSourceID(snRecord.Source, snRecord.SourceID)", snRecord.Source, snChangeRecord.Number)
			var httpError int
			maintenance, err, httpError = db.GetMaintenanceByRecordIDStatement(database, recordID)
			if err != nil {
				errMsg := fmt.Sprintf("Error getting sn maintenances from PG: %d %s ", httpError, err.Error())
				ossmon.SetTag(ctx, DBFailedErr, errMsg)
				ossmon.SetError(ctx, DBFailedErr+"-"+errMsg)
				log.Printf(tlog.Log() + errMsg)

			} else {
				newSNMaintenanceMap := map[string]*datastore.MaintenanceReturn{}
				newSNMaintenanceMap[maintenance.Source+maintenance.SourceID] = maintenance
				snMaintenanceMap.swap(newSNMaintenanceMap)
			}
		}
		if strings.ToLower(snChangeRecord.State) == "implement" || strings.ToLower(snChangeRecord.State) == "scheduled" || maintenance != nil {
			if snChangeRecord.Number != "" {
				snRecord.SourceID = snChangeRecord.Number
			}
			if snChangeRecord.SysCreatedOn != "" {
				snRecord.SourceCreationTime = snChangeRecord.SysCreatedOn + utcSuffix
			}
			if snChangeRecord.SysUpdatedOn != "" {
				snRecord.SourceUpdateTime = snChangeRecord.SysUpdatedOn + utcSuffix
			}

			{
				outageTimeSecs := snChangeRecord.UOutageDuration
				if snChangeRecord.UOutageDuration < 60 {
					snRecord.DisruptionDuration = 1
				} else {
					snRecord.DisruptionDuration = outageTimeSecs / 60
				}
			}

			if snChangeRecord.ShortDescription != "" {
				snRecord.ShortDescription = snChangeRecord.ShortDescription
			}
			// Cleanup and remove crn's as needed
			if len(snChangeRecord.Crn) > 0 {
				// modify CRN's in case of category parent:
				log.Print(tlog.Log()+"crns = ", snChangeRecord.Crn)
				snChangeRecord.Crn, err = normalizeCRNServiceNames(snChangeRecord.Crn)
				if err != nil {
					errMsg := "Error during crn normalize" + err.Error()
					log.Print(tlog.Log(), errMsg)
					ossmon.SetTag(ctx, CRNErr, errMsg)
					ossmon.SetError(ctx, CRNErr+"-"+errMsg)
					return
				}
				validCRNs := []string{}
				for _, crn := range snChangeRecord.Crn {
					svc, err := api.GetServiceFromCRN(crn)
					if err != nil {
						log.Println(tlog.Log(), err)
						ossmon.SetError(ctx, "GetServiceFromCRN "+err.Error())
					}
					if api.IsCrnPnpValid(crn) || svc == "pnp-api-oss" {
						validCRNs = append(validCRNs, crn)
					}
				}
				snRecord.CRNFull = snChangeRecord.Crn
			}
			// no crns.  Nothing to process
			if len(snChangeRecord.Crn) == 0 {
				return maintenanceMaps, notifications
			}

			if snChangeRecord.TargetedURL != "" {
				snRecord.TargetedURL = snChangeRecord.TargetedURL
			}
			if snChangeRecord.Audience != "" {
				snRecord.Audience = snChangeRecord.Audience
			} else {
				snRecord.Audience = db.SNnill2PnP
			}
			if snChangeRecord.EndDate != "" {
				snRecord.PlannedEndTime = snChangeRecord.EndDate + utcSuffix
			}

			if snChangeRecord.StartDate != "" {
				snRecord.PlannedStartTime = snChangeRecord.StartDate + utcSuffix
			}
			layout := "2006-01-02 15:04:05"
			endTime, err := time.Parse(layout, snChangeRecord.EndDate)
			if err != nil {
				log.Print(tlog.Log(), "Error parsing endTime: ", err)
			}
			startTime, err := time.Parse(layout, snChangeRecord.StartDate)
			if err != nil {
				log.Print(tlog.Log(), "Error parsing startTime: ", err)
			}
			diff := endTime.Sub(startTime)
			snRecord.MaintenanceDuration = int(diff.Nanoseconds()/int64(time.Millisecond)) / 1000 / 60
			if snChangeRecord.State != "" {
				if strings.ToLower(snChangeRecord.State) == "implement" {
					snRecord.State = "in-progress"
				} else {
					snRecord.State = strings.ToLower(snChangeRecord.State)
				}

				if strings.ToLower(snRecord.State) == "closed" || strings.ToLower(snRecord.State) == "cancelled" {
					if completionCode, ok := completionCodes[strings.ToLower(snChangeRecord.CloseCode)]; ok {
						snRecord.CompletionCode = completionCode
					} else {
						snRecord.CompletionCode = "successful"
					}
					snRecord.State = "complete"
				}
			}
			// Post any communications.
			// First update v if needed
			if resourceMapCache.updateTime.Add(time.Second * 900).Before(now) {
				log.Print(tlog.Log(), "Updating resourceMapCache")
				newResourceMapCache := map[string]*datastore.ResourceReturn{}
				resourceArray, err := db.GetAllResources(database)
				if err != nil {
					log.Panic(tlog.Log(), "db.GetAllResources error: ", err)
				}
				log.Print(tlog.Log(), "Number of resources: ", len(*resourceArray))
				for i := range *resourceArray {
					newResourceMapCache[(*resourceArray)[i].SourceID] = &(*resourceArray)[i]
				}
				resourceMapCache.swap(newResourceMapCache)
			}
			ctx, _ := buildContext()
			lastTime := time.Time{}
			// https://github.ibm.com/cloud-sre/ToolsPlatform/issues/13229
			// Add the following lines in once the ServiceNow team adds publish date to the incomming communications.
			//atLeastOnePublishDate := checkCommunicationsPublishDate(snChangeRecord.Communications)
			//if !atLeastOnePublishDate {
			//	log.Println(tlog.Log()+"there is no publish date along with the change record number: ",snChangeRecord.Number)
			//	return nil, nil
			//}
			// for long description.  We should only ever have 1 communication.
			// But if we have more than 1, then we use the last communication by time/date.
			for _, v := range snChangeRecord.Communications {
				notificationTime, err := dateparse.ParseAny(v.SysUpdatedOn)
				log.Println(tlog.Log()+"Parsed time: ", notificationTime, "\nLast Time: ", lastTime)
				if err != nil {
					log.Println(tlog.Log(), "Error while trying to parse the updated on timestamp ", v.SysUpdatedOn, err)
				} else {
					if notificationTime.After(lastTime) {
						maintenanceLongDescription = v.Text
						lastTime = notificationTime
					}
				}
				log.Printf(tlog.Log()+"Communication record: %+v", v)
				if v.Stage == "active" || v.Stage == "sent" {
					for _, crn := range snChangeRecord.Crn {
						// gavila - 2020/09/24
						// issue 294: subscription-consumer can send out notifications that do not exist
						// This can happen when CRN validation fails before a record is inserted. We add
						// the same check here to avoid a notification without a corresponding record.
						if !api.IsCrnPnpValid(crn) {
							log.Printf("Invalid CRN %q encountered. Will not add to notifications", crn)
							continue
						}
						log.Printf(tlog.Log() + "Communication record is active or sent.   Will process")
						notificationInsert := datastore.NotificationMsg{}
						notificationInsert.MsgType = datastore.Update
						if !notificationInsert.IsPrimary && !notificationInsert.PnPRemoved {
							notificationInsert.IsPrimary = true
						}

						notificationInsert.SourceCreationTime = v.SysCreatedOn + utcSuffix
						notificationInsert.SourceUpdateTime = v.SysUpdatedOn + utcSuffix
						notificationInsert.EventTimeStart = snChangeRecord.StartDate + utcSuffix
						notificationInsert.EventTimeEnd = snChangeRecord.EndDate + utcSuffix
						notificationInsert.Source = "servicenow"
						notificationInsert.SourceID = v.Number
						notificationInsert.Type = "maintenance"
						notificationInsert.IncidentID = snChangeRecord.Number
						notificationInsert.MaintenanceDuration = snRecord.MaintenanceDuration
						notificationInsert.DisruptionDuration = snRecord.DisruptionDuration
						notificationInsert.DisruptionDescription = snRecord.DisruptionDescription
						notificationInsert.DisruptionType = snRecord.DisruptionType
						notificationInsert.Category = ""
						notificationInsert.CRNFull = crn

						if rd, ok := resourceMapCache.get(crn); ok {
							notificationInsert.ResourceDisplayNames = rd.DisplayNames
							serviceName := strings.Split(crn, ":")[4]

							categoryID, err := osscatalog.ServiceNameToCategoryID(ctx, serviceName)
							if err != nil {
								log.Panic(tlog.Log(), "ServiceNameToCategoryID error: ", err, "\n ctx: ", ctx, "\t servicename: ", serviceName)
							}

							entryType, err := getEntryType(ctx, categoryID)
							if err != nil {
								log.Panic(tlog.Log(), "getEntryType error: ", err, "\n categoryId: ", categoryID)
							}
							log.Println(tlog.Log(), "Entry Type:", entryType)
							notificationInsert.Category = entryType
							tags := ""
							for index, tag := range rd.Tags {
								if index == 0 {
									tags += tag.ID
								} else {
									tags += "," + tag.ID
								}
							}
						}
						notificationInsert.ShortDescription = append(notificationInsert.ShortDescription, datastore.DisplayName{Language: "en", Name: v.ShortDescription})
						notificationInsert.LongDescription = append(notificationInsert.LongDescription, datastore.DisplayName{Language: "en", Name: v.Text})
						log.Printf(tlog.Log()+"Notification to append to array for processing: %+v", notificationInsert)
						notifications = append(notifications, notificationInsert)
					}
				} else {
					log.Println(tlog.Log() + "Communication record (" + snRecord.State + ") is not active or sent. Will ignore")
				}
			}
			snRecord.LongDescription = maintenanceLongDescription
			log.Printf(tlog.Log()+"snRecord: %+v", snRecord)
			snRecords.Results = append(snRecords.Results, &snRecord)
		} else {
			log.Print(tlog.Log(), "Change is not implement or scheduled. Nothing to do. ")
			return maintenanceMaps, notifications
		}
	}

	// Bypass local storage: https://github.ibm.com/cloud-sre/toolsplatform/issues/9491
	//Make sure we have the latest list of maintenances.    If it's 5 minutes old, just get a new one from the db
	//if isBulk {
	//	newSNMaintenanceMap := map[string]*datastore.MaintenanceReturn{}
	//	maintenanceArray, err, httpError := db.GetAllSNMaintenances(database)
	//	if err != nil {
	//		errMsg := fmt.Sprintf("Error getting sn maintenances from PG: %d %s ", httpError, err.Error())
	//		ossmon.SetError(ctx, DBFailedErr+"-"+errMsg)
	//		ossmon.SetTag(ctx, DBFailedErr, errMsg)
	//		log.Print(tlog.Log() + errMsg)
	//		return nil, nil
	//	}
	//	log.Print(tlog.Log()+"Total sn maintenances from PG: ", len(maintenanceArray))
	//	for i, v := range maintenanceArray {
	//		newSNMaintenanceMap[v.Source+v.SourceID] = maintenanceArray[i]
	//	}
	//	snMaintenanceMap.swap(newSNMaintenanceMap)
	//	log.Printf(tlog.Log()+"snMaintenanceMap Length: %d ", len(snMaintenanceMap.SNMaintenances))
	//}

	skipped := 0
	errored := 0
	// Loop through the MaintenanceInserts that were in the payload.  Identify any maintenances that should be updated in the db
	// Maintenances in the snMaintenanceMap will be updated if the hash doesn't match.  Otherwise we need to insert a new record
	for i := range snRecords.Results {
		// Ignore camel cased version of source.   Getting a bad message. It should be servicenow instead.
		if snRecords.Results[i].Source == "servicenow" {
			snRecords.Results[i].Source = "servicenow"
		}

		log.Print(tlog.Log()+"map length: ", len(snMaintenanceMap.SNMaintenances), "\t iteration: ", i, "/", len(snRecords.Results))
		if smm, ok := snMaintenanceMap.get(snRecords.Results[i].Source + snRecords.Results[i].SourceID); ok {
			// This is required due to the table contraints
			if len(snRecords.Results[i].CRNFull) == 0 {
				snRecords.Results[i].CRNFull = smm.CRNFull
			}
			// Calculate the new record hash
			snRecords.Results[i] = calculateHash(smm, snRecords.Results[i])
			log.Printf(tlog.Log()+"snMaintenanceMap: %+v", smm)
			log.Printf(tlog.Log()+"\nsnRecords.Results[i]: %+v", snRecords.Results[i])
			log.Println(tlog.Log()+"maintenanceLongDescription:", maintenanceLongDescription, "snMaintenanceMap Description:", smm.LongDescription)
			if (maintenanceLongDescription != "" && maintenanceLongDescription != smm.LongDescription) || shouldProcessMaintenance(snRecords.Results[i], snRecords.Results[i].RecordHash) {
				if !shared.BypassLocalStorage {
					err, httpError := db.UpdateMaintenance(database, snRecords.Results[i])
					if err != nil {
						errMsg := fmt.Sprintf("Error updating sn maintenances to PG:  %d %s ", httpError, err.Error())
						ossmon.SetError(ctx, DBFailedErr+"-"+errMsg)
						log.Print(tlog.Log() + errMsg)
						errored++
						continue
					} else {
						log.Printf(tlog.Log()+"Updated SN maintenance to PG: %s/%s/%s", snRecords.Results[i].SourceID, smm.RecordID, smm.RecordHash)
						updated = append(updated, snRecords.Results[i].SourceID)
					}
				}

				readyToPost = append(readyToPost, snRecords.Results[i].SourceID)

				// Add to maintenance returns that will get posted
				maintenanceReturns = append(maintenanceReturns, *smm)
			} else {
				skipped++
				//log.Printf(tlog.Log()+"Did not need to process:  %s ", snRecords.Results[i].SourceID)
			}
		} else {
			log.Printf(tlog.Log()+"SN maintenance not found for : %+v", snRecords.Results[i])
			snRecords.Results[i].RecordHash = db.ComputeMaintenanceRecordHash(snRecords.Results[i])
			// Bypass local storage: https://github.ibm.com/cloud-sre/toolsplatform/issues/9491
			if !shared.BypassLocalStorage {
				maintenanceRecordID, err, httpError := db.InsertMaintenance(database, snRecords.Results[i])
				if err != nil {
					errMsg := fmt.Sprintf("Error inserting SN maintenances to PG: %d %s \n %+v", httpError, err.Error(), snRecords.Results[i])
					ossmon.SetError(ctx, DBFailedErr+"-"+errMsg)
					log.Print(tlog.Log() + errMsg)
					errored++
					continue
				} else {
					log.Print(tlog.Log()+"Inserted SN maintenance to PG: ", maintenanceRecordID)
					inserted = append(inserted, snRecords.Results[i].SourceID)
				}
				// Have to get the maintenance as we don't get it in the insert
				maintenanceReturn, err, httpError := db.GetMaintenanceByRecordIDStatement(database, maintenanceRecordID)
				if err != nil {
					log.Printf(tlog.Log()+"Error getting sn maintenances from PG: %d %s ", httpError, err.Error())
					continue
				}
				// Add to maintenance returns that will get posted
				maintenanceReturns = append(maintenanceReturns, *maintenanceReturn)
			}
			readyToPost = append(readyToPost, snRecords.Results[i].SourceID)
		}
	}
	if len(maintenanceReturns) > 0 {
		for i := range maintenanceReturns {
			// TODO: This needs to be tweaked per pnp-nq2ds/issues/184
			ShouldHaveNotification := false
			// WILL NOT PROCESS NOTIFICATIONS THROUGH NOTIFICATION_CONSUMER.   We handle this separately for SN
			// if (strings.ToLower(maintenanceReturns[i].State) == "in-progress" ||
			// 	// TODO: this may not be accurate.  Will probably need to be tweaked
			// 	strings.ToLower(maintenanceReturns[i].State) == "scheduled" ||
			// 	strings.ToLower(maintenanceReturns[i].State) == "complete") &&
			// 	maintenanceReturns[i].DisruptionDuration > 0 {
			// 	ShouldHaveNotification = true
			// } else {
			// 	log.Printf(FCT+`Notification not necessary for %s due to either state is %s(not in-progress||scheduled||complete) or no disruption time(%d)`, maintenanceReturns[i].SourceID, strings.ToLower(maintenanceReturns[i].State), maintenanceReturns[i].DisruptionDuration)
			// }
			maintMap := maintReturnTomaintMap(&maintenanceReturns[i], ShouldHaveNotification)
			maintenanceMaps = append(maintenanceMaps, *maintMap)
		}
	}

	log.Printf(tlog.Log()+"Total in message: %d \n\tTotal in error: %d \n\tTotal skipped: %d \n\tTotal ready to Post:  %d \n\tTotal updated: %d \n\tTotal Inserted: %d ",
		len(snRecords.Results), errored, skipped, len(readyToPost), len(updated), len(inserted))

	txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRprocMantSN, nil, nil)
	defer func() {
		err := txn.End()
		if err != nil {
			log.Println(tlog.Log(), err)
		}
	}()
	ctx = newrelic.NewContext(ctx, txn)
	if !isBulk {
		addMonitorAttributes(ctx, maintenanceMaps)
	}
	return maintenanceMaps, notifications
}

func addMonitorAttributes(ctxParent context.Context, maintenanceMaps []datastore.MaintenanceMap) {
	// Add New Relic and Instana monitoring points
	if len(maintenanceMaps) > 0 {
		source := maintenanceMaps[0].Source
		sourceID := maintenanceMaps[0].SourceID
		sourceCreationTime, _ := parseStringToTime(maintenanceMaps[0].SourceCreationTime)
		sourceUpdateTime, _ := parseStringToTime(maintenanceMaps[0].SourceUpdateTime)
		pnpCreationTime, _ := parseStringToTime(maintenanceMaps[0].PnpCreationTime)
		pnpUpdateTime, _ := parseStringToTime(maintenanceMaps[0].PnpUpdateTime)
		wait := pnpUpdateTime - sourceUpdateTime
		ossmon.SetTagsKV(ctxParent,
			"pnp-Source", source,
			"pnp-SourceID", sourceID,
			"pnp-Kind", "maintenance",
			"sourceCreationTimeString", maintenanceMaps[0].SourceCreationTime,
			"sourceUpdateTimeString", maintenanceMaps[0].SourceUpdateTime,
			"pnpCreationTimeString", maintenanceMaps[0].PnpCreationTime,
			"pnpUpdateTimeString", maintenanceMaps[0].PnpUpdateTime,
			"sourceCreationTime", sourceCreationTime,
			"sourceUpdateTime", sourceUpdateTime,
			"pnpCreationTime", pnpCreationTime,
			"pnpUpdateTime", pnpUpdateTime,
			"pnp-WaitTime", wait,
			"apiKubeClusterRegion", monitor.NRregion,
			"apiKubeAppDeployedEnv", monitor.NRenvironment,
		)
	}
}

func DecodeAndMapMaintenceMessage(message []byte, mon *ossmon.OSSMon) (messageMap map[string]interface{}, decryptedMsg []byte) {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRdecodeAndMap, nil, nil)
	defer func() {
		err := txn.End()
		if err != nil {
			log.Println(tlog.Log(), err)
		}
	}()
	ctx = newrelic.NewContext(ctx, txn)
	defer span.Finish()
	// Record start time:
	startTime := time.Now()
	ossmon.SetTagsKV(ctx,
		ossmon.TagRegion, monitor.REGION,
		ossmon.TagEnv, monitor.ENVIRONMENT,
		ossmon.TagZone, monitor.ZONE,
		"apiKubeClusterRegion", monitor.NRregion,
		"apiKubeAppDeployedEnv", monitor.NRenvironment,
		"pnp-StartTime", startTime.Unix(),
	)

	decryptedMsg, err := encryption.Decrypt(message)
	if err != nil {
		log.Println(tlog.Log()+"Message could not be decrypted, err = ", err)
		ossmon.SetTag(ctx, DecryptionErr, true)
		ossmon.SetError(ctx, DecryptionErr+"-"+err.Error()) //Replaces NewRelic NQLR api-pnp-nq2ds_DecodeMapMaintDecryptionErr
		return nil, nil
	}
	messageMap = make(map[string]interface{})
	err = json.Unmarshal(decryptedMsg, &messageMap)
	// Capture the duration and record it in New Relic and Instana
	ossmon.SetTag(ctx, "pnp-Duration", time.Since(startTime))
	if err != nil {
		log.Print(tlog.Log()+"Error occurred unmarshaling maintenance message, err = ", err)
		log.Print(tlog.Log()+"decryptedMsg: ", string(api.RedactAttributes(decryptedMsg)))
		return nil, decryptedMsg
	}
	return messageMap, decryptedMsg
}

func getExistingMaintenance(ctxParent context.Context, database *sql.DB, recordID string) (*datastore.MaintenanceInsert, bool) {
	doesMaintenanceAlreadyExist := true
	existingMaintenance, err, httpStatusCode := db.GetMaintenanceByRecordIDStatement(database, recordID)
	if err != nil && httpStatusCode == 200 {
		doesMaintenanceAlreadyExist = false
		ossmon.SetTag(ctxParent, DBFailedErr, false)
		return nil, doesMaintenanceAlreadyExist
	} else if err != nil {
		doesMaintenanceAlreadyExist = false
		errMsg := fmt.Sprintf("Error getting maintenance from PG, err = %s, http status code: %d ", err, httpStatusCode)
		ossmon.SetTag(ctxParent, DBFailedErr, true)
		log.Printf(tlog.Log() + errMsg)
		ossmon.SetError(ctxParent, DBFailedErr+"-"+errMsg) //Replaces NewRelic NQRL monitor api-pnp-nq2ds_ProcMaintDBFailure
		return nil, doesMaintenanceAlreadyExist
	}
	existingMaintInsert := maintReturnToMaintInsert(existingMaintenance)
	return &existingMaintInsert, doesMaintenanceAlreadyExist
}

func validateMaintInsert(maintInsert *datastore.MaintenanceInsert) (errorMsg string) {
	// Only attempts to check the minimal needed fields:
	if maintInsert == nil {
		errorMsg = "Input message is nil"
	} else if maintInsert.Source == "" {
		errorMsg = "Bad message: source is empty"
	} else if maintInsert.SourceID == "" {
		errorMsg = "Bad message: source ID is empty"
	} else if len(maintInsert.CRNFull) == 0 {
		errorMsg = "Bad message full CRN is empty"
	} else if maintInsert.State == "" {
		errorMsg = "Bad message State is empty"
	} else if maintInsert.SourceCreationTime == "" {
		errorMsg = "Bad message source creation time is empty"
	}
	return
}

func maintReturnToMaintInsert(maintenanceReturn *datastore.MaintenanceReturn) (maintenanceInsert datastore.MaintenanceInsert) {
	maintenanceInsert = datastore.MaintenanceInsert{
		SourceCreationTime: maintenanceReturn.SourceCreationTime,
		SourceUpdateTime:   maintenanceReturn.SourceUpdateTime,
		PlannedStartTime:   maintenanceReturn.PlannedStartTime,
		PlannedEndTime:     maintenanceReturn.PlannedEndTime,
		ShortDescription:   maintenanceReturn.ShortDescription,
		// Will not update long desciption for change records per KenP
		// LongDescription:       maintenanceReturn.LongDescription,
		CRNFull:               maintenanceReturn.CRNFull,
		State:                 maintenanceReturn.State,
		Disruptive:            maintenanceReturn.Disruptive,
		SourceID:              maintenanceReturn.SourceID,
		Source:                maintenanceReturn.Source,
		RegulatoryDomain:      maintenanceReturn.RegulatoryDomain,
		RecordHash:            maintenanceReturn.RecordHash,
		MaintenanceDuration:   maintenanceReturn.MaintenanceDuration,
		DisruptionType:        maintenanceReturn.DisruptionType,
		DisruptionDescription: maintenanceReturn.DisruptionDescription,
		DisruptionDuration:    maintenanceReturn.DisruptionDuration,
		CompletionCode:        maintenanceReturn.CompletionCode,
		TargetedURL:           maintenanceReturn.TargetedURL,
		Audience:              maintenanceReturn.Audience,
	}
	return
}

func maintReturnTomaintMap(maintenanceReturn *datastore.MaintenanceReturn, shouldHaveNotification bool) *datastore.MaintenanceMap {
	maintMapObj := datastore.MaintenanceMap{
		RecordID:               maintenanceReturn.RecordID,
		PnpCreationTime:        maintenanceReturn.PnpCreationTime,
		PnpUpdateTime:          maintenanceReturn.PnpUpdateTime,
		SourceCreationTime:     maintenanceReturn.SourceCreationTime,
		SourceUpdateTime:       maintenanceReturn.SourceUpdateTime,
		PlannedStartTime:       maintenanceReturn.PlannedStartTime,
		PlannedEndTime:         maintenanceReturn.PlannedEndTime,
		ShortDescription:       maintenanceReturn.ShortDescription,
		LongDescription:        maintenanceReturn.LongDescription,
		CRNFull:                maintenanceReturn.CRNFull,
		State:                  maintenanceReturn.State,
		Disruptive:             maintenanceReturn.Disruptive,
		SourceID:               maintenanceReturn.SourceID,
		Source:                 maintenanceReturn.Source,
		RegulatoryDomain:       maintenanceReturn.RegulatoryDomain,
		RecordHash:             maintenanceReturn.RecordHash,
		MaintenanceDuration:    maintenanceReturn.MaintenanceDuration,
		DisruptionType:         maintenanceReturn.DisruptionType,
		DisruptionDescription:  maintenanceReturn.DisruptionDescription,
		DisruptionDuration:     maintenanceReturn.DisruptionDuration,
		CompletionCode:         maintenanceReturn.CompletionCode,
		ShouldHaveNotification: shouldHaveNotification,
	}
	return &maintMapObj
}

//
//func shouldChangeState(record datastore.MaintenanceInsert) string {
//	state := record.State
//	current := time.Now()
//	currentStr := current.UTC().Format("2006-01-02T15:04:05Z")
//	// Note: Don't change the state of unscheduled maintenance records (i.e. maintenance records in 'new' state)
//	if record.State != "new" && record.PlannedStartTime != "" && !helper.IsNewTimeAfterExistingTime(record.PlannedStartTime, time.RFC3339, currentStr, time.RFC3339) && record.State != "in-progress" {
//		state = "in-progress"
//	}
//	if record.State != "new" && record.PlannedEndTime != "" && !helper.IsNewTimeAfterExistingTime(record.PlannedEndTime, time.RFC3339, currentStr, time.RFC3339) && record.State != "complete" {
//		state = "complete"
//	}
//	if record.State == "complete" {
//		state = "complete"
//	}
//	log.Print(tlog.Log()+"shouldChangeState: ", state)
//	return state
//}

// Will return true if the resource is
// 1) a Deployment or a Main/Status Resource
// 2) The update time in the database is before the update time in the record
func shouldProcessMaintenance(pnpresource *datastore.MaintenanceInsert, hashString string) bool {
	if smm, ok := snMaintenanceMap.SNMaintenances[pnpresource.Source+pnpresource.SourceID]; ok {
		// If the time in the db is after the time in the msg, then we already have the latest
		timeIsAfter, err := helper.CompareTime(pnpresource.SourceUpdateTime, "", smm.SourceUpdateTime, "")
		if err != nil {
			log.Print(tlog.Log()+"Failed to compare time : ", err)
			return false
		}
		if timeIsAfter < 1 {
			log.Println(tlog.Log()+"Update not needed : Time in PG is after the time in the message: "+pnpresource.SourceID,
				"\nsnMap time: ", smm.SourceUpdateTime,
				"\nresource time: ", pnpresource.SourceUpdateTime)
			return false
		}
		if hashString != snMaintenanceMap.SNMaintenances[pnpresource.Source+pnpresource.SourceID].RecordHash {
			log.Printf(tlog.Log()+" MaintenanceMap hash: %s \t Hash being passed in: %s", snMaintenanceMap.SNMaintenances[pnpresource.Source+pnpresource.SourceID].RecordHash, hashString)
			log.Println(tlog.Log()+"Update needed :", pnpresource.Source, pnpresource.SourceID)
			log.Printf(tlog.Log()+"MaintenanceMap version: %+v \n Incoming version: %+v", snMaintenanceMap.SNMaintenances[pnpresource.Source+pnpresource.SourceID], pnpresource)
			return true
		}
		log.Println(tlog.Log() + "Update not needed : " + pnpresource.SourceID)
		return false
	}
	log.Println(tlog.Log() + "Update needed - MaintenanceMap for record not found: " + pnpresource.SourceID)
	return true
}

// setField sets field of v with given name to given value.
func calculateHash(maintReturn *datastore.MaintenanceReturn, maintInsert *datastore.MaintenanceInsert) *datastore.MaintenanceInsert {
	snRecord, _ := db.ComputeMaintenanceRecordHashUsingReturn(maintInsert, maintReturn)
	return snRecord
}

var ctxCount = 0
var nrmon *exmon.Monitor

// DO WE STILL NEED THIS ONE?
func buildContext() (ctx ctxt.Context, err error) {
	logID := fmt.Sprintf("nq2ds_maintenance%d", ctxCount)
	ctxCount++
	if nrmon == nil {
		initadapter.Initialize() // Note for UT you can replace this function via function pointer. See initialize.go
		nrmon, err = exmon.CreateMonitor()
		if err != nil {
			log.Printf("ERROR (%s, %s): Cannot create NR monitor. New Relic monitoring disabled!! %s", tlog.Log(), logID, err.Error())
		}
	}
	return ctxt.Context{LogID: logID, NRMon: nrmon}, nil
}

func getEntryType(ctx ctxt.Context, categoryID string) (string, error) {
	et, err := osscatalog.CategoryIDToEntryType(ctx, categoryID)
	if err != nil {
		return "", err
	}
	switch et {
	case string(ossrecord.PLATFORMCOMPONENT):
		et = "platform"
	case string(ossrecord.SERVICE):
		et = "services"
	case string(ossrecord.RUNTIME):
		et = "runtimes"
	default:
		et = strings.ToLower(et)
	}
	return et, nil
}

func getCrUpdateTime(tmpJSON []byte) (crUpdateTimeString string, err error) {
	log.Print(tlog.Log(), string(tmpJSON))
	type maintenanceTimes struct {
		SourceUpdateTime        string `json:"source_update_time,omitempty"`
		ChangeRequestUpdateTime string `json:"cr_update_time,omitempty"`
	}

	type maintenanceMsg struct {
		Info maintenanceTimes `json:"info"`
	}

	var tinfo = maintenanceMsg{}
	err = json.Unmarshal(tmpJSON, &tinfo)
	if err != nil {
		return "", err
	}

	var crUpdateTime = tinfo.Info.ChangeRequestUpdateTime
	var sourceUpdateTime = tinfo.Info.SourceUpdateTime

	isCRNewer, err := helper.CompareTime(crUpdateTime, "", sourceUpdateTime, "")
	if err != nil {
		return "", errors.New("getCrUpdateTime: error comparing change request update time: " + err.Error())
	} else if isCRNewer > 0 {
		return crUpdateTime, err
	}

	return
}
