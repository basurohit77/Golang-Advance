package utils

import "testing"

func TestCompare(t *testing.T) {

	tm := "2018-07-31T16:18:12Z"
	if !CompareTimeStr(tm, tm) {
		t.Error("Failed to match times ", tm)
	}

	tm = "2018-07-10T20:53:37.359Z"
	if !CompareTimeStr(tm, tm) {
		t.Error("Failed to match times ", tm)
	}

	tm = "2018-07-31T16:18:12.892-05:00"
	if !CompareTimeStr(tm, tm) {
		t.Error("Failed to match times ", tm)
	}

	tm = "2018-07-31T16:18:12.892+05:00"
	if !CompareTimeStr(tm, tm) {
		t.Error("Failed to match times ", tm)
	}

	tm = "2018-10-31T16:18:12.892+05:00"
	tm2 := "2018-11-01T16:18:12.892+05:00"
	if CompareTimeStr(tm, tm2) {
		t.Error("Failed to find differences ", tm)
	}

}

func TestCompareError(t *testing.T) {

	tm := "2018-07-31T16:18:12Z"
	if CompareTimeStr(tm, "") {
		t.Error("Failed to see empty time")
	}

	tm2 := "2018-07-31T16:18:12Z"
	tm = "foobar"
	if CompareTimeStr(tm, tm2) {
		t.Error("Failed to see bad time")
	}

	tm = "2018-07-31T16:18:12Z"
	tm2 = "foobar"
	if CompareTimeStr(tm, tm2) {
		t.Error("Failed to see bad time")
	}

	tm = "2018-07-31T16:18:12EST"
	tm2 = "2018-07-31T16:18:12Z"
	if CompareTimeStr(tm, tm2) {
		t.Error("Failed to see bad time")
	}

	tm = "2018-07-31T16:18:12Z"
	tm2 = "2018-07-31T16:18:12EST"
	if CompareTimeStr(tm, tm2) {
		t.Error("Failed to see bad time")
	}
}
