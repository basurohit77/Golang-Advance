package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const catalogURL = "https://resource-catalog.bluemix.net/api/v1"

const testEntryName = "appid"

func main() {
	searchURL := fmt.Sprintf("%s?q=name:%s&include=metadata.service,metadata.rc_compatible", catalogURL, testEntryName)
	id := DoGet(searchURL, "search API", true)
	getURL := fmt.Sprintf("%s/%s?include=metadata.service,metadata.rc_compatible", catalogURL, id)
	DoGet(getURL, "direct get(id) API", false)
}

// DoGet tests a GET from the Global Catalog
func DoGet(url string, description string, indirect bool) string {
	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		panic(fmt.Errorf("Error from http.Get: %v", err))
	}
	if resp.StatusCode != 200 {
		panic(fmt.Errorf("Error in HTTP GET: StatusCode=%v   %v", resp.StatusCode, resp.Status))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Errorf("Error reading response body: %v", err))
	}
	var result interface{}
	if indirect {
		result = CatalogGet{}
	} else {
		result = CatalogResource{}
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		panic(fmt.Errorf("Error in json.Unmarshal: %v", err))
	}
	str, _ := json.MarshalIndent(result, "DEBUG ", "   ")
	fmt.Printf("result from %s=\nDEBUG %s", description, string(str))
	fmt.Println()

	var resource CatalogResource
	if indirect {
		count := int(result.(CatalogGet).ResourceCount)
		if count != 1 {
			panic(fmt.Errorf("expected count=1 got %d", count))
		}
		resource = result.(CatalogGet).Resources[0]
	} else {
		resource = result.(CatalogResource)
	}
	id := resource.ID
	fmt.Printf("ID=%s", id)
	fmt.Println()
	fmt.Println()
	return id
}

// CatalogGet represents the response from GET calls in Global Catalog (in JSON)
type CatalogGet struct {
	Count         int64             `json:"count"`
	Limit         int64             `json:"limit"`
	Offset        int64             `json:"offset"`
	ResourceCount int64             `json:"resource_count"`
	Resources     []CatalogResource `json:"resources"`
}

// CatalogResource represents the key items that we are about in a Global Catalog entry (in JSON)
type CatalogResource struct {
	Kind           string `json:"kind"`
	Name           string `json:"name"`
	Active         bool   `json:"active"`
	Disabled       bool   `json:"disabled"`
	Group          bool   `json:"group"`
	ID             string `json:"id"`
	ObjectMetaData *struct {
		RCCompatible bool `json:"rc_compatible"`
		Service      *struct {
			IAMCompatible   bool `json:"iam_compatible"`
			RCProvisionable bool `json:"rc_provisionable"`
		} `json:"service"`
	} `json:"metadata"`
}
