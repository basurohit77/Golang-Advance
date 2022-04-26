package pep

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubjectPermutations_requirement1(t *testing.T) {
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getBulkPermitResponderWithResourceObligations(),
		getBulkDenyResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment:     Staging,
		APIKey:          os.Getenv("API_KEY"),
		LogLevel:        LevelError,
		CacheDefaultTTL: cacheTTL,
		HTTPClient:      mocker.httpClient,
	}

	runTest := func(t *testing.T, config Config, request Request) {
		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}

		t.Run("call #0, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#0-non-cached-permit"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #1, same request", func(t *testing.T) {
			trace := "txid-call#1-same-request"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #2, same request with different supported action", func(t *testing.T) {
			trace := "txid-call#2-same-request-with-different-action"
			requests := &Requests{
				{
					"action":   "cloud-object-storage.object.copy",
					"resource": request["resource"],
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with different action should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call with different action should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #3, same request with missing scope", func(t *testing.T) {
			trace := "txid-call#3-same-request-missing-scope"
			subject := Attributes{
				"id": "IBMid-270003GUSX",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with missing scope should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with missing scope should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #4, same request with different scope", func(t *testing.T) {
			trace := "txid-call#4-same-request-different-scope"
			subject := Attributes{
				"id":    "IBMid-270003GUSX",
				"scope": "ibm sso",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different scope should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different scope should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #5, same request with missing id", func(t *testing.T) {
			trace := "txid-call#5-same-request-missing-id"
			subject := Attributes{
				"scope": "ibm openid",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with missing id should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with missing id should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #6, same request with different id", func(t *testing.T) {
			trace := "txid-call#6-same-request-different-id"
			subject := Attributes{
				"id":    "IBMid-457779ABCD",
				"scope": "ibm openid",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different id should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different id should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})
	}

	runTest(t, config, getRequestWithAccountAndService(t))
}

func TestSubjectPermutations_requirement2(t *testing.T) {
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getBulkPermitResponderWithSubjectID(),
		getBulkDenyResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment:     Staging,
		APIKey:          os.Getenv("API_KEY"),
		LogLevel:        LevelError,
		CacheDefaultTTL: cacheTTL,
		HTTPClient:      mocker.httpClient,
	}

	runTest := func(t *testing.T, config Config, request Request) {
		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}

		t.Run("call #0, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#0-non-cached-permit"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #1, same request", func(t *testing.T) {
			trace := "txid-call#1-same-request"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #2, same request with different supported action", func(t *testing.T) {
			trace := "txid-call#2-same-request-with-different-action"
			requests := &Requests{
				{
					"action":   "cloud-object-storage.object.copy",
					"resource": request["resource"],
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with different action should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call with different action should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #3, same request with added scope", func(t *testing.T) {
			trace := "txid-call#3-same-request-with-added-scope"
			subject := Attributes{
				"id":    "IBMid-270003GUSX",
				"scope": "ibm openid",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with added scope should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call with added scope should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #4, same request with missing id", func(t *testing.T) {
			trace := "txid-call#4-same-request-missing-id"
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  Attributes{},
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with missing id should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with missing id should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #5, same request with different id", func(t *testing.T) {
			trace := "txid-call#5-same-request-different-id"
			subject := Attributes{
				"id": "IBMid-457779ABCD",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different id should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different id should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})
	}

	runTest(t, config, getRequestWithOnlySubjectID(t))
}

func getRequestWithOnlySubjectID(t *testing.T) Request {
	subject := Attributes{
		"id": "IBMid-270003GUSX",
	}
	action := "cloud-object-storage.object.get"
	resource := Attributes{
		"accountId":   "12345",
		"serviceName": "cloud-object-storage",
	}

	request := Request{
		"action":   action,
		"resource": resource,
		"subject":  subject,
	}

	return request
}

func getBulkPermitResponderWithSubjectID() JSONMockerFunc {
	mockedJSON := `
	{
		"decisions": [
			{
				"decision": "Permit",
				"obligation": {
					"actions": [
						"cloud-object-storage.object.get",
						"cloud-object-storage.object.put_acl",
						"cloud-object-storage.bucket.get_metrics_monitoring",
						"cloud-object-storage.object.copy_get_version",
						"cloud-object-storage.object.post_md",
						"cloud-object-storage.object.get_tagging",
						"cloud-object-storage.bucket.put_lifecycle",
						"cloud-object-storage.object.copy",
						"cloud-object-storage.object.copy_part",
						"cloud-object-storage.bucket.get_basic",
						"cloud-object-storage.bucket.list_crk_id",
						"cloud-object-storage.object.head"
					],
					"maxCacheAgeSeconds": 600,
					"subject": {
						"attributes": {
							"id": "IBMid-270003GUSX"
						}
					},
					"resource": {
						"attributes": {
							"serviceName": "cloud-object-storage",
							"accountId": "12345"
						}
					}
				}
			}
		],
		"cacheKeyPattern": {
			"order": [
				"subject",
				"resource",
				"action"
			],
			"subject": [
				[
					"id"
				],
				[
					"id",
					"scope"
				]
			],
			"resource": [
				[],
				[
					"serviceName"
				],
				[
					"serviceName",
					"accountId"
				],
				[
					"serviceName",
					"accountId",
					"serviceInstance"
				],
				[
					"serviceName",
					"accountId",
					"serviceInstance",
					"resourceType"
				],
				[
					"serviceName",
					"accountId",
					"serviceInstance",
					"resourceType",
					"resource"
				]
			]
		}
	}
	`
	responder := func(req *http.Request) mockedResponse {
		return mockedResponse{
			json:       mockedJSON,
			statusCode: 200,
		}
	}
	return responder
}
