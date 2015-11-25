package database

import (
    "log"

    _ "github.com/mattn/go-sqlite3"
    sq "github.com/Masterminds/squirrel"
)

// Wrapper for inserting through a Database struct
func (db *Database) Insert(analytic Analytic) error {
    insertQuery := sq.
        Insert("visits").
        Columns("time", "event", "path", "ip", "platform", "refererDomain", "countryCode").
        Values(analytic.Time.Unix(),
            analytic.Event,
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
}