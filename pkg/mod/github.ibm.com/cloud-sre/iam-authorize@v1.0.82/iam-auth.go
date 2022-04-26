package iamauth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	instana "github.com/instana/go-sensor"
	newrelicPreModules "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/xid"
	"github.ibm.com/IAM/pep/v3"
	"github.ibm.com/cloud-sre/iam-authorize/bgcache"
	"github.ibm.com/cloud-sre/iam-authorize/iam"
	"github.ibm.com/cloud-sre/iam-authorize/monitoring"
	"github.ibm.com/cloud-sre/iam-authorize/plugins"
	"github.ibm.com/cloud-sre/iam-authorize/utils"
)

type tokenStruct struct {
	AccessToken string `json:"access_token"`
}

type decodedToken struct {
	IAMID   string `json:"iam_id"`
	Account struct {
		BSS string `json:"bss"`
	} `json:"account"`
	Email      string `json:"email"`
	Expiration int64  `json:"exp"`
	Sub        string `json:"sub"`
}

// cache is used as the caching struct that holds api keys and tokens
type cache struct {
	sync.RWMutex
	apikCache       map[string]apikCache
	maxSize         int
	sizeThreshold   int
	cleanupInterval time.Duration
}

// each api key in the cache has a token, its expiration date in unix time and the source of authorization
type apikCache struct {
	token    string
	tokenExp int64
	source   string
}

// IAMAuth is the base type that implements all public functions in this lib
type IAMAuth struct {
	iamURL                  string
	serviceAPIKey           string
	serviceName             string
	iamResourceActionPlugin plugins.IAMResourceActionPlugin
	resourceType            string
	cache                   cache
	iamHealthCheck          *iam.HealthCheck
	monConfig               *iam.NRWrapperConfig
	enableAutoIAMBreakGlass bool
	mode                    string
}

// Authorization struct used for returning authorization values of a request containing an API key/token
type Authorization struct {
	Email  string
	IamID  string
	Source string
}

// NewIAMAuthConfig is used to provide configuration data when creating a new IAM auth instance
type NewIAMAuthConfig struct {
	SvcName                 string               // Service name used to build the CRN when checking for the access in IAM
	IAMURL                  string               // Used by IAM monitor to test if IAM is available
	IAMAPIKey               string               // Used by IAM monitor to test if IAM is available
	NRConfig                *iam.NRWrapperConfig // Used by IAM monitor to send metrics to NewRelic (can be nil if not required)
	EnableAutoIAMBreakGlass bool                 // Whether break glass should be enabled when IAM monitoring detects IAM is down
}

// iamError is a custom error that indicates that there is something potentially wrong with the health of IAM
type iamError struct {
	rootError error
}

func (err iamError) Error() string {
	if err.rootError != nil {
		return err.rootError.Error()
	}
	return ""
}

const (
	// defaultServiceName is the name of the service to make requests against IAM.
	defaultServiceName string = "pnp-api-oss"
	// defaultIAMURL will be used to call IAM if no IAM URL has been set
	defaultIAMURL string = "https://iam.test.cloud.ibm.com"
	// productionIAMURL defines the IBM Cloud production IAM URL
	productionIAMURL string = "https://iam.cloud.ibm.com"
	// privateStagingIAMURL is the private staging IAM URL
	privateStagingIAMURL = "https://private.iam.test.cloud.ibm.com"
	// privateProdIAMURL is the private production IAM URL
	privateProdIAMURL = "https://private.iam.cloud.ibm.com"
	// cacheDefaultSizeThreshold is the threshold in percentage of when the cache should be cleaned up(remove expired tokens)
	cacheDefaultSizeThreshold int = 80
	// cacheDefaultMaxSize is the maximum number of entries allowed in the cache
	cacheDefaultMaxSize int = 10000
	// cacheDefaultCleanupInterval is the default cache clean up interval
	cacheDefaultCleanupInterval time.Duration = 5 * time.Minute
	// defaultResourceType is the resource type that is set in the IAM policy
	defaultResourceType string = "endpoint"
	// AuthSourcePublic defines the authorization source for Public IAM
	AuthSourcePublic string = "public-iam"
	// Audit holds the log prefix value for audit logs
	Audit = "AUDIT: "
	// AuthorizeAccountOnly is used to check if the incoming request for authorization is only to ensure the account is
	// an IBM cloud account.
	AuthorizeAccountOnly = "oss-authorize-account-only"
)

// BypassFlag allow anyone access to any resource when bypass is true
var BypassFlag = false

