package catalog

const (
	// catalogMainURL is the base URL for accessing top-level service/component entries in Global Catalog
	// (currently set to go to YS1)
	//	catalogMainURL = "https://resource-catalog.stage1.bluemix.net/api/v1"
	//catalogMainURL = "https://resource-catalog.bluemix.net/api/v1"
	catalogMainURL = "https://globalcatalog.cloud.ibm.com/api/v1"

	// catalogMainKeyName is the name of the API key from which to obtain a IAM token for access to main Global Catalog entries
	//	catalogMainKeyNameStaging = "ys1"
	catalogMainKeyName = "catalog-yp"

	// catalogOSSURLStaging is the base URL for accessing OSS entries in Global Catalog
	// (currently set to go to YS1)
	//catalogOSSURLStaging = "https://resource-catalog.stage1.bluemix.net/api/v1"
	catalogOSSURLStaging = "https://globalcatalog.test.cloud.ibm.com/api/v1"

	//  catalogOSSURLProduction = "https://dev-resource-catalog.stage1.ng.bluemix.net/api/v1"
	//catalogOSSURLProduction = "https://resource-catalog.bluemix.net/api/v1"
	catalogOSSURLProduction = "https://globalcatalog.cloud.ibm.com/api/v1"

	// CatalogOSSKeyName is the name of the API key from which to obtain a IAM token for access to main Global Catalog entries
	catalogOSSKeyNameStaging = "catalog-ys1"
	//  catalogOSSKeyNameProduction = "catalog-ys1"
	catalogOSSKeyNameProduction = "catalog-yp"

	// pricingURL is the URL for fetching information from the Pricing Catalog
	//pricingURL = "https://pricing-catalog.stage1.ng.bluemix.net/v1/pricing/plan_definitions?plan_id=%s"
	//pricingURL = "https://pricing.test.cloud.ibm.com/v1/pricing/plan_definitions?plan_id=%s"
	pricingURL = "https://pricing-catalog.bluemix.net/v1/pricing/plan_definitions?plan_id=%s"
	//pricingURL = "https://pricing-catalog.cloud.ibm.com/v1/pricing/plan_definitions?plan_id=%s"
	//pricingURL = "https://cloud.ibm.com/v1/pricing/plan_definitions?plan_id=%s"

	// pricingKeyName is the name of the API key to use when accessing the Pricing Catalog
	//pricingKeyName = "pricing-ys1"
	pricingKeyName = "pricing-yp"

	// osscatOwner is the owner attribute for Catalog entries owned by osscat@us.ibm.com (which is supposed to be the owner for all legitimate OSS entries)
	osscatOwnerStaging    = "a/1cbbf83d5a3a4dc49af8d276e4630620"
	osscatOwnerProduction = "a/79f09349b04648bf8efad29de029baa6"
)
