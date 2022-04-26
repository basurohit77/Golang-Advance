package snowauth

import (
	"log"
	"sync"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/api"
)

// cache is used as the caching struct
type cache struct {
	m                                 sync.RWMutex              // mutex to control read/write ops
	snAuthCache                       map[string][]snAuthCache  // email as key and slice of SNow Auth as value
	MaxSize                           int                       // maximum size allowed for the cache
	CleanupInterval                   time.Duration             // how often the cache is checked for cleanup
	ExpirationTime                    time.Duration             // the expiration the user will be given from now(eg. 30min)
	ExpirationTimeUnauthorized        time.Duration             // the expiration the user will be given from now(eg. 5min). Only applies for users that have been unauthorized to a service
	snAuthUserTypeCache               map[string]snAuthUsers    // cache the email as key, and user type as value
	SNAuthUserTypeCacheExpirationTime time.Duration             // expiration time for the results returned by ServiceNow
	snAuthServiceTypeCache            map[string]snAuthServices // cache the service as key, and value is the serviceType e.g cloud
	SNAuthServiceTypeExpirationTime   time.Duration             // expiration time for the results returned by ServiceNow
}

// snAuthCache holds information for checking user authorization to a specific crn
type snAuthCache struct {
	crn         string
	allowed     bool
	timeExpires int64
}

// snAuthUsers holds information on each email, as returned by SNow
type snAuthUsers struct {
	UserType    string
	timeExpires int64
}

// snAuthServices holds information on each service, as returned by SNow
type snAuthServices struct {
	ServiceType string
	timeExpires int64
}

func newSNAuthCache() *cache {

	var c cache
	c.snAuthCache = make(map[string][]snAuthCache)
	c.MaxSize = 10000
	c.CleanupInterval = 3 * time.Minute
	c.ExpirationTime = 30 * time.Minute
	c.ExpirationTimeUnauthorized = 5 * time.Minute

	c.snAuthUserTypeCache = make(map[string]snAuthUsers)
	c.SNAuthUserTypeCacheExpirationTime = c.ExpirationTime

	c.snAuthServiceTypeCache = make(map[string]snAuthServices)
	c.SNAuthServiceTypeExpirationTime = 1 * time.Hour

	go c.cleanUp()

	return &c

}

// getUserType returns the user type
func (c *cache) getUserType(email string) string {

	c.m.RLock()
	defer c.m.RUnlock()

	cacheEntry := c.snAuthUserTypeCache[email]

	if cacheEntry != (snAuthUsers{}) {
		// only return if the cache entry is still valid
		if cacheEntry.timeExpires >= time.Now().Unix() {
			return cacheEntry.UserType
		}
	}

	return ""

}

func (c *cache) addToUsers(r result) {

	if len(c.snAuthUserTypeCache) < c.MaxSize {

		var cacheEntry snAuthUsers
		cacheEntry.UserType = r.UserType
		cacheEntry.timeExpires = time.Now().Add(c.SNAuthUserTypeCacheExpirationTime).Unix()

		c.m.Lock()
		defer c.m.Unlock()

		// add entry to the cache
		c.snAuthUserTypeCache[r.UserName] = cacheEntry

	} else {
		log.Printf("SNowAuth: unable to add to snAuthUserTypeCache, cache is at the maximum size. current size=%d. Increase the cache size if needed.\n", len(c.snAuthCache))
	}

}

// getServiceType returns the service type
func (c *cache) getServiceType(service string) string {

	c.m.RLock()
	defer c.m.RUnlock()

	cacheEntry := c.snAuthServiceTypeCache[service]

	if cacheEntry != (snAuthServices{}) {

		// only return if the cache entry is still valid
		if cacheEntry.timeExpires >= time.Now().Unix() {
			return cacheEntry.ServiceType
		}
	}

	return ""

}

