package collections

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestStringSet(t *testing.T) {
	var set = NewStringSet()
	var numDups int

	originalmapThreshold := _stringSetMapThreshold
	_stringSetMapThreshold = 3
	defer func() { _stringSetMapThreshold = originalmapThreshold }()

	numDups = set.Add("tag1", "tag3")
	testhelper.AssertEqual(t, "add to empty set", []string{"tag1", "tag3"}, set.Slice())
	testhelper.AssertEqual(t, "add to empty set - numDups", 0, numDups)

	numDups = set.Add("tag2", "tag4")
	testhelper.AssertEqual(t, "insert + add to non-empty set, no dups", []string{"tag1", "tag2", "tag3", "tag4"}, set.Slice())
	testhelper.AssertEqual(t, "insert + add to non-empty set, no dups - numDups", 0, numDups)

	numDups = set.Add("tag2", "tag3.5")
	testhelper.AssertEqual(t, "insert + add dup", []string{"tag1", "tag2", "tag3", "tag3.5", "tag4"}, set.Slice())
	testhelper.AssertEqual(t, "insert + add dup - numDups", 1, numDups)

	testhelper.AssertEqual(t, "Len()", 5, set.Len())
	testhelper.AssertEqual(t, "Contains()", true, set.Contains("tag2"))
	testhelper.AssertEqual(t, "not Contains()", false, set.Contains("tag5"))
}

func TestStringSetCompare(t *testing.T) {
	originalmapThreshold := _stringSetMapThreshold
	_stringSetMapThreshold = 3
	defer func() { _stringSetMapThreshold = originalmapThreshold }()

	var left = NewStringSet("iaas", "service")
	var right = NewStringSet("service", "iaas")

	leftOnly, rightOnly := left.Compare(right)
	testhelper.AssertEqual(t, "leftOnly(1)", nil, leftOnly)
	testhelper.AssertEqual(t, "rightOnly(1)", nil, rightOnly)

	left.Add("runtime")
	right.Add("deployment")
	leftOnly, rightOnly = left.Compare(right)
	testhelper.AssertEqual(t, "leftOnly(2)", `"runtime"`, leftOnly.String())
	testhelper.AssertEqual(t, "rightOnly(2)", `"deployment"`, rightOnly.String())
}