// IsIAMAuthorized checks auth of the request.
// it returns the email address of the user only if the user is authorized, or an error if not.
func (iam *IAMAuth) IsIAMAuthorized(req *http.Request, res http.ResponseWriter) (*Authorization, error) {
	const fct = "IsIAMAuthorized: "
	const isIAMAuthorized = "iam-authorize-IsIAMAuthorized"

	log.Println(Audit+"Request source IP address:", getReqSourceIP(req))

	// Instana
	var sensor *instana.Sensor
	if iam.monConfig != nil && iam.monConfig.InstanaSensor != nil {
		sensor = iam.monConfig.InstanaSensor
	}
	span, ctx := monitoring.NewSpan(context.Background(), sensor, isIAMAuthorized)
	if span == nil {
		log.Println(fct, "couldn't retrieve or create a span")
	} else {
		defer span.Finish()
	}

	// New Relic
	txnPreModules, ok := res.(newrelicPreModules.Transaction)
	if !ok {
		log.Println(fct, "couldn't convert response to type newrelicPreModules.Transaction")
	}

	var txn *newrelic.Transaction
	if iam.monConfig != nil {
		txn = iam.monConfig.NRAppV3.StartTransaction(isIAMAuthorized)
		txn.SetWebRequestHTTP(req)
		txn.SetWebResponse(res)
	}

	// Add New Relic transaction to context in order to retrieve it from ctx
	ctx = newrelicPreModules.NewContext(ctx, txnPreModules)
	ctx = newrelic.NewContext(ctx, txn)

	if iam.monConfig != nil {
		monitoring.SetTagsKV(ctx,
			"environment", iam.monConfig.Environment,
			"region", iam.monConfig.Region,
			"sourceServiceName", iam.monConfig.SourceServiceName,
			"url", iam.monConfig.URL)
	}

	if iam.enableAutoIAMBreakGlass == true {
		monitoring.SetTag(ctx, "iamBreakGlassMode", "enabled")
	} else {
		monitoring.SetTag(ctx, "iamBreakGlassMode", "disabled")
	}

	var resource, permission string
	if iam.iamResourceActionPlugin == nil {
		// No plugin, use the default code:
		resource, permission = utils.GetIAMResourceAndPermission(req)
	} else {
		// Plugin, use it instead of the default code:
		resource = iam.iamResourceActionPlugin.GetIAMResourceFromRequest(req)
		permission = iam.iamResourceActionPlugin.GetIAMActionFromRequest(req)
	}

	authHeader := req.Header.Get("Authorization")
	if utils.IsBadAuth(authHeader) {
		return nil, errors.New(fct + "Bad authorization key found in cache")
	}

	var auth *Authorization
	var token string
	var err error

	// Check whether we can go through the normal flow and call IAM:
	isIAMHealthy := iam.iamHealthCheck != nil && (*iam.iamHealthCheck).IsHealthy()
	if !iam.enableAutoIAMBreakGlass || (iam.enableAutoIAMBreakGlass && isIAMHealthy) {
		// Break glass is disabled, or break glass is enabled and IAM check was OK. Use normal flow and call IAM:
		monitoring.SetTag(ctx, "iamBreakGlass", false)

		username, password, ok := req.BasicAuth()
		if ok && username == "apikey" {
			authHeader = password
		}

		// EDB workaround, should be better served by adding a resource for reading info or healthz endpoints.
		// Ideally this should be a resource for all APIs that need to protect their informational or health
		// endpoints or those endpoints which do not have a resource and only need an account authorization check.
		if permission == AuthorizeAccountOnly {
			auth, token, err = iam.tokenAuthForAccount(authHeader, resource, ctx)
		} else {
			auth, token, err = iam.tokenAuth(authHeader, resource, permission, ctx)
		}

		if err == nil && auth == nil {
			err = errors.New(fct + "unauthorized")
		}

		statusMsg := "authorization successful"
		if err != nil {
			statusMsg = "unauthorized"
		}
		log.Print(Audit, fct, "\n\tpath: ", req.URL.Path, "\n\tmethod: ", req.Method, "\n\tresource: ", resource,
			"\n\tpermission: ", permission, "\n\tauth: ", auth, "\n\tStatus: ", statusMsg)

		returnAfterBypassCheck := false
		if err == nil && auth != nil {
			addDataToRequestContext(req, "iamToken", token) // Add the token in the request context
			addDataToRequestContext(req, "iamTokenEmail", auth.Email)
			addDataToRequestContext(req, "iamID", auth.IamID)
		} else if err != nil {
			_, ok := err.(*iamError)
			if ok && iam.enableAutoIAMBreakGlass && iam.iamHealthCheck != nil && authHeader != "" {
				log.Printf("INFO: The following error indicates that there is a problem with IAM health, so an IAM health check will be performed. Err = %v", err)
				isIAMHealthy = isIAMHealthy && (*iam.iamHealthCheck).CheckHealthAndTrackRecovery()
			} else {
				log.Printf("INFO: The following error indicates there is no problem with IAM health, so an IAM health check will not be performed even if breakglass is enabled. Err = %v", err)
				returnAfterBypassCheck = true // return right after checking the bypass as there is nothing else that can be done
			}
		}

		if BypassFlag {
			// START OVERRIDE TO ENSURE WE ALWAYS AUTHORIZE
			if err != nil {
				log.Println(Audit+fct, "Overriding error", err)
			}
			err = nil
			if auth == nil {
				auth = &Authorization{Email: "code.override@ibm.com", IamID: "code.override", Source: "code.override"}
			}
			// END   OVERRIDE TO ENSURE WE ALWAYS AUTHORIZE
		} else if returnAfterBypassCheck {
			return nil, err // nothing more that can be done, just return error
		}
	}

	if iam.enableAutoIAMBreakGlass == true && authHeader != "" && isAPIkey(authHeader) {
		bgc := bgcache.GetBGCache()

		isAuthorized := auth != nil && err == nil
		if isAuthorized && auth.Source == AuthSourcePublic {
			go bgc.AddAuthorization(authHeader, auth.IamID, auth.Email, auth.Source, token, resource, permission, time.Now().Unix(), false)
		} else if !isIAMHealthy && !isAuthorized { // Get the authorization from break glass cache (if any)
			monitoring.SetTagsKV(ctx,
				"iamAuthResource", resource,
				"iamAuthPermission", permission)

			bga := bgc.GetAuthorization(authHeader, resource, permission)
			if bga != nil {
				err = nil
				auth = &Authorization{Email: bga.User, IamID: bga.IamID, Source: bga.Source}
				addDataToRequestContext(req, "iamToken", bga.Token)
				addDataToRequestContext(req, "iamTokenEmail", auth.Email)
				addDataToRequestContext(req, "iamID", auth.IamID)

				monitoring.SetTagsKV(ctx,
					"iamBreakGlass", true,
					"iamAuthEmail", auth.Email,
					"iamAuthIamID", auth.IamID,
					"iamAuthSource", auth.Source)
			} else {
				monitoring.SetTag(ctx, "iamBreakGlassFail", true)
			}
			log.Print(Audit+fct+"Get authorization from break glass when isAPIkey\n\tauth: ", auth, "\n\tError: ", err)
		}
	} else if iam.enableAutoIAMBreakGlass == true && authHeader != "" && !isAPIkey(authHeader) {
		bgc := bgcache.GetBGCache()

		isAuthorized := auth != nil && err == nil
		if isAuthorized && auth.Source == AuthSourcePublic {
			go bgc.AddUser(auth.IamID, auth.Email, auth.Source, resource, permission, time.Now().Unix(), false)
		} else if !isIAMHealthy && !isAuthorized { // Check break glass cache to see if we have seen this user call before
			monitoring.SetTagsKV(ctx,
				"iamAuthResource", resource,
				"iamAuthPermission", permission)

			tokenFromAuthHeader := getToken(authHeader)
			tokenDecoded, err := decodeToken(tokenFromAuthHeader)
			if err == nil && bgc.IsUserAuthorized(tokenDecoded.IAMID, AuthSourcePublic, resource, permission) {
				err = nil
				auth = &Authorization{Email: tokenDecoded.Email, IamID: tokenDecoded.IAMID, Source: AuthSourcePublic}
				addDataToRequestContext(req, "iamToken", tokenFromAuthHeader)
				addDataToRequestContext(req, "iamTokenEmail", auth.Email)
				addDataToRequestContext(req, "iamID", auth.IamID)

				monitoring.SetTagsKV(ctx,
					"iamBreakGlass", true,
					"iamAuthEmail", auth.Email,
					"iamAuthIamID", auth.IamID,
					"iamAuthSource", auth.Source)
			} else {
				monitoring.SetTag(ctx, "iamBreakGlassFail", true)
			}
			log.Print(Audit+fct+"Get authorization from break glass when !isAPIkey\n\tauth: ", auth, "\n\tError: ", err)
		}
	}

	return auth, err
}

