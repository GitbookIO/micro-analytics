package manager

import (
	"database/sql"
	"os"
	"path"
	"time"

	"github.com/GitbookIO/go-sqlpool"
	"github.com/azer/logger"
	_ "github.com/mattn/go-sqlite3"

	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/utils"
)

type Opts struct {
	database.DriverOpts
}

type DBManager struct {
	StartTime time.Time
	Logger    *logger.Logger
	Pool      *sqlpool.Pool
}

// Get a new DBManager
func New(opts Opts) *DBManager {
	// Configure sqlPool
	poolOpts := sqlpool.Opts{
		Max:         int64(opts.MaxDBs),
		IdleTimeout: int64(opts.IdleTimeout),
		PreInit:     createDirectory,
		PostInit:    initializeDatabase,
	}
	pool := sqlpool.NewPool(poolOpts)

	// Start a new DBManager
	manager := DBManager{
		StartTime: time.Now(),
		Logger:    logger.New("[DBManager]"),
		Pool:      pool,
	}

	// Handle closing connections when app is killed
	go func() {
		<-opts.ClosingChannel
		manager.Pool.ForceClose()
		opts.ClosingChannel <- true
	}()

	return &manager
}

// Acquire a DB connection
func (manager *DBManager) Acquire(dbPath DBPath) (*sqlpool.Resource, error) {
	return manager.Pool.Acquire("sqlite3", dbPath.FileName())
}

// Release a DB connection
func (manager *DBManager) Release(resource *sqlpool.Resource) error {
	return manager.Pool.Release(resource)
}

// Check if the DB folder exists physically
func (manager *DBManager) DBExists(dbPath DBPath) (bool, error) {
	return utils.PathExists(dbPath.String())
}

// Fully delete a DB on disk system
func (manager *DBManager) DeleteDB(dbPath DBPath) error {
	return os.RemoveAll(dbPath.String())
}

// PreInit function for sqlPool
func createDirectory(driver, url string) error {
	dbExists, err := utils.PathExists(url)
	if err != nil {
		return err
	}

	if !dbExists {
		err = os.MkdirAll(path.Dir(url), os.ModePerm)
	}

	return err
}

// PostInit function for sqlPool
func initializeDatabase(db *sql.DB) error {
	const dbSchema = `
    CREATE TABLE visits (
        time            INTEGER,
        event           TEXT,
        path            TEXT,
        ip              TEXT,
        platform        TEXT,
        refererDomain   TEXT,
        countryCode     TEXT
    )`

	tableExists, err := tableExists(db)
	if err != nil {
		return err
	}

	if !tableExists {
		_, err = db.Exec(dbSchema)
	}

	return err
}

// Check wether the visits table already exists
func tableExists(db *sql.DB) (bool, error) {
	// Query
	existsQuery := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='visits'`

	// Execute query
	var count int
	err := db.QueryRow(existsQuery).Scan(&count)

	if err != nil {
		return false, err
	}

	return count == 1, nil
}
