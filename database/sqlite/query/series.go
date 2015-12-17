package query

import (
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/GitbookIO/micro-analytics/database"
)

// Wrapper for querying a Database struct over a time interval
func Series(db *sql.DB, interval int, timeRange *database.TimeRange) (*database.Intervals, error) {
	// Query
	queryBuilder := sq.
		Select(fmt.Sprintf("(time / %d) * %d AS startTime", interval, interval), "COUNT(*)").
		From("visits")

	// Add time constraints if timeRange provided
	if timeRange != nil {
		if !timeRange.Start.Equal(time.Time{}) {
			timeQuery := fmt.Sprintf("time >= %d", timeRange.Start.Unix())
			queryBuilder = queryBuilder.Where(timeQuery)
		}
		if !timeRange.End.Equal(time.Time{}) {
			timeQuery := fmt.Sprintf("time <= %d", timeRange.End.Unix())
			queryBuilder = queryBuilder.Where(timeQuery)
		}
	}

	// Set query Group By condition
	query, _, err := queryBuilder.GroupBy("startTime").ToSql()
	if err != nil {
		return nil, err
	}

	// Exec query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Format results
	intervals := database.Intervals{}
	for rows.Next() {
		result := database.Interval{}
		var startTime int

		rows.Scan(&startTime, &result.Total)

		// Format Start and End from TIMESTAMP to ISO time
		result.Start = time.Unix(int64(startTime), 0).UTC().Format(time.RFC3339)
		result.End = time.Unix(int64(startTime+interval), 0).UTC().Format(time.RFC3339)

		intervals.List = append(intervals.List, result)
	}

	return &intervals, nil
}

// Wrapper for querying a Database struct over a time interval
func SeriesUniq(db *sql.DB, interval int, timeRange *database.TimeRange) (*database.Intervals, error) {
	// Subquery for counting unique IPs
	subqueryBuilder := sq.
		Select(fmt.Sprintf("(time / %d) * %d AS sqStartTime", interval, interval), "COUNT(DISTINCT ip) AS uniqueCount").
		From("visits")

	// Add time constraints if timeRange provided
	if timeRange != nil {
		if !timeRange.Start.Equal(time.Time{}) {
			timeQuery := fmt.Sprintf("time >= %d", timeRange.Start.Unix())
			subqueryBuilder = subqueryBuilder.Where(timeQuery)
		}
		if !timeRange.End.Equal(time.Time{}) {
			timeQuery := fmt.Sprintf("time <= %d", timeRange.End.Unix())
			subqueryBuilder = subqueryBuilder.Where(timeQuery)
		}
	}

	// Format subquery
	subquery, _, err := subqueryBuilder.GroupBy("sqStartTime").ToSql()
	if err != nil {
		return nil, err
	}

	subquery = fmt.Sprintf("(%s) AS subquery", subquery)

	// Query
	joinClause := fmt.Sprintf("%s ON sqStartTime = startTime", subquery)
	queryBuilder := sq.
		Select(fmt.Sprintf("(time / %d) * %d AS startTime", interval, interval), "COUNT(*) AS total", "uniqueCount").
		From("visits").
		Join(joinClause)

	// Add time constraints if timeRange provided
	if timeRange != nil {
		if !timeRange.Start.Equal(time.Time{}) {
			timeQuery := fmt.Sprintf("time >= %d", timeRange.Start.Unix())
			queryBuilder = queryBuilder.Where(timeQuery)
		}
		if !timeRange.End.Equal(time.Time{}) {
			timeQuery := fmt.Sprintf("time <= %d", timeRange.End.Unix())
			queryBuilder = queryBuilder.Where(timeQuery)
		}
	}

	// Set query Group By condition
	query, _, err := queryBuilder.GroupBy("startTime").ToSql()
	if err != nil {
		return nil, err
	}

	// Exec query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Format results
	intervals := database.Intervals{}
	for rows.Next() {
		result := database.Interval{}
		var startTime int

		rows.Scan(&startTime, &result.Total, &result.Unique)

		// Format Start and End from TIMESTAMP to ISO time
		result.Start = time.Unix(int64(startTime), 0).UTC().Format(time.RFC3339)
		result.End = time.Unix(int64(startTime+interval), 0).UTC().Format(time.RFC3339)

		intervals.List = append(intervals.List, result)
	}

	return &intervals, nil
}
