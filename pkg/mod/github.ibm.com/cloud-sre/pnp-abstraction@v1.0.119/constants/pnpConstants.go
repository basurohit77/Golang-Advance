package crn

// This file contains defaul constant values user for crn in the cloud for pnp

// The cname segment identifies the cloud instance. This is an alphanumeric identifier that uniquely identifies
// the cloud instance that contains the resource. A cname effectively identifies an independent control plane that
// is owns the identified resource.
// https://github.ibm.com/ibmcloud/builders-guide/blob/master/specifications/crn/CRN.md
const (
	// CrnVersion The version segment identifies the version of the CRN format. Currently the only valid version segment value is v1
	CrnVersion = "v1"
	// CrnNamePrdCustIBMCloud All production environments. This is the only one exposed to external users.
	CrnNamePrdCustIBMCloud = "bluemix"
	// CrnNamePrdIntIBMCloud IBM CIO office environments.
	CrnNamePrdIntIBMCloud = "internal"
	// CrnNameIntStgIBMCloud Internal dev/test/staging environments before going in production.
	CrnNameIntStgIBMCloud = "staging"
)

// The ctype segments identifies the type of cloud instance represented by the specified cname
const (
	// CrnTypePublic All services that are available from the public catalog: CF services, Softlayer, etc..
	CrnTypePublic = "public"
	// CrnTypeDedicated Only for current IBM Cloud Dedicated environments.
	CrnTypeDedicated = "dedicated"
	// CrnTypeLocal All services deployed in customer's own environments: e.g IBM Private Cloud, BlueBox Local.
	CrnTypeLocal = "local"
)

const (
	// GenericCRNMask is a generic version of the CRN mask that encompasses all resources
	GenericCRNMask = "crn:v1:" + CrnNamePrdCustIBMCloud + ":" + CrnTypePublic + "::::::"
	// GenericCRNMaskShort is a generic short version of the CRN mask that encompasses all resources
	GenericCRNMaskShort = "crn:" + CrnVersion + ":" + CrnNamePrdCustIBMCloud + ":" + CrnTypePublic + ":"
)

const (

	// APIManagementAdminApisURL API management host
	APIManagementAdminApisURL = "https://api-oss.bluemix.net/apis"
)
