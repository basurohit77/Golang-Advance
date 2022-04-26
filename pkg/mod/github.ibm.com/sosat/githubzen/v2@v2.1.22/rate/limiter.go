package rate

import (
	"log"
	"time"

	"github.com/google/go-github/github"
	"github.ibm.com/sosat/githubzen/v2/utils"
)

// Limiter contains the basic information for rate limiting
type Limiter struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// NewLimiter creates a new default rate limiter
// Pass an initial value for the limit.  Requests will determine the limit dynamically.
func NewLimiter(limit, remaining int, reset time.Time) *Limiter {
	return &Limiter{Limit: limit, Remaining: remaining, Reset: reset}
}

// HasRateError will use the limiter to check if a rate limit has occured. This should be
// called after every call to github so that it can acurately track the progress of rate limiting
func (l *Limiter) HasRateError(resp *github.Response, err error) bool {
	l.Limit = resp.Rate.Limit
	l.Remaining = resp.Rate.Remaining
	l.Reset = utils.GithubTimestampToTime(resp.Rate.Reset)
	return isRatelimitError(resp, err)
}

// Wait will examine the current condition of the limiter and apply a wait as needed.
// This function will return immediately if no wait is necessary.
func (l *Limiter) Wait() {
	// log.Printf("wait, limiter=%s\n", lim.Name)
	d := 30 * time.Second // add a bit of time to account for time drift

	// threshold := l.Threshold * float64(lim.Limit)
	var threshold float64 = 5 // instead of percentage, sleep when we only have 5 remaining
	if float64(l.Remaining) < threshold {
		delay := time.Until(l.Reset) + d
		log.Printf("Wait: rateRemaining=%d, threshold=%f, reset=%q, resetDuration=%v, resetWithDelay=%v\n", l.Remaining, threshold, l.Reset.String(), time.Until(l.Reset), delay)

		if delay > d {
			time.Sleep(delay)
		} else {
			time.Sleep(d) // minimum sleep time
		}
	}
}

func isRatelimitError(resp *github.Response, err error) bool {
	if err, ok := err.(*github.RateLimitError); ok {
		log.Printf("isRatelimitError: %#v\n", err)
		return true
	}

	if err != nil && resp.StatusCode == 403 {
		log.Printf("isRatelimitError: statusCode=%d, err: %#v\n", resp.StatusCode, err)

		// 2019/07/01 18:37:04.649008 gh-repos.go:153: getRepository: statusCode=403, err: &github.ErrorResponse{Response:(*http.Response)(0xc000180120), Message:"", Errors:[]github.Error(nil),     Block:(*struct { Reason string "json:\"reason,omitempty\""; CreatedAt *github.Timestamp "json:\"created_at,omitempty\"" })(nil), DocumentationURL:""}

		// 2019/07/01 18:37:04.649008 gh-repos.go:153: getRepository: statusCode=403,
		// err: &github.ErrorResponse{Response:(*http.Response)(0xc000180120),
		// Message:"",
		// Errors:[]github.Error(nil),
		// Block:(*struct { Reason string "json:\"reason,omitempty\""; CreatedAt *github.Timestamp "json:\"created_at,omitempty\"" })(nil),
		// DocumentationURL:""}

		// 2019/07/01 18:30:57.487500 gh-repos.go:154: GET https://github.ibm.com/api/v3/repos/cloud-sre/pnp-abstraction: 403  []

		return true
	}

	return false
}
