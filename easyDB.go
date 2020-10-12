package main

import (
	"database/sql"
	"errors"
	"os"
	"reflect"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const dbname = "database.db"

// Create db if not exists
func createDB(initialQuery string) {
	if !fileExists(dbname) {
		file, createError := os.Create(dbname)
		err(createError)
		file.Close()

		// Execute initial query
		execQuery(initialQuery)
	}
}

// Function for INSERT OR CREATE queries
func execQuery(query string, args ...interface{}) {
	db, databaseError := sql.Open("sqlite3", dbname)
	err(databaseError)
	defer db.Close()
	_, execError := db.Exec(query, args...)
	err(execError)
}

// Conditional Insert
func conditionalInsert(condition bool, query string, args ...interface{}) error {
	db, databaseError := sql.Open("sqlite3", dbname)
	err(databaseError)
	defer db.Close()
	tx, txError := db.Begin()
	err(txError)
	_, execError := tx.Exec(query, args...)
	if execError != nil || condition {
		rollbackError := tx.Rollback()
		err(rollbackError)
		if execError != nil {
			return execError
		} else {
			return errors.New("not nil")
		}
	}
	commitError := tx.Commit()
	err(commitError)
	return nil
}

// Function for SELECT queries
func JSONfromDB(query string, args ...interface{}) []interface{} {
	db, databaseError := sql.Open("sqlite3", dbname)
	err(databaseError)
	defer db.Close()

	// Prepare statement
	statement, stmError := db.Prepare(query)
	err(stmError)

	// Get relevant rows by making query
	rows, rowsError := statement.Query(args...)
	err(rowsError)
	defer rows.Close()

	// Get field names from query
	fieldsRaw := regexp.MustCompile(`".*?"`).FindAllString(query, -1)
	var fields []string

	// Trim quotes (need to find solution to not include qoutes using regex)
	for _, x := range fieldsRaw {
		fields = append(fields, strings.Trim(x, `"`))
	}

	// Make range variables with pointer for rows.Scan function
	tmp := make([]interface{}, len(fields))
	dest := make([]interface{}, len(fields))
	for i := range tmp {
		dest[i] = &tmp[i]
	}

	var data []interface{}
	for rows.Next() {
		// Get values from row
		scanError := rows.Scan(dest...)
		err(scanError)

		// Put values to map with string-keys
		m := make(map[string]interface{})
		for i, t := range tmp {
			m[fields[i]] = t
		}
		data = append(data, m)
	}
	return data
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Function for SELECT queries
func JSONfromDB2(model interface{}, query string, args ...interface{}) []interface{} {
	db, databaseError := sql.Open("sqlite3", dbname)
	err(databaseError)
	defer db.Close()

	// Prepare statement
	statement, stmError := db.Prepare(query)
	err(stmError)

	// Get relevant rows by making query
	rows, rowsError := statement.Query(args...)
	err(rowsError)
	defer rows.Close()

	// get fields lenght
	v := reflect.ValueOf(model)
	len := v.NumField()

	// Make range variables with pointer for rows.Scan function
	tmp := make([]interface{}, len)
	dest := make([]interface{}, len)
	for i := range tmp {
		dest[i] = &tmp[i]
	}

	var data []interface{}
	for rows.Next() {
		// Get values from row
		scanError := rows.Scan(dest...)
		err(scanError)

		// Put values to map with string-keys
		row := reflect.New(reflect.TypeOf(model))
		for i, t := range tmp {
			row.Elem().Field(i).Set(reflect.ValueOf(t))
		}
		data = append(data, row.Interface())
	}
	return data
}

func JSONfromDB3(model interface{}, query string, args ...interface{}) []interface{} {
	db, databaseError := sql.Open("sqlite3", dbname)
	err(databaseError)
	defer db.Close()

	// Prepare statement
	statement, stmError := db.Prepare(query)
	err(stmError)

	// Get relevant rows by making query
	rows, rowsError := statement.Query(args...)
	err(rowsError)
	defer rows.Close()

	// get fields lenght
	v := reflect.ValueOf(model).Type().Elem()
	len := v.NumField()

	// Make range variables with pointer for rows.Scan function
	tmp := make([]interface{}, len)
	dest := make([]interface{}, len)
	for i := range tmp {
		dest[i] = &tmp[i]
	}

	var data []interface{}
	for rows.Next() {
		// Get values from row
		scanError := rows.Scan(dest...)
		err(scanError)

		// Put values to map with string-keys
		row := reflect.New(reflect.ValueOf(model).Type().Elem())
		for i, t := range tmp {
			row.Elem().Field(i).Set(reflect.ValueOf(t))
		}
		data = append(data, row.Interface())
	}
	return data
}
