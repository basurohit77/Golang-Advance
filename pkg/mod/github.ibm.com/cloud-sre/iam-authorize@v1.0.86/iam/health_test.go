package iam

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mockIAMServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		return
	}))
)

// NewPepWrapperForUnitTest creates a Pep wrapper test stub. The arguments are functions that provide the
// implementations of the wrapper functions. This allows different test cases to be tried and verified.
func NewPepWrapperForUnitTest(
	getServiceTokenValidator getServiceTokenValidator,
	verifyUserTokenValidator verifyUserTokenValidator) pepWrapperForUnitTest {
	pepWrapperForUnitTest := pepWrapperForUnitTest{}
	pepWrapperForUnitTest.getServiceTokenValidator = getServiceTokenValidator
	pepWrapperForUnitTest.verifyUserTokenValidator = verifyUserTokenValidator
	return pepWrapperForUnitTest
}

type getServiceTokenValidator func(HealthCheckConfig) (string, error)
type verifyUserTokenValidator func(tokenString string) (bool, error)
type pepWrapperForUnitTest struct {
	getServiceTokenValidator getServiceTokenValidator
	verifyUserTokenValidator verifyUserTokenValidator
}

// func (pepWrapper pepWrapperForUnitTest) GetToken(apiKey string) (string, error) {
func (pepWrapper pepWrapperForUnitTest) GetToken(healthCheck HealthCheckConfig) (string, error) {
	return pepWrapper.getServiceTokenValidator(healthCheck)
}
func (pepWrapper pepWrapperForUnitTest) GetClaims(tokenString string) (bool, error) {
	return pepWrapper.verifyUserTokenValidator(tokenString)
}

// NewNRWrapperForUnitTest creates a NewRelic wrapper test stub. The arguments are functions that provide the
// implementations of the wrapper functions. This allows different test cases to be tried and verified.
func NewNRWrapperForUnitTest(sendTransactionValidator sendTransactionValidator) nrWrapperForTest {
	nrWrapperForTest := nrWrapperForTest{}
	nrWrapperForTest.sendTransactionValidator = sendTransactionValidator
	return nrWrapperForTest
}

type sendTransactionValidator func(transactionName string, success bool, err string)
type nrWrapperForTest struct {
	sendTransactionValidator sendTransactionValidator
}

func (nrWrapper nrWrapperForTest) SendTransaction(transactionName string, success bool, err string) {
	nrWrapper.sendTransactionValidator(transactionName, success, err)
}

func TestHealthCheckSuccess(t *testing.T) {
	// Tests case where everything is successful the first time around and IAM is considered up.

	// Expected values:
	apiKeyExpected := "1234567890"
	tokenStringExpected := "abcd.abcd.abcd"
	transactionNameExpected := "IAMHealthCheck"
	successExpected := true
	errStringExpected := ""
	isHealthyExpected := true
	numberOfTimesGetServiceTokenCalledExpected := 1
	numberOfTimesVerifyUserTokenCalledExpected := 1

	// Counters tracking how many times the Pep functions are called:
	numberOfTimesGetServiceTokenCalled := 0
	numberOfTimesVerifyUserTokenCalled := 0

	// Create health check configuation:
	healthCheckConfig := createConfig(apiKeyExpected)

	getServiceTokenValidator := func(healthCheck HealthCheckConfig) (string, error) {
		numberOfTimesGetServiceTokenCalled++
		assert.Equal(t, apiKeyExpected, healthCheck.IAMApiKey)
		return tokenStringExpected, nil // Success
	}
	verifyUserTokenValidator := func(tokenString string) (bool, error) {
		numberOfTimesVerifyUserTokenCalled++
		assert.Equal(t, tokenStringExpected, tokenString)
		return true, nil // Success
	}
	pepWrapper := NewPepWrapperForUnitTest(getServiceTokenValidator, verifyUserTokenValidator)

	// Create test NewRelic wrapper with implementations of the wrapper functions:
	sendTransactionValidator := func(transactionName string, success bool, err string) {
		assert.Equal(t, transactionNameExpected, transactionName)
		assert.Equal(t, successExpected, success)
		assert.Equal(t, errStringExpected, err)
	}
	nrWrapperForTest := NewNRWrapperForUnitTest(sendTransactionValidator)

	// Create health check and verify result:
	healthCheck := NewHealthCheck(healthCheckConfig, pepWrapper, nrWrapperForTest)
	assert.Equal(t, isHealthyExpected, healthCheck.IsHealthy())
	assert.Equal(t, numberOfTimesGetServiceTokenCalledExpected, numberOfTimesGetServiceTokenCalled)
	assert.Equal(t, numberOfTimesVerifyUserTokenCalledExpected, numberOfTimesVerifyUserTokenCalled)
}

