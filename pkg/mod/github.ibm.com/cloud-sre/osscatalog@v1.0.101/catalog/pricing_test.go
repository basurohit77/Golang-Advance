package catalog

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestListPricingInfoFromCatalog(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListPricingInfoFromCatalog() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Catalog | debug.Pricing)
	}

	options.LoadGlobalOptions("-keyfile DEFAULT -visibility ibm_only", true)

	var rootName ossrecord.CRNServiceName = "appid"
	root, err := ReadMainCatalogEntry(rootName)
	if err != nil {
		t.Errorf("TestListPricingInfoFromCatalog - could not get root entry (%s): %v", rootName, err)
		t.FailNow()
	}

	countResults := 0

	err = ListPricingInfoFromCatalog(root, func(p *catalogapi.Pricing) {
		countResults++
		if *testhelper.VeryVerbose /* || true /* XXX */ {
			fmt.Printf(" -> found entry: %+v\n", p)
		}
	})

	if err != nil {
		t.Errorf("ListPricingInfoFromCatalog failed: %v", err)
	}
	if countResults < 5 {
		t.Errorf("ListPricingInfoFromCatalog returned only %d entries -- fewer than expected", countResults)
	}
	if *testhelper.VeryVerbose {
		fmt.Printf("Read %d Pricing entries from Global Catalog\n", countResults)
	}

}

func TestListPricingInfo(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListPricingInfo() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Catalog | debug.Pricing /* | debug.Fine | debug.IAM */)
	}

	options.LoadGlobalOptions("-keyfile DEFAULT -visibility ibm_only", true)

	var rootName ossrecord.CRNServiceName = "appid"
	root, err := ReadMainCatalogEntry(rootName)
	if err != nil {
		t.Errorf("TestListPricingInfo - could not get root entry (%s): %v", rootName, err)
		t.FailNow()
	}

	countResults := 0

	err = ListPricingInfo(root, func(p *BSSPricing) {
		countResults++
		if *testhelper.VeryVerbose /* || true /* XXX */ {
			fmt.Printf(" -> found entry: %+v\n", p)
		}
	})

	if err != nil {
		t.Errorf("ListPricingInfo failed: %v", err)
	}
	if countResults < 2 {
		t.Errorf("ListPricingInfo returned only %d entries -- fewer than expected", countResults)
	}
	if *testhelper.VeryVerbose {
		fmt.Printf("Read %d Pricing entries from Pricing Catalog\n", countResults)
	}

}
