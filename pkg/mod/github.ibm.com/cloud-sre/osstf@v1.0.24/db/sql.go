package db

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.ibm.com/cloud-sre/osstf/lg"
)

// var openRowsCount = 0
// var closeRowsCount = 0

var pgVariables = regexp.MustCompile(`\?`)

func convertToPostgres(query string) string {
	idx := 0
	return pgVariables.ReplaceAllStringFunc(query, func(inp string) string {
		idx += 1
		return fmt.Sprintf("$%d", idx)
	})
}

func dbQueryRows(database *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	rows, error := database.Query(convertToPostgres(query), args...)
	// openRowsCount++
	// lg.OpenRows(openDbCount, openRowsCount)
	return rows, error
}

func dbExec(database *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	res, err := database.Exec(convertToPostgres(query), args...)
	return res, err
}

func dbCloseRows(rows *sql.Rows) {
	if rows != nil {
		_ = rows.Close()
		// closeRowsCount++
		// lg.CloseRows(openDbCount, closeRowsCount)
	}
}

var isAlnum = regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString

func Alnum(checkme string) string {
	checkme = strings.TrimSpace(checkme)
	if !isAlnum(checkme) {
		panic("Unexpected column or table name: " + checkme)
	}
	return checkme
}

func string2interface(argv []string) []interface{} {
	ret := make([]interface{}, len(argv))
	for i, v := range argv {
		ret[i] = v
	}
	return ret
}

// Execute multiple statements, abort as soon as an error occurs
// DEPRECATED, DO NOT USE, DELETE WHEN NO LONGER IN USE, USE ExecuteQueriesWithArgs INSTEAD
func ExecuteQueries(database *sql.DB, statements []string) {
	// DEPRECATED, DO NOT USE, DELETE WHEN NO LONGER IN USE, USE ExecuteQueriesWithArgs INSTEAD
	// execute queries
	for index := range statements {
		result, err := database.Exec(statements[index])
		lg.SqlExecuteStatement(statements[index], result, err)

		if err != nil {
			// failed to execute one of several SQL statements. Exit
			log.Fatal(statements[index], " ", err)
		}
	}
}

// Execute multiple statements, abort as soon as an error occurs
func ExecuteQueriesWithArgs(database *sql.DB, statements []string, args [][]interface{}) {
	// execute queries
	for index := range statements {
		result, err := dbExec(database, statements[index], args[index]...)
		lg.SqlExecuteStatement(statements[index], result, err)

		if err != nil {
			// failed to execute one of several SQL statements. Exit
			log.Fatal(statements[index], " ", err)
		}
	}
}

// Insert into table, log error without terminating the process
func Insert(database *sql.DB, table string, columns []string, values []string) {

	// compose the query
	var query bytes.Buffer
	query.WriteString("INSERT INTO ")
	query.WriteString(Alnum(table))
	query.WriteString(" ( ")
	last := len(columns) - 1
	for index := range columns {
		query.WriteString(Alnum(columns[index]))
		if index < last {
			query.WriteString(" , ")
		} else {
			query.WriteString(" ) VALUES ( ")
		}
	}
	last = len(values) - 1
	for index := range values {
		query.WriteString("?")
		if index < last {
			query.WriteString(" , ")
		} else {
			query.WriteString(" ) ")
		}
	}

	// execute
	queryStr := query.String()
	result, err := dbExec(database, queryStr, string2interface(values)...)
	lg.SqlExecuteTable("INSERT", table, queryStr, result, err)

}

// Update a row, log error without terminating the process. All columns being updated are strings
func Update(database *sql.DB, table string, keyColumn string, keyValue string, columns []string, values []string) {
	if len(columns) != len(values) {
		panic("Unmatching col val pairs, check lengths")
	}
	// compose the query
	var query bytes.Buffer
	query.WriteString("UPDATE ")
	query.WriteString(Alnum(table))
	query.WriteString(" SET ")
	last := len(columns) - 1
	for index := range columns {
		query.WriteString(Alnum(columns[index]))
		query.WriteString(" = ")
		query.WriteString("?")
		if index < last {
			query.WriteString(" , ")
		} else {
			query.WriteString(" ")
		}
	}
	query.WriteString(" WHERE ")
	query.WriteString(Alnum(keyColumn))
	query.WriteString(" = ?")
	values = append(values, keyValue)

	// execute
	queryStr := query.String()
	result, err := dbExec(database, queryStr, string2interface(values)...)
	lg.SqlExecuteTable("UPDATE", table, queryStr, result, err)

}

// Update an integer column in a row, log error without terminating the process
func UpdateInt(database *sql.DB, table string, keyColumn string, keyValue string, column string, value int) {

	// compose the query
	var query bytes.Buffer
	query.WriteString("UPDATE ")
	query.WriteString(Alnum(table))
	query.WriteString(" SET ")
	query.WriteString(Alnum(column))
	query.WriteString(" = ")
	query.WriteString(strconv.Itoa(value))
	query.WriteString(" WHERE ")
	query.WriteString(Alnum(keyColumn))
	query.WriteString(" = ?")

	// execute
	queryStr := query.String()
	result, err := dbExec(database, queryStr, keyValue)
	lg.SqlExecuteTable("UPDATE", table, queryStr, result, err)

}

