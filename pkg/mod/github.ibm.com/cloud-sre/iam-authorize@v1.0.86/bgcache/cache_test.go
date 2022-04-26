package bgcache

import (
	"encoding/json"
	"errors"
	instana "github.com/instana/go-sensor"
	newrelicPreModules "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.ibm.com/cloud-sre/iam-authorize/elasticsearch"
	"github.ibm.com/cloud-sre/iam-authorize/iam"
	"reflect"
	"testing"
	"time"
)

var (
	apiKeyOrToken = "1234567890-aaabbbccc_dddeee-AAABBBCCCDDDEEE-"
	currentKeyID int64
	timeToBeAdded = time.Now().Unix() + 1 // this ensures that unit tests can pass without violating bgDurationLimit
)

type testElasticsearch struct {
	BulkIndexImpl func(input interface{}, index string, id string) error
	IndexImpl     func(input interface{}, index string) (elasticsearch.IndexResponse, error)
	SearchImpl    func(input interface{}, index string) (elasticsearch.SearchResponse, error)
}

func (es *testElasticsearch) BulkIndex(input interface{}, index string, id string) error {
	return es.BulkIndexImpl(input, index, id)
}

func (es *testElasticsearch) Index(input interface{}, index string) (elasticsearch.IndexResponse, error) {
	return es.IndexImpl(input, index)
}

func (es *testElasticsearch) Search(input interface{}, index string) (elasticsearch.SearchResponse, error) {
	return es.SearchImpl(input, index)
}

// TestGetCurrentKeyID explicitly initializes currentKeyID
func TestGetCurrentKeyID(t *testing.T) {
	setEnv(t)
	_, currentKeyID, _ = getMasterKey(time.Now().Unix())
	t.Log(currentKeyID)
}

// TestInitBgCache explicitly calls InitBgCache
func TestInitBgCache(t *testing.T) {
	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nil,
		InstanaSensor: nil,
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	InitBgCache(nrWrapperConfig)
}

// TestEncryptForAndDecryptFromElasticsearch tests that a string can be encrypted and decrypted successfully
func TestEncryptForAndDecryptFromElasticsearch(t *testing.T) {
	encryptedApiKeyOrToken, apiKeyOrToKenEncryptionKeyID, _ := encryptForElasticsearch(apiKeyOrToken)
	decryptedApiKeyOrToken, _ := decryptFromElasticsearch(encryptedApiKeyOrToken, apiKeyOrToKenEncryptionKeyID)
	AssertEqual(t, "Decrypted string should be equal to the original string", apiKeyOrToken, decryptedApiKeyOrToken)
	AssertEqual(t, "Encryption key id should equal to the current encryption key id", currentKeyID, apiKeyOrToKenEncryptionKeyID)
}

// TestAddAuthorizationGetElasticsearchFailure tests whether getElasticsearch works in AddAuthorization without being created
// and send data to New Relic if it fails
func TestAddAuthorizationGetElasticsearchFailure(t *testing.T) {
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	enableElasticsearchMux.Lock()
	enableElasticsearch = true
	enableElasticsearchMux.Unlock()

	bgcGot.AddAuthorization("1234567890", "IBMid-1234567890", "fred@ibm.com", "public-iam", "2345678901", "gcor1-segments", "pnp-api-oss.rest.get", 1608070135, false)
}

// TestAddUserGetElasticsearchFailure tests whether getElasticsearch works in AddUser without being created
func TestAddUserGetElasticsearchFailure(t *testing.T) {
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.AddUser("IBMid-1234567890", "fred@ibm.com", "public-iam", "gcor1-segments", "pnp-api-oss.rest.get", 1608070135, false)
}

// TestSyncWithElasticsearchInitializeBgcFailure tests whether InitBgCache initializes bgc successfully
func TestSyncWithElasticsearchInitializeBgcFailure(t *testing.T) {
	bgc = nil

	enableElasticsearchMux.Lock()
	enableElasticsearch = false
	enableElasticsearchMux.Unlock()

	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	syncWithElasticsearch(nrWrapperConfig)

	bgcGot := GetBGCache()
	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
	}

	AssertEqual(t, "bgCache", bgcExpected, bgcGot)
	AssertEqual(t, "enableElasticsearch", false, enableElasticsearch)
}

// TestSyncWithElasticsearchGetElasticsearchFailure tests whether getElasticsearch works in syncWithElasticsearch without being created
func TestSyncWithElasticsearchGetElasticsearchFailure(t *testing.T) {
	bgc = nil

	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	GetBGCache()
	if bgc.monConfig == nil {
		bgc.monConfig = nrWrapperConfig
	}
	syncWithElasticsearch(nrWrapperConfig)

	bgcGot := GetBGCache()
	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	AssertEqual(t, "bgCache", bgcExpected, bgcGot)
	AssertEqual(t, "enableElasticsearch", false, enableElasticsearch)
}

// TestSyncWithElasticsearchNoHit tests whether in-memory cache will be empty when there is no hit from Elasticsearch
func TestSyncWithElasticsearchNoHit(t *testing.T) {
	bgc = nil
	testElasticsearch := &testElasticsearch{
		SearchImpl: func(input interface{}, index string) (elasticsearch.SearchResponse, error) {
			response := elasticsearch.SearchResponse{}
			sampleResponse := `{
			  "took": 1,
			  "timed_out": false,
			  "_shards": {
				"total": 1,
				"successful": 1,
				"skipped": 0,
				"failed": 0
			  },
			  "hits": {
				"total": {
				  "value": 0,
				  "relation": "eq"
				},
				"max_score": null,
				"hits": []
			  }
			}`
			json.Unmarshal([]byte(sampleResponse), &response)

			return response, nil
		},
	}

	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	setElasticsearch(testElasticsearch)
	GetBGCache()
	if bgc.monConfig == nil {
		bgc.monConfig = nrWrapperConfig
	}
	syncWithElasticsearch(nrWrapperConfig)

	bgcGot := GetBGCache()
	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	AssertEqual(t, "bgCache", bgcExpected, bgcGot)
	AssertEqual(t, "enableElasticsearch", true, enableElasticsearch)
}

// TestSyncWithElasticsearchSingleAPIRecordSuccess tests that a single API record can be synced from Elasticsearch
// and be saved to the in-memory cache successfully
func TestSyncWithElasticsearchSingleAPIRecordSuccess(t *testing.T) {
	bgc = nil
	enableElasticsearch = false
	testElasticsearch := &testElasticsearch{
		SearchImpl: func(input interface{}, index string) (elasticsearch.SearchResponse, error) {
			response := elasticsearch.SearchResponse{}
			hit := elasticsearch.SearchResponseHit{}
			sampleRecord := `{
				"_index": "breakglass",
				"_type": "_doc",
				"_id": "api+IBMid-1234567890",
				"_score": 1,
				"_source": {
				  "user": "fred@ibm.com",
				  "isapi": "true",
				  "api_keys": [
					{
				  	"source": "public-iam",
				  	"api_key": "fI4PSunG4C3zvQWYqZ49AUh/ivCkY6PhEu44TU6CisHIWU6ZeHA",
				  	"token": "1s6/Arj7yMXXNC8f9mMiaVxvInbi6TgjcQ4+elj1WNrtByiEUeE",
					"key_id": 2366841599,
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070135
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070135
					}
				]
			  	}
			  }`
			json.Unmarshal([]byte(sampleRecord), &hit)

			var hits [1]elasticsearch.SearchResponseHit
			hits[0] = hit
			response.Hits.Hits = hits[0:]
			return response, nil
		},
	}

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nil,
		InstanaSensor: nil,
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	setElasticsearch(testElasticsearch)
	GetBGCache()
	if bgc.monConfig == nil {
		bgc.monConfig = nrWrapperConfig
	}
	syncWithElasticsearch(nrWrapperConfig)

	bgcGot := GetBGCache()
	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}
	bgcExpected.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": 1608070135, "pnp-api-oss.rest.post": 1608070135}, "tip-hooks-concern": map[string]int64{"tip.rest.get": 1608070135, "tip.rest.post": 1608070135}},
	}

	AssertEqual(t, "bgCache", bgcExpected, bgcGot)
	AssertEqual(t, "enableElasticsearch", true, enableElasticsearch)
}