// getToken removes "Bearer " from the provided authorization header value
func getToken(authHeaderVal string) string {
	token := ""
	authHeaderVal = strings.TrimSpace(authHeaderVal)
	tokenSplit := strings.Split(authHeaderVal, " ")
	if len(tokenSplit) > 1 {
		token = tokenSplit[len(tokenSplit)-1]
	}
	return token
}

// GetAccountID expects an IAM access token, and returns the associated IBM Cloud accountID and an error.
func (iam *IAMAuth) GetAccountID(token string) (string, error) {
	return getAccountID(token)
}

// GetEmail gets the email associated with the API key/token in the Authorization header
func (iam *IAMAuth) GetEmail(req *http.Request) (string, error) {
	return iam.getEmail(req.Header.Get("Authorization"))
}

// GetTokenScope gets the scope of the associated API key/token in the Authorization header
func (iam *IAMAuth) GetTokenScope(req *http.Request) (string, error) {
	const fct = "GetTokenScope: "
	token, ok := req.Context().Value(ContextKeyString("iamToken")).(string)
	if !ok {
		log.Print(fct + "Unable to get token from request context. Calling getTokenFromRequest...")

		var err error
		token, err = iam.getTokenFromRequest(req.Header.Get("Authorization"), nil)
		if err != nil {
			return "", err
		}
	}

	claims, err := pep.GetClaims(token, iam.getTokenValidationMode())
	if err != nil {
		return "", err
	}

	return claims.Scope, nil
}

// getTokenValidationMode returns whether token validation should be performed or not
func (iam *IAMAuth) getTokenValidationMode() bool {
	// validation enabled by default
	var skipValidation = false
	if iam.mode == "test" {
		skipValidation = true
	}
	return skipValidation
}

// GetAccessToken expects an IBM Cloud API Key, and returns an IAM access token and an error.
func (iam *IAMAuth) GetAccessToken(apiKey string) (string, error) {

	// check api key cache before generating a new token
	if token := iam.checkAPIKeyCache(apiKey); token != "" {
		return token, nil
	}

	return iam.getToken(apiKey)
}

// SetServiceAPIKey sets the Cloud Service API Key that is used to check whether an account has the right access.
// Optionaly you can set env variable CLOUD_SERVICE_API_KEY to set the key.
func (iam *IAMAuth) SetServiceAPIKey(apikey string) { // TODO: will be removed in the future
	if apikey != "" {
		iam.serviceAPIKey = apikey
		iam.configPep()
	}
}

// SetIAMURL sets a the Public IAM URL. The default is the IAM staging URL.
func (iam *IAMAuth) SetIAMURL(iamurl string) {
	if iamurl != "" {
		iam.iamURL = iamurl
		// iam.configPep() // no use case found that must reconfigure IAM config when URL is updated. This causes problem with edb-mapping-consumer.
		if iam.iamHealthCheck != nil {
			(*iam.iamHealthCheck).UpdateIAMURL(iam.iamURL)
		}
	}
}

