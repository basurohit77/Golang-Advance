package postgres

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
)

func TestSimpleExercise(t *testing.T) {
	setupTestFunctions()

	db, err := dbConnect("host", 8080, "databaseName", "user", "password", "sslmode")
	if err != nil {
		t.Fatal(err)
	}

	isActiveTest(db)
	getNotificationByQueryTest(db, "crn=crn:v1:bluemix:public::::::", -1, 0)

	disconnectTest(db)
}

func setupTestFunctions() {
	dbConnect = fConnect
	dbDisconnect = fDisconnect
	dbIsActive = fIsActive
	dbGetNotificationByQuery = fGetNotificationByQuery
	dbGetNotificationByRecordID = fGetNotificationByRecordID
	dbInsertNotification = fInsertNotification
	dbUpdateNotification = fUpdateNotification

}

func connectTest(host string, port int, databaseName string, user string, password string, sslmode string) (database *sql.DB, err error) {
	db := new(sql.DB)
	return db, nil
}

func disconnectTest(database *sql.DB) {
	return
}

func isActiveTest(database *sql.DB) (err error) {
	if database == nil {
		return errors.New("no db")
	}
	return nil
}

var queryTestReturnZero = false
var queryTestReturnBadHTTP = false

func getNotificationByQueryTest(database *sql.DB, query string, limit int, offset int) (result *[]datastore.NotificationReturn, int, httpResponse int, err error) {

	if queryTestReturnZero {
		return nil, 0, http.StatusOK, nil
	}

	list := make([]datastore.NotificationReturn, 0, 1)

	list = append(list, CreateNotificationReturn("cloudant", "123", "announcement", "services", "INC001001", "crn:v1:bluemix:public:cloudantnosqldb:::::", time.Minute*5, time.Minute*5, time.Minute*10, time.Minute*10, time.Minute*11, time.Minute*11, "cloudantnosqldb", "Test CIE", "This is a CIE for the test program"))

	return &list, 1, http.StatusOK, nil
}

func getNotificationByRecordIDTest(database *sql.DB, recordID string) (result *datastore.NotificationReturn, httpResponse int, err error) {
	if queryTestReturnZero {
		return nil, http.StatusOK, nil
	}
	if queryTestReturnBadHTTP {
		return nil, http.StatusInternalServerError, nil
	}
	nr := CreateNotificationReturn("cloudant", "123", "announcement", "services", "INC001001", "crn:v1:bluemix:public:cloudantnosqldb:::::", time.Minute*5, time.Minute*5, time.Minute*10, time.Minute*10, time.Minute*11, time.Minute*11, "cloudantnosqldb", "Test CIE", "This is a CIE for the test program")
	return &nr, http.StatusOK, nil
}

func insertNotificationTest(database *sql.DB, itemToInsert *datastore.NotificationInsert) (recordID string, httpStatus int, err error) {
	return "", http.StatusOK, nil
}

func updateNotificationTest(database *sql.DB, itemToInsert *datastore.NotificationInsert) (int, error) {
	return http.StatusOK, nil
}

func TestApi(t *testing.T) {
	Open("", "", "", "", "", "")
	Open("host", "", "", "", "", "")
	Open("host", "port", "", "", "", "")
	Open("host", "port", "db", "", "", "")
	Open("host", "port", "db", "user", "", "")
	Open("host", "port", "db", "user", "pw", "")
	db, _ := Open("host", "8080", "db", "user", "pw", "require")
	ctx := ctxt.Context{}
	db.GetAllNotifications(ctx, "", true)
	db.GetNotificationByRecordID(ctx, "foobar", "")
	db.GetNotificationByQuery(ctx, "crn=crn:v1::::::::", 200, 0, "")
	queryTestReturnZero = true
	db.GetNotificationByQuery(ctx, "crn=crn:v1::::::::", 200, 0, "")
	db.GetNotificationByRecordID(ctx, "foobar", "")

	queryTestReturnZero = false
	queryTestReturnBadHTTP = true
	db.GetNotificationByQuery(ctx, "crn=crn:v1::::::::", 200, 0, "")
	db.GetNotificationByRecordID(ctx, "foobar", "")
	queryTestReturnBadHTTP = false

	queryTestReturnZero = false
	dbDeleteNotification = mySimpleDeleteNotification
	db.DeleteAllNotifications(ctx)
	mySimpleDeleteNotificationErr1 = true
	db.DeleteAllNotifications(ctx)
	mySimpleDeleteNotificationErr1 = false
	mySimpleDeleteNotificationErr2 = true
	db.DeleteAllNotifications(ctx)

	db.Close()

}

var mySimpleDeleteNotificationErr1 = false
var mySimpleDeleteNotificationErr2 = false

func mySimpleDeleteNotification(dbConnection *sql.DB, recordID string) (error, int) {
	if mySimpleDeleteNotificationErr1 {
		return errors.New("Fake Error"), 200
	}
	if mySimpleDeleteNotificationErr2 {
		return nil, 500
	}
	return nil, 200
}
