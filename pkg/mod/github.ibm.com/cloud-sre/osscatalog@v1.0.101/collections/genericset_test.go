package collections

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

type TestType string

func TestGenericSet(t *testing.T) {
	var set = NewGenericSet(func(val interface{}) string {
		return string(val.(TestType))
	})
	var numDups int

	originalmapThreshold := _genericSetMapThreshold
	_genericSetMapThreshold = 2
	defer func() { _genericSetMapThreshold = originalmapThreshold }()

	numDups = set.Add(TestType("tag1"), TestType("tag3"))
	testhelper.AssertEqual(t, "add to empty set", []interface{}{TestType("tag1"), TestType("tag3")}, set.Slice())
	testhelper.AssertEqual(t, "add to empty set - numDups", 0, numDups)

	numDups = set.Add(TestType("tag2"), TestType("tag4"))
	testhelper.AssertEqual(t, "insert + add to non-empty set, no dups", []interface{}{TestType("tag1"), TestType("tag2"), TestType("tag3"), TestType("tag4")}, set.Slice())
	testhelper.AssertEqual(t, "insert + add to non-empty set, no dups - numDups", 0, numDups)

	actualTag2 := TestType("tag2")
	numDups = set.Add(actualTag2, TestType("tag3.5"))
	testhelper.AssertEqual(t, "insert + add dup", []interface{}{TestType("tag1"), TestType("tag2"), TestType("tag3"), TestType("tag3.5"), TestType("tag4")}, set.Slice())
	testhelper.AssertEqual(t, "insert + add dup - numDups", 1, numDups)

	testhelper.AssertEqual(t, "Len()", 5, set.Len())
	testhelper.AssertEqual(t, "Contains()", true, set.Contains(TestType("tag2")))
	testhelper.AssertEqual(t, "not Contains()", false, set.Contains(TestType("tag5")))

	testhelper.AssertEqual(t, "Find()", actualTag2, set.Find(TestType("tag2")))
	testhelper.AssertEqual(t, "not Find()", nil, set.Find(TestType("tag5")))
}
