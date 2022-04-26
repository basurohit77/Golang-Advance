package token

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"sync"
	"time"

	gojwks "github.ibm.com/IAM/go-jwks"
)

// created to update from the gojwks internal storage since gojwks blocks on HTTP requests
type cacheKey struct {
	jwk            gojwks.Keys // json web keys, must be used with #mutex
	endpoint       string
	initMutex      sync.RWMutex
	keyCacheExpiry float64
	client         *gojwks.Client
	singleCacheUtils
}

const defaultKeyCacheExpiry float64 = 3600 // Time until expiry of gojwks cache expiry

var config *gojwks.ClientConfig
var keyCacheMutex sync.RWMutex
var keyCache = make(map[string]*cacheKey)

func (c *cacheKey) initCache() {
	debugLogging := false
	config = gojwks.NewConfig()
	//this gets translated to seconds inside the framework even though it takes a time.Duration
	config.WithCacheTimeout(time.Duration(c.keyCacheExpiry))

	debugLoggingEnvVar := os.Getenv("JWKSDEBUGLOGGING")
	if debugLoggingEnvVar != "" {
		parsedDebugLogging, err := strconv.ParseBool(debugLoggingEnvVar)
		if err != nil {
			fmt.Println("DebugLogging could not be set", err)
		} else {
			debugLogging = parsedDebugLogging
		}
	}
	config.WithDebugLogging(debugLogging, nil)

	c.client = gojwks.NewClient(c.endpoint, config) // start stage and prod thread and independent local? Also create new init for those and way to identify the keys?
	c.expiryTime = (time.Duration(c.keyCacheExpiry) * time.Second)

	cacheInterval(c)
}

func (c *cacheKey) updateCache() {
	keys, err := c.client.GetKeys()
	if err != nil {
		_ = fmt.Errorf(err.Error())
		fmt.Println(err)
		return
	}

	if len(keys) == 0 {
		_ = fmt.Errorf("fetch returned empty keys slice")
	} else {
		c.writeKeys(keys)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.utilsInitialized = true
}

func (c *cacheKey) writeKeys(keys []gojwks.Key) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	//update keyCache with a copy over
	c.jwk.Keys = make([]gojwks.Key, len(keys))

	for i, k := range keys {
		c.jwk.Keys[i].Alg = k.Alg
		c.jwk.Keys[i].E = k.E
		c.jwk.Keys[i].Kid = k.Kid
		c.jwk.Keys[i].Kty = k.Kty
		c.jwk.Keys[i].N = k.N
		c.jwk.Keys[i].Use = k.Use
		c.jwk.Keys[i].X5t = k.X5t

		// using copy so that pointers are copied by value
		copy(c.jwk.Keys[i].X5c, k.X5c)
	}
}

func (c *cacheKey) initializeIfNeeded() {
	c.initMutex.Lock()
	defer c.initMutex.Unlock()
	if !c.isInitialized() {
		// key cache is empty
		c.initCache()
		for !c.isInitialized() {
		}
	}
}

func initializeEnvKeyCacheIfNeeded(endpoint string, expiry float64) {
	// TODO improve initialization by adding env check to make sure endpoint is reachable and return an error after a few retries
	keyCacheMutex.Lock()
	defer keyCacheMutex.Unlock()
	key, ok := keyCache[endpoint]

	if ok {
		key.keyCacheExpiry = expiry
		key.initializeIfNeeded()
		return
	}

	k := &cacheKey{endpoint: endpoint, keyCacheExpiry: expiry}
	k.initializeIfNeeded()
	keyCache[endpoint] = k
}

func getRSAPubKey(jwKey gojwks.Key) *rsa.PublicKey {

	n, err := base64.RawURLEncoding.DecodeString(jwKey.N)
	if err != nil {
		_ = fmt.Errorf(err.Error())
		return nil
	}

	// The default exponent is usually 65537, so just compare the
	// base64 for [1,0,1] or [0,1,0,1]
	e := 65537
	if jwKey.E != "AQAB" { //&& jwKey.E != "AAEAAQ" {
		// still need to decode the big-endian int

		eBytes, err := base64.RawURLEncoding.DecodeString(jwKey.E)
		e = int((new(big.Int).SetBytes(eBytes)).Uint64())
		if err != nil {
			_ = fmt.Errorf(err.Error())
			return nil
		}
	}

	RSApk := &rsa.PublicKey{
		N: new(big.Int).SetBytes(n),
		E: e,
	}

	return RSApk
}

// GetKeys retrieves a copy of the public keys cache without affecting the cache
// It returns a gojwks.Keys struct object containing the public keys
func GetKeys(environmentHost string) (readKeys gojwks.Keys) {

	initializeEnvKeyCacheIfNeeded(environmentHost, defaultKeyCacheExpiry)

	keyCacheMutex.RLock()
	key := keyCache[environmentHost]
	keyCacheMutex.RUnlock()

	key.mutex.RLock()
	defer key.mutex.RUnlock()

	readKeys.Keys = make([]gojwks.Key, len(key.jwk.Keys))

	for i, k := range key.jwk.Keys {
		readKeys.Keys[i].Alg = k.Alg
		readKeys.Keys[i].E = k.E
		readKeys.Keys[i].Kid = k.Kid
		readKeys.Keys[i].Kty = k.Kty
		readKeys.Keys[i].N = k.N
		readKeys.Keys[i].Use = k.Use
		readKeys.Keys[i].X5t = k.X5t

		// using copy so that pointers are copied by value
		copy(readKeys.Keys[i].X5c, k.X5c)
	}
	return
}

// GetKeyByKID returns a single public key with the supplied KID if one is found
// returns a nil interface and error if the key is not found in the cache
func GetKeyByKID(kid string, environmentEndpoint string) (interface{}, error) {
	keys := GetKeys(environmentEndpoint)

	for _, v := range keys.Keys {
		if kid == v.Kid {
			return v, nil
		}
	}

	return nil, fmt.Errorf("KID %v not found for environment %v. Available keys for this environment are: %+v", kid, environmentEndpoint, keys)
}
