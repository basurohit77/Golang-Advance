package globalcatalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/exmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/utils"
)

var cloudResourcesCache *CloudResourceCache

// GetCloudResourcesCache will return the cached version of the cloud resources.  If the cache does not exist
// then it will use the input URL to populate the cache and return the resources.
func GetCloudResourcesCache(ctx ctxt.Context, url string) (cache *CloudResourceCache, err error) {

	if cloudResourcesCache == nil || cloudResourcesCache.CacheExpire.Before(time.Now()) {
		list, err := GetAllCloudResources(ctx, url)
		if err != nil {
			return cloudResourcesCache, err
		}

		setCloudResourcesCache(list, url)
	}
	return cloudResourcesCache, err
}

func setCloudResourcesCache(list []*CloudResource, url string) *CloudResourceCache {

	if list != nil {

		cache := new(CloudResourceCache)
		cache.Resources = make(map[string]*CloudResource)

		for _, i := range list {

			if cache.Resources[i.Name] != nil {
				log.Println("ERROR: The following service or component name appears twice in the global catalog:", i.Name)
			} else {
				cache.Resources[i.Name] = i
			}
		}

		cache.CacheURL = url
		cache.CacheExpire = time.Now().Add(utils.GetEnvMinutes(GCResourceCacheTime, 120))

		cloudResourcesCache = cache
	}
	return cloudResourcesCache
}

// GetAllCloudResources is similar to GetCloudResources except this method will walk
// all of the next pointers and pull all resources into a single list.
func GetAllCloudResources(ctx ctxt.Context, url string) (list []*CloudResource, err error) {

	METHOD := "GetAllCloudResources"

	list = make([]*CloudResource, 0)

	pageURL := url
	for pageURL != "" {
		page, err := getCloudResources(ctx, pageURL)

		if err != nil {
			log.Printf("ERROR (%s): retrieving global catalog url (%s)", METHOD, pageURL)
			return nil, err
		}

		for _, r := range page.Resources {
			if r.Active {
				getLanguageStrings(r)
				list = append(list, r)
			}
		}

		pageURL = page.Next

	}

	setCloudResourcesCache(list, url)

	return list, nil
}

// getCloudResources is used to pull all the cloud notification records
func getCloudResources(ctx ctxt.Context, url string) (list *CloudResourcesList, err error) {

	METHOD := "GetCloudResources"

	txn := ctx.NRMon.StartTransaction(exmon.GlobalCatalogGetRecords)
	defer txn.End()
	txn.AddCustomAttribute(exmon.Operation, exmon.OperationQuery)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		msg := fmt.Sprintf("ERROR (%s): Failed to get resources from global catalog url (%s): %s", METHOD, url, err.Error())
		log.Println(msg)
		txn.AddBoolCustomAttribute(exmon.GlobalCatalogFailed, true)
		return nil, errors.New(msg)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := fmt.Sprintf("ERROR (%s): Unexpected status code (%s) returned when retrieving resources from global catalog url (%s): %s", METHOD, resp.Status, url, resp.Body)
		log.Println(msg)
		txn.AddBoolCustomAttribute(exmon.GlobalCatalogFailed, true)
		return nil, errors.New(msg)
	}

	txn.AddBoolCustomAttribute(exmon.GlobalCatalogFailed, false)

	defer resp.Body.Close()

	// Decode the json result into a struct
	list = new(CloudResourcesList)
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {

		msg := fmt.Sprintf("ERROR (%s): Failed to decode the result from cloud notifications API: %s", METHOD, err.Error())
		log.Println(msg)
		txn.AddBoolCustomAttribute(exmon.ParseFailed, true)
		return nil, errors.New(msg)

	}

	txn.AddBoolCustomAttribute(exmon.ParseFailed, false)
	txn.AddIntCustomAttribute(exmon.RecordCount, len(list.Resources))

	return list, err

}

func getLanguageStrings(r *CloudResource) {
	METHOD := "getLanguageStrings"

	r.LanguageStrings = make([]*ResourceLang, 0)

	var f interface{}

	err := json.Unmarshal(r.OverviewUI, &f)
	if err != nil {
		log.Fatal(err)
	}

	// JSON object parses into a map with string keys
	itemsMap := f.(map[string]interface{})

	for k, v := range itemsMap {

		switch jsonObj := v.(type) {
		// we expect objects
		case interface{}:

			rl := &ResourceLang{Language: k}

			for itemKey, itemValue := range jsonObj.(map[string]interface{}) {
				switch itemKey {
				case "description":
					rl.Description = itemValue.(string)
				case "long_description":
					rl.LongDescription = itemValue.(string)
				case "display_name":
					rl.DisplayName = itemValue.(string)
				default:
				}
			}
			r.LanguageStrings = append(r.LanguageStrings, rl)

		default:
			log.Printf("INFO (%s): Did not get an object, but not a big deal. (%s)", METHOD, k)
		}
	}

}
