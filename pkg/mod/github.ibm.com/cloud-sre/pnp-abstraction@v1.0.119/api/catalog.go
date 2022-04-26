package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.ibm.com/cloud-sre/tip-data-model/catalogdatamodel"
)

const catalogAPIPathImpl = "/catalog/api/catalog/impl"
const catalogAPIPathImpls = "/catalog/api/catalog/impls"
const catalogCategoryPath = "/catalog/api/catalog/category"
const catalogCategoriesPath = "/catalog/api/catalog/categories"

var catalogAuthToken = os.Getenv("CATALOG_AUTH_TOKEN")

var debugMode = false

// RegisterAPIWithCatalog registers the provided API (without its extension) with the API Catalog
func RegisterAPIWithCatalog(kongURL string, catalogURL string, catalogTimeout int, catalogCategoryID string, catalogCategoryName string, catalogCategoryDescription string, catalogClientID string) (success bool) {
	const FCT = "RegisterAPI: "
	var rimplInfo catalogdatamodel.RoleImplInfo

	// Check if category is registed with the catalog:
	if debugMode {
		log.Println(FCT+"DEBUG: About to create category ("+catalogCategoryID+") with catalog at", catalogURL, "...")
	}
	category, stat, err := getCategory(catalogURL, catalogTimeout, catalogCategoryID, false)
	if category == nil {
		// Category is not register, register it:
		if !registerCategory(catalogURL, catalogTimeout, catalogCategoryID, catalogCategoryName, catalogCategoryDescription, false) {
			log.Println(FCT + "Category could not be registered")
			return false
		}
	} else if debugMode {
		log.Println(FCT + "DEBUG: Category (" + catalogCategoryID + ") already registered")
	}

	if debugMode {
		log.Println(FCT+"DEBUG: About to register API (URL:", kongURL, ") with catalog at", catalogURL, "...")
	}

	implInfo, err := createRoleImplInfo(kongURL, catalogClientID, catalogCategoryID)
	if err != nil {
		log.Println(err)
	}

	// Get implementation list from catalog (use clientid & category as search criteria)
	implsList, stat, err := getCatalogImpls(kongURL, catalogURL, catalogTimeout, catalogClientID, catalogCategoryID, !debugMode)

	if implsList != nil {
		impls := implsList.Impls
		if debugMode {
			log.Printf(FCT+"Found %d implementations", len(impls))
		}
		if len(impls) > 0 {
			// implementation already registered, update registration
			id := impls[0].Id
			if err := json.Unmarshal(implInfo.Bytes(), &rimplInfo); err != nil {
				fmt.Println("Unmarshal run into an ERROR!")
			}
			if rimplInfo.ClientId == impls[0].ClientId &&
				rimplInfo.Categories[0] == impls[0].Categories[0] &&
				reflect.DeepEqual(rimplInfo.SourceInfo, impls[0].SourceInfo) {
				if debugMode {
					log.Println(FCT + "API has been registered, No need to update!")
				}
				return true
			}
			if debugMode {
				log.Print("DEBUG: id = ", id)
			}
			if ok := registerAPIUpdate(kongURL, catalogURL, catalogTimeout, id, catalogCategoryID); ok {
				return true
			}
			log.Println(FCT+"Error: Update API failed, id=", id)
			return false
		}
		// no implementation registered, register it
		ok := registerAPINew(kongURL, catalogURL, catalogTimeout, catalogClientID, catalogCategoryID)
		if ok {
			return true
		}
		log.Println(FCT + "Error: Register API failed")
		return false

	}
	log.Println(FCT+"Error: failed to read implementations from catalog err,stat:", err, stat)
	return false
}

