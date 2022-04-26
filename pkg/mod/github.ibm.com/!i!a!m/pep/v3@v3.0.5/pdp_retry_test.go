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

func TestRetryOnce_whenTooManyRequests(t *testing.T) {

	responders := []JSONMockerFunc{
		getTooManyRequestsResponder(),
		getBulkPermitResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
		HTTPClient:  mocker.httpClient,
		AuthzRetry:  true,
	}

	err := Configure(&config)
	require.NoError(t, err)

	requests := &Requests{getBulkRequest()}
	trace := "txid-retry-once-when-too-many-requests"
	response, err := PerformAuthorization(requests, trace)
	require.NoError(t, err)

	require.NotEmpty(t, response.Decisions)
	assert.True(t, response.Decisions[0].Permitted, "should get a permit")
	assert.False(t, response.Decisions[0].Cached, "should not hit the cache ")

	assert.Equal(t, 2, mocker.counter(), "the 2nd call to pdp should be successful")
	assert.Equal(t, 1, response.Decisions[0].RetryCount, "should retry once")
}

func TestRetryTwice_whenTooManyRequests(t *testing.T) {

	responders := []JSONMockerFunc{
		getTooManyRequestsResponder(),
		getTooManyRequestsResponder(),
		getBulkPermitResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
		HTTPClient:  mocker.httpClient,
		AuthzRetry:  true,
	}

	err := Configure(&config)
	require.NoError(t, err)

	requests := &Requests{getBulkRequest()}
	trace := "txid-retry-twice-when-too-many-requests"
	response, err := PerformAuthorization(requests, trace)
	require.NoError(t, err)

	require.NotEmpty(t, response.Decisions)
	assert.True(t, response.Decisions[0].Permitted, "should get a permit")
	assert.False(t, response.Decisions[0].Cached, "should not hit the cache ")

	assert.Equal(t, 3, mocker.counter(), "the 3rd call to pdp should be successful")
	assert.Equal(t, 2, response.Decisions[0].RetryCount, "should retry twice")

}
func TestRetryAfter_whenTooManyRequests(t *testing.T) {

	responders := []JSONMockerFunc{
		getTooManyRequestsWithRetryAfterResponder(),
		getBulkPermitResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
		HTTPClient:  mocker.httpClient,
		AuthzRetry:  true,
	}

	err := Configure(&config)
	require.NoError(t, err)

	requests := &Requests{getBulkRequest()}
	trace := "txid-retry-after-when-too-many-requests"

	start := time.Now()
	response, err := PerformAuthorization(requests, trace)
	elapsed := time.Since(start)

	require.NoError(t, err)

	require.NotEmpty(t, response.Decisions)
	assert.True(t, response.Decisions[0].Permitted, "should get a permit")
	assert.False(t, response.Decisions[0].Cached, "should not hit the cache ")

	afterRetryHeader := time.Duration(2 * time.Second)
	assert.Greater(t, elapsed, afterRetryHeader, "should respect the After-Retry header")

	assert.Equal(t, 2, mocker.counter(), "should call twice to the pdp")
	assert.Equal(t, 1, response.Decisions[0].RetryCount, "should retry once")

}

func TestRetry_whenInternalServerError(t *testing.T) {

	responders := []JSONMockerFunc{
		getInternalServerErrorResponder(),
		getBulkPermitResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
		HTTPClient:  mocker.httpClient,
		AuthzRetry:  true,
	}

	err := Configure(&config)
	require.NoError(t, err)

	requests := &Requests{getBulkRequest()}
	trace := "txid-retry-when-internal-server-error"
	response, err := PerformAuthorization(requests, trace)
	require.NoError(t, err)

	require.NotEmpty(t, response.Decisions)
	assert.True(t, response.Decisions[0].Permitted, "should get a permit")
	assert.False(t, response.Decisions[0].Cached, "should not hit the cache ")

	assert.Equal(t, 2, mocker.counter(), "the second call to pdp should be successful")
	assert.Equal(t, 1, response.Decisions[0].RetryCount, "should retry once")

}