// SetCRNServiceName sets a service name that will be needed to build the CRN
// when checking for the access in IAM.
//
// The default value is "pnp-api-oss"
func (iam *IAMAuth) SetCRNServiceName(svcName string) {
	if svcName != "" {
		iam.serviceName = svcName
	}

	if iam.serviceName == "" {
		iam.serviceName = defaultServiceName
	}
}

// SetResourceTypeName sets the resourceType value that will be needed to build the CRN
// when checking for the access in IAM.
//
// The default value is "endpoint"
func (iam *IAMAuth) SetResourceTypeName(resourceType string) {
	iam.resourceType = resourceType
}

// SetCacheMaxSize sets the maximum cache size in number of records in the cache.
//
// The default value is 10k records. The minimum is 100 records.
func (iam *IAMAuth) SetCacheMaxSize(size int) {
	if size >= 100 {
		iam.cache.maxSize = size
	}
}

// SetCacheSizeThreshold sets the number(in percentage) of the cache size threshold.
// When the threshold has been reached then all expired tokens will be removed from the cache.
//
// The default value is 80. The minimum is 10 and max is 99.
func (iam *IAMAuth) SetCacheSizeThreshold(threshold int) {
	if threshold >= 10 && threshold < 100 {
		iam.cache.sizeThreshold = threshold
	}
}

// SetCacheCleanupInterval sets how often the cache is checked to be cleaned up(remove expired tokens).
//
// The default value is 5 minutes. The minimum value is 10 seconds.
func (iam *IAMAuth) SetCacheCleanupInterval(interval time.Duration) {
	if interval >= (10 * time.Second) {
		iam.cache.cleanupInterval = interval
	}
}

// SetIAMResourceActionPlugin sets the plugin interface used to obtain IAM resources and
// actions from requests. If set, this plugin overrides the default built-in code.
func (iam *IAMAuth) SetIAMResourceActionPlugin(plugin plugins.IAMResourceActionPlugin) {
	iam.iamResourceActionPlugin = plugin
}

// SetNRWrapperConfig sets the monitoring wrapper config used for IAM monitoring.
func (iam *IAMAuth) SetNRWrapperConfig(monConfig *iam.NRWrapperConfig) {
	iam.monConfig = monConfig
}

// NewIAMAuthWithIAMMonitor returns the IAMAuth struct that holds all functions to interact with this lib.
// This method creates an IAMAuth instance that automatically monitors IAM availability in the background,
// can optionally notify NewRelic of IAM availability, and can optionally use IAM break glass support in
// the event that IAM is not available.
func NewIAMAuthWithIAMMonitor(newIAMAuthConfig NewIAMAuthConfig) *IAMAuth {
	iamAuth := NewIAMAuth(newIAMAuthConfig.SvcName)
	iamAuth.SetIAMURL(newIAMAuthConfig.IAMURL)
	iamHealthCheck := iam.NewDefaultHealthCheck(newIAMAuthConfig.IAMURL, newIAMAuthConfig.IAMAPIKey, newIAMAuthConfig.NRConfig)
	iamAuth.iamHealthCheck = &iamHealthCheck
	iamAuth.enableAutoIAMBreakGlass = newIAMAuthConfig.EnableAutoIAMBreakGlass
	iamAuth.monConfig = newIAMAuthConfig.NRConfig

	// Initialize break glass cache here
	if iamAuth.enableAutoIAMBreakGlass == true {
		if newIAMAuthConfig.NRConfig != nil {
			newIAMAuthConfig.NRConfig.URL = newIAMAuthConfig.IAMURL
		}
		bgcache.InitBgCache(newIAMAuthConfig.NRConfig)
	}

	return iamAuth
}

// NewIAMAuth returns the IAMAuth struct that holds all functions to interact with this lib
func NewIAMAuth(svcName string) *IAMAuth {

	const fct = "NewIAMAuth: "

	iamauth := &IAMAuth{
		serviceName: (func() string {
			if svcName != "" {
				return svcName
			}
			return defaultServiceName
		})(),
		iamURL: (func() string {
			if envIAMURL := os.Getenv("IAM_URL"); envIAMURL != "" {
				return envIAMURL
			}
			return defaultIAMURL
		})(),
		resourceType: defaultResourceType,
		cache: cache{
			apikCache:       make(map[string]apikCache, 1),
			maxSize:         cacheDefaultMaxSize,
			sizeThreshold:   cacheDefaultSizeThreshold,
			cleanupInterval: cacheDefaultCleanupInterval,
		},
		mode: os.Getenv("IAM_MODE"),
	}

	// configure pep
	iamauth.configPep()

	// start cache cleanup job
	iamauth.cleanUpAPIKeyCache()

	// check whether the bypass flag is set
	if flag := os.Getenv("IAM_BYPASS_FLAG"); flag == "true" {
		log.Println(fct + "BYPASS FLAG is active. No resource authorization is performed under this mode.")
		BypassFlag = true
	}

	return iamauth
}

