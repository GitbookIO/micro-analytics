package database

import (
    "log"
    "time"

    _ "github.com/mattn/go-sqlite3"
    sq "github.com/Masterminds/squirrel"
)

// Wrapper for inserting through a Database struct
func (db *Database) Insert(analytic Analytic) error {
    // Query
    // insertQuery := `
    // INSERT INTO visits(time, type, path, ip, platform, refererDomain, countryCode)
    // VALUES(?, ?, ?, ?, ?, ?, ?)`
    insertQuery := sq.
        Insert("visits").
        Columns("time", "type", "path", "ip", "platform", "refererDomain", "countryCode").
        Values(time.Now(),
            analytic.Type,
            analytic.Path,
            analytic.Ip,
            analytic.Platform,
            analytic.RefererDomain,
            analytic.CountryCode).
        RunWith(db.Conn)

    _, err := insertQuery.Exec()
    if err != nil {
        log.Printf("Error inserting analytic: %#v\n", analytic)
        log.Printf("%v\n", err)
        return err
    }

    return nil
    // // Create transaction
    // tx, err := db.Begin()
    // if err != nil {
    //     log.Fatal("Error creating transaction", err)
    // }

    // // Create statement
    // stmt, err := tx.Prepare(insertQuery)
    // if err != nil {
    //     log.Fatal("Error preparing transaction", err)
    // }
    // defer stmt.Close()

    // _, err = stmt.Exec(time.Now(),
    //     analytic.Type,
    //     analytic.Path,
    //     analytic.Ip,
    //     analytic.Platform,
    //     analytic.RefererDomain,
    //     analytic.CountryCode)

    // if err != nil {
    //     log.Fatal("Error inserting row", err)
    // }

    // tx.Commit()
}