package pep

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.ibm.com/IAM/pep/v3/cache"
	"github.ibm.com/IAM/pep/v3/stats"
	"github.ibm.com/IAM/token/v5"
)

// DeploymentEnvironment is an alias representing the environment on which the pep is running, e.g. stating, production or Custom
type DeploymentEnvironment = token.DeploymentEnvironment

// Staging, Production, PrivateStaging, PrivateProduction and Custom environment types denoted using iota
const (
	Production DeploymentEnvironment = iota
	Staging
	Custom
	PrivateProduction
	PrivateStaging
)

const stagingBaseURL = "https://iam.test.cloud.ibm.com"
const prodBaseURL = "https://iam.cloud.ibm.com"
const privateStagingBaseURL = "https://private.iam.test.cloud.ibm.com"
const privateProdBaseURL = "https://private.iam.cloud.ibm.com"

const (
	// endpoints
	stagingAuthzEndpoint = stagingBaseURL + "/v2/authz"
	stagingListEndpoint  = stagingBaseURL + "/v2/authz/bulk"
	stagingTokenEndpoint = stagingBaseURL + "/identity/token"
	stagingKeyEndpoint   = stagingBaseURL + "/identity/keys"
	prodAuthzEndpoint    = prodBaseURL + "/v2/authz"
	prodListEndpoint     = prodBaseURL + "/v2/authz/bulk"
	prodTokenEndpoint    = prodBaseURL + "/identity/token"
	prodKeyEndpoint      = prodBaseURL + "/identity/keys"
	//only identity is ready with private hostnames at the moment.
	privateStagingAuthzEndpoint = stagingBaseURL + "/v2/authz"
	privateStagingListEndpoint  = stagingBaseURL + "/v2/authz/bulk"
	privateStagingTokenEndpoint = privateStagingBaseURL + "/identity/token"
	privateStagingKeyEndpoint   = privateStagingBaseURL + "/identity/keys"
	privateProdAuthzEndpoint    = prodBaseURL + "/v2/authz"
	privateProdListEndpoint     = prodBaseURL + "/v2/authz/bulk"
	privateProdTokenEndpoint    = privateProdBaseURL + "/identity/token"
	privateProdKeyEndpoint      = privateProdBaseURL + "/identity/keys"
	rolesRoute                  = "/v2/authz/roles"

	// stats/metrics
	requestsServicedByPEP = "requestsServicedByPEP"
	originalRequestsToPDP = "originalRequestsToPDP"
	failedUserRequests    = "failedUserRequests"
	apiErr                = "APIerr"
	//	timeouts              = "timeouts"

	//Average PDP request latency coming later
)

// Config contains the required configuration variables needed for the PEP library to run
type Config struct {

	// There are required
	Environment DeploymentEnvironment
	APIKey      string

	// Optional, but recommended for production token exchange
	ClientID     string
	ClientSecret string

	// Optional, but recommended to use tokens authorized against resources of a specific service.
	TokenScope string

	// These will be automatically filled out when the Environment is specified if it is staging or production
	AuthzEndpoint string
	ListEndpoint  string
	TokenEndpoint string
	KeyEndpoint   string
	RolesEndpoint string

	// AuthzRetry set to true will enable a maximum three time retry for authz calls. Incrementally attempting at 1, 2, and finally 6 seconds before an error is returned to the caller.
	AuthzRetry bool

	// These will be automatically filled out with default values if not specified by the user. This is the size of the cache in MB. The minium size is 32MB.
	DecisionCacheSize int

	// Cache is enabled by default
	DisableCache bool

	// Denied cache is enabled by default and is independent of DisableCache
	DisableDeniedCache bool

	// Enable expired cache entries being returned when authz endpoint has errors
	EnableExpiredCache bool

	// This is the optional configurable TTL for permitted cache entries.
	// If this is not configured (or set to 0), the maxCacheAgeSeconds from PDP will be used.
	// If for some reason, the maxCacheAgeSeconds from PDP is not available, 10 minutes will be used.
	// Warning: the IAM team does not encourage the usage of this setting because it could have unintended impacts on authorization decisions.
	// If you think that you need this settings, it is highly suggested to discuss with the IAM team with your usecase.
	CacheDefaultTTL time.Duration

	// Default TTL for denied cache entries (2 minutes)
	CacheDefaultDeniedTTL time.Duration

	// The user provided implementation of the decision cache
	CachePlugin cache.DecisionCache

	// A configurable logger that the user can hook into their own logging implementation
	Logger Logger

	// The log level for the default logger, available levels are `LevelDebug`, `LevelInfo`, and `LevelError`. Default setting is info level
	LogLevel Level

	// The output destination for the default logger
	LogOutput io.Writer

	// The statistics aggregator
	Statistics *stats.Counter

	// HTTP client timeout for PDP calls. This configuration will be ignored when a custom HTTPClient is supplied.
	PDPTimeout time.Duration

	IsInitialized bool

	tokenManager *token.TokenManager

	/* A custom Http Client for users to provide their own implementation. Can be used to pass a custom Http client with Instana
	 * sensor embedded to the transport for tracing the service communications. Make sure to set a proper Timeout for the client.
	 */
	HTTPClient *http.Client

	// ServiceIdentifer
	// ClientID Client secret
	// Request Rate Limit

	// keyCacheExpiry int
}

