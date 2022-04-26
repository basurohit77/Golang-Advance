package token

import (
	"sync"
	"time"
)

type singleCache interface {
	initCache()
	updateCache()
	initializeIfNeeded()
	setQuitUpdateLoop(chan bool)
	getExpiryTime() time.Duration
	isInitialized() bool
}

type singleCacheUtils struct {
	mutex            sync.RWMutex
	expiryTime       time.Duration // duration until the cache attempts a refresh
	quitCacheLoop    chan bool     // used to stop the cache refresh cycle
	utilsInitialized bool          // true if cache initialized with default values and update was attempted
}

// schedules the cache refresh interval for the `updateCache()` function
func cacheInterval(o singleCache) {
	o.setQuitUpdateLoop(schedule(o.updateCache, o.getExpiryTime()))
}

func (c *singleCacheUtils) setQuitUpdateLoop(quitChan chan bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.quitCacheLoop = quitChan
}

func (c *singleCacheUtils) getExpiryTime() time.Duration {
	return c.expiryTime
}

func (c *singleCacheUtils) isInitialized() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.utilsInitialized
}
