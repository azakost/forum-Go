package main

import (
	"database/sql"
	"errors"
	"os"
	"reflect"

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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func sliceFromDB(model interface{}, query string, args ...interface{}) {
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

	container := reflect.Indirect(reflect.ValueOf(model))
	v := container.Type().Elem()
	len := v.NumField()

	// Make range variables with pointer for rows.Scan function
	tmp := make([]interface{}, len)
	dest := make([]interface{}, len)
	for i := range tmp {
		dest[i] = &tmp[i]
	}

	for rows.Next() {
		// Get values from row
		scanError := rows.Scan(dest...)
		err(scanError)

		// Put values to struct
		row := reflect.Indirect(reflect.New(v))
		for i, t := range tmp {
			row.Field(i).Set(reflect.ValueOf(t))
		}
		container.Set(reflect.Append(container, row))
	}
}

func structFromDB(model interface{}, query string, args ...interface{}) error {
	db, databaseError := sql.Open("sqlite3", dbname)
	err(databaseError)
	defer db.Close()
	row := db.QueryRow(query, args...)
	container := reflect.Indirect(reflect.ValueOf(model))
	v := container.Type()
	len := v.NumField()
	tmp := make([]interface{}, len)
	dest := make([]interface{}, len)
	for i := range tmp {
		dest[i] = &tmp[i]
	}
	scanError := row.Scan(dest...)
	if scanError != nil {
		if scanError == sql.ErrNoRows {
			return scanError
		}
		panic(scanError)
	}
	for i, t := range tmp {
		container.Field(i).Set(reflect.ValueOf(t))
	}
	return nil
}

func isInDB(query string, data interface{}) bool {
	db, databaseError := sql.Open("sqlite3", dbname)
	err(databaseError)
	defer db.Close()

	queryError := db.QueryRow(query, data).Scan(&data)
	if queryError != nil {
		if queryError != sql.ErrNoRows {
			panic(err)
		}
		return false
	}
	return true
}
