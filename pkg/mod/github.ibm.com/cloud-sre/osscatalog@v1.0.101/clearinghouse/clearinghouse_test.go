package clearinghouse

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestParseCHLabel(t *testing.T) {
	chname, chid := ParseCHLabel(`{Some Entry Name [chid:12345GH56]}`)

	testhelper.AssertEqual(t, "chname", "Some Entry Name", chname)
	testhelper.AssertEqual(t, "chid", DeliverableID("12345GH56"), chid)
}

func TestGetFullRecordByID(t *testing.T) {
	if testing.Short() /* &&  false /* XXX */ {
		t.Skip("Skipping test TestGetFullRecordByID() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ClearingHouse)
		*testhelper.VeryVerbose = true
	}

	options.LoadGlobalOptions("-keyfile DEFAULT", true)

	resetFullRecordsCache()

	result, err := GetFullRecordByID("2052E430379B11E58B2CB2A838CE4F20") // cloudantnosqldb
	//result, err := GetFullRecordByID("4440E450C2C811E6A98AAE81A233E762") // containers-kubernetes
	if err != nil {
		t.Errorf("GetFullRecordByID failed: %v", err)
	}

	if *testhelper.VeryVerbose {
		data, err := json.MarshalIndent(result, "  ", "    ")
		if err != nil {
			fmt.Printf(" -> *ERROR* %v\n", err)
		} else {
			fmt.Printf(" -> %s\n", data)
		}
	}

}

func TestSearchRecordsByName(t *testing.T) {
	if testing.Short() /* && false /* */ {
		t.Skip("Skipping test TestSearchRecordsByName() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* */ {
		debug.SetDebugFlags(debug.ClearingHouse)
		*testhelper.VeryVerbose = true
	}

	options.LoadGlobalOptions("-keyfile DEFAULT", true)

	resetFullRecordsCache()

	result, err := SearchRecordsByName("Cloudant")
	if err != nil {
		t.Errorf("SearchRecordsByName failed: %v", err)
	}

	if *testhelper.VeryVerbose {
		if len(result) > 0 {
			for i, e := range result {
				data, err := json.MarshalIndent(e, "  ", "    ")
				if err != nil {
					fmt.Printf(" -> Entry %d *ERROR* %v\n", i, err)
				} else {
					fmt.Printf(" -> Entry %d %s\n", i, data)
				}
			}
		}
	}

}
