package cloudant

import (
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/utils"
)

type notificationsCache struct {
	ExpireTime time.Time
	Cache      *NotificationsResult
}

var noteCacheLookup map[string]*notificationsCache

func getNotificationsCache(url string) *NotificationsResult {

	if noteCacheLookup == nil {
		noteCacheLookup = make(map[string]*notificationsCache)
	}

	noteCache := noteCacheLookup[url]
	if noteCache != nil {
		if time.Now().After(noteCache.ExpireTime) {
			noteCacheLookup[url] = nil
		}
	}

	if noteCache == nil {
		return nil
	}
	return noteCache.Cache
}

func setNotificationsCache(data *NotificationsResult, url string) {

	cacheExpireInterval := utils.GetEnvMinutes(NotificationCacheMinutes, 15)
	noteCache := &notificationsCache{ExpireTime: time.Now().Add(cacheExpireInterval), Cache: data}

	if noteCacheLookup == nil {
		noteCacheLookup = make(map[string]*notificationsCache)
	}
	noteCacheLookup[url] = noteCache

}

type nameMappingCache struct {
	ExpireTime time.Time
	Cache      *NameMapping
}

var nameMapLookup map[string]*nameMappingCache
var nameMapCache *nameMappingCache

func getNameMappingCache(url string) *NameMapping {

	if nameMapLookup == nil {
		nameMapLookup = make(map[string]*nameMappingCache)
	}

	nameMapCache := nameMapLookup[url]
	if nameMapCache != nil {
		if time.Now().After(nameMapCache.ExpireTime) {
			nameMapCache = nil
		}
	}

	if nameMapCache == nil {
		return nil
	}
	return nameMapCache.Cache
}

func setNameMappingCache(data *NameMapping, url string) {

	cacheExpireInterval := utils.GetEnvMinutes(NameMapCacheMinutes, 15)
	nameMapCache = &nameMappingCache{ExpireTime: time.Now().Add(cacheExpireInterval), Cache: data}

	if nameMapLookup == nil {
		nameMapLookup = make(map[string]*nameMappingCache)
	}
	nameMapLookup[url] = nameMapCache
}
