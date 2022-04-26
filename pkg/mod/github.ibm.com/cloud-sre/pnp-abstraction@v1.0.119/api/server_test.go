package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/testutils"
)

var (
	region                  = "testRegion"
	environment             = "testEnv"
	port                    = "3210"
	monitoringKey           = "xxxx"
	monitoringAppName       = ""
	skipCatalogRegistration bool
	catalogCheckRegInterval = 5
	endpointHandlers        []EndpointHandler
	stacktraceBuffer        []byte
	testApi_path            = "/api/v1/test1"
	testApi_path_2          = "/api/v1/test2"
	//	kongURL                    = "http://kongurl"
	//	catalogURL                 = "http://catalogurl"
	//	catalogTimeout             = 5
	//	catalogCategoryID          = "catalogCategoryID"
	//	catalogCategoryName        = "catalogCategoryName"
	//	catalogCategoryDescription = "Test catalogCategory description"
	//	catalogClientID            = "testCategoryApi"
)

func createServerObject() *Server {
	server := NewServer(monitoringKey)
	server.Environment = environment
	server.Region = region
	//server.Port = port
	server.KongURL = kongURL
	server.CatalogURL = catalogURL
	server.CatalogTimeout = catalogTimeout
	server.CatalogCatagoryID = catalogCategoryID
	server.CatalogCatagoryName = catalogCategoryName
	server.CatalogCatagoryDescription = catalogCategoryDescription
	server.CatalogClientID = catalogClientID
	server.CatalogCheckRegInterval = catalogCheckRegInterval
	server.MonitoringAppName = monitoringAppName
	server.SkipCatalogRegistration = true
	return server
}

func createServerObjectWithMonitor() *Server {
	server := NewServerWithMonitor(monitoringAppName, monitoringKey, region, environment)
	server.KongURL = kongURL
	server.CatalogURL = catalogURL
	server.CatalogTimeout = catalogTimeout
	server.CatalogCatagoryID = catalogCategoryID
	server.CatalogCatagoryName = catalogCategoryName
	server.CatalogCatagoryDescription = catalogCategoryDescription
	server.CatalogClientID = catalogClientID
	server.CatalogCheckRegInterval = catalogCheckRegInterval
	server.SkipCatalogRegistration = true
	return server
}

func Test_newServer(t *testing.T) {
	serv := createServerObject()
	serv.SkipCatalogRegistration = true

	validMethods := make(map[string]string)
	validMethods[http.MethodGet] = http.MethodGet

	var endpointHandlers []EndpointHandler
	eph := CreateEndpointHandler(testApi_path, testHandlerFunc, validMethods)
	endpointHandlers = append(endpointHandlers, *eph)
	serv.EndpointHandlers = endpointHandlers

	go func() {
		serv.Start()
	}()

	time.Sleep(1000 * time.Millisecond)
	expectedResponse := `{"test":"ok"}`
	testutils.HttpGet(t, eph.Path, expectedResponse, http.StatusOK, eph.HandlerFunc)

	time.Sleep(1000 * time.Millisecond)
}

func Test_newServerWithMonitor(t *testing.T) {
	serv := createServerObjectWithMonitor()
	serv.SkipCatalogRegistration = true

	validMethods := make(map[string]string)
	validMethods[http.MethodGet] = http.MethodGet

	var endpointHandlers []EndpointHandler
	eph := CreateEndpointHandler(testApi_path_2, testHandlerFunc, validMethods)
	endpointHandlers = append(endpointHandlers, *eph)
	serv.EndpointHandlers = endpointHandlers

	go func() {
		serv.StartWithoutMonitor()
	}()

	time.Sleep(1000 * time.Millisecond)
	expectedResponse := `{"test":"ok"}`
	testutils.HttpGet(t, eph.Path, expectedResponse, http.StatusOK, eph.HandlerFunc)

	time.Sleep(1000 * time.Millisecond)
}

func testHandlerFunc(resp http.ResponseWriter, req *http.Request) {
	// intercept panics: print error and stacktrace
	defer HandlePanics(resp)
	responseBody := make(map[string]interface{})
	responseBody["test"] = "ok"

	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(responseBody); err != nil {
		log.Print("handle request -- error")
		resp.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Print("handle request -- ok")
		// return successful response
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		if _, err := resp.Write(buffer.Bytes()); err != nil {
			log.Println(err)
		}
	}
}