// TestSyncWithElasticsearchMultipleAPIRecordSuccessForTheSameUser tests that more than one API records can be synced from Elasticsearch
// and be saved to the in-memory cache successfully for the same user
func TestSyncWithElasticsearchMultipleAPIRecordSuccessForTheSameUser(t *testing.T) {
	bgc = nil
	enableElasticsearch = false
	testElasticsearch := &testElasticsearch{
		SearchImpl: func(input interface{}, index string) (elasticsearch.SearchResponse, error) {
			response := elasticsearch.SearchResponse{}
			hit := elasticsearch.SearchResponseHit{}
			sampleRecord := `{
				"_index": "breakglass",
				"_type": "_doc",
				"_id": "api+IBMid-1234567890",
				"_score": 1,
				"_source": {
				  "user": "fred@ibm.com",
				  "isapi": "true",
				  "api_keys": [
					{
				  	"source": "public-iam",
				  	"api_key": "fI4PSunG4C3zvQWYqZ49AUh/ivCkY6PhEu44TU6CisHIWU6ZeHA",
				  	"token": "1s6/Arj7yMXXNC8f9mMiaVxvInbi6TgjcQ4+elj1WNrtByiEUeE",
					"key_id": 2366841599,
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070135
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070135
					},
					{
				  	"source": "public-iam",
				  	"api_key": "FmlzSBMQ10xJ1DsLzQhMeMKf2JeyAEAIE2H2jxlFhXmWfTF+obw",
				  	"token": "1s6/Arj7yMXXNC8f9mMiaVxvInbi6TgjcQ4+elj1WNrtByiEUeE",
					"key_id": 2366841599,
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070135
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070135
					}
				]
			  	}
			  }`
			json.Unmarshal([]byte(sampleRecord), &hit)

			var hits [1]elasticsearch.SearchResponseHit
			hits[0] = hit
			response.Hits.Hits = hits[0:]
			return response, nil
		},
	}

	// Set New Relic app here to mock sending transactions to New Relic
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	setElasticsearch(testElasticsearch)
	GetBGCache()
	if bgc.monConfig == nil {
		bgc.monConfig = nrWrapperConfig
	}
	syncWithElasticsearch(nrWrapperConfig)

	bgcGot := GetBGCache()
	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}
	bgcExpected.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": 1608070135, "pnp-api-oss.rest.post": 1608070135}, "tip-hooks-concern": map[string]int64{"tip.rest.get": 1608070135, "tip.rest.post": 1608070135}},
	}
	bgcExpected.memCacheRecords["0987654321"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": 1608070135, "pnp-api-oss.rest.post": 1608070135}, "tip-hooks-concern": map[string]int64{"tip.rest.get": 1608070135, "tip.rest.post": 1608070135}},
	}

	AssertEqual(t, "bgCache", bgcExpected, bgcGot)
	AssertEqual(t, "enableElasticsearch", true, enableElasticsearch)
}

// TestSyncWithElasticsearchMultipleAPIRecordSuccessForTheSameUserTestOverwrite tests that more than one API records can be synced from Elasticsearch
// and be saved to the in-memory cache successfully for the same user and whether the new record overwrites the old record
// Permissions.Time should only be updated when the time parameter is larger than the existing time in the map
func TestSyncWithElasticsearchMultipleAPIRecordSuccessForTheSameUserTestOverwrite(t *testing.T) {
	bgc = nil
	enableElasticsearch = false
	testElasticsearch := &testElasticsearch{
		SearchImpl: func(input interface{}, index string) (elasticsearch.SearchResponse, error) {
			response := elasticsearch.SearchResponse{}
			hit := elasticsearch.SearchResponseHit{}
			// Time for permission pnp-api-oss.rest.post should be updated since it is larger (1608070136) than 1608070135
			// Time for permission tip.rest.get should not be updated since it is smaller (1608070134) than 1608070135
			sampleRecord := `{
				"_index": "breakglass",
				"_type": "_doc",
				"_id": "api+IBMid-1234567890",
				"_score": 1,
				"_source": {
				  "user": "fred@ibm.com",
				  "isapi": "true",
				  "api_keys": [
					{
				  	"source": "public-iam",
				  	"api_key": "fI4PSunG4C3zvQWYqZ49AUh/ivCkY6PhEu44TU6CisHIWU6ZeHA",
				  	"token": "1s6/Arj7yMXXNC8f9mMiaVxvInbi6TgjcQ4+elj1WNrtByiEUeE",
					"key_id": 2366841599,
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070135
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070135
					},
					{
				  	"source": "private-iam",
				  	"api_key": "BseYCFOz0MFjvN7QhUhFbOPZXbX0uryZKMucUg67sJ86Bo4ek2I",
				  	"token": "hSltdP6wwoYD8gL0N1Z24oHnU/na6NeFtMp16sEii13Kq+3/XDE",
					"key_id": 2366841599,
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070136
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070134
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070136
					}
				]
			  	}
			  }`
			json.Unmarshal([]byte(sampleRecord), &hit)

			var hits [1]elasticsearch.SearchResponseHit
			hits[0] = hit
			response.Hits.Hits = hits[0:]
			return response, nil
		},
	}

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nil,
		InstanaSensor: nil,
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	setElasticsearch(testElasticsearch)
	GetBGCache()
	if bgc.monConfig == nil {
		bgc.monConfig = nrWrapperConfig
	}
	syncWithElasticsearch(nrWrapperConfig)

	bgcGot := GetBGCache()
	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}
	bgcExpected.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "private-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": 1608070135, "pnp-api-oss.rest.post": 1608070136}, "tip-hooks-concern": map[string]int64{"tip.rest.get": 1608070135, "tip.rest.post": 1608070135}},
	}

	AssertEqual(t, "bgCache", bgcExpected, bgcGot)
	AssertEqual(t, "enableElasticsearch", true, enableElasticsearch)
}

