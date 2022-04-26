package snowauth

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	emailCache = "email@email.com"
	crn        = "crn:v1:bluemix:public:svc1:::::"
	crn2       = "crn:v1:bluemix:public:svc2:::::"
)

func TestCache(t *testing.T) {
	cache := newSNAuthCache()
	cache.CleanupInterval = 1 * time.Second
	cache.ExpirationTime = 1 * time.Second
	cache.SNAuthServiceTypeExpirationTime = 2 * time.Second
	cache.SNAuthUserTypeCacheExpirationTime = 2 * time.Second

	// cache crn
	cache.add(emailCache, crn, true)
	// should find the crn added above
	auth, found := cache.search(emailCache, crn)
	// should return "true" because email and crn are cached
	assert.True(t, found)
	assert.True(t, auth)

	// allow time for cache to remove expired entry
	time.Sleep(3 * time.Second)

	// search for the same crn but should not find it as it was removed from cache(expired)
	auth, found = cache.search(emailCache, crn)
	// should return "false" because email and crn are no longer cached
	assert.False(t, found)
	assert.False(t, auth)

	// search crn that is not cached
	auth, found = cache.search(emailCache, crn2)
	// should return "false" because crn is cached
	assert.False(t, found)
	assert.False(t, auth)

	r := result{
		UserName:    email,
		CRN:         crns[1],
		UserType:    "cloud",
		ServiceType: "noncloud",
		Authorized: authorized{
			Valid:   false,
			Message: "some msg",
		},
	}

	cache.addToServices(r)
	assert.Equal(t, r.ServiceType, cache.getServiceType(strings.Split(r.CRN, ":")[4]))

	cache.addToUsers(r)
	assert.Equal(t, r.UserType, cache.getUserType(r.UserName))

	// allow time for cache to remove expired entry
	time.Sleep(3 * time.Second)

	// test again and the cache should have been cleanup
	assert.Equal(t, "", cache.getServiceType(strings.Split(r.CRN, ":")[4]))
	assert.Equal(t, "", cache.getUserType(r.UserName))

}
