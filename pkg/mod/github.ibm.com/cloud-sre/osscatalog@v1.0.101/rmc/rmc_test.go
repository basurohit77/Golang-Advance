package rmc

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestReadRMCSummaryEntry(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestReadRMCSummaryEntry() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.RMC | debug.Fine /* XXX */)
	}

	name := "appid"
	//name := "AppID"
	//name := "cloudantnosqldb"
	//name := "pnp-api-oss"
	//name := "is.vpc"
	//name := "is"
	//name := "databases-for-redis"
	//name := "does-not-exist-678298"
	//name := "apiconnect"
	//name := "APIConnect"
	//name := "rmc-1p-asc-test1004"

	rest.LoadDefaultKeyFile()

	result, err := ReadRMCSummaryEntry(ossrecord.CRNServiceName(name), false)

	if err != nil {
		t.Errorf("ReadRMCSummaryEntry failed: %v", err)
	} else if string(result.CRNServiceName) != name {
		t.Errorf(`ReadRMCSummaryEntry did not return a record with the expected name "%s" (actual="%s")`, name, result.CRNServiceName)
	}
}

func DISABLEDTestReadRMCSummaryEntryTestMode(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestReadRMCSummaryEntryTestMode() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.RMC | debug.Fine /* XXX */)
	}

	name := "shane-oss-demo"

	rest.LoadDefaultKeyFile()

	result, err := ReadRMCSummaryEntry(ossrecord.CRNServiceName(name), true)

	if err != nil {
		t.Errorf("ReadRMCSummaryEntry failed: %v", err)
	}
	if string(result.CRNServiceName) != name {
		t.Errorf(`ReadRMCSummaryEntry did not return a record with the expected name "%s" (actual="%s")`, name, result.CRNServiceName)
	}
}

func DISABLEDTestListRMCEntriesTestMode(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListRMCEntriesTestMode() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.RMC /* | debug.Fine /* XXX */)
	}

	rest.LoadDefaultKeyFile()

	err := ListRMCEntries(true)

	testhelper.AssertError(t, err)
}
