package sqlite

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GitbookIO/diskache"
	"github.com/GitbookIO/go-sqlpool"

	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/database/errors"

	"github.com/GitbookIO/micro-analytics/database/sqlite/manager"
	"github.com/GitbookIO/micro-analytics/database/sqlite/query"
)

type Sharded struct {
	DBManager *manager.DBManager
	directory string
	cache     *diskache.Diskache
}

func NewShardedDriver(driverOpts database.DriverOpts) (*Sharded, error) {
	manager := manager.New(manager.Opts{driverOpts})

	cacheOpts := &diskache.Opts{
		Directory: driverOpts.CacheDirectory,
	}
	cache, err := diskache.New(cacheOpts)
	if err != nil {
		return nil, err
	}

	driver := &Sharded{
		DBManager: manager,
		directory: driverOpts.Directory,
		cache:     cache,
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
		driver.DBManager.Logger.Error("Error executing Query/DBExists on DB %s: %v\n", dbPath, err)
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
	cachedRequest := cachedRequest(params.URL)

	// Read from each shard
	for _, shardName := range shards {

		// Don't include shard if not in timerange
		shardInt, err := shardNameToInt(shardName)
		if err != nil {
			return nil, err
		}

		startInt, endInt := timeRangeToInt(params.TimeRange)
		if shardInt < startInt || shardInt > endInt {
			continue
		}

		// Get result if is cached
		var shardAnalytics *database.Analytics

		cacheURL, err := formatURLForCache(params.URL, shardInt, startInt, endInt, params.TimeRange)
		if err != nil {
			return nil, err
		}

		cached, inCache := driver.cache.Get(cacheURL)
		if inCache {
			err = json.Unmarshal(cached, &shardAnalytics)
			if err != nil {
				driver.DBManager.Logger.Error("Error unmarshaling from cache: %v\n", err)
				return nil, err
			}
		} else {
			// Else query shard
			// Construct each shard DBPath
			shardPath := manager.DBPath{
				Name:      shardName,
				Directory: dbPath.String(),
			}

			// Get DB shard from manager
			db, err := driver.DBManager.Acquire(shardPath)
			if err != nil {
				driver.DBManager.Logger.Error("Error executing Query/Acquire on DB %s: %v\n", shardPath, err)
				return nil, &errors.InternalError
			}
			defer driver.DBManager.Release(db)

			// Return query result
			shardAnalytics, err = query.Query(db.DB, params.TimeRange)
			if err != nil {
				driver.DBManager.Logger.Error("Error executing Query on DB %s: %v\n", shardPath, err)
				return nil, &errors.InternalError
			}

			// Set shard result in cache if asked
			if cachedRequest {
				if data, err := json.Marshal(shardAnalytics); err == nil {
					err = driver.cache.Set(cacheURL, data)
					if err != nil {
						driver.DBManager.Logger.Error("Error adding to cache: %v\n", err)
					}
				}
			}
		}

		// Add shard result to analytics
		for _, analytic := range shardAnalytics.List {
			analytics.List = append(analytics.List, analytic)
		}
	}

	return &analytics, nil
}

func (driver *Sharded) Count(params database.Params) (*database.Count, error) {
	// Construct DBPath
	dbPath := manager.DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Check if DB file exists
	dbExists, err := driver.DBManager.DBExists(dbPath)
	if err != nil {
		driver.DBManager.Logger.Error("Error executing Count/DBExists on DB %s: %v\n", dbPath, err)
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
	analytics := database.Count{}

	cachedRequest := cachedRequest(params.URL)

	// Read from each shard
	for _, shardName := range shards {
		// Don't include shard if not in timerange
		shardInt, err := shardNameToInt(shardName)
		if err != nil {
			return nil, err
		}

		startInt, endInt := timeRangeToInt(params.TimeRange)
		if shardInt < startInt || shardInt > endInt {
			continue
		}

		// Get result if is cached
		var shardAnalytics *database.Count

		cacheURL, err := formatURLForCache(params.URL, shardInt, startInt, endInt, params.TimeRange)
		if err != nil {
			return nil, err
		}

		cached, inCache := driver.cache.Get(cacheURL)
		if inCache {
			err = json.Unmarshal(cached, &shardAnalytics)
			if err != nil {
				driver.DBManager.Logger.Error("Error unmarshaling from cache: %v\n", err)
				return nil, err
			}
		} else {
			// Else query shard
			// Construct each shard DBPath
			shardPath := manager.DBPath{
				Name:      shardName,
				Directory: dbPath.String(),
			}

			// Get DB shard from manager
			db, err := driver.DBManager.Acquire(shardPath)
			if err != nil {
				driver.DBManager.Logger.Error("Error executing Count/Acquire on DB %s: %v\n", shardPath, err)
				return nil, &errors.InternalError
			}
			defer driver.DBManager.Release(db)

			// Launch query
			shardAnalytics, err = query.Count(db.DB, params.TimeRange)
			if err != nil {
				driver.DBManager.Logger.Error("Error executing Count on DB %s: %v\n", shardPath, err)
				return nil, &errors.InternalError
			}

			// Set shard result in cache if asked
			if cachedRequest {
				if data, err := json.Marshal(shardAnalytics); err == nil {
					err = driver.cache.Set(cacheURL, data)
					if err != nil {
						driver.DBManager.Logger.Error("Error adding to cache: %v\n", err)
					}
				}
			}
		}

		// Add shard result to main result
		analytics.Total += shardAnalytics.Total
		analytics.Unique += shardAnalytics.Unique
	}

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
		driver.DBManager.Logger.Error("Error executing GroupBy/DBExists on DB %s: %v\n", dbPath, err)
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

	cachedRequest := cachedRequest(params.URL)

	// Read from each shard
	for _, shardName := range shards {
		// Don't include shard if not in timerange
		shardInt, err := shardNameToInt(shardName)
		if err != nil {
			return nil, err
		}

		startInt, endInt := timeRangeToInt(params.TimeRange)
		if shardInt < startInt || shardInt > endInt {
			continue
		}

		// Get result if is cached
		var shardAnalytics *database.Aggregates

		cacheURL, err := formatURLForCache(params.URL, shardInt, startInt, endInt, params.TimeRange)
		if err != nil {
			return nil, err
		}

		cached, inCache := driver.cache.Get(cacheURL)
		if inCache {
			err = json.Unmarshal(cached, &shardAnalytics)
			if err != nil {
				driver.DBManager.Logger.Error("Error unmarshaling from cache: %v\n", err)
				return nil, err
			}
		} else {
			// Else query shard
			// Construct each shard DBPath
			shardPath := manager.DBPath{
				Name:      shardName,
				Directory: dbPath.String(),
			}

			// Get DB shard from manager
			db, err := driver.DBManager.Acquire(shardPath)
			if err != nil {
				driver.DBManager.Logger.Error("Error executing GroupBy/Acquire on DB %s: %v\n", shardPath, err)
				return nil, &errors.InternalError
			}
			defer driver.DBManager.Release(db)

			// Check for unique query parameter to call function accordingly
			if params.Unique {
				shardAnalytics, err = query.GroupByUniq(db.DB, params.Property, params.TimeRange)
				if err != nil {
					driver.DBManager.Logger.Error("Error executing GroupByUniq on DB %s: %v\n", shardPath, err)
					return nil, &errors.InternalError
				}
			} else {
				shardAnalytics, err = query.GroupBy(db.DB, params.Property, params.TimeRange)
				if err != nil {
					driver.DBManager.Logger.Error("Error executing GroupBy on DB %s: %v\n", shardPath, err)
					return nil, &errors.InternalError
				}
			}

			// Set shard result in cache if asked
			if cachedRequest {
				if data, err := json.Marshal(shardAnalytics); err == nil {
					err = driver.cache.Set(cacheURL, data)
					if err != nil {
						driver.DBManager.Logger.Error("Error adding to cache: %v\n", err)
					}
				}
			}
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

	// Convert analyticsMap to an AggregateList struct
	aggregateList := database.AggregateList{}
	for _, analytic := range analyticsMap {
		aggregateList = append(aggregateList, analytic)
	}

	// Sort and set
	sort.Sort(aggregateList)
	analytics.List = aggregateList

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
		driver.DBManager.Logger.Error("Error executing Series/DBExists on DB %s: %v\n", dbPath, err)
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

	cachedRequest := cachedRequest(params.URL)

	// Read from each shard
	for _, shardName := range shards {
		// Don't include shard if not in timerange
		shardInt, err := shardNameToInt(shardName)
		if err != nil {
			return nil, err
		}

		startInt, endInt := timeRangeToInt(params.TimeRange)
		if shardInt < startInt || shardInt > endInt {
			continue
		}

		// Get result if is cached
		var shardAnalytics *database.Intervals

		cacheURL, err := formatURLForCache(params.URL, shardInt, startInt, endInt, params.TimeRange)
		if err != nil {
			return nil, err
		}

		cached, inCache := driver.cache.Get(cacheURL)
		if inCache {
			err = json.Unmarshal(cached, &shardAnalytics)
			if err != nil {
				driver.DBManager.Logger.Error("Error unmarshaling from cache: %v\n", err)
				return nil, err
			}
		} else {
			// Else query shard
			// Construct each shard DBPath
			shardPath := manager.DBPath{
				Name:      shardName,
				Directory: dbPath.String(),
			}

			// Get DB shard from manager
			db, err := driver.DBManager.Acquire(shardPath)
			if err != nil {
				driver.DBManager.Logger.Error("Error executing Series/Acquire on DB %s: %v\n", shardPath, err)
				return nil, &errors.InternalError
			}
			defer driver.DBManager.Release(db)

			// Check for unique query parameter to call function accordingly
			if params.Unique {
				shardAnalytics, err = query.SeriesUniq(db.DB, params.Interval, params.TimeRange)
				if err != nil {
					driver.DBManager.Logger.Error("Error executing SeriesUniq on DB %s: %v\n", shardPath, err)
					return nil, &errors.InternalError
				}
			} else {
				shardAnalytics, err = query.Series(db.DB, params.Interval, params.TimeRange)
				if err != nil {
					driver.DBManager.Logger.Error("Error executing Series on DB %s: %v\n", shardPath, err)
					return nil, &errors.InternalError
				}
			}

			// Set shard result in cache if asked
			if cachedRequest {
				if data, err := json.Marshal(shardAnalytics); err == nil {
					err = driver.cache.Set(cacheURL, data)
					if err != nil {
						driver.DBManager.Logger.Error("Error adding to cache: %v\n", err)
					}
				}
			}
		}

		// Add shard result to analyticsMap
		for _, analytic := range shardAnalytics.List {
			analytics.List = append(analytics.List, analytic)
		}
	}

	// Merge time series by Start and End date
	analytics.Merge()
	return &analytics, nil
}

func (driver *Sharded) Insert(params database.Params, analytic database.Analytic) error {
	// Construct DBPath
	dbPath := manager.DBPath{
		Name:      params.DBName,
		Directory: driver.directory,
	}

	// Push to right shard based on analytic time
	shardName := timeToShardName(analytic.Time)

	// Construct shard DBPath
	shardPath := manager.DBPath{
		Name:      shardName,
		Directory: dbPath.String(),
	}

	// Get DB from manager
	db, err := driver.DBManager.Acquire(shardPath)
	if err != nil {
		driver.DBManager.Logger.Error("Error executing Insert/Acquire on DB %s: %v\n", shardPath, err)
		return &errors.InternalError
	}
	defer driver.DBManager.Release(db)

	// Insert data if everything's OK
	err = query.Insert(db.DB, analytic)

	if err != nil {
		driver.DBManager.Logger.Error("Error executing Insert on DB %s: %v\n", shardPath, err)
		return &errors.InsertFailed
	}

	return nil
}

func (driver *Sharded) BulkInsert(analytics map[string][]database.Analytic) error {
	var acquireErr, insertErr error
	var db *sqlpool.Resource
	// Run a bulk insert query for each database
	for dbName, _analytics := range analytics {
		// Group database analytics by shards
		shardedAnalytics := make(map[string][]database.Analytic)
		for _, analytic := range _analytics {
			shardName := timeToShardName(analytic.Time)
			shardedAnalytics[shardName] = append(shardedAnalytics[shardName], analytic)
		}

		// Run a bulk insert query for each shard
		for shardName, shardAnalytics := range shardedAnalytics {
			// Construct DBPath
			dbPath := manager.DBPath{
				Name:      dbName,
				Directory: driver.directory,
			}

			// Construct shard DBPath
			shardPath := manager.DBPath{
				Name:      shardName,
				Directory: dbPath.String(),
			}

			// Get DB from manager
			db, acquireErr = driver.DBManager.Acquire(shardPath)
			if acquireErr != nil {
				driver.DBManager.Logger.Error("Error executing Insert/Acquire on DB %s: %v\n", shardPath, acquireErr)
				continue
			}
			defer driver.DBManager.Release(db)

			// Insert data if everything's OK
			insertErr = query.BulkInsert(db.DB, shardAnalytics)
			if insertErr != nil {
				driver.DBManager.Logger.Error("Error executing Insert on DB %s: %v\n", shardPath, insertErr)
			}
		}
	}

	if insertErr != nil {
		return &errors.InsertFailed
	}
	if acquireErr != nil {
		return &errors.InternalError
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
		driver.DBManager.Logger.Error("Error executing Delete/DBExists on DB %s: %v\n", dbPath, err)
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
// 2015-12-08T00:00:00.000Z -> 2015-12
func timeToShardName(timeValue time.Time) string {
	layout := "2006-01"
	return timeValue.Format(layout)
}

// Convert a shard name to an int
// 2015-12 -> 201512
func shardNameToInt(shardName string) (int, error) {
	parts := strings.Split(shardName, "-")
	shardName = strings.Join(parts, "")
	shardInt, err := strconv.Atoi(shardName)
	return shardInt, err
}

// Return the list of all shards in a DBPath
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
func timeRangeToInt(timeRange *database.TimeRange) (int, int) {
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
func formatURLForCache(uRL *url.URL, shardName int, startMonth int, endMonth int, timeRange *database.TimeRange) (string, error) {
	// Extract URL query parameters
	queryParams := uRL.Query()

	// Remove start
	if startMonth < shardName {
		queryParams.Del("start")
	}

	// Remove start if timeRange.Start is the start of current month
	if startMonth == shardName && isStartOfMonth(timeRange.Start) {
		queryParams.Del("start")
	}

	// Remove end
	if endMonth > shardName {
		queryParams.Del("end")
	}

	// Remove cache for months before current month
	currentMonth, err := shardNameToInt(timeToShardName(time.Now().UTC()))
	if err != nil {
		return "", err
	}

	if shardName < currentMonth {
		queryParams.Del("cache")
	}

	// Add shard=shardName query parameter
	queryParams.Add("shard", strconv.Itoa(shardName))

	// Create new modified URL
	cacheURL := *uRL
	cacheURL.RawQuery = queryParams.Encode()

	return cacheURL.String(), nil
}

func isStartOfMonth(t time.Time) bool {
	y := t.Year()
	m := t.Month()
	d := t.Day()
	utcLoc, _ := time.LoadLocation("UTC")
	startOfMonth := time.Date(y, m, d, 0, 0, 0, 0, utcLoc)
	return t.Equal(startOfMonth)
}

// Return true if cache query parameter passed
func cachedRequest(uRL *url.URL) bool {
	// Extract query parameters
	queryParams := uRL.Query()
	return len(queryParams.Get("cache")) > 0
}

var _ database.Driver = &Sharded{}
