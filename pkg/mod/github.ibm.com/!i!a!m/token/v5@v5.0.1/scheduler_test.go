package token

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAFunctionStartedAsScheduled(t *testing.T) {
	var mutex sync.RWMutex
	callCount := 0

	task := func() {
		mutex.Lock()
		defer mutex.Unlock()
		callCount++
	}

	stop := schedule(task, 1*time.Second)
	stop <- true
	time.Sleep(1 * time.Second)
	mutex.RLock()
	defer mutex.RUnlock()
	assert.Equal(t, callCount, 1)
}

func TestFunctionCalledAtEachInterval(t *testing.T) {
	var mutex sync.RWMutex
	callCount := 0

	task := func() {
		mutex.Lock()
		defer mutex.Unlock()
		callCount++
	}

	stop := schedule(task, 1*time.Second)
	time.Sleep(2 * time.Second)
	stop <- true
	mutex.RLock()
	defer mutex.RUnlock()
	// There is currently an issue with testify where GreaterOrEqual is undefined. Until that issue is fixed.
	//assert.GreaterOrEqual(t, callCount, 2)
	if callCount < 2 {
		t.Errorf("Task should be called at least twice. It is called %d", callCount)
	}

}

func TestSingleRunScheduler(t *testing.T) {
	var mutex sync.RWMutex
	callCount := 0

	task := func() {
		mutex.Lock()
		defer mutex.Unlock()
		callCount++
	}

	singleSchedule(task, 1)

	time.Sleep(2 * time.Second)

	mutex.RLock()
	defer mutex.RUnlock()
	assert.Equal(t, callCount, 1)

}
