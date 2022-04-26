package dbcreate

import (
	"database/sql"
	"log"
	"time"

	"github.ibm.com/cloud-sre/osstf/lg"
	db "github.ibm.com/cloud-sre/pnp-abstraction/db"
)

var urlLength = "2000"

// *********************************** Drop UDTs ***********************************

// DropUDTIncidentState user defined type already exist, drop it and redefine it
func DropUDTIncidentState(database *sql.DB, recreate bool) {
	if recreate {
		// user defined type already exist, drop it and redefine it
		err := ExecuteQuery(database, DROP_UDT_INCIDENT_STATE_STMT)
		if err != nil {
			log.Fatal("Cannot drop UDT: " + UDT_INCIDENT_STATE)
		}
	}
}

// DropUDTIncidentClassification user defined type already exist, drop it and redefine it
func DropUDTIncidentClassification(database *sql.DB, recreate bool) {
	if recreate {
		// user defined type already exist, drop it and redefine it
		err := ExecuteQuery(database, DROP_UDT_INCIDENT_CLASSIFICATION_STMT)
		if err != nil {
			log.Fatal("Cannot drop UDT: " + UDT_INCIDENT_CLASSIFICATION)
		}
	}
}

// DropUDTIncidentSeverity user defined type already exist, drop it and redefine it
func DropUDTIncidentSeverity(database *sql.DB, recreate bool) {
	if recreate {
		// user defined type already exist, drop it and redefine it
		err := ExecuteQuery(database, DROP_UDT_INCIDENT_SEVERITY_STMT)
		if err != nil {
			log.Fatal("Cannot drop UDT: " + UDT_INCIDENT_SEVERITY)
		}
	}
}

// DropUDTMaintenanceState user defined type already exist, drop it and redefine it
func DropUDTMaintenanceState(database *sql.DB, recreate bool) {
	if recreate {
		// user defined type already exist, drop it and redefine it
		err := ExecuteQuery(database, DROP_UDT_MAINTENANCE_STATE_STMT)
		if err != nil {
			log.Fatal("Cannot drop UDT: " + UDT_MAINTENANCE_STATE)
		}
	}
}

//DropUDTWatchWildcards user defined type already exist, drop it and redefine it
func DropUDTWatchWildcards(database *sql.DB, recreate bool) {
	if recreate {
		// user defined type already exist, drop it and redefine it
		err := ExecuteQuery(database, DROP_UDT_WATCH_WILDCARDS_STMT)
		if err != nil {
			log.Fatal("Cannot drop UDT: " + UDT_WATCH_WILDCARDS)
		}
	}
}

// *********************************** Create UDTs ***********************************

// CreateUDTIncidentState creates user defined type for Incident State
func CreateUDTIncidentState(database *sql.DB) {
	err := ExecuteQuery(database, CREATE_UDT_INCIDENT_STATE_STMT)
	if err != nil {
		log.Fatal("Cannot create UDT: " + UDT_INCIDENT_STATE)
	}
}

// CreateUDTIncidentClassification creates user defined type for Incident Classification
func CreateUDTIncidentClassification(database *sql.DB) {
	err := ExecuteQuery(database, CREATE_UDT_INCIDENT_CLASSIFICATION_STMT)
	if err != nil {
		log.Fatal("Cannot create UDT: " + UDT_INCIDENT_CLASSIFICATION)
	}
}

// CreateUDTIncidentSeverity creates user defined type for Incident Severity
func CreateUDTIncidentSeverity(database *sql.DB) {
	err := ExecuteQuery(database, CREATE_UDT_INCIDENT_SEVERITY_STMT)
	if err != nil {
		log.Fatal("Cannot create UDT: " + UDT_INCIDENT_SEVERITY)
	}
}

// CreateUDTMaintenanceState creates user defined type for Maintenance State
func CreateUDTMaintenanceState(database *sql.DB) {
	err := ExecuteQuery(database, CREATE_UDT_MAINTENANCE_STATE_STMT)
	if err != nil {
		log.Fatal("Cannot create UDT: " + UDT_MAINTENANCE_STATE)
	}
}

// CreateUDTWatchWildcards creates user defined type for Watch Wildcards
func CreateUDTWatchWildcards(database *sql.DB) {
	err := ExecuteQuery(database, CREATE_UDT_WATCH_WILDCARDS_STMT)
	if err != nil {
		log.Fatal("Cannot create UDT: " + UDT_WATCH_WILDCARDS)
	}
}

// *********************************** Drop Tables **********************************

// DropResourceTable drops resource table and their dependencies
func DropResourceTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_RESOURCE_CNAME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_RESOURCE_CNAME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_RESOURCE_CTYPE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_RESOURCE_CTYPE_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_RESOURCE_SERVICE_NAME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_RESOURCE_SERVICE_NAME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_RESOURCE_LOCATION_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_RESOURCE_LOCATION_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_RESOURCE_SCOPE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_RESOURCE_SCOPE_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_RESOURCE_SERVICE_INSTANCE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_RESOURCE_SERVICE_INSTANCE_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_RESOURCE_RESOURCE_TYPE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_RESOURCE_RESOURCE_TYPE_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_RESOURCE_RESOURCE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_RESOURCE_RESOURCE_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_RESOURCE_SOURCE_SOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_RESOURCE_SOURCE_SOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_RESOURCE_CATALOG_PARENT_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_RESOURCE_CATALOG_PARENT_ID_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_RESOURCE_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.RESOURCE_TABLE_NAME)
	}

}

