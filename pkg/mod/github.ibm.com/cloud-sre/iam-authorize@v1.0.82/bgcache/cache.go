package bgcache

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.ibm.com/cloud-sre/iam-authorize/elasticsearch"
	"github.ibm.com/cloud-sre/iam-authorize/iam"
	"github.ibm.com/cloud-sre/iam-authorize/monitoring"
	"log"
	"strings"
	"sync"

	instana "github.com/instana/go-sensor"
	newrelicPreModules "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/v3/newrelic"
	t "time"
)

// BGCache is the interface implemented by the break glass cache and represents the functions that are available.
type BGCache interface {
	GetAuthorization(authKey, resource, permission string) *bgAuthorization
	AddAuthorization(authKey, iamID, user, source, token, resource, permission string, time int64, atPodStartup bool)
	IsUserAuthorized(iamID, source, resource, permission string) bool
	AddUser(iamID, user, source, resource, permission string, time int64, atPodStartup bool)
}

const bgDurationLimit = 5184000          // Maximum number seconds to allow BG access
type bgScope map[string]map[string]int64 // The time a certain permission has been give to a resource

// bgAuthorization struct used to return authorization values of a request
type bgAuthorization struct {
	IamID    string
	User     string
	Source   string
	Token    string
	resource bgScope
}

// bgCache is the implementation of the break glass cache interface.
type bgCache struct {
	memCacheRecords     map[string]bgAuthorization
	encryptionKey       string
	mux                 sync.Mutex
	userMemCacheRecords map[string]bgAuthorization
	userMemCachMux      sync.Mutex
	monConfig           *iam.NRWrapperConfig
}

var (
	bgc *bgCache
	enableElasticsearch bool
	enableElasticsearchMux sync.Mutex
)

func InitBgCache(monConfig *iam.NRWrapperConfig) {
	// Retrieve maps for storing cache data and add nrConfig
	GetBGCache()
	if bgc.monConfig == nil {
		bgc.monConfig = monConfig
	}

	// Disable writing to Elasticsearch first
	enableElasticsearchMux.Lock()
	enableElasticsearch = false
	enableElasticsearchMux.Unlock()

	// Start background go routine to load cache entries from Elasticsearch at pod startup
	go syncWithElasticsearch(monConfig)
}

