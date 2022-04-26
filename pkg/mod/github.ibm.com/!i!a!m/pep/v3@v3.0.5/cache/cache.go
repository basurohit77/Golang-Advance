package cache

import (
	"encoding/json"
	"time"

	"github.com/VictoriaMetrics/fastcache"
)

// The default time-to-live for permit cache entries
const DefaultTTL time.Duration = time.Duration(10 * time.Minute)

// The default time-to-live for deny cache entries
const DefaultDenyTTL time.Duration = time.Duration(2 * time.Minute)

// DecisionCacheConfig stores cache configuration information
type DecisionCacheConfig struct {
	CacheSize        int           // The cache size in MB
	TTL              time.Duration // The user-defined time-to-live for permit cache entries
	DeniedTTL        time.Duration // The user-defined time-to-live for deny cache entries
	DisableDenied    bool          // Indicate whether denied decision should be stored or not.
	DisablePermitted bool          // Indicate whether permitted decision should be stored or not.
}

// CachedDecision stores the policy evaluation decision and the unix time at which the decision expires.
type CachedDecision struct {
	Permitted bool
	ExpiresAt time.Time
	Reason    int
}

// Stats represents cache statistics
type Stats struct {
	// Hits is the number of successful cache hits
	Hits uint64

	// Misses is the number of cache misses
	Misses uint64

	// EntriesCount is the current number of entries in the cache
	EntriesCount uint64

	// Capacity is the total size of the cache, including unused space, in bytes
	Capacity uint64

	// BytesSize is the current size of the cache in bytes
	BytesSize uint64
}

// Expired returns true if the decision has expired its TTL, false otherwise
func (d CachedDecision) Expired() bool {
	return time.Now().UnixNano() >= d.ExpiresAt.UnixNano()
}

// DecisionCache is the interface that wraps the decision cache Get/Set methods.
//
// The default implementation is based on fastcache. The user can also provide their own implemention that satisfies this interface.
//
// Get will return a CachedDecision and a boolean indicating if the key exists or not.
//
// Set will store the CachedDecision
//
type DecisionCache interface {
	// Given the key, returns the CachedDecision as a pointer. Nil if not found.
	Get(key []byte) *CachedDecision
	// Stores the CachedDecision associated with the key
	Set(key []byte, values bool, ttl time.Duration, reason int)
	// GetConfig returns the stored configuration
	GetConfig() *DecisionCacheConfig
	// GetStatistics returns statistics about the cache represented by the `Stats` struct
	GetStatistics() Stats
	// Reset removes all the items from the cache.
	Reset()
}

// This is the one pep implementation of the cache, which is based on fastcache
type decisionCache struct {
	config *DecisionCacheConfig
	fcache *fastcache.Cache
}

func (c decisionCache) Get(key []byte) *CachedDecision {
	b := c.fcache.Get(nil, key)
	if b == nil {
		return nil
	}
	v := CachedDecision{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		//TODO logs whenever the logger is implemented
		return nil
	}
	return &v
}

func (c decisionCache) Set(key []byte, v bool, ttl time.Duration, r int) {
	d := CachedDecision{
		Permitted: v,
		ExpiresAt: time.Now().Add(ttl),
		Reason:    r,
	}
	b, _ := json.Marshal(d)
	c.fcache.Set(key, b)
}

func (c decisionCache) GetConfig() *DecisionCacheConfig {
	return c.config
}

func (c decisionCache) GetStatistics() Stats {
	var fs fastcache.Stats
	var s Stats
	c.fcache.UpdateStats(&fs)

	s.BytesSize = fs.BytesSize
	if c.GetConfig().CacheSize < 32 {
		s.Capacity = 32
	} else {
		s.Capacity = uint64(c.GetConfig().CacheSize)
	}
	s.EntriesCount = fs.EntriesCount
	s.Hits = fs.GetCalls - fs.Misses
	s.Misses = fs.Misses

	return s
}

func (c decisionCache) Reset() {
	c.fcache.Reset()
}

// NewDecisionCache returns an implementation of the DecisionCache that is used to store policy evaluation decision
//
// The cache engine used in the implementation is fastcache (https://github.com/VictoriaMetrics/fastcache)
func NewDecisionCache(config *DecisionCacheConfig) DecisionCache {

	cache := decisionCache{}
	if config == nil {
		return nil
	}

	cache.config = config
	cache.fcache = fastcache.New(config.CacheSize)
	return cache
}
