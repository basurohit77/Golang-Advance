package testhelper

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

// TestDebugFlag is a global declaration to enable debugging in all tests for this package
// TODO: TestDebugFlag should not be present when running a main progam instead of tests
var TestDebugFlag = flag.Bool("testdebug", false, "enable test debugging information")

// VeryVerbose is a global declaration to enable very verbose debugging in all tests
// TODO: need a way to force test.verbose mode if enabling VeryVerbose
var VeryVerbose = flag.Bool("vv", false, "enable very verbose test output")

// AssertError reports a testing failure when a given error is non nil
func AssertError(t *testing.T, err error) {
	if err != nil {
		t.Logf("%s operation failed: %v", t.Name(), err)
		t.Fail()
	}
}

// AssertEqual reports a testing failure when a given actual item (interface{}) is not equal to the expected value
func AssertEqual(t *testing.T, item string, expected, actual interface{}) {
	if actual == nil {
		if expected != nil {
			t.Logf("%s %s mismatch -\nexpected %#v\ngot      %#v", t.Name(), item, expected, actual)
			t.Fail()
		}
	} else if reflect.TypeOf(actual).Comparable() {
		if expected != actual {
			t.Logf("%s %s mismatch -\nexpected %#v\ngot      %#v", t.Name(), item, expected, actual)
			t.Fail()
		}
	} else {
		if !reflect.DeepEqual(expected, actual) {
			t.Logf("%s %s mismatch -\nexpected %#v\ngot      %#v", t.Name(), item, expected, actual)
			t.Fail()
		}
	}
}

// AssertNotEqual reports a testing failure when a given actual item (interface{}) is equal to the expected value
func AssertNotEqual(t *testing.T, item string, expected, actual interface{}) {
	if actual == nil {
		if expected == nil {
			t.Logf("%s %s should not be %#v", t.Name(), item, actual)
			t.Fail()
		}
	} else if reflect.TypeOf(actual).Comparable() {
		if expected == actual {
			t.Logf("%s %s should not be %#v", t.Name(), item, actual)
			t.Fail()
		}
	} else {
		if reflect.DeepEqual(expected, actual) {
			t.Logf("%s %s should not be %#v", t.Name(), item, actual)
			t.Fail()
		}
	}
}

// IsReleaseTest check command argument 'releaseTest' for skip or run long run test
func IsReleaseTest() bool {
	args := os.Args
	var runTest = false
	for _, v := range args {
		if v == "releaseTest" {
			runTest = true
			break
		}
	}
	return runTest
}