// syncWithElasticsearch will load cache content from Elasticsearch and put into in-memory cache
func syncWithElasticsearch(monConfig *iam.NRWrapperConfig) {
	const fct = "syncWithElasticsearch: "
	const syncWithElasticsearchName = "iam-authorize-syncWithElasticsearch"

	// Wait for connection to Instana agent
	t.Sleep(30 * t.Second)

	log.Println(fct, "Starting after 30s wait...")

	// Create or retrieve a span in Instana
	var sensor *instana.Sensor
	if monConfig != nil && monConfig.InstanaSensor != nil {
		sensor = monConfig.InstanaSensor
	}
	bgCacheSpan, ctx := monitoring.NewSpan(context.Background(), sensor, syncWithElasticsearchName)
	if bgCacheSpan == nil {
		log.Println(fct, "couldn't retrieve or create a span")
	} else {
		defer bgCacheSpan.Finish()
	}

	// Start a transaction in New Relic
	var txnPreModules newrelicPreModules.Transaction
	var txn *newrelic.Transaction
	if monConfig != nil && monConfig.NRApp != nil {
		// Wait for nrConfig.NRApp to connect when pod starts
		if err := monConfig.NRApp.WaitForConnection(5 * t.Second); err != nil {
			log.Println(fct, "Failed to connect to New Relic: ", err)
		}

		txnPreModules = monConfig.NRApp.StartTransaction(syncWithElasticsearchName, nil, nil)
		ctx = newrelicPreModules.NewContext(ctx, txnPreModules)
		defer txnPreModules.End()
	}

	if monConfig != nil {
		// Wait for nrConfig.NRAppV3 to connect when pod starts
		if err := monConfig.NRAppV3.WaitForConnection(5 * t.Second); err != nil {
			log.Println(fct, "Failed to connect to New Relic V3: ", err)
		}

		txn = monConfig.NRAppV3.StartTransaction(syncWithElasticsearchName)
		ctx = newrelic.NewContext(ctx, txn)
		defer txn.End()

		// Send data to Instana and New Relic
		monitoring.SetTagsKV(ctx,
			"environment", monConfig.Environment,
			"region", monConfig.Region,
			"sourceServiceName", monConfig.SourceServiceName,
			"url", monConfig.URL)
	}

    // Make sure bgc is initialized
    if bgc == nil {
		errMsg := "Failed to initialize bgc at pod startup"
		log.Println(fct, errMsg)

		monitoring.SetTagsKV(ctx,
			"bgErr", errMsg,
			"syncWithElasticsearchSuccess", false)
		return
	}

	// Get the query
	query := queryAllCacheEntries

	// Get (or create) an Elasticsearch instance
	es, err := getElasticsearch()
	if err != nil {
		errMsg := "Failed to get Elasticsearch instance at pod startup"
		log.Println(fct, errMsg)

		monitoring.SetTagsKV(ctx,
			"bgErr", errMsg,
			"syncWithElasticsearchSuccess", false)
		return
	}

	// Elasticsearch Search
	response, err := es.Search(query, breakglassIndex)
	if err != nil {
		errMsg := "Failed to search from Elasticsearch at pod startup"
		log.Println(fct, errMsg)

		monitoring.SetTagsKV(ctx,
			"bgErr", errMsg,
			"syncWithElasticsearchSuccess", false)
		return
	}

	// No action if there is no record in Elasticsearch
	if len(response.Hits.Hits) < 1 {
		enableElasticsearchMux.Lock()
		enableElasticsearch = true
		enableElasticsearchMux.Unlock()

		monitoring.SetTag(ctx, "syncWithElasticsearchSuccess", true)
		return
	}

	// Convert response to cache records and save to in-memory cache maps
	for _, hit := range response.Hits.Hits {
		// Convert each hit to a cache record
		esCacheRecord, err := hitToESCacheRecord(hit)
		if err != nil {
			// Discard bad hit and send id and error message to Instana and New Relic
			monitoring.SetTagsKV(ctx,
				"id", hit.ID,
				"hitToESCacheRecordFailure", err)
			continue
		}

		// Add the record to in-memory cache maps
		for _, key := range esCacheRecord.APIKeys {
			if esCacheRecord.IsAPI == "true" {
				// It is an API cache record
				for _, resource := range key.ESResources {
					for _, permission := range resource.ResourcePermissions {
						// APIKey and Token are decrypted here
						apiKeyDecrypted, apiKeyDecryptionErr := decryptFromElasticsearch(key.APIKey, key.KeyID)
						tokenDecrypted, tokenDecryptionErr := decryptFromElasticsearch(key.Token, key.KeyID)

						// If decryption fails, report error to Instana and New Relic
						if apiKeyDecryptionErr != nil {
							monitoring.SetTagsKV(ctx,
								"user", esCacheRecord.User,
								"source", key.Source,
								"resource", resource.Resource,
								"permission", permission.Permission,
								"time", permission.Time,
								"atPodStartup", true,
								"bgErr", "Failed to decrypt API key")
							continue
						}
						if tokenDecryptionErr != nil {
							monitoring.SetTagsKV(ctx,
								"user", esCacheRecord.User,
								"source", key.Source,
								"resource", resource.Resource,
								"permission", permission.Permission,
								"time", permission.Time,
								"atPodStartup", true,
								"bgErr", "Failed to decrypt token")
							continue
						}

						// If iamID cannot be retrieved, or iamID is empty, or it is an old record that contains
						// user information, report error to Instana and New Relic
						// Allows empty user at this point
						iamID, err := getIamID(hit.ID)
						if err != nil || iamID == "" || (strings.Contains(iamID, esCacheRecord.User) && esCacheRecord.User != "") {
							monitoring.SetTagsKV(ctx,
								"user", esCacheRecord.User,
								"source", key.Source,
								"resource", resource.Resource,
								"permission", permission.Permission,
								"time", permission.Time,
								"atPodStartup", true,
								"bgErr", "Failed to get iamID")
							continue
						}

						bgc.AddAuthorization(apiKeyDecrypted, iamID, esCacheRecord.User, key.Source, tokenDecrypted, resource.Resource, permission.Permission, permission.Time, true)
					}
				}
			} else {
				// It is a token cache record
				for _, resource := range key.ESResources {
					for _, permission := range resource.ResourcePermissions {
						// No decryption needed
						iamID := hit.ID

						// If iamID is empty, or it is an old record that contains user information (including empty user),
						// report error to Instana and New Relic
						if iamID == "" || strings.Contains(hit.ID, esCacheRecord.User) {
							monitoring.SetTagsKV(ctx,
								"user", esCacheRecord.User,
								"source", key.Source,
								"resource", resource.Resource,
								"permission", permission.Permission,
								"time", permission.Time,
								"atPodStartup", true,
								"bgErr", "Invalid or empty iamID/user")
							continue
						}

						bgc.AddUser(iamID, esCacheRecord.User, key.Source, resource.Resource, permission.Permission, permission.Time, true)
					}
				}
			}
		}
	}

	// syncWithElasticsearch success
	// Enable writes to Elasticsearch
	enableElasticsearchMux.Lock()
	enableElasticsearch = true
	enableElasticsearchMux.Unlock()

	monitoring.SetTag(ctx, "syncWithElasticsearchSuccess", true)
	log.Println(fct, "Success")
}

