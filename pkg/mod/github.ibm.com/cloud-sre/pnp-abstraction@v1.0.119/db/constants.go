package db

// PnP Dictionary metadata declarations
// Add table names and tables columns and depencies here, such as indexes and contrains
// use all lower case in column names since they are case-insensitive inside the database

const (
	// DATABASE_CONFIG_FILE Database configuration file
	DATABASE_CONFIG_FILE = ".dbconfig.json"

	// DefaultRetryDelay defualt 1 second
	DefaultRetryDelay = 1 // seconds
	// DefaultMaxRetries defult maximun of entries defaul is 20;
	// 20 * defaultRetryDelay = 20 seconds max in retry
	DefaultMaxRetries = 20

	// DefaultIBMCloudCname IBM Public Cloud default cname, ctype
	DefaultIBMCloudCname = "bluemix"

	// PnP tables names add a new table here

	RESOURCE_TABLE_NAME                 = "resource_table"
	DISPLAY_NAMES_TABLE_NAME            = "display_names_table"
	VISIBILITY_JUNCTION_TABLE_NAME      = "visibility_junction_table"
	VISIBILITY_TABLE_NAME               = "visibility_table"
	TAG_JUNCTION_TABLE_NAME             = "tag_junction_table"
	TAG_TABLE_NAME                      = "tag_table"
	INCIDENT_TABLE_NAME                 = "incident_table"
	INCIDENT_JUNCTION_TABLE_NAME        = "incident_junction_table"
	MAINTENANCE_TABLE_NAME              = "maintenance_table"
	MAINTENANCE_JUNCTION_TABLE_NAME     = "maintenance_junction_table"
	CASE_TABLE_NAME                     = "case_table"
	SUBSCRIPTION_TABLE_NAME             = "subscription_table"
	WATCH_TABLE_NAME                    = "watch_table"
	WATCH_JUNCTION_TABLE_NAME           = "watch_junction_table"
	NOTIFICATION_TABLE_NAME             = "notification_table"
	NOTIFICATION_DESCRIPTION_TABLE_NAME = "notification_description_table"

	// Resource table columns names metadata

	RESOURCE_COLUMN_RECORD_ID            = "record_id"
	RESOURCE_COLUMN_PNP_CREATION_TIME    = "pnp_creation_time"
	RESOURCE_COLUMN_PNP_UPDATE_TIME      = "pnp_update_time"
	RESOURCE_COLUMN_CRN_FULL             = "crn_full"
	RESOURCE_COLUMN_STATE                = "state"
	RESOURCE_COLUMN_OPERATIONAL_STATUS   = "operational_status"
	RESOURCE_COLUMN_SOURCE               = "source"
	RESOURCE_COLUMN_SOURCE_ID            = "source_id"
	RESOURCE_COLUMN_STATUS               = "status"
	RESOURCE_COLUMN_STATUS_UPDATE_TIME   = "status_update_time"
	RESOURCE_COLUMN_REGULATORY_DOMAIN    = "regulatory_domain"
	RESOURCE_COLUMN_VERSION              = "version"
	RESOURCE_COLUMN_CNAME                = "cname"
	RESOURCE_COLUMN_CTYPE                = "ctype"
	RESOURCE_COLUMN_SERVICE_NAME         = "service_name"
	RESOURCE_COLUMN_LOCATION             = "location"
	RESOURCE_COLUMN_SCOPE                = "scope"
	RESOURCE_COLUMN_SERVICE_INSTANCE     = "service_instance"
	RESOURCE_COLUMN_RESOURCE_TYPE        = "resource_type"
	RESOURCE_COLUMN_RESOURCE             = "resource"
	RESOURCE_COLUMN_SOURCE_CREATION_TIME = "source_creation_time"
	RESOURCE_COLUMN_SOURCE_UPDATE_TIME   = "source_update_time"
	RESOURCE_COLUMN_CATEGORY_ID          = "category_id"
	RESOURCE_COLUMN_CATEGORY_PARENT      = "category_parent"
	RESOURCE_COLUMN_IS_CATALOG_PARENT    = "is_catalog_parent"
	RESOURCE_COLUMN_CATALOG_PARENT_ID    = "catalog_parent_id"
	RESOURCE_COLUMN_RECORD_HASH          = "record_hash"

	// Display Names table columns names metadata

	DISPLAYNAMES_COLUMN_RECORD_ID   = "record_id"
	DISPLAYNAMES_COLUMN_RESOURCE_ID = "resource_id"
	DISPLAYNAMES_COLUMN_NAME        = "name"
	DISPLAYNAMES_COLUMN_LANGUAGE    = "language"

	// Visibility table columns names metadata

	VISIBILITY_COLUMN_RECORD_ID   = "record_id"
	VISIBILITY_COLUMN_NAME        = "name"
	VISIBILITY_COLUMN_DESCRIPTION = "description"

	// Visibility Junction table columns names metadata

	VISIBILITYJUNCTION_COLUMN_RECORD_ID     = "record_id"
	VISIBILITYJUNCTION_COLUMN_RESOURCE_ID   = "resource_id"
	VISIBILITYJUNCTION_COLUMN_VISIBILITY_ID = "visibility_id"

	// Tag table columns names metadata

	TAG_COLUMN_RECORD_ID = "record_id"
	TAG_COLUMN_ID        = "id"

	// Tag Junction table columns names metadata

	TAGJUNCTION_COLUMN_RECORD_ID   = "record_id"
	TAGJUNCTION_COLUMN_RESOURCE_ID = "resource_id"
	TAGJUNCTION_COLUMN_TAG_ID      = "tag_id"

	// Incident table columns
	INCIDENT_COLUMN_RECORD_ID                   = "record_id"
	INCIDENT_COLUMN_PNP_CREATION_TIME           = "pnp_creation_time"
	INCIDENT_COLUMN_PNP_UPDATE_TIME             = "pnp_update_time"
	INCIDENT_COLUMN_SOURCE_CREATION_TIME        = "source_creation_time"
	INCIDENT_COLUMN_SOURCE_UPDATE_TIME          = "source_update_time"
	INCIDENT_COLUMN_START_TIME                  = "start_time"
	INCIDENT_COLUMN_END_TIME                    = "end_time"
	INCIDENT_COLUMN_SHORT_DESCRIPTION           = "short_description"
	INCIDENT_COLUMN_LONG_DESCRIPTION            = "long_description"
	INCIDENT_COLUMN_STATE                       = "state"
	INCIDENT_COLUMN_CLASSIFICATION              = "classification"
	INCIDENT_COLUMN_SEVERITY                    = "severity"
	INCIDENT_COLUMN_CRN_FULL                    = "crn_full"
	INCIDENT_COLUMN_SOURCE_ID                   = "source_id"
	INCIDENT_COLUMN_SOURCE                      = "source"
	INCIDENT_COLUMN_REGULATORY_DOMAIN           = "regulatory_domain"
	INCIDENT_COLUMN_AFFECTED_ACTIVITY           = "affected_activity"
	INCIDENT_COLUMN_CUSTOMER_IMPACT_DESCRIPTION = "customer_impact_description"
	INCIDENT_COLUMN_PNP_REMOVED                 = "pnp_removed"
	INCIDENT_COLUMN_TARGETED_URL                = "targeted_url"
	INCIDENT_COLUMN_AUDIENCE                    = "audience"

	// Incident Junction table columns names metadata

	INCIDENTJUNCTION_COLUMN_RECORD_ID   = "record_id"
	INCIDENTJUNCTION_COLUMN_RESOURCE_ID = "resource_id"
	INCIDENTJUNCTION_COLUMN_INCIDENT_ID = "incident_id"

	// Maintenance table columns names metadata

	MAINTENANCE_COLUMN_RECORD_ID              = "record_id"
	MAINTENANCE_COLUMN_PNP_CREATION_TIME      = "pnp_creation_time"
	MAINTENANCE_COLUMN_PNP_UPDATE_TIME        = "pnp_update_time"
	MAINTENANCE_COLUMN_SOURCE_CREATION_TIME   = "source_creation_time"
	MAINTENANCE_COLUMN_SOURCE_UPDATE_TIME     = "source_update_time"
	MAINTENANCE_COLUMN_SHORT_DESCRIPTION      = "short_description"
	MAINTENANCE_COLUMN_LONG_DESCRIPTION       = "long_description"
	MAINTENANCE_COLUMN_CRN_FULL               = "crn_full"
	MAINTENANCE_COLUMN_START_TIME             = "start_time"
	MAINTENANCE_COLUMN_END_TIME               = "end_time"
	MAINTENANCE_COLUMN_STATE                  = "state"
	MAINTENANCE_COLUMN_DISRUPTIVE             = "disruptive"
	MAINTENANCE_COLUMN_SOURCE_ID              = "source_id"
	MAINTENANCE_COLUMN_SOURCE                 = "source"
	MAINTENANCE_COLUMN_REGULATORY_DOMAIN      = "regulatory_domain"
	MAINTENANCE_COLUMN_RECORD_HASH            = "record_hash"
	MAINTENANCE_COLUMN_MAINTENANCE_DURATION   = "maintenance_duration"
	MAINTENANCE_COLUMN_DISRUPTION_TYPE        = "disruption_type"
	MAINTENANCE_COLUMN_DISRUPTION_DESCRIPTION = "disruption_description"
	MAINTENANCE_COLUMN_DISRUPTION_DURATION    = "disruption_duration"
	MAINTENANCE_COLUMN_COMPLETION_CODE        = "completion_code"
	MAINTENANCE_COLUMN_PNP_REMOVED            = "pnp_removed"
	MAINTENANCE_COLUMN_TARGETED_URL           = "targeted_url"
	MAINTENANCE_COLUMN_AUDIENCE               = "audience"

	// Maintenance Junction table columns names metadata

	MAINTENANCEJUNCTION_COLUMN_RECORD_ID      = "record_id"
	MAINTENANCEJUNCTION_COLUMN_RESOURCE_ID    = "resource_id"
	MAINTENANCEJUNCTION_COLUMN_MAINTENANCE_ID = "maintenance_id"

	// Case table columns names metdata

	CASE_COLUMN_RECORD_ID     = "record_id"
	CASE_COLUMN_SOURCE        = "source"
	CASE_COLUMN_SOURCE_ID     = "source_id"
	CASE_COLUMN_SOURCE_SYS_ID = "source_sys_id"

	// Subscription table columns names metadata

	SUBSCRIPTION_COLUMN_RECORD_ID      = "record_id"
	SUBSCRIPTION_COLUMN_NAME           = "name"
	SUBSCRIPTION_COLUMN_TARGET_ADDRESS = "target_address"
	SUBSCRIPTION_COLUMN_TARGET_TOKEN   = "target_token"
	SUBSCRIPTION_COLUMN_EXPIRATION     = "expiration"

	// Watch table columns names metadata

	WATCH_COLUMN_RECORD_ID          = "record_id"
	WATCH_COLUMN_SUBSCRIPTION_ID    = "subscription_id"
	WATCH_COLUMN_KIND               = "kind"
	WATCH_COLUMN_PATH               = "path"
	WATCH_COLUMN_CRN_FULL           = "crn_full"
	WATCH_COLUMN_WILDCARDS          = "wildcards"
	WATCH_COLUMN_RECORD_ID_TO_WATCH = "record_id_to_watch"
	WATCH_COLUMN_SUBSCRIPTION_EMAIL = "subscription_email"

	// Watch Junction table columns names metadata

	WATCHJUNCTION_COLUMN_RECORD_ID   = "record_id"
	WATCHJUNCTION_COLUMN_RESOURCE_ID = "resource_id"
	WATCHJUNCTION_COLUMN_WATCH_ID    = "watch_id"

	// Notification table columns names metadata

	NOTIFICATION_COLUMN_RECORD_ID              = "record_id"
	NOTIFICATION_COLUMN_PNP_CREATION_TIME      = "pnp_creation_time"
	NOTIFICATION_COLUMN_PNP_UPDATE_TIME        = "pnp_update_time"
	NOTIFICATION_COLUMN_SOURCE_CREATION_TIME   = "source_creation_time"
	NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME     = "source_update_time"
	NOTIFICATION_COLUMN_EVENT_TIME_START       = "event_time_start"
	NOTIFICATION_COLUMN_EVENT_TIME_END         = "event_time_end"
	NOTIFICATION_COLUMN_SOURCE                 = "source"
	NOTIFICATION_COLUMN_SOURCE_ID              = "source_id"
	NOTIFICATION_COLUMN_TYPE                   = "type"
	NOTIFICATION_COLUMN_CATEGORY               = "category"
	NOTIFICATION_COLUMN_SHORT_DESCRIPTION      = "short_description"
	NOTIFICATION_COLUMN_INCIDENT_ID            = "incident_id"
	NOTIFICATION_COLUMN_RESOURCE_DISPLAY_NAMES = "resource_display_name"
	NOTIFICATION_COLUMN_CRN_FULL               = "crn_full"
	NOTIFICATION_COLUMN_VERSION                = "version"
	NOTIFICATION_COLUMN_CNAME                  = "cname"
	NOTIFICATION_COLUMN_CTYPE                  = "ctype"
	NOTIFICATION_COLUMN_SERVICE_NAME           = "service_name"
	NOTIFICATION_COLUMN_LOCATION               = "location"
	NOTIFICATION_COLUMN_SCOPE                  = "scope"
	NOTIFICATION_COLUMN_SERVICE_INSTANCE       = "service_instance"
	NOTIFICATION_COLUMN_RESOURCE_TYPE          = "resource_type"
	NOTIFICATION_COLUMN_RESOURCE               = "resource"
	NOTIFICATION_COLUMN_TAGS                   = "tags"
	NOTIFICATION_COLUMN_RECORD_RETRACTION_TIME = "record_retraction_time"
	NOTIFICATION_COLUMN_PNP_REMOVED            = "pnp_removed"
	NOTIFICATION_COLUMN_RELEASE_NOTE_URL       = "release_note_url"

	// Notification Description table columns names metadata

	NOTIFICATIONDESCRIPTION_COLUMN_RECORD_ID        = "record_id"
	NOTIFICATIONDESCRIPTION_COLUMN_NOTIFICATION_ID  = "notification_id"
	NOTIFICATIONDESCRIPTION_COLUMN_LONG_DESCRIPTION = "long_description"
	NOTIFICATIONDESCRIPTION_COLUMN_LANGUAGE         = "language"

	NOTIFICATION_LANGUAGE_LENGTH = 20

	//----------------------------------------------------------
	// Query columns in GetResourceByQuery function

	RESOURCE_QUERY_CRN                     = "crn"
	RESOURCE_QUERY_VERSION                 = "version"
	RESOURCE_QUERY_CNAME                   = "cname"
	RESOURCE_QUERY_CTYPE                   = "ctype"
	RESOURCE_QUERY_SERVICE_NAME            = "service_name"
	RESOURCE_QUERY_LOCATION                = "location"
	RESOURCE_QUERY_SCOPE                   = "scope"
	RESOURCE_QUERY_SERVICE_INSTANCE        = "service_instance"
	RESOURCE_QUERY_RESOURCE_TYPE           = "resource_type"
	RESOURCE_QUERY_RESOURCE                = "resource"
	RESOURCE_QUERY_VISIBILITY              = "visibility"
	RESOURCE_QUERY_CATALOG_PARENT_ID       = "catalog_parent_resource_id"
	RESOURCE_QUERY_CREATION_TIME_START     = "creation_time_start"
	RESOURCE_QUERY_CREATION_TIME_END       = "creation_time_end"
	RESOURCE_QUERY_PNP_CREATION_TIME_START = "pnp_creation_time_start"
	RESOURCE_QUERY_PNP_CREATION_TIME_END   = "pnp_creation_time_end"
	RESOURCE_QUERY_UPDATE_TIME_START       = "update_time_start"
	RESOURCE_QUERY_UPDATE_TIME_END         = "update_time_end"
	RESOURCE_QUERY_PNP_UPDATE_TIME_START   = "pnp_update_time_start"
	RESOURCE_QUERY_PNP_UPDATE_TIME_END     = "pnp_update_time_end"

	// Query columns in GetIncidentByQuery function

	INCIDENT_QUERY_CREATION_TIME_START     = "creation_time_start"
	INCIDENT_QUERY_CREATION_TIME_END       = "creation_time_end"
	INCIDENT_QUERY_PNP_CREATION_TIME_START = "pnp_creation_time_start"
	INCIDENT_QUERY_PNP_CREATION_TIME_END   = "pnp_creation_time_end"
	INCIDENT_QUERY_UPDATE_TIME_START       = "update_time_start"
	INCIDENT_QUERY_UPDATE_TIME_END         = "update_time_end"
	INCIDENT_QUERY_PNP_UPDATE_TIME_START   = "pnp_update_time_start"
	INCIDENT_QUERY_PNP_UPDATE_TIME_END     = "pnp_update_time_end"
	INCIDENT_QUERY_OUTAGE_START_START      = "outage_start_start"
	INCIDENT_QUERY_OUTAGE_START_END        = "outage_start_end"

	// Query columns in GetMaintenanceByQuery function

	MAINTENANCE_QUERY_SOURCE                  = "source"
	MAINTENANCE_QUERY_CREATION_TIME_START     = "creation_time_start"
	MAINTENANCE_QUERY_CREATION_TIME_END       = "creation_time_end"
	MAINTENANCE_QUERY_PNP_CREATION_TIME_START = "pnp_creation_time_start"
	MAINTENANCE_QUERY_PNP_CREATION_TIME_END   = "pnp_creation_time_end"
	MAINTENANCE_QUERY_UPDATE_TIME_START       = "update_time_start"
	MAINTENANCE_QUERY_UPDATE_TIME_END         = "update_time_end"
	MAINTENANCE_QUERY_PNP_UPDATE_TIME_START   = "pnp_update_time_start"
	MAINTENANCE_QUERY_PNP_UPDATE_TIME_END     = "pnp_update_time_end"
	MAINTENANCE_QUERY_PLANNED_START_START     = "planned_start_start"
	MAINTENANCE_QUERY_PLANNED_START_END       = "planned_start_end"
	MAINTENANCE_QUERY_PLANNED_END_START       = "planned_end_start"
	MAINTENANCE_QUERY_PLANNED_END_END         = "planned_end_end"

	// Query columns in GetWatchesByQuery function

	WATCH_QUERY_SUBSCRIPTION_ID = "subscription_id"
	WATCH_QUERY_KIND            = "kind"

	// Query columns in GetNotificationByQuery function

	NOTIFICATION_QUERY_CRN                     = "crn"
	NOTIFICATION_QUERY_VERSION                 = "version"
	NOTIFICATION_QUERY_CNAME                   = "cname"
	NOTIFICATION_QUERY_CTYPE                   = "ctype"
	NOTIFICATION_QUERY_SERVICE_NAME            = "service_name"
	NOTIFICATION_QUERY_LOCATION                = "location"
	NOTIFICATION_QUERY_SCOPE                   = "scope"
	NOTIFICATION_QUERY_SERVICE_INSTANCE        = "service_instance"
	NOTIFICATION_QUERY_RESOURCE_TYPE           = "resource_type"
	NOTIFICATION_QUERY_RESOURCE                = "resource"
	NOTIFICATION_QUERY_CREATION_TIME_START     = "creation_time_start"
	NOTIFICATION_QUERY_CREATION_TIME_END       = "creation_time_end"
	NOTIFICATION_QUERY_UPDATE_TIME_START       = "update_time_start"
	NOTIFICATION_QUERY_UPDATE_TIME_END         = "update_time_end"
	NOTIFICATION_QUERY_PNP_CREATION_TIME_START = "pnp_creation_time_start"
	NOTIFICATION_QUERY_PNP_CREATION_TIME_END   = "pnp_creation_time_end"
	NOTIFICATION_QUERY_PNP_UPDATE_TIME_START   = "pnp_update_time_start"
	NOTIFICATION_QUERY_PNP_UPDATE_TIME_END     = "pnp_update_time_end"

	// Errors declaration

	ERR_BAD_CRN_FORMAT     = "CRNFull has incorrect format. "
	ERR_NO_CRN_VERSION     = "CRN Version is empty. "
	ERR_NO_CNAME           = "CRN Cname is empty. "
	ERR_NO_CTYPE           = "CRN Ctype is empty. "
	ERR_NO_SERVICE         = "CRN ServiceName is empty. "
	ERR_NO_LOCATION        = "CRN Location is empty. "
	ERR_NO_CRN             = "IncidentInsert.CRNFull cannot be nil or empty"
	ERR_BAD_CLASSIFICATION = "Classification is not valid"
	ERR_BAD_STATE          = "State is not valid"
	ERR_NO_SOURCEID        = "SourceId is not valid"
	ERR_NO_SOURCE          = "Source is not valid"

	// SourcePNPDB use as resource.source when inserting generic CRN such as "crn:v1:::" + serviceName + ":::::"
	// https://github.ibm.com/cloud-sre/toolsplatform/issues/8257
	SourcePNPDB = "pnpdb"
	// SNnill2PnP ServiceNow sends nil when the value is set to -- None -- PnP use "none" instead
	SNnill2PnP = "none"
)