// TestSyncWithElasticsearchMultipleAPIRecordSuccessForDifferentUsers tests that more than one API records can be synced from Elasticsearch
// and be saved to the in-memory cache successfully for different users
func TestSyncWithElasticsearchMultipleAPIRecordSuccessForDifferentUsers(t *testing.T) {
	bgc = nil
	enableElasticsearch = false
	testElasticsearch := &testElasticsearch{
		SearchImpl: func(input interface{}, index string) (elasticsearch.SearchResponse, error) {
			response := elasticsearch.SearchResponse{}
			hit1 := elasticsearch.SearchResponseHit{}
			sampleRecord1 := `{
				"_index": "breakglass",
				"_type": "_doc",
				"_id": "api+IBMid-1234567890",
				"_score": 1,
				"_source": {
				  "user": "fred@ibm.com",
				  "isapi": "true",
				  "api_keys": [
					{
				  	"source": "public-iam",
				  	"api_key": "fI4PSunG4C3zvQWYqZ49AUh/ivCkY6PhEu44TU6CisHIWU6ZeHA",
				  	"token": "1s6/Arj7yMXXNC8f9mMiaVxvInbi6TgjcQ4+elj1WNrtByiEUeE",
					"key_id": 2366841599,
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070135
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070135
					}
				]
			  	}
			  }`
			json.Unmarshal([]byte(sampleRecord1), &hit1)

			hit2 := elasticsearch.SearchResponseHit{}
			// With a different user freed@ibm.com
			sampleRecord2 := `{
				"_index": "breakglass",
				"_type": "_doc",
				"_id": "api+IBMid-0123456789",
				"_score": 1,
				"_source": {
				  "user": "freed@ibm.com",
				  "isapi": "true",
				  "api_keys": [
					{
				  	"source": "public-iam",
				  	"api_key": "FmlzSBMQ10xJ1DsLzQhMeMKf2JeyAEAIE2H2jxlFhXmWfTF+obw",
				  	"token": "1s6/Arj7yMXXNC8f9mMiaVxvInbi6TgjcQ4+elj1WNrtByiEUeE",
					"key_id": 2366841599,
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070135
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070135
					}
				]
			  	}
			  }`
			json.Unmarshal([]byte(sampleRecord2), &hit2)

			var hits [2]elasticsearch.SearchResponseHit
			hits[0] = hit1
			hits[1] = hit2

			response.Hits.Hits = hits[0:2]
			return response, nil
		},
	}

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nil,
		InstanaSensor: nil,
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	setElasticsearch(testElasticsearch)
	GetBGCache()
	if bgc.monConfig == nil {
		bgc.monConfig = nrWrapperConfig
	}
	syncWithElasticsearch(nrWrapperConfig)

	bgcGot := GetBGCache()
	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}
	bgcExpected.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": 1608070135, "pnp-api-oss.rest.post": 1608070135}, "tip-hooks-concern": map[string]int64{"tip.rest.get": 1608070135, "tip.rest.post": 1608070135}},
	}
	bgcExpected.memCacheRecords["0987654321"] = bgAuthorization{
		IamID:    "IBMid-0123456789",
		User:     "freed@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": 1608070135, "pnp-api-oss.rest.post": 1608070135}, "tip-hooks-concern": map[string]int64{"tip.rest.get": 1608070135, "tip.rest.post": 1608070135}},
	}

	AssertEqual(t, "bgCache", bgcExpected, bgcGot)
	AssertEqual(t, "enableElasticsearch", true, enableElasticsearch)
}

// TestSyncWithElasticsearchMultipleAPIRecordSuccessForDifferentUsersAPIDecryptionFailure tests that more than one API records can be synced from Elasticsearch
// and be saved to the in-memory cache successfully for different users
// If decryption for APIKey or Token fails, record is discarded and New Relic is notified
func TestSyncWithElasticsearchMultipleAPIRecordSuccessForDifferentUsersAPIDecryptionFailure(t *testing.T) {
	bgc = nil
	enableElasticsearch = false
	testElasticsearch := &testElasticsearch{
		SearchImpl: func(input interface{}, index string) (elasticsearch.SearchResponse, error) {
			response := elasticsearch.SearchResponse{}
			hit1 := elasticsearch.SearchResponseHit{}
			// api_key and token are both bad for decryption
			sampleRecord1 := `{
				"_index": "breakglass",
				"_type": "_doc",
				"_id": "api+IBMid-1234567890",
				"_score": 1,
				"_source": {
				  "user": "fred@ibm.com",
				  "isapi": "true",
				  "api_keys": [
					{
				  	"source": "public-iam",
				  	"api_key": "fI4PSunG4C3zvQWYqZ49AUh/",
				  	"token": "1s6/Arj7yMXXNC8f9mMiaVxv",
					"key_id": 2366841599,
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070135
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070135
					}
				]
			  	}
			  }`
			json.Unmarshal([]byte(sampleRecord1), &hit1)

			hit2 := elasticsearch.SearchResponseHit{}
			// With a different user freed@ibm.com
			// token is bad for decryption
			sampleRecord2 := `{
				"_index": "breakglass",
				"_type": "_doc",
				"_id": "api+IBMid-0123456789",
				"_score": 1,
				"_source": {
				  "user": "freed@ibm.com",
				  "isapi": "true",
				  "api_keys": [
					{
				  	"source": "public-iam",
				  	"api_key": "FmlzSBMQ10xJ1DsLzQhMeMKf2JeyAEAIE2H2jxlFhXmWfTF+obw",
				  	"token": "1s6/Arj7yMXXNC8f9mMi",
					"key_id": 2366841599,
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070135
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070135
					}
				]
			  	}
			  }`
			json.Unmarshal([]byte(sampleRecord2), &hit2)

			var hits [2]elasticsearch.SearchResponseHit
			hits[0] = hit1
			hits[1] = hit2

			response.Hits.Hits = hits[0:2]
			return response, nil
		},
	}

	// Set New Relic app here to mock sending transactions to New Relic
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	setElasticsearch(testElasticsearch)
	GetBGCache()
	if bgc.monConfig == nil {
		bgc.monConfig = nrWrapperConfig
	}
	syncWithElasticsearch(nrWrapperConfig)

	bgcGot := GetBGCache()
	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	// Both records are discarded since decryption fails for both
	AssertEqual(t, "bgCache", bgcExpected, bgcGot)
	AssertEqual(t, "enableElasticsearch", true, enableElasticsearch)
}