// isElasticsearchEnabled returns the last result of enableElasticsearch
func isElasticsearchEnabled() bool {
	enableElasticsearchMux.Lock()
	defer enableElasticsearchMux.Unlock()
	return enableElasticsearch
}

// getIamID returns iamID
// The _id for API keys is api+iamID and the _id for tokens is iamID
func getIamID(id string) (string, error) {
	// Wrong format if there is no api+ prefix
	if !strings.HasPrefix(id, "api+") {
		return id, errors.New("invalid _id")
	}

	return strings.TrimPrefix(id, "api+"), nil
}

// decryptFromElasticsearch decrypts a AES_256_GCM encrypted string and returns the decrypted string
func decryptFromElasticsearch (field string, keyID int64) (string, error) {
	decodedField, err := base64.RawStdEncoding.DecodeString(field)
	if err != nil {
		return "", err
	}

	decryptedFieldInByte, err := Decrypt(decodedField, keyID)
	if err != nil {
		return "", err
	}

	return string(decryptedFieldInByte), nil
}

// hitToESCacheRecord returns a cache record instance converted from a search hit
func hitToESCacheRecord(hit elasticsearch.SearchResponseHit) (ESCacheRecord, error) {
	record := ESCacheRecord{}
	hitSource, err := json.Marshal(hit.Source)
	if err != nil {
		return record, err
	}
	err = json.Unmarshal(hitSource, &record)
	if err != nil {
		return record, err
	}
	return record, nil
}

func GetBGCache() BGCache {
	// With syncWithElasticsearch(), bgc should not be nil at this point
	if bgc == nil {
		bgc = &bgCache{
			memCacheRecords: make(map[string]bgAuthorization),
			userMemCacheRecords: make(map[string]bgAuthorization)}
	}

	return bgc
}

// GetAuthorization returns the authorization record for the provided auth key.
// If found the authorization is returned. If not found, return nil.
func (bgc *bgCache) GetAuthorization(authKey, resource, permission string) *bgAuthorization {
	bgc.mux.Lock()
	defer bgc.mux.Unlock()
	bga, keyExist := bgc.memCacheRecords[authKey]
	// Get the authorization from in-memory cache (if any)
	if keyExist {
		resourcePermissions, keyExist := bga.resource[resource]
		if keyExist {
			authorizationTime, keyExist := resourcePermissions[permission]
			if keyExist && (t.Now().Unix()-authorizationTime) < bgDurationLimit {
				return &bga
			}
		}
	}
	return nil
}

