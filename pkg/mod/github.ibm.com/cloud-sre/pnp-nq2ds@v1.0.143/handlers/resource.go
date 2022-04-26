package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	newrelic "github.com/newrelic/go-agent"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/ossmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
	"github.ibm.com/cloud-sre/pnp-nq2ds/helper"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
)

// NOTE: Once NewRelic get completed decommissioned we need to remove all instances
// ossmon.SetTag(ctx, DBFailedErr, bool), ossmon.SetTag(ctx, DecryptionErr, bool), ossmon.SetTag(ctx, EncryptionErr, bool)
// SetError calls replaces the NQRL monitors

const (
	RESOURCE_STATUS_ARCHIVED = "archived"
	DecryptionErr            =  monitor.SrvPrfx+"decryption-err"
	EncryptionErr            =  monitor.SrvPrfx+"encryption-err"
	DBFailedErr              =  monitor.SrvPrfx+"pnp-db-failed"
	CRNErr                   =  monitor.SrvPrfx+"CRN-normalize-err"
)

type ResourceMap struct {
	Resources      map[string]*datastore.ResourceReturn
	lastUpdateTime time.Time
}

var resourceMap = ResourceMap{Resources: make(map[string]*datastore.ResourceReturn), lastUpdateTime: time.Now()}

// ConvertToResourceInsertArray converts the main resource and all the associated deployment records into an array of records ready to be inserted
func ConvertToResourceInsertArray(pnpRec *datastore.PnpStatusResource) (resInserts []datastore.ResourceInsert) {
	counter := 0
	log.Println(tlog.Log(), "Number of deployments: ", len(pnpRec.Deployments))
	for _, v := range pnpRec.Deployments {
		resInsert := db.ConvertPnpDeploymentToResourceInsert(v, pnpRec.DisplayName)
		// Need to check this here since it will need to get set later anyway in insertUpdateResourceInDB()
		if _, ok := resourceMap.Resources[pnpRec.SourceID]; ok && pnpRec.Status == "" {
			pnpRec.Status = resourceMap.Resources[pnpRec.SourceID].Status
		}
		if v.Status != RESOURCE_STATUS_ARCHIVED {
			hashString := db.ComputeResourceRecordHash(resInsert)
			currentCount := fmt.Sprint(counter, "/", len(pnpRec.Deployments))
			log.Printf(tlog.Log() + "counter->" + currentCount + ": ************" + v.SourceID +
				"\n\t: " + "Hash from msg = " + hashString)
			if shouldProcess(v, hashString) {
				resInsert.RecordHash = hashString
				resInserts = append(resInserts, *resInsert)
				counter++
			}
		} else {
			resInserts = append(resInserts, *resInsert)
		}
	}
	return resInserts
}

// Will return true if the resource is
// 1) a Deployment or a Main/Status Resource
// 2) The update time in the database is before the update time in the record
func shouldProcess(pnpresource interface{}, hashString string) bool {
	var pnpSourceID string
	v := reflect.TypeOf(pnpresource).String()
	if strings.Contains(v, "PnpDeployment") {
		pnpSourceID = pnpresource.(*datastore.PnpDeployment).SourceID
	} else {
		log.Println(tlog.Log() + "The record is not a deployment.  Skipping as we only process deployments. ")
		return false
	}
	if _, ok := resourceMap.Resources[pnpSourceID]; ok {
		if hashString != resourceMap.Resources[pnpSourceID].RecordHash {
			log.Printf(tlog.Log()+"Update needed : "+pnpSourceID+"\n\tcached: %+v", resourceMap.Resources[pnpSourceID])
			return true
		}
		log.Println(tlog.Log() + "Update not needed : " + pnpSourceID)
		return false
	}
	log.Println(tlog.Log() + "Update needed : " + pnpSourceID)
	return true
}

