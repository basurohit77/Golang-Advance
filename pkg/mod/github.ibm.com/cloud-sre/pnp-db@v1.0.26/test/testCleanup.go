package test

import (
	"database/sql"
	"log"
	"strconv"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
)


func TestCleanup(dbConnection *sql.DB) int {
	FCT := "TestCleanup: "
	failed := 0

	// DELETE FROM resource_table WHERE source_id LIKE '%source%';
	strQuery := "DELETE FROM "+db.RESOURCE_TABLE_NAME+" WHERE "+db.RESOURCE_COLUMN_SOURCE_ID+" LIKE '%test_source%'"
	result, err := dbConnection.Exec(strQuery)
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" Error Deleting record: ", err)
		failed++
	}
	affected, errAffected := result.RowsAffected()
	if errAffected != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" parsing result: ", errAffected)
	}
	log.Println("Result of resource table delete: ", strconv.FormatInt(affected, 10))

	//DELETE FROM visibility_table WHERE name LIKE 'vis%';
	strQuery = "DELETE FROM "+db.VISIBILITY_TABLE_NAME+" WHERE "+db.VISIBILITY_COLUMN_NAME+" LIKE 'test_vis%'"
	result, err = dbConnection.Exec(strQuery)
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" Error Deleting record: ", err)
		failed++
	}
	affected, errAffected = result.RowsAffected()
	if errAffected != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" parsing result: ", errAffected)
	}
	log.Println("Result of visibility table delete: ", strconv.FormatInt(affected, 10))

	//DELETE FROM tag_table WHERE id LIKE 'tag%';
	strQuery = "DELETE FROM "+db.TAG_TABLE_NAME+" WHERE "+db.TAG_COLUMN_ID+" LIKE 'test_tag%'"
	result, err = dbConnection.Exec(strQuery)
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" Error Deleting record: ", err)
		failed++
	}
	affected, errAffected = result.RowsAffected()
	if errAffected != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" parsing result: ", errAffected)
	}
	log.Println("Result of tag table delete: ", strconv.FormatInt(affected, 10))

	//DELETE FROM incident_table WHERE source_id LIKE 'source%';
	strQuery = "DELETE FROM "+db.INCIDENT_TABLE_NAME+" WHERE "+db.INCIDENT_COLUMN_SOURCE_ID+" LIKE '%test_source%'"
	result, err = dbConnection.Exec(strQuery)
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" Error Deleting record: ", err)
		failed++
	}
	affected, errAffected = result.RowsAffected()
	if errAffected != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" parsing result: ", errAffected)
	}
	log.Println("Result of incident table delete: ", strconv.FormatInt(affected, 10))

	//DELETE FROM maintenance_table WHERE source_id LIKE 'source%';
	strQuery = "DELETE FROM "+db.MAINTENANCE_TABLE_NAME+" WHERE "+db.MAINTENANCE_COLUMN_SOURCE_ID+" LIKE '%test_source%'"
	result, err = dbConnection.Exec(strQuery)
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" Error Deleting record: ", err)
		failed++
	}
	affected, errAffected = result.RowsAffected()
	if errAffected != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" parsing result: ", errAffected)
	}
	log.Println("Result of maintenance table delete: ", strconv.FormatInt(affected, 10))

	//DELETE FROM subscription_table WHERE name LIKE 'name%';
	strQuery = "DELETE FROM "+db.SUBSCRIPTION_TABLE_NAME+" WHERE "+db.SUBSCRIPTION_COLUMN_NAME+" LIKE '&test_name%'"
	result, err = dbConnection.Exec(strQuery)
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" Error Deleting record: ", err)
		failed++
	}
	affected, errAffected = result.RowsAffected()
	if errAffected != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" parsing result: ", errAffected)
	}
	log.Println("Result of subscription table delete: ", strconv.FormatInt(affected, 10))

	//DELETE FROM notification_table WHERE source_id LIKE 'source%';
	strQuery = "DELETE FROM "+db.NOTIFICATION_TABLE_NAME+" WHERE "+db.NOTIFICATION_COLUMN_SOURCE_ID+" LIKE '%test_source%'"
	result, err = dbConnection.Exec(strQuery)
	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" Error Deleting record: ", err)
		failed++
	}
	affected, errAffected = result.RowsAffected()
	if errAffected != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+" parsing result: ", errAffected)
	}
	log.Println("Result of notification table delete: ", strconv.FormatInt(affected, 10))

	return failed
}

