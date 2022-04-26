package testutils

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-db/dbcreate"
)

var RECREATE_TABLE = true

func SetupDB(recreate bool) *sql.DB {
	RECREATE_TABLE = recreate
	var err error
	var (
		pgHost   = os.Getenv("PG_DB_IP")
		pgDB     = os.Getenv("PG_DB")
		pgDBUser = os.Getenv("PG_DB_USER")
		pgPass   = os.Getenv("PG_DB_PASS")
		pgPort   = os.Getenv("PG_DB_PORT")
	)
	pgPortInt, _ := strconv.Atoi(pgPort)

	log.Println("Database Connection Info: " + pgHost + " " + pgPort + " " + pgDB + " " + pgDBUser + " **** " + "disable")

	dbConnection, err := db.Connect(pgHost, pgPortInt, pgDB, pgDBUser, pgPass, "disable")
	// dbConnection, err := sql.Open("postgres", "user=postgres password= dbname=postgres sslmode=disable")

	if err != nil {
		log.Print(err)
	}
	// defer db.Disconnect(dbConnection)

	EnsureTablesExist(dbConnection)

	return dbConnection

}

func ConnectDB() *sql.DB {

	var err error
	var (
		pgHost   = os.Getenv("PG_DB_IP")
		pgDB     = os.Getenv("PG_DB")
		pgDBUser = os.Getenv("PG_DB_USER")
		pgPass   = os.Getenv("PG_DB_PASS")
		pgPort   = os.Getenv("PG_DB_PORT")
	)
	pgPortInt, _ := strconv.Atoi(pgPort)

	log.Println("Database Connection Info: " + pgHost + " " + pgPort + " " + pgDB + " " + pgDBUser + " **** " + "disable")

	dbConnection, err := db.Connect(pgHost, pgPortInt, pgDB, pgDBUser, pgPass, "disable")

	if err != nil {
		log.Print(err)
	}

	return dbConnection

}

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
}
