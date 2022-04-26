package osscatalog

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
)

const (
	// RememberMissingIDs is a timer to indicate when we recheck the OSS catalog for missing catalog IDs
	RememberMissingIDs = time.Hour * 12

	desireServiceName = "desireServiceName"
	desireDisplayName = "desireDisplayName"
	desireEntryType   = "desireEntryType"
)

var notFoundList map[string]string
var lastListTime time.Time
var debugOSSCompliance = false

// CategoryIDToServiceName will take a notification Category ID and transform it to a
// CRN service name use the OSS catalog records
func CategoryIDToServiceName(ctx ctxt.Context, categoryID string) (serviceName string, err error) {

	serviceName, _, err = findByCategoryID(ctx, categoryID, desireServiceName)
	return serviceName, err
}

// ServiceNameToCategoryID will take a service name and return its notification Category ID
func ServiceNameToCategoryID(ctx ctxt.Context, servicename string) (categoryID string, err error) {
	categoryID, _, err = findByServiceName(ctx, servicename)
	return categoryID, err
}

// CategoryIDToDisplayName will take a notification Category ID and transform it to a
// display name use the OSS catalog records
func CategoryIDToDisplayName(ctx ctxt.Context, categoryID string) (displayName string, err error) {
	displayName, _, err = findByCategoryID(ctx, categoryID, desireDisplayName)
	return displayName, err
}

// CategoryIDToEntryType will take a notification Category ID and transform it to a
// entry type usng the OSS catalog records
func CategoryIDToEntryType(ctx ctxt.Context, categoryID string) (entryType string, err error) {
	entryType, _, err = findByCategoryID(ctx, categoryID, desireEntryType)
	return entryType, err
}

// CategoryIDHasTag will take a notification Category ID and determine if an
// OSS tag is associated with its entry in the OSS catalog
func CategoryIDHasTag(ctx ctxt.Context, categoryID string, tag osstags.Tag) (result bool, err error) {
	METHOD := "CategoryIDHasTag"
	_, rec, err := findByCategoryID(ctx, categoryID, desireDisplayName)
	if err != nil {
		return false, err
	}
	if rec == nil {
		return false, fmt.Errorf("INFO (%s): Record not found for category ID %s", METHOD, categoryID)
	}
	return (len(rec.GeneralInfo.OSSTags) > 0 && rec.GeneralInfo.OSSTags.Contains(tag)), err
}

// ServiceNameHasTag will take a service name and determine if an
// OSS tag is associated with its entry in the OSS catalog
func ServiceNameHasTag(ctx ctxt.Context, servicename string, tag osstags.Tag) (result bool, err error) {
	METHOD := "ServiceNameHasTag"
	_, rec, err := findByServiceName(ctx, servicename)
	if err != nil {
		return false, err
	}
	if rec == nil {
		return false, fmt.Errorf("INFO (%s): Record not found for service name %s", METHOD, servicename)
	}
	return (len(rec.GeneralInfo.OSSTags) > 0 && rec.GeneralInfo.OSSTags.Contains(tag)), err
}

// CategoryIDToOSSRecord will take a notification Category ID and return the cached oss record
func CategoryIDToOSSRecord(ctx ctxt.Context, categoryID string) (record *ossrecord.OSSService, err error) {
	METHOD := "CategoryIDToOSSRecord"
	_, rec, err := findByCategoryID(ctx, categoryID, desireDisplayName)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, fmt.Errorf("INFO (%s): Record not found for category ID %s", METHOD, categoryID)
	}
	return rec, err
}

// ServiceNameToOSSRecord will take a Service Name and return the cached oss record
func ServiceNameToOSSRecord(ctx ctxt.Context, servicename string) (record *ossrecord.OSSService, err error) {
	METHOD := "ServiceNameToOSSRecord"
	_, rec, err := findByServiceName(ctx, servicename)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, fmt.Errorf("INFO (%s): Record not found for service name %s", METHOD, servicename)
	}
	return rec, err
}

