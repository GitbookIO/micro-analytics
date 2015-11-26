package database

import (
    "database/sql"
    "log"
    "os"
    "path"
    "time"

    "github.com/GitbookIO/analytics/utils"
)

const dbFileName = "analytics.db"

type Database struct {
    Name        string
    Conn        *sql.DB
    StartTime   time.Time
}

type DBManager struct {
    DBs         map[string]*Database
    StartTime   time.Time
    maxDBs      int
    directory   string
}

// Get a new DBManager
func NewManager(directory string, maxDBs int) *DBManager {
    manager := DBManager{
        DBs:        map[string]*Database{},
        StartTime:  time.Now(),
        maxDBs:     maxDBs,
        directory:  directory,
    }

    return &manager
}

// Attach a new DB to the manager
func (manager *DBManager) Register(db *Database) {
    // Unregister a DB if manager is full
    nbActive := len(manager.DBs)
    if nbActive == manager.maxDBs {
        manager.RemoveLongestAlive()
    }

    manager.DBs[db.Name] = db
    log.Printf("[DBManager] Registered DB %s\n", db.Name)
}

// Fully remove a DB from manager
func (manager *DBManager) Unregister(dbName string) {
    // Test that DB was registered
    if _, ok := manager.DBs[dbName]; ok {
        // Close DB
        manager.DBs[dbName].Conn.Close()

        // Unregister DB
        delete(manager.DBs, dbName)
    }
}

// Detach the longest opened DB from manager
func (manager *DBManager) RemoveLongestAlive() {
    var toDelete string
    minTime := time.Now()

    for dbName, db := range manager.DBs {
        if db.StartTime.Before(minTime) {
            toDelete = dbName
            minTime = db.StartTime
        }
    }

    manager.Unregister(toDelete)
    log.Printf("[DBManager] Unregistered DB %s\n", toDelete)
}

// Unregister all attached DBs
func (manager *DBManager) Purge() {
    for dbName, _ := range manager.DBs {
        manager.Unregister(dbName)
    }
}

// Get a DB from manager, register and create if necessary
func (manager *DBManager) GetDB(dbName string) (*Database, error) {
    // Return DB if already registered
    if db, ok := manager.DBs[dbName]; ok {
        log.Printf("[DBManager] Returning registered DB %s\n", dbName)
        return db, nil
    }

    // Create DB directory if doesn't exist
    dbDir := path.Join(manager.directory, dbName)
    dbExists, err := manager.DBExists(dbName)
    if err != nil {
        log.Printf("[DBManager] Error %v reaching for DB %s\n", err, dbName)
        return nil, err
    }
    if !dbExists {
        if err = os.Mkdir(dbDir, os.ModePerm); err != nil {
            log.Printf("[DBManager] Error %v creating directory for DB %s\n", err, dbName)
            return nil, err
        }
    }

    // Open DB connection and returns the full Database
    dbPath := path.Join(dbDir, dbFileName)
    conn, err := OpenAndInitialize(dbPath)
    if err != nil {
        log.Printf("[DBManager] Error %v opening DB %s", err, dbName)
        return nil, err
    }

    database := Database{
        Name:       dbName,
        Conn:       conn,
        StartTime:  time.Now(),
    }

    manager.Register(&database)
    return &database, nil
}

// Check if a DB exists physically
func (manager *DBManager) DBExists(dbName string) (bool, error) {
    dbPath := path.Join(manager.directory, dbName, dbFileName)

    dbExists, err := utils.PathExists(dbPath)
    if err != nil {
        // Error reading file
        log.Printf("[DBManager] Error [%#v] trying to reach file '%s'\n", err, dbPath)
        return false, err
    }

    return dbExists, nil
}

// Fully delete a DB on disk system
func (manager *DBManager) DeleteDB(dbName string) error {
    // Unregister from manager
    manager.Unregister(dbName)

    // Then delete
    dbDir := path.Join(manager.directory, dbName)
    return os.RemoveAll(dbDir)
}