// ProcessResourceMsg processes new or updated resource
// https://github.ibm.com/cloud-sre/pnp-abstraction/blob/master/datastore/Resource.go
//
// Check and update cache of all resources.
// Check the resources in the message  (ConvertToResourceInsertArray)
// 	- must be in cache
// 	- if the hash value matches,  add to the list of resources that need to be processed
// Insert or update the resources in the list (insertUpdateResourceInDB)
// 	- Check whether there is an existing incident or maintenance against the resource for the status
// 	- Calculate a new hash
// 	- insert or update to the db
func ProcessResourceMsg(database *sql.DB, message []byte, mon *ossmon.OSSMon) ([]*datastore.ResourceReturn, error) {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRprocResource,nil,nil)
    ctx = newrelic.NewContext(ctx,txn)
	defer span.Finish()
	defer func() {
		err:= txn.End()
		if err !=nil {
			log.Println(tlog.Log(),err)
		}
	}()
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
	var updateResources []*datastore.ResourceReturn
	//Make sure we have the latest resource list.    If it's 2 minutes old, just get a new one
	if len(resourceMap.Resources) == 0 || time.Now().After(resourceMap.lastUpdateTime.Add(2*time.Minute)) {
		resArray, err := db.GetAllResources(database)
		if err != nil {
			log.Print(tlog.Log()+"Error getting resources from PG: ", err.Error())
			return updateResources, err
		}
		log.Print(tlog.Log()+"Total resources from PG: ", len(*resArray))
		tempResourceMap := ResourceMap{Resources: make(map[string]*datastore.ResourceReturn), lastUpdateTime: time.Now()}
		for i, v := range *resArray {
			tempResourceMap.Resources[v.SourceID] = &(*resArray)[i]
		}
		resourceMap = tempResourceMap
		log.Println(tlog.Log()+"resourceMap length:", len(resourceMap.Resources))
	}
	decryptedMsg, err := encryption.Decrypt(message)
	if err != nil {
		log.Println(tlog.Log()+"Message could not be decrypted, err = ", err)
		ossmon.SetTag(ctx, DecryptionErr, err.Error())
		ossmon.SetError(ctx, DecryptionErr+"-"+err.Error()) //Replaces NewRelic api-pnp-nq2ds_ProcessResDecryptionErr monitor
		return updateResources, err
	}
	log.Println("*********** Message: ************\n Consumed new or updated resource from queue: " + string(decryptedMsg))
	pnpResource := new(datastore.PnpStatusResource)
	err = json.Unmarshal(decryptedMsg, &pnpResource)
	if err != nil {
		log.Println(tlog.Log()+"Error in unmarshal :", err.Error())
		return updateResources, err
	}
	// returns an array of deployments that need processing
	resources := ConvertToResourceInsertArray(pnpResource)
	// Capture the duration time
	ossmon.SetTag(ctx, "pnp-Duration", time.Since(startTime))
	updateResources, err = insertUpdateResourceInDB(ctx, mon,database, resources)
	return updateResources, err
}

