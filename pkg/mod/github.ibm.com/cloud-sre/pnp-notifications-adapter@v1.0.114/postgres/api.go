package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/exmon"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/api"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/convert"
)

const (
	// GenericCRNMask is a generic version of the CRN mask that encompasses all resources
	GenericCRNMask = "crn:v1::::::::"
	// CRN is the parameter to the database query for the crn
	CRN = "crn"
)

var (
	dbMaxOpenConns, isDBMaxOpenConnsSet = lookupIntEnvVariable("DB_MAX_OPEN_CONNS") // optional parameter. default value is no maximum
)

// Database is the representation of the common database contstruct
type Database struct {
	Connection *sql.DB
	Port       int
	Host       string
	ReusedConn bool
}

var connectionMap = make(map[string]*Database)
var connectionMapMutex = new(sync.Mutex)

// Reuse will reuse an existing connection in the database abstraction.  Important, the golang
// `Connection` is actually a connection pool itself and not a single connection, so it becomes
// necessary to reuse ghe connections so we don't have database trouble.
func Reuse(conn *sql.DB) (db *Database) {
	return &Database{Connection: conn, ReusedConn: true}
}

// Open will open a new database connection to be use in subsequent calls to postgres APIs
func Open(host, port, dbName, user, pw, sslMode string) (db *Database, err error) {
	METHOD := "postgres.Open"

	if host == "" {
		return nil, fmt.Errorf("ERROR (%s): No host provided on input", METHOD)
	}
	if port == "" {
		return nil, fmt.Errorf("ERROR (%s): No port provided on input", METHOD)
	}
	if dbName == "" {
		return nil, fmt.Errorf("ERROR (%s): No database name provided on input", METHOD)
	}
	if user == "" {
		return nil, fmt.Errorf("ERROR (%s): No user provided on input", METHOD)
	}
	if pw == "" {
		return nil, fmt.Errorf("ERROR (%s): No password provided on input", METHOD)
	}

	if sslMode == "" {
		sslMode = "require"
	}

	iPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	connectionMapMutex.Lock()
	defer connectionMapMutex.Unlock()
	key := fmt.Sprintf("%s_%s_%s_%s_%s_%s", host, port, dbName, user, pw, sslMode)
	db = connectionMap[key]

	if db != nil {
		if err := dbIsActive(db.Connection); err != nil {
			log.Printf("ERROR (%s): Database not active: %s", METHOD, err.Error())
			dbDisconnect(db.Connection)
			connectionMap[key] = nil
		} else {
			return db, nil
		}
	}

	sqlDB, err := dbConnect(host, iPort, dbName, user, pw, sslMode)

	if err != nil {
		return nil, err
	}

	// Set the maximum number of open connections that are allowed. If this is not set,
	// potentially every single concurrent API call will use it's own connection (Go
	// database connection pooling does not have a maximum number of database connections
	// by default). With this set, a database call will have to wait for a connection in the
	// pool to become free. See https://www.alexedwards.net/blog/configuring-sqldb for
	// additional information.
	if isDBMaxOpenConnsSet {
		log.Print("Setting maximum number of open database connections to ", dbMaxOpenConns)
		sqlDB.SetMaxOpenConns(dbMaxOpenConns)
	}

	db = &Database{Connection: sqlDB, Port: iPort, Host: host}
	connectionMap[key] = db

	return db, nil
}

// Close will close the database connection that was created with Open
func (db *Database) Close() (err error) {

	/* [2019-01-23] We no long disconnect the connection because it's actuall a pool
	if db != nil && db.Connection != nil {
		dbDisconnect(db.Connection)
	}*/

	return err
}

type allNotificationsCacheType struct {
	lastUpdate     time.Time
	expirationTime time.Time
	list           *api.NotificationList
}

// GetAllNotifications will return notifications from the database
func (db *Database) GetAllNotifications(ctx ctxt.Context, tagsParamValue string, includePnPRemoved bool) (result *api.NotificationList, err error) {

	METHOD := "GetAllNotifications:"

	txn := ctx.NRMon.StartTransaction(exmon.PostgresGetNotifications)
	defer txn.End()
	txn.AddCustomAttribute(exmon.Operation, exmon.OperationQuery)

	limit := -1 // negative should be no limit, usually around 150 entries
	offset := 0

	myQuery := CRN + "=" + GenericCRNMask
	if includePnPRemoved {
		myQuery += "&pnp_removed=true,false"
	}

	nr, count, err, httpStatus := dbGetNotificationByQuery(db.Connection, myQuery, limit, offset)

	if err != nil {
		// We should always get some notifications. If none, then there is a problem.
		log.Printf("INFO: %s recieved error %s", METHOD, err.Error())
		txn.AddBoolCustomAttribute(exmon.PostgresFailed, true)
		return nil, err
	}

	if httpStatus != http.StatusOK {
		err = fmt.Errorf("Bad HTTP status: %d", httpStatus)
		log.Printf("INFO: %s recieved HTTP error %d - trying to continue", METHOD, httpStatus)
	}

	txn.AddBoolCustomAttribute(exmon.PostgresFailed, false)
	txn.AddIntCustomAttribute(exmon.RecordCount, count)

	if nr == nil {
		log.Println("INFO:" + METHOD + " did not get any notifications.")
		return nil, err
	}

	result = convert.PGnrToCnList(*nr, tagsParamValue)

	log.Printf("INFO: %s returning %d notifications (collated from %d)", METHOD, len(result.Items), count)

	if err != nil {
		txn.AddBoolCustomAttribute(exmon.ParseFailed, true)
	} else {
		txn.AddBoolCustomAttribute(exmon.ParseFailed, false)
	}

	return result, err

}

