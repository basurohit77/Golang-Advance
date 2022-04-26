package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"strings"
)

//CreateTable for Postgresql, if table is not found, create it
func CreateTable(tableName string, tableCreationStatement string, db *sql.DB) error {

	s := `select tablename from pg_tables where "tablename"=$1`
	var name string
	err := db.QueryRow(s, tableName).Scan(&name)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set"){
		return err
	}

	if name == "" {
		log.Println(tableName, " is not existed, create it")
		_, err = db.Exec(tableCreationStatement)
		if err != nil {
			return err
		}
	}
	return nil
}
