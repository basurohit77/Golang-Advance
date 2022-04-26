package db

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.ibm.com/cloud-sre/oss-globals/tlog"
)

const (
	INCIDENT_JUNCTION_TABLE_PREFIX = "ij"
	INCIDENT_JUNCTION_TABLE_ALIAS  = "ij."

	MAINTENANCE_JUNCTION_TABLE_PREFIX = "mj"
	MAINTENANCE_JUNCTION_TABLE_ALIAS  = "mj."

	WATCH_JUNCTION_TABLE_PREFIX = "wj"
	WATCH_JUNCTION_TABLE_ALIAS  = "wj."

	SELECT_RESOURCE_JOIN_INCIDENT_JUNCTION    = "SELECT " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + " FROM " + RESOURCE_TABLE_NAME + " " + RESOURCE_TABLE_ALIAS + " INNER JOIN " + INCIDENT_JUNCTION_TABLE_NAME + " " + INCIDENT_JUNCTION_TABLE_PREFIX + " ON " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + " = " + INCIDENT_JUNCTION_TABLE_ALIAS + INCIDENTJUNCTION_COLUMN_RESOURCE_ID
	SELECT_RESOURCE_JOIN_MAINTENANCE_JUNCTION = "SELECT " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + " FROM " + RESOURCE_TABLE_NAME + " " + RESOURCE_TABLE_ALIAS + " INNER JOIN " + MAINTENANCE_JUNCTION_TABLE_NAME + " " + MAINTENANCE_JUNCTION_TABLE_PREFIX + " ON " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + " = " + MAINTENANCE_JUNCTION_TABLE_ALIAS + MAINTENANCEJUNCTION_COLUMN_RESOURCE_ID
	SELECT_RESOURCE_JOIN_WATCH_JUNCTION       = "SELECT " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + " FROM " + RESOURCE_TABLE_NAME + " " + RESOURCE_TABLE_ALIAS + " INNER JOIN " + WATCH_JUNCTION_TABLE_NAME + " " + WATCH_JUNCTION_TABLE_PREFIX + " ON " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + " = " + WATCH_JUNCTION_TABLE_ALIAS + WATCHJUNCTION_COLUMN_RESOURCE_ID
)

