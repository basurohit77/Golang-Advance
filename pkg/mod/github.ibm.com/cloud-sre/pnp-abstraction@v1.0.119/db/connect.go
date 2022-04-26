package db

import (
	"database/sql"

	tfdb "github.ibm.com/cloud-sre/osstf/db"
)

// Connect to database to be ready for queries
// caller of this method must do Disconnect()
func Connect(host string, port int, databaseName string, user string, password string, sslmode string) (database *sql.DB, err error) {
	tfdb.DbConnectionInfo = tfdb.GetDatabaseConfiguration3(host, port, databaseName, user, password, sslmode)

	// connect to the database`
	return tfdb.Connect()
}

// ConnectWithSSL - Must provide file path of SSLROOTCERT certificate and SSL mode e.g. "verify-full" or "verify-ca"
func ConnectWithSSL(host string, port int, databaseName string, user string, password string, sslMode string, sslRootCertFilePath string) (database *sql.DB, err error) {
	tfdb.DbConnectionInfo = tfdb.GetDatabaseConfigurationWithSSLRootCert(host, port, databaseName, user, password, sslMode, sslRootCertFilePath)

	// connect to the database`
	return tfdb.Connect()
}

// Disconnect from the database
func Disconnect(database *sql.DB) {
	tfdb.Disconnect(database)
}

// IsActive check if the database connection is still active and usable
func IsActive(database *sql.DB) (err error) {
	return tfdb.IsActive(database)
}
