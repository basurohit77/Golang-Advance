package pep

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthzCacheExpiration_permitEntryShouldExpireAfterTTL(t *testing.T) {

	// Each call should take about 5 us
	// so a TTL of 40 millisecond should be more than enough to run 10 mocked authz requests
	// without getting expired entries
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getAuthzPermitResponder(),
	}

	mocker := NewAuthzMocker(t, responders)

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

		t.Run("call #1, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#1-before-expiration "
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		for n := 2; n <= 10; n++ {
			t.Run(fmt.Sprintf("subsequent call #%d, expect a cached permit", n), func(t *testing.T) {
				trace := fmt.Sprintf("txid-call#%d-before-expiration", n)
				response, err := PerformAuthorization(requests, trace)
				require.NoError(t, err)

				require.NotEmpty(t, response.Decisions)
				assert.True(t, response.Decisions[0].Permitted, fmt.Sprintf("call #%d should get a permit", n))
				assert.True(t, response.Decisions[0].Cached, fmt.Sprintf("call #%d should have a cache hit", n))
				assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
			})
		}

		// We wait as long as the cache TTL so that cached entries are expired.
		time.Sleep(cacheTTL)

		// Now that all cached entries are expired, we should see the same behavior as if we start from the beginning
		t.Run("call #1 after expiration, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#1-after-expiration"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		for n := 2; n <= 10; n++ {
			t.Run(fmt.Sprintf("subsequent call #%d after expration, expect a cached permit", n), func(t *testing.T) {
				trace := fmt.Sprintf("txid-call#%d-after-expiration", n)
				response, err := PerformAuthorization(requests, trace)
				require.NoError(t, err)

				require.NotEmpty(t, response.Decisions)
				assert.True(t, response.Decisions[0].Permitted, fmt.Sprintf("call #%d should get a permit", n))
				assert.True(t, response.Decisions[0].Cached, fmt.Sprintf("call #%d should have a cache hit", n))
				assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
			})
		}
	}

	runTest(t, config, getAuthzRequest())
}

func TestAuthzCacheExpiration_expiredCacheEntryShouldReturnIfEnabled(t *testing.T) {

	// Each call should take about 5 us
	// so a TTL of 40 millisecond should be more than enough to run 10 mocked authz requests
	// without getting expired entries
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getAuthzPermitResponder(),
		getServerErrorResponder(), // we simulate the server error from pdp on the second call
	}

	mocker := NewAuthzMocker(t, responders)

	config := Config{
		Environment:        Staging,
		APIKey:             os.Getenv("API_KEY"),
		LogLevel:           LevelError,
		CacheDefaultTTL:    cacheTTL,
		HTTPClient:         mocker.httpClient,
		EnableExpiredCache: true,
	}

	runTest := func(t *testing.T, config Config, request Request) {

		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}
		t.Run("call #1, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#1-before-expiration"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "1st call should not be expired")
		})

		for n := 2; n <= 10; n++ {
			t.Run(fmt.Sprintf("subsequent call #%d, expect a cached permit", n), func(t *testing.T) {
				trace := fmt.Sprintf("txid-call#%d-before-expiration", n)
				response, err := PerformAuthorization(requests, trace)
				require.NoError(t, err)

				require.NotEmpty(t, response.Decisions)
				assert.True(t, response.Decisions[0].Permitted, fmt.Sprintf("call #%d should get a permit", n))
				assert.True(t, response.Decisions[0].Cached, fmt.Sprintf("call #%d should have a cache hit", n))
				assert.False(t, response.Decisions[0].Expired, "entry should not expire yet")
			})
		}

		// We wait as long as the cache TTL so that cached entries are expired.
		time.Sleep(cacheTTL)

		// Now that all cached entries are expired and pdp is still not reachable we expect expired decision to be returned.

		for n := 1; n <= 10; n++ {
			t.Run(fmt.Sprintf("all call #%d after expiration, expect an expired cached permit", n), func(t *testing.T) {
				trace := fmt.Sprintf("txid-call#%d-after-expiration", n)
				response, err := PerformAuthorization(requests, trace)
				require.NoError(t, err)

				require.NotEmpty(t, response.Decisions)
				assert.True(t, response.Decisions[0].Permitted, fmt.Sprintf("call #%d should get a permit", n))
				assert.True(t, response.Decisions[0].Cached, fmt.Sprintf("call #%d should have a cache hit", n))
				assert.True(t, response.Decisions[0].Expired, "entry should show as expired since we enabled it")
			})
		}

	}

	runTest(t, config, getAuthzRequest())

	// Re-enabling pdp connectivity and default timeout
	responders = []JSONMockerFunc{
		getAuthzPermitResponder(),
	}

	mocker = NewAuthzMocker(t, responders)

	pepConfig.HTTPClient = mocker.httpClient
	pepConfig.CacheDefaultTTL = time.Duration(10) * time.Minute

	runTestAfterRecovery := func(t *testing.T, request Request) {

		requests := &Requests{request}
		t.Run("call #1 after recovery, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#1-after-recovery"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call after recovery should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call after recovery should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "1st call after recovery should not be expired")
		})

	}

	runTestAfterRecovery(t, getAuthzRequest())

}

