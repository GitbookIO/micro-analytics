package database

import (
    "database/sql"
    "log"
    "path"
    "os"

    _ "github.com/mattn/go-sqlite3"

    "github.com/GitbookIO/analytics/utils"
)

// Open a DB and returns it
// Create if necessary
func OpenOrCreate(dbName string) (*sql.DB, error) {
    // DB schema
    createTable := `
    CREATE TABLE visits (
        time            TIMESTAMP,
        type            VARCHAR,
        path            VARCHAR,
        ip              VARCHAR,
        platform        VARCHAR,
        refererDomain   VARCHAR,
        countryCode     VARCHAR
    )`

    // DB index
    // createIndex := `CREATE INDEX visits_object_time_idx ON visits (object, time)`

    // DB file name
    dbDir := path.Join("./dbs", dbName)
    dbPath := path.Join(dbDir, "analytics.db")

    // Create DB directory if inexistant
    dirExists, err := utils.PathExists(dbDir)
    if err != nil {
        log.Fatal("DB directory path error", err)
    }
    if !dirExists {
        log.Printf("Creating new DB directory at: %s", dbDir)
        os.Mkdir(dbDir, os.ModePerm)
    }

    // DB connection
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatal("Connection error\n", err)
    }

    if !TableExists(db) {
        // Create DB
        _, err = db.Exec(createTable)
        if err != nil {
            log.Printf("Error %q creating table %s\n", err, dbPath)
            return nil, err
        }

        // Create index
        // _, err = db.Exec(createIndex)
        // if err != nil {
        //     log.Printf("Error %q creating index for table %s\n", err, dbPath)
        //     return nil, err
        // }
    }

    return db, nil
}