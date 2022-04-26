package chkdbcon

import (
	"database/sql"
	"log"

	"github.ibm.com/cloud-sre/oss-globals/tlog"
)

// CheckDBConnection A simple sql statement 'SELECT 1' to verrify the database connection is valid
func CheckDBConnection(database *sql.DB) error {

	log.Println(tlog.Log() + "Testing database connection...")
	_, err := database.Exec("Select 1")
	if err != nil {
		log.Println(tlog.Log() + "Database connection test failed" + err.Error())
		return err
	}
	log.Println(tlog.Log() + "Database connection passed")
	return nil
}
