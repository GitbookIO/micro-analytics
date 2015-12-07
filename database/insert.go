package database

import (
    sq "github.com/Masterminds/squirrel"
    "github.com/azer/logger"
    _ "github.com/mattn/go-sqlite3"
)

// Wrapper for inserting through a Database struct
func (db *Database) Insert(analytic Analytic) error {
    var log = logger.New("[Database.Insert]")

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
        log.Error("Error [%v] inserting analytic %#v", err, analytic)
        return err
    }

    return nil
}