// DropNotificationTable drops notification table and their dependencies
func DropNotificationTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_NOTIFICATION_CNAME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_CNAME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_CTYPE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_CTYPE_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_SERVICE_NAME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_SERVICE_NAME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_LOCATION_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_LOCATION_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_SCOPE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_SCOPE_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_SERVICE_INSTANCE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_SERVICE_INSTANCE_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_RESOURCE_TYPE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_RESOURCE_TYPE_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_RESOURCE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_RESOURCE_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_SOURCE_SOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_SOURCE_SOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_SOURCE_CREATION_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_SOURCE_CREATION_TIME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_SOURCE_UPDATE_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_SOURCE_UPDATE_TIME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_EVENT_TIME_START_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_EVENT_TIME_START_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_EVENT_TIME_END_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_EVENT_TIME_END_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_TYPE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_TYPE_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_NOTIFICATION_RELASE_NOTE_URL_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATION_RELASE_NOTE_URL_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_NOTIFICATION_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.NOTIFICATION_TABLE_NAME)
	}
}

// DropNotificationDescriptionTable drops notification_description table and their dependencies
func DropNotificationDescriptionTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_NOTIFICATIONDESCRIPTION_NOTIFICATION_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_NOTIFICATIONDESCRIPTION_NOTIFICATION_ID_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_NOTIFICATION_DESCRIPTION_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.NOTIFICATION_DESCRIPTION_TABLE_NAME)
	}
}

// DropIncidentTable drops incident  table and their dependencies
func DropIncidentTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_INCIDENT_SOURCE_CREATION_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_INCIDENT_SOURCE_CREATION_TIME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_INCIDENT_SOURCE_UPDATE_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_INCIDENT_SOURCE_UPDATE_TIME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_INCIDENT_START_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_INCIDENT_START_TIME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_INCIDENT_END_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_INCIDENT_END_TIME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_INCIDENT_SOURCE_SOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_INCIDENT_SOURCE_SOURCE_ID_STMT)
	}

	// ATR Nov,2019 Added
	err = ExecuteQuery(database, DROP_INDEX_INCIDENT_TARGETED_URL_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_INCIDENT_TARGETED_URL_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_INCIDENT_AUDIENCE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_INCIDENT_AUDIENCE_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_INCIDENT_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.INCIDENT_TABLE_NAME)
	}

	time.Sleep(10 * time.Second) // ensure postgres has finished dropping the tables

	// drop user defined types already exist
	DropUDTIncidentState(database, true)
	DropUDTIncidentClassification(database, true)
	DropUDTIncidentSeverity(database, true)
}

// DropIncidentJunctionTable drops incident_junction  table and their dependencies
func DropIncidentJunctionTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_INCIDENTJUNCTION_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_INCIDENTJUNCTION_RESOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_INCIDENTJUNCTION_INCIDENT_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_INCIDENTJUNCTION_INCIDENT_ID_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_INCIDENT_JUNCTION_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.INCIDENT_JUNCTION_TABLE_NAME)
	}
}

// DropMaintenanceTable drops incident_junction  table and their dependencies
func DropMaintenanceTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_MAINTENANCE_SOURCE_CREATION_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_MAINTENANCE_SOURCE_CREATION_TIME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_MAINTENANCE_SOURCE_UPDATE_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_MAINTENANCE_SOURCE_UPDATE_TIME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_MAINTENANCE_START_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_MAINTENANCE_START_TIME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_MAINTENANCE_END_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_MAINTENANCE_END_TIME_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_MAINTENANCE_SOURCE_SOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_MAINTENANCE_SOURCE_SOURCE_ID_STMT)
	}

	// ATR Nov,2019 Added
	err = ExecuteQuery(database, DROP_INDEX_MAINTENANCE_TARGETED_URL_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_MAINTENANCE_TARGETED_URL_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_MAINTENANCE_AUDIENCE_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_MAINTENANCE_AUDIENCE_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_MAINTENANCE_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.MAINTENANCE_TABLE_NAME)
	}

	time.Sleep(10 * time.Second) // ensure postgres has finished dropping the tables

	// drop user defined types already exist
	DropUDTMaintenanceState(database, true)
}

// DropMaintenanceJunctionTable drops maintenance_junction  table and their dependencies
func DropMaintenanceJunctionTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_MAINTENANCEJUNCTION_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_MAINTENANCEJUNCTION_RESOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_MAINTENANCEJUNCTION_MAINTENANCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_MAINTENANCEJUNCTION_MAINTENANCE_ID_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_MAINTENANCE_JUNCTION_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.MAINTENANCE_JUNCTION_TABLE_NAME)
	}
}

// DropDisplayNamesTable drops display_names  table and their dependencies
func DropDisplayNamesTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_DISPLAYNAMES_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_DISPLAYNAMES_RESOURCE_ID_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_DISPLAY_NAMES_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.DISPLAY_NAMES_TABLE_NAME)
	}
}

// DropVisibilityTable drops visibility  table and their dependencies
func DropVisibilityTable(database *sql.DB) {
	// drop the tables already exist
	err := ExecuteQuery(database, DROP_VISIBILITY_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.VISIBILITY_TABLE_NAME)
	}
}

// DropVisibilityJunctionTable drops visibility_juntion  table and their dependencies
func DropVisibilityJunctionTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_VISIBILITYJUNCTION_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_VISIBILITYJUNCTION_RESOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_VISIBILITYJUNCTION_VISIBILITY_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_VISIBILITYJUNCTION_VISIBILITY_ID_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_VISIBILITY_JUNCTION_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.VISIBILITY_JUNCTION_TABLE_NAME)
	}
}

