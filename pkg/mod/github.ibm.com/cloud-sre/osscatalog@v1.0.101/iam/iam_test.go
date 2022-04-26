package iam

import (
	"fmt"
	"regexp"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestListIAMServices(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListIAMServices() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.IAM /* | debug.Fine /* XXX */)
		/* *testhelper.VeryVerbose = true /* XXX */
	}

	err := rest.LoadDefaultKeyFile()
	if err != nil {
		t.Errorf("TestListIAMServices.LoadDefaultKeyFile: %v", err)
	}

	pattern := regexp.MustCompile(".*")

	countResults := 0

	err = ListIAMServices(pattern, func(e *Service) {
		countResults++
		if *testhelper.VeryVerbose {
			fmt.Printf(" -> found entry: %+v\n", e)
		}
	})

	if err != nil {
		t.Errorf("TestListIAMServices failed: %v", err)
	}
	if countResults < 100 {
		t.Errorf("TestListIAMServices returned only %d entries -- fewer than expected", countResults)
	}
	if *testhelper.VeryVerbose {
		fmt.Printf("Read %d entries from IAM\n", countResults)
	}

}