func TestHealthCheckGetServiceTokenError(t *testing.T) {
	// Tests case where the Pep GetServiceToken function fails twice resulting in IAM being considered down.

	// Expected values:
	apiKeyExpected := "1234567890"
	tokenStringExpected := "abcd.abcd.abcd"
	transactionNameExpected := "IAMHealthCheck"
	successExpected := false
	errStringExpected := "Something bad happened"
	isHealthyExpected := false
	numberOfTimesGetServiceTokenCalledExpected := 2
	numberOfTimesVerifyUserTokenCalledExpected := 0

	// Counters tracking how many times the Pep functions are called:
	numberOfTimesGetServiceTokenCalled := 0
	numberOfTimesVerifyUserTokenCalled := 0

	// Create health check configuation:
	healthCheckConfig := createConfig(apiKeyExpected)

	getServiceTokenValidator := func(healthCheck HealthCheckConfig) (string, error) {
		numberOfTimesGetServiceTokenCalled++
		assert.Equal(t, apiKeyExpected, healthCheck.IAMApiKey)
		return "", errors.New(errStringExpected) // Error
	}
	verifyUserTokenValidator := func(tokenString string) (bool, error) {
		numberOfTimesVerifyUserTokenCalled++
		assert.Equal(t, tokenStringExpected, tokenString)
		return true, nil // Success
	}
	pepWrapper := NewPepWrapperForUnitTest(getServiceTokenValidator, verifyUserTokenValidator)

	// Create test NewRelic wrapper with implementations of the wrapper functions:
	sendTransactionValidator := func(transactionName string, success bool, err string) {
		assert.Equal(t, transactionNameExpected, transactionName)
		assert.Equal(t, successExpected, success)
		assert.Equal(t, errStringExpected, err)
	}
	nrWrapperForTest := NewNRWrapperForUnitTest(sendTransactionValidator)

	// Create health check and verify result:
	healthCheck := NewHealthCheck(healthCheckConfig, pepWrapper, nrWrapperForTest)
	assert.Equal(t, isHealthyExpected, healthCheck.IsHealthy())
	assert.Equal(t, numberOfTimesGetServiceTokenCalledExpected, numberOfTimesGetServiceTokenCalled)
	assert.Equal(t, numberOfTimesVerifyUserTokenCalledExpected, numberOfTimesVerifyUserTokenCalled)
}

func TestHealthVerifyUserTokenError(t *testing.T) {
	// Tests case where the Pep VerifyUserToken function fails twice resulting in IAM being considered down.

	// Expected values:
	apiKeyExpected := "1234567890"
	tokenStringExpected := "abcd.abcd.abcd"
	transactionNameExpected := "IAMHealthCheck"
	successExpected := false
	errStringExpected := "Something bad happened"
	isHealthyExpected := false
	numberOfTimesGetServiceTokenCalledExpected := 2
	numberOfTimesVerifyUserTokenCalledExpected := 2

	// Counters tracking how many times the Pep functions are called:
	numberOfTimesGetServiceTokenCalled := 0
	numberOfTimesVerifyUserTokenCalled := 0

	// Create health check configuation:
	healthCheckConfig := createConfig(apiKeyExpected)

	getServiceTokenValidator := func(healthCheck HealthCheckConfig) (string, error) {
		numberOfTimesGetServiceTokenCalled++
		assert.Equal(t, apiKeyExpected, healthCheck.IAMApiKey)
		return tokenStringExpected, nil // Success
	}
	verifyUserTokenValidator := func(tokenString string) (bool, error) {
		numberOfTimesVerifyUserTokenCalled++
		assert.Equal(t, tokenStringExpected, tokenString)
		return false, errors.New(errStringExpected) // Error
	}
	pepWrapper := NewPepWrapperForUnitTest(getServiceTokenValidator, verifyUserTokenValidator)

	// Create test NewRelic wrapper with implementations of the wrapper functions:
	sendTransactionValidator := func(transactionName string, success bool, err string) {
		assert.Equal(t, transactionNameExpected, transactionName)
		assert.Equal(t, successExpected, success)
		assert.Equal(t, errStringExpected, err)
	}
	nrWrapperForTest := NewNRWrapperForUnitTest(sendTransactionValidator)

	// Create health check and verify result:
	healthCheck := NewHealthCheck(healthCheckConfig, pepWrapper, nrWrapperForTest)
	assert.Equal(t, isHealthyExpected, healthCheck.IsHealthy())
	assert.Equal(t, numberOfTimesGetServiceTokenCalledExpected, numberOfTimesGetServiceTokenCalled)
	assert.Equal(t, numberOfTimesVerifyUserTokenCalledExpected, numberOfTimesVerifyUserTokenCalled)
}