// TestSyncWithElasticsearchMultipleAPIRecordSuccessForDifferentUsersGetIamIDFailure tests that more than one API records can be synced from Elasticsearch
// and be saved to the in-memory cache successfully for different users
// If it fails to get iamID, record is discarded and New Relic is notified
func TestSyncWithElasticsearchMultipleAPIRecordSuccessForDifferentUsersGetIamIDFailure(t *testing.T) {
	bgc = nil
	enableElasticsearch = false
	testElasticsearch := &testElasticsearch{
		SearchImpl: func(input interface{}, index string) (elasticsearch.SearchResponse, error) {
			response := elasticsearch.SearchResponse{}
			hit1 := elasticsearch.SearchResponseHit{}
			// api_key and token are both bad for decryption
			sampleRecord1 := `{
				"_index": "breakglass",
				"_type": "_doc",
				"_id": "IBMid-1234567890",
				"_score": 1,
				"_source": {
				  "user": "fred@ibm.com",
				  "isapi": "true",
				  "api_keys": [
					{
				  	"source": "public-iam",
				  	"api_key": "fI4PSunG4C3zvQWYqZ49AUh/ivCkY6PhEu44TU6CisHIWU6ZeHA",
				  	"token": "1s6/Arj7yMXXNC8f9mMiaVxvInbi6TgjcQ4+elj1WNrtByiEUeE",
					"key_id": 2366841599,
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070135
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070135
					}
				]
			  	}
			  }`
			json.Unmarshal([]byte(sampleRecord1), &hit1)

			hit2 := elasticsearch.SearchResponseHit{}
			// With a different user freed@ibm.com
			// token is bad for decryption
			sampleRecord2 := `{
				"_index": "breakglass",
				"_type": "_doc",
				"_id": "freed@ibm.com",
				"_score": 1,
				"_source": {
				  "user": "",
				  "isapi": "false",
				  "api_keys": [
					{
				  	"source": "public-iam",
				  	"api_key": "",
				  	"token": "",
					"key_id": 0,
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070135
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070135
					}
				]
			  	}
			  }`
			json.Unmarshal([]byte(sampleRecord2), &hit2)

			var hits [2]elasticsearch.SearchResponseHit
			hits[0] = hit1
			hits[1] = hit2

			response.Hits.Hits = hits[0:2]
			return response, nil
		},
	}

	// Set New Relic app here to mock sending transactions to New Relic
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	setElasticsearch(testElasticsearch)
	GetBGCache()
	if bgc.monConfig == nil {
		bgc.monConfig = nrWrapperConfig
	}
	syncWithElasticsearch(nrWrapperConfig)

	bgcGot := GetBGCache()
	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	// Both records are discarded since decryption fails for both
	AssertEqual(t, "bgCache", bgcExpected, bgcGot)
	AssertEqual(t, "enableElasticsearch", true, enableElasticsearch)
}

// TestSyncWithElasticsearchSingleAPIRecordFailure tests that records cannot be synced from Elasticsearch
// and fail to be saved to the in-memory cache when there is an error returned from Elasticsearch
func TestSyncWithElasticsearchSingleAPIRecordFailure(t *testing.T) {
	bgc = nil
	enableElasticsearch = false
	testElasticsearch := &testElasticsearch{
		SearchImpl: func(input interface{}, index string) (elasticsearch.SearchResponse, error) {
			response := elasticsearch.SearchResponse{}

			return response, errors.New("an error occurred")
		},
	}

	// Set New Relic app here to mock sending transactions to New Relic
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	setElasticsearch(testElasticsearch)
	GetBGCache()
	if bgc.monConfig == nil {
		bgc.monConfig = nrWrapperConfig
	}
	syncWithElasticsearch(nrWrapperConfig)

	bgcGot := GetBGCache()
	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	AssertEqual(t, "bgCache", bgcExpected, bgcGot)
	AssertEqual(t, "enableElasticsearch", false, enableElasticsearch)
}

// TestSyncWithElasticsearchSingleTokenRecordSuccess tests that a single token record can be synced from Elasticsearch
// and be saved to the in-memory cache successfully
func TestSyncWithElasticsearchSingleTokenRecordSuccess(t *testing.T) {
	bgc = nil
	enableElasticsearch = false
	testElasticsearch := &testElasticsearch{
		SearchImpl: func(input interface{}, index string) (elasticsearch.SearchResponse, error) {
			response := elasticsearch.SearchResponse{}
			hit := elasticsearch.SearchResponseHit{}
			sampleRecord := `{
				"_index": "breakglass",
				"_type": "_doc",
				"_id": "IBMid-1234567890",
				"_score": 1,
				"_source": {
				  "user": "fred@ibm.com",
				  "isapi": "false",
				  "api_keys": [
					{
				  	"source": "public-iam",
				  	"api_key": "",
				  	"token": "",
				  	"resources": [
						{
					  	"resource": "gcor1-segments",
					  	"permissions": [
							{
						  	"permission": "pnp-api-oss.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "pnp-api-oss.rest.post",
						  	"time": 1608070135
							}
					  	]
						},
						{
					  	"resource": "tip-hooks-concern",
					  	"permissions": [
							{
						  	"permission": "tip.rest.get",
						  	"time": 1608070135
							},
							{
						  	"permission": "tip.rest.post",
						  	"time": 1608070135
							}
					  	]
						}
				  	],
				  	"last_updated": 1608070135
					}
				]
			  	}
			  }`
			json.Unmarshal([]byte(sampleRecord), &hit)

			var hits [1]elasticsearch.SearchResponseHit
			hits[0] = hit
			response.Hits.Hits = hits[0:]
			return response, nil
		},
	}

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nil,
		InstanaSensor: nil,
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	setElasticsearch(testElasticsearch)
	GetBGCache()
	if bgc.monConfig == nil {
		bgc.monConfig = nrWrapperConfig
	}
	syncWithElasticsearch(nrWrapperConfig)

	bgcGot := GetBGCache()
	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}
	bgcExpected.userMemCacheRecords["IBMid-1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": 1608070135, "pnp-api-oss.rest.post": 1608070135}, "tip-hooks-concern": map[string]int64{"tip.rest.get": 1608070135, "tip.rest.post": 1608070135}},
	}

	AssertEqual(t, "bgCache", bgcExpected, bgcGot)
	AssertEqual(t, "enableElasticsearch", true, enableElasticsearch)
}

// TestAddAuthorizationSuccessWithEmptyInMemoryCache tests that AddAuthorization function can add or update an API record to in-memory cache,
// find an existing API record from in-memory cache, and bulk index the cache record to Elasticsearch successfully
func TestAddAuthorizationSuccessWithEmptyInMemoryCache(t *testing.T) {
	permission := Permissions{
		Permission: "pnp-api-oss.rest.get",
		Time:       timeToBeAdded,
	}
	resource := Resources{
		Resource:            "gcor1-segments",
		ResourcePermissions: []Permissions {
			permission,
		},
	}

	keyExpected := Key{
		APIKey:      "1234567890",
		Source:      "public-iam",
		Token:       "2345678901",
		KeyID:       currentKeyID,
		ESResources: []Resources{resource},
		LastUpdated: timeToBeAdded,
	}

	esCacheRecordExpected := ESCacheRecord{
		User:    "fred@ibm.com",
		IsAPI:   "true",
		APIKeys: []Key{keyExpected},
	}

	testElasticsearch := &testElasticsearch{
		BulkIndexImpl: func(input interface{}, index string, id string) error {
			AssertEqual(t, "index", breakglassIndex, index)
			AssertEqual(t, "_id", "api+IBMid-1234567890", id)
			esCacheRecord, ok := input.(ESCacheRecord)
			if !ok {
				t.Error("Input is not of type ESCacheRecord")
			}

			AssertEqual(t, "User", esCacheRecordExpected.User, esCacheRecord.User)
			AssertEqual(t, "IsAPI", esCacheRecordExpected.IsAPI, esCacheRecord.IsAPI)
			// The order doesn't matter here since there is only one record
			esCacheRecord.APIKeys[0].APIKey, _ = decryptFromElasticsearch(esCacheRecord.APIKeys[0].APIKey, currentKeyID)
			esCacheRecord.APIKeys[0].Token, _ = decryptFromElasticsearch(esCacheRecord.APIKeys[0].Token, currentKeyID)
			AssertEqual(t, "APIKeys", esCacheRecordExpected.APIKeys, esCacheRecord.APIKeys)

			return nil
		},
	}

	setElasticsearch(testElasticsearch)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nil,
		InstanaSensor: nil,
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.AddAuthorization("1234567890", "IBMid-1234567890", "fred@ibm.com", "public-iam", "2345678901", "gcor1-segments", "pnp-api-oss.rest.get", timeToBeAdded, false)

	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization)}
	bgcExpected.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded}},
	}

	// Looking for an existing record
	AssertEqual(t, "bgCache", bgcExpected.memCacheRecords["1234567890"], *bgcGot.GetAuthorization("1234567890", "gcor1-segments", "pnp-api-oss.rest.get"))

	// Trying to look for a missing record
	var bgcMissingRecordExpected *bgAuthorization
	bgcMissingRecord := bgcGot.GetAuthorization("1111111111", "gcor1-segments", "pnp-api-oss.rest.get")
	AssertEqual(t, "bgCache", bgcMissingRecordExpected, bgcMissingRecord)
}

