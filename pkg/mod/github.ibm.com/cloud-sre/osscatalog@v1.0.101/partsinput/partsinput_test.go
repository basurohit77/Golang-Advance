package partsinput

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestReadPartsInputFile(t *testing.T) {
	inputFile := "testdata/partsinput-TEST.xlsx"
	minPartsRecordsCount = 15

	err := ReadPartsInputFile(inputFile)
	testhelper.AssertError(t, err)

	if *testhelper.VeryVerbose /* || true /* XXX */ {
		for _, e := range allPartNumbers {
			fmt.Printf("  --> got entry %+v\n", e)
		}
	}

	testhelper.AssertEqual(t, "Number of entries", 15, len(allPartNumbers))
}

func TestReadPartsInputFileLONG(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestReadPartsInputFileLONG() in short mode")
	}
	if *testhelper.TestDebugFlag {
		debug.SetDebugFlags(debug.Pricing)
	}

	inputFile := "testdata/partsinput-LONG.xlsx"

	err := ReadPartsInputFile(inputFile)
	testhelper.AssertError(t, err)

	if *testhelper.VeryVerbose /* || true /* XXX */ {
		for _, e := range allPartNumbers {
			fmt.Printf("  --> got entry %+v\n", e)
		}
	}

	testhelper.AssertEqual(t, "Number of entries", 5446, len(allPartNumbers))
}
