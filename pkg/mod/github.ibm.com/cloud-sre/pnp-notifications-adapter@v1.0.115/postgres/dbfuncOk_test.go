package postgres

import (
	"database/sql"
	"testing"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

func TestToPreventError(t *testing.T) {
}

func fConnect(host string, port int, databaseName string, user string, password string, sslmode string) (database *sql.DB, err error) {
	return connectTest(host, port, databaseName, user, password, sslmode)
}

func fDisconnect(database *sql.DB) {
	disconnectTest(database)
}

func fIsActive(database *sql.DB) (err error) {
	return isActiveTest(database)
}

func fGetNotificationByQuery(database *sql.DB, query string, limit int, offset int) (*[]datastore.NotificationReturn, int, error, int) {
	nr, i1, i2, err := getNotificationByQueryTest(database, query, limit, offset)
	return nr, i1, err, i2
}

func fGetNotificationByRecordID(database *sql.DB, recordID string) (*datastore.NotificationReturn, error, int) {
	nr, i1, err := getNotificationByRecordIDTest(database, recordID)
	return nr, err, i1
}

func fInsertNotification(database *sql.DB, itemToInsert *datastore.NotificationInsert) (string, error, int) {
	str1, i1, err := insertNotificationTest(database, itemToInsert)
	return str1, err, i1
}

func fUpdateNotification(database *sql.DB, itemToInsert *datastore.NotificationInsert) (error, int) {
	i1, err := updateNotificationTest(database, itemToInsert)
	return err, i1
}
