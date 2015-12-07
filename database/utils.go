package database

import (
    "database/sql"
    "github.com/azer/logger"
)

// Check wether the table already exists
func TableExists(db *sql.DB) bool {
    var log = logger.New("[Database.TableExists]")

    // Query
    existsQuery := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='visits'`

    // Execute query
    rows, err := db.Query(existsQuery)
    if err != nil {
        log.Error("Error [%v] checking if table exists", err)
        return false
    }
    defer rows.Close()

    // Get result
    var count int
    for rows.Next() {
        rows.Scan(&count)
    }

    return count == 1
}
