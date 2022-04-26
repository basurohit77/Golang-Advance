package testutils

import "time"

const format = "2006-01-02T15:04:05.000Z"

// GetDate will return a date string based on offset from now
func GetDate(offset time.Duration) string {
	return time.Now().Add(offset).UTC().Format(format)
}
