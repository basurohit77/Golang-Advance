package catalog

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestCRUDOSSEnvironment(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestCRUDOSSEnvironment() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Catalog | debug.Fine /* XXX */)
	}

	var err error

	rest.LoadDefaultKeyFile()
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		t.Logf("Cannot initialize Context: %v", err)
		t.FailNow()
	}

	envName := "OSSCatalog Test Environment Name"
	envID := ossrecord.EnvironmentID("crn:v1:osscatalog:staging::osscatalog-test-environment::::")
	env := ossrecordextended.NewOSSEnvironmentExtended(envID)
	env.DisplayName = envName
	envEntryID := ossrecord.CatalogID(env.GetOSSEntryID())

	//sleepTime := 30 * time.Second
	sleepTime := time.Duration(1)

	testhelper.AssertEqual(t, "envEntryID", ossrecord.CatalogID("oss_environment."+envID), envEntryID)

	// dt := debugT{t}
	dt := t

	dt.Run("delete-initial-nonnative", func(t *testing.T) {
		err = DeleteOSSEntry(env)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("create-nonnative", func(t *testing.T) {
		err = CreateOSSEntry(env, IncludeAll|IncludeEnvironments)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	defer DeleteOSSEntryByIDWithContext(ctx, envEntryID)
	dt.Run("read-after-create-nonnative", func(t *testing.T) {
		env1, err := ReadOSSEnvironment(envID)
		testhelper.AssertError(t, err)
		if env1 != nil {
			testhelper.AssertEqual(t, "envName", envName, env1.DisplayName)
		}
		time.Sleep(sleepTime)
	})
	dt.Run("update-nonnative", func(t *testing.T) {
		env.DisplayName = envName + " UPDATE1"
		err = UpdateOSSEntry(env, IncludeAll|IncludeEnvironments)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("read-after-update-nonnative", func(t *testing.T) {
		env1, err := ReadOSSEnvironment(envID)
		testhelper.AssertError(t, err)
		if env1 != nil {
			testhelper.AssertEqual(t, "envName", envName+" UPDATE1", env1.DisplayName)
		}
		time.Sleep(sleepTime)
	})
	dt.Run("delete-final-nonnative", func(t *testing.T) {
		err = DeleteOSSEntry(env)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
		//		fmt.Println("Sleeping 30 second after delete")
		time.Sleep(1 * time.Second)
	})
	dt.Run("read-after-delete-nonative", func(t *testing.T) {
		_, err = ReadOSSEnvironment(envID)
		if err != nil {
			if !rest.IsEntryNotFound(err) {
				t.Logf("%s operation failed: %v", t.Name(), err)
				t.FailNow()
			}
		} else {
			t.Logf("%s operation unexpectedly succeeded", t.Name())
			t.FailNow()
		}
		time.Sleep(sleepTime)
	})
}

func TestListOSSEnvironmentsWithContext(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test ListOSSEnvironmentsWithContext() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Catalog /* | debug.Fine /* XXX */)
	}
	/* *testhelper.VeryVerbose = true /* XXX */

	options.LoadGlobalOptions("-keyfile DEFAULT -lenient", true)
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	//ctx, err := setupContextForMainEntries(productionFlagDisabled, false)
	testhelper.AssertError(t, err)

	//	pattern := regexp.MustCompile(".*node.*")
	pattern := regexp.MustCompile("^.*")

	countResults := 0

	err = ListOSSEnvironmentsWithContext(ctx, pattern, IncludeNone, func(r ossrecord.OSSEntry) {
		countResults++
		switch r1 := r.(type) {
		case *ossrecord.OSSEnvironment:
			if *testhelper.VeryVerbose {
				fmt.Print(" -> found entry ")
				fmt.Print(r1.Header())
			}
		default:
			t.Errorf("* Unexpected entry type: %#v\n", r)
		}
	})

	if err != nil {
		t.Errorf("ListOSSEnvironmentsWithContext failed: %v", err)
	}
	if countResults < 40 {
		t.Errorf("ListOSSEnvironmentsWithContext returned only %d entries -- fewer than expected", countResults)
	}
	if *testhelper.VeryVerbose {
		fmt.Printf("Read %d OSSEnvironment entries from Global Catalog\n", countResults)
	}

}