// DropTagTable drops tab table and their dependencies
func DropTagTable(database *sql.DB) {
	// drop the tables already exist
	err := ExecuteQuery(database, DROP_TAG_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.TAG_TABLE_NAME)
	}
}

// DropTagJunctionTable drops tab_junction  table and their dependencies
func DropTagJunctionTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_TAGJUNCTION_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_TAGJUNCTION_RESOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_TAGJUNCTION_TAG_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_TAGJUNCTION_TAG_ID_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_TAG_JUNCTION_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.TAG_JUNCTION_TABLE_NAME)
	}
}

// DropCaseTable drops case table and their dependencies
func DropCaseTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_CASE_SOURCE_SOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_CASE_SOURCE_SOURCE_ID_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_CASE_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.CASE_TABLE_NAME)
	}
}

// DropSubscriptionTable drops suscription table and their dependencies
func DropSubscriptionTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_SUBSCRIPTION_NAME_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_SUBSCRIPTION_NAME_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_SUBSCRIPTION_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.SUBSCRIPTION_TABLE_NAME)
	}
}

// DropWatchTable drops watch table and their dependencies
func DropWatchTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_WATCH_SUBSCRIPTION_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_WATCH_SUBSCRIPTION_ID_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_WATCH_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.WATCH_TABLE_NAME)
	}

	time.Sleep(10 * time.Second) // ensure postgres has finished dropping the tables

	// drop user defined types already exist
	DropUDTWatchWildcards(database, true)

}

// DropWatchJunctionTable drops watch_junction table and their dependencies
func DropWatchJunctionTable(database *sql.DB) {
	// drop indexes already exist
	err := ExecuteQuery(database, DROP_INDEX_WATCHJUNCTION_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_WATCHJUNCTION_RESOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, DROP_INDEX_WATCHJUNCTION_WATCH_ID_STMT)
	if err != nil {
		log.Fatal("Cannot drop index: " + DROP_INDEX_WATCHJUNCTION_WATCH_ID_STMT)
	}

	// drop the tables already exist
	err = ExecuteQuery(database, DROP_WATCH_JUNCTION_TABLE_STMT)
	if err != nil {
		log.Fatal("Cannot drop table " + db.WATCH_JUNCTION_TABLE_NAME)
	}
}

// *********************************** Create Tables *********************************

// CreateResourceTable creates resource table and their dependencies
func CreateResourceTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateResourceTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.RESOURCE_TABLE_NAME)
	}

	err = ExecuteQuery(database, CREATE_INDEX_RESOURCE_CNAME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_RESOURCE_CNAME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_RESOURCE_CTYPE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_RESOURCE_CTYPE_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_RESOURCE_SERVICE_NAME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_RESOURCE_SERVICE_NAME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_RESOURCE_LOCATION_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_RESOURCE_LOCATION_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_RESOURCE_SCOPE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_RESOURCE_SCOPE_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_RESOURCE_SERVICE_INSTANCE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_RESOURCE_SERVICE_INSTANCE_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_RESOURCE_RESOURCE_TYPE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_RESOURCE_RESOURCE_TYPE_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_RESOURCE_RESOURCE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_RESOURCE_RESOURCE_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_RESOURCE_SOURCE_SOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_RESOURCE_SOURCE_SOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_RESOURCE_CATALOG_PARENT_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_RESOURCE_CATALOG_PARENT_ID_STMT)
	}
}

// CreateNotificationTable creates notification table and their dependencies
func CreateNotificationTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateNotificationTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.NOTIFICATION_TABLE_NAME)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_CNAME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_CNAME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_CTYPE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_CTYPE_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_SERVICE_NAME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_SERVICE_NAME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_LOCATION_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_LOCATION_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_SCOPE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_SCOPE_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_SERVICE_INSTANCE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_SERVICE_INSTANCE_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_RESOURCE_TYPE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_RESOURCE_TYPE_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_RESOURCE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_RESOURCE_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_SOURCE_SOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_SOURCE_SOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_SOURCE_CREATION_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_SOURCE_CREATION_TIME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_SOURCE_UPDATE_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_SOURCE_UPDATE_TIME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_EVENT_TIME_START_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_EVENT_TIME_START_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_EVENT_TIME_END_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_EVENT_TIME_END_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_TYPE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_TYPE_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATION_RELASE_NOTE_URL_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATION_RELASE_NOTE_URL_STMT)
	}

}

// CreateNotificationDescriptionTable creates notification_description table and their dependencies
func CreateNotificationDescriptionTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateNotificationDescriptionTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.NOTIFICATION_DESCRIPTION_TABLE_NAME)
	}

	err = ExecuteQuery(database, CREATE_INDEX_NOTIFICATIONDESCRIPTION_NOTIFICATION_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_NOTIFICATIONDESCRIPTION_NOTIFICATION_ID_STMT)
	}
}

