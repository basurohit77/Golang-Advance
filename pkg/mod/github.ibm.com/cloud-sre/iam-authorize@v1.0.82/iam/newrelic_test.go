package iam

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/cloud-sre/iam-authorize/monitoring"
)

// TestNewRelicWithNoApp tests when no monitoring app is set
func TestNewRelicWithNoApp(t *testing.T) {
	// Create config:
	nrWrapperConfig := NRWrapperConfig{
		Environment: "staging",
		Region: "us-east",
	}

	// Create wrapper:
	nrWrapper := NewNRWrapper(nrWrapperConfig)
	nrWrapper.SendTransaction("IAMHealthCheck", true, "")

	// Send transactions:
	nrWrapper.SendTransaction("IAMHealthCheck-unit-test", false, "Something bad happened")
}

// TestNewRelicWithApp tests when monitoring app is set
func TestNewRelicWithApp(t *testing.T) {
	// Create New Relic test app (using fake NewRelic licence key):
	nrApp, nrErr := NewDefaultNRApp("", "1234567890123456789012345678901234567890")
	assert.NotNil(t, nrApp, "nrApp should not be nil, but is nil")
	assert.Nil(t, nrErr, "Error should be nil, but is not")

	// Create New Relic prod app (using fake NewRelic licence key):
	nrAppProd, nrErrProd := NewDefaultNRApp(NRAppEnvProd, "1234567890123456789012345678901234567890")
	assert.NotNil(t, nrAppProd, "nrAppProd should not be nil, but is nil")
	assert.Nil(t, nrErrProd, "nrErrProd should be nil, but is not")

	// Create New Relic staging app (using fake NewRelic licence key):
	nrAppStage, nrErrStage := NewDefaultNRApp(NRAppEnvStage, "1234567890123456789012345678901234567890")
	assert.NotNil(t, nrAppStage, "nrAppStage should not be nil, but is nil")
	assert.Nil(t, nrErrStage, "nrErrStage should be nil, but is not")

	// Create New Relic dev app (using fake NewRelic licence key):
	nrAppDev, nrErrDev := NewDefaultNRApp(NRAppEnvDev, "1234567890123456789012345678901234567890")
	assert.NotNil(t, nrAppDev, "nrAppDev should not be nil, but is nil")
	assert.Nil(t, nrErrDev, "nrErrDev should be nil, but is not")

	// Create Instana test sensor
	sensor := monitoring.NewDefaultMonitoringApp("")
	assert.NotNil(t, sensor, "sensor should not be nil, but is nil")

	// Create Instana prod sensor
	sensorProd := monitoring.NewDefaultMonitoringApp(monitoring.AppEnvProd)
	assert.NotNil(t, sensorProd, "sensorProd should not be nil, but is nil")

	// Create Instana staging sensor
	sensorStage := monitoring.NewDefaultMonitoringApp(monitoring.AppEnvStage)
	assert.NotNil(t, sensorStage, "sensorStage should not be nil, but is nil")

	// Create Instana dev sensor
	sensorDev := monitoring.NewDefaultMonitoringApp(monitoring.AppEnvDev)
	assert.NotNil(t, sensorDev, "sensorDev should not be nil, but is nil")

	// Create config:
	nrWrapperConfig := NRWrapperConfig{
		NRAppV3: nrAppStage,
		InstanaSensor: sensorStage,
		Environment: "staging",
		Region: "us-east",
	}

	// Create wrapper:
	nrWrapper := NewNRWrapper(nrWrapperConfig)

	// Send transactions:
	nrWrapper.SendTransaction("IAMHealthCheck-unit-test", true, "")
	nrWrapper.SendTransaction("IAMHealthCheck-unit-test", false, "Something bad happened")
}
