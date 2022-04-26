package testutils

import (
	"database/sql"
	"github.ibm.com/cloud-sre/pnp-nq2ds/monitor"
	"log"
	"os"
	"strconv"
	"testing"

	"github.ibm.com/cloud-sre/oss-globals/tlog"

	instana "github.com/instana/go-sensor"
	newrelic "github.com/newrelic/go-agent"
	"github.ibm.com/cloud-sre/ossmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
	"github.ibm.com/cloud-sre/pnp-db/dbcreate"
)

var (
	Source        = "servicenow"
	SourceID      = "123456"
	RecreateTable = false
)

func PrepareTestInc(t *testing.T, msg string) (*sql.DB, []byte, ossmon.OSSMon) {
	log.Println(tlog.Log())
	var (
		err      error
		pgHost   = os.Getenv("PG_DB_IP")
		pgDB     = os.Getenv("PG_DB")
		pgDBUser = os.Getenv("PG_DB_USER")
		pgPass   = os.Getenv("PG_DB_PASS")
		pgPort   = os.Getenv("PG_DB_PORT")
		encMsg   = []byte("")
	)
	pgPortInt, _ := strconv.Atoi(pgPort)
	t.Logf(tlog.Log() + "Database Connection Info: " + pgHost + " " + pgPort + " " + pgDB + " " + pgDBUser + " **** " + "disable")
	dbConnection, err := db.Connect(pgHost, pgPortInt, pgDB, pgDBUser, pgPass, "disable")
	if err != nil {
		t.Errorf(tlog.Log()+"Failed to connect to DB, %d", err)
	}
	if RecreateTable {
		EnsureTablesExist(dbConnection)
	}
	if msg != "" {
		encMsg, err = encryption.Encrypt(msg) // Encryption
		if err != nil {
			log.Println(tlog.Log(), err)
		}
	}
	nrConfig := newrelic.NewConfig(monitor.SrvPrfx+"testing", "0000000000000000000000000000000000000000")
	nrConfig.Enabled = false
	nrApp, err := newrelic.NewApplication(nrConfig)
	mon := ossmon.OSSMon{
		NewRelicApp: nrApp,
		Sensor:      instana.NewSensor(monitor.SrvPrfx + "testing"),
	}
	if err != nil {
		t.Fatal(err)
	}
	osscatalog.CatalogCheckBypass = true // Bypass checks to oss catalog for pnpenabled
	return dbConnection, encMsg, mon
}

func EnsureTablesExist(database *sql.DB) {
	log.Println(tlog.Log() + "Creating User-defined types, tables and indexes.")
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
	RecreateTable = false
}