func TestAuthzCacheExpiration_denyEntryShouldExpireAfterTTL(t *testing.T) {

	// Each call should take about 5 us
	// so a TTL of 40 millisecond should be more than enough to run 10 mocked authz requests
	// without getting expired entries
	denyTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getAuthzDenyResponder(),
	}

	mocker := NewAuthzMocker(t, responders)

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
		t.Run("call #1, expect a non-cached deny", func(t *testing.T) {
			trace := "txid-call#1-of-expired-cache-test"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "1st call should get a deny")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "expired cache not enabled")
		})

		for n := 1; n < 8; n++ {
			t.Run(fmt.Sprintf("call #%d, expect a cached deny", n), func(t *testing.T) {
				trace := fmt.Sprintf("txid-call#%d-of-cached-deny", n)

				response, err := PerformAuthorization(requests, trace)
				require.NoError(t, err)

				require.NotEmpty(t, response.Decisions)
				assert.False(t, response.Decisions[0].Permitted, fmt.Sprintf("call #%d should get a deny", n))
				assert.True(t, response.Decisions[0].Cached, fmt.Sprintf("call #%d should have no cache hit since the entry expired", n))
				assert.False(t, response.Decisions[0].Expired, "expired cache not enabled")
			})
		}

		// We wait as long as the deny TTL so that cached entries are expired.
		time.Sleep(denyTTL)

		t.Run("calls after all cache expired, expect a non-cached deny", func(t *testing.T) {
			trace := "txid-call-after-cache-expired"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "1st call should get a deny")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "expired cache not enabled")
		})

		for n := 3; n < 8; n++ {
			t.Run(fmt.Sprintf("call #%d, expect a cached deny", n), func(t *testing.T) {
				trace := fmt.Sprintf("txid-call#%d-of-expired-deny", n)

				response, err := PerformAuthorization(requests, trace)
				require.NoError(t, err)

				require.NotEmpty(t, response.Decisions)
				assert.False(t, response.Decisions[0].Permitted, fmt.Sprintf("call #%d should get a deny", n))
				assert.True(t, response.Decisions[0].Cached, fmt.Sprintf("call #%d should have no cache hit since the entry expired", n))
				assert.False(t, response.Decisions[0].Expired, "expired cache not enabled")
			})
		}
	}

	runTest(t, config, getAuthzRequest())
}

