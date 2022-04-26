package osscatalog

import (
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/exmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/utils"
)

var debugCatalog = false

// ListOSSFunction is the prototype for the function used by the osscatalog to enumerate the
// OSS catalog entries
type ListOSSFunction func(*regexp.Regexp, catalog.IncludeOptions, func(r ossrecord.OSSEntry)) error

var globalListingFunction = GetDefaultListingFunction()

// OSSRecordCache A cache of the OSS records from the osscatlog libraray
type OSSRecordCache struct {
	CreateTime      time.Time // The time this cache was created
	Records         []*ossrecord.OSSService
	CategoryMap     map[string]*ossrecord.OSSService // records keyed by notification category ID
	ServiceMap      map[string]*ossrecord.OSSService // records keyed by service name
	ServiceInOSSMap map[string]bool                  // used to check if the service exists in OSSCatalog
}

var recordCache *OSSRecordCache
var tempCache *OSSRecordCache

// GetCache retrieves the OSSRecord cache or returns nil of one does not exist
func GetCache(ctx ctxt.Context) (*OSSRecordCache, error) {
	return NewCache(ctx, globalListingFunction) // New cache checks expiration and rebuilds only if expired
}

// NewCache creates a new OSSRecord cache
func NewCache(ctx ctxt.Context, listingFunc ListOSSFunction) (*OSSRecordCache, error) {

	METHOD := "osscatalog.NewCache"

	if recordCache != nil {
		// Safety valve to ensure we don't pound the OSS record library
		if recordCache.CreateTime.Add(utils.GetEnvMinutes(OSSCatalogCacheTime, 120)).After(time.Now()) {
			return recordCache, nil
		}
	}

	log.Printf("INFO (%s,%s): Retreiving new OSS Catalog cache", METHOD, ctx.LogID)

	txnStartTime := time.Now()

	tempCache = new(OSSRecordCache)
	tempCache.Records = make([]*ossrecord.OSSService, 0, 50)
	tempCache.CategoryMap = make(map[string]*ossrecord.OSSService)
	tempCache.ServiceMap = make(map[string]*ossrecord.OSSService)
	tempCache.ServiceInOSSMap = make(map[string]bool)
	tempCache.CreateTime = time.Now()

	// Call the osscatalog library listing function
	// The function is abstracted for unit test purposes
	globalListingFunction = listingFunc
	err := listingFunc(regexp.MustCompile(".*"), catalog.IncludeServices, OSSResourceHandler)
	if err != nil {
		log.Printf("ERROR (%s,%s): error calling listing function: %s", METHOD, ctx.LogID, err.Error())
		// Note: Required to put the new relic transaction down here like this because it somehow doesn't work when we
		// use the osscatalog listing funciton. Suspect it has to do with the separate callbacks in the function.
		txn := ctx.NRMon.StartTransaction(exmon.OSSCatGetRecords)
		txn.Start = txnStartTime
		txn.AddCustomAttribute(exmon.Operation, exmon.OperationQuery)
		txn.AddBoolCustomAttribute(exmon.OSSCatalogFailed, true)
		txn.AddIntCustomAttribute(exmon.RecordCount, 0)
		txn.End()
		if recordCache != nil {
			log.Printf("WARN (%s,%s): using the older cache due to the error calling listing function", METHOD, ctx.LogID)
			return recordCache, nil
		} else {
			return nil, err
		}
	}

	// Note: Required to put the new relic transaction down here like this because it somehow doesn't work when we
	// use the osscatalog listing funciton. Suspect it has to do with the separate callbacks in the function.
	txn := ctx.NRMon.StartTransaction(exmon.OSSCatGetRecords)
	txn.Start = txnStartTime
	txn.AddCustomAttribute(exmon.Operation, exmon.OperationQuery)
	txn.AddBoolCustomAttribute(exmon.ParseFailed, false)
	txn.AddBoolCustomAttribute(exmon.OSSCatalogFailed, false)
	txn.AddIntCustomAttribute(exmon.RecordCount, len(tempCache.Records))
	txn.End()

	recordCache = tempCache
	tempCache = nil

	return recordCache, nil
}

// OSSResourceHandler is the function called by the osscatalog library that provides the OSS records
// found in the global catalog
func OSSResourceHandler(ossE ossrecord.OSSEntry) {

	var r *ossrecord.OSSService

	switch ossR := ossE.(type) {
	case *ossrecord.OSSService:
		r = ossR
	default:
		r = nil
	}

	if r != nil {
		referenceResourceName := strings.TrimSpace(string(r.ReferenceResourceName))

		tempCache.Records = append(tempCache.Records, r)
		tempCache.ServiceInOSSMap[referenceResourceName] = true

		categoryID := strings.TrimSpace(r.StatusPage.CategoryID)

		// We only deal with pnp-enabled non-retired resources that have a category ID
		// Note: Ideally a retired service should NOT be pnp-enabled. But there can be a time-window
		// between setting of 'retired' status and removal of pnp-enabled tag. So we need additional
		// RETIRED check below.
		if categoryID != "" && r.GeneralInfo.OSSTags.Contains(osstags.PnPEnabled) &&
			r.GeneralInfo.OperationalStatus != ossrecord.RETIRED {
			categoryParent := strings.TrimSpace(string(r.StatusPage.CategoryParent))
			if debugCatalog {
				log.Println("DEBUG: ================================= ", categoryID)
				log.Println("DEBUG: Name=", referenceResourceName)
				log.Println("DEBUG: ClientFacing=", r.GeneralInfo.ClientFacing)
				log.Println("DEBUG: EntryType=", r.GeneralInfo.EntryType)
				log.Println("DEBUG: OperationalStatus=", r.GeneralInfo.OperationalStatus)
				log.Println("DEBUG: OSSTags=", r.GeneralInfo.OSSTags)
				log.Println("DEBUG: CategoryParent:" + categoryParent)
			}
			// OSS Catalog code will identify a parent of a notification category ID. If multiple resources have the same
			// notification category ID, then the OSS records "CategoryParent" attribute will have the serviceName of the
			// parent.  If the current record is the parent, then the CategoryParent matches the service name of this record.
			// If no other service has this category ID, then the CategoryParent is empty.
			if categoryParent == "" || categoryParent == referenceResourceName {
				tempCache.CategoryMap[categoryID] = r
			}
			tempCache.ServiceMap[referenceResourceName] = r
		}
	}
}

// GetDefaultListingFunction will retrieve what is the default value
// for a listing function
func GetDefaultListingFunction() ListOSSFunction {

	env := os.Getenv("KUBE_APP_DEPLOYED_ENV")

	if env != "prod" {
		log.Println("Choosing default OSS Catalog listing function for staging.")
		return catalog.ListOSSEntries
	}
	log.Println("Choosing default OSS Catalog listing function for production.")
	return catalog.ListOSSEntriesProduction
}
