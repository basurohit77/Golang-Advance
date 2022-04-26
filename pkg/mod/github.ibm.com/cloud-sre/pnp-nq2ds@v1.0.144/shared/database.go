package shared

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.ibm.com/cloud-sre/oss-globals/chkdbcon"
	"github.ibm.com/cloud-sre/oss-globals/tlog"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
)

var (
	pgHost                              = os.Getenv("PG_HOST")
	pgDB                                = os.Getenv("PG_DB")
	pgPort                              = os.Getenv("PG_PORT")
	pgDBUser                            = os.Getenv("PG_DB_USER")
	pgPass                              = os.Getenv("PG_DB_PASS")
	dbSSLMode                           = os.Getenv("PG_SSLMODE")
	pgSSLRootCertFilePath               = os.Getenv("PG_SSLROOTCERTFILEPATH")
	dbMaxOpenConns, isDBMaxOpenConnsSet = lookupIntEnvVariable("DB_MAX_OPEN_CONNS") // optional parameter. default value is no maximum
	// DBConn stores the database connection, and it is used by the all queries to improve performance
	DBConn *sql.DB

	//BypassLocalStorage is used to avoid storing SN data in Postgres. FedRAMP requirement.
	BypassLocalStorage = os.Getenv("BYPASS_LOCAL_STORAGE") == "true"
)

// ConnectDatabase connects to a PostgreSQL database
func ConnectDatabase() error {
	port, err := strconv.Atoi(pgPort)
	if err != nil {
		log.Println("Make sure the follow parameter had been set DB_PORT,DB_IP,DB_DBNAME, DB_USERNAME , DB_PASSWORD at database.go", err.Error())
		return err
	}

	var database *sql.DB

	log.Println(tlog.Log() + "Database Connection Info: " + pgHost + ":" + pgPort + " database:" + pgDB + " with user: " + pgDBUser + " SSL: " + dbSSLMode)

	database, err = db.ConnectWithSSL(pgHost, port, pgDB, pgDBUser, pgPass, dbSSLMode, pgSSLRootCertFilePath)

	if err != nil {
		log.Println(err.Error())
		return err
	}
	// Check that the database parameters passed are valid and database connection is valid
	err = chkdbcon.CheckDBConnection(database)
	if err != nil {
		log.Println(tlog.Log() + err.Error())
		return err
	}
	log.Println(tlog.Log() + "Successfully connected to Postgres server:" + pgHost + ":" + pgPort)

	// Set the maximum number of open connections that are allowed. If this is not set,
	// potentially every single concurrent API call will use it's own connection (Go
	// database connection pooling does not have a maximum number of database connections
	// by default). With this set, a database call will have to wait for a connection in the
	// pool to become free. See https://www.alexedwards.net/blog/configuring-sqldb for
	// additional information.
	if isDBMaxOpenConnsSet {
		log.Print("Setting maximum number of open database connections to ", dbMaxOpenConns)
		database.SetMaxOpenConns(dbMaxOpenConns)
	}

	DBConn = database
	return err
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
