package token

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	guuid "github.com/google/uuid"
	"github.com/pkg/errors"
	gojwks "github.ibm.com/IAM/go-jwks"
)

const (
	// DefaultTokenExpiry is the percent of one hour (45 minutes) until automatic token fetch
	DefaultTokenExpiry = 0.75

	// DefaultTokenRetryDelay is the number seconds to wait between token fetch retries
	DefaultTokenRetryDelay = 15

	// DefaultFetchIntervalSec a default fetch time in seconds for edge case fetch failures
	DefaultFetchIntervalSec = int64(300)

	// CRNPREFIX is the realmid for a crn token
	CRNPREFIX = "crn-"
)

var jwtTime sync.RWMutex

// IAMToken represents the contents of a token object from the IAM identity service
type IAMToken struct {
	AccessToken  string `json:"access_token"`  //base64 encoded string containing the access token
	RefreshToken string `json:"refresh_token"` //base64 encoded string containing the refresh token
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Expiration   int    `json:"expiration"`
	Scope        string `json:"scope"`
}

// IAMAccessTokenClaims represents the claims in a JWT from the IAM identity service
type IAMAccessTokenClaims struct {
	*jwt.StandardClaims
	IAMID              string   `json:"iam_id"`
	ID                 string   `json:"id"`
	RealmID            string   `json:"realmid"`
	Identifier         string   `json:"identifier"`
	GivenName          string   `json:"given_name"`
	FamilyName         string   `json:"family_name"`
	Name               string   `json:"name"`
	Email              string   `json:"email"`
	Account            Account  `json:"account"`
	GrantType          string   `json:"grant_type"`
	Scope              string   `json:"scope"`
	ClientID           string   `json:"client_id"`
	ACR                int      `json:"acr"`
	AMR                []string `json:"amr"`
	Sub                string   `json:"sub"`
	SubType            string   `json:"sub_type"`
	UniqueInstanceCrns []string `json:"unique_instance_crns"`
	Authn              Authn    `json:"authn"`
}

