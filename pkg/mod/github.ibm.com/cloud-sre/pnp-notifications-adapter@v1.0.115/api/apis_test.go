package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	cattestutils "github.ibm.com/cloud-sre/pnp-abstraction/testutils"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/cloudant"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/globalcatalog"
)

const (
	cloudantNotificationsFile   = "../cloudant/testdata/production_test.json"
	cloudantServiceNameMapFile  = "../cloudant/testdata/serviceNames_test.json"
	cloudantPlatformNameMapFile = "../cloudant/testdata/platformNames_test.json"
	cloudantRuntimesNameMapFile = "../cloudant/testdata/runtimesNames_test.json"
)

// TestNotifications is used to make a simple test of notifications feed
//func TestNotifications(t *testing.T) {
//
//	EliminateBadOSSRecords = false
//
//	sourceConfig, servers := buildSourceConfig(t)
//	for _, s := range servers {
//		defer s.Close()
//	}
//	result, err := GetNotifications(ctxt.Context{}, sourceConfig)
//
//	if err != nil {
//		log.Printf("ERROR: %s\n", err.Error())
//		t.Fatal(err)
//		return
//	}
//
//	expected := 150
//	if len(result.Items) != expected {
//		t.Fatal("Mismatch of expected number of items. Should be ", expected, " was ", len(result.Items))
//	}
//
//	// Here we just perform a random check of one of the elements to see if we parsed OK
//	randomCheck := false
//	for _, i := range result.Items {
//
//		if i.ShortDescription == "Security Bulletin: A vulnerability in OpenSSL affect IBM SDK for Node.js in IBM Cloud (CVE-2018-0739)" && i.CategoryNotificationID == "cloudoe.sop.enum.paratureCategory.literal.l12" {
//			randomCheck = true
//		}
//	}
//
//	if !randomCheck {
//		t.Fatal("Did not locate an expected record")
//	}
//
//}

//  TestCategoryIDMatch is used to test the category IDs that we can match
//func TestCategoryIDMatch(t *testing.T) {
//
//	sourceConfig, servers := buildSourceConfig(t)
//	for _, s := range servers {
//		defer s.Close()
//	}
//
//	notes, err := GetNotifications(ctxt.Context{}, sourceConfig)
//
//	if err != nil {
//		log.Printf("ERROR: %s\n", err.Error())
//		t.Fatal(err)
//		return
//	}
//
//	//t.Error("Resources")
//	for _, note := range notes.Items {
//
//		if note.NotificationType == "security" {
//			t.Log("ShortDescription:", note.ShortDescription)
//			//t.Log("LongDescription:", note.LongDescription)
//			//t.Log("EventTime:", note.EventTime)
//			t.Log("CreationTime:", note.CreationTime)
//			t.Log("UpdateTime:", note.UpdateTime)
//			//t.Log("CategoryNotificationID:", note.CategoryNotificationID)
//			//t.Log("Source:", note.Source)
//			//t.Log("SourceID:", note.SourceID)
//			//t.Log("NotificationType:", note.NotificationType)
//
//			//for _, crn := range note.CRNs {
//			//	t.Log("CRN:", crn)
//			//}
//
//			//for _, dn := range note.DisplayName {
//			//	t.Log("Lang:", dn.Language, "Text:", dn.Text)
//			//}
//		}
//
//	}
//}

func TestOSSCatConfig(t *testing.T) {

	creds := new(SourceConfig)

	err := SetupOSSCatalogCredentials(creds)
	if err == nil {
		t.Fatal("Should have gotten error about no credentials")
	}

	creds.AddOSSCatalogCredential("osscat-service-ys1", "testerkey", "my-special-key")

	err = SetupOSSCatalogCredentials(creds)
	if err != nil {
		t.Fatal(err)
	}

}

//func buildSourceConfig(t *testing.T) (sourceConfig *SourceConfig, servers []*httptest.Server) {
//
//	// Create a test HTTP server to return the notification cloudant data
//	ts1 := getFileServer(t, cloudantNotificationsFile)
//	servers = append(servers, ts1)
//	os.Setenv(cloudant.NotificationsURLEnv, ts1.URL) // Sets env to be picked up by implementation: Not really needed, but just for safety
//
//	sourceConfig = new(SourceConfig)
//	sourceConfig.Cloudant.NotificationsURL = ts1.URL
//	sourceConfig.Cloudant.AccountID = "someID"
//	sourceConfig.Cloudant.Password = "somePW"
//
//	// Create a test HTTP server to return the service name mapping
//	ts2 := getFileServer(t, cloudantServiceNameMapFile)
//	sourceConfig.Cloudant.ServiceNamesURL = ts2.URL
//	servers = append(servers, ts2)
//
//	// Create a test HTTP server to return the runtime name mapping
//	ts3 := getFileServer(t, cloudantRuntimesNameMapFile)
//	sourceConfig.Cloudant.RuntimeNamesURL = ts3.URL
//	servers = append(servers, ts3)
//
//	// Create a test HTTP server to return the platform name mapping
//	ts4 := getFileServer(t, cloudantPlatformNameMapFile)
//	sourceConfig.Cloudant.PlatformNamesURL = ts4.URL
//	servers = append(servers, ts4)
//
//	// Create Global Catalog URL for resources
//	ts5, page1URL := testutils.GetGCResourceServer(t)
//	sourceConfig.GlobalCatalog.ResourcesURL = page1URL
//	servers = append(servers, ts5)
//
//	primeOSSCatalogCache(t)
//
//	// We fill these in even though we don't actually use these credentials during UT.
//	sourceConfig.AddOSSCatalogCredential("osscat-service-ys1", "testerkey", "my-special-key")
//
//	return sourceConfig, servers
//}

func primeOSSCatalogCache(t *testing.T) {
	cache, err := osscatalog.NewCache(ctxt.Context{}, cattestutils.MyTestListFunction)
	if err != nil || cache == nil {
		t.Fatal("Unable to prime cache", err)
	}
}

// getFileServer returns a server that can provide data from a file
func getFileServer(t *testing.T, fn string) *httptest.Server {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		data, err := ioutil.ReadFile(fn)
		if err != nil {
			t.Fatal("Error reading file:", err)
		}

		// Need this code to replace dates since we cut at 90 days, this test would have different results as time marches on
		now := time.Now().Add(time.Hour * 24 * 30 * -1)
		dateSub := fmt.Sprintf("%d-%02d", now.Year(), int(now.Month()))
		data = bytes.Replace(data, []byte("XXXX-XX"), []byte(dateSub), -1)
		_, err = w.Write(data)

		if err != nil {
			t.Fatal("Error reading file:", err)
		}
	}))

	return ts
}

func TestEnvs(t *testing.T) {

	GetCredentials()
	os.Setenv(cloudant.AccountID, "12345")
	GetCredentials()
	os.Setenv(cloudant.AccountPW, "myPW")
	GetCredentials()
	os.Setenv(cloudant.NotificationsURLEnv, "https://ibm.com")
	GetCredentials()
	os.Setenv(cloudant.ServicesURLEnv, "https://ibm.com")
	GetCredentials()
	os.Setenv(cloudant.RuntimesURLEnv, "https://ibm.com")
	GetCredentials()
	os.Setenv(cloudant.PlatformURLEnv, "https://ibm.com")
	GetCredentials()
	os.Setenv(globalcatalog.GCResourceURL, "https://ibm.com")
	GetCredentials()
	os.Setenv(osscatalog.OSSCatalogCredentialLookup[0], "https://ibm.com")
	GetCredentials()

}
