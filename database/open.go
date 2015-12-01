package database

import (
    "database/sql"
    "log"

    _ "github.com/mattn/go-sqlite3"
)

// Open a DB and returns it
func Open(dbPath string) (*sql.DB, error) {

    // Make DB connection
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Printf("[DBOpen] Connection at DB %s in error: %v\n", dbPath, err)
        return nil, err
    }

    return db, nil
}

// Open a DB and returns it
// Create if necessary
func OpenAndInitialize(dbPath string) (*sql.DB, error) {
    // DB schema
    createTable := `
    CREATE TABLE visits (
        time            INTEGER,
        event           TEXT,
        path            TEXT,
        ip              TEXT,
        platform        TEXT,
        refererDomain   TEXT,
        countryCode     TEXT
    )`

    // Make DB connection
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Printf("[DBOpen] Connection at DB %s in error: %v\n", dbPath, err)
        return nil, err
    }

    // Create table at initialization
    if !TableExists(db) {
        _, err = db.Exec(createTable)
        if err != nil {
            log.Printf("Error %v creating table %s\n", err, dbPath)
            return nil, err
        }
    }

    return db, nil
}