// Delete a row, log error without terminating the process
func Delete(database *sql.DB, table string, keyColumn string, keyValue string) {

	// compose the query
	var query bytes.Buffer
	query.WriteString("DELETE FROM ")
	query.WriteString(Alnum(table))
	query.WriteString(" WHERE ")
	query.WriteString(Alnum(keyColumn))
	query.WriteString(" = ?")

	// execute
	queryStr := query.String()
	result, err := dbExec(database, queryStr, keyValue)
	lg.SqlExecuteTable("DELETE", table, queryStr, result, err)

}

// Execute a SQL select to get one column in one row
// If Select fails, log SQL error without terminating the process, return nil
func SelectColumnWithPrimaryKey(database *sql.DB, table string, keyColumn string, keyValue string, columnToSelect string) interface{} {

	// the select statement
	query := "SELECT " + Alnum(columnToSelect) + " FROM " + Alnum(table) + " WHERE " + Alnum(keyColumn) + " = ?"

	// execute select
	rows, err := dbQueryRows(database, query, keyValue)
	defer dbCloseRows(rows)
	lg.SqlQuery(table, query, err)

	if err != nil {
		return nil
	}

	if rows.Next() {
		var content interface{}
		contentPtr := &content
		err = rows.Scan(contentPtr)
		if err != nil {
			lg.SqlError("sql.Rows.Scan()", err)
			return nil
		}
		return *contentPtr
	}

	return nil
}

// Execute a SQL select by the primary key to select all columns of a single row
// Return a map in which the keys are column names and values are column contents
// If Select fails, log SQL error without terminating the process, return empty result set
func SelectByPrimaryKeyStatus(database *sql.DB, table string, keyColumn string, keyValue string, statusColumn string, status int) map[string]string {

	// the select statement
	query := "SELECT * FROM " + Alnum(table) + " WHERE " + Alnum(keyColumn) + " = ? AND " + Alnum(statusColumn) + " = " + strconv.Itoa(status)

	return selectByPrimaryKeyInternal(database, query, table, keyValue)
}

// Execute a SQL select by the primary key to select all columns of a single row
// Return a map in which the keys are column names and values are column contents
// If Select fails, log SQL error without terminating the process, return empty result set
func SelectByPrimaryKey(database *sql.DB, table string, keyColumn string, keyValue string) map[string]string {

	// the select statement
	query := "SELECT * FROM " + Alnum(table) + " WHERE " + Alnum(keyColumn) + " = ?"

	return selectByPrimaryKeyInternal(database, query, table, keyValue)
}

func selectByPrimaryKeyInternal(database *sql.DB, query string, table string, args ...interface{}) map[string]string {
	result := make(map[string]string)

	// execute select
	rows, err := dbQueryRows(database, query, args...)
	defer dbCloseRows(rows)
	lg.SqlQuery(table, query, err)
	if err != nil {
		return result
	}

	// get table columns
	columns, err := rows.Columns()
	if err != nil {
		lg.SqlError("sql.Rows.Columns()", err)
		return result
	}

	if rows.Next() {
		// create pointers to hold column values
		var contentPtrs = make([]interface{}, len(columns))
		for c := range contentPtrs {
			var content string
			contentPtrs[c] = &content
		}

		// point pointers to column contents
		err = rows.Scan(contentPtrs...)
		if err != nil {
			lg.SqlError("sql.Rows.Scan()", err)
			return result
		}

		// go through the pointers to get column content
		for i, c := range columns {
			// convert the content into string
			content := *(contentPtrs[i]).(*string)
			// put the content in the map
			result[c] = content
		}
	}

	return result
}

// Execute a SQL select query to retrive a column in all rows
// Return a slice of strings
// If Select fails, log SQL error without terminating the process, return empty slice
func SelectColumn(database *sql.DB, table string, column string) []string {

	empty := []string{}

	// the select statement
	query := "SELECT " + Alnum(column) + " FROM " + Alnum(table)

	// execute select
	rows, err := dbQueryRows(database, query)
	defer dbCloseRows(rows)
	lg.SqlQuery(table, query, err)
	if err != nil {
		log.Printf("Error occurred executing query '%s', error = %s", query, err.Error())
		return empty
	}

	var result = make([]string, 0)
	for rows.Next() {
		// get column content
		var content interface{}
		err = rows.Scan(&content)
		if err != nil {
			lg.SqlError("sql.Rows.Scan()", err)
			return result
		}

		contentStr := content.(string)
		result = append(result, contentStr)
	}

	return result
}

// Execute a SQL select query to select all rows and all columns
// Return a slice of maps where the keys of the maps are column names and values of the maps are column contents
// If Select fails, log SQL error without terminating the process, return empty result set
func SelectAll(database *sql.DB, table string) []map[string]string {

	query := "SELECT * FROM " + Alnum(table)
	return selectQuery(database, query, table)

}