// TestAddAuthorizationSuccessWithExistingCacheRecords tests that AddAuthorization function can add or update an API record to in-memory cache,
// find an existing API record from in-memory cache, and bulk index the cache record along with the existing records to Elasticsearch successfully.
// And send custom attributes to New Relic when nrApp is not nil
func TestAddAuthorizationSuccessWithExistingCacheRecords(t *testing.T) {
	permission := Permissions{
		Permission: "pnp-api-oss.rest.get",
		Time:       timeToBeAdded,
	}
	permission2 := Permissions{
		Permission: "tip.rest.get",
		Time:       timeToBeAdded,
	}
	resource := Resources{
		Resource:            "gcor1-segments",
		ResourcePermissions: []Permissions {
			permission,
		},
	}
	resource2 := Resources{
		Resource:            "tip-hooks-concern",
		ResourcePermissions: []Permissions {
			permission2,
		},
	}

	keyExpected := Key{
		APIKey:      "1234567890",
		Source:      "public-iam",
		Token:       "2345678901",
		KeyID:       currentKeyID,
		ESResources: []Resources{resource, resource2},
		LastUpdated: timeToBeAdded,
	}

	esCacheRecordExpected := ESCacheRecord{
		User:    "fred@ibm.com",
		IsAPI:   "true",
		APIKeys: []Key{keyExpected},
	}

	testElasticsearch := &testElasticsearch{
		BulkIndexImpl: func(input interface{}, index string, id string) error {
			AssertEqual(t, "index", breakglassIndex, index)
			AssertEqual(t, "_id", "api+IBMid-1234567890", id)
			esCacheRecord, ok := input.(ESCacheRecord)
			if !ok {
				t.Error("Input is not of type ESCacheRecord")
			}

			jsonIn, _ := toJSON(input)
			t.Log(jsonIn)

			AssertEqual(t, "User", esCacheRecordExpected.User, esCacheRecord.User)
			AssertEqual(t, "IsAPI", esCacheRecordExpected.IsAPI, esCacheRecord.IsAPI)
			// Comment this out since the order might be different
			// esCacheRecord.APIKeys[0].APIKey, _ = decryptFromElasticsearch(esCacheRecord.APIKeys[0].APIKey, currentKeyID)
			// esCacheRecord.APIKeys[0].Token, _ = decryptFromElasticsearch(esCacheRecord.APIKeys[0].Token, currentKeyID)
			// AssertEqual(t, "APIKeys", esCacheRecordExpected.APIKeys, esCacheRecord.APIKeys)

			return nil
		},
	}

	setElasticsearch(testElasticsearch)

	// Set New Relic app here to mock sending transactions to New Relic
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded}},
	}

	bgcGot.AddAuthorization("1234567890", "IBMid-1234567890", "fred@ibm.com", "public-iam", "2345678901", "tip-hooks-concern", "tip.rest.get", timeToBeAdded, false)

	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization)}
	bgcExpected.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded}, "tip-hooks-concern": map[string]int64{"tip.rest.get": timeToBeAdded}},
	}

	// Looking for an existing record
	AssertEqual(t, "bgCache", bgcExpected.memCacheRecords["1234567890"], *bgcGot.GetAuthorization("1234567890", "gcor1-segments", "pnp-api-oss.rest.get"))
	AssertEqual(t, "bgCache", bgcExpected.memCacheRecords["1234567890"], *bgcGot.GetAuthorization("1234567890", "tip-hooks-concern", "tip.rest.get"))

	// Trying to look for a missing record
	var bgcMissingRecordExpected *bgAuthorization
	bgcMissingRecord := bgcGot.GetAuthorization("1111111111", "gcor1-segments", "pnp-api-oss.rest.get")
	AssertEqual(t, "bgCache", bgcMissingRecordExpected, bgcMissingRecord)
}

// TestAddAuthorizationSuccessWithExistingCacheRecordsMultipleAPIKeys tests that AddAuthorization function can add or update an API record to in-memory cache,
// find an existing API record from in-memory cache, and bulk index the cache record along with the existing records (multiple APIKeys) to Elasticsearch successfully
func TestAddAuthorizationSuccessWithExistingCacheRecordsMultipleAPIKeys(t *testing.T) {
	permission := Permissions{
		Permission: "pnp-api-oss.rest.get",
		Time:       timeToBeAdded,
	}
	permission2 := Permissions{
		Permission: "tip.rest.get",
		Time:       timeToBeAdded,
	}
	resource := Resources{
		Resource:            "gcor1-segments",
		ResourcePermissions: []Permissions {
			permission,
		},
	}
	resource2 := Resources{
		Resource:            "tip-hooks-concern",
		ResourcePermissions: []Permissions {
			permission2,
		},
	}

	keyExpected1 := Key{
		APIKey:      "1234567890",
		Source:      "private-iam",
		Token:       "yyyyyyyyyy",
		KeyID:       currentKeyID,
		ESResources: []Resources{resource, resource2},
		LastUpdated: timeToBeAdded,
	}

	keyExpected2 := Key{
		APIKey:      "0123456789",
		Source:      "public-iam",
		Token:       "xxxxxxxxxx",
		KeyID:       currentKeyID,
		ESResources: []Resources{resource, resource2},
		LastUpdated: timeToBeAdded,
	}

	esCacheRecordExpected := ESCacheRecord{
		User:    "fred@ibm.com",
		IsAPI:   "true",
		APIKeys: []Key{keyExpected1, keyExpected2},
	}

	testElasticsearch := &testElasticsearch{
		BulkIndexImpl: func(input interface{}, index string, id string) error {
			AssertEqual(t, "index", breakglassIndex, index)
			AssertEqual(t, "_id", "api+IBMid-1234567890", id)
			esCacheRecord, ok := input.(ESCacheRecord)
			if !ok {
				t.Error("Input is not of type ESCacheRecord")
			}

			jsonIn, _ := toJSON(input)
			t.Log(jsonIn)

			AssertEqual(t, "User", esCacheRecordExpected.User, esCacheRecord.User)
			AssertEqual(t, "IsAPI", esCacheRecordExpected.IsAPI, esCacheRecord.IsAPI)
			// Comment this out since the order might be different
			// esCacheRecord.APIKeys[0].APIKey, _ = decryptFromElasticsearch(esCacheRecord.APIKeys[0].APIKey, currentKeyID)
			// esCacheRecord.APIKeys[0].Token, _ = decryptFromElasticsearch(esCacheRecord.APIKeys[0].Token, currentKeyID)
			// esCacheRecord.APIKeys[1].APIKey, _ = decryptFromElasticsearch(esCacheRecord.APIKeys[1].APIKey, currentKeyID)
			// esCacheRecord.APIKeys[1].Token, _ = decryptFromElasticsearch(esCacheRecord.APIKeys[1].Token, currentKeyID)
			// AssertEqual(t, "APIKeys", esCacheRecordExpected.APIKeys, esCacheRecord.APIKeys)

			return nil
		},
	}

	setElasticsearch(testElasticsearch)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nil,
		InstanaSensor: nil,
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded}},
	}
	bgcGot.memCacheRecords["0123456789"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "xxxxxxxxxx",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded}, "tip-hooks-concern": map[string]int64{"tip.rest.get": timeToBeAdded}},
	}

	bgcGot.AddAuthorization("1234567890", "IBMid-1234567890", "fred@ibm.com", "private-iam", "yyyyyyyyyy", "tip-hooks-concern", "tip.rest.get", timeToBeAdded, false)

	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization)}
	bgcExpected.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "private-iam", // source should be updated
		Token:    "yyyyyyyyyy", // token should be updated
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded}, "tip-hooks-concern": map[string]int64{"tip.rest.get": timeToBeAdded}},
	}
	bgcExpected.memCacheRecords["0123456789"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "xxxxxxxxxx",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded}, "tip-hooks-concern": map[string]int64{"tip.rest.get": timeToBeAdded}},
	}

	// Looking for an existing record
	AssertEqual(t, "bgCache", bgcExpected.memCacheRecords["1234567890"], *bgcGot.GetAuthorization("1234567890", "gcor1-segments", "pnp-api-oss.rest.get"))
	AssertEqual(t, "bgCache", bgcExpected.memCacheRecords["1234567890"], *bgcGot.GetAuthorization("1234567890", "tip-hooks-concern", "tip.rest.get"))
	AssertEqual(t, "bgCache", bgcExpected.memCacheRecords["0123456789"], *bgcGot.GetAuthorization("0123456789", "gcor1-segments", "pnp-api-oss.rest.get"))
	AssertEqual(t, "bgCache", bgcExpected.memCacheRecords["0123456789"], *bgcGot.GetAuthorization("0123456789", "tip-hooks-concern", "tip.rest.get"))

	// Trying to look for a missing record
	var bgcMissingRecordExpected *bgAuthorization
	bgcMissingRecord := bgcGot.GetAuthorization("1111111111", "gcor1-segments", "pnp-api-oss.rest.get")
	AssertEqual(t, "bgCache", bgcMissingRecordExpected, bgcMissingRecord)
}