// configPep configures PEP lib with the desired IAM configuration
func (iam *IAMAuth) configPep() {

	var env pep.DeploymentEnvironment
	if strings.HasPrefix(iam.iamURL, defaultIAMURL) {
		env = pep.Staging
	} else if strings.HasPrefix(iam.iamURL, productionIAMURL) {
		env = pep.Production
	} else if strings.HasPrefix(iam.iamURL, privateStagingIAMURL) {
		env = pep.PrivateStaging
	} else if strings.HasPrefix(iam.iamURL, privateProdIAMURL) {
		env = pep.PrivateProduction
	} else {
		env = pep.Custom
	}

	var apiKey = os.Getenv("CLOUD_SERVICE_API_KEY")
	if apiKey == "" {
		log.Println("env variable CLOUD_SERVICE_API_KEY is empty, it must be set for this lib to work with IAM")
	}
	if iam.serviceAPIKey != "" { // TODO: will be removed in the future
		apiKey = iam.serviceAPIKey
	}

	// must have an API Key set to configure pep, or else it exists the application.
	// this is needed specifically for test cases as it's not set and they end up failing.
	// for running applications this should never occur.
	if apiKey == "" {
		apiKey = "placeholder-api-key"
	}

	config := &pep.Config{
		Environment:        env,
		APIKey:             apiKey,
		LogLevel:           pep.LevelError,
		AuthzRetry:         true,
		EnableExpiredCache: true,
	}

	if env == pep.Custom {
		config.AuthzEndpoint = iam.iamURL + "/v2/authz"
		config.ListEndpoint = iam.iamURL + "/v2/authz/bulk"
		config.TokenEndpoint = iam.iamURL + "/identity/token"
		config.KeyEndpoint = iam.iamURL + "/identity/keys"
	}

	err := pep.Configure(config)
	if err != nil {
		log.Fatalln("could not configure pep iam lib, error=", err)
	}

}

func isAPIkey(authKey string) bool {
	return !strings.HasPrefix(strings.ToLower(authKey), "bearer")
}

// ctx is only used during the auth.  Otherwise it should be nil.
func (iam *IAMAuth) getTokenFromRequest(authHeaderVal string, ctx context.Context) (token string, err error) {

	if !isAPIkey(authHeaderVal) {
		// it is a token because it is prefixed by "Bearer" and it's considered to be a Public IAM token
		authHeaderVal = strings.TrimSpace(authHeaderVal)
		tokenSplit := strings.Split(authHeaderVal, " ")
		if len(tokenSplit) > 1 {
			token = tokenSplit[len(tokenSplit)-1]
			// verify Public IAM token is valid using the public key
			_, err = pep.GetClaims(token, iam.getTokenValidationMode())
			if err != nil {
				// IAM caches public keys and attempts to update the public keys by calling IAM every 60 minutes by default. If
				// there is a problem the cached public keys are used, so we can assume any error from the call to GetClaims is
				// most likely a bad token and not a situation where IAM is down. So set the token to an empty string and fall
				// down to the "unauthorized" return below:
				// log.Println(err)
				token = ""
			}
		}
	} else {
		// not a token, consider it is an API Key
		token, err = iam.GetAccessToken(authHeaderVal)
		if err != nil {
			// log.Println(err)
			return
		}

		// Only needed during auth. It's the only time we will pass the ctx in.
		if ctx != nil {
			// get email from token
			log.Print("Getting email from token")
			tokenDecoded, err := decodeToken(token)
			if err != nil {
				// This indicates that IAM has given us a bad IAM token from the IAM API key, so something is potentially wrong with IAM:
				return "", &iamError{rootError: err}
			}

			if !isTokenValid(tokenDecoded.Expiration) {
				// This indicates that IAM has given us a bad IAM token from the IAM API key, so something is potentially wrong with IAM:
				return "", &iamError{rootError: errors.New("token is expired")}
			}
			log.Print("email from token: ", tokenDecoded.Email)
			monitoring.SetTag(ctx, "iamPrivateUser", tokenDecoded.Email)
		}
	}

	if token != "" {
		return
	}

	// Return a normal error instead of an iamError because the user is unauthorized:
	return "", errors.New("unauthorized")
}

func (iam *IAMAuth) getAPIKeySource(apik string) string {
	iam.cache.RLock()
	defer iam.cache.RUnlock()

	if token := iam.cache.apikCache[apik]; token != (apikCache{}) && isTokenValid(token.tokenExp) {
		// log.Println("valid token found in cache")
		return token.source
	}

	return ""
}