func TestHealthCheckGetServiceTokenErrorFirstTime(t *testing.T) {
	// Tests case where the Pep GetServiceToken function fails the first time but is successful the second time
	// resulting in IAM being considered up.

	// Expected values:
	apiKeyExpected := "1234567890"
	tokenStringExpected := "abcd.abcd.abcd"
	transactionNameExpected := "IAMHealthCheck"
	successExpected := true
	errStringExpected := ""
	isHealthyExpected := true
	numberOfTimesGetServiceTokenCalledExpected := 2
	numberOfTimesVerifyUserTokenCalledExpected := 1

	// Counters tracking how many times the Pep functions are called:
	numberOfTimesGetServiceTokenCalled := 0
	numberOfTimesVerifyUserTokenCalled := 0

	// Create health check configuation:
	healthCheckConfig := createConfig(apiKeyExpected)

	getServiceTokenValidator := func(healthCheck HealthCheckConfig) (string, error) {
		numberOfTimesGetServiceTokenCalled++
		assert.Equal(t, apiKeyExpected, healthCheck.IAMApiKey)
		if numberOfTimesGetServiceTokenCalled == 1 {
			return "", errors.New(errStringExpected) // Error first time called
		}
		return tokenStringExpected, nil // Success second time called
	}
	verifyUserTokenValidator := func(tokenString string) (bool, error) {
		numberOfTimesVerifyUserTokenCalled++
		assert.Equal(t, tokenStringExpected, tokenString)
		return true, nil
	}
	pepWrapper := NewPepWrapperForUnitTest(getServiceTokenValidator, verifyUserTokenValidator)

	// Create test NewRelic wrapper with implementations of the wrapper functions:
	sendTransactionValidator := func(transactionName string, success bool, err string) {
		assert.Equal(t, transactionNameExpected, transactionName)
		assert.Equal(t, successExpected, success)
		assert.Equal(t, errStringExpected, err)
	}
	nrWrapperForTest := NewNRWrapperForUnitTest(sendTransactionValidator)

	// Create health check and verify result:
	healthCheck := NewHealthCheck(healthCheckConfig, pepWrapper, nrWrapperForTest)
	assert.Equal(t, isHealthyExpected, healthCheck.IsHealthy())
	assert.Equal(t, numberOfTimesGetServiceTokenCalledExpected, numberOfTimesGetServiceTokenCalled)
	assert.Equal(t, numberOfTimesVerifyUserTokenCalledExpected, numberOfTimesVerifyUserTokenCalled)
}

func TestHealthUserVerifyTokenErrorFirstTime(t *testing.T) {
	// Tests case where the Pep VerifyUserToken function fails the first time but is successful the second time
	// resulting in IAM being considered up.

	// Expected values:
	apiKeyExpected := "1234567890"
	tokenStringExpected := "abcd.abcd.abcd"
	transactionNameExpected := "IAMHealthCheck"
	successExpected := true
	errStringExpected := ""
	isHealthyExpected := true
	numberOfTimesGetServiceTokenCalledExpected := 2
	numberOfTimesVerifyUserTokenCalledExpected := 2

	// Counters tracking how many times the Pep functions are called:
	numberOfTimesGetServiceTokenCalled := 0
	numberOfTimesVerifyUserTokenCalled := 0

	// Create health check configuation:
	healthCheckConfig := createConfig(apiKeyExpected)

	getServiceTokenValidator := func(healthCheck HealthCheckConfig) (string, error) {
		numberOfTimesGetServiceTokenCalled++
		assert.Equal(t, apiKeyExpected, healthCheck.IAMApiKey)
		return tokenStringExpected, nil
	}
	verifyUserTokenValidator := func(tokenString string) (bool, error) {
		numberOfTimesVerifyUserTokenCalled++
		assert.Equal(t, tokenStringExpected, tokenString)
		if numberOfTimesVerifyUserTokenCalled == 1 {
			return false, errors.New(errStringExpected) // Error first time called
		}
		return true, nil // Success second time called
	}
	pepWrapper := NewPepWrapperForUnitTest(getServiceTokenValidator, verifyUserTokenValidator)

	// Create test NewRelic wrapper with implementations of the wrapper functions:
	sendTransactionValidator := func(transactionName string, success bool, err string) {
		assert.Equal(t, transactionNameExpected, transactionName)
		assert.Equal(t, successExpected, success)
		assert.Equal(t, errStringExpected, err)
	}
	nrWrapperForTest := NewNRWrapperForUnitTest(sendTransactionValidator)

	// Create health check and verify result:
	healthCheck := NewHealthCheck(healthCheckConfig, pepWrapper, nrWrapperForTest)
	assert.Equal(t, isHealthyExpected, healthCheck.IsHealthy())
	assert.Equal(t, numberOfTimesGetServiceTokenCalledExpected, numberOfTimesGetServiceTokenCalled)
	assert.Equal(t, numberOfTimesVerifyUserTokenCalledExpected, numberOfTimesVerifyUserTokenCalled)
}

