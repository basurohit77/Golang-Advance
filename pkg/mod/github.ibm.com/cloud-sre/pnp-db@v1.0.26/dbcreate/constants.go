package dbcreate

import (
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
)

// use all lower case in column names since they are case-insensitive inside the database
const (
	// Database configuration file
	DATABASE_CONFIG_FILE = ".dbconfig.json"

	// Create table statement prefix
	CREATE_TABLE_IF_NOT_EXISTS = "CREATE TABLE IF NOT EXISTS "

	// Resource table indexes
	INDEX_RESOURCE_CNAME             = "resource_cname_index"
	INDEX_RESOURCE_CTYPE             = "resource_ctype_index"
	INDEX_RESOURCE_SERVICE_NAME      = "resource_service_name_index"
	INDEX_RESOURCE_LOCATION          = "resource_location_index"
	INDEX_RESOURCE_SCOPE             = "resource_scope_index"
	INDEX_RESOURCE_SERVICE_INSTANCE  = "resource_service_instance_index"
	INDEX_RESOURCE_RESOURCE_TYPE     = "resource_resource_type_index"
	INDEX_RESOURCE_RESOURCE          = "resource_resource_index"
	INDEX_RESOURCE_SOURCE_SOURCE_ID  = "resource_source_source_id_index"
	INDEX_RESOURCE_CATALOG_PARENT_ID = "resource_catalog_parent_id_index"

	// Display Names table indexes
	INDEX_DISPLAYNAMES_RESOURCE_ID = "displaynames_resource_id_index"

	// Visibility table indexes
	INDEX_VISIBILITY_NAME = "visibility_name_index"

	// Visibility Junction table indexes
	INDEX_VISIBILITYJUNCTION_RESOURCE_ID   = "visibilityjunction_resource_id_index"
	INDEX_VISIBILITYJUNCTION_VISIBILITY_ID = "visibilityjunction_visibility_id_index"

	// Tag Junction table indexes
	INDEX_TAGJUNCTION_RESOURCE_ID = "tagjunction_resource_id_index"
	INDEX_TAGJUNCTION_TAG_ID      = "tagjunction_tag_id_index" // tag's record_id index

	// Tag table indexes
	INDEX_TAG_ID = "tag_id_index"

	// Incident table indexes
	INDEX_INCIDENT_SOURCE_CREATION_TIME = "incident_source_creation_time_index"
	INDEX_INCIDENT_SOURCE_UPDATE_TIME   = "incident_source_update_time_index"
	INDEX_INCIDENT_START_TIME           = "incident_start_time_index"
	INDEX_INCIDENT_END_TIME             = "incident_end_time_index"
	INDEX_INCIDENT_SOURCE_SOURCE_ID     = "incident_source_source_id_index"
	// ATR Nov,2019 Added
	INDEX_INCIDENT_TARGETED_URL = "incident_targeted_url_index"
	INDEX_INCIDENT_AUDIENCE     = "incident_audience_index"

	// Incident Junction table indexes
	INDEX_INCIDENTJUNCTION_RESOURCE_ID = "incidentjunction_resource_id_index"
	INDEX_INCIDENTJUNCTION_INCIDENT_ID = "incidentjunction_incident_id_index"

	// Maintenance table indexes
	INDEX_MAINTENANCE_SOURCE_CREATION_TIME = "maintenance_source_creation_time_index"
	INDEX_MAINTENANCE_SOURCE_UPDATE_TIME   = "maintenance_source_update_time_index"
	INDEX_MAINTENANCE_START_TIME           = "maintenance_start_time_index"
	INDEX_MAINTENANCE_END_TIME             = "maintenance_end_time_index"
	INDEX_MAINTENANCE_SOURCE_SOURCE_ID     = "maintenance_source_source_id_index"
	// ATR Nov,2019 Added
	INDEX_MAINTENANCE_TARGETED_URL = "maintenance_targeted_url_index"
	INDEX_MAINTENANCE_AUDIENCE     = "maintenance_audience_index"

	// Maintenance Junction table indexes
	INDEX_MAINTENANCEJUNCTION_RESOURCE_ID    = "maintenancejunction_resource_id_index"
	INDEX_MAINTENANCEJUNCTION_MAINTENANCE_ID = "maintenancejunction_maintenance_id_index"

	// Case table indexes
	INDEX_CASE_SOURCE_SOURCE_ID = "case_source_source_id_index"

	// Subscription table indexes
	INDEX_SUBSCRIPTION_NAME = "subscription_name_index"

	// Watch table indexes
	INDEX_WATCH_SUBSCRIPTION_ID = "watch_subscription_id_index"
	INDEX_WATCH_KIND            = "watch_kind_index"

	// Watch Junction table indexes
	INDEX_WATCHJUNCTION_RESOURCE_ID = "watchjunction_resource_id_index"
	INDEX_WATCHJUNCTION_WATCH_ID    = "watchjunction_watch_id_index"

	// Notification table indexes
	INDEX_NOTIFICATION_CNAME                = "notification_cname_index"
	INDEX_NOTIFICATION_CTYPE                = "notification_ctype_index"
	INDEX_NOTIFICATION_SERVICE_NAME         = "notification_service_name_index"
	INDEX_NOTIFICATION_LOCATION             = "notification_location_index"
	INDEX_NOTIFICATION_SCOPE                = "notification_scope_index"
	INDEX_NOTIFICATION_SERVICE_INSTANCE     = "notification_service_instance_index"
	INDEX_NOTIFICATION_RESOURCE_TYPE        = "notification_resource_type_index"
	INDEX_NOTIFICATION_RESOURCE             = "notification_resource_index"
	INDEX_NOTIFICATION_SOURCE_SOURCE_ID     = "notification_source_source_id_index"
	INDEX_NOTIFICATION_SOURCE_CREATION_TIME = "notification_source_creation_time_index"
	INDEX_NOTIFICATION_SOURCE_UPDATE_TIME   = "notification_source_update_time_index"
	INDEX_NOTIFICATION_EVENT_TIME_START     = "notification_event_time_start_index"
	INDEX_NOTIFICATION_EVENT_TIME_END       = "notification_event_time_end_index"
	INDEX_NOTIFICATION_TYPE                 = "notification_type_index"
	// ATR Mar,2022 Added
	INDEX_NOTIFICATION_RELASE_NOTE_URL		= "notification_release_note_url_index"

	// Notification Description table indexes
	INDEX_NOTIFICATIONDESCRIPTION_NOTIFICATION_ID = "notificationdescription_notification_id_index"

	// Drop table Statements
	DROP_TABLE_IF_EXISTS                 = "DROP TABLE IF EXISTS "
	DROP_RESOURCE_TABLE_STMT             = DROP_TABLE_IF_EXISTS + db.RESOURCE_TABLE_NAME
	DROP_INCIDENT_TABLE_STMT             = DROP_TABLE_IF_EXISTS + db.INCIDENT_TABLE_NAME
	DROP_INCIDENT_JUNCTION_TABLE_STMT    = DROP_TABLE_IF_EXISTS + db.INCIDENT_JUNCTION_TABLE_NAME
	DROP_MAINTENANCE_TABLE_STMT          = DROP_TABLE_IF_EXISTS + db.MAINTENANCE_TABLE_NAME
	DROP_MAINTENANCE_JUNCTION_TABLE_STMT = DROP_TABLE_IF_EXISTS + db.MAINTENANCE_JUNCTION_TABLE_NAME
	//	DROP_MEMBER_TABLE_STMT = DROP_TABLE_IF_EXISTS + db.MEMBER_TABLE_NAME
	DROP_DISPLAY_NAMES_TABLE_STMT            = DROP_TABLE_IF_EXISTS + db.DISPLAY_NAMES_TABLE_NAME
	DROP_VISIBILITY_TABLE_STMT               = DROP_TABLE_IF_EXISTS + db.VISIBILITY_TABLE_NAME
	DROP_VISIBILITY_JUNCTION_TABLE_STMT      = DROP_TABLE_IF_EXISTS + db.VISIBILITY_JUNCTION_TABLE_NAME
	DROP_TAG_TABLE_STMT                      = DROP_TABLE_IF_EXISTS + db.TAG_TABLE_NAME
	DROP_TAG_JUNCTION_TABLE_STMT             = DROP_TABLE_IF_EXISTS + db.TAG_JUNCTION_TABLE_NAME
	DROP_CASE_TABLE_STMT                     = DROP_TABLE_IF_EXISTS + db.CASE_TABLE_NAME
	DROP_SUBSCRIPTION_TABLE_STMT             = DROP_TABLE_IF_EXISTS + db.SUBSCRIPTION_TABLE_NAME
	DROP_WATCH_TABLE_STMT                    = DROP_TABLE_IF_EXISTS + db.WATCH_TABLE_NAME
	DROP_WATCH_JUNCTION_TABLE_STMT           = DROP_TABLE_IF_EXISTS + db.WATCH_JUNCTION_TABLE_NAME
	DROP_NOTIFICATION_TABLE_STMT             = DROP_TABLE_IF_EXISTS + db.NOTIFICATION_TABLE_NAME
	DROP_NOTIFICATION_DESCRIPTION_TABLE_STMT = DROP_TABLE_IF_EXISTS + db.NOTIFICATION_DESCRIPTION_TABLE_NAME

	// Resource User-defined type names
	//UDT_RESOURCE_STATE = "resource_state_type"
	//UDT_RESOURCE_OPERATIONAL_STATUS = "resource_operational_status_type"
	//UDT_RESOURCE_STATUS = "resource_status_type"

	// Incident User-defined type names
	UDT_INCIDENT_STATE          = "incident_state_type"
	UDT_INCIDENT_CLASSIFICATION = "incident_classification_type"
	UDT_INCIDENT_SEVERITY       = "incident_severity_type"

	// Maintenance User-defined type names
	UDT_MAINTENANCE_STATE = "maintenance_state_type"

	// Watch User-defined type names
	UDT_WATCH_WILDCARDS = "watch_wildcards_type"

	// Drop user-defined types
	DROP_UDT_IF_EXISTS = "DROP TYPE IF EXISTS "
	//	DROP_UDT_RESOURCE_STATE_STMT = DROP_UDT_IF_EXISTS + UDT_RESOURCE_STATE
	//	DROP_UDT_RESOURCE_OPERATIONAL_STATUS_STMT = DROP_UDT_IF_EXISTS + UDT_RESOURCE_OPERATIONAL_STATUS
	//	DROP_UDT_RESOURCE_STATUS_STMT = DROP_UDT_IF_EXISTS + UDT_RESOURCE_STATUS

	DROP_UDT_INCIDENT_STATE_STMT          = DROP_UDT_IF_EXISTS + UDT_INCIDENT_STATE
	DROP_UDT_INCIDENT_CLASSIFICATION_STMT = DROP_UDT_IF_EXISTS + UDT_INCIDENT_CLASSIFICATION
	DROP_UDT_INCIDENT_SEVERITY_STMT       = DROP_UDT_IF_EXISTS + UDT_INCIDENT_SEVERITY

	//	DROP_UDT_MAINTENANCE_DISRUPTIVE_STMT = DROP_UDT_IF_EXISTS + UDT_MAINTENANCE_DISRUPTIVE
	DROP_UDT_MAINTENANCE_STATE_STMT = DROP_UDT_IF_EXISTS + UDT_MAINTENANCE_STATE

	DROP_UDT_WATCH_WILDCARDS_STMT = DROP_UDT_IF_EXISTS + UDT_WATCH_WILDCARDS

	// Create User-defined types
	CREATE_UDT = "CREATE TYPE "
	//	CREATE_UDT_RESOURCE_MODE_STMT = CREATE_UDT + UDT_RESOURCE_MODE + " AS ENUM ('resource', 'category')"
	//	CREATE_UDT_RESOURCE_STATE_STMT = CREATE_UDT + UDT_RESOURCE_STATE + " AS ENUM ('ok', 'archived')"
	//	CREATE_UDT_RESOURCE_OPERATIONAL_STATUS_STMT = CREATE_UDT + UDT_RESOURCE_OPERATIONAL_STATUS + " AS ENUM ('none', 'ga', 'experiment', 'deprecated')"
	//	CREATE_UDT_RESOURCE_STATUS_STMT = CREATE_UDT + UDT_RESOURCE_STATUS + " AS ENUM ('ok', 'degraded', 'failed')"

	CREATE_UDT_INCIDENT_STATE_STMT          = CREATE_UDT + UDT_INCIDENT_STATE + " AS ENUM ('new', 'in-progress', 'resolved')"
	CREATE_UDT_INCIDENT_CLASSIFICATION_STMT = CREATE_UDT + UDT_INCIDENT_CLASSIFICATION + " AS ENUM ('confirmed-cie', 'potential-cie', 'normal')"
	CREATE_UDT_INCIDENT_SEVERITY_STMT       = CREATE_UDT + UDT_INCIDENT_SEVERITY + " AS ENUM ('1', '2', '3', '4')"

	//	CREATE_UDT_MAINTENANCE_DISRUPTIVE_STMT = CREATE_UDT + UDT_MAINTENANCE_DISRUPTIVE + " AS ENUM ('true', 'false')"
	CREATE_UDT_MAINTENANCE_STATE_STMT = CREATE_UDT + UDT_MAINTENANCE_STATE + " AS ENUM ('new', 'scheduled', 'in-progress', 'complete')"

	CREATE_UDT_WATCH_WILDCARDS_STMT = CREATE_UDT + UDT_WATCH_WILDCARDS + " AS ENUM ('true', 'false')"

	// Create/Drop resource index statements
	CREATE_INDEX        = "CREATE INDEX IF NOT EXISTS "
	CREATE_UNIQUE_INDEX = "CREATE UNIQUE INDEX IF NOT EXISTS "
	DROP_INDEX          = "DROP INDEX IF EXISTS "

	CREATE_INDEX_RESOURCE_CNAME_STMT = CREATE_INDEX + INDEX_RESOURCE_CNAME + " ON " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_CNAME + ")"
	DROP_INDEX_RESOURCE_CNAME_STMT   = DROP_INDEX + INDEX_RESOURCE_CNAME

	CREATE_INDEX_RESOURCE_CTYPE_STMT = CREATE_INDEX + INDEX_RESOURCE_CTYPE + " ON " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_CTYPE + ")"
	DROP_INDEX_RESOURCE_CTYPE_STMT   = DROP_INDEX + INDEX_RESOURCE_CTYPE

	CREATE_INDEX_RESOURCE_SERVICE_NAME_STMT = CREATE_INDEX + INDEX_RESOURCE_SERVICE_NAME + " ON " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_SERVICE_NAME + ")"
	DROP_INDEX_RESOURCE_SERVICE_NAME_STMT   = DROP_INDEX + INDEX_RESOURCE_SERVICE_NAME

	CREATE_INDEX_RESOURCE_LOCATION_STMT = CREATE_INDEX + INDEX_RESOURCE_LOCATION + " ON " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_LOCATION + ")"
	DROP_INDEX_RESOURCE_LOCATION_STMT   = DROP_INDEX + INDEX_RESOURCE_LOCATION

	CREATE_INDEX_RESOURCE_SCOPE_STMT = CREATE_INDEX + INDEX_RESOURCE_SCOPE + " ON " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_SCOPE + ")"
	DROP_INDEX_RESOURCE_SCOPE_STMT   = DROP_INDEX + INDEX_RESOURCE_SCOPE

	CREATE_INDEX_RESOURCE_SERVICE_INSTANCE_STMT = CREATE_INDEX + INDEX_RESOURCE_SERVICE_INSTANCE + " ON " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_SERVICE_INSTANCE + ")"
	DROP_INDEX_RESOURCE_SERVICE_INSTANCE_STMT   = DROP_INDEX + INDEX_RESOURCE_SERVICE_INSTANCE

	CREATE_INDEX_RESOURCE_RESOURCE_TYPE_STMT = CREATE_INDEX + INDEX_RESOURCE_RESOURCE_TYPE + " ON " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_RESOURCE_TYPE + ")"
	DROP_INDEX_RESOURCE_RESOURCE_TYPE_STMT   = DROP_INDEX + INDEX_RESOURCE_RESOURCE_TYPE

	CREATE_INDEX_RESOURCE_RESOURCE_STMT = CREATE_INDEX + INDEX_RESOURCE_RESOURCE + " ON " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_RESOURCE + ")"
	DROP_INDEX_RESOURCE_RESOURCE_STMT   = DROP_INDEX + INDEX_RESOURCE_RESOURCE

	CREATE_INDEX_RESOURCE_SOURCE_SOURCE_ID_STMT = CREATE_INDEX + INDEX_RESOURCE_SOURCE_SOURCE_ID + " ON " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_SOURCE + "," + db.RESOURCE_COLUMN_SOURCE_ID + ")"
	DROP_INDEX_RESOURCE_SOURCE_SOURCE_ID_STMT   = DROP_INDEX + INDEX_RESOURCE_SOURCE_SOURCE_ID

	CREATE_INDEX_RESOURCE_CATALOG_PARENT_ID_STMT = CREATE_INDEX + INDEX_RESOURCE_CATALOG_PARENT_ID + " ON " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_CATALOG_PARENT_ID + ")"
	DROP_INDEX_RESOURCE_CATALOG_PARENT_ID_STMT   = DROP_INDEX + INDEX_RESOURCE_CATALOG_PARENT_ID

	// Create/Drop incident index statements
	CREATE_INDEX_INCIDENT_SOURCE_CREATION_TIME_STMT = CREATE_INDEX + INDEX_INCIDENT_SOURCE_CREATION_TIME + " ON " + db.INCIDENT_TABLE_NAME + "(" + db.INCIDENT_COLUMN_SOURCE_CREATION_TIME + ")"
	DROP_INDEX_INCIDENT_SOURCE_CREATION_TIME_STMT   = DROP_INDEX + INDEX_INCIDENT_SOURCE_CREATION_TIME

	CREATE_INDEX_INCIDENT_SOURCE_UPDATE_TIME_STMT = CREATE_INDEX + INDEX_INCIDENT_SOURCE_UPDATE_TIME + " ON " + db.INCIDENT_TABLE_NAME + "(" + db.INCIDENT_COLUMN_SOURCE_UPDATE_TIME + ")"
	DROP_INDEX_INCIDENT_SOURCE_UPDATE_TIME_STMT   = DROP_INDEX + INDEX_INCIDENT_SOURCE_UPDATE_TIME

	CREATE_INDEX_INCIDENT_START_TIME_STMT = CREATE_INDEX + INDEX_INCIDENT_START_TIME + " ON " + db.INCIDENT_TABLE_NAME + "(" + db.INCIDENT_COLUMN_START_TIME + ")"
	DROP_INDEX_INCIDENT_START_TIME_STMT   = DROP_INDEX + INDEX_INCIDENT_START_TIME

	CREATE_INDEX_INCIDENT_END_TIME_STMT = CREATE_INDEX + INDEX_INCIDENT_END_TIME + " ON " + db.INCIDENT_TABLE_NAME + "(" + db.INCIDENT_COLUMN_END_TIME + ")"
	DROP_INDEX_INCIDENT_END_TIME_STMT   = DROP_INDEX + INDEX_INCIDENT_END_TIME

	CREATE_INDEX_INCIDENT_SOURCE_SOURCE_ID_STMT = CREATE_INDEX + INDEX_INCIDENT_SOURCE_SOURCE_ID + " ON " + db.INCIDENT_TABLE_NAME + "(" + db.INCIDENT_COLUMN_SOURCE + "," + db.INCIDENT_COLUMN_SOURCE_ID + ")"
	DROP_INDEX_INCIDENT_SOURCE_SOURCE_ID_STMT   = DROP_INDEX + INDEX_INCIDENT_SOURCE_SOURCE_ID
	// ATR Nov,2019 Added
	CREATE_INDEX_INCIDENT_TARGETED_URL_STMT = CREATE_INDEX + INDEX_INCIDENT_TARGETED_URL + " ON " + db.INCIDENT_TABLE_NAME + "(" + db.INCIDENT_COLUMN_SOURCE + "," + db.INCIDENT_COLUMN_TARGETED_URL + ")"
	DROP_INDEX_INCIDENT_TARGETED_URL_STMT   = DROP_INDEX + INDEX_INCIDENT_TARGETED_URL

	CREATE_INDEX_INCIDENT_AUDIENCE_STMT = CREATE_INDEX + INDEX_INCIDENT_AUDIENCE + " ON " + db.INCIDENT_TABLE_NAME + "(" + db.INCIDENT_COLUMN_AUDIENCE + ")"
	DROP_INDEX_INCIDENT_AUDIENCE_STMT   = DROP_INDEX + INDEX_INCIDENT_AUDIENCE

	CREATE_INDEX_INCIDENTJUNCTION_RESOURCE_ID_STMT = CREATE_INDEX + INDEX_INCIDENTJUNCTION_RESOURCE_ID + " ON " + db.INCIDENT_JUNCTION_TABLE_NAME + "(" + db.INCIDENTJUNCTION_COLUMN_RESOURCE_ID + ")"
	DROP_INDEX_INCIDENTJUNCTION_RESOURCE_ID_STMT   = DROP_INDEX + INDEX_INCIDENTJUNCTION_RESOURCE_ID

	CREATE_INDEX_INCIDENTJUNCTION_INCIDENT_ID_STMT = CREATE_INDEX + INDEX_INCIDENTJUNCTION_INCIDENT_ID + " ON " + db.INCIDENT_JUNCTION_TABLE_NAME + "(" + db.INCIDENTJUNCTION_COLUMN_INCIDENT_ID + ")"
	DROP_INDEX_INCIDENTJUNCTION_INCIDENT_ID_STMT   = DROP_INDEX + INDEX_INCIDENTJUNCTION_INCIDENT_ID

	// Create/Drop maintenance index statements
	CREATE_INDEX_MAINTENANCE_SOURCE_CREATION_TIME_STMT = CREATE_INDEX + INDEX_MAINTENANCE_SOURCE_CREATION_TIME + " ON " + db.MAINTENANCE_TABLE_NAME + "(" + db.MAINTENANCE_COLUMN_SOURCE_CREATION_TIME + ")"
	DROP_INDEX_MAINTENANCE_SOURCE_CREATION_TIME_STMT   = DROP_INDEX + INDEX_MAINTENANCE_SOURCE_CREATION_TIME

	CREATE_INDEX_MAINTENANCE_SOURCE_UPDATE_TIME_STMT = CREATE_INDEX + INDEX_MAINTENANCE_SOURCE_UPDATE_TIME + " ON " + db.MAINTENANCE_TABLE_NAME + "(" + db.MAINTENANCE_COLUMN_SOURCE_UPDATE_TIME + ")"
	DROP_INDEX_MAINTENANCE_SOURCE_UPDATE_TIME_STMT   = DROP_INDEX + INDEX_MAINTENANCE_SOURCE_UPDATE_TIME

	CREATE_INDEX_MAINTENANCE_START_TIME_STMT = CREATE_INDEX + INDEX_MAINTENANCE_START_TIME + " ON " + db.MAINTENANCE_TABLE_NAME + "(" + db.MAINTENANCE_COLUMN_START_TIME + ")"
	DROP_INDEX_MAINTENANCE_START_TIME_STMT   = DROP_INDEX + INDEX_MAINTENANCE_START_TIME

	CREATE_INDEX_MAINTENANCE_END_TIME_STMT = CREATE_INDEX + INDEX_MAINTENANCE_END_TIME + " ON " + db.MAINTENANCE_TABLE_NAME + "(" + db.MAINTENANCE_COLUMN_END_TIME + ")"
	DROP_INDEX_MAINTENANCE_END_TIME_STMT   = DROP_INDEX + INDEX_MAINTENANCE_END_TIME

	CREATE_INDEX_MAINTENANCE_SOURCE_SOURCE_ID_STMT = CREATE_INDEX + INDEX_MAINTENANCE_SOURCE_SOURCE_ID + " ON " + db.MAINTENANCE_TABLE_NAME + "(" + db.MAINTENANCE_COLUMN_SOURCE + "," + db.MAINTENANCE_COLUMN_SOURCE_ID + ")"
	DROP_INDEX_MAINTENANCE_SOURCE_SOURCE_ID_STMT   = DROP_INDEX + INDEX_MAINTENANCE_SOURCE_SOURCE_ID
	// ATR Nov,2019 Added
	CREATE_INDEX_MAINTENANCE_TARGETED_URL_STMT = CREATE_INDEX + INDEX_MAINTENANCE_TARGETED_URL + " ON " + db.MAINTENANCE_TABLE_NAME + "(" + db.MAINTENANCE_COLUMN_SOURCE + "," + db.MAINTENANCE_COLUMN_TARGETED_URL + ")"
	DROP_INDEX_MAINTENANCE_TARGETED_URL_STMT   = DROP_INDEX + INDEX_MAINTENANCE_TARGETED_URL

	CREATE_INDEX_MAINTENANCE_AUDIENCE_STMT = CREATE_INDEX + INDEX_MAINTENANCE_AUDIENCE + " ON " + db.MAINTENANCE_TABLE_NAME + "(" + db.MAINTENANCE_COLUMN_AUDIENCE + ")"
	DROP_INDEX_MAINTENANCE_AUDIENCE_STMT   = DROP_INDEX + INDEX_MAINTENANCE_AUDIENCE

	CREATE_INDEX_MAINTENANCEJUNCTION_RESOURCE_ID_STMT = CREATE_INDEX + INDEX_MAINTENANCEJUNCTION_RESOURCE_ID + " ON " + db.MAINTENANCE_JUNCTION_TABLE_NAME + "(" + db.MAINTENANCEJUNCTION_COLUMN_RESOURCE_ID + ")"
	DROP_INDEX_MAINTENANCEJUNCTION_RESOURCE_ID_STMT   = DROP_INDEX + INDEX_MAINTENANCEJUNCTION_RESOURCE_ID

	CREATE_INDEX_MAINTENANCEJUNCTION_MAINTENANCE_ID_STMT = CREATE_INDEX + INDEX_MAINTENANCEJUNCTION_MAINTENANCE_ID + " ON " + db.MAINTENANCE_JUNCTION_TABLE_NAME + "(" + db.MAINTENANCEJUNCTION_COLUMN_MAINTENANCE_ID + ")"
	DROP_INDEX_MAINTENANCEJUNCTION_MAINTENANCE_ID_STMT   = DROP_INDEX + INDEX_MAINTENANCEJUNCTION_MAINTENANCE_ID

	//	CREATE_INDEX_MEMBER_PARENT_STMT = CREATE_INDEX + INDEX_MEMBER_PARENT + " ON " + db.MEMBER_TABLE_NAME + "(" + db.MEMBER_COLUMN_PARENT + ")"
	//	DROP_INDEX_MEMBER_PARENT_STMT = DROP_INDEX + INDEX_MEMBER_PARENT

	//	CREATE_INDEX_MEMBER_MEMBER_STMT = CREATE_INDEX + INDEX_MEMBER_MEMBER + " ON " + db.MEMBER_TABLE_NAME + "(" + db.MEMBER_COLUMN_MEMBER + ")"
	//	DROP_INDEX_MEMBER_MEMBER_STMT = DROP_INDEX + INDEX_MEMBER_MEMBER

	CREATE_INDEX_DISPLAYNAMES_RESOURCE_ID_STMT = CREATE_INDEX + INDEX_DISPLAYNAMES_RESOURCE_ID + " ON " + db.DISPLAY_NAMES_TABLE_NAME + "(" + db.DISPLAYNAMES_COLUMN_RESOURCE_ID + ")"
	DROP_INDEX_DISPLAYNAMES_RESOURCE_ID_STMT   = DROP_INDEX + INDEX_DISPLAYNAMES_RESOURCE_ID

	CREATE_INDEX_VISIBILITYJUNCTION_RESOURCE_ID_STMT = CREATE_INDEX + INDEX_VISIBILITYJUNCTION_RESOURCE_ID + " ON " + db.VISIBILITY_JUNCTION_TABLE_NAME + "(" + db.VISIBILITYJUNCTION_COLUMN_RESOURCE_ID + ")"
	DROP_INDEX_VISIBILITYJUNCTION_RESOURCE_ID_STMT   = DROP_INDEX + INDEX_VISIBILITYJUNCTION_RESOURCE_ID

	CREATE_INDEX_VISIBILITYJUNCTION_VISIBILITY_ID_STMT = CREATE_INDEX + INDEX_VISIBILITYJUNCTION_VISIBILITY_ID + " ON " + db.VISIBILITY_JUNCTION_TABLE_NAME + "(" + db.VISIBILITYJUNCTION_COLUMN_VISIBILITY_ID + ")"
	DROP_INDEX_VISIBILITYJUNCTION_VISIBILITY_ID_STMT   = DROP_INDEX + INDEX_VISIBILITYJUNCTION_VISIBILITY_ID

	CREATE_INDEX_VISIBILITY_NAME_STMT = CREATE_UNIQUE_INDEX + INDEX_VISIBILITY_NAME + " ON " + db.VISIBILITY_TABLE_NAME + "(" + db.VISIBILITY_COLUMN_NAME + ")"
	DROP_INDEX_VISIBILITY_NAME_STMT   = DROP_INDEX + INDEX_VISIBILITY_NAME

	CREATE_INDEX_TAGJUNCTION_RESOURCE_ID_STMT = CREATE_INDEX + INDEX_TAGJUNCTION_RESOURCE_ID + " ON " + db.TAG_JUNCTION_TABLE_NAME + "(" + db.TAGJUNCTION_COLUMN_RESOURCE_ID + ")"
	DROP_INDEX_TAGJUNCTION_RESOURCE_ID_STMT   = DROP_INDEX + INDEX_TAGJUNCTION_RESOURCE_ID

	CREATE_INDEX_TAGJUNCTION_TAG_ID_STMT = CREATE_INDEX + INDEX_TAGJUNCTION_TAG_ID + " ON " + db.TAG_JUNCTION_TABLE_NAME + "(" + db.TAGJUNCTION_COLUMN_TAG_ID + ")"
	DROP_INDEX_TAGJUNCTION_TAG_ID_STMT   = DROP_INDEX + INDEX_TAGJUNCTION_TAG_ID

	CREATE_INDEX_TAG_ID_STMT = CREATE_UNIQUE_INDEX + INDEX_TAG_ID + " ON " + db.TAG_TABLE_NAME + "(" + db.TAG_COLUMN_ID + ")"
	DROP_INDEX_TAG_ID_STMT   = DROP_INDEX + INDEX_TAG_ID

	CREATE_INDEX_SUBSCRIPTION_NAME_STMT = CREATE_INDEX + INDEX_SUBSCRIPTION_NAME + " ON " + db.SUBSCRIPTION_TABLE_NAME + "(" + db.SUBSCRIPTION_COLUMN_NAME + ")"
	DROP_INDEX_SUBSCRIPTION_NAME_STMT   = DROP_INDEX + INDEX_SUBSCRIPTION_NAME

	CREATE_INDEX_WATCH_SUBSCRIPTION_ID_STMT = CREATE_INDEX + INDEX_WATCH_SUBSCRIPTION_ID + " ON " + db.WATCH_TABLE_NAME + "(" + db.WATCH_COLUMN_SUBSCRIPTION_ID + ")"
	DROP_INDEX_WATCH_SUBSCRIPTION_ID_STMT   = DROP_INDEX + INDEX_WATCH_SUBSCRIPTION_ID

	//	CREATE_INDEX_WATCH_KIND_STMT = CREATE_INDEX + INDEX_WATCH_KIND + " ON " + db.WATCH_TABLE_NAME + "(" + db.WATCH_COLUMN_KIND + ")"
	//	DROP_INDEX_WATCH_KIND_STMT = DROP_INDEX + INDEX_WATCH_KIND

	CREATE_INDEX_WATCHJUNCTION_RESOURCE_ID_STMT = CREATE_INDEX + INDEX_WATCHJUNCTION_RESOURCE_ID + " ON " + db.WATCH_JUNCTION_TABLE_NAME + "(" + db.WATCHJUNCTION_COLUMN_RESOURCE_ID + ")"
	DROP_INDEX_WATCHJUNCTION_RESOURCE_ID_STMT   = DROP_INDEX + INDEX_WATCHJUNCTION_RESOURCE_ID

	CREATE_INDEX_WATCHJUNCTION_WATCH_ID_STMT = CREATE_INDEX + INDEX_WATCHJUNCTION_WATCH_ID + " ON " + db.WATCH_JUNCTION_TABLE_NAME + "(" + db.WATCHJUNCTION_COLUMN_WATCH_ID + ")"
	DROP_INDEX_WATCHJUNCTION_WATCH_ID_STMT   = DROP_INDEX + INDEX_WATCHJUNCTION_WATCH_ID

	CREATE_INDEX_CASE_SOURCE_SOURCE_ID_STMT = CREATE_INDEX + INDEX_CASE_SOURCE_SOURCE_ID + " ON " + db.CASE_TABLE_NAME + "(" + db.CASE_COLUMN_SOURCE + "," + db.CASE_COLUMN_SOURCE_ID + ")"
	DROP_INDEX_CASE_SOURCE_SOURCE_ID_STMT   = DROP_INDEX + INDEX_CASE_SOURCE_SOURCE_ID

	CREATE_INDEX_NOTIFICATION_CNAME_STMT = CREATE_INDEX + INDEX_NOTIFICATION_CNAME + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_CNAME + ")"
	DROP_INDEX_NOTIFICATION_CNAME_STMT   = DROP_INDEX + INDEX_NOTIFICATION_CNAME

	CREATE_INDEX_NOTIFICATION_CTYPE_STMT = CREATE_INDEX + INDEX_NOTIFICATION_CTYPE + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_CTYPE + ")"
	DROP_INDEX_NOTIFICATION_CTYPE_STMT   = DROP_INDEX + INDEX_NOTIFICATION_CTYPE

	CREATE_INDEX_NOTIFICATION_SERVICE_NAME_STMT = CREATE_INDEX + INDEX_NOTIFICATION_SERVICE_NAME + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_SERVICE_NAME + ")"
	DROP_INDEX_NOTIFICATION_SERVICE_NAME_STMT   = DROP_INDEX + INDEX_NOTIFICATION_SERVICE_NAME

	CREATE_INDEX_NOTIFICATION_LOCATION_STMT = CREATE_INDEX + INDEX_NOTIFICATION_LOCATION + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_LOCATION + ")"
	DROP_INDEX_NOTIFICATION_LOCATION_STMT   = DROP_INDEX + INDEX_NOTIFICATION_LOCATION

	CREATE_INDEX_NOTIFICATION_SCOPE_STMT = CREATE_INDEX + INDEX_NOTIFICATION_SCOPE + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_SCOPE + ")"
	DROP_INDEX_NOTIFICATION_SCOPE_STMT   = DROP_INDEX + INDEX_NOTIFICATION_SCOPE

	CREATE_INDEX_NOTIFICATION_SERVICE_INSTANCE_STMT = CREATE_INDEX + INDEX_NOTIFICATION_SERVICE_INSTANCE + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_SERVICE_INSTANCE + ")"
	DROP_INDEX_NOTIFICATION_SERVICE_INSTANCE_STMT   = DROP_INDEX + INDEX_NOTIFICATION_SERVICE_INSTANCE

	CREATE_INDEX_NOTIFICATION_RESOURCE_TYPE_STMT = CREATE_INDEX + INDEX_NOTIFICATION_RESOURCE_TYPE + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_RESOURCE_TYPE + ")"
	DROP_INDEX_NOTIFICATION_RESOURCE_TYPE_STMT   = DROP_INDEX + INDEX_NOTIFICATION_RESOURCE_TYPE

	CREATE_INDEX_NOTIFICATION_RESOURCE_STMT = CREATE_INDEX + INDEX_NOTIFICATION_RESOURCE + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_RESOURCE + ")"
	DROP_INDEX_NOTIFICATION_RESOURCE_STMT   = DROP_INDEX + INDEX_NOTIFICATION_RESOURCE

	CREATE_INDEX_NOTIFICATION_SOURCE_SOURCE_ID_STMT = CREATE_INDEX + INDEX_NOTIFICATION_SOURCE_SOURCE_ID + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_SOURCE + "," + db.NOTIFICATION_COLUMN_SOURCE_ID + ")"
	DROP_INDEX_NOTIFICATION_SOURCE_SOURCE_ID_STMT   = DROP_INDEX + INDEX_NOTIFICATION_SOURCE_SOURCE_ID

	CREATE_INDEX_NOTIFICATION_SOURCE_CREATION_TIME_STMT = CREATE_INDEX + INDEX_NOTIFICATION_SOURCE_CREATION_TIME + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_SOURCE_CREATION_TIME + ")"
	DROP_INDEX_NOTIFICATION_SOURCE_CREATION_TIME_STMT   = DROP_INDEX + INDEX_NOTIFICATION_SOURCE_CREATION_TIME

	CREATE_INDEX_NOTIFICATION_SOURCE_UPDATE_TIME_STMT = CREATE_INDEX + INDEX_NOTIFICATION_SOURCE_UPDATE_TIME + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME + ")"
	DROP_INDEX_NOTIFICATION_SOURCE_UPDATE_TIME_STMT   = DROP_INDEX + INDEX_NOTIFICATION_SOURCE_UPDATE_TIME

	CREATE_INDEX_NOTIFICATION_EVENT_TIME_START_STMT = CREATE_INDEX + INDEX_NOTIFICATION_EVENT_TIME_START + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_EVENT_TIME_START + ")"
	DROP_INDEX_NOTIFICATION_EVENT_TIME_START_STMT   = DROP_INDEX + INDEX_NOTIFICATION_EVENT_TIME_START

	CREATE_INDEX_NOTIFICATION_EVENT_TIME_END_STMT = CREATE_INDEX + INDEX_NOTIFICATION_EVENT_TIME_END + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_EVENT_TIME_END + ")"
	DROP_INDEX_NOTIFICATION_EVENT_TIME_END_STMT   = DROP_INDEX + INDEX_NOTIFICATION_EVENT_TIME_END

	CREATE_INDEX_NOTIFICATION_TYPE_STMT = CREATE_INDEX + INDEX_NOTIFICATION_TYPE + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_TYPE + ")"
	DROP_INDEX_NOTIFICATION_TYPE_STMT   = DROP_INDEX + INDEX_NOTIFICATION_TYPE

	// ATR Mar,2022 Added
	CREATE_INDEX_NOTIFICATION_RELASE_NOTE_URL_STMT = CREATE_INDEX + INDEX_NOTIFICATION_RELASE_NOTE_URL + " ON " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_RELEASE_NOTE_URL + ")"
	DROP_INDEX_NOTIFICATION_RELASE_NOTE_URL_STMT   = DROP_INDEX + INDEX_NOTIFICATION_RELASE_NOTE_URL

	CREATE_INDEX_NOTIFICATIONDESCRIPTION_NOTIFICATION_ID_STMT = CREATE_INDEX + INDEX_NOTIFICATIONDESCRIPTION_NOTIFICATION_ID + " ON " + db.NOTIFICATION_DESCRIPTION_TABLE_NAME + "(" + db.NOTIFICATIONDESCRIPTION_COLUMN_NOTIFICATION_ID + ")"
	DROP_INDEX_NOTIFICATIONDESCRIPTION_NOTIFICATION_ID_STMT   = DROP_INDEX + INDEX_NOTIFICATIONDESCRIPTION_NOTIFICATION_ID

)
