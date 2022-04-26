package shared

const (
	// APIDoctorMaintenancePath is the Doctor maintenances API path
	APIDoctorMaintenancePath = "/api/v1/doctor/maintenances"
	// APISnowPathPrefix is the base path for all SNow v1 endpoints
	APISnowPathPrefix = "/api/v1/snow"
	// APIGhePathPrefix is the base path for all Ghe v1 endpoints
	APIGhePathPrefix = "/api/v1/ghe"
	// APISnowCasesPath is the cases API path
	APISnowCasesPath = APISnowPathPrefix + "/cases"
	// APISnowIncidentsPath is the incidents API path
	APISnowIncidentsPath = APISnowPathPrefix + "/incidents"
	// APISnowBSPNPath is the bspn API path
	APISnowBSPNPath = APISnowPathPrefix + "/bspn"
	// APISnowChangesPath is the changes API path
	APISnowChangesPath = APISnowPathPrefix + "/changes"
	// APIGheAnnouncementsPath is the security notice and announcement API path
	APIGheAnnouncementsPath = APIGhePathPrefix + "/announcements"
	// APIHealthzPath is the healthz API path
	APIHealthzPath = "/api/v1/pnp/hooks/healthz"
)