var (
	pepConfigMutex = make(chan struct{}, 1)
	pepConfig      = &Config{}
)

// GetConfig returns a Config object containing the current configuration
func GetConfig() interface{} {
	pepConfigMutex <- struct{}{}
	c := *pepConfig
	<-pepConfigMutex
	return &c
}

// Configure initializes the current deployment configuration.
// Required fields are: Environment and APIKey
// If the environment is custom, the AuthzEndpoint, ListEndpoint and TokenEndpoint must be specified.
// Otherwise, they are filled out automatically.
func Configure(c *Config) error {
	if c.Environment != Staging && c.Environment != Production && c.Environment != PrivateStaging &&
		c.Environment != PrivateProduction && c.Environment != Custom {
		return fmt.Errorf("environment is not valid. Supported environments: %s, %s, %s, %s and %s", Staging.String(),
			Production.String(), PrivateStaging.String(), PrivateProduction.String(), Custom.String())
	}

	if c.Environment == Custom &&
		(c.AuthzEndpoint == "" || c.ListEndpoint == "" || c.TokenEndpoint == "" || c.KeyEndpoint == "") {
		return fmt.Errorf("AuthzEndpont, ListEndpoint, TokenEndpoint, and KeyEndpoint are required for Custom")
	}

	// If the cache is enabled but the user does not provide one, we create one
	if !c.DisableCache || !c.DisableDeniedCache {

		if c.CachePlugin == nil {
			cacheConfig := cache.DecisionCacheConfig{
				CacheSize:        c.DecisionCacheSize,
				TTL:              c.CacheDefaultTTL,
				DeniedTTL:        c.CacheDefaultDeniedTTL,
				DisableDenied:    c.DisableDeniedCache,
				DisablePermitted: c.DisableCache,
			}

			if cacheConfig.CacheSize < 1 {
				cacheConfig.CacheSize = 1
			}
			c.CachePlugin = cache.NewDecisionCache(&cacheConfig)
		}
	}
	// Configure statistics
	s := stats.NewStatsCounter()

	pepConfigMutex <- struct{}{}
	defer func() {
		<-pepConfigMutex
	}()

	pepConfig = &Config{
		Environment:           c.Environment,
		APIKey:                c.APIKey,
		ClientID:              c.ClientID,
		ClientSecret:          c.ClientSecret,
		AuthzEndpoint:         c.AuthzEndpoint,
		ListEndpoint:          c.ListEndpoint,
		TokenEndpoint:         c.TokenEndpoint,
		KeyEndpoint:           c.KeyEndpoint,
		DecisionCacheSize:     c.DecisionCacheSize,
		DisableCache:          c.DisableCache,
		DisableDeniedCache:    c.DisableDeniedCache,
		CacheDefaultTTL:       c.CacheDefaultTTL,
		CacheDefaultDeniedTTL: c.CacheDefaultDeniedTTL,
		CachePlugin:           c.CachePlugin,
		EnableExpiredCache:    c.EnableExpiredCache,
		Statistics:            s,
		AuthzRetry:            c.AuthzRetry,
	}

	if c.Logger == nil {
		var err error

		if c.LogOutput == nil {
			pepConfig.LogOutput = os.Stdout
		} else {
			pepConfig.LogOutput = c.LogOutput
		}

		if c.LogLevel == LevelNotSet {
			pepConfig.LogLevel = LevelInfo
		} else {
			pepConfig.LogLevel = c.LogLevel
		}

		pepConfig.Logger, err = NewOnePEPLogger(pepConfig.LogLevel, pepConfig.LogOutput)
		if err != nil {
			return err
		}
	} else {
		pepConfig.Logger = c.Logger
	}

	tokenConfig := &token.ExtendedConfig{
		ClientID:     pepConfig.ClientID,
		ClientSecret: pepConfig.ClientSecret,
		Scope:        pepConfig.TokenScope,
		Logger:       pepConfig.Logger,
	}

	if pepConfig.Environment == Staging {
		pepConfig.AuthzEndpoint = stagingAuthzEndpoint
		pepConfig.ListEndpoint = stagingListEndpoint
		pepConfig.TokenEndpoint = stagingTokenEndpoint
		pepConfig.KeyEndpoint = stagingKeyEndpoint
		pepConfig.RolesEndpoint = stagingBaseURL + rolesRoute
	} else if pepConfig.Environment == Production {
		pepConfig.AuthzEndpoint = prodAuthzEndpoint
		pepConfig.ListEndpoint = prodListEndpoint
		pepConfig.TokenEndpoint = prodTokenEndpoint
		pepConfig.KeyEndpoint = prodKeyEndpoint
		pepConfig.RolesEndpoint = prodBaseURL + rolesRoute
	} else if pepConfig.Environment == PrivateStaging {
		pepConfig.AuthzEndpoint = privateStagingAuthzEndpoint
		pepConfig.ListEndpoint = privateStagingListEndpoint
		pepConfig.TokenEndpoint = privateStagingTokenEndpoint
		pepConfig.KeyEndpoint = privateStagingKeyEndpoint
		pepConfig.RolesEndpoint = stagingBaseURL + rolesRoute
	} else if pepConfig.Environment == PrivateProduction {
		pepConfig.AuthzEndpoint = privateProdAuthzEndpoint
		pepConfig.ListEndpoint = privateProdListEndpoint
		pepConfig.TokenEndpoint = privateProdTokenEndpoint
		pepConfig.KeyEndpoint = privateProdKeyEndpoint
		pepConfig.RolesEndpoint = prodBaseURL + rolesRoute
	} else {
		tokenConfig.Endpoints = token.Endpoints{
			TokenEndpoint: c.TokenEndpoint,
			KeyEndpoint:   c.KeyEndpoint,
		}
	}

	if c.PDPTimeout == 0 {
		pepConfig.PDPTimeout = 15 * time.Second
	} else {
		pepConfig.PDPTimeout = c.PDPTimeout
	}

	if c.HTTPClient == nil {
		pepConfig.HTTPClient = &http.Client{
			Timeout: pepConfig.PDPTimeout,
		}
	} else {
		pepConfig.HTTPClient = c.HTTPClient
	}

	if c.APIKey != "" {
		tm, err := token.NewTokenManager(c.APIKey, pepConfig.Environment, tokenConfig)
		// Error handling
		if err != nil {
			return errors.Wrap(err, "Could not fetch token.")
		}

		pepConfig.tokenManager = tm

		pepConfig.IsInitialized = true
	}

	currentCacheKeyInfo.storeCacheKeyPattern(CacheKeyPattern{})

	pepConfig.Logger.Info("DeploymentEnvironment: ", pepConfig.Environment.String(), "\nAuthzEndpoint: ", pepConfig.AuthzEndpoint, "\nListEndpoint: ", pepConfig.ListEndpoint, "\nTokenEndpoint: ", pepConfig.TokenEndpoint, "\nKeyEndpoint: ", pepConfig.KeyEndpoint, "\nRolesEndpoint: ", pepConfig.RolesEndpoint, "\nDecisionCacheSize: ", pepConfig.DecisionCacheSize, "\nDisableCache: ", pepConfig.DisableCache, "\nDisableDeniedCache: ", pepConfig.DisableDeniedCache, "\nEnableExpiredCache: ", pepConfig.EnableExpiredCache, "\nCacheDefaultTTL: ", pepConfig.CacheDefaultTTL, "\nCacheDefaultDeniedTTL: ", pepConfig.CacheDefaultDeniedTTL, "\nLogLevel: ", pepConfig.LogLevel, "\nPDPTimeout: ", pepConfig.PDPTimeout, "\nAuthzRetry: ", pepConfig.AuthzRetry)

	return nil
}

