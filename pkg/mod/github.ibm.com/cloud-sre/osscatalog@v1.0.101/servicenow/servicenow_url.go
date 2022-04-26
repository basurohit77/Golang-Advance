package servicenow

const (
	// serviceNowURLv1 is the base URL for accessing ServiceNow (API v1)
	serviceNowURLv1     = "https://watson.service-now.com/api/ibmwc/cloud_offering_api"
	serviceNowDevURLv1  = "https://watsondev.service-now.com/api/ibmwc/cloud_offering_api"
	serviceNowTestURLv1 = "https://watsontest.service-now.com/api/ibmwc/cloud_offering_api"

	// serviceNow*URLv2 is the base URL for accessing ServiceNow (API v2)
	serviceNowWatsonURLv2       = "https://watson.service-now.com/api/ibmwc/service"
	serviceNowWatsonDevURLv2    = "https://watsondev.service-now.com/api/ibmwc/service"
	serviceNowWatsonTestURLv2   = "https://watsontest.service-now.com/api/ibmwc/service"
	serviceNowCloudfedURLv2     = "https://cloudfed.servicenowservices.com/api/ibmwc/service"
	serviceNowCloudfedDevURLv2  = "https://cloudfeddev.servicenowservices.com/api/ibmwc/service"
	serviceNowCloudfedTestURLv2 = "https://cloudfedtest.servicenowservices.com/api/ibmwc/service"

	// serviceNow*TokenName is the name of the token in the keyfile, used to authenticate to ServiceNow
	serviceNowWatsonTokenName       = "servicenow-watson"
	serviceNowWatsonDevTokenName    = "servicenow-watsondev"
	serviceNowWatsonTestTokenName   = "servicenow-watsontest"
	serviceNowCloudfedTokenName     = "servicenow-cloudfed"
	serviceNowCloudfedDevTokenName  = "servicenow-cloudfeddev"
	serviceNowCloudfedTestTokenName = "servicenow-cloudfedtest"

	// serviceNowUser is the value to be passed in the "updated_by" header when using the API v2
	serviceNowUser = "RMC /Global Catalog.SA"
)

//const apiVersion = apiV1

const apiVersion = apiV2
