package iam

import (
	"fmt"
	"log"
	"sync"
	"time"
)

/*
Checks the health of IAM periodically in a background go routine. The current health result is available via the
IsHealthy function and is optionally reported to monitor so that alerting can be setup. How the testing is performed,
including how retries are handled, is determined by the health check configuration. Under the covers the IAM PnP
library is used without caching to verify whether IAM is up and running.

Example usage:

import (
    "log"
	"os"
	"time"
    "github.ibm.com/cloud-sre/iam-authorize/iam"
    newrelicPreModules "github.com/newrelic/go-agent"
    "github.com/newrelic/go-agent/v3/newrelic"
)

...

nrApp, err := iam.NewDefaultNRApp(iam.NRAppEnvDev, monitoringLicense)
if err != nil {
  log.Println(err.Error())
  return
}

nrWrapperConfig := iam.NRWrapperConfig{
	NRApp: nrApp,
	NRAppV3: nrAppV3,
	InstanaSensor: sensor,
	Environment: "staging",
	Region: "us-east",
}

iamHealthCheck := iam.NewDefaultHealthCheck(os.Getenv("IAM_PUBLIC_URL"), os.Getenv("API_PLATFORM_API_KEY"), nrWrapperConfig)
*/

// HealthCheckConfig is the configuration needed to create a health check instance.
type HealthCheckConfig struct {
	IAMUrl                  string // IAM URL - for example, https://iam.cloud.ibm.com
	IAMApiKey               string // IAM API Key
	IAMTimeoutInSecs        int    // Timeout in seconds of IAM Pep API calls
	HealthMaxAttempts       int    // Maximum times this health monitor will try testing IAM before reporting down
	HealthSecsBeforeRetry   int    // Number of seconds between retries (related to HealthMaxAttempts above)
	HealthSecsBetweenChecks int    // Number of seconds to wait before testing IAM again
	TestMode                bool   // Whether or not we are running in unit test mode or not
}

// HealthCheck is the interface implemented by the health check and represents the functions that are available. Note
// that typically the CheckHealth and MonitorIAMUntilHealthy functions would not be called by code outside of this file.
type HealthCheck interface {
	CheckHealth() (bool, error)
	IsHealthy() bool
	IsMonitor() bool
	CheckHealthAndTrackRecovery() bool
	MonitorIAMUntilHealthy()
	UpdateIAMURL(iamURL string)
}

// NewDefaultTestHealthCheck creates a health check that tests the health of the IAM test (staging) environment. The
// monConfig parameter is optional and can be nil if reporting to monitoring tool is not needed. The iamAPIKey parameter must
// be a valid test IAM API key. Other default configuration values are used.
func NewDefaultTestHealthCheck(iamAPIKey string, monConfig *NRWrapperConfig) HealthCheck {
	return NewDefaultHealthCheck("https://iam.test.cloud.ibm.com", iamAPIKey, monConfig)
}

// NewDefaultProductionHealthCheck creates a health check that tests the health of the IAM production environment. The
// monConfig parameter is optional and can be nil if reporting to monitoring tool is not needed. The iamAPIKey parameter must
// be a valid production IAM API key. Other default configuration values are used.
func NewDefaultProductionHealthCheck(iamAPIKey string, monConfig *NRWrapperConfig) HealthCheck {
	return NewDefaultHealthCheck("https://iam.cloud.ibm.com", iamAPIKey, monConfig)
}

// NewDefaultHealthCheck creates a health check for testing IAM with default values. The iamURL parameter must be a
// valid IAM URL, the iamAPIKey parameter must be a valid IAM API key and must match the environment of the IAM URL,
// and the monConfig parameter is optional and can be nil if reporting to monitoring tool is not needed.
func NewDefaultHealthCheck(iamURL string, iamAPIKey string, monConfig *NRWrapperConfig) HealthCheck {
	// Create health check configuration using default values:
	healthCheckConfig := HealthCheckConfig{
		IAMUrl:                  iamURL,
		IAMApiKey:               iamAPIKey,
		IAMTimeoutInSecs:        30,
		HealthMaxAttempts:       2,
		HealthSecsBeforeRetry:   120, // 2 minutes
		HealthSecsBetweenChecks: 300, // 5 minutes
		TestMode:                false,
	}
	// Default pep wrapper talks to IAM through the IAM Pep API:
	pep := NewPepWrapper()
	if monConfig != nil {
		monConfig.URL = iamURL
		mon := NewNRWrapper(*monConfig)
		return NewHealthCheck(healthCheckConfig, pep, mon)
	}
	return NewHealthCheck(healthCheckConfig, pep, nil)
}

// NewHealthCheck creates a health check for testing IAM using the provided health check config, Pep wrapper, and
// optional monitor wrapper (nil if not needed).
func NewHealthCheck(config HealthCheckConfig, pep PepWrapper, mon NRWrapper) HealthCheck {
	healthCheck := &healthCheck{}
	healthCheck.config = config
	healthCheck.pep = pep
	healthCheck.mon = mon
	healthCheck.isHealthy = true  // assume initially healthy
	healthCheck.isMonitor = false // assume initially no background health check monitor running
	if healthCheck.config.TestMode {
		healthCheck.MonitorIAMUntilHealthy() // for testing, run in current go routine
	}
	return healthCheck
}