// GetNotificationByRecordID will return a notification given a record ID
func (db *Database) GetNotificationByRecordID(ctx ctxt.Context, recordID string, tagsURLParam string) (result *api.Notification, err error) {

	METHOD := "GetNotificationByRecordID(" + recordID + "):"

	nr, err, httpStatus := dbGetNotificationByRecordID(db.Connection, recordID)

	if nr == nil {
		if httpStatus != http.StatusOK {
			log.Printf("INFO: %s recieved HTTP error %d", METHOD, httpStatus)
			return nil, err
		}
		log.Println("INFO:" + METHOD + " did not get any notifications.")
		return nil, nil
	}

	if tagsURLParam == "" {
		// tags should always be empty when returning JSON with search by Record ID without "tags" search parameter
		nr.Tags = ""
	}

	result = convert.PGnrToCn(nr)
	log.Printf("INFO: %s found notification == %t", METHOD, result != nil)

	// Must populate CRNs since there could be other records for this notification with other CRNs
	if result != nil {
		result, err = db.populateCRN(ctx, result, tagsURLParam)
		if err != nil {
			log.Printf("%s: error trying to collate - %s", METHOD, err.Error())
			return nil, err
		}
	}

	return result, nil
}

// GetNotificationByQuery will return notifications given a query string
func (db *Database) GetNotificationByQuery(ctx ctxt.Context, query string, limit, offset int, tagsParamValue string) (result *api.NotificationList, err error) {

	METHOD := fmt.Sprintf("%s: (query=%s, limit=%d, offset=%d)", "GetNotificationByQuery", query, limit, offset)

	if tagsParamValue != "" {
		log.Println("tags parameter value=", tagsParamValue)
	}

	nrList, count, err, httpStatus := dbGetNotificationByQuery(db.Connection, query, limit, offset)

	if count == 0 {
		if httpStatus != http.StatusOK {
			log.Printf("%s recieved HTTP error %d", METHOD, httpStatus)
			return nil, err
		}
		log.Println(METHOD + " did not get any notifications.")
		return nil, nil
	}

	result = convert.PGnrToCnList(*nrList, tagsParamValue)
	log.Printf("%s returning %d notifications (collated from %d)", METHOD, len(result.Items), count)

	vals, err := url.ParseQuery(query)
	if err == nil && len(vals[CRN]) != 0 && vals[CRN][0] != GenericCRNMask && vals[CRN][0] != "crn:v1::::::::" {
		result, err = db.populateCRNList(ctx, result, "")
		if err != nil {
			log.Printf("%s: error trying to collate - %s", METHOD, err.Error())
			return nil, err
		}
	}

	return result, nil
}

// populateCRNList is used to ensure the notification records have all the CRNs.  This should only be used when the original
// notification query uses a specific CRN mask instead of the general CRN mask that captures all resources.
// The input set is the result set from a query, and this will return that same set, except the CRN list will be filled in.
func (db *Database) populateCRNList(ctx ctxt.Context, inputSet *api.NotificationList, tagsParamValue string) (result *api.NotificationList, err error) {

	fullList, err := db.GetAllNotifications(ctx, tagsParamValue, false)
	if err != nil {
		return nil, err
	}

	lookup := make(map[string]*api.Notification)
	if fullList != nil {
		for _, n := range fullList.Items {
			lookup[n.Source+":"+n.SourceID] = n
		}
	}

	for _, n := range inputSet.Items {
		n.CRNs = make([]string, 0)
	}

	for _, n := range inputSet.Items {
		i := lookup[n.Source+":"+n.SourceID]
		if i != nil {
			n.CRNs = append(n.CRNs, i.CRNs...)
			if tagsParamValue != "" {
				n.Tags = i.Tags
			}
		}

	}

	return inputSet, nil
}

func (db *Database) populateCRN(ctx ctxt.Context, input *api.Notification, tagsParamValue string) (result *api.Notification, err error) {

	inputSet := &api.NotificationList{}
	inputSet.Items = append(inputSet.Items, input)
	resultSet, err := db.populateCRNList(ctx, inputSet, tagsParamValue)
	if err != nil {
		return nil, err
	}

	return resultSet.Items[0], nil
}

// DeleteAllNotifications is used to delete all notification records from the database.
// This is really only to be used for testing when we want to clear the database.
func (db *Database) DeleteAllNotifications(ctx ctxt.Context) (err error) {
	list, err := db.GetAllNotifications(ctx, "", true)
	if err != nil {
		return err
	}

	for _, r := range list.Items {
		err, status := dbDeleteNotification(db.Connection, r.RecordID)
		if err != nil {
			log.Println("ERROR! =", err)
		}
		if status != http.StatusOK {
			log.Println("HTTP Status = ", status)
		}
	}

	return nil
}

func lookupIntEnvVariable(enVariableName string) (int, bool) {
	result := -1
	value, exists := os.LookupEnv(enVariableName)
	if exists {
		var err error
		result, err = strconv.Atoi(value)
		if err != nil {
			log.Print("Environment variable "+enVariableName+" could not be parsed to an int, err=", err.Error())
			exists = false
		}
	}
	return result, exists
}