// Usage representing the usage stats
type Usage struct {
	// RequestsServicedByPEP is the number of requests serviced by the PEP
	RequestsServicedByPEP int

	// OriginalRequestsToPDP is the number of requests sent to the PDP
	OriginalRequestsToPDP int

	// FailedUserRequests is the number of requests that failed
	FailedUserRequests int

	// StatusCodeCounts is the list of all status codes returned by PDP and respective count
	StatusCodeCounts map[string]int
}

// Stats representing usage and cache stats
type Stats struct {

	// Usage is the statistics that is tracked by the PEP
	Usage Usage

	// Cache is the cache statistics tracked by internal caching library
	Cache cache.Stats
}

// ReportStatistics returns the available runtime statistics of the PEP
func (c *Config) reportStatisticsStats() Stats {

	if !c.IsInitialized || c.Statistics == nil {
		return Stats{}
	}

	usageStatsMap := c.Statistics.GetStatsAsMap()

	UsageStat := Usage{RequestsServicedByPEP: usageStatsMap[requestsServicedByPEP],
		OriginalRequestsToPDP: usageStatsMap[originalRequestsToPDP],
		FailedUserRequests:    usageStatsMap[failedUserRequests],
		StatusCodeCounts:      make(map[string]int)}

	// Iterate the usageStatsMap for status codes and store them in UsageStat
	for key, count := range usageStatsMap {
		if strings.Contains(key, "status-code") {
			UsageStat.StatusCodeCounts[key] = count
		}
	}

	statsToReport := Stats{}
	statsToReport.Usage = UsageStat
	if c.CachePlugin != nil {
		statsToReport.Cache = c.CachePlugin.GetStatistics()
	}
	return statsToReport
}
