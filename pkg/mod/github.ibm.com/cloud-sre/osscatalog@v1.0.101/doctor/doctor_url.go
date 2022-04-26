package doctor

// URLs for access to Environment information in Doctor
const (
	//doctorAllEnvURL = "https://cloud-oss-metadata.bluemix.net/cloud-oss/metadata/allEnv"
	//doctorAllEnvURL = "https://cloud-oss-metadata.cloud.ibm.com/cloud-oss/metadata/allEnv"
	doctorAllEnvURL = "https://169.46.226.211/cloud-oss/metadata/allEnv"

	//doctorNameMappingURL = "https://api-oss.bluemix.net/doctorapi/api/doctor/namemapping"
	doctorNameMappingURL = "https://pnp-api-oss.cloud.ibm.com/doctorapi/api/doctor/namemapping"

	//doctorRegionIDURL = "https://api-oss.bluemix.net/doctorapi/api/doctor/regionids"
	doctorRegionIDURL = "https://pnp-api-oss.cloud.ibm.com/doctorapi/api/doctor/regionids"

	doctorKeyName = "oss-apiplatform"
	//doctorKeyName = "catalog-yp"
)

/*
https://cloud-oss-metadata.bluemix.net/cloud-oss/metadata/allEnv
[
{
crn: "resona:dedicated:as-jp",
deployment: "resona-as-jp",
doctor: "D_RESONA",
env_name: "Resona Bank - decommissioned",
monitor: "resona.env1",
new_crn: "crn:v1:d-resona:dedicated::jp-tok::::",
region_id: "resona:prod:as-jp",
ucd: "ded-prd-sl-tok02-resona"
},
*/

/*
https://api-oss.bluemix.net/doctorapi/api/doctor/namemapping
{
result: "success",
detail: [
{
doctor_env_name: "YP_DALLAS",
jml_repo_name: "jml-pub_dalyp_278460.git",
deployment_name: "scas-yz-prod",
monitor_env_name: "ibm.env5",
ucd_env_name: "pub-prd-sl-dal09-yp",
account: "278460",
mccp_region_id: "ibm:yp:us-south",
bluemix_env_domain: "ng.bluemix.net",
crn_name: "yp_dallas:public:us-south",
new_crn_name: "crn:v1:bluemix:public::us-south::::",
EU_MANAGED: null,
bluemix_apps_domain: null,
cf_api: "https://api.ng.bluemix.net",
ace_console: "https://console.ng.bluemix.net"
},
*/

/*
https://cloud-oss-metadata.bluemix.net/cloud-oss/metadata/envTypeList
{
detail: {
BMX_DEDICATED: [
{
crn: "aa1:dedicated:us-ne",
deployment: "aa1-us-ne",
doctor: "D_AA1",
env_name: "American Airlines 1",
monitor: "aa.env1",
new_crn: "crn:v1:d-aa1:dedicated::us-east::::",
region_id: "aa1:prod:us-ne",
ucd: "ded-prd-sl-wdc04-aa1"
},
],
BMX_IBM_INTERNAL: null,
BMX_LOCAL: [],
BMX_PUBLIC: [],
CFS_DEDICATED: null,
CFS_LOCAL: null,
CFS_PUBLIC: [],
CMS: [],
CTO_ARMADA_CLUSTER: [],
DEDICATED_SHARED: [],
EU_BMX_DEDICATED: null,
EU_BMX_LOCAL: null,
EU_BMX_PUBLIC: null,
EU_WATSON_DEDICATED: null,
EU_WATSON_LOCAL: null,
EU_WATSON_PUBLIC: null,
MIS: [],
OSS_ARMADA_CLUSTER: [],
SERVICE: null,
SOFTLAYER_IACS: [],
SOFTLAYER_IIGDB2: [],
SOFTLAYER_NETWORK: [],
SOFTLAYER_SAP: [],
WATSON_ARMADA_CLUSTER: [],
WATSON_CMFRAPRD: null,
WATSON_CMPRD: [],
WATSON_CSFDEV: [],
WATSON_DDEV: [],
WATSON_DEDICATED: null,
WATSON_FRAPRD: [],
WATSON_GENERAL: [],
WATSON_HEALTH_HIPPA: [],
WATSON_HEALTH_LSC_SANDBOX: [],
WATSON_ICPDEV: [],
WATSON_LOCAL: null,
WATSON_LONPRD: [],
WATSON_PPRD: [],
WATSON_PSTG: [],
WATSON_PUBLIC: null,
WATSON_SBPRD: [],
WATSON_SKPRD: [],
WATSON_SLOPS: [],
WATSON_STAGING: null,
WATSON_SYDPRD: [],
WATSON_TOKPRD: [],
WATSON_TRPRD: [],
WDP_CLOUDANT: [],
WDP_COMPOSE: [
{
doctor: "WDP_COMPOSE"
}
]
},
result: "success"
}
*/

/*
https://api-oss.bluemix.net/doctorapi/api/doctor/regionids
"resources": [
	DEBUG:       {
	DEBUG:          "crn": "",
	DEBUG:          "decommissioned": "true",
	DEBUG:          "id": "vmware-46-chrisahl2:prod:eu-gb",
	DEBUG:          "name": "Rack46 2 - decommissioned"
	DEBUG:       },
	DEBUG:       {
	DEBUG:          "crn": "crn:v1:d-showroom-3:dedicated::us-south::::",
	DEBUG:          "decommissioned": "true",
	DEBUG:          "id": "sr3:prod:us-south",
	DEBUG:          "name": "Showroom 3 - decommissioned"
	DEBUG:       },
	      {
DEBUG:          "crn": "crn:v1:softlayer:public::ams01::::",
DEBUG:          "decommissioned": "false",
DEBUG:          "id": "SoftLayer AMS01 (265592)",
DEBUG:          "name": "SoftLayer AMS01 (265592)"
DEBUG:       },
*/

/*
http://api.opscenter.bluemix.net:4574/dlt/environments
*/
