package token

import (
	"time"
)

// Schedule executes the specified task at every interval
// A channel is returned and must be used to stop the running task.
func schedule(task func(), interval time.Duration) chan bool {
	stop := make(chan bool)
	go func() {
		for {
			go task()
			select {
			case <-time.After(interval):
			case <-stop:
				return
			}
		}
	}()
	return stop
}

// task is the function to schedule
// waitTime is an int64 representing the amount of time in seconds to wait until running the task
func singleSchedule(task func(), waitTime int64) {
	go func() {
		<-time.After(time.Duration(float64(waitTime)) * time.Second)
		task()
	}()
}
