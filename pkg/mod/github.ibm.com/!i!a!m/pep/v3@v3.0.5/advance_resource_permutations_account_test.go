package pep

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvanceResourcePermutations_accountPolicyRequirement4(t *testing.T) {
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

		t.Run("call #3, same request with different serviceInstance", func(t *testing.T) {
			trace := "txid-call#3-same-request-different-serviceInstance"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-test-bucket-01",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with different serviceInstance should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call with different serviceInstance should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #4, same request with unsupported action", func(t *testing.T) {
			trace := "txid-call#4-same-request-with-unsupported-action"
			requests := &Requests{
				{
					"action":   "cloud-object-storage.unsupported.action",
					"resource": request["resource"],
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with unsupported action should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with unsupported action should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #5, same request with different serviceName", func(t *testing.T) {
			trace := "txid-call#5-same-request-with-different-serviceName"
			resource := Attributes{
				"accountId":   "12345",
				"serviceName": "gopep",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different serviceName should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different serviceName should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #6, same request with different accountId", func(t *testing.T) {
			trace := "txid-call#6-same-request-with-different-accountId"
			resource := Attributes{
				"accountId":   "654321",
				"serviceName": "cloud-object-storage",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different accountId should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different accountId should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})
	}

	runTest(t, config, getRequestWithAccountAndService(t))
}

func TestAdvanceResourcePermutations_accountPolicyRequirement5(t *testing.T) {
	cacheTTL := time.Duration(40) * time.Millisecond
	responders := []JSONMockerFunc{
		getBulkPermitResponderWithServiceInstance(),
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

		t.Run("call #3, same request with different resource", func(t *testing.T) {
			trace := "txid-call#3-same-request-different-resource"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-test-bucket-01",
				"resource":        "some-cos-resource",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with different resource should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call with different resource should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #4, same request with different serviceInstance", func(t *testing.T) {
			trace := "txid-call#4-same-request-different-serviceInstance"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-another-cos-bucket-01",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different serviceInstance should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different serviceInstance should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #5, same request with different serviceName", func(t *testing.T) {
			trace := "txid-call#5-same-request-with-different-serviceName"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "gopep",
				"serviceInstance": "my-test-bucket-01",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different serviceName should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different serviceName should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #6, same request with different accountId", func(t *testing.T) {
			trace := "txid-call#6-same-request-with-different-accountId"
			resource := Attributes{
				"accountId":       "654321",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-test-bucket-01",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different accountId should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different accountId should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #7, same request with unsupported action", func(t *testing.T) {
			trace := "txid-call#7-same-request-with-unsupported-action"
			requests := &Requests{
				{
					"action":   "cloud-object-storage.unsupported.action",
					"resource": request["resource"],
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with unsupported action should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with unsupported action should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})
	}

	runTest(t, config, getRequestWithServiceInstance(t))
}

func TestAdvanceResourcePermutations_accountPolicyRequirement6(t *testing.T) {
	cacheTTL := time.Duration(40) * time.Millisecond
	responders := []JSONMockerFunc{
		getBulkPermitResponderWithResourceType(),
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

		t.Run("call #3, same request with different resource", func(t *testing.T) {
			trace := "txid-call#3-same-request-different-resource"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-test-bucket-01",
				"resourceType":    "bucket",
				"resource":        "some-cos-resource",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with different resource should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call with different resource should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #4, same request with different resourceType", func(t *testing.T) {
			trace := "txid-call#4-same-request-different-resourceType"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-test-bucket-01",
				"resourceType":    "some-other-type",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different resourceType should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different resourceType should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #5, same request with different serviceInstance", func(t *testing.T) {
			trace := "txid-call#5-same-request-different-serviceInstance"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-another-cos-bucket-01",
				"resourceType":    "bucket",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different serviceInstance should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different serviceInstance should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #6, same request with different serviceName", func(t *testing.T) {
			trace := "txid-call#6-same-request-with-different-serviceName"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "gopep",
				"serviceInstance": "my-another-cos-bucket-01",
				"resourceType":    "bucket",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different serviceName should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different serviceName should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #7, same request with different accountId", func(t *testing.T) {
			trace := "txid-call#7-same-request-with-different-accountId"
			resource := Attributes{
				"accountId":       "654321",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-test-bucket-01",
				"resourceType":    "bucket",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different accountId should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different accountId should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #8, same request with unsupported action", func(t *testing.T) {
			trace := "txid-call#8-same-request-with-unsupported-action"
			requests := &Requests{
				{
					"action":   "cloud-object-storage.unsupported.action",
					"resource": request["resource"],
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with unsupported action should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with unsupported action should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})
	}

	runTest(t, config, getRequestWithResourceType(t))
}

func TestAdvanceResourcePermutations_accountPolicyRequirement7(t *testing.T) {
	cacheTTL := time.Duration(40) * time.Millisecond
	responders := []JSONMockerFunc{
		getBulkPermitResponderWithResource(),
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

		t.Run("call #3, same request with custom attribute", func(t *testing.T) {
			trace := "txid-call#3-same-request-custom-attribute"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-test-bucket-01",
				"resourceType":    "bucket",
				"resource":        "my-cos-resource",
				"resourceGroup":   "my-group",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with custom attributee should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call with custom attribute should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #4, same request with different resource", func(t *testing.T) {
			trace := "txid-call#4-same-request-different-resource"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-test-bucket-01",
				"resourceType":    "bucket",
				"resource":        "some-cos-resource",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different resource should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different resource should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #5, same request with different resourceType", func(t *testing.T) {
			trace := "txid-call#5-same-request-different-resourceType"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-test-bucket-01",
				"resourceType":    "some-other-type",
				"resource":        "my-cos-resource",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different resourceType should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different resourceType should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #6, same request with different serviceInstance", func(t *testing.T) {
			trace := "txid-call#6-same-request-different-serviceInstance"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-another-cos-bucket-01",
				"resourceType":    "bucket",
				"resource":        "my-cos-resource",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different serviceInstance should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different serviceInstance should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #7, same request with different serviceName", func(t *testing.T) {
			trace := "txid-call#7-same-request-with-different-serviceName"
			resource := Attributes{
				"accountId":       "12345",
				"serviceName":     "gopep",
				"serviceInstance": "my-test-bucket-01",
				"resourceType":    "bucket",
				"resource":        "my-cos-resource",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different serviceName should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different serviceName should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #8, same request with different accountId", func(t *testing.T) {
			trace := "txid-call#8-same-request-with-different-accountId"
			resource := Attributes{
				"accountId":       "654321",
				"serviceName":     "cloud-object-storage",
				"serviceInstance": "my-test-bucket-01",
				"resourceType":    "bucket",
				"resource":        "my-cos-resource",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": resource,
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with different accountId should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different accountId should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #9, same request with unsupported action", func(t *testing.T) {
			trace := "txid-call#9-same-request-with-unsupported-action"
			requests := &Requests{
				{
					"action":   "cloud-object-storage.unsupported.action",
					"resource": request["resource"],
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with unsupported action should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with unsupported action should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})
	}

	runTest(t, config, getRequestWithResource(t))
}

func getRequestWithAccountAndService(t *testing.T) Request {
	subject := Attributes{
		"id":    "IBMid-270003GUSX",
		"scope": "ibm openid",
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

func getRequestWithServiceInstance(t *testing.T) Request {
	subject := Attributes{
		"id":    "IBMid-270003GUSX",
		"scope": "ibm openid",
	}
	action := "cloud-object-storage.object.get"
	resource := Attributes{
		"accountId":       "12345",
		"serviceName":     "cloud-object-storage",
		"serviceInstance": "my-test-bucket-01",
	}

	request := Request{
		"action":   action,
		"resource": resource,
		"subject":  subject,
	}

	return request
}

func getRequestWithResourceType(t *testing.T) Request {
	subject := Attributes{
		"id":    "IBMid-270003GUSX",
		"scope": "ibm openid",
	}
	action := "cloud-object-storage.object.get"
	resource := Attributes{
		"accountId":       "12345",
		"serviceName":     "cloud-object-storage",
		"serviceInstance": "my-test-bucket-01",
		"resourceType":    "bucket",
	}

	request := Request{
		"action":   action,
		"resource": resource,
		"subject":  subject,
	}

	return request
}

func getRequestWithResource(t *testing.T) Request {
	subject := Attributes{
		"id":    "IBMid-270003GUSX",
		"scope": "ibm openid",
	}
	action := "cloud-object-storage.object.get"
	resource := Attributes{
		"accountId":       "12345",
		"serviceName":     "cloud-object-storage",
		"serviceInstance": "my-test-bucket-01",
		"resourceType":    "bucket",
		"resource":        "my-cos-resource",
	}

	request := Request{
		"action":   action,
		"resource": resource,
		"subject":  subject,
	}

	return request
}

func getBulkPermitResponderWithServiceInstance() JSONMockerFunc {
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
							"scope": "ibm openid",
							"id": "IBMid-270003GUSX"
						}
					},
					"resource": {
						"attributes": {
							"accountId": "12345",
							"serviceName": "cloud-object-storage",
							"serviceInstance": "my-test-bucket-01"
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

func getBulkPermitResponderWithResourceType() JSONMockerFunc {
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
							"scope": "ibm openid",
							"id": "IBMid-270003GUSX"
						}
					},
					"resource": {
						"attributes": {
							"accountId": "12345",
							"serviceName": "cloud-object-storage",
							"serviceInstance": "my-test-bucket-01",
							"resourceType": "bucket"
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

func getBulkPermitResponderWithResource() JSONMockerFunc {
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
							"scope": "ibm openid",
							"id": "IBMid-270003GUSX"
						}
					},
					"resource": {
						"attributes": {
							"accountId": "12345",
							"serviceName": "cloud-object-storage",
							"serviceInstance": "my-test-bucket-01",
							"resourceType": "bucket",
							"resource": "my-cos-resource"
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
