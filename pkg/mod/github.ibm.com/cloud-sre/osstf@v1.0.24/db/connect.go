package db

import (
	"database/sql"
	"errors"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // a pure Go Postgres driver for the database/sql package
	"github.ibm.com/cloud-sre/osstf/lg"
)

const (
	maxIdleCount int           = 20               // zero means defaultMaxIdleConns; negative means 0
	maxIdleTime  time.Duration = 20 * time.Second // maximum amount of time a connection may be idle before being closed  20 seconds
)

// The information needed to connect to a database
var DbConnectionInfo string

// Sensor needed to monitor database activity (optional)
var Sensor *instana.Sensor

var openDbCount = 0
var closeDbCount = 0

// connect to database to be ready for queries
// caller of this method must do Disconnect()
func Connect() (*sql.DB, error) {

	var err error

	// attempt to connect n time
	n := 3
	for i := 0; i < n; i++ {
		database, err := tryConnect(DbConnectionInfo)
		if err == nil {
			return database, err
		} else if i < (n - 1) {
			// sleep a bit, will retry
			time.Sleep(time.Second) // one second
		}
	}

	// do not terminate, let caller deal with it
	if err == nil {
		// sql.Db didn't return an error. Make sure err!=nil when database==nil
		err = errors.New("failed to Open, Ping, Begin() the database. No error returned")
	}
	return nil, err
}

func tryConnect(dbinfo string) (*sql.DB, error) {
	// connect to database
	var database *sql.DB
	var err error
	if Sensor == nil {
		// Sensor not available, use regular sql Open function:
		database, err = sql.Open("postgres", dbinfo)
	} else {
		// Sensor is avaialble, wrap the driver and use the instana SQLOpen function:
		instana.InstrumentSQLDriver(Sensor, "postgres", &pq.Driver{})
		database, err = instana.SQLOpen("postgres", dbinfo)
	}
	if err != nil {
		lg.SqlError("sql.Open()", err)
		return nil, err
	}

	database.SetMaxIdleConns(maxIdleCount)   // Keep only the number of idle connections as max set at maxIdleCount
	database.SetConnMaxIdleTime(maxIdleTime) // Keep an idle connections for the maximum amount of time set at maxIdleTime

	// We just did an Open(), a Ping() may not be necessary
	//	// confirm connection is alive
	//	if err := database.Ping(); err != nil {
	//		database.Close()
	//		lg.SqlError("sql.DB.Ping()", err)
	//		return nil, err
	//	}

	// Not sure if Begin() is required.
	// It seems to leave behind an idle transaction as seem from "select * from pg_stat_activity;"
	//	// get ready to begin transactions
	//	_, err = database.Begin()
	//	if err != nil {
	//		database.Close()
	//		lg.SqlError("sql.DB.Begin()", err)
	//		return nil, err
	//	}

	openDbCount++
	lg.OpenDatabse(openDbCount)

	return database, err
}

func Disconnect(database *sql.DB) {
	if database != nil {
		_ = database.Close()
		closeDbCount++
		lg.CloseDatabase(closeDbCount)
		// openRowsCount = 0
		// closeRowsCount = 0
	}
}

// check if the database connection is still active and usable
func IsActive(database *sql.DB) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = errors.New("an error occurred when executing the SQL query 'SELECT WHERE 1=0'")
		}
	}()

	if database != nil {
		// confirm connection is alive
		if err = database.Ping(); err != nil {
			return err
		}
		// per https://stackoverflow.com/questions/41618428/golang-ping-succeed-the-second-time-even-if-database-is-down
		// Ping() is not reliable
		// we need to execute a query to be sure
		_, err = database.Exec("SELECT WHERE 1=0")
	}

	return err
}
