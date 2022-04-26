package osscatalog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	osscatcrn "github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// environmentListingFunction is the function that returns the environments to this processor
var globalEnvironmentListingFunc = catalog.ListOSSEnvironmentsProduction

func init() {
	if os.Getenv("test") == "true" {
		globalEnvironmentListingFunc = catalog.ListOSSEnvironments
	}
}

// EnvironmentList is simply is list of OSS environments
type EnvironmentList []ossrecord.OSSEnvironment

// environmentCacheType and environmentCache define a simple cache to hold environment data retrieved from the OSS catalog
type environmentCacheType struct {
	Expiration   time.Time
	IsRefreshing bool
	List         EnvironmentList
	CRNMap       map[string]bool // Map of environment CRNs to PNPEnabled flag
}

var environmentCache *environmentCacheType
var cacheMutex = &sync.Mutex{}

// cacheTimeout is the time that the cache has to live before refreshing
var cacheTimeout = 60 * 60 * time.Second

// IsEnvPnPEnabled will take any CRN and determine if the environment part of the CRN
// is OK to be included in PNP
func IsEnvPnPEnabled(crn string) bool {
	// Removed guard rail - only check that the CRN is public:
	return isPublicCRN(crn)
}

// isPublicCRN - returns true if the provided crn is a public CRN, false otherwise
func isPublicCRN(crn string) bool {
	crnMask, err := osscatcrn.Parse(crn)
	if err != nil {
		return false
	}
	return crnMask.IsIBMPublicCloud()
}

// GetEnvironments returns a list of OSS environments that are deemed OK for PnP.
// Note that you need to initialize the catalog prior to calling this.  See init-adapter.go
func GetEnvironments() (EnvironmentList, error) {

	cache, err := getCache()

	if err != nil {
		return nil, err
	}

	result := make(EnvironmentList, len(cache.List))
	copy(result, cache.List)

	return result, nil
}

// GetCloudServiceEnvironments returns a list of environments in which Cloud
// services may be located.  As an example, GaaS environments are not returned
// because GaaS environments only contain GaaS resources.
func GetCloudServiceEnvironments() (EnvironmentList, error) {
	return GetFilteredEnvironments(
		[]ossrecord.EnvironmentType{
			ossrecord.EnvironmentIBMCloudRegion,
			ossrecord.EnvironmentIBMCloudDatacenter,
			ossrecord.EnvironmentIBMCloudPOP})
}

// GetRegionEnvironments returns a list of environments that are region
// environments only.  This does not include datacenters or GaaS environments
func GetRegionEnvironments() (EnvironmentList, error) {
	return GetFilteredEnvironments(
		[]ossrecord.EnvironmentType{
			ossrecord.EnvironmentIBMCloudRegion})
}

// GetFilteredEnvironments is just like GetEnvironments, except you can pass a type
// and the returned list will have only the requested types of environment. You provide
// one or more types in the provided array
func GetFilteredEnvironments(envType []ossrecord.EnvironmentType) (EnvironmentList, error) {
	list, err := GetEnvironments()
	if err != nil {
		return nil, err
	}

	result := make(EnvironmentList, 0)

	for _, e := range list { // Walk through list of all envs

		for _, v := range envType { // Match against input types
			if e.Type == v {
				result = append(result, e)
				break
			}

		}
	}
	return result, nil
}

// acceptableTypes are all the acceptable environment types.  We allow:
// Cloud Regions, Datacenters, POP (Point of Presence), and GaaS
var acceptableTypes = map[ossrecord.EnvironmentType]bool{
	ossrecord.EnvironmentIBMCloudRegion:     true,
	ossrecord.EnvironmentIBMCloudDatacenter: true,
	ossrecord.EnvironmentIBMCloudPOP:        true,
	ossrecord.EnvironmentIBMCloudZone:       false,
	ossrecord.EnvironmentIBMCloudDedicated:  false,
	ossrecord.EnvironmentIBMCloudLocal:      false,
	ossrecord.EnvironmentIBMCloudStaging:    false,
	ossrecord.EnvironmentTypeSpecial:        true,
	ossrecord.EnvironmentGAAS:               true}

// acceptableStatus contains all the acceptable status for an environment
var acceptableStatus = map[ossrecord.EnvironmentStatus]bool{ossrecord.EnvironmentActive: true, ossrecord.EnvironmentSelectAvailability: true}

// envIsPNPEnabled will determine if a given OSSEnvironment is OK to use in PnP
// If not, it will return a small msg indicating why not
func envIsPNPEnabled(e ossrecord.OSSEnvironment) (msg string, ok bool) {
	// Removed guard rail - only check that the CRN is public:
	crnOfEnv := string(e.EnvironmentID)
	ok = isPublicCRN(crnOfEnv)
	if !ok {
		msg = fmt.Sprintf("Environment CRN mask is not public. Env CRN mask=%s",crnOfEnv)
	}
	return msg, ok
}

