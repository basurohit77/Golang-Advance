package monitoringinfo

import (
	"fmt"
	"regexp"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestListEDBMetricData(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListEDBMetricData() in short mode")
	}
	/* *testhelper.VeryVerbose = true /* XXX */

	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Monitoring /* | debug.Fine /* XXX */)
	}

	rest.LoadDefaultKeyFile()

	pattern := regexp.MustCompile(".*")

	countResults := 0

	err := ListEDBMetricData(pattern, func(e *EDBMetricData) {
		countResults++
		if *testhelper.VeryVerbose {
			fmt.Printf(" -> found entry %s\n", e.String())
			/*
				if e.Availability != 0 {
					fmt.Printf("      Availability=%f", e.Availability)
				}
				if e.Date != "" {
					fmt.Printf("      Date=%s", e.Date)
				}
			*/
		}
	})

	testhelper.AssertError(t, err)
	const minResults = 1000
	if countResults < minResults {
		t.Errorf("Expected at least %d results but got %d", minResults, countResults)
	}

	if *testhelper.VeryVerbose {
		fmt.Printf("Read %d entries from EDB\n", countResults)
	}
}
