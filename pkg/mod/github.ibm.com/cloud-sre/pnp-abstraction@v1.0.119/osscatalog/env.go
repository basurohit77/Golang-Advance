package osscatalog

const (

	// OSSCatalogCacheTime is the minimum number of minutes between cache refresh time
	OSSCatalogCacheTime = "PNP_OSSCATALOG_CACHE_TIME"

	// OSSCatalogCatYPKeyLabel is the NAME used for catalog-yp (see below)
	OSSCatalogCatYPKeyLabel = "PNP_CATALOG_CATYP_NAME"
	// OSSCatalogCatYPKeyValue is the KEY used for catalog-yp (see below)
	OSSCatalogCatYPKeyValue = "PNP_CATALOG_CATYP_VALUE"

	// OSSCatalogCatYS1KeyLabel is the NAME used for catalog-ys1 (see below)
	OSSCatalogCatYS1KeyLabel = "PNP_CATALOG_CATYS1_NAME"
	// OSSCatalogCatYS1KeyValue is the KEY used for catalog-ys1 (see below)
	OSSCatalogCatYS1KeyValue = "PNP_CATALOG_CATYS1_VALUE"
)

// For the OSS catalog key information, the following is an example
// of the keyfile and helps to understand where everything is placed.
// Currently we are only defining the values we need at this point, more
// may be added later.
//                             | {
//                             |     "yp": {
//                             |         "name": "osscatYPPlatformKey1",
//                             |         "key": "*************"
//                             |     },
//                             |     "ys1": {
//                             |         "name": "osscatPlatformKey1",
//                             |         "key": "**************"
//                             |     },
//                             |     "catalog-yp": {
// OSSCatalogCatYPKeyName      |         "name": "osscatYPPlatformKey1",
// OSSCatalogCatYPKeyValue     |         "key": "************"
//                             |     },
//                             |     "catalog-ys1": {
// OSSCatalogCatYS1KeyName     |         "name": "osscatPlatformKey1",
// OSSCatalogCatYS1KeyValue    |         "key": "bwfNwWgFkHqasgiG_xcZmChpLuMO2CaiPPUO35IQ2SyD"
//                             |     },
//                             |     "osscat-service-ys1": {
//                             |         "name": "osscatServiceKey1",
//                             |         "id": "ServiceId-f041f3f2-1fb0-4f0b-afb4-3f46ee937e1b",
//                             |         "key": "***"
//                             |     },
//                             |     "servicenow-watsondev": {
//                             |         "name": "ServiceNow Dev token for dpj",
//                             |         "token": "***"
//                             |     },
//                             |     "oss-apiplatform": {
//                             |         "name": "API Plaform API Key (Doctor) for dpj",
//                             |         "key": "Y-82WATEABChIrlvlTPEXwTN-c7AnOBJkjuMCmbeooef"
//                             |     }
//                             | }

// OSSCatalogCredentialLookup provides an array of all the settings we should attempt to find for the OSS catalog
// Strings come in 3's.  First in the main key identifyer, then the name (aka label) for that key, then the value.
var OSSCatalogCredentialLookup = []string{"catalog-yp", OSSCatalogCatYPKeyLabel, OSSCatalogCatYPKeyValue, "catalog-ys1", OSSCatalogCatYS1KeyLabel, OSSCatalogCatYS1KeyValue}
