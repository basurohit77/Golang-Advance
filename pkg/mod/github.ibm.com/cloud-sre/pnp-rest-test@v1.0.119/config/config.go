package config

import (
	"database/sql"
	"os"
	"time"
)

var (
	Start                 = time.Now()
	PgHost                = os.Getenv("PG_HOST")
	PgDB                  = os.Getenv("PG_DB")
	PgPort                = os.Getenv("PG_PORT")
	PgDBUser              = os.Getenv("PG_DB_USER")
	PgPass                = os.Getenv("PG_DB_PASS")
	DBSSLMode             = os.Getenv("PG_SSLMODE")
	PgSSLRootCertFilePath = os.Getenv("PG_SSLROOTCERTFILEPATH")
	Pdb                   *sql.DB
	Debug                 = false
)