// AddAuthorization adds or updates an authorization record in the cache,
// and synchronizes in-memory cache records with Elasticsearch periodically (default to 300 seconds).
// Source is where the token or API key came from. For example, "public-iam".
func (bgc *bgCache) AddAuthorization(authKey, iamID, user, source, token, resource, permission string, time int64, atPodStartup bool) {
	const addAuthorizationName = "iam-authorize-AddAuthorization"

	bgc.mux.Lock()
	defer bgc.mux.Unlock()

	// Elasticsearch _id is api+<iamID>
	// Add the api+ prefix to differentiate API key and Token documents
	id := "api+" + iamID

	// Create or retrieve a span in Instana
	var sensor *instana.Sensor
	if bgc.monConfig != nil && bgc.monConfig.InstanaSensor != nil {
		sensor = bgc.monConfig.InstanaSensor
	}
	bgCacheSpan, ctx := monitoring.NewSpan(context.Background(), sensor, addAuthorizationName)
	if bgCacheSpan != nil {
		defer bgCacheSpan.Finish()
	}

	// Send parameters as custom attributes to New Relic
	var txnPreModules newrelicPreModules.Transaction
	var txn *newrelic.Transaction
	if bgc.monConfig != nil && bgc.monConfig.NRApp != nil {
		txnPreModules = bgc.monConfig.NRApp.StartTransaction(addAuthorizationName, nil, nil)
		ctx = newrelicPreModules.NewContext(ctx, txnPreModules)
		defer txnPreModules.End()
	}

	if bgc.monConfig != nil {
		txn = bgc.monConfig.NRAppV3.StartTransaction(addAuthorizationName)
		ctx = newrelic.NewContext(ctx, txn)
		defer txn.End()

		monitoring.SetTagsKV(ctx,
			"environment", bgc.monConfig.Environment,
			"region", bgc.monConfig.Region,
			"url", bgc.monConfig.URL,
			"sourceServiceName", bgc.monConfig.SourceServiceName,
			"user", user,
			"iamID", iamID,
			"source", source,
			"resource", resource,
			"permission", permission,
			"time", time,
			"atPodStartup", atPodStartup,
			"documentID", id,
			"bulkIndexNeeded", !atPodStartup)
	}

	// Update in-memory cache
	bga, keyExist := bgc.memCacheRecords[authKey] // bga stands for BreakGlass Authorization
	if keyExist {
		// Update resource, permission and time
		resourcePermissions, keyExist := bga.resource[resource]
		if keyExist {
			// Update time only when it is larger
			if time > resourcePermissions[permission] {
				resourcePermissions[permission] = time
			}
		} else {
			bga.resource[resource] = map[string]int64{permission: time}
		}
		// Update iamID, user, source and token (if any)
		if bga.resource != nil {
			bgc.memCacheRecords[authKey] = bgAuthorization{
				IamID:    iamID,
				User:     user,
				Source:   source,
				Token:    token,
				resource: bga.resource,
			}
		}
	} else {
		bgc.memCacheRecords[authKey] = bgAuthorization{
			IamID:    iamID,
			User:     user,
			Source:   source,
			Token:    token,
			resource: bgScope{resource: map[string]int64{permission: time}},
		}
	}

	// Do not post to Elasticsearch if iamID is empty
	if iamID == "" {
		errMsg := "AddAuthorization: invalid iamID"

		monitoring.SetTag(ctx, "bgErr", errMsg)
		return
	}

	// Post to Elasticsearch through Bulk Index every 5 minutes
	// Do not post to Elasticsearch when synchronizing records from Elasticsearch at pod startup or at an emergency brake
	if !atPodStartup && isElasticsearchEnabled() {
		// Prepare Bulk Index input

		/* The structure of ESResources
		esResources := []Resources{
			{
				Resource: resource,
				ResourcePermissions: []Permissions{
					{
						Permission: permission,
						Time:       time,
					},
				},
			},
		}
		*/

		// Iterate through cache map memCacheRecords and retrieve all records under the target iamID as input of Bulk Index
		// so that we can add/update the document in Elasticsearch with an id without losing data
		var apiKeys []Key
		// Find all Key records of a iamID
		for apiKey, bga := range bgc.memCacheRecords {
			if bga.IamID == iamID {
				// Get resources under a Key
				var resources []Resources
				var mostRecentTime int64
				for resource := range bgc.memCacheRecords[apiKey].resource {
					var permissions []Permissions
					for permission, time := range bgc.memCacheRecords[apiKey].resource[resource] {
						perm := Permissions{
							Permission: permission,
							Time:       time,
						}
						// Update mostRecentTime if one permission has more recent access
						if perm.Time > mostRecentTime {
							mostRecentTime = perm.Time
						}
						permissions = append(permissions, perm)
					}
					res := Resources{
						Resource:            resource,
						ResourcePermissions: permissions,
					}
					resources = append(resources, res)
				}

				// Get and update other info under a Key
				// Encrypt APIKey and Token here
				apiKeyEncrypted, apiKeyEncryptionKeyID, apiKeyEncryptionErr := encryptForElasticsearch(apiKey)
				tokenEncrypted, tokenEncryptionKeyID, tokenEncryptionErr := encryptForElasticsearch(bga.Token)

				// If encryption fails, report error to Instana and New Relic
				if apiKeyEncryptionErr != nil {
					monitoring.SetTagsKV(ctx,
						"bulkIndexSuccess", false,
						"bgErr", "Failed to encrypt API key")
					return
				}
				if tokenEncryptionErr != nil {
					monitoring.SetTagsKV(ctx,
						"bulkIndexSuccess", false,
						"bgErr", "Failed to encrypt token")
					return
				}
				if apiKeyEncryptionKeyID != tokenEncryptionKeyID {
					monitoring.SetTagsKV(ctx,
						"bulkIndexSuccess", false,
						"bgErr", "Failed to get encryption key ID")
					return
				}

				key := Key{
					APIKey:      apiKeyEncrypted,
					Source:      bga.Source,
					Token:       tokenEncrypted,
					KeyID:       apiKeyEncryptionKeyID,
					ESResources: resources,
					LastUpdated: mostRecentTime,
				}
				// There are multiple keys in apiKeys array since one iamID could have multiple APIKeys
				apiKeys = append(apiKeys, key)
			}
		}

		esCacheRecord := ESCacheRecord{
			User:        user,
			IsAPI:       "true",
			APIKeys:     apiKeys,
		}

		// Get (or create) an Elasticsearch instance
		es, err := getElasticsearch()
		if err != nil {
			errMsg := "Failed to get Elasticsearch instance for AddAuthorization"
			log.Println(errMsg)

			monitoring.SetTagsKV(ctx,
				"bulkIndexSuccess", false,
				"bgErr", errMsg)
			return
		}

		// Elasticsearch Bulk Index
		err = es.BulkIndex(esCacheRecord, breakglassIndex, id)
		if err != nil {
			// Elasticsearch Bulk Index failure
			errMsg := "Elasticsearch bulk index failure for AddAuthorization"
			log.Println(errMsg)

			monitoring.SetTagsKV(ctx,
				"bulkIndexSuccess", false,
				"bgErr", errMsg)
			return
		}

		// Elasticsearch Bulk Index success
		monitoring.SetTagsKV(ctx,
			"bulkIndexSuccess", true,
			"bgErr", "")
	}
}

