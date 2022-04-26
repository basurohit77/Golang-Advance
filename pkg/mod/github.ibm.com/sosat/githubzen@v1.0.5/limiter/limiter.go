package limiter

import (
	"log"
	"time"
)

type Limiter struct {
	Name      string
	Limit     int
	Remaining int
	Reset     time.Time
	Threshold float64
}

func (lim *Limiter) Set(limit int, remaining int, reset time.Time) {
	lim.Limit = limit
	lim.Remaining = remaining
	lim.Reset = reset
}

func (lim *Limiter) Wait() {
	// log.Printf("wait, limiter=%s\n", lim.Name)
	d := 30 * time.Second // add a bit of time to account for time drift

	// threshold := lim.Threshold * float64(lim.Limit)
	var threshold float64 = 5 // instead of percentage, sleep when we only have 5 remaining
	if float64(lim.Remaining) < threshold {
		delay := time.Until(lim.Reset) + d
		log.Printf("Wait: limiterName=%s, rateRemaining=%d, threshold=%f, reset=%q, resetDuration=%v, resetWithDelay=%v\n", lim.Name, lim.Remaining, threshold, lim.Reset.String(), time.Until(lim.Reset), delay)

		if delay > d {
			time.Sleep(delay)
		} else {
			time.Sleep(d) // minimum sleep time
		}
	}
}
