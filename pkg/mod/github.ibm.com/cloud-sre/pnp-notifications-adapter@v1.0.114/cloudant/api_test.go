package cloudant

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
)

const (
	cloudantNotificationsFile   = "../cloudant/testdata/production_test.json"
	cloudantServiceNameMapFile  = "../cloudant/testdata/serviceNames_test.json"
	cloudantPlatformNameMapFile = "../cloudant/testdata/platformNames_test.json"
	cloudantRuntimesNameMapFile = "../cloudant/testdata/runtimesNames_test.json"
)

func TestGettingCloudantData1(t *testing.T) {

	// Exercise code before running any other tests
	GetNotificationsURL()
	GetPlatformIDsURL()
	GetRuntimesIDsURL()
	GetServicesIDsURL()
	cloudantCredentialsFromUserHome()
	GetCloudantPassword()

}

// TestMatchCategories will test that we seem to have a good match of notification category ids
func TestBadDataResults(t *testing.T) {

	fmt.Println("TestBadData")

	ts1 := getFileServer(t, "testdata/baddata_test.json")
	defer ts1.Close()

	log.Println("The error message that follows is expected")
	_, err := GetNotifications(ctxt.Context{}, ts1.URL, "someID", "somePass")
	if err == nil {
		t.Fatal("Expected error, but no error happened")
	}

	log.Println("The error message that follows is expected")
	_, err = GetNameMapping(ts1.URL, "someID", "somePass")
	if err == nil {
		t.Fatal("Expected error, but no error happened")
	}
}

// TestGet will test a basic get of data as it comes from Cloudant
func TestGet(t *testing.T) {

	fmt.Println("TestGet")
	ts := getFileServer(t, "testdata/production_test.json")
	defer ts.Close()

	os.Setenv(AccountID, "tempID")
	os.Setenv(AccountPW, "tempPW")
	data, err := pullCloudantCredentials()
	if err != nil {
		t.Fatal(err)
	}

	result, err := GetNotifications(ctxt.Context{}, ts.URL, data.ID, data.Password)
	if err != nil {
		t.Fatal(err)
	}

	expected := 1792
	if len(result.Rows) != expected {
		t.Fatal("Did not receive expected rows", expected, "got this:", len(result.Rows))
	}

	fmt.Println("Rows=", len(result.Rows))

}

func TestNilCache(t *testing.T) {
	setNotificationsCache(nil, "foo")
	getNotificationsCache("foo")
	setNameMappingCache(nil, "foo")
	getNameMappingCache("foo")

}

func TestBadCloudantRequest(t *testing.T) {
	_, err := GetFromCloudant("TestNilCache", "GET", "http://127.0.0.1:6767", "user", "password")

	if err == nil {
		t.Fatal("did not get expected failure")
	}

	ts := getErrorServer(t)
	defer ts.Close()
	_, err = GetFromCloudant("TestNilCache", ts.URL, "", "", "")

	if err == nil {
		t.Fatal("did not get expected failure")
	}
}

func TestGettingCloudantData2(t *testing.T) {

	GetNotificationsURL()
	GetPlatformIDsURL()
	GetRuntimesIDsURL()
	GetServicesIDsURL()
	cloudantCredentialsFromUserHome()
	GetCloudantPassword()

}

// TestMatchCategories will test that we seem to have a good match of notification category ids
//func TestMatchCategories(t *testing.T) {
//
//	fmt.Println("TestMatchCategories")
//	ts1 := getFileServer(t, "testdata/production_test.json")
//	defer ts1.Close()
//
//	os.Setenv(AccountID, "tempID")
//	os.Setenv(AccountPW, "tempPW")
//	data, err := pullCloudantCredentials()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	notes, err := GetNotifications(ctxt.Context{}, ts1.URL, data.ID, data.Password)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ts2 := getFileServer(t, "testdata/serviceNames_test.json")
//	defer ts2.Close()
//
//	ts3 := getFileServer(t, "testdata/runtimesNames_test.json")
//	defer ts3.Close()
//
//	ts4 := getFileServer(t, "testdata/platformNames_test.json")
//	defer ts4.Close()
//
//	nm, err := NewNameMap(data.ID, data.Password, ts2.URL, ts3.URL, ts4.URL)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// Run all notifications and be sure we find a notification category ID match
//	foundMap := make(map[string]string)
//	notFoundMap := make(map[string]string)
//	for _, i := range notes.Rows {
//		if i.Doc.SubCategory != "" && (i.Doc.Type == "SECURITY" || i.Doc.Type == "ANNOUNCEMENT") && (i.Doc.Category == "PLATFORM" || i.Doc.Category == "RUNTIMES" || i.Doc.Category == "SERVICES") {
//
//			createt, err := time.Parse("2006-01-02T15:04:05.000Z", i.Doc.Creation.Time)
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			if createt.After(time.Now().Add(time.Hour * 24 * -100)) {
//
//				if nm.MatchCategoryID(i.Doc.SubCategory) == nil {
//					notFoundMap[i.Doc.SubCategory] = fmt.Sprintf("[%s][%s][%s] %s", i.Doc.Category, i.Doc.Type, i.Doc.Creation.Time, i.Doc.Title)
//				} else {
//					foundMap[i.Doc.SubCategory] = "yes"
//				}
//			}
//		}
//	}
//
//	if len(notFoundMap) > 0 {
//		t.Error("Could not find", len(notFoundMap), "unique IDs")
//		t.Error("Found", len(foundMap), "unique IDs")
//
//		t.Error("Unfound IDs")
//		for k, v := range notFoundMap {
//			t.Error(k, " = ", v)
//		}
//	}
//
//}

// TestNameMapFind will find some expected categoryIDs
func TestNameMapFind(t *testing.T) {

	// Create a test HTTP server to return the service name mapping
	ts2 := getFileServer(t, cloudantServiceNameMapFile)
	defer ts2.Close()

	// Create a test HTTP server to return the runtime name mapping
	ts3 := getFileServer(t, cloudantRuntimesNameMapFile)
	defer ts3.Close()

	// Create a test HTTP server to return the platform name mapping
	ts4 := getFileServer(t, cloudantPlatformNameMapFile)
	defer ts4.Close()

	myMap, err := NewNameMap("someID", "somePW", ts2.URL, ts3.URL, ts4.URL)

	if err != nil {
		t.Fatal(err)
	}

	ids := []string{"cloudoe.sop.enum.paratureCategory.literal.l72"}

	for _, categoryID := range ids {
		if myMap.MatchCategoryID(categoryID) == nil {
			t.Error("Did not find categoryID:", categoryID)
		}
		//fmt.Println("DisplayName ", myMap.MatchCategoryID(categoryID).DisplayName)
		//fmt.Println("ServiceName ", myMap.MatchCategoryID(categoryID).ServiceName)
	}

}

// platformNames_test.json  production_test.json  runtimesNames_test.json  serviceNames_test.json
// getFileServer returns a server that can return a file.
func getFileServer(t *testing.T, fn string) *httptest.Server {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		data, err := ioutil.ReadFile(fn) // File in source directory with test file
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

// This always returns an error
func getErrorServer(t *testing.T) *httptest.Server {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
	}))

	return ts
}
