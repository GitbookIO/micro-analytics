package database

import (
    "database/sql"
    "log"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

func Insert(db *sql.DB, analytic Analytic) {
    // Query
    insertQuery := `
    INSERT INTO visits(time, type, path, ip, platform, refererDomain, countryCode)
    VALUES(?, ?, ?, ?, ?, ?, ?)`

    // Create transaction
    tx, err := db.Begin()
    if err != nil {
        log.Fatal("Error creating transaction", err)
    }

    // Create statement
    stmt, err := tx.Prepare(insertQuery)
    if err != nil {
        log.Fatal("Error preparing transaction", err)
    }
    defer stmt.Close()

    _, err = stmt.Exec(time.Now(),
        analytic.Type,
        analytic.Path,
        analytic.Ip,
        analytic.Platform,
        analytic.RefererDomain,
        analytic.CountryCode)

    if err != nil {
        log.Fatal("Error inserting row", err)
    }

    tx.Commit()
}