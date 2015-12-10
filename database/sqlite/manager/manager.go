package manager

import (
	"database/sql"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/azer/logger"

	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/utils"
)

type Opts struct {
	database.DriverOpts
}

type DBManager struct {
	sync.Mutex
	DBs       map[string]*Database
	StartTime time.Time
	maxDBs    int
	Logger    *logger.Logger
}

// Get a new DBManager
func New(opts Opts) *DBManager {
	manager := DBManager{
		DBs:       map[string]*Database{},
		StartTime: time.Now(),
		maxDBs:    opts.MaxDBs,
		Logger:    logger.New("[DBManager]"),
	}

	// Handle cleaning connections
	// Allow for more than maxDBs to run at full charge
	go func() {
		for {
			time.Sleep(time.Second * 5)
			manager.cleanConnections()
		}
	}()

	// Handle closing connections when app is killed
	go func() {
		<-opts.ClosingChannel
		manager.purge()
		opts.ClosingChannel <- true
	}()

	return &manager
}

// Get a DB from manager, register and create if necessary
func (manager *DBManager) GetDB(dbPath DBPath) (*Database, error) {
	manager.Lock()
	defer manager.Unlock()

	// Return DB if already registered
	if db, ok := manager.DBs[dbPath.String()]; ok {
		return db, nil
	}

	// Create/open
	conn, err := manager.openOrCreate(dbPath)
	if err != nil {
		return nil, err
	}

	// Register DB
	database := Database{
		Path:      dbPath,
		Conn:      conn,
		StartTime: time.Now(),
	}
	manager.register(&database)

	return &database, nil
}

// Check if the DB folder exists physically
func (manager *DBManager) DBExists(dbPath DBPath) (bool, error) {
	dbExists, err := utils.PathExists(dbPath.String())
	if err != nil {
		// Error reading file
		manager.Logger.Error("Error [%v] trying to reach file '%s'", err, dbPath.FileName())
		return false, err
	}

	return dbExists, nil
}

// Fully delete a DB on disk system
func (manager *DBManager) DeleteDB(dbPath DBPath) error {
	manager.Lock()
	defer manager.Unlock()

	// Unregister from manager
	manager.unregister(dbPath)

	// Then delete
	return os.RemoveAll(dbPath.String())
}

// Attach a new DB to the manager
func (manager *DBManager) register(db *Database) {
	dbName := db.Path.String()
	manager.DBs[dbName] = db
}

// Fully remove a DB from manager
func (manager *DBManager) unregister(dbPath DBPath) {
	// Test that DB was registered
	if db, ok := manager.DBs[dbPath.String()]; ok {
		// Lock DB
		db.Lock()
		defer db.Unlock()

		// Close DB
		db.Conn.Close()

		// Unregister DB
		delete(manager.DBs, db.Path.String())
	}
}

// Unregister all attached DBs
func (manager *DBManager) purge() {
	manager.Lock()
	defer manager.Unlock()

	for _, db := range manager.DBs {
		manager.unregister(db.Path)
	}
}

// Detach the longest opened DB from manager
func (manager *DBManager) removeUnpending() error {
	manager.Lock()
	defer manager.Unlock()

	var toDelete string
	minTime := time.Now()

	for dbName, db := range manager.DBs {
		if db.StartTime.Before(minTime) {
			toDelete = dbName
			minTime = db.StartTime
		}
	}

	if len(toDelete) == 0 {
		return errors.New("All registered DBs are busy at this time")
	}

	manager.unregister(manager.DBs[toDelete].Path)
	return nil
}

// Ensure number of alive connections drops to maxDBs
func (manager *DBManager) cleanConnections() {
	nbActive := len(manager.DBs)
	var err error

	for nbActive > manager.maxDBs && err == nil {
		manager.Logger.Info("Cleaning alive connections: %v / %v available", nbActive, manager.maxDBs)
		err = manager.removeUnpending()
		nbActive = len(manager.DBs)
	}

	if err != nil {
		manager.Logger.Info("%v", err)
	}
}

func (manager *DBManager) openOrCreate(dbPath DBPath) (*sql.DB, error) {
	if exists, err := manager.DBExists(dbPath); err != nil {
		return nil, err
	} else if exists {
		return manager.openDB(dbPath)
	}
	return manager.createDB(dbPath)
}

func (manager *DBManager) openDB(dbPath DBPath) (*sql.DB, error) {
	conn, err := OpenConnection(dbPath.FileName())
	if err != nil {
		manager.Logger.Error("Error [%v] opening DB %s", err, dbPath)
		return nil, err
	}
	return conn, nil
}

func (manager *DBManager) createDB(dbPath DBPath) (*sql.DB, error) {
	if err := os.MkdirAll(dbPath.String(), os.ModePerm); err != nil {
		manager.Logger.Error("Error [%v] creating directory for DB %s", err, dbPath)
		return nil, err
	}

	// Open DB connection and returns the full Database
	conn, err := InitializeDatabase(dbPath.FileName())
	if err != nil {
		manager.Logger.Error("Error [%v] opening DB %s", err, dbPath)
		return nil, err
	}

	return conn, nil
}