func TestAuthzCacheExpiration_denyEntryShouldReturnAfterExpiration(t *testing.T) {

	// TTL of 40 millisecond should be enough to run 10 mocked authz requests
	denyTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getAuthzDenyResponder(),
		getServerErrorResponder(),
	}

	mocker := NewAuthzMocker(t, responders)

	config := Config{
		Environment:           Staging,
		APIKey:                os.Getenv("API_KEY"),
		LogLevel:              LevelError,
		CacheDefaultDeniedTTL: denyTTL,
		HTTPClient:            mocker.httpClient,
		EnableExpiredCache:    true,
	}

	runTest := func(t *testing.T, config Config, request Request) {

		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}
		t.Run("call #1, expect a non-cached deny", func(t *testing.T) {
			trace := "txid-call#1-of-expired-cache-test"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "1st call should get a deny")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "1st call should not be expired")
		})

		t.Run("call #2, expect a cached deny", func(t *testing.T) {
			trace := "txid-call#2-of-expired-cache-test"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "call #2 should get a deny")
			assert.True(t, response.Decisions[0].Cached, "call #2 should hit the cache")
			assert.False(t, response.Decisions[0].Expired, "call #2 should not be expired")
		})

		// We wait as long as the deny TTL so that cached entries are expired.
		time.Sleep(denyTTL)

		for n := 3; n < 8; n++ {
			t.Run(fmt.Sprintf("call #%d, expect a cached deny even if expired", n), func(t *testing.T) {
				trace := fmt.Sprintf("txid-call#%d-of-expired-cache-test", n)

				// We wait as long as the cache TTL so that cached entries are expired.
				response, err := PerformAuthorization(requests, trace)
				require.NoError(t, err)

				require.NotEmpty(t, response.Decisions)
				assert.False(t, response.Decisions[0].Permitted, fmt.Sprintf("call #%d should get a deny", n))
				assert.True(t, response.Decisions[0].Cached, fmt.Sprintf("call #%d should have a cache hit even expired", n))
				assert.True(t, response.Decisions[0].Expired, "call #%d should be expired")
			})
		}
	}

	runTest(t, config, getAuthzRequest())

	// Simulating pdp connection recovery
	responders = []JSONMockerFunc{
		getAuthzDenyResponder(),
	}

	mocker = NewAuthzMocker(t, responders)

	pepConfig.HTTPClient = mocker.httpClient
	pepConfig.CacheDefaultDeniedTTL = time.Duration(2) * time.Minute

	runTestAfterRecovery := func(t *testing.T, request Request) {

		requests := &Requests{request}
		t.Run("call #1 after recovery, expect a non-cached deny", func(t *testing.T) {
			trace := "txid-call#1-after-recovery"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "1st call after recovery should get a deny")
			assert.False(t, response.Decisions[0].Cached, "1st call after recovery should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "1st call after recovery should not be expired")
		})

		t.Run("call #2 after recovery, expect a cached deny", func(t *testing.T) {
			trace := "txid-call#2-after-recovery"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "call #2 after recovery should get a deny")
			assert.True(t, response.Decisions[0].Cached, "call #2 after recovery should hit the cache")
			assert.False(t, response.Decisions[0].Expired, "call #2 after recovery should not be expired")
		})

	}

	runTestAfterRecovery(t, getAuthzRequest())

}

func TestAuthzCacheExpiration_pdpDictateTheCacheTTL(t *testing.T) {

	// maxCacheAgeSeconds from the pdp
	maxCacheAgeSeconds := 1
	responders := []JSONMockerFunc{
		getAuthzPermitWithMaxCacheAgeResponder(maxCacheAgeSeconds),
	}

	mocker := NewAuthzMocker(t, responders)

	// When the CacheDefaultTTL is not specified, the maxCacheAgeSeconds from the pdp should be used
	config := Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
		HTTPClient:  mocker.httpClient,
	}

	runTest := func(t *testing.T, config Config, request Request) {

		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}
		t.Run("call #1, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#1-before-expiration"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "1st call should not be expired")
		})

		for n := 2; n <= 10; n++ {
			t.Run(fmt.Sprintf("subsequent call #%d, expect a cached permit", n), func(t *testing.T) {
				trace := fmt.Sprintf("txid-call#%d-before-expiration", n)
				response, err := PerformAuthorization(requests, trace)
				require.NoError(t, err)

				require.NotEmpty(t, response.Decisions)
				assert.True(t, response.Decisions[0].Permitted, fmt.Sprintf("call #%d should get a permit", n))
				assert.True(t, response.Decisions[0].Cached, fmt.Sprintf("call #%d should have a cache hit", n))
				assert.False(t, response.Decisions[0].Expired, "entry should not expire yet")
			})
		}

		// We wait as long as the maxCacheAgeSeconds from the pdp so that cached entries are expired.
		time.Sleep(time.Duration(maxCacheAgeSeconds) * time.Second)

		t.Run("call after expiration, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call-after-expiration"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "call after cache expired should get a permit")
			assert.False(t, response.Decisions[0].Cached, "call after cache expired should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "call after cache expired should not be expired")
		})

	}

	runTest(t, config, getAuthzRequest())

}