func (c *cache) addToServices(r result) {

	if len(c.snAuthServiceTypeCache) < c.MaxSize {

		svcName, err := api.GetServiceFromCRN(r.CRN)
		if err != nil {
			log.Println(err)
		} else {

			if svcName != "" {

				var cacheEntry snAuthServices
				cacheEntry.ServiceType = r.ServiceType
				cacheEntry.timeExpires = time.Now().Add(c.SNAuthServiceTypeExpirationTime).Unix()

				c.m.Lock()
				defer c.m.Unlock()
				// add entry to the cache
				c.snAuthServiceTypeCache[svcName] = cacheEntry
			}
		}

	} else {
		log.Printf("SNowAuth: unable to add to snAuthServiceTypeCache, cache is at the maximum size. current size=%d. Increase the cache size if needed.\n", len(c.snAuthCache))
	}

}

// cleanUp removes all entries that are older than what's defined in cacheEntry.timeExpires
func (c *cache) cleanUp() {

	go (func() {

		for {
			// log.Printf("SNowAuth: current cache size=%d; maximum cache size allowed=%d\n", len(c.snAuthCache), c.MaxSize)
			now := time.Now().Unix()

			c.m.RLock()

			// clean up the snAuthCache cache
			for email, cacheEntries := range c.snAuthCache {
				for _, cacheEntry := range cacheEntries {
					if now >= cacheEntry.timeExpires {
						if len(cacheEntries) > 1 { // delete only one entry and leave the rest

							var updatedEntry []snAuthCache
							for _, entry := range c.snAuthCache[email] {
								if entry.crn != cacheEntry.crn {
									updatedEntry = append(updatedEntry, entry)
								}
							}

							c.m.RUnlock()
							c.m.Lock()
							c.snAuthCache[email] = updatedEntry
							c.m.Unlock()
							c.m.RLock()

						} else if len(cacheEntries) == 1 { // delete whole entry if there is only one

							c.m.RUnlock()
							c.m.Lock()
							delete(c.snAuthCache, email)
							c.m.Unlock()
							c.m.RLock()

						}
					}
				}
			}

			// clean up the snAuthUserTypeCache cache
			for email, cacheEntry := range c.snAuthUserTypeCache {
				if now >= cacheEntry.timeExpires {

					c.m.RUnlock()
					c.m.Lock()
					delete(c.snAuthUserTypeCache, email)
					c.m.Unlock()
					c.m.RLock()

				}
			}

			// clean up the snAuthServiceType cache
			for svc, cacheEntry := range c.snAuthServiceTypeCache {
				if now >= cacheEntry.timeExpires {

					c.m.RUnlock()
					c.m.Lock()
					delete(c.snAuthServiceTypeCache, svc)
					c.m.Unlock()
					c.m.RLock()

				}
			}

			c.m.RUnlock()

			time.Sleep(c.CleanupInterval)
		}

	})()

}

// search searches the cache for the provided CRN and email
// the bool from the left returns the authorization,
// and the bool from the right tells you whether the entry was found in cache AND is not expired(true=found in cache AND is not expired; false=not found in cache or is expired)
func (c *cache) search(email, crn string) (bool, bool) {
	c.m.RLock()
	defer c.m.RUnlock()

	cacheEntries := c.snAuthCache[email]

	for _, entry := range cacheEntries {
		if entry.crn == crn {
			if entry.timeExpires >= time.Now().Unix() { // only return non-expired authorizations
				return entry.allowed, true // found in cache and is valid
			}
		}
	}

	// not authorized, and either not found in cache or found in cache but it's expired
	return false, false
}

// add adds a new entry to the cache
func (c *cache) add(email, crn string, allowed bool) {

	// add to cache only if there is room for that
	if len(c.snAuthCache) < c.MaxSize {

		var cacheEntry snAuthCache
		cacheEntry.allowed = allowed
		cacheEntry.crn = crn
		if allowed { // add a different expiration time for "true" auth
			cacheEntry.timeExpires = time.Now().Add(c.ExpirationTime).Unix()
		} else { // add a different expiration time for "false" auth
			cacheEntry.timeExpires = time.Now().Add(c.ExpirationTimeUnauthorized).Unix()
		}

		c.m.Lock()
		defer c.m.Unlock()

		// add entry to the cache
		c.snAuthCache[email] = append(c.snAuthCache[email], cacheEntry)

	} else {
		log.Printf("SNowAuth: unable to add to snAuthCache, cache is at the maximum size. current size=%d. Increase the cache size if needed.\n", len(c.snAuthCache))
	}

}
