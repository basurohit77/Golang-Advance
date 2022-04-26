package compare

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestOutputGetDiffs(t *testing.T) {
	out := Output{IncludeEqual: true}
	struct1 := struct {
		KeyValue string
		KeyEqual string
		KeyLeft  string
		KeyType  string
	}{
		KeyValue: "value1",
		KeyEqual: "valueEqual",
		KeyLeft:  "valueLeft",
		KeyType:  "valueKind",
	}
	struct2 := struct {
		KeyValue string
		KeyEqual string
		KeyRight string
		KeyType  int
	}{
		KeyValue: "value2",
		KeyEqual: "valueEqual",
		KeyRight: "valueRight",
		KeyType:  123,
	}

	DeepCompare("left", struct1, "right", struct2, &out)
	diffs := out.GetDiffs()

	expectedDiffs := []OneDiff{
		{DiffValue, "left.KeyValue", "\"value1\"", "right.KeyValue", "\"value2\""},
		{DiffEqual, "left.KeyEqual", "\"valueEqual\"", "right.KeyEqual", "\"valueEqual\""},
		{DiffLOnly, "left.KeyLeft", "\"valueLeft\"", "left.KeyLeft", ""},
		{DiffType, "left.KeyType", "Type(string)=string", "right.KeyType", "Type(int)=int"},
		{DiffROnly, "right.KeyRight", "", "right.KeyRight", "\"valueRight\""},
	}

	for i, d := range diffs {
		if i < len(expectedDiffs) {
			ed := expectedDiffs[i]
			if *d != ed {
				t.Errorf("diff[%d]: got %#v  expected %#v", i, d, ed)
			}
		} else {
			t.Errorf("diff[%d]: got %#v  - unexpected", i, d)
		}
	}
	if len(expectedDiffs) > len(diffs) {
		for i := len(diffs); i < len(expectedDiffs); i++ {
			t.Errorf("diff[%d]: got nothing  - expected %#v", i, expectedDiffs[i])
		}
	}

	//	fmt.Println(out.StringWithPrefix("DEBUGTEST: "))
}

func TestOutputSummary(t *testing.T) {
	out := Output{IncludeEqual: true}
	struct1 := struct {
		KeyValue      string
		KeyEqual      string
		KeyLeft       string
		KeyType       string
		SchemaVersion string
	}{
		KeyValue:      "value1",
		KeyEqual:      "valueEqual",
		KeyLeft:       "valueLeft",
		KeyType:       "valueKind",
		SchemaVersion: "1.1",
	}
	struct2 := struct {
		KeyValue      string
		KeyEqual      string
		KeyRight      string
		KeyType       int
		SchemaVersion string
	}{
		KeyValue:      "value2",
		KeyEqual:      "valueEqual",
		KeyRight:      "valueRight",
		KeyType:       123,
		SchemaVersion: "1.2",
	}

	DeepCompare("left", struct1, "right", struct2, &out)

	summary := out.Summary()
	testhelper.AssertEqual(t, "", "LOnly:1  ROnly:1  Diff:3  (core):4  SchemaVersion:1", summary)
}
