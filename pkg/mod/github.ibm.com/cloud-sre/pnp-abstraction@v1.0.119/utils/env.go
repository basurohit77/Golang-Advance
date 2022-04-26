package utils

import (
	"os"
	"strconv"
	"time"
)

// GetEnvMinutes provides a convenient way to get a Duration from an env with a default
func GetEnvMinutes(envVar string, defaultMinutes time.Duration) (num time.Duration) {
	num = -1

	strVar := os.Getenv(envVar)
	if strVar != "" {
		iNum, err := strconv.ParseInt(strVar, 10, 64)
		if err == nil {
			num = time.Duration(iNum)
		}
	}

	if num == -1 {
		num = defaultMinutes
	}

	return num * time.Minute
}
