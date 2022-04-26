package utils

import (
	"log"
	"sync"
	"time"
)

// authCache contains information about caches bad tokens
// authCache acts as a filter to tell you bad tokens or keys only
type authCache struct {
	tokens map[string]time.Time
}

var (
	// This is the global instance of the auth cache
	authCacheInstance *authCache

	// Mutext to protect cache changes
	authCacheMutex = &sync.Mutex{}
)

// AddBadAuth will add an invalid token to the cache
// Use this so if this token is used again, we can quickly
// determine that it is bad.
func AddBadAuth(authKey string) {

	if authKey == "" {
		return
	}

	authCacheMutex.Lock()
	defer authCacheMutex.Unlock()
	cache := getAuthCache()

	if cache != nil {
		cache.tokens[authKey] = time.Now().Add(30 * time.Second)
	}

}

// IsBadAuth will return true if the input key was seen before
// and is bad.  If false is returned, we don't know if the input
// key is bad or good.
func IsBadAuth(authKey string) bool {
	authCacheMutex.Lock()
	defer authCacheMutex.Unlock()
	cache := getAuthCache()

	if cache != nil {
		if expireTime, ok := cache.tokens[authKey]; ok {
			if time.Now().Before(expireTime) {
				log.Println("Found bad key.  Bad entry expires at ", expireTime.Format(time.RFC3339))
				return true
			}
			delete(cache.tokens, authKey)
		}
	}

	return false
}

// getAuthCache will retrieve the authorization cache.
func getAuthCache() *authCache {

	if authCacheInstance == nil {
		authCacheInstance = makeAuthCache()
		return authCacheInstance
	}

	return authCacheInstance
}

// makeAuthCache will create a new auth cache
func makeAuthCache() *authCache {

	cache := new(authCache)
	cache.tokens = make(map[string]time.Time)

	return cache
}
