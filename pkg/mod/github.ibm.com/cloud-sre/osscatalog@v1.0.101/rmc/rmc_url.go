package rmc

const (
	// rmcKeyName is the name of the key in the keyfile to use to access RMC
	rmcKeyName = "rmc-staging"

	// rmcSummaryURL is the URL for accessing service summary info in RMC
	rmcSummaryURL  = rmcSummaryURL3
	rmcSummaryURL1 = "https://api.rmc.test.cloud.ibm.com/v1/services/%s/summary?onlyData=true"
	rmcSummaryURL2 = "https://api.rmconsole.test.cloud.ibm.com/v1/services/%s/summary?onlyData=true"
	rmcSummaryURL3 = "https://rmc.api.cloud.ibm.com/test/v1/services/%s/summary?onlyData=true"

	// rmcTestSummaryURL is the URL for accessing service summary info in the test instance of RMC (in test mode)
	rmcTestSummaryURL  = rmcTestSummaryURL3
	rmcTestSummaryURL1 = "https://api-test.rmc.test.cloud.ibm.com/v1/services/%s/summary?onlyData=true"
	rmcTestSummaryURL2 = "https://api.dev.rmconsole.test.cloud.ibm.com/v1/services/%s/summary?onlyData=true"
	rmcTestSummaryURL3 = "https://rmc-dev.api.test.cloud.ibm.com/test/v1/services/%s/summary?onlyData=true"
)
