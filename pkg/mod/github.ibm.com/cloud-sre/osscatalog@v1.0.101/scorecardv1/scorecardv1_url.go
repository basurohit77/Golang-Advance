package scorecardv1

const (
	// scorecardv1KeyName is the name of API key in the keyfile, used to authenticate to Doctor ScorecardV1
	scorecardv1KeyName = "catalog-yp"
	//scorecardv1KeyName = "oss-apiplatform-NEW2"
	//scorecardv1KeyName = "oss-apiplatform"
	//scorecardv1KeyName = "oss-special-key"

	// scorecardv1ServiceEntryURL is the URL for fetching one service record from Doctor ScorecardV1 (using the OSS API)
	scorecardv1ServiceEntryURL = "https://pnp-api-oss.cloud.ibm.com/scorecard/api/segmenttribe/v1/services?name=%s"

	// scorecardv1ServiceListURL is the URL for listing all service records from Doctor ScorecardV1 (using the OSS API)
	scorecardv1ServiceListURL = "https://pnp-api-oss.cloud.ibm.com/scorecard/api/segmenttribe/v1/services?nopagination=true"

	// scorecardv1ServicesNamesURL is the URL to get a list of all service names from Doctor ScorecardV1
	scorecardv1ServicesNamesURL = "https://pnp-api-oss.cloud.ibm.com/scorecard/api/internal/v1/services_crn_list"

	// scorecardv1SegmentsURL is the base URL for listing all Segments registered in Doctor ScorecardV1,
	// and from there access the Tribes and individual services
	scorecardv1SegmentsURL = "https://pnp-api-oss.cloud.ibm.com/scorecard/api/segmenttribe/v1/segments?_limit=200"
	// /* INTERNAL API */ scorecardv1SegmentsURL = "https://doctor.cloud.ibm.com/taishan_api/router/scorecard/api/segmenttribe/v1/segments"
	//scorecardv1SegmentsURL = "https://pnp-api-oss.cloud.ibm.com/proxy_scorecard/api/segmenttribe/v1/segments"

	// scorecardv1TribesURL is the base URL for listing all Tribes associated with one given segment in ScorecardV1.
	// Note that this URL should not need to be explicitly specified, because it is normally included in the response
	// when reading segment information from ScorecardV1.
	// However, we allow a non-empty value to override the value returned by ScorecardV1 itself.
	scorecardv1TribesURL = "" // empty value is the default
	// /* INTERNAL API */ scorecardv1TribesURL = "https://doctor.cloud.ibm.com/taishan_api/router/scorecard/api/segmenttribe/v1/segments/%s/tribes"

	// scorecardv1ServiceDetailURL is the URL for getting a service full details from Doctor ScorecardV1 (using the internal Doctor API)
	/* INTERNAL API */
	scorecardv1ServiceDetailURL = "https://doctor.bluemix.net/taishan_api/router/servicedb/cache/service_configs?name=%s"

	// scorecardv1DetailListURL is the URL for getting a lisr of services with full details from Doctor ScorecardV1 (using the internal Doctor API)
	// /* INTERNAL API */ scorecardv1DetailListURL = "https://doctor.bluemix.net/taishan_api/router/servicedb/cache/services"
	scorecardv1DetailListURL = "https://pnp-api-oss.cloud.ibm.com/scorecard/api/segmenttribe/v1/rawservices"
)
