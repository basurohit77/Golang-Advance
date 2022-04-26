# pnp-db
Creates PnP Postgres database and all the tables, indexes and user-defined types. This repository also contains test cases for Database Wrapper Functions in [pnp-abstraction](https://github.ibm.com/cloud-sre/pnp-abstraction).

## When to deploy this repo
Normally, you only need to deploy this repository once in an Armada environment for dev/stage/prod. The major function calls in the main.go are all commented out, it is to prevent people from accidentally deploy this repository and recreate all the database tables unexpectedly.
```
	// Create tables for PNP if they are not already exist
	// ** NOTE** : EnsureTablesExist(database) WILL RECREATE ALL DATABASE TABLES
	//EnsureTablesExist(database) 

	// These test cases will be run on real postgres database
	//RunTests(database) 
	
	// This for testing Postgres Failover, the whole testing last 3-4 minutes
	//RunTestsForPostgresFailover(database) 
```

In the main.go, EnsureTablesExist(database) will recreate all the tables, indexes, and user-defined types in the PnP database. If this is an initial set up of an environment, and you need to create all the PnP tables, then uncomment out the line 
```
	//EnsureTablesExist(database) 
```
**Do not uncomment the two RunTests... lines, unless you really want to insert records into the newly created tables.**

For double protection, you also have to uncomment the lines inside **EnsureTablesExist** function, including the lines inside the if statement.
```
	if RECREATE_TABLE {
		// Tables have to be in a certain order. All the dependencies have to be dropped first
/*		dbcreate.DropIncidentJunctionTable(database)
		dbcreate.DropMaintenanceJunctionTable(database)
		dbcreate.DropDisplayNamesTable(database)
		dbcreate.DropVisibilityJunctionTable(database)
		dbcreate.DropVisibilityTable(database)
		dbcreate.DropTagJunctionTable(database)
		dbcreate.DropTagTable(database)
		dbcreate.DropWatchJunctionTable(database)
		dbcreate.DropResourceTable(database)
		dbcreate.DropIncidentTable(database)
		dbcreate.DropMaintenanceTable(database)
		dbcreate.DropCaseTable(database)
		dbcreate.DropWatchTable(database)
		dbcreate.DropSubscriptionTable(database)
		dbcreate.DropNotificationDescriptionTable(database)
		dbcreate.DropNotificationTable(database)
*/
	}
	
	// All tables being referenced as foreign keys have to be created first
/*	dbcreate.CreateResourceTable(database)
	dbcreate.CreateDisplayNamesTable(database)
	dbcreate.CreateVisibilityTable(database)
	dbcreate.CreateVisibilityJunctionTable(database)
	dbcreate.CreateTagTable(database)
	dbcreate.CreateTagJunctionTable(database)
	dbcreate.CreateIncidentTable(database)
	dbcreate.CreateIncidentJunctionTable(database)
	dbcreate.CreateMaintenanceTable(database)
	dbcreate.CreateMaintenanceJunctionTable(database)
	dbcreate.CreateCaseTable(database)
	dbcreate.CreateSubscriptionTable(database)
	dbcreate.CreateWatchTable(database)
	dbcreate.CreateWatchJunctionTable(database)
	dbcreate.CreateNotificationTable(database)
	dbcreate.CreateNotificationDescriptionTable(database)
*/
}
```
After you have uncommented the above mentioned lines for **EnsureTablesExist**, you can deploy the repository using kdep command. This repository will create a Kubernetes job, and will create all the PnP database tables if they are not already exist in the database. If tables are already exist in the database, it will RECREATE everything, and all the existing data in the database will be wiped out. Therefore we should **never** deploy this onto staging or production environment if database tables are already there. When the kubernetes job completes, do 
```
	kubectl logs <api-pnp-db job name>
```
to check the logs. 

## What to do if the definition of a database table needs to be changed in Staging or Production
If we need to make changes to any of the tables in Staging or Production, like adding a new column, changing the column size, or change the definition of a user-defined type. You have to update the CREATE statement in /dbcreate/create.go in this repository. You also need appropriate changes to the database wrapper functions in [pnp-abstraction](https://github.ibm.com/cloud-sre/pnp-abstraction). <br>
`DO NOT` redeploy api-pnp-db in staging and production environment. Instead, go to one of the Kubernetes pod container, using psql command to change the table definitions manually using `ALTER` command. For example, 
```
	ALTER TABLE test ADD COLUMN IF NOT EXISTS description varchar(30);
```
You can do the same for development environment, unless you would like to recreate all tables and start from scratch, then you redeploy this component.


## Unit Tests
The test cases in /pnp-db/test directory are to test on a real database tables, indexes, and user-defined types created by this component. They also test the database wrapper functions in pnp-abstraction repository. These test cases are not mocked, and they are intended to be run in **development** environment on development Postgres database only. 
For extra precaution, it is recommended to create a new set of tables to run these test cases even though we are testing on development environment, so that it will not mess up the data that other people are testing, therefore there are some manual steps to set this up.
1. In pnp-abstraction/db/constants.go, change the names of PnP tables to suffix with '1', for example:
```
	// PNP tables
	RESOURCE_TABLE_NAME                 = "resource_table1"
	DISPLAY_NAMES_TABLE_NAME            = "display_names_table1"
	VISIBILITY_JUNCTION_TABLE_NAME      = "visibility_junction_table1"
	VISIBILITY_TABLE_NAME               = "visibility_table1"
	TAG_JUNCTION_TABLE_NAME             = "tag_junction_table1"
	TAG_TABLE_NAME                      = "tag_table1"
	INCIDENT_TABLE_NAME                 = "incident_table1"
	INCIDENT_JUNCTION_TABLE_NAME        = "incident_junction_table1"
	MAINTENANCE_TABLE_NAME              = "maintenance_table1"
	MAINTENANCE_JUNCTION_TABLE_NAME     = "maintenance_junction_table1"
	CASE_TABLE_NAME                     = "case_table1"
	SUBSCRIPTION_TABLE_NAME             = "subscription_table1"
	WATCH_TABLE_NAME                    = "watch_table1"
	WATCH_JUNCTION_TABLE_NAME           = "watch_junction_table1"
	NOTIFICATION_TABLE_NAME             = "notification_table1"
	NOTIFICATION_DESCRIPTION_TABLE_NAME = "notification_description_table1"
```
2. In pnp-db/dbcreate/constants.go, change the names of all indexes to suffix with '1', for example:
```
	// Resource table indexes
	INDEX_RESOURCE_CNAME             = "resource_cname_index1"
	INDEX_RESOURCE_CTYPE             = "resource_ctype_index1"
	INDEX_RESOURCE_SERVICE_NAME      = "resource_service_name_index1"
	INDEX_RESOURCE_LOCATION          = "resource_location_index1"
	INDEX_RESOURCE_SCOPE             = "resource_scope_index1"
	INDEX_RESOURCE_SERVICE_INSTANCE  = "resource_service_instance_index1"
	INDEX_RESOURCE_RESOURCE_TYPE     = "resource_resource_type_index1"
	INDEX_RESOURCE_RESOURCE          = "resource_resource_index1"
	INDEX_RESOURCE_SOURCE_SOURCE_ID  = "resource_source_source_id_index1"
	INDEX_RESOURCE_CATALOG_PARENT_ID = "resource_catalog_parent_id_index1"

	// Display Names table indexes
	INDEX_DISPLAYNAMES_RESOURCE_ID = "displaynames_resource_id_index1"

	// Visibility table indexes
	INDEX_VISIBILITY_NAME = "visibility_name_index1"

	// Visibility Junction table indexes
	INDEX_VISIBILITYJUNCTION_RESOURCE_ID   = "visibilityjunction_resource_id_index1"
	INDEX_VISIBILITYJUNCTION_VISIBILITY_ID = "visibilityjunction_visibility_id_index1"

	// Tag Junction table indexes
	INDEX_TAGJUNCTION_RESOURCE_ID = "tagjunction_resource_id_index1"
	INDEX_TAGJUNCTION_TAG_ID      = "tagjunction_tag_id_index1" // tag's record_id index

	// Tag table indexes
	INDEX_TAG_ID = "tag_id_index1"

	// Incident table indexes
	INDEX_INCIDENT_SOURCE_CREATION_TIME = "incident_source_creation_time_index1"
	INDEX_INCIDENT_SOURCE_UPDATE_TIME   = "incident_source_update_time_index1"
	INDEX_INCIDENT_START_TIME           = "incident_start_time_index1"
	INDEX_INCIDENT_END_TIME             = "incident_end_time_index1"
	INDEX_INCIDENT_SOURCE_SOURCE_ID     = "incident_source_source_id_index1"

	// Incident Junction table indexes
	INDEX_INCIDENTJUNCTION_RESOURCE_ID = "incidentjunction_resource_id_index1"
	INDEX_INCIDENTJUNCTION_INCIDENT_ID = "incidentjunction_incident_id_index1"

	// Maintenance table indexes
	INDEX_MAINTENANCE_SOURCE_CREATION_TIME = "maintenance_source_creation_time_index1"
	INDEX_MAINTENANCE_SOURCE_UPDATE_TIME   = "maintenance_source_update_time_index1"
	INDEX_MAINTENANCE_START_TIME           = "maintenance_start_time_index1"
	INDEX_MAINTENANCE_END_TIME             = "maintenance_end_time_index1"
	INDEX_MAINTENANCE_SOURCE_SOURCE_ID     = "maintenance_source_source_id_index1"

	// Maintenance Junction table indexes
	INDEX_MAINTENANCEJUNCTION_RESOURCE_ID    = "maintenancejunction_resource_id_index1"
	INDEX_MAINTENANCEJUNCTION_MAINTENANCE_ID = "maintenancejunction_maintenance_id_index1"

	// Case table indexes
	INDEX_CASE_SOURCE_SOURCE_ID = "case_source_source_id_index1"

	// Subscription table indexes
	INDEX_SUBSCRIPTION_NAME = "subscription_name_index1"

	// Watch table indexes
	INDEX_WATCH_SUBSCRIPTION_ID = "watch_subscription_id_index1"
	INDEX_WATCH_KIND            = "watch_kind_index1"

	// Watch Junction table indexes
	INDEX_WATCHJUNCTION_RESOURCE_ID = "watchjunction_resource_id_index1"
	INDEX_WATCHJUNCTION_WATCH_ID    = "watchjunction_watch_id_index1"

	// Notification table indexes
	INDEX_NOTIFICATION_CNAME                = "notification_cname_index1"
	INDEX_NOTIFICATION_CTYPE                = "notification_ctype_index1"
	INDEX_NOTIFICATION_SERVICE_NAME         = "notification_service_name_index1"
	INDEX_NOTIFICATION_LOCATION             = "notification_location_index1"
	INDEX_NOTIFICATION_SCOPE                = "notification_scope_index1"
	INDEX_NOTIFICATION_SERVICE_INSTANCE     = "notification_service_instance_index1"
	INDEX_NOTIFICATION_RESOURCE_TYPE        = "notification_resource_type_index1"
	INDEX_NOTIFICATION_RESOURCE             = "notification_resource_index1"
	INDEX_NOTIFICATION_SOURCE_SOURCE_ID     = "notification_source_source_id_index1"
	INDEX_NOTIFICATION_SOURCE_CREATION_TIME = "notification_source_creation_time_index1"
	INDEX_NOTIFICATION_SOURCE_UPDATE_TIME   = "notification_source_update_time_index1"
	INDEX_NOTIFICATION_EVENT_TIME_START     = "notification_event_time_start_index1"
	INDEX_NOTIFICATION_EVENT_TIME_END       = "notification_event_time_end_index1"
	INDEX_NOTIFICATION_TYPE                 = "notification_type_index1"

	// Notification Description table indexes
	INDEX_NOTIFICATIONDESCRIPTION_NOTIFICATION_ID = "notificationdescription_notification_id_index1"
```
3. In pnp-db/dbcreate/constants.go, change the names of all types to suffix with '1', for example:
```
	// Incident User-defined type names
	UDT_INCIDENT_STATE          = "incident_state_type1"
	UDT_INCIDENT_CLASSIFICATION = "incident_classification_type1"
	UDT_INCIDENT_SEVERITY       = "incident_severity_type1"

	// Maintenance User-defined type names
	UDT_MAINTENANCE_STATE = "maintenance_state_type1"

	// Watch User-defined type names
	UDT_WATCH_WILDCARDS = "watch_wildcards_type1"
```
4. Uncomment the following lines in main() in pnp-db/main.go
```
	EnsureTablesExist(database) 

	RunTests(database) 
```
5. Uncomment the following lines in EnsureTablesExist(database *sql.DB) in pnp-db/main.go
```
func EnsureTablesExist(database *sql.DB) {
	log.Println("Creating User-defined types, tables and indexes.")

	if RECREATE_TABLE {
		// Tables have to be in a certain order. All the dependencies have to be dropped first
		dbcreate.DropIncidentJunctionTable(database)
		dbcreate.DropMaintenanceJunctionTable(database)
		dbcreate.DropDisplayNamesTable(database)
		dbcreate.DropVisibilityJunctionTable(database)
		dbcreate.DropVisibilityTable(database)
		dbcreate.DropTagJunctionTable(database)
		dbcreate.DropTagTable(database)
		dbcreate.DropWatchJunctionTable(database)
		dbcreate.DropResourceTable(database)
		dbcreate.DropIncidentTable(database)
		dbcreate.DropMaintenanceTable(database)
		dbcreate.DropCaseTable(database)
		dbcreate.DropWatchTable(database)
		dbcreate.DropSubscriptionTable(database)
		dbcreate.DropNotificationDescriptionTable(database)
		dbcreate.DropNotificationTable(database)

	}
	
	// All tables being referenced as foreign keys have to be created first
	dbcreate.CreateResourceTable(database)
	dbcreate.CreateDisplayNamesTable(database)
	dbcreate.CreateVisibilityTable(database)
	dbcreate.CreateVisibilityJunctionTable(database)
	dbcreate.CreateTagTable(database)
	dbcreate.CreateTagJunctionTable(database)
	dbcreate.CreateIncidentTable(database)
	dbcreate.CreateIncidentJunctionTable(database)
	dbcreate.CreateMaintenanceTable(database)
	dbcreate.CreateMaintenanceJunctionTable(database)
	dbcreate.CreateCaseTable(database)
	dbcreate.CreateSubscriptionTable(database)
	dbcreate.CreateWatchTable(database)
	dbcreate.CreateWatchJunctionTable(database)
	dbcreate.CreateNotificationTable(database)
	dbcreate.CreateNotificationDescriptionTable(database)

}
```
6. Before running the test cases, we need to find out what are the Gaas services first, and set the services in **gaasServices** variable in "RunTests(database *sql.DB)" in pnp-db/main.go, otherwise test cases will fail.
To find out what are Gaas services in OSSCatalog, set "debugCatalog=true" in pnp-abstraction/osscatalog/ossrecord.go, then gomake and deploy api-pnp-db in dev environment using `kdep ./development-values.yaml` command. 
This will dump all the osscatalog to the logs, and test cases will fail.
In the log, `kubectl -napi logs <pod_id>` look for services with "EntryType==GAAS", then replace 7 names of Gaas services in the **gaasServices** variable in the code. Then gomake and redeploy api-pnp-db again, test cases should all passed.

If all the test cases passed, you should see `ALL TEST RESULT: SUCCESSFUL` at the end of the log. The tests will do the cleanup at the end, it will remove all the records inserted to the database tables during the tests. If you want to see the data in Postgres database to see why a test cases failed, you can comment out the cleanup section in main.go.

If you have made changes to the test cases, make sure you comment those lines that you have uncommented in the above steps, and remove the suffices that you added, before checking them in to github master.

# Unit Tests For Postgres Failover
If you want to test the Postgres Failover, you have to test it together with Technical Foundation Team. Change the database configuration in main.go `db.Connect` line to point to the failover test database, and uncomment this line,
```
	//RunTestsForPostgresFailover(database) 
```
Then deploy the repository on **development** environment, then Technical Foundation team will stop the failover test database. Check the logs to see the result. All test cases should pass even when there is an outage in database for 10-15 seconds for the failover. This `RunTestsForPostgresFailover` function will run the same set of test cases in `RunTests` function repeatedly for 10 times, to ensure the failover occurs while the testing is still running, and can continue on successfully after the failover is complete.