func insertUpdateResourceInDB(ctxParent context.Context, mon *ossmon.OSSMon,database *sql.DB, resources []datastore.ResourceInsert) ([]*datastore.ResourceReturn, error) {
	var updateResources []*datastore.ResourceReturn
	var err error
	counter := 0
	log.Println(tlog.Log()+"Number of resources:", len(resources))

	if len(resources) == 0 {
		txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRinsertUpdateResource,nil,nil)
		ctxParent = newrelic.NewContext(ctxParent,txn)
		defer func() {
			err:= txn.End()
			if err !=nil {
				log.Println(tlog.Log(),err)
			}
		}()
		// Record start time:
		startTime := time.Now()
		log.Print(tlog.Log() + "No Resources need to be processed")
		ossmon.SetTagsKV(ctxParent,
			ossmon.TagRegion, monitor.REGION,
			ossmon.TagEnv, monitor.ENVIRONMENT,
			ossmon.TagZone, monitor.ZONE,
			"apiKubeClusterRegion", monitor.NRregion,
			"apiKubeAppDeployedEnv", monitor.NRenvironment,
			"pnp-StartTime", startTime.Unix(),
			"pnp-Source", "",
			"pnp-SourceID", "",
			"pnp-Operation", "",
			"pnp-Kind", "resource",
			DBFailedErr, false,
			"pnp-Duration", time.Since(startTime),
		)
		return updateResources, err
	}

	for i := range resources {
		resource := resources[i]
		log.Print(tlog.Log()+"counter->", counter, "/", len(resources), " - ", resource.CRNFull)
		counter = counter + 1
		txn := mon.NewRelicApp.StartTransaction(monitor.TxnNRinsertUpdateResource,nil,nil)
		ctxParent = newrelic.NewContext(ctxParent,txn)
		defer func() {
			err:= txn.End()
			if err !=nil {
				log.Println(tlog.Log(),err)
			}
		}()
		// Record start time:
		startTime := time.Now()
		// add attributes to New Relic and Instana transaction
		ossmon.SetTagsKV(ctxParent,
			ossmon.TagRegion, monitor.REGION,
			ossmon.TagEnv, monitor.ENVIRONMENT,
			ossmon.TagZone, monitor.ZONE,
			"apiKubeClusterRegion", monitor.NRregion,
			"apiKubeAppDeployedEnv", monitor.NRenvironment,
			"pnp-StartTime", startTime.Unix(),
			"pnp-Source", resource.Source,
			"pnp-SourceID", resource.SourceID,
			"pnp-Kind", "resource",
		)
		hasStatus := false
		for _, v := range resource.Visibility {
			if v == "hasStatus" {
				hasStatus = true
				break
			}
		}
		if resource.Status == "" && hasStatus {
			resource.Status = checkResourceStatus(database, resource.CRNFull)
		}
		log.Println(tlog.Log()+"checked resource status:", resource.Status)
		if _, ok := resourceMap.Resources[resource.SourceID]; ok {
			ossmon.SetTag(ctxParent, "pnp-Operation", "update")
			err, _ = db.UpdateResource(database, &resource)
			// Keep track of inserted and updated resources:
			if err == nil {
				ossmon.SetTag(ctxParent, DBFailedErr, false)
				resourceReturn, err, _ := db.GetResourceBySourceID(database, resource.Source, resource.SourceID)
				if err == nil && resourceReturn != nil {
					updateResources = append(updateResources, resourceReturn)
				}
			} else {
				ossmon.SetError(ctxParent, DBFailedErr+"-"+err.Error()) // replaces NewRelic monitor api-pnp-nq2ds_InsertUpdateResDBFailure
				ossmon.SetTag(ctxParent, DBFailedErr, true)
			}
		} else {
			existingResource, doesResourceAlreadyExist := getExistingResource(ctxParent, database, resource.Source, resource.SourceID)
			log.Println(tlog.Log()+"existing resource:", helper.GetJson(existingResource))
			if doesResourceAlreadyExist {
				ossmon.SetTag(ctxParent, "pnp-Operation", "update")
				err, _ = db.UpdateResource(database, &resource)
				// Keep track of inserted and updated resources:
				if err == nil {
					ossmon.SetTag(ctxParent, DBFailedErr, false)
					resourceReturn, err, _ := db.GetResourceBySourceID(database, resource.Source, resource.SourceID)
					if err == nil && resourceReturn != nil {
						updateResources = append(updateResources, resourceReturn)
					}
				} else {
					ossmon.SetTag(ctxParent, DBFailedErr, true)
					ossmon.SetError(ctxParent, DBFailedErr+"-"+err.Error()) // replaces NewRelic monitor api-pnp-nq2ds_InsertUpdateResDBFailure
				}
			} else {
				ossmon.SetTag(ctxParent, "pnp-Operation", "insert")
				recordID, err, _ := db.InsertResource(database, &resources[i]) //Fix G601 (CWE-118): Implicit memory aliasing in for loop: recordID, err, _ := db.InsertResource(database, &resource)
				if err != nil {
					log.Printf(tlog.Log()+"\n %+v \n", &resources[i]) //Fix G601 (CWE-118): Implicit memory aliasing in for loop: log.Printf(tlog.Log()+"\n %+v \n", &resource)
					log.Println(tlog.Log()+"Error in insert :", err.Error())
					ossmon.SetTag(ctxParent, DBFailedErr, true)
					ossmon.SetError(ctxParent, DBFailedErr+"-"+err.Error()) // replaces NewRelic monitor api-pnp-nq2ds_InsertUpdateResDBFailure
					return updateResources, err
				}
				log.Println(tlog.Log()+"Record Id is ", recordID)
				ossmon.SetTag(ctxParent, DBFailedErr, false)
				// Keep track of inserted and updated resources:
				resourceReturn, err, _ := db.GetResourceBySourceID(database, resource.Source, resource.SourceID)
				if err == nil && resourceReturn != nil {
					updateResources = append(updateResources, resourceReturn)
					resourceMap.Resources[resource.SourceID] = resourceReturn
				}
			}
		}
		// Determine duration to report to NewRelic and Instana
		ossmon.SetTag(ctxParent, "pnp-Duration", time.Since(startTime))
	}
	return updateResources, err
}

