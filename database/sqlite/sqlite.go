package sqlite

import (
	"github.com/GitbookIO/go-sqlpool"

	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/database/errors"

	"github.com/GitbookIO/micro-analytics/database/sqlite/manager"
	"github.com/GitbookIO/micro-analytics/database/sqlite/query"
)

type SQLite struct {
	DBManager *manager.DBManager
	directory string
}

func NewSimpleDriver(driverOpts database.DriverOpts) *SQLite {
	manager := manager.New(manager.Opts{driverOpts})
	return &SQLite{
		DBManager: manager,
		directory: driverOpts.Directory,
	}
}

func (driver *SQLite) Query(params database.Params) (*database.Analytics, error) {
	// Construct DBPath
	dbPath := manager.DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Check if DB file exists
	dbExists, err := driver.DBManager.DBExists(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}

	// DB doesn't exist
	if !dbExists {
		return nil, &errors.InvalidDatabaseName
	}

	// Get DB from manager
	db, err := driver.DBManager.Acquire(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}
	defer driver.DBManager.Release(db)

	// Return query result
	analytics, err := query.Query(db.DB, params.TimeRange)
	if err != nil {
		return nil, &errors.InternalError
	}

	return analytics, nil
}

func (driver *SQLite) Count(params database.Params) (*database.Count, error) {
	// Construct DBPath
	dbPath := manager.DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Check if DB file exists
	dbExists, err := driver.DBManager.DBExists(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}

	// DB doesn't exist
	if !dbExists {
		return nil, &errors.InvalidDatabaseName
	}

	// Get DB from manager
	db, err := driver.DBManager.Acquire(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}
	defer driver.DBManager.Release(db)

	// Return query result
	analytics, err := query.Count(db.DB, params.TimeRange)
	if err != nil {
		return nil, &errors.InternalError
	}

	return analytics, nil
}

func (driver *SQLite) GroupBy(params database.Params) (*database.Aggregates, error) {
	// Construct DBPath
	dbPath := manager.DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Check if DB file exists
	dbExists, err := driver.DBManager.DBExists(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}

	// DB doesn't exist
	if !dbExists {
		return nil, &errors.InvalidDatabaseName
	}

	// Get DB from manager
	db, err := driver.DBManager.Acquire(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}
	defer driver.DBManager.Release(db)

	// Check for unique query parameter to call function accordingly
	var analytics *database.Aggregates

	if params.Unique {
		analytics, err = query.GroupByUniq(db.DB, params.Property, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	} else {
		analytics, err = query.GroupBy(db.DB, params.Property, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	}

	return analytics, nil
}

func (driver *SQLite) Series(params database.Params) (*database.Intervals, error) {
	// Construct DBPath
	dbPath := manager.DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Check if DB file exists
	dbExists, err := driver.DBManager.DBExists(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}

	// DB doesn't exist
	if !dbExists {
		return nil, &errors.InvalidDatabaseName
	}

	// Get DB from manager
	db, err := driver.DBManager.Acquire(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}
	defer driver.DBManager.Release(db)

	// Check for unique query parameter to call function accordingly
	var analytics *database.Intervals

	if params.Unique {
		analytics, err = query.SeriesUniq(db.DB, params.Interval, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	} else {
		analytics, err = query.Series(db.DB, params.Interval, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	}

	return analytics, nil
}

func (driver *SQLite) Insert(params database.Params, analytic database.Analytic) error {
	// Construct DBPath
	dbPath := manager.DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Get DB from manager
	db, err := driver.DBManager.Acquire(dbPath)
	if err != nil {
		return &errors.InternalError
	}
	defer driver.DBManager.Release(db)

	// Insert data if everything's OK
	err = query.Insert(db.DB, analytic)

	if err != nil {
		return &errors.InsertFailed
	}

	return nil
}

func (driver *SQLite) BulkInsert(analytics map[string][]database.Analytic) error {
	var acquireErr, insertErr error
	var db *sqlpool.Resource
	// Run a bulk insert query for each database
	for dbName, _analytics := range analytics {
		// Construct DBPath
		dbPath := manager.DBPath{
			Name:      dbName,
			Directory: driver.directory,
		}

		// Get DB from manager
		db, acquireErr = driver.DBManager.Acquire(dbPath)
		if acquireErr != nil {
			// Impossible to get database, process next analytics
			continue
		}
		defer driver.DBManager.Release(db)

		// Insert data if everything's OK
		insertErr = query.BulkInsert(db.DB, _analytics)
	}

	if insertErr != nil {
		return &errors.InsertFailed
	}
	if acquireErr != nil {
		return &errors.InternalError
	}

	return nil
}

func (driver *SQLite) Delete(params database.Params) error {
	// Construct DBPath
	dbPath := manager.DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Check if DB file exists
	dbExists, err := driver.DBManager.DBExists(dbPath)
	if err != nil {
		return &errors.InternalError
	}

	// DB doesn't exist
	if !dbExists {
		return &errors.InvalidDatabaseName
	}

	// Delete full DB directory
	err = driver.DBManager.DeleteDB(dbPath)
	return err
}

var _ database.Driver = &SQLite{}