func getCategory(catalogURL string, catalogTimeout int, categoryID string, hidelog bool) (category *catalogdatamodel.Category, status int, err error) {
	const FCT = "getCategory: "

	client := &http.Client{
		Timeout: time.Duration(catalogTimeout) * time.Second,
	}

	req, err := http.NewRequest("GET", catalogURL, nil)
	if err != nil {
		log.Println(err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Set("Connection", "close")

	req.URL.Path = catalogCategoryPath + "/" + categoryID
	q := req.URL.Query()
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		if hidelog == false {
			log.Println(FCT+"Error: failed to query api catalog, err= ", err)
		}
		return nil, -1, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		// Decode json response body
		if err := json.NewDecoder(res.Body).Decode(&category); err != nil {
			if hidelog == false {
				log.Println(FCT+"Failed to decode response body, err= ", err.Error())
			}
			return nil, res.StatusCode, err
		}

		return category, res.StatusCode, nil
	}

	err = errors.New(FCT + "Error: failed to query api catalog category, status=" + res.Status)
	if hidelog == false {
		log.Println(err.Error(), res.StatusCode)
	}

	return nil, res.StatusCode, err
}

func registerCategory(catalogURL string, catalogTimeout int, categoryID string, categoryName string, categoryDescription string, hidelog bool) (success bool) {
	const FCT = "registerCategory: "

	// Create API category info for registration
	categoryInfo := &catalogdatamodel.CategoryInfo{ID: categoryID, Name: categoryName, Description: categoryDescription}
	PrettyPrintJSON(FCT+"DEBUG: CategoryInfo", categoryInfo)
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(categoryInfo); err != nil {
		log.Println(FCT+" Error: cannot encode CategoryInfo, err=", err) // programming error
		return false
	}

	client := &http.Client{
		Timeout: time.Duration(catalogTimeout) * time.Second,
	}
	req, err := http.NewRequest("POST", catalogURL+catalogCategoriesPath, buffer)
	if err != nil {
		log.Println(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")
	req.Header.Set("Authorization", catalogAuthToken)

	log.Println(FCT + "DEBUG: About to register category with catalog: POST " + req.URL.String())

	res, err := client.Do(req)
	if err != nil {
		if hidelog == false {
			log.Println(FCT+"Error: failed to register api category, err= ", err)
		}
		return false
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		log.Println(FCT + "DEBUG: Category successfully added, id: " + categoryInfo.ID)
		return true
	}
	// http status other than 200
	log.Println(FCT+"Error: Category creation request failed, status ", res.StatusCode)
	return false

}

//
// getCatalogImpls - get registered implementations for the api
// Parameters:
//  hidelog (set to true to omit all log messages, we want that in the self monitor case)
// Return values:
//  impls   pointer to list of implementations registered with catalog for IM
//          category and clientID of this component.
//          nil if request to catalog failed
//  err     error
//  status  status code, httpstatus if > 0
//
func getCatalogImpls(kongURL string, catalogURL string, catalogTimeout int, catalogClientID string, catalogCategory string, hidelog bool) (impls *catalogdatamodel.RoleImplList, status int, err error) {
	const FCT = "GetCatalogImpls: "

	client := &http.Client{
		Timeout: time.Duration(catalogTimeout) * time.Second,
	}
	req, err := http.NewRequest("GET", catalogURL, nil)
	if err != nil {
		log.Println(err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Set("Connection", "close")
	req.URL.Path = catalogAPIPathImpls
	q := req.URL.Query()
	q.Add("clientId", catalogClientID)
	q.Add("category", catalogCategory)
	req.URL.RawQuery = q.Encode()

	if hidelog == false {
		log.Println(FCT + "GET " + req.URL.String())
	}

	var riList *catalogdatamodel.RoleImplList

	res, err := client.Do(req)
	if err != nil {
		if hidelog == false {
			log.Println(FCT+"Error: failed to query api catalog, err= ", err)
		}
		return nil, -1, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		// Decode json response body
		if err := json.NewDecoder(res.Body).Decode(&riList); err != nil {
			if hidelog == false {
				log.Println(FCT+"Failed to decode response body, err= ", err.Error())
			}
			return nil, res.StatusCode, err
		}
		// Decoding done, return implementation list
		if hidelog == false {
			PrettyPrintJSON(FCT+"ImplList in response body ", riList)
		}
		return riList, res.StatusCode, nil
	}

	err = errors.New(FCT + "Error: failed to query api catalog impls, status=" + res.Status)
	if hidelog == false {
		log.Println(err.Error(), res.StatusCode)
	}

	return nil, res.StatusCode, err

}

//
// RegisterApiNew - Register api implementation with API catalog. Implementation
//   was not yet registered.
// Return Value: true if successful, false otherwise
//
func registerAPINew(kongURL string, catalogURL string, catalogTimeout int, clientid string, category string) (success bool) {
	const FCT = "RegisterApiNew: "
	log.Println(FCT + "Debug: Register API Impl with catalog. id= " + clientid + "; category= " + category)

	// Create API implementation info for registration
	buffer, err := createRoleImplInfo(kongURL, clientid, category)
	if err != nil {
		log.Println(FCT+"Error: failed to create API implementation info, err=", err)
		return false
	}

	// Prepare http request to catalog API
	client := &http.Client{
		Timeout: time.Duration(catalogTimeout) * time.Second,
	}
	req, err := http.NewRequest("POST", catalogURL+catalogAPIPathImpls, buffer)
	if err != nil {
		log.Println(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")
	req.Header.Set("Authorization", catalogAuthToken)

	log.Println(FCT + "DEBUG: About to register API with catalog: POST " + req.URL.String())

	res, err := client.Do(req)
	if err != nil {
		log.Println(FCT+"Error: failed to register implementation with api catalog, err=", err)
		return false
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		// Some sanity checks ....
		var ri *catalogdatamodel.RoleImpl
		if err := json.NewDecoder(res.Body).Decode(&ri); err != nil {
			log.Println(FCT+"Error: cannot decode response body, err=", err)
			return false
		}
		if ri.ClientId != clientid {
			log.Println(FCT + "Error: returned clientId:" + ri.ClientId + " does not match request id:" + clientid)
			return false
		}
		log.Println(FCT+"DEBUG: API successfully registered, id:", ri.Id)
		return true
	} else {
		// http status other than 200
		log.Println(FCT+"Error: register request failed, status ", res.StatusCode)
		return false
	}
}

//
// RegisterApiUpdate - update a previously registered API implementation
// in the API catalog
// Return Value: true if successful, false otherwise
//
func registerAPIUpdate(kongURL string, catalogURL string, catalogTimeout int, id string, category string) (success bool) {
	const FCT = "RegisterApiUpdate: "
	log.Println(FCT + "DEBUG: Update API Impl in catalog, id= " + id + ", category= " + category)

	// Create API implementation info for catalog update
	buffer, err := createRoleImplInfo(kongURL, id, category)
	if err != nil {
		log.Println(FCT+"Error: failed to create API implementation info, err=", err)
		return false
	}

	client := &http.Client{
		Timeout: time.Duration(catalogTimeout) * time.Second,
	}
	req, err := http.NewRequest("PUT", catalogURL+catalogAPIPathImpl+"/"+id, buffer)
	if err != nil {
		log.Println(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")
	req.Header.Set("Authorization", catalogAuthToken)

	log.Println(FCT + "http request: PUT " + req.URL.String())

	res, err := client.Do(req)
	if err != nil {
		log.Println(FCT+"Error: update api impl in catalog failed, err=", err)
		return false
	}
	defer res.Body.Close() // close body only if err was nil

	if res.StatusCode == http.StatusOK {
		log.Println(FCT + "DEBUG: APIs successfully updated, id: " + id)
		return true
	} else {
		// http status other than 200
		log.Println(FCT+"Error: update request failed, status ", res.StatusCode)
		return false
	}

}

// Create implementation info for registering the API with the catalog
func createRoleImplInfo(kongURL string, clientid string, category string) (*bytes.Buffer, error) {
	const FCT = "createRoleImplInfo: "
	if debugMode {
		log.Println(FCT + "DEBUG: createRoleImplInfo for Client ID: " + clientid + "; category: " + category)
	}

	// API category
	categories := []string{category}
	// Href for API api/info
	// Format is 	"https://api.opscenter.bluemix.net/CLIENT_ID/api/info"

	sourceinfo := &catalogdatamodel.Source{Href: kongURL + "/" + clientid + "/api/info"}

	// Role implementation
	r := &catalogdatamodel.RoleImplInfo{ClientId: clientid, Categories: categories, SourceInfo: sourceinfo}
	if debugMode {
		PrettyPrintJSON(FCT+"RoleImplInfo", r)
	}
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(r); err != nil {
		log.Println(FCT+" Error: cannot encode RoleImplInfo, err=", err) // programming error
		return nil, err
	}
	return buffer, nil
}
