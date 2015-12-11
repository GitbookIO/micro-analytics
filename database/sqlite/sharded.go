package sqlite

import (
	"io/ioutil"
	"net/url"
	"strconv"
	"time"

	"github.com/azer/logger"
	"github.com/hashicorp/golang-lru"

	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/database/errors"

	"github.com/GitbookIO/micro-analytics/database/sqlite/manager"
	"github.com/GitbookIO/micro-analytics/database/sqlite/query"
)

type Sharded struct {
	DBManager *manager.DBManager
	directory string
	cache     *lru.Cache
	logger    *logger.Logger
}

func NewShardedDriver(driverOpts database.DriverOpts) (*Sharded, error) {
	manager := manager.New(manager.Opts{driverOpts})

	cache, err := lru.New(driverOpts.CacheSize)
	if err != nil {
		return nil, err
	}

	driver := &Sharded{
		DBManager: manager,
		directory: driverOpts.Directory,
		cache:     cache,
		logger:    logger.New("[Sharded]"),
	}

	return driver, nil
}

func (driver *Sharded) Query(params database.Params) (*database.Analytics, error) {
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

	// At this point, there should be shards to query
	// Get list of shards by reading directory
	shards := listShards(dbPath)
	analytics := database.Analytics{}

	// Read from each shard
	for _, shard := range shards {
		// Don't include shard if not in timerange
		shardInt, err := strconv.Atoi(shard)
		if err != nil {
			driver.logger.Error("Error [%v] converting shard %s name to an integer", err, shard)
		}

		startInt, endInt := convertToInt(params.TimeRange)
		if shardInt < startInt || shardInt > endInt {
			continue
		}

		// Construct each shard DBPath
		shardPath := manager.DBPath{
			Name:      shard,
			Directory: dbPath.String(),
		}

		// Get DB shard from manager
		db, err := driver.DBManager.GetDB(shardPath)
		if err != nil {
			return nil, &errors.InternalError
		}

		// Return query result
		shardAnalytics, err := query.Query(db.Conn, params.TimeRange)
		if err != nil {
			return nil, &errors.InternalError
		}

		// Add shard result to analytics
		for _, analytic := range shardAnalytics.List {
			analytics.List = append(analytics.List, analytic)
		}
	}

	// // If value is in Cache, return directly
	// cached, inCache := driver.DBManager.Cache.Get(params.URL)
	// if inCache {
	// 	if response, ok := cached.(*database.Analytics); ok {
	// 		driver.DBManager.UnlockDB <- dbPath
	// 		return response, nil
	// 	}
	// }

	// // Store response in Cache before sending
	// driver.DBManager.Cache.Add(params.URL, analytics)

	return &analytics, nil
}

