package servicenow

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestReadServiceNowImportFile(t *testing.T) {
	filename := "testdata/snimport_test.csv"
	err := ReadServiceNowImportFile(filename)
	testhelper.AssertError(t, err)

	/*
		cwd, _ := os.Getwd()
		fmt.Println("DEBUG: CWD=", cwd)
	*/

	testhelper.AssertEqual(t, "Number of SN Import records", 257, len(allImportedRecords))

	entry, ok := GetServiceNowImport("appid")
	testhelper.AssertEqual(t, `lookup(appid)`, true, ok)
	if ok {
		testhelper.AssertEqual(t, "entry(appid).Name", "appid", entry.Name)
		testhelper.AssertEqual(t, "entry(appid).CustomerFacing", true, entry.CustomerFacing)
		if *testhelper.VeryVerbose {
			fmt.Printf("Entry(appid): %#v\n", entry)
		}
	}
}
