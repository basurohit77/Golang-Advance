package doctor

import (
	"fmt"
	"regexp"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestListEnvironments(t *testing.T) {
	t.Skip("Skipping test TestListEnvironments() - Doctor disabled")

	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListEnvironments() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Doctor /* | debug.Fine /* XXX */)
		/* *testhelper.VeryVerbose = true /* XXX */
	}

	pattern := regexp.MustCompile(".*")

	countResults := 0

	err := ListEnvironments(pattern, func(e *EnvironmentEntry) {
		countResults++
		if *testhelper.VeryVerbose {
			fmt.Println(" -> found entry", e.String())
		}
	})

	if err != nil {
		t.Errorf("TestListEnvironments failed: %v", err)
	}
	if countResults < 100 {
		t.Errorf("TestListEnvironments returned only %d entries -- fewer than expected", countResults)
	}
	if *testhelper.VeryVerbose {
		fmt.Printf("Read %d environment entries from Doctor\n", countResults)
	}
}
