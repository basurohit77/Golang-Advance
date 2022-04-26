package utils

import (
	"errors"
	"log"
	"time"

	"github.com/google/go-github/github"
)

// IsSuccessfulRequest is a general function to check if the combination of the
// error and http response indicate a successful request.
func IsSuccessfulRequest(resp *github.Response, inError error) error {
	if inError != nil {
		return inError
	}

	if resp.StatusCode < 200 && resp.StatusCode > 299 {
		return errors.New("Bad http status code: " + resp.Status)
	}

	cacheHitVal := resp.Header["X-From-Cache"]
	if len(cacheHitVal) > 0 {
		log.Println("Cache hit value:", cacheHitVal)
	}

	return nil
}

// GithubTimestampToTime converts the github package version of Timstamp to a standard time.Time
func GithubTimestampToTime(ts github.Timestamp) time.Time {
	e := ts.UTC().Unix()
	t := time.Unix(e, 0)
	return t
}