// healthCheck is the implementation of the health check interface.
type healthCheck struct {
	config    HealthCheckConfig
	muConfig  sync.Mutex
	pep       PepWrapper
	mon       NRWrapper
	mu        sync.Mutex
	isHealthy bool
	muMonitor sync.Mutex
	isMonitor bool
}

// CheckHealthAndTrackRecovery performs a single check of the IAM health, and if unhealthy launches a background thread to
// continually check the IAM health until IAM is deemed healthy.
func (check *healthCheck) CheckHealthAndTrackRecovery() bool {
	FCT := "CheckHealthAndTrackRecovery: "
	health, err := check.CheckHealth()
	log.Println(fmt.Sprintf("%sINFO: Result of checking IAM health. Err = %v ; health = %t", FCT, err, health))
	if err == nil && health {
		// Notify New Relic and Instana:
		if check.mon != nil {
			check.mon.SendTransaction("IAMHealthCheck", health, "")
		}
		// Healthy, just return:
		return true
	}
	// Unhealthy, so set the isHealthy value to false:
	check.mu.Lock()
	check.isHealthy = health
	check.mu.Unlock()
	// Notify New Relic and Instana:
	if check.mon != nil {
		errString := ""
		if err != nil {
			errString = err.Error()
		}
		check.mon.SendTransaction("IAMHealthCheck", health, errString)
	}

	// Check if there is an existing health check monitor
	// Start a new instance of MonitorIAMUntilHealthy() only when there is no existing background thread running
	check.muMonitor.Lock()
	if check.isMonitor == false {
		// When there is no existing monitor running, launch background thread:
		check.isMonitor = true
		check.muMonitor.Unlock()

		go check.MonitorIAMUntilHealthy()
	} else {
		// Monitor exists, no action:
		check.muMonitor.Unlock()
	}
	return false
}

// MonitorIAMUntilHealthy runs an endless loop that periodically checks the IAM health and updates the health status that is
// returned by the isHealthy function and optionally reports the status to monitor. When the IAM health is determined to be
// OK, this function ends.
func (check *healthCheck) MonitorIAMUntilHealthy() {
	FCT := "MonitorIAMUntilHealthy: "
	for {
		healthConfig := check.getConfiguration()
		health := false // assume unhealthy to start
		var err error
		for i := 1; i <= healthConfig.HealthMaxAttempts; i++ {
			if i != 1 {
				// Sleep between retries:
				time.Sleep(time.Duration(healthConfig.HealthSecsBeforeRetry) * time.Second)
			}
			health, err = check.CheckHealth()
			if err == nil && health {
				// Check was successful, break out of retry loop:
				break
			}
			// Check was unsuccessful, log information and sleep before retrying:
			log.Println(fmt.Sprintf("%sERROR: Problem occurred checking health. Err = %s ; health = %t",
				FCT, err, health))
		}
		log.Println(fmt.Sprintf("%sINFO: Result of checking %s IAM health: Err = %v ; health = %t", FCT, healthConfig.IAMUrl,
			err, health))
		// Update health status:
		check.mu.Lock()
		check.isHealthy = health
		check.mu.Unlock()
		// Notify New Relic and Instana:
		if check.mon != nil {
			errString := ""
			if err != nil {
				errString = err.Error()
			}
			check.mon.SendTransaction("IAMHealthCheck", check.isHealthy, errString)
		}
		if healthConfig.TestMode || health {
			// Stop background check if we are in test mode or IAM is found to be healthy:
			check.muMonitor.Lock()
			check.isMonitor = false
			check.muMonitor.Unlock()
			break
		} else {
			// Sleep until next check:
			time.Sleep(time.Duration(healthConfig.HealthSecsBetweenChecks) * time.Second)
		}
	}
}

// CheckHealth does a single check of the health of IAM. The check consists of obtaining an IAM token from an IAM API
// key and then validating the obtained user token. No retries are done except for the call to obtain an IAM token
// from an IAM API key which is done by the IAM Pep library.
func (check *healthCheck) CheckHealth() (bool, error) {

	healthConfig := check.getConfiguration()

	// Get IAM token from IAM API key (note that this call will automatically and quickly be retried by the IAM Pep
	// library if there is a failure):
	token, err := check.pep.GetToken(healthConfig)
	if err != nil {
		return false, err
	}
	// Verify that the token is valid (there is no retry here in the Pep library):
	isTokenValid, err := check.pep.GetClaims(token)
	if err != nil {
		return false, err
	}

	return isTokenValid, nil
}

// IsMonitor returns if there is a background health check running.
func (check *healthCheck) IsMonitor() bool {
	check.mu.Lock()
	defer check.mu.Unlock()
	return check.isMonitor
}

// IsHealthy returns the last result of the IAM health check test.
func (check *healthCheck) IsHealthy() bool {
	check.mu.Lock()
	defer check.mu.Unlock()
	return check.isHealthy
}

// getConfiguration returns the current health configuration
func (check *healthCheck) getConfiguration() HealthCheckConfig {
	check.muConfig.Lock()
	defer check.muConfig.Unlock()
	return check.config
}

// UpdateIAMURL updates the IAM URL used by the health check.
func (check *healthCheck) UpdateIAMURL(iamURL string) {
	if iamURL != "" {
		check.muConfig.Lock()
		defer check.muConfig.Unlock()
		fmt.Printf("UpdateIAMURL: Updating IAM URL to: %s\n", iamURL)
		check.config.IAMUrl = iamURL
	}
}