// CreateIncidentTable creates incident table and their dependencies
func CreateIncidentTable(database *sql.DB) {
	// create UDT
	CreateUDTIncidentState(database)
	CreateUDTIncidentClassification(database)
	CreateUDTIncidentSeverity(database)

	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateIncidentTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.INCIDENT_TABLE_NAME)
	}

	// create indexes only if it is not exist
	err = ExecuteQuery(database, CREATE_INDEX_INCIDENT_SOURCE_CREATION_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_INCIDENT_SOURCE_CREATION_TIME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_INCIDENT_SOURCE_UPDATE_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_INCIDENT_SOURCE_UPDATE_TIME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_INCIDENT_START_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_INCIDENT_START_TIME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_INCIDENT_END_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_INCIDENT_END_TIME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_INCIDENT_SOURCE_SOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_INCIDENT_SOURCE_SOURCE_ID_STMT)
	}
	// ATR Nov,2019 Added
	err = ExecuteQuery(database, CREATE_INDEX_INCIDENT_TARGETED_URL_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_INCIDENT_TARGETED_URL_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_INCIDENT_AUDIENCE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_INCIDENT_AUDIENCE_STMT)
	}
}

// CreateIncidentJunctionTable creates incident_junction table and their dependencies
func CreateIncidentJunctionTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateIncidentJunctionTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.INCIDENT_JUNCTION_TABLE_NAME)
	}

	// create indexes only if it is not exist
	err = ExecuteQuery(database, CREATE_INDEX_INCIDENTJUNCTION_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_INCIDENTJUNCTION_RESOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_INCIDENTJUNCTION_INCIDENT_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_INCIDENTJUNCTION_INCIDENT_ID_STMT)
	}
}

// CreateMaintenanceTable creates maintenance table and their dependencies
func CreateMaintenanceTable(database *sql.DB) {
	// Create UDT
	CreateUDTMaintenanceState(database)

	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateMaintenanceTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.MAINTENANCE_TABLE_NAME)
	}

	// create indexes only if it is not exist
	err = ExecuteQuery(database, CREATE_INDEX_MAINTENANCE_SOURCE_CREATION_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_MAINTENANCE_SOURCE_CREATION_TIME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_MAINTENANCE_SOURCE_UPDATE_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_MAINTENANCE_SOURCE_UPDATE_TIME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_MAINTENANCE_START_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_MAINTENANCE_START_TIME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_MAINTENANCE_END_TIME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_MAINTENANCE_END_TIME_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_MAINTENANCE_SOURCE_SOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_MAINTENANCE_SOURCE_SOURCE_ID_STMT)
	}
	// ATR Nov,2019 Added
	err = ExecuteQuery(database, CREATE_INDEX_MAINTENANCE_TARGETED_URL_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_MAINTENANCE_TARGETED_URL_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_MAINTENANCE_AUDIENCE_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_MAINTENANCE_AUDIENCE_STMT)
	}
}

// CreateMaintenanceJunctionTable creates maintenance_junction table and their dependencies
func CreateMaintenanceJunctionTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateMaintenanceJunctionTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.MAINTENANCE_JUNCTION_TABLE_NAME)
	}

	// create indexes only if it is not exist
	err = ExecuteQuery(database, CREATE_INDEX_MAINTENANCEJUNCTION_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_MAINTENANCEJUNCTION_RESOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_MAINTENANCEJUNCTION_MAINTENANCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_MAINTENANCEJUNCTION_MAINTENANCE_ID_STMT)
	}
}

// CreateDisplayNamesTable creates display_names table and their dependencies
func CreateDisplayNamesTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateDisplayNamesTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.DISPLAY_NAMES_TABLE_NAME)
	}

	err = ExecuteQuery(database, CREATE_INDEX_DISPLAYNAMES_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_DISPLAYNAMES_RESOURCE_ID_STMT)
	}
}

// CreateVisibilityTable creates visibility table and their dependencies
func CreateVisibilityTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateVisibilityTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.VISIBILITY_TABLE_NAME)
	}
}

// CreateVisibilityJunctionTable creates visibility_junction table and their dependencies
func CreateVisibilityJunctionTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateVisibilityJunctionTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.VISIBILITY_JUNCTION_TABLE_NAME)
	}

	// create indexes only if it is not exist
	err = ExecuteQuery(database, CREATE_INDEX_VISIBILITYJUNCTION_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_VISIBILITYJUNCTION_RESOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_VISIBILITYJUNCTION_VISIBILITY_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_VISIBILITYJUNCTION_VISIBILITY_ID_STMT)
	}
}

// CreateTagTable creates tag table and their dependencies
func CreateTagTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateTagTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.TAG_TABLE_NAME)
	}
}

// CreateTagJunctionTable creates tag_junction table and their dependencies
func CreateTagJunctionTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateTagJunctionTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.TAG_JUNCTION_TABLE_NAME)
	}

	// create indexes only if it is not exist
	err = ExecuteQuery(database, CREATE_INDEX_TAGJUNCTION_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_TAGJUNCTION_RESOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_TAGJUNCTION_TAG_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_TAGJUNCTION_TAG_ID_STMT)
	}
}

// CreateCaseTable creates case table and their dependencies
func CreateCaseTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateCaseTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.CASE_TABLE_NAME)
	}

	// create indexes only if it is not exist
	err = ExecuteQuery(database, CREATE_INDEX_CASE_SOURCE_SOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_CASE_SOURCE_SOURCE_ID_STMT)
	}
}

// CreateSubscriptionTable creates subsciption table and their dependencies
func CreateSubscriptionTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateSubscriptionTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.SUBSCRIPTION_TABLE_NAME)
	}

	// create indexes only if it is not exist
	err = ExecuteQuery(database, CREATE_INDEX_SUBSCRIPTION_NAME_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_SUBSCRIPTION_NAME_STMT)
	}
}