// IAMTokenError represents an error received from the IAM identity service API for a response
// with a status code other than 200
type IAMTokenError struct {
	Context struct {
		RequestID   string `json:"requestId"`
		RequestType string `json:"requestType"`
		UserAgent   string `json:"userAgent"`
		ClientIP    string `json:"clientIp"`
		URL         string `json:"url"`
		InstanceID  string `json:"instanceId"`
		ThreadID    string `json:"threadId"`
		Host        string `json:"host"`
		StartTime   string `json:"startTime"`
		EndTime     string `json:"endTime"`
		ElapsedTime string `json:"elapsedTime"`
		Locale      string `json:"locale"`
		ClusterName string `json:"clusterName"`
	} `json:"context"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

// Account claims in an IAM token
type Account struct {
	Valid bool   `json:"valid"`
	Ims   string `json:"ims"`
	Bss   string `json:"bss"`
}

// Authn claims in an IAM token
type Authn struct {
	AuthnId   string `json:"iam_id"`
	AuthnName string `json:"name"`
}

// TokenManager is the struct used to define and retrieve a particular token created with #NewToken()
type TokenManager struct {
	token        string               // encoded json web token, must be used with a #mutex TODO: change this to be an IAMToken so we can make use of the refresh token
	claims       IAMAccessTokenClaims // decoded claims from the token, should also be used with a #mutex
	config       Config               // configuration of the token
	currentError error
	initMutex    sync.RWMutex
	singleCacheUtils
}

type serviceToken struct {
	Token string `json:"access_token" binding:"required"`
}
type tokenServiceResponse struct {
	tokenString      string
	tokenRespDetails tokenResponseDetails
}

type tokenResponseDetails struct {
	respCode   int
	retryAfter int
}

type tokenCallDetails struct {
	body     url.Values
	tmConfig *ExtendedConfig
	insecure bool
}

func (tm *TokenManager) initCache() {
	tm.expiryTime = (time.Duration(float64(3600)*(tm.config.ExtendedConfig.TokenExpiry)) * time.Second)
	tm.currentError = nil

	tm.updateCache()
}

func (tm *TokenManager) updateCache() {
	var err error
	tokenString := ""
	var claims *IAMAccessTokenClaims

	var fetchInterval int64

	tm.mutex.RLock()
	if reflect.DeepEqual(IAMAccessTokenClaims{}, tm.claims) || tm.config.ExtendedConfig.expirySeconds != 0 {
		fetchInterval = tm.config.ExtendedConfig.expirySeconds
		if fetchInterval == 0 {
			fetchInterval = DefaultFetchIntervalSec
		}
	} else {
		fetchInterval = int64(float64(tm.claims.ExpiresAt-tm.claims.IssuedAt) * 0.75)
	}
	tm.mutex.RUnlock()
	for {
		// make call
		retryAfter := 0
		config := tm.getConfig().(Config)
		var tsResponse tokenServiceResponse

		tsResponse, err = fetchToken(tm.config.APIKey, false, config.ExtendedConfig)
		retryAfter = tsResponse.tokenRespDetails.retryAfter
		tokenString = tsResponse.tokenString

		// analyze error
		if err != nil && (tsResponse.tokenRespDetails.respCode == http.StatusTooManyRequests || tsResponse.tokenRespDetails.respCode >= http.StatusInternalServerError) {

			retryDelay := config.ExtendedConfig.retryTimeout

			if tsResponse.tokenRespDetails.respCode == http.StatusTooManyRequests && retryDelay == 0 { // retry with header
				retryDelay = retryAfter
			}

			if retryDelay == 0 {
				retryDelay = DefaultTokenRetryDelay
			}
			config.ExtendedConfig.Logger.Error("unable to retrieve token, retrying in (s): ", retryDelay, err)
			time.Sleep(time.Duration(retryDelay) * time.Second)

		} else {
			break
		}
	}

	if err != nil {
		_ = fmt.Errorf("fetch token err: %s", err.Error())
		tm.mutex.Lock()
		defer tm.mutex.Unlock()
		tm.config.ExtendedConfig.Logger.Error(err, " unable to retrieve token")
		tm.currentError = errors.Wrap(err, "unable to retrieve token")
		tm.utilsInitialized = true

		nextScheduledRun := DefaultFetchIntervalSec

		if fetchInterval < DefaultFetchIntervalSec {
			nextScheduledRun = fetchInterval
		}
		singleSchedule(tm.updateCache, nextScheduledRun)
		return
	}

	claims, err = parseAndValidateToken(tokenString, tm.config.ExtendedConfig.Endpoints.KeyEndpoint, true)

	if err != nil {
		_ = fmt.Errorf("validation err %s", err)

		nextScheduledRun := DefaultFetchIntervalSec

		if fetchInterval < DefaultFetchIntervalSec {
			nextScheduledRun = fetchInterval
		}

		tm.mutex.Lock()
		defer tm.mutex.Unlock()
		tm.config.ExtendedConfig.Logger.Error(err, " unable to validate token")
		tm.currentError = errors.Wrap(err, "Unable to validate token")
		tm.utilsInitialized = true
		singleSchedule(tm.updateCache, nextScheduledRun)
		return
	}
	tm.writeToken(tokenString, claims)

	tm.mutex.RLock()
	if tm.config.ExtendedConfig.expirySeconds <= 0 {
		fetchInterval = int64(float64(claims.ExpiresAt-claims.IssuedAt) * 0.75)
	}
	tm.mutex.RUnlock()
	singleSchedule(tm.updateCache, fetchInterval)
}

func (tm *TokenManager) writeToken(tokenString string, claims *IAMAccessTokenClaims) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	//update claims with a copy over
	tm.token = tokenString
	tm.claims = *claims
	tm.currentError = nil
	tm.utilsInitialized = true
}

// returns true if token is expired
func (tm *TokenManager) expiredToken() bool {
	now := jwt.TimeFunc().Unix()

	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if tm.claims.StandardClaims == nil {
		return false
	}

	return !tm.claims.VerifyExpiresAt(now, false)
}

func (tm *TokenManager) initializeIfNeeded() {
	tm.initMutex.Lock()
	defer tm.initMutex.Unlock()
	if !tm.isInitialized() {
		// token cache is empty
		tm.initCache()
		for !tm.isInitialized() {
		}
	}
}

func (tm *TokenManager) isInitialized() bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.utilsInitialized
}

// FetchToken Fetches a single access token for the given apiKey. No cacheing, retries, or other token management is performed.
func FetchToken(apiKey string, env DeploymentEnvironment, optionalConfig *ExtendedConfig) (token string, respCode int, err error) {
	if apiKey == "" {
		return "", 0, errors.New("The API key is needed to get a token")
	}

	tsResponse := tokenServiceResponse{}

	if optionalConfig == nil {
		optionalConfig = &ExtendedConfig{}
	}

	// Create TokenManager but do not initialize so no cache or refresh schedule is created
	tm := &TokenManager{}

	// Handle endpoints
	err = tm.envConfigure(env, &optionalConfig.Endpoints)

	if err != nil {
		return "", 0, err
	}

	// Configure other extended params
	err = tm.tokenParamConfigure(optionalConfig)

	if err != nil {
		return "", 0, err
	}

	insecure := false
	if env == Custom {
		insecure = true
	}

	tsResponse, err = fetchToken(apiKey, insecure, tm.config.ExtendedConfig)

	token = tsResponse.tokenString
	respCode = tsResponse.tokenRespDetails.respCode
	return
}

// TODO: return entire token instead of just access token string
/*
curl -k -X POST \
  --header "Content-Type: application/x-www-form-urlencoded" \
  --header "Accept: application/json" \
  --data-urlencode "grant_type=urn:ibm:params:oauth:grant-type:apikey" \
  --data-urlencode "apikey=<my key>" \
  "https://iam.cloud.ibm.com/identity/token"
*/
func fetchToken(apiKey string, insecure bool, tmConfig ExtendedConfig) (tsResponse tokenServiceResponse, err error) { //, retrySeconds int, err error) {

	var result IAMToken //map[string]interface{}
	tsResponse = tokenServiceResponse{}
	callDetails := tokenCallDetails{
		tmConfig: &tmConfig,
		insecure: insecure,
	}
	if apiKey == "" {
		return tsResponse, errors.New("The API key is needed to get a token")
	}

	if tmConfig.Endpoints.TokenEndpoint == "" {
		return tsResponse, errors.New("The token endpoint must be specified")
	}

	body := url.Values{}
	body.Set("grant_type", "urn:ibm:params:oauth:grant-type:apikey")
	body.Set("apikey", apiKey)

	if tmConfig.Scope != "" {
		body.Set("scope", tmConfig.Scope)
	}

	callDetails.body = body
	tsResponse.tokenRespDetails, err = doTokenCall(callDetails, &result)

	if err == nil {
		tsResponse.tokenString = result.AccessToken
		return tsResponse, err
	}

	return tsResponse, err

}

func doTokenCall(callDetails tokenCallDetails, tokenStruct interface{}) (respDetails tokenResponseDetails, err error) {

	tokenURL := callDetails.tmConfig.Endpoints.TokenEndpoint
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(callDetails.body.Encode()))
	if err != nil {
		return respDetails, err
	}

	if callDetails.tmConfig != nil && (callDetails.tmConfig.ClientID != "" || callDetails.tmConfig.ClientSecret != "") {
		req.SetBasicAuth(callDetails.tmConfig.ClientID, callDetails.tmConfig.ClientSecret)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	txID := guuid.New().String()
	req.Header.Add("Transaction-Id", txID)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: callDetails.insecure},
	}

	// pre-set a default HTTP timeout value in case it is not configured
	httpTimeout := 15
	if callDetails.tmConfig != nil && callDetails.tmConfig.HttpTimeout != 0 {
		httpTimeout = callDetails.tmConfig.HttpTimeout
	}

	client := &http.Client{
		Timeout:   time.Duration(httpTimeout) * time.Second,
		Transport: transport}

	resp, err := client.Do(req)

	if err, ok := err.(net.Error); ok && err.Timeout() {
		respDetails.respCode = http.StatusGatewayTimeout
		return respDetails, errors.Wrap(err, " Transaction-Id: "+txID)
	} else if ok && err.Temporary() {
		respDetails.respCode = http.StatusBadGateway
		return respDetails, errors.Wrap(err, " Transaction-ID"+txID)
	}

	if err != nil {
		respDetails.respCode = 999
		if err, ok := err.(*url.Error); ok {
			if _, ok := err.Err.(x509.UnknownAuthorityError); ok {
				// tls error for unknown certificate signer
				return respDetails, errors.Wrap(err, "TLS error, Transaction-Id: "+txID)
			}
			if err, ok := err.Err.(*net.OpError); ok {
				if _, ok := err.Err.(*net.DNSError); ok {
					// DNS errors ( dropped connection)
					return respDetails, errors.Wrap(err, "DNS error, Transaction-Id: "+txID)
				}
			}
		}

		if resp == nil {
			return respDetails, errors.Wrap(err, " Transaction-Id: "+txID)
		}
		respDetails.respCode = resp.StatusCode
		return respDetails, errors.Wrap(err, " Transaction-Id: "+txID)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var IAMTokenErr IAMTokenError
		err = json.NewDecoder(resp.Body).Decode(&IAMTokenErr)

		if resp.StatusCode == 429 {
			retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
			respDetails.respCode = resp.StatusCode
			respDetails.retryAfter = retryAfter
			return respDetails, errors.Wrap(err, "response status code "+resp.Status+" Transaction-Id: "+txID)
		}

		if err != nil {
			respDetails.respCode = resp.StatusCode
			return respDetails, errors.Wrap(err, "response status code "+resp.Status+" Transaction-Id: "+txID)
		}

		respDetails.respCode = resp.StatusCode
		return respDetails, errors.New("Transaction-Id: " + txID + " Status Code " + resp.Status + " " + IAMTokenErr.ErrorCode + " " + IAMTokenErr.ErrorMessage)
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenStruct)

	if err != nil {
		respDetails.respCode = resp.StatusCode
		err = fmt.Errorf("Transaction-Id: "+txID+" Failed to decode IAM Token fetch response %s", err.Error())
		return respDetails, err
	}

	respDetails.respCode = resp.StatusCode
	return respDetails, nil
}

func parseAndValidateToken(tokenString string, keyEndpoint string, skipValidation bool) (claims *IAMAccessTokenClaims, err error) {

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &IAMAccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify

		if _, ok := token.Claims.(*IAMAccessTokenClaims); ok {
			if !skipValidation {
				// validate if the alg is RSA
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				// This is kid to get right public key for signature validation.
				kid, ok := token.Header["kid"].(string)
				if !ok {
					return nil, fmt.Errorf("Invalid kid")
				}

				// Match KID to public key
				jwKey, err := GetKeyByKID(kid, keyEndpoint)
				if err != nil {
					// add log here
					return nil, fmt.Errorf("Key does not exist for the token's KID: "+kid+", error detail: %v", err)
				}

				verifyKey := getRSAPubKey((jwKey).(gojwks.Key))
				return verifyKey, nil

			}
			return nil, nil
		}
		return nil, fmt.Errorf("Error retrieving claims from token")
	})

	if err != nil {
		validationError, ok := err.(*jwt.ValidationError)

		if ok {
			if skipValidation && validationError.Errors != jwt.ValidationErrorMalformed {
				return token.Claims.(*IAMAccessTokenClaims), nil
			}

			errorString := ""

			if validationError.Inner != nil {
				errorString += validationError.Inner.Error()
			}

			if validationError.Errors == jwt.ValidationErrorMalformed {
				errorString += "token is malformed:"
			}

			return nil, errors.New(errorString + " " + validationError.Error())
		}
		return nil, err
	}
	return token.Claims.(*IAMAccessTokenClaims), nil
}

// NewToken creates a TokenManager object which is used to make calls to PDP (ExtendedConfig is optional, but recommended)
func NewTokenManager(APIKey string, d DeploymentEnvironment, optionalConfig *ExtendedConfig) (*TokenManager, error) {
	if APIKey == "" {
		return nil, fmt.Errorf("APIKey is required")
	}

	tm := &TokenManager{}

	// Sane values if token extended config not provided
	if optionalConfig == nil {
		optionalConfig = &ExtendedConfig{
			ClientID:     "",
			ClientSecret: "",
			TokenExpiry:  DefaultTokenExpiry,
			Endpoints:    Endpoints{},
		}
	}

	// Handle endpoints
	err := tm.envConfigure(d, &(optionalConfig.Endpoints))

	if err != nil {
		return nil, err
	}

	// configure token
	tm.config.APIKey = APIKey

	// Configure other extended params
	err = tm.tokenParamConfigure(optionalConfig)

	if err != nil {
		return nil, err
	}

	for i := 0; i < 3; i++ {
		var tsResponse tokenServiceResponse
		// try credentials/configuration for validity
		tsResponse, err = fetchToken(tm.config.APIKey, false, tm.config.ExtendedConfig)

		if err == nil { //init cache with freshly fetched token
			tm.initializeIfNeeded()

			// return token manager
			return tm, nil
		}

		config := tm.getConfig().(Config)
		if err != nil && (tsResponse.tokenRespDetails.respCode == http.StatusTooManyRequests || tsResponse.tokenRespDetails.respCode >= http.StatusInternalServerError) {
			retryDelay := DefaultTokenRetryDelay
			retryAfter := tsResponse.tokenRespDetails.retryAfter
			if tsResponse.tokenRespDetails.respCode == http.StatusTooManyRequests && retryAfter > 0 { // retry with header
				retryDelay = retryAfter
			}
			config.ExtendedConfig.Logger.Error("unable to retrieve token, retrying in (s): ", retryDelay, err)
			time.Sleep(time.Duration(retryDelay) * time.Second)
		} else {
			// break from the loop for 4xx errors
			config.ExtendedConfig.Logger.Error("unable to retrieve token. Status code: ", tsResponse.tokenRespDetails.respCode, err)
			break
		}
	}

	return nil, err

}

// Valid Validates time based claims "exp" and "iss".
// There is no accounting for clock skew. This function is created to fulfill the go-jwt Claims interface
// We added this implementation since clock skew in adopters environments was causeing issus with the iat and nbf claims
// which the default Valid() function takes into account
// More information: https://github.com/dgrijalva/jwt-go/blob/9742bd7fca1c67ba2eb793750f56ee3094d1b04f/claims.go#L9
// returns a jwt.TokenExpiredError error exp has passed
// returns a jwt.InvalidIssuerError iss is invalid
func (tm IAMAccessTokenClaims) Valid() error {

	vErr := new(jwt.ValidationError)
	now := jwt.TimeFunc().Unix()

	// The claims below are optional, by default, so if they are set to the
	// default value in Go, let's not fail the verification for them.
	if !tm.VerifyExpiresAt(now, false) {
		delta := time.Unix(now, 0).Sub(time.Unix(tm.ExpiresAt, 0))
		vErr.Inner = fmt.Errorf("token is expired by %v", delta)
		vErr.Errors |= jwt.ValidationErrorExpired
	}

	if vErr.Errors == 0 {
		return nil
	}

	return vErr
}

// GetToken returns the cached access token
func (tm *TokenManager) GetToken() (string, error) {

	tm.initializeIfNeeded()

	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if tm.expiredToken() {
		return "", errors.New("token is expired, trying to fetch new token")
	}
	return tm.token, nil
}

// DeleteToken deletes the TokenManager object and the data it contains
// func (tm *TokenManager) DeleteToken() {
// 	tm.mutex.Lock()
// 	tm.quitCacheLoop <- true
// 	tm = nil
// }

// TokenValidator struct represents the validator containing the token and key host endpoint
// key resource paths are assumed to be "/identity/keys"
type TokenValidator struct {
	hostEndpoint string // endpoint of the token and public key
}

// NewTokenValidator takes a hostEndpoint representing the key and token sources
// returns a pointer to a TokenValidator and error
func NewTokenValidator(hostEndpoint string) (*TokenValidator, error) {

	tv := &TokenValidator{}

	if hostEndpoint == "" {
		return nil, errors.Errorf("host or key endpoint not specified")
	}

	tv.hostEndpoint = hostEndpoint

	GetKeys(hostEndpoint) // this will initialize the key cache for this host if it has not yet been initialized

	return tv, nil
}

// GetHost returns the host endpoint configured for the TokenValidator
func (tv *TokenValidator) GetHost() string {
	return tv.hostEndpoint
}

// GetClaims returns an IAMAccessTokenClaims object
// param token the token string whose claims will be returned
// param skipValidation skips the validation of the token expiration and signature if set to true
func (tv *TokenValidator) GetClaims(token string, skipValidation bool) (*IAMAccessTokenClaims, error) {
	return parseAndValidateToken(token, tv.hostEndpoint, skipValidation)
}

// GetClaims consumes a JWT and returns a list of token claims. Can skip JWT validation.
func (tm *TokenManager) GetClaims(token string, skipValidation bool) (*IAMAccessTokenClaims, error) {
	return parseAndValidateToken(token, tm.config.ExtendedConfig.Endpoints.KeyEndpoint, skipValidation)
}

// GetSubjectAsIAMIDClaim Consumes a JWT and returns the iam_id claim. Can skip JTW validation.
func (tm *TokenManager) GetSubjectAsIAMIDClaim(token string, skipValidation bool) (string, error) {
	claims, err := parseAndValidateToken(token, tm.config.ExtendedConfig.Endpoints.KeyEndpoint, skipValidation)

	if err != nil {
		return "", err
	}

	return claims.IAMID, nil
}

// GetDelegationToken - a type method that retrieves an access token representing the desiredIAMId
//		Args:  desiredIAMId: the crn of the targeted IAM ID
//
// Note: this function uses the accessToken found in the TokenManager that it is called on.
// 				The access token representing the identity originating the
//              delegation request. This identity, generally a serviceId, needs to have a
//              crn:v1:bluemix:public:iam-identity-platform::::role:Delegate role over
//              its serviceName
func (tm *TokenManager) GetDelegationToken(desiredIAMId string) (string, error) {
	accessToken, err := tm.GetToken()

	if err != nil {
		return "", err
	}

	return fetchDelegationToken(accessToken, desiredIAMId, tm.config.ExtendedConfig)
}

// GetDelegationToken - A generic convenience function that retrieves an access token representing the desiredIAMId
//    Args:  accessToken: The access token representing the identity originating the
//              delegation request. This identity, generally a serviceId, needs to have a
//              crn:v1:bluemix:public:iam-identity-platform::::role:Delegate role over
//              its serviceName
//           desiredIAMId: the crn of the targeted IAM ID
//			 ec: the configuration for the request. The ExtendedConfig.Endpoints.TokenEndpoint is required
func GetDelegationToken(accessToken string, desirdesiredIAMId string, ec ExtendedConfig) (string, error) {
	return fetchDelegationToken(accessToken, desirdesiredIAMId, ec)
}

func fetchDelegationToken(accessToken string, desiredIAMId string, ec ExtendedConfig) (string, error) {
	var token serviceToken

	if (ExtendedConfig{}) == ec {
		return "", errors.New("missing configuration parameter to fetch delegation token")
	}

	if (Endpoints{} == ec.Endpoints) || (len(ec.Endpoints.TokenEndpoint) == 0) {
		return "", errors.New("missing token endpoint in config")
	}

	// IAM-Identity authz grant_type for delgation token
	const AUTHZGRANT = "urn:ibm:params:oauth:grant-type:iam-authz"

	// Very often people forget that realmId is needed to prefix the crn. Auto add it if not present
	if !strings.HasPrefix(desiredIAMId, CRNPREFIX) {
		desiredIAMId = CRNPREFIX + desiredIAMId
	}

	body := url.Values{}
	body.Set("grant_type", AUTHZGRANT)
	body.Set("access_token", accessToken)

	if ec.Scope != "" {
		body.Set("scope", ec.Scope)
	}

	body.Set("desired_iam_id", desiredIAMId)
	body.Set("grant_type", AUTHZGRANT)

	tcDetails := tokenCallDetails{
		body:     body,
		tmConfig: &ec,
		insecure: false,
	}

	_, err := doTokenCall(tcDetails, &token)

	if err == nil {
		return token.Token, err
	}

	return "", err

}
