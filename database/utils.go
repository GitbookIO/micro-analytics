package database

import (
    "database/sql"
    "log"
)

// Check wether the table already exists
func TableExists(db *sql.DB) bool {
    // Query
    existsQuery := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='visits'`

    // Execute query
    rows, err := db.Query(existsQuery)
    if err != nil {
        log.Fatal("[DBUtils] Error checking if table exists\n", err)
    }
    defer rows.Close()

    // Get result
    var count int
    for rows.Next() {
        rows.Scan(&count)
    }

    return count == 1
}