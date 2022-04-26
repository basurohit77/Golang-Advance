package pep

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoAdvancedResourceObligation_requirement1(t *testing.T) {
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getBulkPermitResponderWithoutResourceObligations(),
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

		t.Run("call #3, same request with unsupported action", func(t *testing.T) {
			trace := "txid-call#3-same-request-with-unsupported-action"
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

		t.Run("call #4, same request with different serviceName", func(t *testing.T) {
			trace := "txid-call#4-same-request-with-different-serviceName"
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

		t.Run("call #5, same request with different subject id", func(t *testing.T) {
			trace := "txid-call#5-same-request-different-id"
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

func TestNoAdvancedSubjectObligation_requirement2(t *testing.T) {
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getBulkPermitResponderWithoutSubjectObligations(),
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

		/* t.Run("call #2, same request with different supported action", func(t *testing.T) {
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
		}) */

		t.Run("call #3, same request with unsupported action", func(t *testing.T) {
			trace := "txid-call#3-same-request-with-unsupported-action"
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

		t.Run("call #4, same request with different serviceName", func(t *testing.T) {
			trace := "txid-call#4-same-request-with-different-serviceName"
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

		t.Run("call #5, same request with different subject id", func(t *testing.T) {
			trace := "txid-call#5-same-request-different-id"
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

func TestNoAdvancedActionObligation_requirement3(t *testing.T) {
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getBulkPermitResponderWithoutActionObligations(),
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

		t.Run("call #2, same request with different action", func(t *testing.T) {
			trace := "txid-call#2-same-request-with-action"
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
			assert.False(t, response.Decisions[0].Permitted, "same call with different action should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different action should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #3, same request with different serviceName", func(t *testing.T) {
			trace := "txid-call#3-same-request-with-different-serviceName"
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

		t.Run("call #4, same request with different subject id", func(t *testing.T) {
			trace := "txid-call#4-same-request-different-id"
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

func TestUnsupportedResourceAttributes_requirement4(t *testing.T) {
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getBulkPermitResponderWithUnsupportedResourceObligations(),
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
			assert.False(t, response.Decisions[0].Permitted, "same call with different action should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different action should be a cache miss")
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
			assert.False(t, response.Decisions[0].Permitted, "same call with different serviceInstance should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different serviceInstance should be a cache miss")
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

func TestUnsupportedSubjectAttributes_requirement5(t *testing.T) {
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getBulkPermitResponderWithUnsupportedSubjectObligations(),
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
			assert.False(t, response.Decisions[0].Permitted, "same call with different action should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different action should be a cache miss")
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
			assert.False(t, response.Decisions[0].Permitted, "same call with different serviceInstance should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with different serviceInstance should be a cache miss")
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

		t.Run("call #7, same request with different scope", func(t *testing.T) {
			trace := "txid-call#7-same-request-different-scope"
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

		t.Run("call #8, same request with different id", func(t *testing.T) {
			trace := "txid-call#8-same-request-different-id"
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

func getBulkPermitResponderWithoutSubjectObligations() JSONMockerFunc {
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

func getBulkPermitResponderWithoutActionObligations() JSONMockerFunc {
	mockedJSON := `
	{
		"decisions": [
			{
				"decision": "Permit",
				"obligation": {
					"maxCacheAgeSeconds": 600,
					"subject": {
						"attributes": {
							"scope": "ibm openid",
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

func getBulkPermitResponderWithUnsupportedResourceObligations() JSONMockerFunc {
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
							"something": "fake"
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

func getBulkPermitResponderWithUnsupportedSubjectObligations() JSONMockerFunc {
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
							"somethingelse": "whatever"
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
