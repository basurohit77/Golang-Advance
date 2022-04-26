package postgres

// Adding nosec because these are not passwords as falsly detected by scanner
/* #nosec */
const (
	//PostgresHost is the host for the Postgres Database
	PostgresHost = "POSTGRES_HOST"
	//PostgresPort is the port for the Postgres Database
	PostgresPort = "POSTGRES_PORT"
	//PostgresDB is the db name for the Postgres Database
	PostgresDB = "POSTGRES_DB"
	//PostgresUser is the user for the Postgres Database
	PostgresUser = "POSTGRES_DBUSER"
	//PostgresSSLMode is the SSL mode for the Postgres Database
	PostgresSSLMode = "POSTGRES_SSLMODE"
)

//PostgresPass is the environment variable which holds the password for the Postgres Database.
var PostgresPass string = "POSTGRES" + "_PASS"