// tokenAuth checks authorization given a user's IAM token or IBM Cloud API Key
// or a resource name and an action/permission name
func (iam *IAMAuth) tokenAuth(authHeaderVal string, resource, permission string, ctx context.Context) (*Authorization, string, error) {

	const fct = "tokenAuth: "
	if authHeaderVal == "" {
		// Return a normal error instead of an iamError because the error is not due to IAM being unavailable:
		return nil, ``, errors.New("no access token or api key provided")
	}

	if permission == "" {
		// Return a normal error instead of an iamError because the error is not due to IAM being unavailable:
		return nil, ``, errors.New("no permission name provided")
	}

	if resource == "" {
		// Return a normal error instead of an iamError because the error is not due to IAM being unavailable:
		return nil, ``, errors.New("no resource name provided")
	}

	accessToken, err := iam.getTokenFromRequest(authHeaderVal, ctx)
	if err != nil {
		log.Println(err)
		utils.AddBadAuth(authHeaderVal)
		return nil, ``, err
	}

	log.Printf("Checking permission=[%s] for resource=[%s] and resource type=[%s]\n", permission, resource, iam.resourceType)

	// if the bypass flag is set to true, then return authorized for any valid IBM Coud API key or token
	if BypassFlag {
		log.Println(fct + "BYPASS FLAG is true. No resource authorization is being performed.")
		_, err := iam.authorized(accessToken, AuthSourcePublic, permission, resource, ctx)
		if err != nil {
			log.Println(fct+"Error during authorization", err)
			return &Authorization{Email: "bypass_user@test.com", IamID: "testID", Source: "testSource"}, accessToken, nil
		}
		return nil, ``, err
	}

	// ensure Cloud Service API Key is set, if not return error
	err = iam.checkServiceAPIKey() // TODO: will be removed in the future
	if err != nil {
		// Should never be an error here because the pep library would never initialize without a service API key.
		// Return a normal error instead of an iamError:
		return nil, ``, err
	}

	iamID, err := pep.GetSubjectAsIAMIDClaim(accessToken, false)
	if err != nil {
		// IAM caches public keys and attempts to update the public keys by calling IAM every 60 minutes by default. If
		// there is a problem the cached public keys are used, so we can assume any error from the call to GetClaims is
		// most likely a bad token and not a situation where IAM is down, so return a normal error instead of a iamError:
		return nil, ``, err
	}

	subject := pep.Attributes{
		"id": iamID,
	}

	svcToken, err := pep.GetToken()
	if err != nil {
		// An error indicates a problem with IAM not renewing the service token, so a potential problem with IAM:
		return nil, ``, &iamError{rootError: err}
	}

	accID, err := iam.GetAccountID(svcToken)
	if err != nil {
		// An error here indicates a problem with the IAM service token that is auto renewed by IAM the pep library. The problem
		// could be structural or that the token is expired - both indicate a potential problem with IAM:
		return nil, "", &iamError{rootError: err}
	}

	resources := pep.Attributes{
		"serviceName":  iam.serviceName,
		"accountId":    accID,
		"resourceType": iam.resourceType,
		"resource":     resource,
	}

	requests := pep.Requests{
		{
			"action":   permission,
			"resource": resources,
			"subject":  subject,
		},
	}

	log.Printf("Checking IAM authorization. URL=%s\n", iam.iamURL+"/v2/authz/bulk")

	traceID := xid.New().String()

	tTraceStart := time.Now()

	response, err := pep.PerformAuthorization(&requests, traceID)
	if err != nil {
		log.Printf("ERROR: Response: %+v", response)
		// Note that if a user is unauthorized, an error is NOT returned so we should check the IAM health:
		return nil, "", &iamError{rootError: err}
	}

	tTraceEnd := time.Now()
	log.Printf("Request time: %s", tTraceEnd.Sub(tTraceStart).String())

	log.Printf("Transaction-Id: %s\n", response.Trace)
	log.Printf("Allowed: %t", response.Decisions[0].Permitted)

	if response.Decisions[0].Permitted {
		auth, err := iam.authorized(accessToken, AuthSourcePublic, permission, resource, ctx)
		// Error should be nil here because the token has already been validated at this point. An error would indicate that
		// the token is malformed so return a normal error instead of an iamError:
		return auth, accessToken, err
	}

	// send authorized result to NR
	iam.addIAMAuthNRAttributes(ctx, accessToken, permission, resource, response.Decisions[0].Permitted)

	// Return a normal error instead of an iamError because the user is unauthorized:
	return nil, ``, errors.New("unauthorized")
}

// Gabriel Avila 2020-11-12
// Change prompted by EDB case
// tokenAuthForAccount checks whether the incoming authorization token is part of an account with access to the endpoint
// for which we seek authorization.
func (iam *IAMAuth) tokenAuthForAccount(authHeaderVal, resource string, ctx context.Context) (*Authorization, string, error) {

	const fct = "tokenAuthForAccount: "
	var auth *Authorization

	accessToken, err := iam.getTokenFromRequest(authHeaderVal, ctx)
	if err != nil {
		utils.AddBadAuth(authHeaderVal)
		return nil, ``, err
	}

	log.Printf("Checking account permission for resource type=[%s]\n", iam.resourceType)

	// return authorized for any valid IBM Cloud API key or token
	auth, err = iam.authorizedForAccount(accessToken, resource, AuthSourcePublic, ctx)
	if err != nil {
		log.Println(fct+"Error during authorization", err)
		return nil, accessToken, err
	}

	return auth, accessToken, err
}

func (iam *IAMAuth) authorized(token, source, permission, resource string, ctx context.Context) (*Authorization, error) {
	tokenDecoded, err := decodeToken(token)
	if err != nil {
		// send authorized result to NR
		iam.addIAMAuthNRAttributes(ctx, token, permission, resource, false)

		return nil, err
	}

	// send authorized result to NR
	iam.addIAMAuthNRAttributes(ctx, token, permission, resource, true)

	return &Authorization{Email: tokenDecoded.Email, IamID: tokenDecoded.IAMID, Source: source}, nil
}

// Gabriel Avila 2020-11-12
// Change for EDB case where resources need not be checked for informational endpoints such as healthz or info.
// authorizedForAccount checks the supplied token and a source to authenticate an account against IAM.
// this is used to protect informational or healthz endpoints.
func (iam *IAMAuth) authorizedForAccount(token, resource, source string, ctx context.Context) (*Authorization, error) {
	ibmCloudAccountID, err := getAccountID(token)
	if err != nil {
		return nil, err
	}

	if ibmCloudAccountID != resource {
		iam.addIAMAuthNRAttributes(ctx, token, source, AuthorizeAccountOnly, false)
		// Account id from the token does not match the desired resource, so the token is unauthorized. Return
		// a normal error instead of an iamError because the error is not due to IAM being unavailable:
		return nil, errors.New("not authorized for account")
	}

	tokenDecoded, err := decodeToken(token)
	if err != nil {
		iam.addIAMAuthNRAttributes(ctx, token, source, AuthorizeAccountOnly, false)
		return nil, err
	}

	return &Authorization{Email: tokenDecoded.Email, IamID: tokenDecoded.IAMID, Source: source}, nil
}

