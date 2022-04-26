package pep

import (
	"sync"
)

// cacheKeyInfo contains informations needed to build the cache key
type cacheKeyInfo struct {
	rw sync.RWMutex
	CacheKeyPattern
}

type sliceOfString []string
type doubleSliceOfString [][]string

var currentCacheKeyInfo cacheKeyInfo

func (src doubleSliceOfString) duplicate() (dest doubleSliceOfString) {
	if src == nil {
		return
	}

	dest = make(doubleSliceOfString, len(src))
	for i, v := range src {
		dest[i] = make(sliceOfString, len(v))
		copy(dest[i], v)
	}
	return
}

func (src sliceOfString) duplicate() (dest sliceOfString) {
	if src == nil {
		return
	}
	dest = make(sliceOfString, len(src))
	copy(dest, src)
	return
}

func (c *cacheKeyInfo) getCacheKeyPattern() (pattern CacheKeyPattern) {

	c.rw.RLock()

	pattern.Order = sliceOfString(c.Order).duplicate()

	pattern.Subject = doubleSliceOfString(c.Subject).duplicate()

	pattern.Resource = doubleSliceOfString(c.Resource).duplicate()

	c.rw.RUnlock()
	return
}

func (c *cacheKeyInfo) storeCacheKeyPattern(src CacheKeyPattern) {

	c.rw.Lock()

	c.Order = sliceOfString(src.Order).duplicate()

	c.Subject = doubleSliceOfString(src.Subject).duplicate()

	c.Resource = doubleSliceOfString(src.Resource).duplicate()

	c.rw.Unlock()
}