// CategoryIDToOSSCompliance will return compliance indication from the OSS catalog
// ok = can be included
// bad = do not include
// notFound = do not include
// error = do not include, an error condition occurred
func CategoryIDToOSSCompliance(ctx ctxt.Context, categoryID string) (status string) {
	result := "bad"

	// Perhaps check for OSS tag "oss_status_green"
	serviceName, record, err := findByCategoryID(ctx, categoryID, desireServiceName)

	if err != nil {
		result = "error"
		log.Println("WARNING: error looking for category id ", categoryID, ", service name ", serviceName, ": ", err)
	} else if serviceName == "" || record == nil {
		result = "notFound"
		log.Println("WARNING: Record not found for category id ", categoryID, ", service name ", serviceName)
	} else {
		if debugOSSCompliance {
			// DEBUG
			log.Println("DEBUG: >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
			log.Println("DEBUG: ServiceName:" + record.ReferenceResourceName)
			log.Println("DEBUG: Group:" + record.StatusPage.Group)
			log.Println("DEBUG: CategoryID:" + record.StatusPage.CategoryID)
			log.Println("DEBUG: CategoryParent:" + record.StatusPage.CategoryParent)
			log.Println("DEBUG: OSSTags:", record.GeneralInfo.OSSTags)
			log.Println("DEBUG: EntryType:", record.GeneralInfo.EntryType)

			log.Println("DEBUG: <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
			// DEBUG
		}
		if record.GeneralInfo.OSSTags.Contains(osstags.PnPEnabled) {
			result = "ok"
		} else {
			log.Println("WARNING: Category id ", categoryID, " is not PNP enabled, ", "service name ", serviceName)
		}
	}

	return result
}

func makeCache(ctx ctxt.Context) (cache *OSSRecordCache, err error) {
	cache, err = NewCache(ctx, globalListingFunction)
	if err != nil || cache == nil {
		log.Printf("ERROR (makeCache): Getting new cache")
		return nil, err
	}
	return cache, err
}

var lookupLock = sync.Mutex{}

func findByCategoryID(ctx ctxt.Context, categoryID, desired string) (result string, record *ossrecord.OSSService, err error) {
	METHOD := "findByCategoryID"

	lookupLock.Lock()
	defer lookupLock.Unlock()

	cache, err := GetCache(ctx)
	if cache == nil {
		log.Printf("ERROR (%s): Getting cache: cache is nil", METHOD)
		return "", nil, errors.New("cache is nil")
	}
	if err != nil {
		log.Printf("ERROR (%s): Getting cache: %s", METHOD, err.Error())
		return "", nil, err
	}

	record = cache.CategoryMap[categoryID]

	if record != nil {
		result = string(record.ReferenceResourceName)
		if desired == desireDisplayName {
			result = record.ReferenceDisplayName
		}
		if desired == desireEntryType {
			result = string(record.GeneralInfo.EntryType)
		}
	}

	// If the service name was not found, then we might need to refresh
	// the OSS records.  But only refresh on new category IDs that we
	// encounter.  We don't want to keep refreshing everytime we find the
	// same missing category ID
	if notFoundList == nil || lastListTime.Add(RememberMissingIDs).Before(time.Now()) {
		notFoundList = make(map[string]string)
		lastListTime = time.Now()
	}
	if result == "" && notFoundList[categoryID] == "" {

		cache, err = makeCache(ctx)
		if err != nil {
			return "", nil, err
		}

		record := cache.CategoryMap[categoryID]

		if record != nil {
			result = string(record.ReferenceResourceName)
			if desired == desireDisplayName {
				result = record.ReferenceDisplayName
			}
		}

	}

	if result == "" {
		notFoundList[categoryID] = time.Now().String()
	}

	return result, record, nil
}

func findByServiceName(ctx ctxt.Context, serviceName string) (result string, record *ossrecord.OSSService, err error) {
	METHOD := "findByServiceName"

	lookupLock.Lock()
	defer lookupLock.Unlock()

	cache, err := GetCache(ctx)
	if cache == nil {
		log.Printf("ERROR (%s): Getting cache: cache is nil", METHOD)
		return "", nil, errors.New("cache is nil")
	}
	if err != nil {
		log.Printf("ERROR (%s): Getting cache: %s", METHOD, err.Error())
		return "", nil, err
	}

	record = cache.ServiceMap[serviceName]

	if record != nil {
		result = string(record.StatusPage.CategoryID)
	}

	return result, record, err
}

func ServiceExistedInOSS(ctx ctxt.Context, serviceName string) ( existedInOSS bool, record *ossrecord.OSSService, err error) {
	METHOD := "ServiceExistedInOSS"

	lookupLock.Lock()
	defer lookupLock.Unlock()

	cache, err := GetCache(ctx)
	if cache == nil {
		log.Printf("ERROR (%s): Getting cache: cache is nil", METHOD)
		return false, nil, errors.New("cache is nil")
	}
	if err != nil {
		log.Printf("ERROR (%s): Getting cache: %s", METHOD, err.Error())
		return false, nil, err
	}
	record = cache.ServiceMap[serviceName]
	existedInOSS = cache.ServiceInOSSMap[serviceName]

	return existedInOSS, record, err
}
