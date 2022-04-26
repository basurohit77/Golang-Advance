package debug

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestMakeDuplicateName(t *testing.T) {
	testhelper.AssertEqual(t, "Simple", "Simple (1)", MakeDuplicateName("Simple"))
	testhelper.AssertEqual(t, "Simple (1)", "Simple (2)", MakeDuplicateName("Simple (1)"))
	testhelper.AssertEqual(t, "Simple(2)", "Simple(2) (1)", MakeDuplicateName("Simple(2)"))
	testhelper.AssertEqual(t, "Simple (3)", "Simple (4)", MakeDuplicateName("Simple (3)"))
	testhelper.AssertEqual(t, "Simple (foo)", "Simple (foo) (1)", MakeDuplicateName("Simple (foo)"))
}

func TestCompareDuplicateNames(t *testing.T) {
	testhelper.AssertEqual(t, "empty/(1)", true, CompareDuplicateNames("Simple", "Simple (1)"))
	testhelper.AssertEqual(t, "(1)/empty", true, CompareDuplicateNames("Simple (1)", "Simple"))
	testhelper.AssertEqual(t, "(1)/(2)", true, CompareDuplicateNames("Simple (1)", "Simple (2)"))
	testhelper.AssertEqual(t, "empty/empty", true, CompareDuplicateNames("Simple", "Simple"))
	testhelper.AssertEqual(t, "(1)/diff(1)", false, CompareDuplicateNames("Simple (1)", "Other (1)"))
}
