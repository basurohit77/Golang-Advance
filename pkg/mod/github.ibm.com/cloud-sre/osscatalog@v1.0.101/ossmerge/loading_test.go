package ossmerge

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestLoadScorecardV1SegmentTribes(t *testing.T) {
	t.Skip("Skipping test TestLoadScorecardV1SegmentTribes() - Doctor disabled")

	if testing.Short() {
		t.Skip("Skipping test TestLoadScorecardV1SegmentTribes() in short mode")
	}
	if *testhelper.TestDebugFlag {
		debug.SetDebugFlags(debug.Catalog)
	}
	options.LoadGlobalOptions("-keyfile DEFAULT -lenient", true)

	err := LoadScorecardV1SegmentTribes()

	testhelper.AssertError(t, err)
}
