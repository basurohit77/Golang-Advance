package catalog

import (
	"fmt"
	"regexp"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestReadMainCatalogEntry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test TestReadMainCatalogEntry() in short mode")
	}
	if *testhelper.TestDebugFlag {
		debug.SetDebugFlags(debug.Catalog)
	}

	options.LoadGlobalOptions("-keyfile DEFAULT", true)

	result, err := ReadMainCatalogEntry("appid")

	if err != nil {
		t.Errorf("ReadMainCatalogEntry failed: %v", err)
	}
	if result.Name != "appid" {
		t.Errorf("ReadMainCatalogEntry did not return a record with the expected name \"appid\"")
	}
	testhelper.AssertEqual(t, "Visibility", string(catalogapi.VisibilityPublic), result.Visibility.Restrictions)
}

func TestListMainCatalogEntries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test TestListMainCatalogEntries() in short mode")
	}
	if *testhelper.TestDebugFlag {
		debug.SetDebugFlags(debug.Catalog)
	}

	options.LoadGlobalOptions("-keyfile DEFAULT -visibility ibm_only", true)

	//	pattern := regexp.MustCompile(".*node.*")
	pattern := regexp.MustCompile(".*")

	countResults := 0

	err := ListMainCatalogEntries(pattern, func(r *catalogapi.Resource) {
		countResults++
		if *testhelper.VeryVerbose {
			fmt.Println(" -> found entry", r.Kind, r.Name)
		}
		if r.Visibility.Restrictions == "" {
			t.Errorf("ListMainCatalogEntries() did not return a Visibility.Restrictions field for entry \"%s\"", r.Name)
		}
		if r.EffectiveVisibility.Restrictions == "" {
			t.Errorf("ListMainCatalogEntries() did not return a EffectiveVisibility.Restrictions field for entry \"%s\"", r.Name)
		}
		// TODO: Check that we also get ibm_only and private entries
	})

	if err != nil {
		t.Errorf("ListMainCatalogEntries failed: %v", err)
	}
	if countResults < 100 {
		t.Errorf("ListMainCatalogEntries returned only %d entries -- fewer than expected", countResults)
	}
	if *testhelper.VeryVerbose {
		fmt.Printf("Read %d entries from Global Catalog\n", countResults)
	}

}

func TestListDataCenterEntries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test TestListDataCenterEntries() in short mode")
	}
	if *testhelper.TestDebugFlag {
		debug.SetDebugFlags(debug.Catalog)
	}

	options.LoadGlobalOptions("-keyfile DEFAULT -visibility ibm_only", true)

	err := ListDataCenterEntries(func(r *catalogapi.Resource) {

		if *testhelper.VeryVerbose {
			fmt.Println(" -> found entry", r.Kind, r.Name)
		}
		if r.Visibility.Restrictions == "" {
			t.Errorf("ListDataCenterEntries() did not return a Visibility.Restrictions field for entry \"%s\"", r.Name)
		}

	})

	if err != nil {
		t.Errorf("ListDataCenterEntries failed: %v", err)
	}

}
func TestGetMainCatalogUI(t *testing.T) {
	url, err := GetMainCatalogEntryUI("appid")
	if err != nil {
		t.Errorf("GetMainCatalogEntryUI failed: %v", err)
	}
	//	testhelper.AssertEqual(t, "MainCatalogUI", `https://resource-catalog.bluemix.net/update/appid`, url)
	testhelper.AssertEqual(t, "MainCatalogUI", `https://globalcatalog.cloud.ibm.com/update/appid`, url)
}
