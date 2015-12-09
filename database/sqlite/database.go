package sqlite

import (
	"database/sql"
	"path"
	"sync"
	"time"
)

const dbFileName = "analytics.db"

type Database struct {
	sync.Mutex
	Path      DBPath
	Conn      *sql.DB
	StartTime time.Time
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
