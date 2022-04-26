package ossvalidation

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func setupOSSValidation() *OSSValidation {
	ossv := New("test-entry", "<last ProductInfoParts run timestamp>")

	ossv.AddSource("fromScorecardV1", SCORECARDV1)
	ossv.AddSource("fromCatalog", CATALOG)
	ossv.AddSource("fromCatalog", PRIOROSS)
	ossv.AddSource("fromServiceNow", SERVICENOW)
	ossv.AddSource("fromServiceNow", COMPUTED)
	//ossv.SetSourceNameCanonical("fromCatalog")

	ossv.AddIssue(WARNING, "warning 2", "warning 2 details").TagDataMismatch()
	ossv.AddIssue(DEFERRED, "deferred 1", "deferred 1 details").TagDataMissing()
	ossv.AddIssue(WARNING, "warning 1", "warning 1 details").TagCRN()
	ossv.AddIssue(SEVERE, "severe 1", "severe 1 details").TagSNOverride().TagCRN()
	ossv.AddIssue(IGNORE, "ignore 1", "ignore 1 details")
	ossv.AddIssue(INFO, "info 2", "info 2 details")
	ossv.AddIssue(INFO, "RunAction ProductInfoParts 1", "RunAction ProductInfoParts 1 details").TagRunAction(ossrunactions.ProductInfoParts)
	ossv.AddIssue(WARNING, "warning 3", "warning 3 details")
	ossv.AddIssue(MINOR, "minor 1", "minor 1 details").TagCRN()
	ossv.AddIssue(SEVERE, "severe 2", "severe 2 details")
	ossv.AddIssue(INFO, "RunAction ProductInfoPartsRefresh 1", "RunAction ProductInfoPartsRefresh 1 details").TagRunAction(ossrunactions.ProductInfoPartsRefresh)
	ossv.AddIssue(CRITICAL, "critical 1", "critical 1 details")
	ossv.AddIssue(INFO, "info 1", "info 1 details")

	ossv.RecordRunAction(ossrunactions.ProductInfoParts)

	return ossv
}

func TestCountIssues(t *testing.T) {
	ossv := setupOSSValidation()

	counts := ossv.CountIssues(nil)

	testhelper.AssertEqual(t, "CRITICAL", 1, counts[CRITICAL])
	testhelper.AssertEqual(t, "SEVERE", 2, counts[SEVERE])
	testhelper.AssertEqual(t, "WARNING", 3, counts[WARNING])
	testhelper.AssertEqual(t, "MINOR", 1, counts[MINOR])
	testhelper.AssertEqual(t, "DEFERRED", 1, counts[DEFERRED])
	testhelper.AssertEqual(t, "INFO", 4, counts[INFO])
	testhelper.AssertEqual(t, "IGNORE", 1, counts[IGNORE])
	testhelper.AssertEqual(t, "TOTAL", 8, counts["TOTAL"])
}

func TestCountIssuesWithFilter(t *testing.T) {
	ossv := setupOSSValidation()

	counts := ossv.CountIssues([]Tag{TagCRN})

	testhelper.AssertEqual(t, "CRITICAL", 0, counts[CRITICAL])
	testhelper.AssertEqual(t, "SEVERE", 1, counts[SEVERE])
	testhelper.AssertEqual(t, "WARNING", 1, counts[WARNING])
	testhelper.AssertEqual(t, "MINOR", 1, counts[MINOR])
	testhelper.AssertEqual(t, "DEFERRED", 0, counts[DEFERRED])
	testhelper.AssertEqual(t, "INFO", 0, counts[INFO])
	testhelper.AssertEqual(t, "IGNORE", 0, counts[IGNORE])
	testhelper.AssertEqual(t, "TOTAL", 3, counts["TOTAL"])
}

func TestString(t *testing.T) {
	ossv := setupOSSValidation()

	output := ossv.Details()
	if len(output) < 100 {
		t.Errorf("Output too short - %d chars", len(output))
	}

	if *testhelper.VeryVerbose {
		fmt.Println("-- Raw Results: ")
		fmt.Print(output)
	}
}

func TestSort(t *testing.T) {
	ossv := setupOSSValidation()

	ossv.Sort()

	testhelper.AssertEqual(t, "0", "critical 1", ossv.Issues[0].Title)
	testhelper.AssertEqual(t, "1", "severe 1", ossv.Issues[1].Title)
	testhelper.AssertEqual(t, "2", "severe 2", ossv.Issues[2].Title)
	testhelper.AssertEqual(t, "3", "warning 1", ossv.Issues[3].Title)
	testhelper.AssertEqual(t, "8", "RunAction ProductInfoParts 1", ossv.Issues[8].Title)
	testhelper.AssertEqual(t, "11", "info 2", ossv.Issues[11].Title)
	testhelper.AssertEqual(t, "12", "ignore 1", ossv.Issues[12].Title)

	if *testhelper.VeryVerbose /* || true /* */ {
		output := ossv.Details()
		fmt.Println("-- Sorted Results: ")
		fmt.Print(output)
	}
}

func TestNamedIssue(t *testing.T) {
	ossv := New("test-entry", "test-timestamp")

	ossv.AddNamedIssue(NoValidOwnership, "test")

	testhelper.AssertEqual(t, "title", "No valid ownership or contact information found (no Offering Manager, etc.)", ossv.Issues[0].Title)
	testhelper.AssertEqual(t, "severity", SEVERE, ossv.Issues[0].Severity)
	testhelper.AssertEqual(t, "tags", []Tag{TagCRN}, ossv.Issues[0].Tags)
}

func TestNumTrueSources(t *testing.T) {
	ossv := setupOSSValidation()

	count := ossv.NumTrueSources()
	testhelper.AssertEqual(t, "initial", 3, count)

	ossv.AddSource("catalogIgnored", CATALOGIGNORED)
	ossv.AddSource("fromCatalog", SERVICENOWRETIRED)
	ossv.AddSource("servicenow2", SERVICENOW)
	count = ossv.NumTrueSources()
	testhelper.AssertEqual(t, "+4(1 ignored)", 4, count)
}