// getCache will retrieve the cache, and kickoff a refresh if needed.  Note that
// we do lazy refreshes
func getCache() (cache *environmentCacheType, err error) {
	FCT := "environments.getCache()"

	cacheMutex.Lock()         // lock access to the cache pointer
	defer cacheMutex.Unlock() // defer unlock

	if environmentCache == nil {
		log.Printf("%s INFO: Building environments cache because none exists.", FCT)
		cache, err = buildCache() // if the cache is not available, then build it on the current thread
		if err != nil {
			return nil, fmt.Errorf("FAILURE: %s Could not build environment cache. Error= %s", FCT, err.Error())
		}
		environmentCache = cache
	}

	if err == nil && environmentCache.Expiration.Before(time.Now()) {
		go refreshCache()
	}

	return environmentCache, err

}

func refreshCache() {
	FCT := "enviroments.refreshCache()"

	if environmentCache.IsRefreshing {
		return
	}
	environmentCache.IsRefreshing = true

	cache, err := buildCache()

	if err != nil {
		log.Printf("FAILURE: %s Could not build environment cache. Error= %s", FCT, err.Error())
		return // just return and don't change current cache
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	environmentCache = cache
}

var environmentList []*ossrecord.OSSEnvironment

// buildCache will build and return a new cache pointer
// the cache mutex does not need to be locked while this
// function runs
func buildCache() (*environmentCacheType, error) {

	FCT := "environments.buildCache()"

	log.Println(FCT, "Start building environment cache")
	start := time.Now()

	err := loadEnvironments()
	if err != nil {
		log.Printf("FAILURE: %s could not load environments", FCT)
		return nil, err
	}

	if len(environmentList) < 10 {
		log.Printf("FAILURE: %s something failed with the envrionment load. Only recieved %d", FCT, len(environmentList))
		return nil, fmt.Errorf("FAILURE: %s did not load environments", FCT)
	}

	cache := new(environmentCacheType)
	cache.Expiration = time.Now().Add(cacheTimeout)
	cache.CRNMap = make(map[string]bool)

	for _, e := range environmentList {

		msg, ok := envIsPNPEnabled(*e)
		if ok {
			log.Printf("%s: ACCEPT environment %s.", FCT, e.EnvironmentID)

			cache.List = append(cache.List, *e)
			cache.CRNMap[string(e.EnvironmentID)] = true

		} else {
			log.Printf("%s: REJECT environment %s. reason = %s", FCT, e.EnvironmentID, msg)
		}
	}

	log.Println(time.Since(start).String(), " Finished building environment cache")

	return cache, nil
}

func loadEnvironments() (err error) {

	pattern := regexp.MustCompile(".*")

	// Use the globalEnvironmentListingFunc instead of calling catalog directly because we can UT this
	err = globalEnvironmentListingFunc(pattern, catalog.IncludeEnvironments, environmentProcessor)
	if err != nil {
		log.Printf("loadEnvironments failed: %v", err)
		return err
	}
	return nil
}

func deepEnvCopy(in *ossrecord.OSSEnvironment) (*ossrecord.OSSEnvironment, error) {
	FCT := "deepEnvCopy"
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(in); err != nil {
		log.Println(FCT, "FAILURE: cannot encode", err)
		return nil, err
	}

	out := new(ossrecord.OSSEnvironment)
	if err := json.NewDecoder(buffer).Decode(&out); err != nil {
		log.Println(FCT, "FAILURE: cannot decode", err)
		return nil, err
	}

	return out, nil
}

func environmentProcessor(e ossrecord.OSSEntry) {
	FCT := "environmentProcessor"

	/* OSSEntry.JSON() comes back similar to this. Notice the oss_environment to start, so I don't
	   trust this to be consistent all the time.
	{
	  "oss_environment": {
	      "schema_version": "1.0.9",
	      "id": "crn:v1:yf-dallas:public::us-south::::",
	      "oss_tags": [
	          "ibmcloud_default_segment"
	      ],
	      "parent_id": "",
	      "display_name": "YF DALLAS",
	      "type": "\u003cunknown\u003e",
	      "status": "ACTIVE",
	      "reference_catalog_id": "",
	      "owning_segment": "58eda55b9babda00075a50da",
	      "description": "YF DALLAS\n\n    DoctorEnvironment(YF DALLAS/YF_DALLAS[crn:v1:yf-dallas:public::us-south::::]\n",
	      "ims_id": "",
	      "mccp_id": ""
	  }
	}
	*/

	// We want to copy the incoming object to a new copy of the environment in case OSS Library passes us a pointer

	var ecopy *ossrecord.OSSEnvironment

	switch r1 := e.(type) {
	case *ossrecord.OSSEnvironment:
		var err error
		ecopy, err = deepEnvCopy(r1)

		if err != nil {
			log.Println(FCT, "FAILURE: Cannot copy environment:", e.String())
		}
	default:
		log.Printf("* Unexpected entry type: %#v\n", e)
	}

	environmentList = append(environmentList, ecopy)

}
