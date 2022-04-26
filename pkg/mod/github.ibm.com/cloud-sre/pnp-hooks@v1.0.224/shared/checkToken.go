package shared

import "os"

var SnowToken = os.Getenv("SNOW_TOKEN")

// HasValidToken checks whether the token provided is valid
func HasValidToken(token string) bool {

	if token == SnowToken && token != "" {
		return true
	}

	return false

}