// TestAddAuthorizationToJson tests that AddAuthorization function can convert all the existing cache records
// under an API key from memCacheRecords to a json object ready for Bulk Index
// and bulk index the cache record along with the existing records to Elasticsearch successfully
func TestAddAuthorizationToJson(t *testing.T) {
	permission := Permissions{
		Permission: "pnp-api-oss.rest.get",
		Time:       timeToBeAdded,
	}
	permission2 := Permissions{
		Permission: "pnp-api-oss.rest.post",
		Time:       timeToBeAdded,
	}
	permission3 := Permissions{
		Permission: "tip.rest.get",
		Time:       timeToBeAdded,
	}
	permission4 := Permissions{
		Permission: "tip.rest.post",
		Time:       timeToBeAdded,
	}
	resource := Resources{
		Resource:            "gcor1-segments",
		ResourcePermissions: []Permissions {
			permission,
			permission2,
		},
	}
	resource2 := Resources{
		Resource:            "tip-hooks-concern",
		ResourcePermissions: []Permissions {
			permission3,
			permission4,
		},
	}

	keyExpected := Key{
		APIKey:      "1234567890",
		Source:      "public-iam",
		Token:       "2345678901",
		KeyID:       currentKeyID,
		ESResources: []Resources{resource, resource2},
		LastUpdated: timeToBeAdded,
	}

	esCacheRecordExpected := ESCacheRecord{
		User:    "fred@ibm.com",
		IsAPI:   "true",
		APIKeys: []Key{keyExpected},
	}

	testElasticsearch := &testElasticsearch{
		BulkIndexImpl: func(input interface{}, index string, id string) error {
			AssertEqual(t, "index", breakglassIndex, index)
			AssertEqual(t, "_id", "api+IBMid-1234567890", id)
			esCacheRecord, ok := input.(ESCacheRecord)
			if !ok {
				t.Error("Input is not of type ESCacheRecord")
			}

			jsonIn, _ := toJSON(input)
			t.Log(jsonIn)

			AssertEqual(t, "User", esCacheRecordExpected.User, esCacheRecord.User)
			AssertEqual(t, "IsAPI", esCacheRecordExpected.IsAPI, esCacheRecord.IsAPI)
			// Comment this out since the order might be different
			// esCacheRecord.APIKeys[0].APIKey, _ = decryptFromElasticsearch(esCacheRecord.APIKeys[0].APIKey, currentKeyID)
			// esCacheRecord.APIKeys[0].Token, _ = decryptFromElasticsearch(esCacheRecord.APIKeys[0].Token, currentKeyID)
			// AssertEqual(t, "APIKeys", esCacheRecordExpected.APIKeys, esCacheRecord.APIKeys)

			return nil
		},
	}

	setElasticsearch(testElasticsearch)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nil,
		InstanaSensor: nil,
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded, "pnp-api-oss.rest.post": timeToBeAdded}, "tip-hooks-concern": map[string]int64{"tip.rest.get": timeToBeAdded}},
	}

	bgcGot.AddAuthorization("1234567890", "IBMid-1234567890", "fred@ibm.com", "public-iam", "2345678901", "tip-hooks-concern", "tip.rest.post", timeToBeAdded, false)
}

// TestAddAuthorizationFailure tests that AddAuthorization function can add or update an API record to in-memory cache and
// find an existing API record from in-memory cache successfully but fails to bulk index the cache record to Elasticsearch
func TestAddAuthorizationFailure(t *testing.T) {
	testElasticsearch := &testElasticsearch{
		BulkIndexImpl: func(input interface{}, index string, id string) error {
			AssertEqual(t, "index", breakglassIndex, index)
			return errors.New("an error occurred")
		},
	}

	setElasticsearch(testElasticsearch)

	// Set New Relic app here to mock sending transactions to New Relic
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.AddAuthorization("1234567890", "IBMid-1234567890", "fred@ibm.com", "public-iam", "2345678901", "gcor1-segments", "pnp-api-oss.rest.get", timeToBeAdded, false)

	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization)}
	bgcExpected.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded}},
	}

	// Looking for an existing record
	AssertEqual(t, "bgCache", bgcExpected.memCacheRecords["1234567890"], *bgcGot.GetAuthorization("1234567890", "gcor1-segments", "pnp-api-oss.rest.get"))

	// Trying to look for a missing record
	var bgcMissingRecordExpected *bgAuthorization
	bgcMissingRecord := bgcGot.GetAuthorization("1111111111", "gcor1-segments", "pnp-api-oss.rest.get")
	AssertEqual(t, "bgCache", bgcMissingRecordExpected, bgcMissingRecord)
}