// encryptForElasticsearch encrypts a string using AES_256_GCM and returns the encrypted string with encryption key id
func encryptForElasticsearch(field string) (string, int64, error) {
	encryptedField, keyID, err := Encrypt(field)
	if err != nil {
		// Throw away the record and report to New Relic
		return "", 0, err
	}

	return base64.RawStdEncoding.EncodeToString(encryptedField), keyID, nil
}

// IsUserAuthorized returns true if the provided iamID is allowed to call the provided
// resource with the provided permission.
func (bgc *bgCache) IsUserAuthorized(iamID, source, resource, permission string) bool {
	bgc.userMemCachMux.Lock()
	defer bgc.userMemCachMux.Unlock()
	bga, keyExist := bgc.userMemCacheRecords[iamID]
	// Get the authorization from in-memory cache (if any)
	if keyExist {
		resourcePermissions, keyExist := bga.resource[resource]
		if keyExist {
			authorizationTime, keyExist := resourcePermissions[permission]
			if keyExist && (t.Now().Unix()-authorizationTime) < bgDurationLimit {
				return true
			}
		}
	}
	return false
}

// AddUser adds the user to the cache with the provided source and permission,
// and synchronizes in-memory cache records with Elasticsearch periodically (default to 300 seconds).
// Source is where the token or API key came from. For example, "public-iam".
func (bgc *bgCache) AddUser(iamID, user, source, resource, permission string, time int64, atPodStartup bool) {
	const addUserName = "iam-authorize-AddUser"

	bgc.userMemCachMux.Lock()
	defer bgc.userMemCachMux.Unlock()

	// Elasticsearch _id is the iamID
	id := iamID

	// Create or retrieve a span in Instana
	var sensor *instana.Sensor
	if bgc.monConfig != nil && bgc.monConfig.InstanaSensor != nil {
		sensor = bgc.monConfig.InstanaSensor
	}
	bgCacheSpan, ctx := monitoring.NewSpan(context.Background(), sensor, addUserName)
	if bgCacheSpan != nil {
		defer bgCacheSpan.Finish()
	}

	// Send parameters as custom attributes to New Relic
	var txnPreModules newrelicPreModules.Transaction
	var txn *newrelic.Transaction
	if bgc.monConfig != nil && bgc.monConfig.NRApp != nil {
		txnPreModules = bgc.monConfig.NRApp.StartTransaction(addUserName, nil, nil)
		ctx = newrelicPreModules.NewContext(ctx, txnPreModules)
		defer txnPreModules.End()
	}

	if bgc.monConfig != nil {
		txn = bgc.monConfig.NRAppV3.StartTransaction(addUserName)
		ctx = newrelic.NewContext(ctx, txn)
		defer txn.End()

		monitoring.SetTagsKV(ctx,
			"environment", bgc.monConfig.Environment,
			"region", bgc.monConfig.Region,
			"url", bgc.monConfig.URL,
			"sourceServiceName", bgc.monConfig.SourceServiceName,
			"user", user,
			"iamID", iamID,
			"source", source,
			"resource", resource,
			"permission", permission,
			"time", time,
			"atPodStartup", atPodStartup,
			"documentID", id,
			"bulkIndexNeeded", !atPodStartup)
	}

	// Do not add or update cache entry if iamID is empty
	if iamID == "" {
		errMsg := "AddUser: invalid iamID"

		monitoring.SetTag(ctx, "bgErr", errMsg)
		return
	}

	// Update in-memory cache
	bga, keyExist := bgc.userMemCacheRecords[iamID] // bga stands for BreakGlass Authorization
	if keyExist {
		// Update resource, permission and time
		resourcePermissions, keyExist := bga.resource[resource]
		if keyExist {
			// Update time only when it is larger
			if time > resourcePermissions[permission] {
				resourcePermissions[permission] = time
			}
		} else {
			bga.resource[resource] = map[string]int64{permission: time}
		}
		// Update iamID, user and source (if any)
		if bga.resource != nil {
			bgc.userMemCacheRecords[iamID] = bgAuthorization{
				IamID:    iamID,
				User:     user,
				Source:   source,
				resource: bga.resource,
			}
		}
	} else {
		bgc.userMemCacheRecords[iamID] = bgAuthorization{
			IamID:    iamID,
			User:     user,
			Source:   source,
			resource: bgScope{resource: map[string]int64{permission: time}},
		}
	}

	// Post to Elasticsearch through Bulk Index every 5 minutes
	// Do not post to Elasticsearch when synchronizing records from Elasticsearch at pod startup or at an emergency brake
	if !atPodStartup && isElasticsearchEnabled() {
		// Prepare Bulk Index input

		/* The structure of ESResources
		esResources := []Resources{
			{
				Resource: resource,
				ResourcePermissions: []Permissions{
					{
						Permission: permission,
						Time:       time,
					},
				},
			},
		}
		*/

		// Iterate through cache map userMemCacheRecords and retrieve all records under the target iamID as input of Bulk Index
		// so that we can add/update the document in Elasticsearch with an id without losing data
		var tokens []Key
		var resources []Resources
		var mostRecentTime int64
		// Get resources under a Key
		// Do not iterate through userMemCacheRecords to find all Key record of an iamID
		// Since userMemCacheRecords map uses iamID as map key, userMemCacheRecords[iamID] is our target record
		for resource := range bgc.userMemCacheRecords[iamID].resource {
			var permissions []Permissions
			for permission, time := range bgc.userMemCacheRecords[iamID].resource[resource] {
				perm := Permissions{
					Permission: permission,
					Time:       time,
				}
				// Update mostRecentTime if one permission has more recent access
				if perm.Time > mostRecentTime {
					mostRecentTime = perm.Time
				}
				permissions = append(permissions, perm)
			}
			res := Resources{
				Resource:            resource,
				ResourcePermissions: permissions,
			}
			resources = append(resources, res)
		}

		// Get and update other info under a Key
		// In theory there should be only one key in tokens array since there is only one target iamID
		// APIKey and Token are set to empty
		key := Key{
			APIKey:      "",
			Source:      bgc.userMemCacheRecords[iamID].Source,
			Token:       "",
			KeyID:       0,
			ESResources: resources,
			LastUpdated: mostRecentTime,
		}
		tokens = append(tokens, key)

		esCacheRecord := ESCacheRecord{
			User:        user,
			IsAPI:       "false",
			APIKeys:     tokens,
		}

		// Get (or create) an Elasticsearch instance
		es, err := getElasticsearch()
		if err != nil {
			errMsg := "Failed to get Elasticsearch instance for AddUser"
			log.Println(errMsg)

			monitoring.SetTagsKV(ctx,
				"bulkIndexSuccess", false,
				"bgErr", errMsg)
			return
		}

		// Elasticsearch Bulk Index
		err = es.BulkIndex(esCacheRecord, breakglassIndex, id)
		if err != nil {
			// Elasticsearch Bulk Index failure
			errMsg := "Elasticsearch bulk index failure for AddUser"
			log.Println(errMsg)

			monitoring.SetTagsKV(ctx,
				"bulkIndexSuccess", false,
				"bgErr", errMsg)
			return
		}

		// Elasticsearch Bulk Index success
		monitoring.SetTagsKV(ctx,
			"bulkIndexSuccess", true,
			"bgErr", "")
	}
}