func TestAuthzCacheExpiration_userDictateTheCacheTTL(t *testing.T) {

	// maxCacheAgeSeconds from the pdp
	maxCacheAgeSeconds := 2
	responders := []JSONMockerFunc{
		getAuthzPermitWithMaxCacheAgeResponder(maxCacheAgeSeconds),
	}

	mocker := NewAuthzMocker(t, responders)

	// user specified cache TTL
	userCacheTTL := time.Duration(1 * time.Second)
	config := Config{
		Environment:     Staging,
		APIKey:          os.Getenv("API_KEY"),
		LogLevel:        LevelError,
		HTTPClient:      mocker.httpClient,
		CacheDefaultTTL: userCacheTTL,
	}

	runTest := func(t *testing.T, config Config, request Request) {

		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}
		t.Run("call #1, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#1-before-expiration"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "1st call should not be expired")
		})

		for n := 2; n <= 10; n++ {
			t.Run(fmt.Sprintf("subsequent call #%d, expect a cached permit", n), func(t *testing.T) {
				trace := fmt.Sprintf("txid-call#%d-before-expiration", n)
				response, err := PerformAuthorization(requests, trace)
				require.NoError(t, err)

				require.NotEmpty(t, response.Decisions)
				assert.True(t, response.Decisions[0].Permitted, fmt.Sprintf("call #%d should get a permit", n))
				assert.True(t, response.Decisions[0].Cached, fmt.Sprintf("call #%d should have a cache hit", n))
				assert.False(t, response.Decisions[0].Expired, "entry should not expire yet")
			})
		}

		// We wait as long as the user-provided cacheTTL, from the pdp so that cached entries are expired.
		time.Sleep(userCacheTTL)

		t.Run("call after cache expiration, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#1-after-expiration"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "call after cache expiration should get a permit")
			assert.False(t, response.Decisions[0].Cached, "call after cache expiration should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "call after cache expiration should not be expired")
		})

	}

	runTest(t, config, getAuthzRequest())

}

func TestAuthzCacheExpiration_pdpDontHaveCacheTTL(t *testing.T) {

	// maxCacheAgeSeconds from the pdp
	responders := []JSONMockerFunc{
		getAuthzPermitWithoutMaxCacheAgeResponder(),
	}

	mocker := NewAuthzMocker(t, responders)

	// When the CacheDefaultTTL is not specified, the maxCacheAgeSeconds from the pdp should be used
	config := Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
		HTTPClient:  mocker.httpClient,
	}

	runTest := func(t *testing.T, config Config, request Request) {

		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}
		t.Run("call #1, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#1-before-expiration"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "1st call should not be expired")
		})

		for n := 2; n <= 10; n++ {
			t.Run(fmt.Sprintf("subsequent call #%d, expect a cached permit", n), func(t *testing.T) {
				trace := fmt.Sprintf("txid-call#%d-before-expiration", n)
				response, err := PerformAuthorization(requests, trace)
				require.NoError(t, err)

				require.NotEmpty(t, response.Decisions)
				assert.True(t, response.Decisions[0].Permitted, fmt.Sprintf("call #%d should get a permit", n))
				assert.True(t, response.Decisions[0].Cached, fmt.Sprintf("call #%d should have a cache hit", n))
				assert.False(t, response.Decisions[0].Expired, "entry should not expire yet")
			})
		}

		// Since there is no maxCacheAgeSecond from pdp, the cache entries should still be valid after 1 second
		time.Sleep(time.Duration(1) * time.Second)

		t.Run("call a second, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call-a second-later"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "call a second later should get a permit")
			assert.True(t, response.Decisions[0].Cached, "call a second later should hit the cache")
			assert.False(t, response.Decisions[0].Expired, "call a second later should not be expired")
		})

	}

	runTest(t, config, getAuthzRequest())

}