// TestAddAuthorizationFailureWithEmptyIamID tests that AddAuthorization function can still add or update an API record to in-memory cache,
// but doesn't bulk index the cache record to Elasticsearch when iamID is empty
func TestAddAuthorizationFailureWithEmptyIamID(t *testing.T) {
	// Set New Relic app here to mock sending transactions to New Relic
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.AddAuthorization("1234567890", "", "fred@ibm.com", "public-iam", "2345678901", "gcor1-segments", "pnp-api-oss.rest.get", timeToBeAdded, false)

	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization)}
	bgcExpected.memCacheRecords["1234567890"] = bgAuthorization{
		IamID:    "",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "2345678901",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded}},
	}

	// Looking for an existing record
	AssertEqual(t, "bgCache", bgcExpected.memCacheRecords["1234567890"], *bgcGot.GetAuthorization("1234567890", "gcor1-segments", "pnp-api-oss.rest.get"))

	// Trying to look for a missing record
	var bgcMissingRecordExpected *bgAuthorization
	bgcMissingRecord := bgcGot.GetAuthorization("1111111111", "gcor1-segments", "pnp-api-oss.rest.get")
	AssertEqual(t, "bgCache", bgcMissingRecordExpected, bgcMissingRecord)
}

// TestAddUserFailureWithEmptyIamID tests that AddUser function stops adding or updating a Token record to in-memory cache,
// and doesn't bulk index the cache record to Elasticsearch when iamID is empty
func TestAddUserFailureWithEmptyIamID(t *testing.T) {
	// Set New Relic app here to mock sending transactions to New Relic
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.AddUser("", "fred@ibm.com", "public-iam", "gcor1-segments", "pnp-api-oss.rest.get", timeToBeAdded, false)

	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization)}

	// Trying to look for a missing record
	_, isUserAuthorizedExpected := bgcExpected.userMemCacheRecords[""]
	AssertEqual(t, "isUserAuthorized", isUserAuthorizedExpected, bgcGot.IsUserAuthorized("", "public-iam", "gcor1-segments", "pnp-api-oss.rest.get"))
}

// TestAddUserSuccessWithEmptyInMemoryCache tests that AddUser function can add or update a Token record to in-memory cache,
// check if an user is authorized from in-memory cache, and bulk index the cache record to Elasticsearch successfully
func TestAddUserSuccessWithEmptyInMemoryCache(t *testing.T) {
	permission := Permissions{
		Permission: "pnp-api-oss.rest.get",
		Time:       timeToBeAdded,
	}
	resource := Resources{
		Resource:            "gcor1-segments",
		ResourcePermissions: []Permissions {
			permission,
		},
	}

	keyExpected := Key{
		APIKey:      "",
		Source:      "public-iam",
		Token:       "",
		KeyID:       0,
		ESResources: []Resources{resource},
		LastUpdated: timeToBeAdded,
	}

	esCacheRecordExpected := ESCacheRecord{
		User:    "fred@ibm.com",
		IsAPI:   "false",
		APIKeys: []Key{keyExpected},
	}

	testElasticsearch := &testElasticsearch{
		BulkIndexImpl: func(input interface{}, index string, id string) error {
			AssertEqual(t, "index", breakglassIndex, index)
			AssertEqual(t, "_id", "IBMid-1234567890", id)
			esCacheRecord, ok := input.(ESCacheRecord)
			if !ok {
				t.Error("Input is not of type ESCacheRecord")
			}

			AssertEqual(t, "User", esCacheRecordExpected.User, esCacheRecord.User)
			AssertEqual(t, "IsAPI", esCacheRecordExpected.IsAPI, esCacheRecord.IsAPI)
			// The order doesn't matter here since there is only one record
			AssertEqual(t, "APIKeys", esCacheRecordExpected.APIKeys, esCacheRecord.APIKeys)

			return nil
		},
	}

	setElasticsearch(testElasticsearch)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nil,
		InstanaSensor: nil,
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.AddUser("IBMid-1234567890", "fred@ibm.com", "public-iam", "gcor1-segments", "pnp-api-oss.rest.get", timeToBeAdded, false)

	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization)}
	bgcExpected.userMemCacheRecords["IBMid-1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": 1608070135}},
	}

	// Looking for an existing record
	_, isUserAuthorizedExpected := bgcExpected.userMemCacheRecords["IBMid-1234567890"]
	AssertEqual(t, "isUserAuthorized", isUserAuthorizedExpected, bgcGot.IsUserAuthorized("IBMid-1234567890", "public-iam", "gcor1-segments", "pnp-api-oss.rest.get"))

	// Trying to look for a missing record
	isUserAuthorizedExpected = false
	AssertEqual(t, "isUserAuthorized", isUserAuthorizedExpected, bgcGot.IsUserAuthorized("IBMid-0000000000", "public-iam", "gcor1-segments", "pnp-api-oss.rest.get"))
}

// TestAddUserSuccessWithExistingCacheRecords tests that AddUser function can add or update a Token record to in-memory cache,
// check if an user is authorized from in-memory cache, and bulk index the cache record along with the existing records to Elasticsearch successfully
// It also tests if the LastUpdated field is the latest among all Permissions.Time's inside one Key
func TestAddUserSuccessWithExistingCacheRecords(t *testing.T) {
	permission := Permissions{
		Permission: "pnp-api-oss.rest.get",
		Time:       timeToBeAdded+2,
	}
	permission2 := Permissions{
		Permission: "tip.rest.get",
		Time:       timeToBeAdded,
	}
	resource := Resources{
		Resource:            "gcor1-segments",
		ResourcePermissions: []Permissions {
			permission,
		},
	}
	resource2 := Resources{
		Resource:            "tip-hooks-concern",
		ResourcePermissions: []Permissions {
			permission2,
		},
	}

	// LastUpdated needs to be the largest Time
	keyExpected := Key{
		APIKey:      "",
		Source:      "public-iam",
		Token:       "",
		KeyID:       0,
		ESResources: []Resources{resource, resource2},
		LastUpdated: timeToBeAdded+2,
	}

	esCacheRecordExpected := ESCacheRecord{
		User:    "fred@ibm.com",
		IsAPI:   "false",
		APIKeys: []Key{keyExpected},
	}

	testElasticsearch := &testElasticsearch{
		BulkIndexImpl: func(input interface{}, index string, id string) error {
			AssertEqual(t, "index", breakglassIndex, index)
			AssertEqual(t, "_id", "IBMid-1234567890", id)
			esCacheRecord, ok := input.(ESCacheRecord)
			if !ok {
				t.Error("Input is not of type ESCacheRecord")
			}

			AssertEqual(t, "User", esCacheRecordExpected.User, esCacheRecord.User)
			AssertEqual(t, "IsAPI", esCacheRecordExpected.IsAPI, esCacheRecord.IsAPI)
			// Comment this out since the order might be different
			// AssertEqual(t, "APIKeys", esCacheRecordExpected.APIKeys, esCacheRecord.APIKeys)

			return nil
		},
	}

	setElasticsearch(testElasticsearch)

	// Set New Relic app here to mock sending transactions to New Relic
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.userMemCacheRecords["IBMid-1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded+2}},
	}

	bgcGot.AddUser("IBMid-1234567890", "fred@ibm.com", "public-iam", "tip-hooks-concern", "tip.rest.get", timeToBeAdded, false)

	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization)}
	bgcExpected.userMemCacheRecords["IBMid-1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded+2}, "tip-hooks-concern": map[string]int64{"tip.rest.get": timeToBeAdded}},
	}

	// Looking for an existing record
	_, isUserAuthorizedExpected := bgcExpected.userMemCacheRecords["IBMid-1234567890"]
	AssertEqual(t, "isUserAuthorized", isUserAuthorizedExpected, bgcGot.IsUserAuthorized("IBMid-1234567890", "public-iam", "gcor1-segments", "pnp-api-oss.rest.get"))
	AssertEqual(t, "isUserAuthorized", isUserAuthorizedExpected, bgcGot.IsUserAuthorized("IBMid-1234567890", "public-iam", "tip-hooks-concern", "tip.rest.get"))

	// Trying to look for a missing record
	isUserAuthorizedExpected = false
	AssertEqual(t, "isUserAuthorized", isUserAuthorizedExpected, bgcGot.IsUserAuthorized("IBMid-0000000000", "public-iam", "gcor1-segments", "pnp-api-oss.rest.get"))
}

