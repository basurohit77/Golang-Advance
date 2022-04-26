package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.ibm.com/cloud-sre/oss-globals/tlog"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	"github.ibm.com/cloud-sre/pnp-db/dbcreate"
	"github.ibm.com/cloud-sre/pnp-db/test"
)

const (
	//RECREATE_TABLE flag to set create o not a new DB
	RECREATE_TABLE = true
)

// DbConfiguration JSON stucture to upload dabatase connection info from a config file
type DbConfiguration struct {
	Host                string
	Port                string
	Database            string
	User                string
	Password            string
	SslMode             string
	SslRootCertFilePath string
}

var (
	gaasServices []string
	pgHost                  = os.Getenv("PG_HOST")
	pgDB                    = os.Getenv("PG_DB")
	pgPort                  = os.Getenv("PG_PORT")
)

func main() {
	const FCT = "pnp-db main: "

	t := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// Write startup info to console (stdout)
	log.Println(tlog.Log()+"*** Start Creating PNP-DB...", t)

	// find full path of the current executable including the file name
	ex, e := os.Executable()
	if e != nil {
		log.Println(tlog.Log() + e.Error())
		os.Exit(1)
	}

	// returns the path to the .dbconfig.json file
	pathname := filepath.Dir(ex) + "/" + dbcreate.DATABASE_CONFIG_FILE

	// read the configure file into bytes
	log.Println(tlog.Log()+"Reading ", pathname, " file...")
	configJSON, err := ioutil.ReadFile(pathname) // #nosec G304
	var Config *DbConfiguration
	err = json.Unmarshal(configJSON, &Config)
	if err != nil {
		log.Fatal("Fail to unmarshal database configuration. ", err)
	}

	port, err := strconv.Atoi(pgPort)
	if err != nil {
		log.Println(err.Error())
		return
	}
	var database *sql.DB

	log.Println(tlog.Log()+"Database Connection Info:"+pgHost+":"+pgPort+" database:"+pgDB+" with user: "+Config.User, " SSL: "+Config.SslMode)

	database, _ = db.ConnectWithSSL(pgHost, port, pgDB, Config.User, Config.Password, Config.SslMode, Config.SslRootCertFilePath)

	// connect to the Failover database`
	//	database, _ := db.Connect("pg.oss.cloud.ibm.com", 5432, Config.Database, "postgres", Config.Password, Config.SslMode)

	if database == nil {
		log.Println(tlog.Log() + "pnp-db is unable to connect to database.")
		os.Exit(-1)
	}

	log.Println(tlog.Log()+"Successfully connected to Postgres server:"+pgHost+":"+pgPort+" database:"+pgDB+" with user: "+Config.User, " SSL: "+Config.SslMode)

	// Create tables for PNP if they are not already exist
	// ** NOTE** : EnsureTablesExist(database) WILL RECREATE ALL DATABASE TABLES

	//EnsureTablesExist(database)

	// These test cases will be run on real postgres database
	//RunTests(database)

	// This for testing Postgres Failover, the whole testing last 3-4 minutes
	//RunTestsForPostgresFailover(database)

	db.Disconnect(database)

	log.Println(tlog.Log() + "pnp-db stopped")
	time.Sleep(5 * time.Second)
	os.Exit(0)
}