// CreateWatchJunctionTable creates watch_junction table and their dependencies
func CreateWatchJunctionTable(database *sql.DB) {
	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateWatchJunctionTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.WATCH_JUNCTION_TABLE_NAME)
	}

	// create indexes only if it is not exist
	err = ExecuteQuery(database, CREATE_INDEX_WATCHJUNCTION_RESOURCE_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_WATCHJUNCTION_RESOURCE_ID_STMT)
	}

	err = ExecuteQuery(database, CREATE_INDEX_WATCHJUNCTION_WATCH_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_WATCHJUNCTION_WATCH_ID_STMT)
	}
}

// CreateWatchTable creates watch table and their dependencies
func CreateWatchTable(database *sql.DB) {
	// Create UDT
	CreateUDTWatchWildcards(database)

	// create table only if it is not exist
	err := ExecuteQuery(database, GetCreateWatchTableStatement())
	if err != nil {
		log.Fatal("Cannot create table " + db.WATCH_TABLE_NAME)
	}

	// create indexes only if it is not exist
	err = ExecuteQuery(database, CREATE_INDEX_WATCH_SUBSCRIPTION_ID_STMT)
	if err != nil {
		log.Fatal("Cannot create index: " + CREATE_INDEX_WATCH_SUBSCRIPTION_ID_STMT)
	}

}

// *********************** Create Table Functions ***********************

//GetCreateResourceTableStatement resource table definition only add/update fields
func GetCreateResourceTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.RESOURCE_TABLE_NAME + " (" +
		db.RESOURCE_COLUMN_RECORD_ID + " varchar(64) PRIMARY KEY, " +
		db.RESOURCE_COLUMN_PNP_CREATION_TIME + " timestamptz NOT NULL, " +
		db.RESOURCE_COLUMN_PNP_UPDATE_TIME + " timestamptz NOT NULL, " +
		db.RESOURCE_COLUMN_SOURCE_CREATION_TIME + " timestamptz, " +
		db.RESOURCE_COLUMN_SOURCE_UPDATE_TIME + " timestamptz, " +
		db.RESOURCE_COLUMN_STATE + " " + " varchar(20), " +
		db.RESOURCE_COLUMN_OPERATIONAL_STATUS + " " + " varchar(20), " +
		db.RESOURCE_COLUMN_SOURCE + " varchar(32) NOT NULL, " +
		db.RESOURCE_COLUMN_SOURCE_ID + " varchar(128) NOT NULL, " +
		db.RESOURCE_COLUMN_STATUS + " " + " varchar(20), " +
		db.RESOURCE_COLUMN_STATUS_UPDATE_TIME + " timestamptz, " +
		db.RESOURCE_COLUMN_REGULATORY_DOMAIN + " varchar(32), " +
		db.RESOURCE_COLUMN_CATEGORY_ID + " varchar(128), " +
		db.RESOURCE_COLUMN_CATEGORY_PARENT + " BOOLEAN DEFAULT FALSE, " +
		db.RESOURCE_COLUMN_VERSION + " varchar(10) NOT NULL, " +
		db.RESOURCE_COLUMN_CNAME + " varchar(32) NOT NULL, " +
		db.RESOURCE_COLUMN_CTYPE + " varchar(32) NOT NULL, " +
		db.RESOURCE_COLUMN_SERVICE_NAME + " varchar(64) NOT NULL, " +
		db.RESOURCE_COLUMN_LOCATION + " varchar(32) NOT NULL, " +
		db.RESOURCE_COLUMN_SCOPE + " varchar(32) NOT NULL, " +
		db.RESOURCE_COLUMN_SERVICE_INSTANCE + " varchar(200) NOT NULL, " +
		db.RESOURCE_COLUMN_RESOURCE_TYPE + " varchar(32) NOT NULL, " +
		db.RESOURCE_COLUMN_RESOURCE + " varchar(1024) NOT NULL," +
		db.RESOURCE_COLUMN_IS_CATALOG_PARENT + " BOOLEAN DEFAULT FALSE, " +
		db.RESOURCE_COLUMN_CATALOG_PARENT_ID + " varchar(64), " +
		db.RESOURCE_COLUMN_CRN_FULL + " varchar(1500) NOT NULL," +
		db.RESOURCE_COLUMN_RECORD_HASH + " varchar(64)" +
		")"
}

