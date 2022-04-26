package clearinghouse

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestReadCHInputFile(t *testing.T) {
	//	debug.SetDebugFlags(debug.ClearingHouse)
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestReadCHInputFile() in short mode")
	}
	inputFile := "testdata/chinputfile-TEST.csv"

	err := ReadCHInputFile(inputFile)
	testhelper.AssertError(t, err)

	countEntries := 0
	for pid, elist := range allCHSummaryEntriesByPID {
		for _, e := range elist {
			countEntries++
			if *testhelper.VeryVerbose /* || true /* XXX */ {
				fmt.Printf("  --> got entry PID=%s - %+v\n", pid, e)
			}
		}
	}

	testhelper.AssertEqual(t, "Number of PIDs", 16, len(allCHSummaryEntriesByPID))
	testhelper.AssertEqual(t, "Number of entries", 16, countEntries)
}
