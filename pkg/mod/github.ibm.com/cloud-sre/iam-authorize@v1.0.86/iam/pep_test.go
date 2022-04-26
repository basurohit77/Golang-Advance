package iam

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/IAM/pep/v3"
)

func TestPepWrapper(t *testing.T) {
	// This test only verifies negative cases since we don't want to use a real API key in the test.

	// Create wrapper:
	pepWrapper := NewPepWrapper()

	apiKey := "1234567890"

	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      apiKey,
		LogLevel:    pep.LevelError,
		AuthzRetry:  true,
		PDPTimeout:  time.Duration(30) * time.Second,
	}

	pep.Configure(config)

	hCheck := createConfig(apiKey)

	// Set config:
	// pepWrapper.SetConfig(config)

	// Try to get a token from an invalid API key and verify result:
	token, pepErr := pepWrapper.GetToken(hCheck)
	assert.Equal(t, token, "", "Token value should be an empty string, but is not")
	assert.NotNil(t, pepErr, "Err should not be nil, but is nil")

	// Try to validate invalid token and verify result:
	token = "abcd.abcd.abcd"
	isTokenValid, pepErr := pepWrapper.GetClaims(token)
	assert.Equal(t, isTokenValid, false, "Token check should be false, but is not")
	assert.NotNil(t, pepErr, "Err should not be nil, but is nil")
}
