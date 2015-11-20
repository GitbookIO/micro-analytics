package database

import (
    "log"

    sq "github.com/Masterminds/squirrel"
)

// Wrapper for querying a Database struct
func (db *Database) Query() []Analytic {
    // Query
    // query := `SELECT time, type, path, ip, platform, refererDomain, countryCode FROM visits`
    query, _, err := sq.
        Select("time", "type", "path", "ip", "platform", "refererDomain", "countryCode").
        From("visits").
        ToSql()

    // Exec query
    rows, err := db.Conn.Query(query)
    if err != nil {
        log.Fatal("Error querying DB", err)
    }
    defer rows.Close()

    analytics := []Analytic{}
    for rows.Next() {
        analytic := Analytic{}
        rows.Scan(&analytic.Time,
            &analytic.Type,
            &analytic.Path,
            &analytic.Ip,
            &analytic.Platform,
            &analytic.RefererDomain,
            &analytic.CountryCode)
        analytics = append(analytics, analytic)
    }

    return analytics
}