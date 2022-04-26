package pep

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDenyResponses_requirement1(t *testing.T) {
	denyTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getBulkDenyResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment:           Staging,
		APIKey:                os.Getenv("API_KEY"),
		LogLevel:              LevelError,
		CacheDefaultDeniedTTL: denyTTL,
		HTTPClient:            mocker.httpClient,
	}

	runTest := func(t *testing.T, config Config, request Request) {
		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}

		t.Run("call #0, expect a non-cached deny", func(t *testing.T) {
			trace := "txid-call#0-of-expired-cache-test"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "1st call should get a deny")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "1st call should not be expired")
		})

		t.Run("call #1, same request", func(t *testing.T) {
			trace := "txid-call#1-same-request"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call should get a deny")
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
			assert.False(t, response.Decisions[0].Permitted, "same call with different action should get a deny")
			assert.False(t, response.Decisions[0].Cached, "same call with different action should get a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		// We wait as long as the deny TTL so that cached entries are expired.
		time.Sleep(denyTTL)

		t.Run("call #3, after all cache expired, expect a non-cached deny", func(t *testing.T) {
			trace := "txid-call#3-after-cache-expired"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "1st call should get a deny")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "expired cache not enabled")
		})

		t.Run("verify cache key count", func(t *testing.T) {
			assert.Equal(t, uint64(2), GetStatistics().Cache.EntriesCount, "Expect two entries stored in the deny cache")
		})
	}

	runTest(t, config, getRequestWithAccountAndService(t))
}

func TestDenyResponses_requirement2(t *testing.T) {
	responders := []JSONMockerFunc{
		getBulkDenyResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment:        Staging,
		APIKey:             os.Getenv("API_KEY"),
		LogLevel:           LevelError,
		HTTPClient:         mocker.httpClient,
		DisableDeniedCache: true,
	}

	runTest := func(t *testing.T, config Config, request Request) {
		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}

		t.Run("call #0, expect a non-cached deny", func(t *testing.T) {
			trace := "txid-call#0-of-expired-cache-test"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "1st call should get a deny")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "1st call should not be expired")
		})

		t.Run("call #1, same request", func(t *testing.T) {
			trace := "txid-call#1-same-request"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call should get a deny")
			assert.False(t, response.Decisions[0].Cached, "same call should not hit the cache")
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
			assert.False(t, response.Decisions[0].Permitted, "same call with different action should get a deny")
			assert.False(t, response.Decisions[0].Cached, "same call with different action should get a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("verify cache count", func(t *testing.T) {
			assert.Equal(t, uint64(0), GetStatistics().Cache.EntriesCount, "Expect zero entries stored in the deny cache")
		})
	}

	runTest(t, config, getRequestWithAccountAndService(t))
}