// 502
func TestRetry_whenBadGateway(t *testing.T) {

	responders := []JSONMockerFunc{
		getBadGatewayResponder(),
		getBulkPermitResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
		HTTPClient:  mocker.httpClient,
		AuthzRetry:  true,
	}

	err := Configure(&config)
	require.NoError(t, err)

	requests := &Requests{getBulkRequest()}
	trace := "txid-retry-when-bad-gateway"
	response, err := PerformAuthorization(requests, trace)
	require.NoError(t, err)

	require.NotEmpty(t, response.Decisions)
	assert.True(t, response.Decisions[0].Permitted, "should get a permit")
	assert.False(t, response.Decisions[0].Cached, "should not hit the cache ")

	assert.Equal(t, 2, mocker.counter(), "the second call to pdp should be successful")
	assert.Equal(t, 1, response.Decisions[0].RetryCount, "should retry once")

}

// 503
func TestRetry_whenServiceUnavailable(t *testing.T) {

	responders := []JSONMockerFunc{
		getServiceUnavailableResponder(),
		getBulkPermitResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
		HTTPClient:  mocker.httpClient,
		AuthzRetry:  true,
	}

	err := Configure(&config)
	require.NoError(t, err)

	requests := &Requests{getBulkRequest()}
	trace := "txid-retry-when-service-unavailable"
	response, err := PerformAuthorization(requests, trace)
	require.NoError(t, err)

	require.NotEmpty(t, response.Decisions)
	assert.True(t, response.Decisions[0].Permitted, "should get a permit")
	assert.False(t, response.Decisions[0].Cached, "should not hit the cache ")

	assert.Equal(t, 2, mocker.counter(), "the second call to pdp should be successful")
	assert.Equal(t, 1, response.Decisions[0].RetryCount, "should retry once")

}

// 504
func TestRetry_whenGatewayTimeout(t *testing.T) {

	responders := []JSONMockerFunc{
		getGatewayTimeoutResponder(),
		getBulkPermitResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
		HTTPClient:  mocker.httpClient,
		AuthzRetry:  true,
	}

	err := Configure(&config)
	require.NoError(t, err)

	requests := &Requests{getBulkRequest()}
	trace := "txid-retry-when-gateway-timeout"
	response, err := PerformAuthorization(requests, trace)
	require.NoError(t, err)

	require.NotEmpty(t, response.Decisions)
	assert.True(t, response.Decisions[0].Permitted, "should get a permit")
	assert.False(t, response.Decisions[0].Cached, "should not hit the cache ")

	assert.Equal(t, 2, mocker.counter(), "the second call to pdp should be successful")
	assert.Equal(t, 1, response.Decisions[0].RetryCount, "should retry once")

}

func getTooManyRequestsResponder() JSONMockerFunc {
	return func(req *http.Request) mockedResponse {
		return mockedResponse{
			statusCode: http.StatusTooManyRequests,
		}
	}
}

func getTooManyRequestsWithRetryAfterResponder() JSONMockerFunc {
	afterRetryHeader := 2
	return func(req *http.Request) mockedResponse {
		return mockedResponse{
			header: map[string]string{
				"Retry-After": fmt.Sprintf("%d", afterRetryHeader),
			},
			statusCode: http.StatusTooManyRequests,
		}
	}
}

func getInternalServerErrorResponder() JSONMockerFunc {
	return func(req *http.Request) mockedResponse {
		return mockedResponse{
			statusCode: http.StatusInternalServerError,
		}
	}
}

func getBadGatewayResponder() JSONMockerFunc {
	return func(req *http.Request) mockedResponse {
		return mockedResponse{
			statusCode: http.StatusBadGateway,
		}
	}
}

func getServiceUnavailableResponder() JSONMockerFunc {
	return func(req *http.Request) mockedResponse {
		return mockedResponse{
			statusCode: http.StatusServiceUnavailable,
		}
	}
}

func getGatewayTimeoutResponder() JSONMockerFunc {
	return func(req *http.Request) mockedResponse {
		return mockedResponse{
			statusCode: http.StatusGatewayTimeout,
		}
	}
}
