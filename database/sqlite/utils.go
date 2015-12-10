package sqlite

import (
	"database/sql"
	"io/ioutil"

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

func ListShards(dbPath DBPath) []string {
	folders, err := ioutil.ReadDir(dbPath.String())
	if err != nil {
		return nil
	}

	shards := make([]string, 0)
	for _, folder := range folders {
		shards = append(shards, folder.Name())
	}

	return shards
}