func getExistingResource(ctxParent context.Context, database *sql.DB, source string, sourceID string) (*datastore.ResourceReturn, bool) {
	doesResourceAlreadyExist := true
	existingResource, err, httpStatusCode := db.GetResourceBySourceID(database, source, sourceID)
	if err != nil && httpStatusCode == 200 {
		doesResourceAlreadyExist = false
		ossmon.SetTag(ctxParent, DBFailedErr, false)
		return nil, doesResourceAlreadyExist
	} else if err != nil {
		errMsg := fmt.Sprintf("Error getting resource from PG, err = %s, http status code: %d ", err, httpStatusCode)
		doesResourceAlreadyExist = false
		ossmon.SetTag(ctxParent, DBFailedErr, true)
		ossmon.SetError(ctxParent, DBFailedErr+"-"+errMsg) // replaces NewRelic monitor pnp-nq2ds-getExistingResource
		log.Printf(tlog.Log() + errMsg)
		return nil, doesResourceAlreadyExist
	}
	return existingResource, doesResourceAlreadyExist
}

func checkResourceStatus(database *sql.DB, crn string) string {
	query := "crn=" + crn
	var status string
	isHighSevCIE := false
	isDisruptive := false
	if crn != "" {
		//check if there is a sev 1 CIE for this service
		rQuery, _, err, httpStatus := db.GetIncidentByQuery(database, query, -1, 0)
		if err != nil && httpStatus == 200 {
			log.Println(tlog.Log(), err.Error())
			log.Print(tlog.Log()+"check incident:", httpStatus)
			isHighSevCIE = false
		}
		if len(*rQuery) > 0 {
			for _, inc := range *rQuery {
				if (inc.Severity == "1" || inc.Severity == "2") && inc.Classification == "confirmed-cie" && inc.State != "" && inc.State != "resolved" {
					isHighSevCIE = true
					break
				}
			}
		}
		//check if there is a disruptive maintenance for this service
		qResult, _, errM, httpStatus := db.GetMaintenanceByQuery(database, query, -1, 0)
		if errM != nil && httpStatus == 200 {
			log.Println(tlog.Log(), errM.Error())
			log.Print(tlog.Log()+"check maintenance:", httpStatus)
			isDisruptive = false

		}
		if len(*qResult) > 0 {
			for _, m := range *qResult {
				currentStr := time.Now().UTC().Format("2006-01-02T15:04:05Z")
				if m.Disruptive && !helper.IsNewTimeAfterExistingTime(m.PlannedStartTime, time.RFC3339, currentStr, time.RFC3339) && helper.IsNewTimeAfterExistingTime(m.PlannedEndTime, time.RFC3339, currentStr, time.RFC3339) {
					isDisruptive = true
					break
				}
			}
		}
	}
	log.Print(tlog.Log()+"isHighSevCIE=", isHighSevCIE)
	log.Print(tlog.Log()+"isDisruptive=", isDisruptive)
	if isHighSevCIE || isDisruptive {
		status = "failed"
	} else {
		status = "ok"
	}
	return status
}
