package sqlite

import (
	"database/sql"
	"path"
	"time"
)

const dbFileName = "analytics.db"

type Database struct {
	Path      DBPath
	Conn      *sql.DB
	StartTime time.Time
	Freed     chan bool
	Pending   int
}

type DBPath struct {
	Name      string
	Directory string
}

// Print Database Name (i.e. DBPath)
func (db *Database) Name() string {
	return db.Path.FileName()
}

// Print DBPath with filename
func (dbPath *DBPath) FileName() string {
	return path.Join(dbPath.Directory, dbPath.Name, dbFileName)
}

// Print DBPath directory
func (dbPath *DBPath) String() string {
	return path.Join(dbPath.Directory, dbPath.Name)
}
