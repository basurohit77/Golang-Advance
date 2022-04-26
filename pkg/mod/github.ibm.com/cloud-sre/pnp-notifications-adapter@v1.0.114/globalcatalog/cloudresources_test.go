package globalcatalog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/testutils"
)

func TestLoad(t *testing.T) {

	ts, url := testutils.GetGCResourceServer(t)
	defer ts.Close()

	list, err := GetCloudResourcesCache(ctxt.Context{}, url)

	if err != nil {
		t.Fatal(err)
	}

	if len(list.Resources) != 231 {
		t.Fatal("Received unexpected resource count ", len(list.Resources))
	}
}

func TestError(t *testing.T) {
	_, err := getCloudResources(ctxt.Context{}, "foobar://ibm.com")
	if err == nil {
		t.Fatal("Didn't get an expected error")
	}
}

// TestBadGC will test if the GC provides a bad http response, that the cache is handled correctly
func TestBadGC(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	}))
	defer ts.Close()

	cr := &CloudResource{Active: true, Name: "testResource"}
	rm := make(map[string]*CloudResource)
	rm[cr.Name] = cr
	cloudResourcesCache = &CloudResourceCache{CacheURL: "https://ibm.com", Resources: rm}
	list, _ := GetCloudResourcesCache(ctxt.Context{}, ts.URL)

	if list == nil {
		t.Fatal("Missing list. Old cache should have been retained")
	}

	if len(list.Resources) != 1 {
		t.Fatal("Wrong list length")
	}

}
