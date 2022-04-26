package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	newrelic "github.com/newrelic/go-agent"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	osscatalogCRN "github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/ossmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
)

// ProcessStatus - updates the status of the input resource sent by the Status consumer.
// Returns whether or not the message is a valid message.
func ProcessStatus(database *sql.DB, message []byte, mon *ossmon.OSSMon) (isBadMessage bool) {
	span, ctx := ossmon.StartParentSpan(context.Background(), *mon, monitor.SrvPrfx+tlog.FuncName())
	defer span.Finish()
	isRetryNeeded := true
	for isRetryNeeded {
		isBadMessage, isRetryNeeded = internalProcessStatus(ctx, mon,database, message)
		if isRetryNeeded {
			log.Print(tlog.Log() + "sleep and retry")
			time.Sleep(time.Second * 5)
		}
	}
	return isBadMessage
}

func internalProcessStatus(ctxParent context.Context, mon *ossmon.OSSMon ,database *sql.DB, message []byte) (isBadMessage bool, isRetryNeeded bool) {
	txn :=mon.NewRelicApp.StartTransaction(monitor.TxnNRprocStatus,nil,nil)
	ctxParent = newrelic.NewContext(ctxParent, txn)
	defer func() {
		err:= txn.End()
		if err !=nil {
			log.Println(tlog.Log(),err)
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
	decryptedMsg, err := encryption.Decrypt(message)
	if err != nil {
		log.Println(tlog.Log()+"Message could not be decrypted, err = ", err)
		ossmon.SetTag(ctxParent, DecryptionErr, err.Error())
		ossmon.SetError(ctxParent, DecryptionErr+"-"+err.Error()) //Replaces NewRelic NQRL monitor api-pnp-nq2ds_ProcStatusDecryptionErr
		return true,false
	}
	log.Print(tlog.Log()+"Payload from queue: ", string(decryptedMsg))
	// Unmarshal byte array:
	resourceReturn := &datastore.ResourceReturn{}
	err = json.Unmarshal(decryptedMsg, &resourceReturn)
	if err != nil {
		ossmon.SetTag(ctxParent, DecryptionErr, err.Error())
		ossmon.SetError(ctxParent, DecryptionErr+"-"+"Error occurred trying to unmarshal resource to ResourceReturn, err = "+err.Error())
		log.Print(tlog.Log(), "Error occurred trying to unmarshal resource to ResourceReturn, err = ", err)
		return true, false
	}
	// Validate the incoming message:
	validationErrMsg := validateResourceReturn(resourceReturn)
	if validationErrMsg != "" {
		log.Print(tlog.Log(), validationErrMsg)
		ossmon.SetTag(ctxParent, "validateResourceReturn", validationErrMsg)
		ossmon.SetError(ctxParent, "validateResourceReturn-"+validationErrMsg)
		return true,false
	}
	log.Print(tlog.Log()+"Source = "+resourceReturn.Source+" | Source id = "+resourceReturn.SourceID+" | status = ", resourceReturn.Status)
	// Create the structure that will hold information to update:
	resourceInsert := resourceReturnToResourceInsert(resourceReturn)
	log.Printf(tlog.Log()+"Resource to update: %#v", resourceInsert)
	ossmon.SetTagsKV(ctxParent,
		"pnp-Source", resourceReturn.Source,
		"pnp-SourceID", resourceReturn.SourceID,
		"pnp-Kind", "status",
		"pnp-Operation", "update",
	)
	// Try to update the resource in the database:
	err, httpResponseCode := db.UpdateResource(database, resourceInsert)
	if err != nil && httpResponseCode == http.StatusBadRequest {
		log.Print(tlog.Log(), "Bad message, err = ", err)
		ossmon.SetTag(ctxParent, DBFailedErr, true)
		ossmon.SetError(ctxParent, DBFailedErr+"-"+err.Error())
		return true, false
	} else if err != nil {
		log.Print(tlog.Log(), "Error occurred trying to update resource in database, err = ", err)
		ossmon.SetTag(ctxParent, DBFailedErr, false)
		return false,true
	}
	log.Print(tlog.Log() + "Resource updated in PG")
	// Capture end time:
	endTime := time.Now()
	// Determine duration to report to NewRelic:
	duration := endTime.Unix() - startTime.Unix()
	ossmon.SetTag(ctxParent, "pnp-Duration", duration)
	return false, false // If we have gotten this far everything is OK:
}

func validateResourceReturn(resourceReturn *datastore.ResourceReturn) (errorMsg string) {
	// Only attempts to check the minimal needed fields:
	if resourceReturn == nil {
		errorMsg = "Input message is nil"
	} else if resourceReturn.Source == "" {
		errorMsg = "Bad message: source is empty"
	} else if resourceReturn.SourceID == "" {
		errorMsg = "Bad message: source ID is empty"
	} else if resourceReturn.CRNFull == "" {
		errorMsg = "Bad message full CRN is empty"
	} else if resourceReturn.Status == "" {
		isNotPublicCloud, err := isNotPublicCloudResource(resourceReturn.CRNFull)
		if err != nil {
			errorMsg = err.Error()
		}
		if !isNotPublicCloud {
			errorMsg = "Bad message status is empty"
		}
	} else {
		expectedResourceID := db.CreateRecordIDFromSourceSourceID(resourceReturn.Source, resourceReturn.SourceID)
		if resourceReturn.RecordID != expectedResourceID {
			errorMsg = "Record id does not match the source and source id"
		}
	}
	return
}

func resourceReturnToResourceInsert(resourceReturn *datastore.ResourceReturn) (resourceInsert *datastore.ResourceInsert) {
	resourceInsert = &datastore.ResourceInsert{
		SourceCreationTime: resourceReturn.SourceCreationTime,
		SourceUpdateTime:   resourceReturn.SourceUpdateTime,
		CRNFull:            resourceReturn.CRNFull,
		State:              resourceReturn.State,
		OperationalStatus:  resourceReturn.OperationalStatus,
		Source:             resourceReturn.Source,
		SourceID:           resourceReturn.SourceID,
		Status:             resourceReturn.Status,
		StatusUpdateTime:   resourceReturn.StatusUpdateTime,
		RegulatoryDomain:   resourceReturn.RegulatoryDomain,
		CategoryID:         resourceReturn.CategoryID,
		CategoryParent:     resourceReturn.CategoryParent,
		DisplayNames:       resourceReturn.DisplayNames,
		Visibility:         resourceReturn.Visibility,
		Tags:               resourceReturn.Tags,
		IsCatalogParent:    resourceReturn.IsCatalogParent,
		CatalogParentID:    resourceReturn.CatalogParentID,
	}
	return
}

func isNotPublicCloudResource(crnFull string) (bool, error) {
	crnMask, err := osscatalogCRN.Parse(crnFull)
	if err != nil {
		return false, err
	}
	if !crnMask.IsIBMPublicCloud() {
		return true, nil
	}
	return false, nil
}