// getToken returns an access token generated with the provided api key, and an error
func (iam *IAMAuth) getToken(apiKey string) (string, error) {

	if apiKey == "" {
		// Return a normal error instead of an iamError because the error is not due to IAM being unavailable:
		return "", errors.New("empty api key provided in function call")
	}

	postData := fmt.Sprintf("grant_type=urn:ibm:params:oauth:grant-type:apikey&apikey=%s", apiKey)

	// Public IAM
	pub := iam.iamURL + "/identity/token"

	log.Println("Trying Public IAM")
	// try public IAM if priv failed

	rsp, err := iam.retrieveIAMToken(pub, postData)
	if err != nil {
		return "", err
	}

	defer closeBody(rsp.Body.Close)

	jsonBytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", &iamError{rootError: err}
	}

	// Decode the json result into a struct
	ts := new(tokenStruct)
	if err := json.NewDecoder(bytes.NewReader(jsonBytes)).Decode(ts); err != nil {
		return "", &iamError{rootError: err}
	}

	// add apik and token to the cache
	iam.addAPIKeyToCache(apiKey, ts.AccessToken, AuthSourcePublic)

	return ts.AccessToken, nil
}

// retrieveIAMToken calls IAM URL to retrieve an access token
func (iam *IAMAuth) retrieveIAMToken(url, data string) (*http.Response, error) {

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	if err != nil {
		// An error here indicates a problem in our code. Return a normal error instead of an iamError
		// because the error is not due to IAM being unavailable:
		return nil, err
	}

	req.SetBasicAuth("bx", "bx")

	client := &http.Client{Timeout: 30 * time.Second}

	log.Printf("Getting IAM token. URL=%s\n", url)

	tTraceStart := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return nil, &iamError{rootError: err}
	}

	tTraceEnd := time.Now()
	log.Printf("Request time: %s", tTraceEnd.Sub(tTraceStart).String())

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// In general, check the IAM health for non-200 response codes. This can be improved as needed over time:
		potentialProblemWithIAM := true
		// Read the response message and see if an IAM health check can be skipped:
		defer closeBody(resp.Body.Close)
		data, err := ioutil.ReadAll(resp.Body)
		if err == nil &&
			((resp.StatusCode == http.StatusBadRequest && strings.Contains(string(data), "BXNIM0410E")) || // Provided user not found or active
				(resp.StatusCode == http.StatusBadRequest && strings.Contains(string(data), "BXNIM0415E")) || // Provided API key could not be found
				(resp.StatusCode == http.StatusUnauthorized && strings.Contains(string(data), "BXNIM0436E"))) { // User is suspended in the requested account or invalid.
			// IAM health check is not needed (note that the above IAM error codes are documented at
			// https://github.ibm.com/BlueMix-Fabric/CloudIAM-APIKeys/blob/integration/feature/core/src/main/i18n/exceptions.properties
			// and per the IAM team the error codes can be relied upon as the codes do not change unless there is a bug):
			potentialProblemWithIAM = false
		}
		if err == nil {
			err = fmt.Errorf("Request to URL %s failed with status code %d and body %s. Potential problem with IAM: %t", url, resp.StatusCode, string(data), potentialProblemWithIAM)
		} else {
			err = fmt.Errorf("Request to URL %s failed with status code %d. Potential problem with IAM: %t", url, resp.StatusCode, potentialProblemWithIAM)
		}
		if potentialProblemWithIAM {
			return nil, &iamError{rootError: err}
		}
		return nil, err
	}

	return resp, err

}

// checkServiceAPIKey checks if the Cloud service api key has been set, it attempts to set it with the env variable.
//
// an error is returned if api key is empty and the env variable is not set
func (iam *IAMAuth) checkServiceAPIKey() error { // TODO: will be removed in the future
	if iam.serviceAPIKey == "" {
		if svcAPIKey := os.Getenv("CLOUD_SERVICE_API_KEY"); svcAPIKey != "" {
			iam.serviceAPIKey = svcAPIKey
			iam.configPep()
		} else {
			return errors.New("No Cloud Service API Key has been set. Please set it with the SetServiceAPIKey() function or by setting env variable CLOUD_SERVICE_API_KEY")
		}
	}

	return nil
}

// getAccountID splits the token and take the middle part and decodes it to retrieve the account ID
//
// the function also checks whether a token is still valid
func getAccountID(token string) (string, error) {

	tokenDecoded, err := decodeToken(token)
	if err != nil {
		log.Println(err)
		// Return a normal error instead of an iamError because the error is not due to IAM being unavailable:
		return "", err
	}

	if !isTokenValid(tokenDecoded.Expiration) {
		// Return a normal error instead of an iamError because error is not due to IAM being unavailable:
		return "", errors.New("token is expired")
	}

	return tokenDecoded.Account.BSS, nil

}