func TestHealthUpdateIAMURL(t *testing.T) {
	// Tests that the IAM URL can be changed after the health check is created

	// Expected values:
	apiKeyExpected := "1234567890"
	tokenStringExpected := "abcd.abcd.abcd"
	transactionNameExpected := "IAMHealthCheck"
	successExpected := true
	errStringExpected := ""
	isHealthyExpected := true
	numberOfTimesGetServiceTokenCalledExpected := 1
	numberOfTimesVerifyUserTokenCalledExpected := 1
	updatedIAMURL := mockIAMServer.URL
	hasIAMURLBeenUpdated := false

	// Counters tracking how many times the Pep functions are called:
	numberOfTimesGetServiceTokenCalled := 0
	numberOfTimesVerifyUserTokenCalled := 0

	// Create health check configuation:
	healthCheckConfig := createConfig(apiKeyExpected)

	getServiceTokenValidator := func(healthCheck HealthCheckConfig) (string, error) {
		if hasIAMURLBeenUpdated {
			assert.True(t, strings.HasPrefix(healthCheck.IAMUrl, updatedIAMURL))
		}
		numberOfTimesGetServiceTokenCalled++
		assert.Equal(t, apiKeyExpected, healthCheck.IAMApiKey)
		return tokenStringExpected, nil // Success
	}
	verifyUserTokenValidator := func(tokenString string) (bool, error) {
		numberOfTimesVerifyUserTokenCalled++
		assert.Equal(t, tokenStringExpected, tokenString)
		return true, nil // Success
	}
	pepWrapper := NewPepWrapperForUnitTest(getServiceTokenValidator, verifyUserTokenValidator)

	// Create test NewRelic wrapper with implementations of the wrapper functions:
	sendTransactionValidator := func(transactionName string, success bool, err string) {
		assert.Equal(t, transactionNameExpected, transactionName)
		assert.Equal(t, successExpected, success)
		assert.Equal(t, errStringExpected, err)
	}
	nrWrapperForTest := NewNRWrapperForUnitTest(sendTransactionValidator)

	// Create health check and verify result:
	healthCheck := NewHealthCheck(healthCheckConfig, pepWrapper, nrWrapperForTest)
	assert.Equal(t, isHealthyExpected, healthCheck.IsHealthy())
	assert.Equal(t, numberOfTimesGetServiceTokenCalledExpected, numberOfTimesGetServiceTokenCalled)
	assert.Equal(t, numberOfTimesVerifyUserTokenCalledExpected, numberOfTimesVerifyUserTokenCalled)

	// Update the IAM URL and verify that the updated URL has been picked up by calling the
	// CheckHealth function (note that the verification of the IAM URL happens in the setConfigValidator
	// function above):
	healthCheck.UpdateIAMURL(updatedIAMURL)
	hasIAMURLBeenUpdated = true
	healthCheck.CheckHealth()
}