func TestAuthzCacheExpiration_userDisableTheCache(t *testing.T) {

	// maxCacheAgeSeconds from the pdp
	maxCacheAgeSeconds := 600
	responders := []JSONMockerFunc{
		getAuthzPermitWithMaxCacheAgeResponder(maxCacheAgeSeconds),
	}

	mocker := NewAuthzMocker(t, responders)

	// Settings a very small cacheTTL should have the same effect as disabling the cache
	userCacheTTL := time.Duration(1 * time.Nanosecond)
	config := Config{
		Environment:     Staging,
		APIKey:          os.Getenv("API_KEY"),
		LogLevel:        LevelError,
		HTTPClient:      mocker.httpClient,
		CacheDefaultTTL: userCacheTTL,
	}

	runTest := func(t *testing.T, config Config, request Request) {

		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}
		t.Run("call #1, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#1-before-expiration"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "1st call should not be expired")
		})

		for n := 2; n <= 10; n++ {
			t.Run(fmt.Sprintf("subsequent call #%d, expect a cached permit", n), func(t *testing.T) {
				trace := fmt.Sprintf("txid-call#%d-before-expiration", n)
				response, err := PerformAuthorization(requests, trace)
				require.NoError(t, err)

				require.NotEmpty(t, response.Decisions)
				assert.True(t, response.Decisions[0].Permitted, fmt.Sprintf("call #%d should get a permit", n))
				assert.False(t, response.Decisions[0].Cached, fmt.Sprintf("call #%d should have a cache hit", n))
				assert.False(t, response.Decisions[0].Expired, "entry should not expire yet")
			})
		}

	}

	runTest(t, config, getAuthzRequest())

}

func TestAuthzCacheExpiration_userDictateTheDeniedCacheTTL(t *testing.T) {

	// user-defined deniedTTL to 1 second
	userDeniedTTL := time.Duration(1) * time.Second
	responders := []JSONMockerFunc{
		getAuthzDenyResponder(),
	}

	mocker := NewAuthzMocker(t, responders)

	// user specified cache TTL
	userCacheTTL := time.Duration(1 * time.Second)
	config := Config{
		Environment:           Staging,
		APIKey:                os.Getenv("API_KEY"),
		LogLevel:              LevelError,
		HTTPClient:            mocker.httpClient,
		CacheDefaultDeniedTTL: userDeniedTTL,
	}

	runTest := func(t *testing.T, config Config, request Request) {

		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}
		t.Run("call #1, expect a non-cached deny", func(t *testing.T) {
			trace := "txid-call#1-of-deny-cache-ttl"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "1st call should get a deny")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "1st call should not be expired")
		})

		t.Run("same call expect a cached deny", func(t *testing.T) {
			trace := "txid-call#2-of-deny-cache-ttl"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call should get a deny")
			assert.True(t, response.Decisions[0].Cached, "same call should hit the cache")
			assert.False(t, response.Decisions[0].Expired, "same call should not be expired")
		})

		// We wait as long as the user-provided cacheDeniedTTL
		time.Sleep(userCacheTTL)

		t.Run("call after cache expiration, expect a non-cached deny", func(t *testing.T) {
			trace := "txid-call#1-after-expiration"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "call after cache expiration should get a deny")
			assert.False(t, response.Decisions[0].Cached, "call after cache expiration should not hit the cache")
			assert.False(t, response.Decisions[0].Expired, "call after cache expiration should not be expired")
		})

	}

	runTest(t, config, getAuthzRequest())

}

