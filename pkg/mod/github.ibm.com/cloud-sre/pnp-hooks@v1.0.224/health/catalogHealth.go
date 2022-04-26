package health

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.ibm.com/cloud-sre/tip-data-model/incidentdatamodel"
)

// CatalogHealth data
type CatalogHealth struct {
	isHealthy         bool
	healthDescription string
	url               string
}

// NewCatalogHealth - constructor for API catalog health checker
func NewCatalogHealth(url string) *CatalogHealth {
	catalogHealth := &CatalogHealth{
		isHealthy:         true,
		healthDescription: "",
		url:               url,
	}
	return catalogHealth
}

// CheckCatalogHealth - Get latest API Catalog health check
func (catalogHealth CatalogHealth) CheckCatalogHealth() (bool, string) {
	resultCode, err := catalogHealth.internalCheckCatalogHealth(catalogHealth.url)
	catalogHealth.isHealthy = err == nil && resultCode == http.StatusOK
	catalogHealth.healthDescription = strconv.Itoa(resultCode)
	return catalogHealth.isHealthy, catalogHealth.healthDescription
}

func (catalogHealth CatalogHealth) internalCheckCatalogHealth(url string) (int, error) {
	const FCT = "internalCheckCatalogHealth: "

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("checkAPICatalogHealth: %v\n", err)
		return http.StatusInternalServerError, err
	}

	c := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Printf(FCT+"%v", err)
		return http.StatusInternalServerError, err
	}

	// If status code is not 200 then return 503 Service Unavailable
	if resp.StatusCode != http.StatusOK {
		return http.StatusServiceUnavailable, fmt.Errorf(FCT+"expecting 200 OK, got %d", resp.StatusCode)
	}
	// If body can not be read then return 503 Service Unavailable
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf(FCT+"%v", err)
		return http.StatusInternalServerError, err
	}

	var hr incidentdatamodel.HealthzInfo
	err = json.Unmarshal(b, &hr)
	if err != nil {
		log.Printf(FCT+"%v", err)
		return http.StatusInternalServerError, err
	}
	if hr.Code != 0 {
		log.Printf(FCT+"%v", err)
		return http.StatusServiceUnavailable, fmt.Errorf(FCT + "expected code=0")
	}

	return resp.StatusCode, nil
}