func (driver *Sharded) GroupBy(params database.Params) (*database.Aggregates, error) {
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

	// At this point, there should be shards to query
	// Get list of shards by reading directory
	shards := listShards(dbPath)

	// Aggregated query result
	analytics := database.Aggregates{}
	// Helper map to aggregate
	analyticsMap := map[string]database.Aggregate{}

	// Read from each shard
	for _, shard := range shards {
		// Don't include shard if not in timerange
		shardInt, err := strconv.Atoi(shard)
		if err != nil {
			driver.logger.Error("Error [%v] converting shard %s name to an integer", err, shard)
		}

		startInt, endInt := convertToInt(params.TimeRange)
		if shardInt < startInt || shardInt > endInt {
			continue
		}

		// Get result if is cached
		var shardAnalytics *database.Aggregates

		cached, inCache := driver.cache.Get(formatURLForCache(params.URL, shardInt, startInt, endInt))
		if inCache {
			var ok bool
			if shardAnalytics, ok = cached.(*database.Aggregates); !ok {
				return nil, &errors.InternalError
			}
		} else {
			// Else query shard
			// Construct each shard DBPath
			shardPath := manager.DBPath{
				Name:      shard,
				Directory: dbPath.String(),
			}

			// Get DB shard from manager
			db, err := driver.DBManager.GetDB(shardPath)
			if err != nil {
				return nil, &errors.InternalError
			}

			// Check for unique query parameter to call function accordingly
			if params.Unique {
				shardAnalytics, err = query.GroupByUniq(db.Conn, params.Property, params.TimeRange)
				if err != nil {
					return nil, &errors.InternalError
				}
			} else {
				shardAnalytics, err = query.GroupBy(db.Conn, params.Property, params.TimeRange)
				if err != nil {
					return nil, &errors.InternalError
				}
			}

			// Set shard result in cache if asked
			driver.cache.Add(formatURLForCache(params.URL, shardInt, startInt, endInt))
		}

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

	return &analytics, nil
}

func (driver *Sharded) Series(params database.Params) (*database.Intervals, error) {
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

	// At this point, there should be shards to query
	// Get list of shards by reading directory
	shards := listShards(dbPath)

	// Aggregated query result
	analytics := database.Intervals{}

	// Read from each shard
	for _, shard := range shards {
		// Don't include shard if not in timerange
		shardInt, err := strconv.Atoi(shard)
		if err != nil {
			driver.logger.Error("Error [%v] converting shard %s name to an integer", err, shard)
		}

		startInt, endInt := convertToInt(params.TimeRange)
		if shardInt < startInt || shardInt > endInt {
			continue
		}

		// Construct each shard DBPath
		shardPath := manager.DBPath{
			Name:      shard,
			Directory: dbPath.String(),
		}

		// Get DB shard from manager
		db, err := driver.DBManager.GetDB(shardPath)
		if err != nil {
			return nil, &errors.InternalError
		}

		var shardAnalytics *database.Intervals

		// Check for unique query parameter to call function accordingly
		if params.Unique {
			shardAnalytics, err = query.SeriesUniq(db.Conn, params.Interval, params.TimeRange)
			if err != nil {
				return nil, &errors.InternalError
			}
		} else {
			shardAnalytics, err = query.Series(db.Conn, params.Interval, params.TimeRange)
			if err != nil {
				return nil, &errors.InternalError
			}
		}

		// Add shard result to analyticsMap
		for _, analytic := range shardAnalytics.List {
			analytics.List = append(analytics.List, analytic)
		}
	}

	// // If value is in Cache, return directly
	// cached, inCache := driver.DBManager.Cache.Get(params.URL)
	// if inCache {
	// 	if response, ok := cached.(*database.Intervals); ok {
	// 		driver.DBManager.UnlockDB <- dbPath
	// 		return response, nil
	// 	}
	// }

	// // Store response in Cache before sending
	// driver.DBManager.Cache.Add(params.URL, analytics)

	return &analytics, nil
}

func (driver *Sharded) Insert(params database.Params, analytic database.Analytic) error {
	// Construct DBPath
	dbPath := manager.DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Push to right shard based on analytic time
	shardName := timeToShard(analytic.Time)

	// Construct shard DBPath
	shardPath := manager.DBPath{
		Name:      shardName,
		Directory: dbPath.String(),
	}

	// Get DB from manager
	driver.logger.Info("Request for DB %s", shardPath)
	db, err := driver.DBManager.GetDB(shardPath)
	if err != nil {
		return &errors.InternalError
	}

	driver.logger.Info("Inserting in DB %s", shardPath)
	// Insert data if everything's OK
	err = query.Insert(db.Conn, analytic)

	if err != nil {
		return &errors.InsertFailed
	}

	return nil
}

func (driver *Sharded) Delete(params database.Params) error {
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

// Convert a time to a shard name
// 2015-12-08T00:00:00.000Z -> 201512
func timeToShard(timeValue time.Time) string {
	layout := "200601"
	return timeValue.Format(layout)
}

func listShards(dbPath manager.DBPath) []string {
	folders, err := ioutil.ReadDir(dbPath.String())
	if err != nil {
		return nil
	}

	shards := make([]string, 0)
	for _, folder := range folders {
		shards = append(shards, folder.Name())
	}

	return shards
}

// Helper function to return start and end time as an int in YYYYMM format
// Defaults to 0 for Start and 999999 for End
func convertToInt(timeRange *database.TimeRange) (int, int) {
	var err error
	layout := "200601"

	startDefault := 0
	startInt := 0
	endDefault := 999999
	endInt := 999999

	if timeRange != nil {
		if !timeRange.Start.Equal(time.Time{}) {
			startInt, err = strconv.Atoi(timeRange.Start.Format(layout))
			if err != nil {
				startInt = startDefault
			}
		}
		if !timeRange.End.Equal(time.Time{}) {
			endInt, err = strconv.Atoi(timeRange.End.Format(layout))
			if err != nil {
				endInt = endDefault
			}
		}
	}

	return startInt, endInt
}

// Format URL for a specific shard
// Basically, remove start/end if is is before/after shard time
func formatURLForCache(rawURL string, shardName int, startMonth int, endMonth int) (string, error) {
	// Parse URL and get query parameters
	uRL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	queryParams, err := uRL.Query()
	if err != nil {
		return nil, err
	}

	// Remove start
	if startMonth < shardName {
		queryParams.Del("start")
	}

	// Remove end
	if endMonth > shardName {
		queryParams.Del("end")
	}

	// Remove cache for months before current month
	currentMonth := timeToShard(time.Now())
	if shardName < currentMonth {
		queryParams.Del("cache")
	}

	uRL.RawQuery = queryParams.Encode()
	return uRL.String()
}

var _ database.Driver = &Sharded{}
