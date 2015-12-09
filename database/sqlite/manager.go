package sqlite

import (
    "database/sql"
    "errors"
    "os"
    "time"

    "github.com/azer/logger"
    "github.com/hashicorp/golang-lru"

    "github.com/GitbookIO/micro-analytics/database"
    "github.com/GitbookIO/micro-analytics/utils"
)

type ManagerOpts struct {
    database.DriverOpts
}

type UnlockChannel struct {
    dbPath  DBPath
    removed bool
}

type DBManager struct {
    DBs       map[string]*Database
    StartTime time.Time
    maxDBs    int
    RequestDB chan DBPath
    UnlockDB  chan UnlockChannel
    SendDB    chan *Database
    Cache     *lru.Cache
    Logger    *logger.Logger
}

// Generate an unlockChannel
func NewUnlock(dbPath DBPath, removedOpt ...bool) UnlockChannel {
    removed := false
    if len(removedOpt) > 0 {
        removed = removedOpt[0]
    }

    return UnlockChannel{
        dbPath:  dbPath,
        removed: removed,
    }
}

// Get a new DBManager
func NewManager(opts ManagerOpts) *DBManager {
    manager := DBManager{
        DBs:       map[string]*Database{},
        StartTime: time.Now(),
        maxDBs:    opts.MaxDBs,
        RequestDB: make(chan DBPath),
        UnlockDB:  make(chan UnlockChannel),
        SendDB:    make(chan *Database),
        Logger:    logger.New("[DBManager]"),
    }

    cache, err := lru.New(opts.CacheSize)
    if err != nil {
        manager.Logger.Error("Failed to create cache for DBManager: [%v]", err)
    }
    manager.Cache = cache

    // Handle cleaning connections
    // Allow for more than maxDBs to run at full charge
    go func() {
        for {
            time.Sleep(time.Second * 5)

            nbActive := len(manager.DBs)
            var err error

            // Limit to 15 test
            count := 0

            for nbActive > manager.maxDBs && err == nil && count < 15 {
                count += 1
                manager.Logger.Info("Cleaning alive connections: %v / %v available", nbActive, manager.maxDBs)
                nbActive = len(manager.DBs)
                err = manager.RemoveUnpending()
            }
        }
    }()

    // Handle registering DBs
    go func() {
        for {
            dbPath := <-manager.RequestDB
            db, err := manager.GetDB(dbPath)
            if err != nil {
                manager.Logger.Error("Impossible to get DB %s: Error [%v]", dbPath.FileName(), err)
            }
            manager.SendDB <- db
        }
    }()

    // Handle unlocking DBs
    go func() {
        for {
            unlock := <-manager.UnlockDB
            dbName := unlock.dbPath.String()

            if !unlock.removed {
                manager.DBs[dbName].Pending -= 1
                manager.DBs[dbName].Freed <- true
            }
        }
    }()

    // Handle closing connections when app is killed
    go func() {
        <-opts.ClosingChannel
        manager.Purge()
        opts.ClosingChannel <- true
    }()

    return &manager
}

// Attach a new DB to the manager
func (manager *DBManager) Register(db *Database) {
    dbName := db.Path.String()
    manager.DBs[dbName] = db
}

// Fully remove a DB from manager
func (manager *DBManager) Unregister(dbPath DBPath) {
    // Test that DB was registered
    if _, ok := manager.DBs[dbPath.String()]; ok {
        // Lock DB
        manager.Logger.Info("Sending request for DB %s", dbPath)
        manager.RequestDB <- dbPath
        manager.Logger.Info("Waiting for DB %s", dbPath)
        db := <-manager.SendDB

        // Close DB
        db.Conn.Close()

        // Unlock DB
        unlock := NewUnlock(dbPath, true)
        manager.Logger.Info("Unlocking for DB %s", dbPath)
        manager.UnlockDB <- unlock
        manager.Logger.Info("Unlocked DB %s", dbPath)

        // Unregister DB
        delete(manager.DBs, dbPath.String())
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

    manager.Unregister(manager.DBs[toDelete].Path)
    return nil
}

// Unregister all attached DBs
func (manager *DBManager) Purge() {
    for _, db := range manager.DBs {
        manager.Unregister(db.Path)
    }
}

// Get a DB from manager, register and create if necessary
func (manager *DBManager) GetDB(dbPath DBPath) (*Database, error) {
    dbName := dbPath.String()

    // Return DB if already registered
    if db, ok := manager.DBs[dbName]; ok {
        db.Pending += 1
        // Wait for DB to return from last query
        <-db.Freed
        return db, nil
    }

    // Register DB
    database := Database{
        Path:      dbPath,
        Conn:      nil,
        StartTime: time.Now(),
        Freed:     make(chan bool, 1),
        Pending:   1,
    }

    manager.Register(&database)

    // Create DB directory if doesn't exist
    dbExists, err := manager.DBExists(dbPath)
    if err != nil {
        manager.Logger.Error("Error [%v] checking if DB %s exists", err, dbName)
        return nil, err
    }

    var conn *sql.DB

    if !dbExists {
        if err = os.MkdirAll(dbPath.String(), os.ModePerm); err != nil {
            manager.Logger.Error("Error [%v] creating directory for DB %s", err, dbName)
            return nil, err
        }

        // Open DB connection and returns the full Database
        conn, err = OpenAndInitialize(dbPath.FileName())
        if err != nil {
            manager.Logger.Error("Error [%v] opening DB %s", err, dbName)
            return nil, err
        }
    } else {
        conn, err = Open(dbPath.FileName())
        if err != nil {
            manager.Logger.Error("Error [%v] opening DB %s", err, dbName)
            return nil, err
        }
    }

    database.Conn = conn

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
    // Unregister from manager
    manager.Unregister(dbPath)

    // Then delete
    return os.RemoveAll(dbPath.String())
}