func TestCheckHealthAndTrackRecoveryWhenIsHealthy(t *testing.T) {
	// Tests that IAM health can be tracked when IAM is up

	// Expected values:
	apiKeyExpected := "1234567890"
	tokenStringExpected := "abcd.abcd.abcd"
	transactionNameExpected := "IAMHealthCheck"
	successExpected := true
	errStringExpected := ""
	isHealthyExpected := true
	isMonitorExpected := false
	numberOfTimesGetServiceTokenCalledExpected := 2
	numberOfTimesVerifyUserTokenCalledExpected := 2

	// Counters tracking how many times the Pep functions are called:
	numberOfTimesGetServiceTokenCalled := 0
	numberOfTimesVerifyUserTokenCalled := 0

	// Create health check configuation:
	healthCheckConfig := createConfig(apiKeyExpected)

	// Create test Pep wrapper with implementations of the wrapper functions:
	getServiceTokenValidator := func(healthCheck HealthCheckConfig) (string, error) {
		numberOfTimesGetServiceTokenCalled++
		assert.Equal(t, apiKeyExpected, healthCheck.IAMApiKey)
		return tokenStringExpected, nil // Success
	}
	verifyUserTokenValidator := func(tokenString string) (bool, error) {
		numberOfTimesVerifyUserTokenCalled++
		assert.Equal(t, tokenStringExpected, tokenString)
		return true, nil // Success
	}
	pepWrapper := NewPepWrapperForUnitTest(getServiceTokenValidator, verifyUserTokenValidator)

	// Create test NewRelic wrapper with implementations of the wrapper functions:
	sendTransactionValidator := func(transactionName string, success bool, err string) {
		assert.Equal(t, transactionNameExpected, transactionName)
		assert.Equal(t, successExpected, success)
		assert.Equal(t, errStringExpected, err)
	}
	nrWrapperForTest := NewNRWrapperForUnitTest(sendTransactionValidator)

	// Create health check and verify result:
	healthCheck := NewHealthCheck(healthCheckConfig, pepWrapper, nrWrapperForTest)
	healthCheck.CheckHealthAndTrackRecovery()
	assert.Equal(t, isHealthyExpected, healthCheck.IsHealthy())
	assert.Equal(t, isMonitorExpected, healthCheck.IsMonitor())
	assert.Equal(t, numberOfTimesGetServiceTokenCalledExpected, numberOfTimesGetServiceTokenCalled)
	assert.Equal(t, numberOfTimesVerifyUserTokenCalledExpected, numberOfTimesVerifyUserTokenCalled)
}

func TestCheckHealthAndTrackRecoveryWhenIsNotHealthy(t *testing.T) {
	// Tests that IAM health can be tracked when IAM is down

	// Expected values:
	apiKeyExpected := "1234567890"
	tokenStringExpected := "abcd.abcd.abcd"
	transactionNameExpected := "IAMHealthCheck"
	successExpected := false
	errStringExpected := "Something bad happened"
	isHealthyExpected := false
	isMonitorExpected := true // should be false here if we wait for long enough
	numberOfTimesGetServiceTokenCalledExpected := 3
	numberOfTimesVerifyUserTokenCalledExpected := 0

	// Counters tracking how many times the Pep functions are called:
	numberOfTimesGetServiceTokenCalled := 0
	numberOfTimesVerifyUserTokenCalled := 0

	// Create health check configuation:
	healthCheckConfig := createConfig(apiKeyExpected)

	// Create test Pep wrapper with implementations of the wrapper functions:
	getServiceTokenValidator := func(healthCheck HealthCheckConfig) (string, error) {
		numberOfTimesGetServiceTokenCalled++
		assert.Equal(t, apiKeyExpected, healthCheck.IAMApiKey)
		return "", errors.New(errStringExpected) // Error
	}
	verifyUserTokenValidator := func(tokenString string) (bool, error) {
		numberOfTimesVerifyUserTokenCalled++
		assert.Equal(t, tokenStringExpected, tokenString)
		return true, nil // Success
	}
	pepWrapper := NewPepWrapperForUnitTest(getServiceTokenValidator, verifyUserTokenValidator)

	// Create test NewRelic wrapper with implementations of the wrapper functions:
	sendTransactionValidator := func(transactionName string, success bool, err string) {
		assert.Equal(t, transactionNameExpected, transactionName)
		assert.Equal(t, successExpected, success)
		assert.Equal(t, errStringExpected, err)
	}
	nrWrapperForTest := NewNRWrapperForUnitTest(sendTransactionValidator)

	// Create health check and verify result:
	healthCheck := NewHealthCheck(healthCheckConfig, pepWrapper, nrWrapperForTest)
	healthCheck.CheckHealthAndTrackRecovery()
	assert.Equal(t, isHealthyExpected, healthCheck.IsHealthy())
	//assert.Equal(t, isHealthyExpected, healthCheck.CheckHealthAndTrackRecovery())
	assert.Equal(t, isMonitorExpected, healthCheck.IsMonitor())
	assert.Equal(t, numberOfTimesGetServiceTokenCalledExpected, numberOfTimesGetServiceTokenCalled)
	assert.Equal(t, numberOfTimesVerifyUserTokenCalledExpected, numberOfTimesVerifyUserTokenCalled)
}

func createConfig(apiKey string) HealthCheckConfig {
	healthCheckConfig := HealthCheckConfig{
		IAMUrl:                  "https://iam.test.cloud.ibm.com",
		IAMApiKey:               apiKey,
		IAMTimeoutInSecs:        30,
		HealthMaxAttempts:       2,
		HealthSecsBeforeRetry:   1,
		HealthSecsBetweenChecks: 2,
		TestMode:                true, // running in test mode
	}
	return healthCheckConfig
}
