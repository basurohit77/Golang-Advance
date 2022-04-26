package ossrecord

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestCheckSchemaVersion(t *testing.T) {
	currentSegments := currentSchemaVersion.Segments()
	lowVersion := fmt.Sprintf("%d.%d.%d", currentSegments[0], currentSegments[1], currentSegments[2]-1)
	highVersion := fmt.Sprintf("%d.%d.%d", currentSegments[0], currentSegments[1], currentSegments[2]+1)

	test := func(title string, entryVersion string, expectedError string, expectedUpdatedVersion string) {
		entry := &OSSService{ReferenceResourceName: "test-entry", SchemaVersion: entryVersion}
		err := entry.CheckSchemaVersion()
		if err != nil {
			testhelper.AssertEqual(t, title+".error", expectedError, err.Error())
		} else {
			testhelper.AssertEqual(t, title+".error", expectedError, "")
		}
		testhelper.AssertEqual(t, title+".updatedVersion", expectedUpdatedVersion, entry.SchemaVersion)
	}

	test("compatible", OSSCurrentSchema, "", OSSCurrentSchema)
	test("entryHigh", highVersion, `OSS Resource test-entry has incompatible Schema version: current library OSSSchema "1.0.14" / got "1.0.15"`, `Mismatch{record:"1.0.15"  library:"1.0.14"}`)
	test("entryLow", lowVersion, "", lowVersion)
	test("bogus", "bogus-value", `OSS Resource test-entry has invalid format Schema version (bogus-value): Malformed version: bogus-value`, "bogus-value")
	test("mismatch", `Mismatch{record:"1.2.3"  library:"1.2.4"}`, `OSS Resource test-entry has invalid format Schema version (Mismatch{record:"1.2.3"  library:"1.2.4"}): Malformed version: Mismatch{record:"1.2.3"  library:"1.2.4"}`, `Mismatch{record:"1.2.3"  library:"1.2.4"}`)
	test("empty", "", `OSS Resource test-entry has empty Schema version`, ``)
}
