package ossvalidation

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestRunActionToTag(t *testing.T) {
	testhelper.AssertEqual(t, "", Tag("ProductInfo-Parts"), RunActionToTag(ossrunactions.ProductInfoParts))
	testhelper.AssertEqual(t, "", Tag("ProductInfo-Parts"), RunActionToTag(ossrunactions.ProductInfoPartsRefresh))
}

func TestCopyRunActions(t *testing.T) {
	//	options.LoadGlobalOptions("-keyfile <none>", true)
	ossv := setupOSSValidation()

	ossv2 := New(ossv.CanonicalName, "test-timestamp")
	ossv2.CopyRunAction(ossv, ossrunactions.ProductInfoParts)

	ossv2.Sort()

	testhelper.AssertEqual(t, "number of issues", 2, len(ossv2.Issues))
	testhelper.AssertEqual(t, "0", "RunAction ProductInfoParts 1", ossv2.Issues[0].Title)
	testhelper.AssertEqual(t, "1", "RunAction ProductInfoPartsRefresh 1", ossv2.Issues[1].Title)

	if *testhelper.VeryVerbose /* ||  true /* */ {
		output := ossv2.Details()
		fmt.Println("-- Sorted Results: ")
		fmt.Print(output)
	}
}