// RunTests check if the database is ready
func RunTests(database *sql.DB) {
	failed := 0

	cleanupFailed := 0
	cleanupFailed = test.TestCleanup(database)
	for cleanupFailed > 0 {
		cleanupFailed = test.TestCleanup(database)
	}

	// Initialize and get OSSCatalog
	osscatalog.InitializeCatalogFromScratch()

	// gaasServices must contain the Gaas services retrieved from OSSCatalog, otherwise test cases will fail
	// To find out what are Gaas services in OSSCatalog, set "debugCatalog=true" in pnp-abstraction/osscatalog/ossrecord.go,
	// then deploy api-pnp-db in dev environment, this will dump all the osscatalog to the logs, and test cases will fail.
	// In the log, look for services with "EntryType==GAAS", then replace 7 names of Gaas services in the gaasServices array below.
	// Then redeploy api-pnp-db, test cases should all pass.
	// For example, gaasServices = []string{"asp-net-core", "liberty-for-java", "runtime-for-swift", "sdk-for-node-js", "bluemix-cloudfoundry", "bluemix-developer-experience", "bluemix-login-server"}
	gaasServices = []string{"service1", "service2", "service3", "service4", "service5", "service6", "service7"}

	// Test IBM Public Cloud resources
	failed += test.TestCase(database)
	failed += test.TestResource(database)
	failed += test.TestIncident(database)
	failed += test.TestMaintenance(database)
	failed += test.TestSubscriptionWatch(database)
	failed += test.TestNotification(database)

	// Test Gaas resources
	failed += test.TestResourceForGaas(database, gaasServices)
	failed += test.TestIncidentForGaas(database, gaasServices)
	failed += test.TestMaintenanceForGaas(database, gaasServices)

	if failed > 0 {
		log.Println(tlog.Log()+"ALL TEST RESULT: FAILED, number of failures: ", failed)
	} else {
		log.Println(tlog.Log() + "ALL TEST RESULT: SUCCESSFUL")
	}

	cleanupFailed = 0
	cleanupFailed = test.TestCleanup(database)
	for cleanupFailed > 0 {
		cleanupFailed = test.TestCleanup(database)
	}
}

func RunTestsForPostgresFailover(database *sql.DB) {
	totalFailed := 0

	for i := 0; i < 10; i++ {
		log.Printf(tlog.Log()+"Round %d\n", i)
		failed := 0

		cleanupFailed := 0
		cleanupFailed = test.TestCleanup(database)
		for cleanupFailed > 0 {
			cleanupFailed = test.TestCleanup(database)
		}

		failed += test.TestCase(database)
		failed += test.TestResource(database)
		failed += test.TestIncident(database)
		failed += test.TestMaintenance(database)
		failed += test.TestSubscriptionWatch(database)
		failed += test.TestNotification(database)

		if failed > 0 {
			log.Println(tlog.Log()+"TEST RESULT: FAILED, number of failures: ", failed)
		} else {
			log.Println(tlog.Log() + "TEST RESULT: SUCCESSFUL")
		}
		totalFailed += failed
	}

	cleanupFailed := 0
	cleanupFailed = test.TestCleanup(database)
	for cleanupFailed > 0 {
		cleanupFailed = test.TestCleanup(database)
	}

	if totalFailed > 0 {
		log.Println(tlog.Log()+"ALL TEST RESULT: FAILED, total number of failures: ", totalFailed)
	} else {
		log.Println(tlog.Log() + "ALL TEST RESULT: SUCCESSFUL")
	}

}

//EnsureTablesExist WILL RECREATE ALL DATABASE TABLES Create tables for PNP if they are not already exist
func EnsureTablesExist(database *sql.DB) {
	log.Println(tlog.Log() + "Creating User-defined types, tables and indexes.")

	if RECREATE_TABLE {
		// Tables have to be in a certain order. All the dependencies have to be dropped first
		/*	dbcreate.DropIncidentJunctionTable(database)
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
		*/
	}

	// All tables being referenced as foreign keys have to be created first
	/*	dbcreate.CreateResourceTable(database)
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
	*/
}

/*
var pgVariables = regexp.MustCompile(`\?`)
func convertToPostgres(query string) string {
	idx := 0
	return pgVariables.ReplaceAllStringFunc(query, func(inp string) string {
		idx += 1
		return fmt.Sprintf("$%d", idx)
	})
}
func string2interface(argv []string) []interface{}{
	ret := make([]interface{}, len(argv))

	for i, v := range argv {
		ret[i] = v
	}
	ret[0], _ = uuid.NewV4() // generate uuid for record id

	return ret
}

//func string2interface(argv []string) []interface{}{
//	ret := make([]interface{}, len(argv))
//	for i, v := range argv {
//		ret[i] = v
//	}
//	return ret
//}
*/
func dbExec(database *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	res, err := database.Exec(query, args...)
	return res, err
}
