package osscatalog

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/testutils"
)

// Performs a really basic lookup test
func TestLookup(t *testing.T) {

	debugCatalog = true // Probably will remove this debug code in the future
	PrimeOSSCatalogCache(t)

	serviceName, err := CategoryIDToServiceName(ctxt.Context{}, "cloudoe.sop.enum.paratureCategory.literal.l247")
	if err != nil {
		t.Fatal(err)
	}

	if serviceName != "accesstrail" {
		t.Fatal("Bad service name.  Should be accesstrail, but was", serviceName)
	}

	debugOSSCompliance = true // may eventually remove the debug code
	CategoryIDToOSSCompliance(ctxt.Context{}, "cloudoe.sop.enum.paratureCategory.literal.l247")
}

func PrimeOSSCatalogCache(t *testing.T) {
	cache, err := NewCache(ctxt.Context{}, testutils.MyTestListFunction)
	if err != nil || cache == nil {
		t.Fatal("Unable to prime cache", err)
	}
	// Get a second time
	cache, err = NewCache(ctxt.Context{}, testutils.MyTestListFunction)
	if err != nil || cache == nil {
		t.Fatal("Unable to get a second time", err)
	}
}

func TestDisplayName(t *testing.T) {
	sn, _ := CategoryIDToDisplayName(ctxt.Context{}, "foobar")
	if sn != "" {
		t.Fatal("found something we should not have")
	}
}

func TestErrorCache(t *testing.T) {
	log.Println("Expecting 4 error statements below")
	globalListingFunction = errorListingServer
	recordCache = nil
	tempCache = nil
	_, err := makeCache(ctxt.Context{})
	if err == nil {
		t.Fatal("Did not get expected error")
	}

	recordCache = nil
	tempCache = nil
	_, _, err = findByCategoryID(ctxt.Context{}, "foobar", "foobar")
	if err == nil {
		t.Fatal("Did not get expected error")
	}
}

func errorListingServer(r *regexp.Regexp, cio catalog.IncludeOptions, myFunc func(r ossrecord.OSSEntry)) error {
	return errors.New("The test error server is sending an error for testing")
}

// This always returns an error
func getErrorServer(t *testing.T) *httptest.Server {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
	}))

	return ts
}

func TestLookups(t *testing.T) {
	CategoryIDToOSSCompliance(ctxt.Context{}, "cloudoe.sop.enum.paratureCategory.literal.l133")
}

func TestServerNameLookup(t *testing.T) {

	recordCache = nil
	cache, err := NewCache(ctxt.Context{}, serverNameListingServer)
	if err != nil || cache == nil {
		t.Fatal("Unable to get a cache", err)
	}

	id, _, err := findByServiceName(ctxt.Context{}, "MyService1")
	if id != "MyService1_CategoryID" {
		t.Fatal("did not find id MyService1_CategoryID")
	}

	id, _, err = findByServiceName(ctxt.Context{}, "MyService2")
	if id != "MyService2_CategoryID" {
		t.Fatal("did not find id MyService2_CategoryID")
	}

	id, _, err = findByServiceName(ctxt.Context{}, "MyService3")
	if id != "MyService3_CategoryID" {
		t.Fatal("did not find id MyService3_CategoryID")
	}

	id, _, err = findByServiceName(ctxt.Context{}, "MyService4")
	if id != "MyService3_CategoryID" {
		t.Fatal("did not find id MyService3_CategoryID")
	}

	id, err = ServiceNameToCategoryID(ctxt.Context{}, "MyService3")
	if id != "MyService3_CategoryID" {
		t.Fatal("did not find id MyService3_CategoryID")
	}

}

func serverNameListingServer(r *regexp.Regexp, cio catalog.IncludeOptions, myFunc func(r ossrecord.OSSEntry)) error {

	o := new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyService1" //CRNServiceName
	o.StatusPage.CategoryID = "MyService1_CategoryID"
	o.GeneralInfo.OSSTags.AddTag(osstags.PnPEnabled)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyService2" //CRNServiceName
	o.StatusPage.CategoryID = "MyService2_CategoryID"
	o.GeneralInfo.OSSTags.AddTag(osstags.PnPEnabled)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyService3" //CRNServiceName
	o.StatusPage.CategoryID = "MyService3_CategoryID"
	o.StatusPage.CategoryParent = "MyServiceParent"
	o.GeneralInfo.OSSTags.AddTag(osstags.PnPEnabled)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyService4" //CRNServiceName
	o.StatusPage.CategoryID = "MyService3_CategoryID"
	o.StatusPage.CategoryParent = "MyServiceParent"
	o.GeneralInfo.OSSTags.AddTag(osstags.PnPEnabled)
	myFunc(o)

	return nil
}

