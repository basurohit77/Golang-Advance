package postgres

import (
	"database/sql"
	"log"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
)

// The purpose of this file is to map database functions to actual implementations.  This is primarily helpful to enable unit tests.

// DBConnectFunc is mockable for unit test purposes
type DBConnectFunc func(host string, port int, databaseName string, user string, password string, sslmode string) (database *sql.DB, err error)

// DBDisconnectFunc is mockable for unit test purposes
type DBDisconnectFunc func(database *sql.DB)

// DBIsActiveFunc is mockable for unit test purposes
type DBIsActiveFunc func(database *sql.DB) (err error)

// DBGetNotificationByQueryFunc is mockable for unit test purposes
type DBGetNotificationByQueryFunc func(database *sql.DB, query string, limit int, offset int) (*[]datastore.NotificationReturn, int, error, int)

// DBGetNotificationByRecordIDFunc is mockable for unit test purposes
type DBGetNotificationByRecordIDFunc func(database *sql.DB, recordID string) (*datastore.NotificationReturn, error, int)

// DBInsertNotificationFunc is mockable for unit test purposes
type DBInsertNotificationFunc func(database *sql.DB, itemToInsert *datastore.NotificationInsert) (string, error, int)

// DBUpdateNotificationFunc is mockable for unit test purposes
type DBUpdateNotificationFunc func(database *sql.DB, itemToInsert *datastore.NotificationInsert) (error, int)

// DBDeleteNotificationFunc is mockable for unit test purposes
type DBDeleteNotificationFunc func(dbConnection *sql.DB, recordId string) (error, int)

var dbConnect DBConnectFunc
var dbDisconnect DBDisconnectFunc
var dbIsActive DBIsActiveFunc
var dbGetNotificationByQuery DBGetNotificationByQueryFunc
var dbGetNotificationByRecordID DBGetNotificationByRecordIDFunc
var dbInsertNotification DBInsertNotificationFunc
var dbUpdateNotification DBUpdateNotificationFunc
var dbDeleteNotification DBDeleteNotificationFunc

// SetupDBFunctions assigns functions to the database functions.
func SetupDBFunctions() {
	dbConnect = db.Connect
	dbDisconnect = db.Disconnect
	dbIsActive = db.IsActive
	dbGetNotificationByQuery = db.GetNotificationByQuery
	dbGetNotificationByRecordID = db.GetNotificationByRecordID
	dbInsertNotification = db.InsertNotification
	dbUpdateNotification = db.UpdateNotification
	dbDeleteNotification = db.DeleteNotificationStatement

	if dbConnect == nil {
		log.Println("ERROR (SetupDBFunctions): Unsuccessful setting up DB functions")
	}
}

// SetupDBFunctionsForUT assigns functions to the database functions.
// This is helpful for allowing unit test to mock these functions
func SetupDBFunctionsForUT(f1 DBConnectFunc, f2 DBDisconnectFunc, f3 DBIsActiveFunc, f4 DBGetNotificationByQueryFunc, f5 DBGetNotificationByRecordIDFunc, f7 DBInsertNotificationFunc, f8 DBUpdateNotificationFunc) {
	dbConnect = f1
	dbDisconnect = f2
	dbIsActive = f3
	dbGetNotificationByQuery = f4
	dbGetNotificationByRecordID = f5
	dbInsertNotification = f7
	dbUpdateNotification = f8
	dbDeleteNotification = nil

}
