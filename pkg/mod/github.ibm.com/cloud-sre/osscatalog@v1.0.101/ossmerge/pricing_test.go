package ossmerge

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestLoadPricingFile(t *testing.T) {

	if *testhelper.VeryVerbose /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Fine)
	}

	numEntries, err := LoadPricingFile("testdata/test-pricing.json")

	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "numEntries", 5, numEntries)
}