// DeleteExpiredSubscriptions will deleted any subscription which time expired
func DeleteExpiredSubscriptions(dbConnection *sql.DB) (error, int) {

	now := time.Now()

	log.Println(tlog.Log() + "About to delete Subscriptions that has expired")

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// delete from Subscription Table when expiration time has passed
		result, err := dbConnection.Exec("DELETE FROM "+SUBSCRIPTION_TABLE_NAME+
			" WHERE "+SUBSCRIPTION_COLUMN_EXPIRATION+" < $1", now.UTC().Format("2006-01-02T15:04:05Z"))
		if err != nil {
			log.Println(tlog.Log()+" Error Deleting Expired Subscriptions: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}

		if result != nil {
			//log.Println(tlog.Log()+"Result: ", result)
			rowsAffected, err1 := result.RowsAffected()
			if err1 == nil {
				log.Printf(tlog.Log()+"Rows affected: %d\n", rowsAffected)
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteSubscriptionStatement remove a subscrition from the database
func DeleteSubscriptionStatement(dbConnection *sql.DB, recordID string) (error, int) {
	log.Println(tlog.Log()+"recordId: ", recordID)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		result, err := dbConnection.Exec("DELETE FROM "+SUBSCRIPTION_TABLE_NAME+
			" WHERE "+SUBSCRIPTION_COLUMN_RECORD_ID+" = $1", recordID)

		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error Deleting record: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		if result != nil {
			affected, errAffected := result.RowsAffected()
			if errAffected != nil {
				// failed to execute SQL statements. Exit
				log.Println(tlog.Log()+" parsing result: ", errAffected)
				return err, http.StatusInternalServerError
			}
			log.Println("Result of subscription delete: ", strconv.FormatInt(affected, 10))
		}
		retry = false
	}

	// Retry transaction in case Postgres crashes, and wait for failover
	retry = true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		result, err := dbConnection.Exec("DELETE FROM "+WATCH_TABLE_NAME+
			" WHERE "+WATCH_COLUMN_SUBSCRIPTION_ID+" = $1", recordID)
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error Deleting Subscription: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		if result != nil {
			affected, errAffected := result.RowsAffected()
			if errAffected != nil {
				// failed to execute SQL statements. Exit
				log.Println(tlog.Log()+" parsing result: ", errAffected)
				return err, http.StatusInternalServerError
			}
			log.Println("Result of watch delete: ", strconv.FormatInt(affected, 10))
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteWatchStatement removed a watch record from the databasse
func DeleteWatchStatement(dbConnection *sql.DB, subscriptionID string, recordID string) error {

	query := "DELETE FROM " + WATCH_TABLE_NAME + " WHERE " + WATCH_COLUMN_SUBSCRIPTION_ID + " = $1 AND " + WATCH_COLUMN_RECORD_ID + " = $2"
	log.Printf(tlog.Log()+"Query: %s -> %s, %s\n", query, subscriptionID, recordID)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		result, err := dbConnection.Exec(query, subscriptionID, recordID)
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error Deleting Watch: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err
			}
		}
		if result != nil {
			numrows, errResult := result.RowsAffected()
			if errResult != nil {
				log.Println(tlog.Log()+"Result Error: ", errResult.Error())
				return errResult
			}

			log.Println(tlog.Log()+"Result - Number of rows deleted: ", numrows)
			if numrows == 0 {
				errResult = errors.New("No rows have been deleted")
				return errResult
			}
		}
		retry = false
	}
	return nil
}

// DeleteMaintenanceStatement Only need to delete from Maintenance table, all other tables with FK referencing Maintenance table have ON DELETE CASCADE set up.
func DeleteMaintenanceStatement(dbConnection *sql.DB, recordID string) (error, int) {

	log.Println(tlog.Log()+"recordId: ", recordID)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Delete from Maintenance Table
		result, err := dbConnection.Exec("DELETE FROM "+MAINTENANCE_TABLE_NAME+
			" WHERE "+MAINTENANCE_COLUMN_RECORD_ID+" = $1", recordID)
		if result != nil {
			log.Printf(tlog.Log()+"Result: %+v\n", result.RowsAffected)
		}
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error Deleting Maintenance: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteIncidentStatement Only need to delete from Incident table, all other tables with FK referencing Incident table have ON DELETE CASCADE set up.
func DeleteIncidentStatement(dbConnection *sql.DB, recordID string) (error, int) {

	log.Println(tlog.Log()+"recordId: ", recordID)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Delete from Incident Table
		result, err := dbConnection.Exec("DELETE FROM "+INCIDENT_TABLE_NAME+
			" WHERE "+INCIDENT_COLUMN_RECORD_ID+" = $1", recordID)
		if result != nil {
			log.Printf(tlog.Log()+"Result: %+v\n", result.RowsAffected)
		}
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error Deleting Incident: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteCaseStatement remove case record from the database
func DeleteCaseStatement(dbConnection *sql.DB, recordID string) (error, int) {

	log.Println(tlog.Log()+"recordId: ", recordID)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Delete from Case Table
		result, err := dbConnection.Exec("DELETE FROM "+CASE_TABLE_NAME+
			" WHERE "+CASE_COLUMN_RECORD_ID+" = $1", recordID)
		if result != nil {
			log.Printf(tlog.Log()+"Result: %+v\n", result.RowsAffected)
		}
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error Deleting Case: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteOldResolvedIncidents Delete resolved Incidents with source_update_time older than 30 days, or
// incidents with pnp_removed is true and pnp_update_time older than 30 days
func DeleteOldResolvedIncidents(dbConnection *sql.DB, days int) (error, int) {

	today := time.Now()
	pastTime := today.AddDate(0, 0, -1*days)

	log.Println(tlog.Log()+"About to delete (incidents with state is resolved OR incidents with pnp_removed=true) and source_update_time is before ", pastTime.UTC().Format("2006-01-02T15:04:05Z"))

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Then delete from Incident Table
		result, err := dbConnection.Exec("DELETE FROM "+INCIDENT_TABLE_NAME+
			" WHERE ("+INCIDENT_COLUMN_STATE+" = $1 OR "+INCIDENT_COLUMN_PNP_REMOVED+" = $2) AND "+INCIDENT_COLUMN_SOURCE_UPDATE_TIME+" < $3",
			"resolved", "true", pastTime.UTC().Format("2006-01-02T15:04:05Z"))
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error Deleting Old Incidents: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		if result != nil {
			//log.Println(tlog.Log()+"Result: ", result)
			rowsAffected, err1 := result.RowsAffected()
			if err1 == nil {
				log.Printf(tlog.Log()+"Rows affected: %d\n", rowsAffected)
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteOldCompleteMaintenance Delete completed Maintenance with source_update_time older than 30 days, or
// maintenance with pnp_removed is true and source_update_time older than 30 days
func DeleteOldCompleteMaintenance(dbConnection *sql.DB, days int) (error, int) {

	today := time.Now()
	pastTime := today.AddDate(0, 0, -1*days)

	log.Println(tlog.Log()+"About to delete (maintenance with state is complete OR maintenance with pnp_removed=true) AND source_update_time is before ", pastTime.UTC().Format("2006-01-02T15:04:05Z"))

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Then delete from Incident Table
		result, err := dbConnection.Exec("DELETE FROM "+MAINTENANCE_TABLE_NAME+
			" WHERE ("+MAINTENANCE_COLUMN_STATE+" = $1 OR "+MAINTENANCE_COLUMN_PNP_REMOVED+" = $2) AND "+MAINTENANCE_COLUMN_SOURCE_UPDATE_TIME+" < $3",
			"complete", "true", pastTime.UTC().Format("2006-01-02T15:04:05Z"))
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error in DeleteOldCompleteMaintenance: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}

		if result != nil {
			//log.Println(tlog.Log()+"Result: ", result)
			rowsAffected, err1 := result.RowsAffected()
			if err1 == nil {
				log.Printf(tlog.Log()+"Rows affected: %d\n", rowsAffected)
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

func deleteMaintenanceJunctionStatement(tx *sql.Tx, resourceID string, maintenanceID string) error {
	log.Println(tlog.Log()+"resourceId: ", resourceID, ", maintenanceId: ", maintenanceID)

	// Delete from MaintenanceJunction Table
	result, err := tx.Exec("DELETE FROM "+MAINTENANCE_JUNCTION_TABLE_NAME+
		" WHERE "+MAINTENANCEJUNCTION_COLUMN_RESOURCE_ID+" = $1 AND "+MAINTENANCEJUNCTION_COLUMN_MAINTENANCE_ID+" = $2", resourceID, maintenanceID)
	if result != nil {
		log.Printf(tlog.Log()+"Result: %+v\n", result)
	}
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in deleteMaintenanceJunctionStatement Rollback: ", err1)
			return err1
		}
		log.Println(tlog.Log()+" Error Deleting MaintenanceJunction: ", err)
		return err
	}

	return err
}

func deleteIncidentJunctionStatement(tx *sql.Tx, resourceID string, incidentID string) error {
	log.Println(tlog.Log()+"resourceId: ", resourceID, ", incidentId: ", incidentID)

	// Delete from IncidentJunction Table
	result, err := tx.Exec("DELETE FROM "+INCIDENT_JUNCTION_TABLE_NAME+
		" WHERE "+INCIDENTJUNCTION_COLUMN_RESOURCE_ID+" = $1 AND "+INCIDENTJUNCTION_COLUMN_INCIDENT_ID+" = $2", resourceID, incidentID)
	if result != nil {
		log.Printf(tlog.Log()+"Result: %+v\n", result)
	}
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in deleteIncidentJunctionStatement Rollback: ", err1)
			return err1
		}
		log.Println(tlog.Log()+" Error Deleting IncidentJunction: ", err)
		return err
	}

	return err
}

func deleteVisibilityJunctionStatement(tx *sql.Tx, resourceID string, visibilityID string) error {
	log.Println(tlog.Log()+"resourceId: ", resourceID, ", visibilityId: ", visibilityID)

	result, err := tx.Exec("DELETE FROM "+VISIBILITY_JUNCTION_TABLE_NAME+
		" WHERE "+VISIBILITYJUNCTION_COLUMN_RESOURCE_ID+" = $1 AND "+VISIBILITYJUNCTION_COLUMN_VISIBILITY_ID+" = $2", resourceID, visibilityID)
	if result != nil {
		log.Printf(tlog.Log()+"Result: %+v\n", result)
	}
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in deleteVisibilityJunctionStatement Rollback: ", err1)
			return err1
		}
		log.Println(tlog.Log()+" Error Deleting VisibilityJunction: ", err)
		return err
	}

	return err
}

func deleteTagJunctionStatement(tx *sql.Tx, resourceID string, tagID string) error {
	log.Println(tlog.Log()+"resourceId: ", resourceID, ", tagId: ", tagID)

	result, err := tx.Exec("DELETE FROM "+TAG_JUNCTION_TABLE_NAME+
		" WHERE "+TAGJUNCTION_COLUMN_RESOURCE_ID+" = $1 AND "+TAGJUNCTION_COLUMN_TAG_ID+" = $2", resourceID, tagID)
	if result != nil {
		log.Printf(tlog.Log()+"Result: %+v\n", result)
	}
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in deleteTagJunctionStatement Rollback: ", err1)
			return err1
		}
		log.Println(tlog.Log()+" Error Deleting TagJunction: ", err)
		return err
	}

	return err
}

func deleteDisplayNameStatement(tx *sql.Tx, resourceID string, name string, language string) error {

	log.Println(tlog.Log()+"resourceId: ", resourceID, ", name: ", name, ", language: ", language)

	// Delete from DisplayName Table
	result, err := tx.Exec("DELETE FROM "+DISPLAY_NAMES_TABLE_NAME+
		" WHERE "+DISPLAYNAMES_COLUMN_RESOURCE_ID+" = $1 AND "+DISPLAYNAMES_COLUMN_NAME+" = $2 AND "+DISPLAYNAMES_COLUMN_LANGUAGE+" = $3", resourceID, name, language)
	if result != nil {
		log.Printf(tlog.Log()+"Result: %+v\n", result)
	}
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in deleteDisplayNameStatement Rollback: ", err1)
			return err1
		}
		log.Println(tlog.Log()+" Error Deleting DisplayName: ", err)
		return err
	}

	return err
}

// DeleteArchivedResources Only delete resource with state=archived and no more records in junction tables
// Following diagram https://github.ibm.com/cloud-sre/pnp-status#resource-deletion
// Only need to delete from Resource table, all other tables with FK referencing Resource table have ON DELETE CASCADE set up.
func DeleteArchivedResources(dbConnection *sql.DB) (error, int) {

	/*  Do not do the incidents and maintenance cleanup here, it is done in the db_cleaner, which will
	    cleanup old incidents and maintenance, and expired subscriptions first

		//Delete Old Incidents 30 days
		err, _ := DeleteOldResolvedIncidents(dbConnection, 30)
		if err != nil {
			log.Println(tlog.Log()+" Error Deleting Old Incidents: ", err)
			return err, http.StatusInternalServerError
		}

		//Delete Old Maintenance 30 days
		err, _ = DeleteOldCompleteMaintenance(dbConnection, 30)
		if err != nil {
			log.Println(tlog.Log()+" Error Deleting Old Maintenance: ", err)
			return err, http.StatusInternalServerError
		}
	*/

	log.Println(tlog.Log() + "About to delete archived resources and there is no records in the incident/maintenance/watch junction tables")

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		//https://github.ibm.com/cloud-sre/pnp-abstraction/issues/646
		// result, err := dbConnection.Exec("DELETE FROM " + RESOURCE_TABLE_NAME +
		// 	" WHERE " + RESOURCE_COLUMN_STATE + " = 'archived' AND " + RESOURCE_COLUMN_RECORD_ID +
		// 	" NOT IN (" + SELECT_RESOURCE_JOIN_INCIDENT_JUNCTION + ") AND " + RESOURCE_COLUMN_RECORD_ID +
		// 	" NOT IN (" + SELECT_RESOURCE_JOIN_MAINTENANCE_JUNCTION + ") AND " + RESOURCE_COLUMN_RECORD_ID +
		// 	" NOT IN (" + SELECT_RESOURCE_JOIN_WATCH_JUNCTION + ")")

		// Delete from Resource Table where state='archived' AND resource_id DOES NOT EXIST in junction tables
		result, err := dbConnection.Exec("DELETE FROM " + RESOURCE_TABLE_NAME +
			" WHERE " + RESOURCE_COLUMN_STATE + " = 'archived' " +
			" AND " + RESOURCE_COLUMN_RECORD_ID + " NOT IN (" + SELECT_RESOURCE_JOIN_INCIDENT_JUNCTION + ")" +
			" AND " + RESOURCE_COLUMN_RECORD_ID + " NOT IN (" + SELECT_RESOURCE_JOIN_MAINTENANCE_JUNCTION + ")")

		if err != nil {
			log.Println(tlog.Log()+" Error Deleting Resource: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		if result != nil {
			//log.Println(tlog.Log()+"Result: ", result)
			rowsAffected, err1 := result.RowsAffected()
			if err1 == nil {
				log.Printf(tlog.Log()+"Rows affected: %d\n", rowsAffected)
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteResourceStatement Only delete resource with recordid and state=archived
// Only need to delete from Resource table, all other tables with FK referencing Resource table have ON DELETE CASCADE set up.
func DeleteResourceStatement(dbConnection *sql.DB, recordID string) (error, int) {
	log.Println(tlog.Log()+"recordId: ", recordID)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Delete from Resource Table
		result, err := dbConnection.Exec("DELETE FROM "+RESOURCE_TABLE_NAME+
			" WHERE "+RESOURCE_COLUMN_RECORD_ID+" = $1 AND "+RESOURCE_COLUMN_STATE+" = 'archived'", recordID)
		if result != nil {
			log.Printf(tlog.Log()+"Result: %+v\n", result)
		}
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error Deleting Resource: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteNotificationStatement Only need to delete from Notification table, all other tables with FK referencing Notification table have ON DELETE CASCADE set up.
func DeleteNotificationStatement(dbConnection *sql.DB, recordID string) (error, int) {
	log.Println(tlog.Log()+"recordId: ", recordID)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Delete from Incident Table
		result, err := dbConnection.Exec("DELETE FROM "+NOTIFICATION_TABLE_NAME+
			" WHERE "+NOTIFICATION_COLUMN_RECORD_ID+" = $1", recordID)
		if result != nil {
			log.Printf(tlog.Log()+"Result: %+v\n", result.RowsAffected)
		}
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error Deleting Incident: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteOldNonSecurityNotifications Delete non-security notifications that are older than 90 days, all other tables with FK referencing Notification table have ON DELETE CASCADE set up.
func DeleteOldNonSecurityNotifications(dbConnection *sql.DB, days int) (error, int) {
	today := time.Now()
	pastTime := today.AddDate(0, 0, -1*days)

	log.Println(tlog.Log()+"About to delete Notification which are not security type and source_update_time is before ", pastTime.UTC().Format("2006-01-02T15:04:05Z"))

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Then delete from Incident Table
		result, err := dbConnection.Exec("DELETE FROM "+NOTIFICATION_TABLE_NAME+
			" WHERE "+NOTIFICATION_COLUMN_TYPE+" != $1 AND "+NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME+" < $2", "security", pastTime.UTC().Format("2006-01-02T15:04:05Z"))
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error in deleting old notifications: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		if result != nil {
			//log.Println(tlog.Log()+"Result: ", result)
			rowsAffected, err1 := result.RowsAffected()
			if err1 == nil {
				log.Printf(tlog.Log()+"Rows affected: %d\n", rowsAffected)
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteOldIncidentNotifications - Removes old incident notifications whose associated incident has been removed and that have not been updated in the past number of days provided
func DeleteOldIncidentNotifications(dbConnection *sql.DB, days int) (error, int) {
	today := time.Now()
	pastTime := today.AddDate(0, 0, -1*days)

	log.Println(tlog.Log()+"About to delete incident notifications whose incidents have been removed and whose source_update_time is before ", pastTime.UTC().Format("2006-01-02T15:04:05Z"))

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Then delete from Notification Table (embedded SELECT checks whether the associated incident still exists in the incident table)
		result, err := dbConnection.Exec("DELETE FROM "+NOTIFICATION_TABLE_NAME+
			" WHERE "+NOTIFICATION_COLUMN_TYPE+" = $1 AND "+NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME+" < $2"+
			" AND NOT EXISTS (SELECT 1 FROM "+INCIDENT_TABLE_NAME+" WHERE "+INCIDENT_TABLE_NAME+"."+INCIDENT_COLUMN_SOURCE_ID+
			" = "+NOTIFICATION_TABLE_NAME+"."+NOTIFICATION_COLUMN_INCIDENT_ID+" AND "+INCIDENT_TABLE_NAME+"."+INCIDENT_COLUMN_SOURCE+
			" = "+NOTIFICATION_TABLE_NAME+"."+NOTIFICATION_COLUMN_SOURCE+");", "incident", pastTime.UTC().Format("2006-01-02T15:04:05Z"))
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error in deleting old notifications: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		if result != nil {
			//log.Println(tlog.Log()+"Result: ", result)
			rowsAffected, err1 := result.RowsAffected()
			if err1 == nil {
				log.Printf(tlog.Log()+"Rows affected: %d\n", rowsAffected)
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteOldMaintenanceNotifications - Removes old maintenance notifications whose associated maintenance has been removed and that have not been updated in the past number of days provided
func DeleteOldMaintenanceNotifications(dbConnection *sql.DB, days int) (error, int) {
	today := time.Now()
	pastTime := today.AddDate(0, 0, -1*days)

	log.Println(tlog.Log()+"About to delete maintenance notifications whose maintenance have been removed and whose source_update_time is before ", pastTime.UTC().Format("2006-01-02T15:04:05Z"))

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Then delete from Notification Table (embedded SELECT checks whether the associated maintanance still exists in the maintanance table)
		result, err := dbConnection.Exec("DELETE FROM "+NOTIFICATION_TABLE_NAME+
			" WHERE "+NOTIFICATION_COLUMN_TYPE+" = $1 AND "+NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME+" < $2"+
			" AND NOT EXISTS (SELECT 1 FROM "+MAINTENANCE_TABLE_NAME+" WHERE "+MAINTENANCE_TABLE_NAME+"."+MAINTENANCE_COLUMN_SOURCE_ID+
			" = "+NOTIFICATION_TABLE_NAME+"."+NOTIFICATION_COLUMN_SOURCE_ID+" AND "+MAINTENANCE_TABLE_NAME+"."+MAINTENANCE_COLUMN_SOURCE+
			" = "+NOTIFICATION_TABLE_NAME+"."+NOTIFICATION_COLUMN_SOURCE+");", "maintenance", pastTime.UTC().Format("2006-01-02T15:04:05Z"))
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error in deleting old notifications: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		if result != nil {
			//log.Println(tlog.Log()+"Result: ", result)
			rowsAffected, err1 := result.RowsAffected()
			if err1 == nil {
				log.Printf(tlog.Log()+"Rows affected: %d\n", rowsAffected)
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

// DeleteOldAnnouncmentNotifications deletes announcement notifications that are older than the provided days old
func DeleteOldAnnouncmentNotifications(dbConnection *sql.DB, days int) (error, int) {
	today := time.Now()
	pastTime := today.AddDate(0, 0, -1*days)

	log.Println(tlog.Log()+"About to delete announcement notifications whose source_update_time is before ", pastTime.UTC().Format("2006-01-02T15:04:05Z"))

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Then delete from Incident Table
		result, err := dbConnection.Exec("DELETE FROM "+NOTIFICATION_TABLE_NAME+
			" WHERE "+NOTIFICATION_COLUMN_TYPE+" = $1 AND "+NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME+" < $2", "announcement", pastTime.UTC().Format("2006-01-02T15:04:05Z"))
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error in deleting old notifications: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}
		if result != nil {
			//log.Println(tlog.Log()+"Result: ", result)
			rowsAffected, err1 := result.RowsAffected()
			if err1 == nil {
				log.Printf(tlog.Log()+"Rows affected: %d\n", rowsAffected)
			}
		}
		retry = false
	}
	return nil, http.StatusOK
}

func deleteNotificationDescriptionStatement(tx *sql.Tx, recordID string) error {
	log.Println(tlog.Log()+"recordID: ", recordID)

	// Delete from DisplayName Table
	result, err := tx.Exec("DELETE FROM "+NOTIFICATION_DESCRIPTION_TABLE_NAME+
		" WHERE "+NOTIFICATIONDESCRIPTION_COLUMN_RECORD_ID+" = $1", recordID)
	if result != nil {
		log.Printf(tlog.Log()+"Result: %+v\n", result)
	}
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
			return err1
		}
		log.Println(tlog.Log()+" Error Deleting Notification Description: ", err)
		return err
	}

	return err
}

// DeleteOldMaintenanceByState - Removes old maintenance records whose state is passed  and that have not been updated in the past number of days provided
// https://github.ibm.com/cloud-sre/pnp-status/issues/367
func DeleteOldMaintenanceByState(dbConnection *sql.DB, days int, state string) (int, error) {
	today := time.Now()
	pastTime := today.AddDate(0, 0, -1*days)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		// Then delete from Notification Table (embedded SELECT checks that only not associted maintenance recored with notification are deleted)
		result, err := dbConnection.Exec("DELETE FROM "+MAINTENANCE_TABLE_NAME+
			" WHERE "+MAINTENANCE_COLUMN_STATE+" = $1"+
			" AND "+MAINTENANCE_COLUMN_SOURCE_UPDATE_TIME+" < $2"+
			" AND NOT EXISTS (SELECT 1 FROM "+NOTIFICATION_TABLE_NAME+
			" WHERE "+MAINTENANCE_TABLE_NAME+"."+MAINTENANCE_COLUMN_SOURCE_ID+" = "+NOTIFICATION_TABLE_NAME+"."+NOTIFICATION_COLUMN_SOURCE_ID+
			" AND "+MAINTENANCE_TABLE_NAME+"."+MAINTENANCE_COLUMN_SOURCE+" = "+NOTIFICATION_TABLE_NAME+"."+NOTIFICATION_COLUMN_SOURCE+
			" AND "+NOTIFICATION_COLUMN_CTYPE+"= $3);", state, pastTime.UTC().Format("2006-01-02T15:04:05Z"), "maintenance")

		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+" Error in deleting old notifications: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return http.StatusInternalServerError, err
			}
		}
		if result != nil {
			//log.Println(tlog.Log()+"Result: ", result)
			rowsAffected, err1 := result.RowsAffected()
			if err1 == nil {
				log.Printf(tlog.Log()+"Rows affected: %d\n", rowsAffected)
			}
		}
		retry = false
	}
	return http.StatusOK, nil
}
