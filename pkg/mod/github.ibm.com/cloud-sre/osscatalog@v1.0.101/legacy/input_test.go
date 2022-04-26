package legacy

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestListLegacyRecords(t *testing.T) {
	err := SetCRNValidationFile("testdata/crn-names-TEST-REPORT.xlsx")
	testhelper.AssertError(t, err)

	results := make([]*Entry, 0, 10)

	err = ListLegacyRecords(nil, func(e *Entry) {
		if testing.Verbose() {
			fmt.Printf("  -> Got Entry: %+v\n", e)
		}
		results = append(results, e)
	})
	testhelper.AssertError(t, err)

	testhelper.AssertEqual(t, "number of entries", 4, len(results))
	if len(results) > 0 {
		testhelper.AssertEqual(t, "entry#0", "&{Name:CLOUD_OBJECT_STORAGE EntryType:IAAS OperationalStatus:GA Exceptions:exception-iaas ValidationStatus:Deferred - IaaS DuplicateOf: Notes: AllSources:map[CLOUD_OBJECT_STORAGE:['GHoST(computed)', 'GlobalCatalog']]}", fmt.Sprintf("%+v", results[0]))
	}
}