func getAuthzPermitResponder() JSONMockerFunc {
	mockedJSON := `
	{
		"responses": [
		{
			"status": "200",
			"authorizationDecision": {
				"permitted": true,
				"obligation": {
					"actions": [
					"iam.policy.update",
					"iam.role.read",
					"iam.policy.read",
					"iam.policy.create",
					"iam.service.create",
					"iam.operator.update",
					"iam.delegationPolicy.create",
					"iam.policy.delete",
					"iam.delegationPolicy.update",
					"iam.operator.read",
					"iam.service.delete",
					"iam.service.read",
					"iam.role.assign"
					],
					"environment": {},
					"resources": [
					"crn:v1:staging:public:toolchain:us-south::::"
					],
					"maxCacheAgeSeconds": 600,
					"subject": {
						"attributes": {
							"scope": "ibm openid otc",
							"id": "IBMid-550005146S"
						}
					},
					"resource": {
						"attributes": {}
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
	}`

	responder := func(req *http.Request) mockedResponse {
		return mockedResponse{
			json:       mockedJSON,
			statusCode: 200,
		}
	}
	return responder
}

func getAuthzDenyResponder() JSONMockerFunc {
	mockedJSON := `
	{
		"responses": [
			{
				"status": "200",
				"authorizationDecision": {
					"permitted": false
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
	}`
	responder := func(req *http.Request) mockedResponse {
		return mockedResponse{
			json:       mockedJSON,
			statusCode: 200,
		}
	}
	return responder
}

func getServerErrorResponder() JSONMockerFunc {
	responder := func(req *http.Request) mockedResponse {
		return mockedResponse{
			json:       "{}",
			statusCode: 500,
		}
	}
	return responder
}

func getAuthzRequest() Request {

	subject := Attributes{
		"scope": "ibm openid otc",
		"id":    "IBMid-550005146S",
	}
	action := "iam.policy.read"
	resource := Attributes{
		"crn": "crn:v1:staging:public:toolchain:us-south::::",
	}

	request := Request{
		"action":   action,
		"resource": resource,
		"subject":  subject,
	}

	return request
}

func getAuthzPermitWithMaxCacheAgeResponder(cacheTTL int) JSONMockerFunc {
	mockedJSON := fmt.Sprintf(`
	{
		"responses": [
		{
			"status": "200",
			"authorizationDecision": {
				"permitted": true,
				"obligation": {
					"actions": [
					"iam.policy.update",
					"iam.role.read",
					"iam.policy.read",
					"iam.policy.create",
					"iam.service.create",
					"iam.operator.update",
					"iam.delegationPolicy.create",
					"iam.policy.delete",
					"iam.delegationPolicy.update",
					"iam.operator.read",
					"iam.service.delete",
					"iam.service.read",
					"iam.role.assign"
					],
					"environment": {},
					"resources": [
					"crn:v1:staging:public:toolchain:us-south::::"
					],
					"maxCacheAgeSeconds": %d,
					"subject": {
						"attributes": {
							"scope": "ibm openid otc",
							"id": "IBMid-550005146S"
						}
					},
					"resource": {
						"attributes": {}
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
	}`, cacheTTL)

	responder := func(req *http.Request) mockedResponse {
		return mockedResponse{
			json:       mockedJSON,
			statusCode: 200,
		}
	}
	return responder
}

func getAuthzPermitWithoutMaxCacheAgeResponder() JSONMockerFunc {
	mockedJSON := `
	{
		"responses": [
		{
			"status": "200",
			"authorizationDecision": {
				"permitted": true,
				"obligation": {
					"actions": [
					"iam.policy.update",
					"iam.role.read",
					"iam.policy.read",
					"iam.policy.create",
					"iam.service.create",
					"iam.operator.update",
					"iam.delegationPolicy.create",
					"iam.policy.delete",
					"iam.delegationPolicy.update",
					"iam.operator.read",
					"iam.service.delete",
					"iam.service.read",
					"iam.role.assign"
					],
					"environment": {},
					"resources": [
					"crn:v1:staging:public:toolchain:us-south::::"
					],
					"subject": {
						"attributes": {
							"scope": "ibm openid otc",
							"id": "IBMid-550005146S"
						}
					},
					"resource": {
						"attributes": {}
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
	}`

	responder := func(req *http.Request) mockedResponse {
		return mockedResponse{
			json:       mockedJSON,
			statusCode: 200,
		}
	}
	return responder
}
