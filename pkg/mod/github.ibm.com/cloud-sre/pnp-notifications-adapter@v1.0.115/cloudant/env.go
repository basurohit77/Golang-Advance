package cloudant

const (
	// ServicesURLEnv is the environment variable to find the URL override for the service names
	ServicesURLEnv = "PNP_SERVICES_URL"

	// RuntimesURLEnv is the environment variable to find the URL override for the runtime names
	RuntimesURLEnv = "PNP_RUNTIMES_URL"

	// PlatformURLEnv is the environment variable to find the URL override for the platform names
	PlatformURLEnv = "PNP_PLATFORM_URL"

	// NotificationsURLEnv is the environment variable to find a URL override
	NotificationsURLEnv = "PNP_NOTIFICATIONS_URL"

	// AccountID is the environment variable to pull the account ID
	AccountID = "PNP_CLOUDANT_ID"

	// AccountPW is the environment variable to pull the account PW
	AccountPW = "PNP_CLOUDANT_PW" // #nosec G101

	// NotificationCacheMinutes is the lifespan of notification cache entries in minutes
	NotificationCacheMinutes = "PNP_NOTIFICATION_CACHE_TIME"

	// NameMapCacheMinutes is the lifespan of name map cache entries in minutes
	NameMapCacheMinutes = "PNP_NAMEMAP_CACHE_TIME"
)
