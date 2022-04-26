package utils

import (
	"time"
)

var logSquelch = make(map[string]time.Time)

// LogSquelch provides the function to reduce the number of times some
// logs are emitted to allow log files to not get full.
// Returns true if log is to be squelched, false otherwise
func LogSquelch(msg string, expiration time.Duration) (bSquelch bool) {

	expTime, present := logSquelch[msg]

	if present && time.Now().Before(expTime) {

		bSquelch = true

	} else {

		logSquelch[msg] = time.Now().Add(expiration)

	}

	return bSquelch
}
