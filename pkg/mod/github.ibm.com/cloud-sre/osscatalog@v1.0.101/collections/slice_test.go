package collections

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestCompareSliceString(t *testing.T) {

	f := func(title string, left, right []string, expectedLeftOnly, expectedRightOnly []string) {
		leftOnly, rightOnly := CompareSliceString(left, right)
		title = fmt.Sprintf("%s.CompareSliceString(%v, %v)", title, left, right)
		testhelper.AssertEqual(t, title+".leftOnly", expectedLeftOnly, leftOnly)
		testhelper.AssertEqual(t, title+".rightOnly", expectedRightOnly, rightOnly)
	}

	f("equal", []string{"A", "B", "C", "D"}, []string{"A", "B", "C", "D"}, []string{}, []string{})
	f("left-empty", []string{}, []string{"A", "B", "C", "D"}, []string{}, []string{"A", "B", "C", "D"})
	f("right-empty", []string{"A", "B", "C", "D"}, []string{}, []string{"A", "B", "C", "D"}, []string{})
	f("both-empty", []string{}, []string{}, []string{}, []string{})
	f("left-middle", []string{"A", "B", "C", "D"}, []string{"A", "B", "D"}, []string{"C"}, []string{})
	f("right-middle", []string{"A", "B", "D"}, []string{"A", "B", "C", "D"}, []string{}, []string{"C"})
	f("left-start", []string{"A", "B", "C", "D"}, []string{"B", "C", "D"}, []string{"A"}, []string{})
	f("right-start", []string{"B", "C", "D"}, []string{"A", "B", "C", "D"}, []string{}, []string{"A"})
	f("left-end", []string{"A", "B", "C", "D"}, []string{"A", "B", "C"}, []string{"D"}, []string{})
	f("right-end", []string{"A", "B", "C"}, []string{"A", "B", "C", "D"}, []string{}, []string{"D"})
	f("complex", []string{"A1", "A", "B", "B1", "C", "D", "D1"}, []string{"A2", "A", "A3", "B", "C", "D", "D2"}, []string{"A1", "B1", "D1"}, []string{"A2", "A3", "D2"})
}

func TestAppendSliceStringNoDup(t *testing.T) {
	var slice []string
	var numDups int

	slice, numDups = AppendSliceStringNoDups(slice, "tag1", "tag3")
	testhelper.AssertEqual(t, "add to empty slice", []string{"tag1", "tag3"}, slice)
	testhelper.AssertEqual(t, "add to empty slice - numDups", 0, numDups)

	slice, numDups = AppendSliceStringNoDups(slice, "tag2", "tag4")
	testhelper.AssertEqual(t, "insert + add to non-empty slice, no dups", []string{"tag1", "tag2", "tag3", "tag4"}, slice)
	testhelper.AssertEqual(t, "insert + add to non-empty slice, no dups - numDups", 0, numDups)

	slice, numDups = AppendSliceStringNoDups(slice, "tag2", "tag3.5")
	testhelper.AssertEqual(t, "insert + add dup", []string{"tag1", "tag2", "tag3", "tag3.5", "tag4"}, slice)
	testhelper.AssertEqual(t, "insert + add dup - numDups", 1, numDups)
}