//GetCreateNotificationTableStatement notification table definition only add/update fields
func GetCreateNotificationTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.NOTIFICATION_TABLE_NAME + " (" +
		db.NOTIFICATION_COLUMN_RECORD_ID + " varchar(64) PRIMARY KEY, " +
		db.NOTIFICATION_COLUMN_PNP_CREATION_TIME + " timestamptz NOT NULL, " +
		db.NOTIFICATION_COLUMN_PNP_UPDATE_TIME + " timestamptz NOT NULL, " +
		db.NOTIFICATION_COLUMN_SOURCE_CREATION_TIME + " timestamptz, " +
		db.NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME + " timestamptz, " +
		db.NOTIFICATION_COLUMN_EVENT_TIME_START + " timestamptz, " +
		db.NOTIFICATION_COLUMN_EVENT_TIME_END + " timestamptz, " +
		db.NOTIFICATION_COLUMN_SOURCE + " varchar(32) NOT NULL, " +
		db.NOTIFICATION_COLUMN_SOURCE_ID + " varchar(128) NOT NULL, " +
		db.NOTIFICATION_COLUMN_TYPE + " " + " varchar(20) NOT NULL, " +
		db.NOTIFICATION_COLUMN_CATEGORY + " varchar(20), " +
		db.NOTIFICATION_COLUMN_INCIDENT_ID + " varchar(64), " +
		db.NOTIFICATION_COLUMN_SHORT_DESCRIPTION + " varchar(350)[], " + // first 20 character is language, the rest is text
		db.NOTIFICATION_COLUMN_RESOURCE_DISPLAY_NAMES + " varchar(224)[], " + // first 20 character is language, the rest is text
		db.NOTIFICATION_COLUMN_VERSION + " varchar(10) NOT NULL, " +
		db.NOTIFICATION_COLUMN_CNAME + " varchar(32) NOT NULL, " +
		db.NOTIFICATION_COLUMN_CTYPE + " varchar(32) NOT NULL, " +
		db.NOTIFICATION_COLUMN_SERVICE_NAME + " varchar(64) NOT NULL, " +
		db.NOTIFICATION_COLUMN_LOCATION + " varchar(32) NOT NULL, " +
		db.NOTIFICATION_COLUMN_SCOPE + " varchar(32) NOT NULL, " +
		db.NOTIFICATION_COLUMN_SERVICE_INSTANCE + " varchar(200) NOT NULL, " +
		db.NOTIFICATION_COLUMN_RESOURCE_TYPE + " varchar(32) NOT NULL, " +
		db.NOTIFICATION_COLUMN_RESOURCE + " varchar(1024) NOT NULL," +
		db.NOTIFICATION_COLUMN_CRN_FULL + " varchar(1500) NOT NULL," +
		db.NOTIFICATION_COLUMN_TAGS + " varchar(200), " + // comma separated strings
		db.NOTIFICATION_COLUMN_RECORD_RETRACTION_TIME + " timestamptz, " +
		db.NOTIFICATION_COLUMN_PNP_REMOVED + " BOOLEAN DEFAULT FALSE," +
		db.NOTIFICATION_COLUMN_RELEASE_NOTE_URL + " text " +
		")"
}

// GetCreateNotificationDescriptionTableStatement notification_description table definition only add/update fields
func GetCreateNotificationDescriptionTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.NOTIFICATION_DESCRIPTION_TABLE_NAME + " (" +
		db.NOTIFICATIONDESCRIPTION_COLUMN_RECORD_ID + " varchar(64) PRIMARY KEY, " +
		db.NOTIFICATIONDESCRIPTION_COLUMN_LONG_DESCRIPTION + " text, " +
		db.NOTIFICATIONDESCRIPTION_COLUMN_LANGUAGE + " varchar(20), " +
		db.NOTIFICATIONDESCRIPTION_COLUMN_NOTIFICATION_ID + " varchar(64) references " + db.NOTIFICATION_TABLE_NAME + "(" + db.NOTIFICATION_COLUMN_RECORD_ID + ") ON DELETE CASCADE" +
		")"
}

// GetCreateIncidentTableStatement incident table definition only add/update fields
// ATR Nov,2019 Added db.INCIDENT_COLUMN_TARGETED_URL + " text "
func GetCreateIncidentTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.INCIDENT_TABLE_NAME + " (" +
		db.INCIDENT_COLUMN_RECORD_ID + " varchar(64) PRIMARY KEY, " +
		db.INCIDENT_COLUMN_PNP_CREATION_TIME + " timestamptz NOT NULL, " +
		db.INCIDENT_COLUMN_PNP_UPDATE_TIME + " timestamptz NOT NULL, " +
		db.INCIDENT_COLUMN_SOURCE_CREATION_TIME + " timestamptz NOT NULL, " +
		db.INCIDENT_COLUMN_SOURCE_UPDATE_TIME + " timestamptz, " +
		db.INCIDENT_COLUMN_START_TIME + " timestamptz, " +
		db.INCIDENT_COLUMN_END_TIME + " timestamptz, " +
		db.INCIDENT_COLUMN_SHORT_DESCRIPTION + " varchar(200), " +
		db.INCIDENT_COLUMN_LONG_DESCRIPTION + " text, " +
		db.INCIDENT_COLUMN_STATE + " " + UDT_INCIDENT_STATE + "," +
		db.INCIDENT_COLUMN_CLASSIFICATION + " " + UDT_INCIDENT_CLASSIFICATION + ", " +
		db.INCIDENT_COLUMN_SEVERITY + " " + UDT_INCIDENT_SEVERITY + "," +
		db.INCIDENT_COLUMN_SOURCE_ID + " varchar(64) NOT NULL, " +
		db.INCIDENT_COLUMN_SOURCE + " varchar(32) NOT NULL, " +
		db.INCIDENT_COLUMN_REGULATORY_DOMAIN + " varchar(32)," +
		db.INCIDENT_COLUMN_CRN_FULL + " varchar(1500)[] NOT NULL," +
		db.INCIDENT_COLUMN_AFFECTED_ACTIVITY + " varchar(40)," +
		db.INCIDENT_COLUMN_CUSTOMER_IMPACT_DESCRIPTION + " varchar(4000)," +
		db.INCIDENT_COLUMN_PNP_REMOVED + " BOOLEAN DEFAULT FALSE ," +
		db.INCIDENT_COLUMN_TARGETED_URL + " text ," +
		db.INCIDENT_COLUMN_AUDIENCE + " varchar(20) )"
}

