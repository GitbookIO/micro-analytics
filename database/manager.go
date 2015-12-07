package database

import (
    "database/sql"
    "errors"
    "os"
    "path"
    "time"

    "github.com/azer/logger"

    "github.com/GitbookIO/micro-analytics/utils"
)

const dbFileName = "analytics.db"

type Database struct {
    Name      string
    Conn      *sql.DB
    StartTime time.Time
    Freed     chan bool
    Pending   int
}

type DBManager struct {
    DBs       map[string]*Database
    StartTime time.Time
    maxDBs    int
    directory string
    RequestDB chan string
    SendDB    chan *Database
    UnlockDB  chan string
    Cacher    map[string]interface{}
    Logger    *logger.Logger
}

// Get a new DBManager
func NewManager(directory string, maxDBs int) *DBManager {
    manager := DBManager{
        DBs:       map[string]*Database{},
        StartTime: time.Now(),
        maxDBs:    maxDBs,
        directory: directory,
        RequestDB: make(chan string),
        SendDB:    make(chan *Database),
        UnlockDB:  make(chan string),
        Cacher:    make(map[string]interface{}),
        Logger:    logger.New("[DBManager]"),
    }

    // Handle cleaning connections
    // Allow for more than maxDBs to run at full charge
    go func() {
        for {
            time.Sleep(time.Second * 5)

            nbActive := len(manager.DBs)
            var err error

            for nbActive > manager.maxDBs && err == nil {
                manager.Logger.Info("Cleaning alive connections: %v / %v available", nbActive, manager.maxDBs)
                nbActive = len(manager.DBs)
                err = manager.RemoveUnpending()
            }
        }
    }()

    // Handle registering DBs
    go func() {
        for {
            dbName := <-manager.RequestDB
            db, err := manager.GetDB(dbName)
            if err != nil {
                manager.Logger.Error("Impossible to get DB %s: Error [%v]", dbName, err)
            }
            manager.SendDB <- db
        }
    }()

    // Handle unlocking DBs
    go func() {
        for {
            dbName := <-manager.UnlockDB

            manager.DBs[dbName].Pending -= 1
            manager.DBs[dbName].Freed <- true
        }
    }()

    return &manager
}

// Attach a new DB to the manager
func (manager *DBManager) Register(db *Database) {
    manager.DBs[db.Name] = db
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
func (manager *DBManager) RemoveUnpending() error {
    var toDelete string
    minTime := time.Now()

    for dbName, db := range manager.DBs {
        if db.Pending == 0 && db.StartTime.Before(minTime) {
            toDelete = dbName
            minTime = db.StartTime
        }
    }

    if len(toDelete) == 0 {
        return errors.New("All registered DBs are busy at this time")
    }

    manager.Unregister(toDelete)
    return nil
}

// Unregister all attached DBs
func (manager *DBManager) Purge() {
    for dbName := range manager.DBs {
        manager.Unregister(dbName)
    }
}

// Get a DB from manager, register and create if necessary
func (manager *DBManager) GetDB(dbName string) (*Database, error) {
    // Return DB if already registered
    if db, ok := manager.DBs[dbName]; ok {
        db.Pending += 1
        // Wait for DB to return from last query
        <-db.Freed
        return db, nil
    }

    // Register DB
    database := Database{
        Name:      dbName,
        Conn:      nil,
        StartTime: time.Now(),
        Freed:     make(chan bool, 1),
        Pending:   1,
    }

    manager.Register(&database)

    // Create DB directory if doesn't exist
    dbDir := path.Join(manager.directory, dbName)
    dbExists, err := manager.DBExists(dbName)
    if err != nil {
        manager.Logger.Error("Error [%v] reaching for DB %s", err, dbName)
        return nil, err
    }

    var conn *sql.DB
    dbPath := path.Join(dbDir, dbFileName)

    if !dbExists {
        if err = os.Mkdir(dbDir, os.ModePerm); err != nil {
            manager.Logger.Error("Error [%v] creating directory for DB %s", err, dbName)
            return nil, err
        }

        // Open DB connection and returns the full Database
        conn, err = OpenAndInitialize(dbPath)
        if err != nil {
            manager.Logger.Error("Error [%v] opening DB %s", err, dbName)
            return nil, err
        }
    } else {
        conn, err = Open(dbPath)
        if err != nil {
            manager.Logger.Error("Error [%v] opening DB %s", err, dbName)
            return nil, err
        }
    }

    database.Conn = conn

    return &database, nil
}

// Check if a DB exists physically
func (manager *DBManager) DBExists(dbName string) (bool, error) {
    dbPath := path.Join(manager.directory, dbName, dbFileName)

    dbExists, err := utils.PathExists(dbPath)
    if err != nil {
        // Error reading file
        manager.Logger.Error("Error [%v] trying to reach file '%s'", err, dbPath)
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
