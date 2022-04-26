package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/IAM/basiclog"
)

func TestDefaultConfig(t *testing.T) {
	tm := &TokenManager{}
	optionalConfig := &ExtendedConfig{}

	// Handle endpoints
	err := tm.envConfigure(0, &(optionalConfig.Endpoints))

	assert.Nil(t, err)

	// configure token
	tm.config.APIKey = "APIKey"

	// Configure other extended params
	err = tm.tokenParamConfigure(optionalConfig)
	assert.Nil(t, err)
	assert.Equal(t, tm.config.ExtendedConfig.Endpoints.TokenEndpoint, ProdHostURL+TokenPath)
	assert.Equal(t, tm.config.ExtendedConfig.Endpoints.KeyEndpoint, ProdHostURL+KeyPath)
}

func TestDefaultExpiryConfiguration(t *testing.T) {

	tm := &TokenManager{}
	err := tm.tokenParamConfigure(nil)

	assert.Nil(t, err)
	assert.Equal(t, DefaultTokenExpiry, tm.config.ExtendedConfig.TokenExpiry)
}

func TestClientProvidedConfiguration(t *testing.T) {

	var tokenConfigCustom = &ExtendedConfig{
		ClientID:     "cust1",
		ClientSecret: "secret1",
		Scope:        "serviceScope",
		TokenExpiry:  0.80,
		Endpoints:    Endpoints{},
		retryTimeout: 10,
		HttpTimeout:  15,
		LogLevel:     LevelError,
	}

	tm := &TokenManager{}
	err := tm.tokenParamConfigure(tokenConfigCustom)

	assert.Nil(t, err)
	assert.Equal(t, 0.80, tm.config.ExtendedConfig.TokenExpiry)
	assert.Equal(t, "cust1", tm.config.ExtendedConfig.ClientID)
	assert.Equal(t, "secret1", tm.config.ExtendedConfig.ClientSecret)
	assert.Equal(t, "serviceScope", tm.config.ExtendedConfig.Scope)
	assert.Equal(t, 10, tm.config.ExtendedConfig.retryTimeout)
	assert.Equal(t, 15, tm.config.ExtendedConfig.HttpTimeout)
	assert.Equal(t, LevelError, tm.config.ExtendedConfig.LogLevel)
	assert.IsType(t, &basiclog.BasicLogger{}, tm.config.ExtendedConfig.Logger)
}

func TestMultipleClientProvidedConfiguration(t *testing.T) {

	var tokenConfigCustom1 = &ExtendedConfig{
		ClientID:     "cust1",
		ClientSecret: "secret1",
		Scope:        "customScope1",
		TokenExpiry:  0.80,
		Endpoints:    Endpoints{},
		retryTimeout: 10,
		HttpTimeout:  11,
		LogLevel:     LevelDebug,
	}

	var tokenConfigCustom2 = &ExtendedConfig{
		ClientID:     "cust2",
		ClientSecret: "secret2",
		Scope:        "customScope2",
		TokenExpiry:  0.90,
		Endpoints:    Endpoints{},
		retryTimeout: 15,
		HttpTimeout:  20,
		LogLevel:     LevelInfo,
	}

	// First instance
	tm1 := &TokenManager{}
	err := tm1.tokenParamConfigure(tokenConfigCustom1)

	assert.Nil(t, err)
	assert.Equal(t, 0.80, tm1.config.ExtendedConfig.TokenExpiry)
	assert.Equal(t, "cust1", tm1.config.ExtendedConfig.ClientID)
	assert.Equal(t, "secret1", tm1.config.ExtendedConfig.ClientSecret)
	assert.Equal(t, "customScope1", tm1.config.ExtendedConfig.Scope)
	assert.Equal(t, 10, tm1.config.ExtendedConfig.retryTimeout)
	assert.Equal(t, 11, tm1.config.ExtendedConfig.HttpTimeout)

	// Second instance
	tm2 := &TokenManager{}
	err = tm2.tokenParamConfigure(tokenConfigCustom2)

	assert.Nil(t, err)
	assert.Equal(t, 0.90, tm2.config.ExtendedConfig.TokenExpiry)
	assert.Equal(t, "cust2", tm2.config.ExtendedConfig.ClientID)
	assert.Equal(t, "secret2", tm2.config.ExtendedConfig.ClientSecret)
	assert.Equal(t, "customScope2", tm2.config.ExtendedConfig.Scope)
	assert.Equal(t, LevelDebug, tm1.config.ExtendedConfig.LogLevel)
	assert.Equal(t, 15, tm2.config.ExtendedConfig.retryTimeout)
	assert.Equal(t, 20, tm2.config.ExtendedConfig.HttpTimeout)

	// ensure first instance is still as expected
	assert.Equal(t, 0.80, tm1.config.ExtendedConfig.TokenExpiry)
	assert.Equal(t, "cust1", tm1.config.ExtendedConfig.ClientID)
	assert.Equal(t, "secret1", tm1.config.ExtendedConfig.ClientSecret)
	assert.Equal(t, "customScope1", tm1.config.ExtendedConfig.Scope)
	assert.Equal(t, LevelInfo, tm2.config.ExtendedConfig.LogLevel)
	assert.Equal(t, 10, tm1.config.ExtendedConfig.retryTimeout)
	assert.Equal(t, 11, tm1.config.ExtendedConfig.HttpTimeout)
}

func TestTokenValidatorConfig(t *testing.T) {
	stageTV, err := NewTokenValidator(StagingHostURL + KeyPath)

	assert.Nil(t, err)
	assert.Equal(t, StagingHostURL+KeyPath, stageTV.hostEndpoint)

	prodTV, err := NewTokenValidator(ProdHostURL + KeyPath)
	assert.Nil(t, err)
	assert.Equal(t, ProdHostURL+KeyPath, prodTV.hostEndpoint)
}