// TestAddUserToJson tests that AddUser function can convert all the existing cache records
// under a token from userMemCacheRecords to a json object ready for Bulk Index
// and bulk index the cache record along with the existing records to Elasticsearch successfully
func TestAddUserToJson(t *testing.T) {
	permission := Permissions{
		Permission: "pnp-api-oss.rest.get",
		Time:       timeToBeAdded,
	}
	permission2 := Permissions{
		Permission: "pnp-api-oss.rest.post",
		Time:       timeToBeAdded,
	}
	permission3 := Permissions{
		Permission: "tip.rest.get",
		Time:       timeToBeAdded,
	}
	permission4 := Permissions{
		Permission: "tip.rest.post",
		Time:       timeToBeAdded,
	}
	resource := Resources{
		Resource:            "gcor1-segments",
		ResourcePermissions: []Permissions {
			permission,
			permission2,
		},
	}
	resource2 := Resources{
		Resource:            "tip-hooks-concern",
		ResourcePermissions: []Permissions {
			permission3,
			permission4,
		},
	}

	keyExpected := Key{
		APIKey:      "",
		Source:      "public-iam",
		Token:       "",
		KeyID:       0,
		ESResources: []Resources{resource, resource2},
		LastUpdated: timeToBeAdded,
	}

	esCacheRecordExpected := ESCacheRecord{
		User:    "fred@ibm.com",
		IsAPI:   "false",
		APIKeys: []Key{keyExpected},
	}

	testElasticsearch := &testElasticsearch{
		BulkIndexImpl: func(input interface{}, index string, id string) error {
			AssertEqual(t, "index", breakglassIndex, index)
			AssertEqual(t, "_id", "IBMid-1234567890", id)
			esCacheRecord, ok := input.(ESCacheRecord)
			if !ok {
				t.Error("Input is not of type ESCacheRecord")
			}

			jsonIn, _ := toJSON(input)
			t.Log(jsonIn)

			AssertEqual(t, "User", esCacheRecordExpected.User, esCacheRecord.User)
			AssertEqual(t, "IsAPI", esCacheRecordExpected.IsAPI, esCacheRecord.IsAPI)
			// Comment this out since the order might be different
			// AssertEqual(t, "APIKeys", esCacheRecordExpected.APIKeys, esCacheRecord.APIKeys)

			return nil
		},
	}

	setElasticsearch(testElasticsearch)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nil,
		InstanaSensor: nil,
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.userMemCacheRecords["IBMid-1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded, "pnp-api-oss.rest.post": timeToBeAdded}, "tip-hooks-concern": map[string]int64{"tip.rest.get": timeToBeAdded}},
	}

	bgcGot.AddUser("IBMid-1234567890", "fred@ibm.com", "public-iam", "tip-hooks-concern", "tip.rest.post", timeToBeAdded, false)
}

// TestAddUserFailure tests that AddUser function can add or update a Token record to in-memory cache and
// check if an user is authorized from in-memory cache, but fails to bulk index the cache record to Elasticsearch
func TestAddUserFailure(t *testing.T) {
	testElasticsearch := &testElasticsearch{
		BulkIndexImpl: func(input interface{}, index string, id string) error {
			AssertEqual(t, "index", breakglassIndex, index)
			return errors.New("an error occurred")
		},
	}

	setElasticsearch(testElasticsearch)

	// Set New Relic app here to mock sending transactions to New Relic
	nrConfig := newrelicPreModules.Config{
		AppName: "fake app",
	}

	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("fake app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)

	nrWrapperConfig := &iam.NRWrapperConfig{
		NRApp: nrApp,
		NRAppV3: nrAppV3,
		InstanaSensor: instana.NewSensor("fake app"),
		Environment: "fake-env",
		Region: "fake-region",
		SourceServiceName: "fake-service",
	}

	bgcGot := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization),
		monConfig: nrWrapperConfig,
	}

	bgcGot.AddUser("IBMid-1234567890", "fred@ibm.com", "public-iam", "gcor1-segments", "pnp-api-oss.rest.get", timeToBeAdded, false)

	bgcExpected := &bgCache{
		memCacheRecords: make(map[string]bgAuthorization),
		userMemCacheRecords: make(map[string]bgAuthorization)}
	bgcExpected.userMemCacheRecords["IBMid-1234567890"] = bgAuthorization{
		IamID:    "IBMid-1234567890",
		User:     "fred@ibm.com",
		Source:   "public-iam",
		Token:    "",
		resource: bgScope{"gcor1-segments": map[string]int64{"pnp-api-oss.rest.get": timeToBeAdded}},
	}

	// Looking for an existing record
	_, isUserAuthorizedExpected := bgcExpected.userMemCacheRecords["IBMid-1234567890"]
	AssertEqual(t, "isUserAuthorized", isUserAuthorizedExpected, bgcGot.IsUserAuthorized("IBMid-1234567890", "public-iam", "gcor1-segments", "pnp-api-oss.rest.get"))

	// Trying to look for a missing record
	isUserAuthorizedExpected = false
	AssertEqual(t, "isUserAuthorized", isUserAuthorizedExpected, bgcGot.IsUserAuthorized("IBMid-0000000000", "public-iam", "gcor1-segments", "pnp-api-oss.rest.get"))
}

// AssertEqual reports a testing failure when a given actual item (interface{}) is not equal to the expected value
func AssertEqual(t *testing.T, item string, expected, actual interface{}) {
	if actual == nil {
		if expected != nil {
			t.Logf("%s %s mismatch - expected %T %v got %T %v", t.Name(), item, expected, expected, actual, actual)
			t.Fail()
		}
	} else if reflect.TypeOf(actual).Comparable() && reflect.TypeOf(actual).Kind() != reflect.Ptr {
		// Note that pointers are comparable but this doesn't apply to comparing de-referenced values
		if expected != actual {
			t.Logf("%s %s mismatch - expected %T %v got %T %v", t.Name(), item, expected, expected, actual, actual)
			t.Fail()
		}
	} else {
		if !reflect.DeepEqual(expected, actual) {
			t.Logf("%s %s mismatch - expected %T %v got %T %v", t.Name(), item, expected, expected, actual, actual)
			t.Fail()
		}
	}
}

// toJSON returns JSON representation of the provided interface
func toJSON(j interface{}) (string, error) {
	result := ""
	var out []byte
	var err error
	out, err = json.MarshalIndent(j, "", "    ")
	if err == nil {
		result = string(out)
	}
	return result, err
}