// GetCreateIncidentJunctionTableStatement incident_junction table definition only add/update fields
func GetCreateIncidentJunctionTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.INCIDENT_JUNCTION_TABLE_NAME + " (" +
		db.INCIDENTJUNCTION_COLUMN_RECORD_ID + " uuid PRIMARY KEY, " +
		db.INCIDENTJUNCTION_COLUMN_RESOURCE_ID + " varchar(64) references " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_RECORD_ID + ") ON DELETE CASCADE, " +
		db.INCIDENTJUNCTION_COLUMN_INCIDENT_ID + " varchar(64) references " + db.INCIDENT_TABLE_NAME + "(" + db.INCIDENT_COLUMN_RECORD_ID + ") ON DELETE CASCADE" +
		")"
}

// GetCreateMaintenanceTableStatement maintenance table definition only add/update fields
// ATR Nov,2019 Added db.MAINTENANCE_COLUMN_TARGETED_URL + " text "
func GetCreateMaintenanceTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.MAINTENANCE_TABLE_NAME + " (" +
		db.MAINTENANCE_COLUMN_RECORD_ID + " varchar(64) PRIMARY KEY," +
		db.MAINTENANCE_COLUMN_PNP_CREATION_TIME + " timestamptz NOT NULL, " +
		db.MAINTENANCE_COLUMN_PNP_UPDATE_TIME + " timestamptz NOT NULL, " +
		db.MAINTENANCE_COLUMN_SOURCE_CREATION_TIME + " timestamptz NOT NULL, " +
		db.MAINTENANCE_COLUMN_SOURCE_UPDATE_TIME + " timestamptz, " +
		db.MAINTENANCE_COLUMN_START_TIME + " timestamptz, " +
		db.MAINTENANCE_COLUMN_END_TIME + " timestamptz, " +
		db.MAINTENANCE_COLUMN_SHORT_DESCRIPTION + " varchar(200), " +
		db.MAINTENANCE_COLUMN_LONG_DESCRIPTION + " text, " +
		db.MAINTENANCE_COLUMN_STATE + " " + UDT_MAINTENANCE_STATE + ", " +
		db.MAINTENANCE_COLUMN_DISRUPTIVE + " BOOLEAN DEFAULT FALSE, " +
		db.MAINTENANCE_COLUMN_SOURCE_ID + " varchar(64) NOT NULL, " +
		db.MAINTENANCE_COLUMN_SOURCE + " varchar(32) NOT NULL, " +
		db.MAINTENANCE_COLUMN_REGULATORY_DOMAIN + " varchar(32), " +
		db.MAINTENANCE_COLUMN_CRN_FULL + " varchar(1500)[] NOT NULL," +
		db.MAINTENANCE_COLUMN_RECORD_HASH + " varchar(64)," +
		db.MAINTENANCE_COLUMN_MAINTENANCE_DURATION + " integer DEFAULT 0," +
		db.MAINTENANCE_COLUMN_DISRUPTION_TYPE + " varchar(200)," +
		db.MAINTENANCE_COLUMN_DISRUPTION_DESCRIPTION + " text," +
		db.MAINTENANCE_COLUMN_DISRUPTION_DURATION + " integer DEFAULT 0," +
		db.MAINTENANCE_COLUMN_COMPLETION_CODE + " varchar(40)," +
		db.MAINTENANCE_COLUMN_PNP_REMOVED + " BOOLEAN DEFAULT FALSE," +
		db.MAINTENANCE_COLUMN_TARGETED_URL + " text ," +
		db.MAINTENANCE_COLUMN_AUDIENCE + " varchar(20) )"
}

// GetCreateMaintenanceJunctionTableStatement maintenance_junction table definition only add/update fields
func GetCreateMaintenanceJunctionTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.MAINTENANCE_JUNCTION_TABLE_NAME + " (" +
		db.MAINTENANCEJUNCTION_COLUMN_RECORD_ID + " uuid PRIMARY KEY, " +
		db.MAINTENANCEJUNCTION_COLUMN_RESOURCE_ID + " varchar(64) references " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_RECORD_ID + ") ON DELETE CASCADE," +
		db.MAINTENANCEJUNCTION_COLUMN_MAINTENANCE_ID + " varchar(64) references " + db.MAINTENANCE_TABLE_NAME + "(" + db.MAINTENANCE_COLUMN_RECORD_ID + ") ON DELETE CASCADE" +
		")"
}

// GetCreateDisplayNamesTableStatement display_names table definition only add/update fields
func GetCreateDisplayNamesTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.DISPLAY_NAMES_TABLE_NAME + " (" +
		db.DISPLAYNAMES_COLUMN_RECORD_ID + " uuid PRIMARY KEY, " +
		db.DISPLAYNAMES_COLUMN_NAME + " varchar(100), " +
		db.DISPLAYNAMES_COLUMN_LANGUAGE + " varchar(20), " +
		db.DISPLAYNAMES_COLUMN_RESOURCE_ID + " varchar(64) references " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_RECORD_ID + ") ON DELETE CASCADE" +
		")"
}

// GetCreateVisibilityTableStatement visibility table definition only add/update fields
func GetCreateVisibilityTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.VISIBILITY_TABLE_NAME + " (" +
		db.VISIBILITY_COLUMN_RECORD_ID + " varchar(64) PRIMARY KEY, " +
		db.VISIBILITY_COLUMN_NAME + " varchar(32) NOT NULL, " +
		db.VISIBILITY_COLUMN_DESCRIPTION + " varchar(100) " +
		")"
}

