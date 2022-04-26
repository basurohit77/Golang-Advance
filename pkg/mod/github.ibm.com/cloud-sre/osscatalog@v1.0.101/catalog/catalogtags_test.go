package catalog

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestScanCategoryTags(t *testing.T) {
	var cat string
	r := &catalogapi.Resource{}

	r.Kind = "service"
	r.Tags = nil
	cat = ScanCategoryTags(r)
	testhelper.AssertEqual(t, "no category", "", cat)

	r.Kind = "service"
	r.Tags = []string{"mobile"}
	cat = ScanCategoryTags(r)
	testhelper.AssertEqual(t, "single category", "*Mobile", cat)

	r.Kind = "runtime"
	r.Tags = []string{"mobile", "apps-services", "whisk", "community", "ibm_beta", "openwhisk", "serverless"}
	cat = ScanCategoryTags(r)
	testhelper.AssertEqual(t, "multiple categories with dups", "*Compute>Serverless Compute, *Mobile, Cloud Foundry Apps", cat)

}