//// Execute a SQL select query to select all rows and all columns
//// Return a slice of maps where the keys of the maps are column names and values of the maps are column contents
//// Only return rows status matching the given value
//// If Select fails, log SQL error without terminating the process, return empty result set
//func SelectAllWithStatus(database *sql.DB, table string, statusColumn string, status int) []map[string]string {
//
//	query := "SELECT * FROM " + table  + " WHERE " + statusColumn + " = " + strconv.Itoa(status)
//	return selectQuery(database, query, table)
//
//}

//// Execute a SQL select query to select all rows and all columns
//// Return a slice of maps where the keys of the maps are column names and values of the maps are column contents
//// Only return rows status different from the given value
//// If Select fails, log SQL error without terminating the process, return empty result set
//func SelectAllWithStatusNot(database *sql.DB, table string, statusColumn string, status int) []map[string]string {
//
//	query := "SELECT * FROM " + table  + " WHERE " + statusColumn + " <> " + strconv.Itoa(status)
//	return selectQuery(database, query, table)
//
//}

// Execute a SQL select query to select all columns and return the rows with matching column values
// The given column is not the primary key, therefore multiple rows may be returned
// Return a slice of maps where the keys of the maps are column names and values of the maps are column contents
// If Select fails, log SQL error without terminating the process, return empty result set
func SelectByColumnValue(database *sql.DB, table string, column string, columnValue string) []map[string]string {

	query := "SELECT * FROM " + Alnum(table) + " WHERE " + Alnum(column) + " = ?"
	return selectQuery(database, query, table, columnValue)

}

//// Execute a SQL select query to select all columns and return the rows with matching column values
//// The given column is not the primary key, therefore multiple rows may be returned
//// Return a slice of maps where the keys of the maps are column names and values of the maps are column contents
//// Only return rows status matcing the given value
//// If Select fails, log SQL error without terminating the process, return empty result set
//func SelectByColumnValue(database *sql.DB, table string, column string, columnValue string, statusColumn string, status int) []map[string]string {
//
//	query := "SELECT * FROM " + table + " WHERE " + column + " = '" + columnValue + "' AND " + statusColumn + " = " + strconv.Itoa(status)
//	return selectQuery(database, query, table)
//
//}

// Execute a SQL select query to select all columns and return the rows with the column containing the specified value
// Return a slice of maps where the keys of the maps are column names and values of the maps are column contents
// If Select fails, log SQL error without terminating the process, return empty result set
func SelectColumnValuesIn(database *sql.DB, table string, column string, columnValues []string) []map[string]string {

	values := strings.Join(strings.Split(strings.Repeat("?", len(columnValues)), ""), ",")
	query := "SELECT * FROM " + Alnum(table) + " WHERE " + Alnum(column) + " IN (" + values + ")"
	return selectQuery(database, query, table, string2interface(columnValues)...)

}

//// Execute a SQL select query to select all columns and return the rows with the column containing the specified value
//// Return a slice of maps where the keys of the maps are column names and values of the maps are column contents
//// Only return rows status matcing the given value
//// If Select fails, log SQL error without terminating the process, return empty result set
//func SelectColumnValuesIn(database *sql.DB, table string, column string, columnValues []string, statusColumn string, status int) []map[string]string {
//
//	values := strings.Join(columnValues, "', '")
//	query := "SELECT * FROM " + table + " WHERE " + column + " IN ('" + values + "')" + " AND " + statusColumn + " = " + strconv.Itoa(status)
//	return selectQuery(database, query, table)
//
//}

func selectQuery(database *sql.DB, query string, table string, args ...interface{}) []map[string]string {
	// this is what's returned in case Select or subsequent scan fails
	empty := []map[string]string{}

	// execute select
	rows, err := dbQueryRows(database, query, args...)
	defer dbCloseRows(rows)
	lg.SqlQuery(table, query, err)
	if err != nil {
		return empty
	}

	// get table columns
	columns, err := rows.Columns()
	if err != nil {
		lg.SqlError("sql.Rows.Columns()", err)
		return empty
	}

	// create pointers to hold column values
	var contentPtrs = make([]interface{}, len(columns))
	for c := range contentPtrs {
		var content string
		contentPtrs[c] = &content
	}

	// make zero size slice to be appended to
	var result = make([](map[string]string), 0)

	// scan rows
	for rows.Next() {

		// create a map to hold column-value pairs within a row
		rowMap := make(map[string]string)

		// point pointers to column contents
		err := rows.Scan(contentPtrs...)
		if err != nil {
			lg.SqlError("sql.Rows.Scan()", err)
			return result
		}

		// go through the pointers to get column content
		for i, c := range columns {

			// convert the content into string
			content := *(contentPtrs[i]).(*string)
			// put the content in the map
			rowMap[c] = content
		}

		if len(rowMap) > 0 {
			result = append(result, rowMap)
		}
	}

	return result
}
