package doctor

import (
	"fmt"
	"regexp"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestListRegionIDs(t *testing.T) {
	t.Skip("Skipping test TestListRegionIDs() - Doctor disabled")

	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListRegionIDs() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Doctor /* | debug.Fine /* XXX */)
		*testhelper.VeryVerbose = true /* XXX */
	}

	rest.LoadDefaultKeyFile()

	pattern := regexp.MustCompile(".*")

	countResults := 0

	err := ListRegionIDs(pattern, func(e *RegionID) {
		countResults++
		if *testhelper.VeryVerbose {
			fmt.Println(" -> found entry", e.String())
		}
	})

	if err != nil {
		t.Errorf("TestListRegionIDs failed: %v", err)
	}
	if countResults < 100 {
		t.Errorf("TestListRegionIDs returned only %d entries -- fewer than expected", countResults)
	}
	if *testhelper.VeryVerbose {
		fmt.Printf("Read %d regionID entries from Doctor\n", countResults)
	}
}