func TestTags(t *testing.T) {

	t.Log("ENTER TestTags")
	recordCache = nil
	cache, err := NewCache(ctxt.Context{}, tagNameListingServer)
	if err != nil || cache == nil {
		t.Fatal("Unable to get a cache", err)
	}

	b, err := CategoryIDHasTag(ctxt.Context{}, "cloudoe.sop.enum.paratureCategory.literal.l247", osstags.PnPEnabled)
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("Did not get expected result")
	}

	b, err = ServiceNameHasTag(ctxt.Context{}, "MyService1", osstags.PnPEnabled)
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("Did not get expected result")
	}

	b, err = ServiceNameHasTag(ctxt.Context{}, "MyService4", osstags.OneCloud)
	if err != nil {
		t.Fatal(err)
	}
	if b {
		t.Fatal("Did not get expected result")
	}

	b, err = CategoryIDHasTag(ctxt.Context{}, "MyService4_CategoryID", osstags.PnPEnabled)
	if err == nil {
		t.Fatal("should not have found this")
	}
}

func TestRecordGet(t *testing.T) {

	t.Log("ENTER TestRecordGet")
	recordCache = nil
	cache, err := NewCache(ctxt.Context{}, tagNameListingServer)
	if err != nil || cache == nil {
		t.Fatal("Unable to get a cache", err)
	}

	rec, err := CategoryIDToOSSRecord(ctxt.Context{}, "cloudoe.sop.enum.paratureCategory.literal.l247")
	if err != nil {
		t.Fatal(err)
	}
	if rec == nil {
		t.Fatal("Did not get record")
	}

	rec, err = ServiceNameToOSSRecord(ctxt.Context{}, "MyService3")
	if err != nil {
		t.Fatal(err)
	}
	if rec == nil {
		t.Fatal("Did not get record")
	}
	if rec.ReferenceResourceName != "MyService3" {
		t.Fatal("Did not get the child record")
	}

	rec, err = CategoryIDToOSSRecord(ctxt.Context{}, "MyService3_CategoryID")
	if err == nil && rec != nil {
		t.Fatal("Should not have found this")
	}

	rec, err = CategoryIDToOSSRecord(ctxt.Context{}, "MyServiceParentCategoryID")
	if err != nil {
		t.Fatal(err)
	}
	if rec == nil {
		t.Fatal("Did not get record")
	}
	if rec.ReferenceResourceName != "MyServiceParent" {
		t.Fatal("Did not get the parent record")
	}
}

func tagNameListingServer(r *regexp.Regexp, cio catalog.IncludeOptions, myFunc func(r ossrecord.OSSEntry)) error {

	o := new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyServiceParent" //CRNServiceName
	o.StatusPage.CategoryID = "MyServiceParentCategoryID"
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.PnPEnabled)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyService1" //CRNServiceName
	o.StatusPage.CategoryID = "cloudoe.sop.enum.paratureCategory.literal.l247"
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.PnPEnabled)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyService2" //CRNServiceName
	o.StatusPage.CategoryID = "MyService2_CategoryID"
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.PnPEnabled)
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.OneCloud)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyService3" //CRNServiceName
	o.StatusPage.CategoryID = "MyServiceParentCategoryID"
	o.StatusPage.CategoryParent = "MyServiceParent"
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.PnPEnabled)
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.OneCloud)
	myFunc(o)

	o = new(ossrecord.OSSService)
	o.ReferenceResourceName = "MyService4" //CRNServiceName
	o.StatusPage.CategoryID = "MyServiceParentCategoryID"
	o.StatusPage.CategoryParent = "MyServiceParent"
	o.GeneralInfo.OSSTags = append(o.GeneralInfo.OSSTags, osstags.PnPEnabled)
	myFunc(o)

	return nil
}
