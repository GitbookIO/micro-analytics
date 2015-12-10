package manager

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const dbSchema = `
    CREATE TABLE visits (
        time            INTEGER,
        event           TEXT,
        path            TEXT,
        ip              TEXT,
        platform        TEXT,
        refererDomain   TEXT,
        countryCode     TEXT
    )`

// Open an SQLite3 DB connection
func OpenConnection(dbPath string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	// Make DB connection
	if db, err = sql.Open("sqlite3", dbPath); err != nil {
		return nil, err
	}

	return db, nil
}

// Open an SQLite3 DB connection and create table schema
func InitializeDatabase(dbPath string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	if db, err = OpenConnection(dbPath); err != nil {
		return nil, err
	}

	if _, err = db.Exec(dbSchema); err != nil {
		return nil, err
	}

	return db, nil
}
