package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type DB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Close() error
}

type Database struct {
	*sql.DB
}

func NewDatabase(driverName, dataSourceName string) (DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	// Check if the connection is working
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db}, nil
}
