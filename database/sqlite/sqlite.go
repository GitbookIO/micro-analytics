package sqlite

import (
	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/database/errors"
	"github.com/GitbookIO/micro-analytics/database/structures"

	"github.com/GitbookIO/micro-analytics/database/sqlite/manager"
	"github.com/GitbookIO/micro-analytics/database/sqlite/query"
)

type SQLite struct {
	DBManager *manager.DBManager
	directory string
}

func NewSimpleDriver(driverOpts database.DriverOpts) *SQLite {
	manager := manager.NewManager(manager.ManagerOpts{driverOpts})
	return &SQLite{
		DBManager: manager,
		directory: driverOpts.Directory,
	}
}

func (driver *SQLite) Query(params structures.Params) (*structures.Analytics, error) {
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
	db, err := driver.DBManager.GetDB(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}

	// If value is in Cache, return directly
	// cached, inCache := driver.DBManager.Cache.Get(params.URL)
	// if inCache {
	// 	if response, ok := cached.(*structures.Analytics); ok {
	// 		driver.DBManager.UnlockDB <- NewUnlock(dbPath)
	// 		return response, nil
	// 	}
	// }

	// Return query result
	analytics, err := query.Query(db.Conn, params.TimeRange)
	if err != nil {
		return nil, &errors.InternalError
	}

	// Store response in Cache before sending
	// driver.DBManager.Cache.Add(params.URL, analytics)

	return analytics, nil
}

func (driver *SQLite) GroupBy(params structures.Params) (*structures.Aggregates, error) {
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
	db, err := driver.DBManager.GetDB(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}

	// If value is in Cache, return directly
	// cached, inCache := driver.DBManager.Cache.Get(params.URL)
	// if inCache {
	// 	if response, ok := cached.(*structures.Aggregates); ok {
	// 		driver.DBManager.UnlockDB <- NewUnlock(dbPath)
	// 		return response, nil
	// 	}
	// }

	// Check for unique query parameter to call function accordingly
	var analytics *structures.Aggregates

	if params.Unique {
		analytics, err = query.GroupByUniq(db.Conn, params.Property, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	} else {
		analytics, err = query.GroupBy(db.Conn, params.Property, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	}

	// Store response in Cache before sending
	// driver.DBManager.Cache.Add(params.URL, analytics)

	return analytics, nil
}

func (driver *SQLite) Series(params structures.Params) (*structures.Intervals, error) {
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
	db, err := driver.DBManager.GetDB(dbPath)
	if err != nil {
		return nil, &errors.InternalError
	}

	// If value is in Cache, return directly
	// cached, inCache := driver.DBManager.Cache.Get(params.URL)
	// if inCache {
	// 	if response, ok := cached.(*structures.Intervals); ok {
	// 		driver.DBManager.UnlockDB <- NewUnlock(dbPath)
	// 		return response, nil
	// 	}
	// }

	// Check for unique query parameter to call function accordingly
	var analytics *structures.Intervals

	if params.Unique {
		analytics, err = query.OverTimeUniq(db.Conn, params.Interval, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	} else {
		analytics, err = query.OverTime(db.Conn, params.Interval, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}
	}

	// Store response in Cache before sending
	// driver.DBManager.Cache.Add(params.URL, analytics)

	return analytics, nil
}

func (driver *SQLite) Insert(params structures.Params, analytic structures.Analytic) error {
	// Construct DBPath
	dbPath := manager.DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Get DB from manager
	db, err := driver.DBManager.GetDB(dbPath)
	if err != nil {
		return &errors.InternalError
	}

	// Insert data if everything's OK
	err = query.Insert(db.Conn, analytic)

	if err != nil {
		return &errors.InsertFailed
	}

	return nil
}

func (driver *SQLite) Delete(params structures.Params) error {
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