// GetCreateVisibilityJunctionTableStatement visibility_junction table definition only add/update fields
func GetCreateVisibilityJunctionTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.VISIBILITY_JUNCTION_TABLE_NAME + " (" +
		db.VISIBILITYJUNCTION_COLUMN_RECORD_ID + " uuid PRIMARY KEY, " +
		db.VISIBILITYJUNCTION_COLUMN_RESOURCE_ID + " varchar(64) references " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_RECORD_ID + ") ON DELETE CASCADE," +
		db.VISIBILITYJUNCTION_COLUMN_VISIBILITY_ID + " varchar(64) references " + db.VISIBILITY_TABLE_NAME + "(" + db.VISIBILITY_COLUMN_RECORD_ID + ") ON DELETE CASCADE" +
		")"
}

// GetCreateTagTableStatement tag table definition only add/update fields
func GetCreateTagTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.TAG_TABLE_NAME + " (" +
		db.TAG_COLUMN_RECORD_ID + " varchar(64) PRIMARY KEY, " +
		db.TAG_COLUMN_ID + " varchar(64) NOT NULL " +
		")"
}

// GetCreateTagJunctionTableStatement tag_junction table definition only add/update fields
func GetCreateTagJunctionTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.TAG_JUNCTION_TABLE_NAME + " (" +
		db.TAGJUNCTION_COLUMN_RECORD_ID + " uuid PRIMARY KEY, " +
		db.TAGJUNCTION_COLUMN_RESOURCE_ID + " varchar(64) references " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_RECORD_ID + ") ON DELETE CASCADE," +
		db.TAGJUNCTION_COLUMN_TAG_ID + " varchar(64) references " + db.TAG_TABLE_NAME + "(" + db.TAG_COLUMN_RECORD_ID + ") ON DELETE CASCADE" +
		")"
}

// GetCreateCaseTableStatement case table definition only add/update fields
func GetCreateCaseTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.CASE_TABLE_NAME + " (" +
		db.CASE_COLUMN_RECORD_ID + " varchar(64) PRIMARY KEY, " +
		db.CASE_COLUMN_SOURCE + " varchar(32) NOT NULL, " +
		db.CASE_COLUMN_SOURCE_ID + " varchar(64) NOT NULL, " +
		db.CASE_COLUMN_SOURCE_SYS_ID + " varchar(128) NOT NULL" +
		")"
}

// GetCreateSubscriptionTableStatement subscription table definition only add/update fields
func GetCreateSubscriptionTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.SUBSCRIPTION_TABLE_NAME + " (" +
		db.SUBSCRIPTION_COLUMN_RECORD_ID + " uuid PRIMARY KEY, " +
		db.SUBSCRIPTION_COLUMN_NAME + " varchar(64) NOT NULL, " +
		db.SUBSCRIPTION_COLUMN_TARGET_ADDRESS + " varchar(" + urlLength + "), " +
		db.SUBSCRIPTION_COLUMN_TARGET_TOKEN + " varchar(7500), " +
		db.SUBSCRIPTION_COLUMN_EXPIRATION + " timestamptz" +
		")"
}

// GetCreateWatchTableStatement watch table definition only add/update fields
func GetCreateWatchTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.WATCH_TABLE_NAME + " (" +
		db.WATCH_COLUMN_RECORD_ID + " uuid PRIMARY KEY, " +
		db.WATCH_COLUMN_SUBSCRIPTION_ID + " uuid references " + db.SUBSCRIPTION_TABLE_NAME + "(" + db.SUBSCRIPTION_COLUMN_RECORD_ID + ") ON DELETE CASCADE, " +
		db.WATCH_COLUMN_KIND + " varchar(20) NOT NULL, " +
		db.WATCH_COLUMN_PATH + " varchar(1000), " +
		db.WATCH_COLUMN_WILDCARDS + " " + UDT_WATCH_WILDCARDS + ", " +
		db.WATCH_COLUMN_RECORD_ID_TO_WATCH + " varchar(64)[], " +
		db.WATCH_COLUMN_CRN_FULL + " varchar(1500)[], " +
		db.WATCH_COLUMN_SUBSCRIPTION_EMAIL + " varchar(128)" +
		")"
}

// GetCreateWatchJunctionTableStatement watch_juction table definition only add/update fields
func GetCreateWatchJunctionTableStatement() string {
	return CREATE_TABLE_IF_NOT_EXISTS + db.WATCH_JUNCTION_TABLE_NAME + " (" +
		db.WATCHJUNCTION_COLUMN_RECORD_ID + " uuid PRIMARY KEY, " +
		db.WATCHJUNCTION_COLUMN_RESOURCE_ID + " varchar(64) references " + db.RESOURCE_TABLE_NAME + "(" + db.RESOURCE_COLUMN_RECORD_ID + ") ON DELETE CASCADE, " +
		db.WATCHJUNCTION_COLUMN_WATCH_ID + " uuid references " + db.WATCH_TABLE_NAME + "(" + db.WATCH_COLUMN_RECORD_ID + ") ON DELETE CASCADE" +
		")"
}

// *************************************************************************************

// ExecuteQuery Execute one statement, it does not abort when error occurs, return error
func ExecuteQuery(database *sql.DB, statement string) error {
	result, err := database.Exec(statement)
	lg.SqlExecuteStatement(statement, result, err)

	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println("Error: "+statement, ": ", err)
	}
	return err
}
