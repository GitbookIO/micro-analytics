package sqlite

import (
	"strconv"
	"time"

	"github.com/azer/logger"

	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/database/errors"
	"github.com/GitbookIO/micro-analytics/database/structures"
)

type Sharded struct {
	DBManager *DBManager
	directory string
	logger    *logger.Logger
}

func NewShardedDriver(driverOpts database.DriverOpts) *Sharded {
	manager := NewManager(ManagerOpts{driverOpts})
	return &Sharded{
		DBManager: manager,
		directory: driverOpts.Directory,
		logger:    logger.New("[Sharded]"),
	}
}

func (driver *Sharded) Query(params structures.Params) (*structures.Analytics, error) {
	// Construct DBPath
	dbPath := DBPath{
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

	// At this point, there should be shards to query
	// Get list of shards by reading directory
	shards := ListShards(dbPath)
	analytics := structures.Analytics{}

	// Read from each shard
	for _, shard := range shards {
		// Don't include shard if not in timerange
		shardInt, err := strconv.Atoi(shard)
		if err != nil {
			driver.logger.Error("Error [%v] converting shard %s name to an integer", err, shard)
		}

		startInt, endInt := params.TimeRange.ConvertToInt()
		if shardInt < startInt || shardInt > endInt {
			continue
		}

		// Construct each shard DBPath
		shardPath := DBPath{
			Name:      shard,
			Directory: dbPath.String(),
		}

		// Get DB shard from manager
		driver.DBManager.RequestDB <- shardPath
		db := <-driver.DBManager.SendDB

		// Return query result
		shardAnalytics, err := db.Query(params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}

		// Unlock DB
		driver.DBManager.UnlockDB <- NewUnlock(shardPath)

		// Add shard result to analytics
		for _, analytic := range shardAnalytics.List {
			analytics.List = append(analytics.List, analytic)
		}
	}

	// // If value is in Cache, return directly
	// cached, inCache := driver.DBManager.Cache.Get(params.URL)
	// if inCache {
	// 	if response, ok := cached.(*structures.Analytics); ok {
	// 		driver.DBManager.UnlockDB <- dbPath
	// 		return response, nil
	// 	}
	// }

	// // Store response in Cache before sending
	// driver.DBManager.Cache.Add(params.URL, analytics)

	return &analytics, nil
}

func (driver *Sharded) GroupBy(params structures.Params) (*structures.Aggregates, error) {
	// Construct DBPath
	dbPath := DBPath{
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

	// At this point, there should be shards to query
	// Get list of shards by reading directory
	shards := ListShards(dbPath)

	// Aggregated query result
	analytics := structures.Aggregates{}
	// Helper map to aggregate
	analyticsMap := map[string]structures.Aggregate{}

	// Read from each shard
	for _, shard := range shards {
		// Don't include shard if not in timerange
		shardInt, err := strconv.Atoi(shard)
		if err != nil {
			driver.logger.Error("Error [%v] converting shard %s name to an integer", err, shard)
		}

		startInt, endInt := params.TimeRange.ConvertToInt()
		if shardInt < startInt || shardInt > endInt {
			continue
		}

		// Construct each shard DBPath
		shardPath := DBPath{
			Name:      shard,
			Directory: dbPath.String(),
		}

		// Get DB shard from manager
		driver.DBManager.RequestDB <- shardPath
		db := <-driver.DBManager.SendDB

		var shardAnalytics *structures.Aggregates

		// Check for unique query parameter to call function accordingly
		if params.Unique {
			shardAnalytics, err = db.GroupByUniq(params.Property, params.TimeRange)
			if err != nil {
				return nil, &errors.InternalError
			}
		} else {
			shardAnalytics, err = db.GroupBy(params.Property, params.TimeRange)
			if err != nil {
				return nil, &errors.InternalError
			}
		}

		// Unlock DB
		driver.DBManager.UnlockDB <- NewUnlock(shardPath)

		// Add shard result to analyticsMap
		for _, analytic := range shardAnalytics.List {
			if total, ok := analyticsMap[analytic.Id]; ok {
				total.Total += analytic.Total
				total.Unique += analytic.Unique
				analyticsMap[analytic.Id] = total
			} else {
				analyticsMap[analytic.Id] = analytic
			}
		}
	}

	// Convert analyticsMap to an Aggregates struct
	for _, analytic := range analyticsMap {
		analytics.List = append(analytics.List, analytic)
	}

	// // If value is in Cache, return directly
	// cached, inCache := driver.DBManager.Cache.Get(params.URL)
	// if inCache {
	// 	if response, ok := cached.(*structures.Aggregates); ok {
	// 		driver.DBManager.UnlockDB <- dbPath
	// 		return response, nil
	// 	}
	// }

	// // Store response in Cache before sending
	// driver.DBManager.Cache.Add(params.URL, analytics)

	return &analytics, nil
}

func (driver *Sharded) OverTime(params structures.Params) (*structures.Intervals, error) {
	// Construct DBPath
	dbPath := DBPath{
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

	// At this point, there should be shards to query
	// Get list of shards by reading directory
	shards := ListShards(dbPath)

	// Aggregated query result
	analytics := structures.Intervals{}

	// Read from each shard
	for _, shard := range shards {
		// Don't include shard if not in timerange
		shardInt, err := strconv.Atoi(shard)
		if err != nil {
			driver.logger.Error("Error [%v] converting shard %s name to an integer", err, shard)
		}

		startInt, endInt := params.TimeRange.ConvertToInt()
		if shardInt < startInt || shardInt > endInt {
			continue
		}

		// Construct each shard DBPath
		shardPath := DBPath{
			Name:      shard,
			Directory: dbPath.String(),
		}

		// Get DB shard from manager
		driver.DBManager.RequestDB <- shardPath
		db := <-driver.DBManager.SendDB

		var shardAnalytics *structures.Intervals

		// Check for unique query parameter to call function accordingly
		if params.Unique {
			shardAnalytics, err = db.OverTimeUniq(params.Interval, params.TimeRange)
			if err != nil {
				return nil, &errors.InternalError
			}
		} else {
			shardAnalytics, err = db.OverTime(params.Interval, params.TimeRange)
			if err != nil {
				return nil, &errors.InternalError
			}
		}

		// Unlock DB
		driver.DBManager.UnlockDB <- NewUnlock(shardPath)

		// Add shard result to analyticsMap
		for _, analytic := range shardAnalytics.List {
			analytics.List = append(analytics.List, analytic)
		}
	}

	// // If value is in Cache, return directly
	// cached, inCache := driver.DBManager.Cache.Get(params.URL)
	// if inCache {
	// 	if response, ok := cached.(*structures.Intervals); ok {
	// 		driver.DBManager.UnlockDB <- dbPath
	// 		return response, nil
	// 	}
	// }

	// // Store response in Cache before sending
	// driver.DBManager.Cache.Add(params.URL, analytics)

	return &analytics, nil
}

func (driver *Sharded) Push(params structures.Params, analytic structures.Analytic) error {
	// Construct DBPath
	dbPath := DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Push to right shard based on analytic time
	shardName := TimeToShard(analytic.Time)

	// Construct shard DBPath
	shardPath := DBPath{
		Name:      shardName,
		Directory: dbPath.String(),
	}

	// Get DB from manager
	driver.DBManager.RequestDB <- shardPath
	db := <-driver.DBManager.SendDB

	// Insert data if everything's OK
	err := db.Insert(analytic)

	// Unlock DB
	driver.DBManager.UnlockDB <- NewUnlock(shardPath)

	if err != nil {
		return &errors.InsertFailed
	}

	return nil
}

func (driver *Sharded) Delete(params structures.Params) error {
	// Construct DBPath
	dbPath := DBPath{
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

// Convert a time to a shard name
// 2015-12-08T00:00:00.000Z -> 201512
func TimeToShard(timeValue time.Time) string {
	layout := "200601"
	return timeValue.Format(layout)
}
