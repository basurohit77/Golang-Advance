package utils

import (
	"fmt"
	"time"
)

// CompareTimeStr will compare two time strings for equality.
// Returns true if they are the same, false otherwise.
func CompareTimeStr(t1, t2 string) bool {

	if (len(t1) == 0 && len(t2) != 0) || (len(t1) != 0 && len(t2) == 0) {
		return false
	}

	// Move to common Format("2006-01-02T15:04:05Z")
	ts1, err := VerifyAndConvertTimestamp(t1)
	if err != nil {
		return false
	}
	ts2, err := VerifyAndConvertTimestamp(t2)
	if err != nil {
		return false
	}

	time1, err := time.Parse(time.RFC3339, ts1)
	if err != nil {
		fmt.Println(err)
		return false
	}
	time2, err := time.Parse(time.RFC3339, ts2)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return time1.Equal(time2)
}