// decodeToken decodes the IAM token and returns it
func decodeToken(token string) (tokenDecoded decodedToken, err error) {

	tkn := strings.SplitN(token, ".", 3)

	if len(tkn) != 3 {
		// Return a normal error instead of an iamError because the error is not due to IAM being unavailable:
		return tokenDecoded, errors.New("invalid token")
	}

	var decoded []byte

	// IAM tokens are JWT tokens as defined by https://datatracker.ietf.org/doc/html/rfc7519.html so
	// the token payload is base64 url-encoded and RawURLEncoding needs to be used for the decoding:
	decoded, err = base64.RawURLEncoding.DecodeString(tkn[1])
	if err != nil {
		// Return a normal error instead of an iamError because the error is not due to IAM being unavailable:
		return
	}

	// log.Println(string(decoded))

	err = json.Unmarshal(decoded, &tokenDecoded)
	if err != nil {
		// Return a normal error instead of an iamError because the error is not due to IAM being unavailable:
		return
	}

	log.Printf("%s User IAM id: %s; User IAM name: %s\n", Audit, tokenDecoded.IAMID, tokenDecoded.Sub)

	return
}

// isTokenValid checks if the token is expired
func isTokenValid(tokenExpiration int64) bool {

	now := time.Now().Unix()

	if now > tokenExpiration {
		log.Printf("[iam-auth.go] token is expired, tokenExpiration: %v time.Unix(): %v", tokenExpiration, now)
		return false
	}

	return true
}

func closeBody(f func() error) {

	if err := f(); err != nil {
		log.Println(err)
	}

}

func (iam *IAMAuth) getEmail(tokenOrAPIkey string) (string, error) {

	token, err := iam.getTokenFromRequest(tokenOrAPIkey, nil)
	if err != nil {
		return "", err
	}

	tokenDecoded, err := decodeToken(token)
	if err != nil {
		return "", err
	}

	if !isTokenValid(tokenDecoded.Expiration) {
		return "", errors.New("token is expired")
	}

	if tokenDecoded.Email != "" {
		return tokenDecoded.Email, nil
	}

	return "", errors.New("no email found in decoded token")
}

// checkAPIKeyCache searches the cache for the provided api key, if it is found return the token,
// empty string otherwise.
func (iam *IAMAuth) checkAPIKeyCache(apikey string) string {
	iam.cache.RLock()
	defer iam.cache.RUnlock()

	if apik := iam.cache.apikCache[apikey]; apik != (apikCache{}) {
		// Remove 5 minutes from token expiration to prevent getting an almost expired token:
		tokenExp := time.Unix(apik.tokenExp, 0)
		tokenExp = tokenExp.Add(time.Minute * time.Duration(-5))
		// Check if the token is still valid based on the updated token expiration:
		if isTokenValid(tokenExp.Unix()) {
			return apik.token
		}
	}

	return ""
}

// addAPIKeyToCache adds a new entry to the api key cache
func (iam *IAMAuth) addAPIKeyToCache(apik, token, authSource string) {

	iam.cache.Lock()
	defer iam.cache.Unlock()

	// add to cache only if there is room in the cache for that
	if len(iam.cache.apikCache) < iam.cache.maxSize {
		var cacheEntry apikCache
		cacheEntry.token = token

		tknDecoded, err := decodeToken(token)
		if err == nil && isTokenValid(tknDecoded.Expiration) {
			// only add to cache if the token is not expired
			cacheEntry.tokenExp = tknDecoded.Expiration
			cacheEntry.source = authSource

			// add or replace entry in the cache
			iam.cache.apikCache[apik] = cacheEntry

		} else if err != nil {
			// do not add to cache, only print out error
			log.Println(err)
		}
	} else {
		log.Printf("unable to add api key to cache, cache is at the maximum size. current size=%d. Increase the cache size if needed.\n", len(iam.cache.apikCache))
	}

}

func (iam *IAMAuth) cleanUpAPIKeyCache() {

	go (func() {

		for {
			iam.cache.RLock()

			log.Printf("current cache size=%d ; maximum cache size allowed=%d\n", len(iam.cache.apikCache), iam.cache.maxSize)

			// check if the current cache size is at 80%(current threshold) or above
			// if it is, then clean up cache(remove expired tokens)
			if len(iam.cache.apikCache) >= ((iam.cache.maxSize * iam.cache.sizeThreshold) / 100) {
				for apik, val := range iam.cache.apikCache {
					if !isTokenValid(val.tokenExp) {
						iam.cache.RUnlock()
						iam.cache.Lock()

						// remove from cache
						delete(iam.cache.apikCache, apik)

						iam.cache.Unlock()
						iam.cache.RLock()
					}
				}
			}

			iam.cache.RUnlock()

			time.Sleep(iam.cache.cleanupInterval)
		}

	})()

}

// ContextKeyString is a dedicated type used to add values to an http request context as recommended in the
// description of function WithValue()
type ContextKeyString string

// Adds data to the request context and updates the request with the modified context
func addDataToRequestContext(req *http.Request, key ContextKeyString, value string) {
	ctx := context.WithValue(req.Context(), key, value)
	*req = *req.Clone(ctx)
}

// addIAMAuthNRAttributes add all attributes for each user that has a valid API key/token
func (iam *IAMAuth) addIAMAuthNRAttributes(ctx context.Context, accessToken, permission, resource string, auth bool) {
	email, err := iam.getEmail("Bearer " + accessToken)
	if err != nil {
		log.Println(err)
	} else {
		monitoring.SetTagsKV(ctx,
			"iamAuthEmail", email,
			"iamAuthAuthorized", auth,
			"iamAuthPermission", permission,
			"iamAuthResource", resource)
	}
}

// getReqSourceIP returns the request source IP
func getReqSourceIP(r *http.Request) string {
	return r.Header.Get("X-Forwarded-For")
}
