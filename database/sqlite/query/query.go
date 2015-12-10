package query

import (
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	. "github.com/GitbookIO/micro-analytics/database/structures"
	"github.com/GitbookIO/micro-analytics/utils/geoip"
)

// Wrapper for querying a Database struct
func Query(db *sql.DB, timeRange *TimeRange) (*Analytics, error) {
	// Query
	queryBuilder := sq.
		Select("time", "event", "path", "ip", "platform", "refererDomain", "countryCode").
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

	query, _, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	// Exec query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	analytics := Analytics{}
	for rows.Next() {
		analytic := Analytic{}
		var analyticTime int64
		rows.Scan(&analyticTime,
			&analytic.Event,
			&analytic.Path,
			&analytic.Ip,
			&analytic.Platform,
			&analytic.RefererDomain,
			&analytic.CountryCode)

		analytic.Time = time.Unix(analyticTime, 0)
		analytics.List = append(analytics.List, analytic)
	}

	return &analytics, nil
}

// Wrapper for querying a Database struct grouped by a property
func GroupBy(db *sql.DB, property string, timeRange *TimeRange) (*Aggregates, error) {
	// Query
	queryBuilder := sq.
		Select(property, "COUNT(*)").
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
	query, _, err := queryBuilder.GroupBy(property).ToSql()
	if err != nil {
		return nil, err
	}

	// Exec query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := Aggregates{}
	for rows.Next() {
		aggregate := Aggregate{}
		rows.Scan(&aggregate.Id, &aggregate.Total)

		// For countries, get fullname as Label
		if property == "countryCode" {
			aggregate.Label = geoip.GetCountry(aggregate.Id)
		} else {
			aggregate.Label = aggregate.Id
		}

		list.List = append(list.List, aggregate)
	}

	return &list, nil
}

// Wrapper for querying a Database struct grouped by a property
func GroupByUniq(db *sql.DB, property string, timeRange *TimeRange) (*Aggregates, error) {
	// Subquery for counting unique IPs
	subqueryBuilder := sq.
		Select(property, "COUNT(DISTINCT ip) AS uniqueCount").
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
	subquery, _, err := subqueryBuilder.GroupBy(property).ToSql()
	if err != nil {
		return nil, err
	}

	subquery = fmt.Sprintf("(%s) AS subquery", subquery)

	// Query
	tableProperty := fmt.Sprintf("visits.%s", property)
	subqueryProperty := fmt.Sprintf("subquery.%s", property)
	joinClause := fmt.Sprintf("%s ON %s = %s", subquery, tableProperty, subqueryProperty)

	queryBuilder := sq.
		Select(tableProperty, "COUNT(*) AS total", "uniqueCount").
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

	// Format query
	query, _, err := queryBuilder.GroupBy(tableProperty).ToSql()
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
	list := Aggregates{}
	for rows.Next() {
		aggregate := Aggregate{}
		rows.Scan(&aggregate.Id, &aggregate.Total, &aggregate.Unique)

		// For countries, get fullname as Label
		if property == "countryCode" {
			aggregate.Label = geoip.GetCountry(aggregate.Id)
		} else {
			aggregate.Label = aggregate.Id
		}

		list.List = append(list.List, aggregate)
	}

	return &list, nil
}

// Wrapper for querying a Database struct over a time interval
func OverTime(db *sql.DB, interval int, timeRange *TimeRange) (*Intervals, error) {
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
	intervals := Intervals{}
	for rows.Next() {
		result := Interval{}
		var startTime int

		rows.Scan(&startTime, &result.Total)

		// Format Start and End from TIMESTAMP to ISO time
		result.Start = time.Unix(int64(startTime), 0).Format(time.RFC3339)
		result.End = time.Unix(int64(startTime+interval), 0).Format(time.RFC3339)

		intervals.List = append(intervals.List, result)
	}

	return &intervals, nil
}

// Wrapper for querying a Database struct over a time interval
func OverTimeUniq(db *sql.DB, interval int, timeRange *TimeRange) (*Intervals, error) {
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
	intervals := Intervals{}
	for rows.Next() {
		result := Interval{}
		var startTime int

		rows.Scan(&startTime, &result.Total, &result.Unique)

		// Format Start and End from TIMESTAMP to ISO time
		result.Start = time.Unix(int64(startTime), 0).Format(time.RFC3339)
		result.End = time.Unix(int64(startTime+interval), 0).Format(time.RFC3339)

		intervals.List = append(intervals.List, result)
	}

	return &intervals, nil
}

// Wrapper for querying a Database struct grouped by a property for which IP is unique
func CountUniqueWhere(db *sql.DB, property string, value string, timeRange *TimeRange) (int, error) {
	// Query
	queryBuilder := sq.
		Select("COUNT(DISTINCT ip)").
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

	// Set condition on value
	queryBuilder = queryBuilder.Where(fmt.Sprintf("%s = '%s'", property, value))

	// Set query Group By condition
	query, _, err := queryBuilder.GroupBy(property).ToSql()
	if err != nil {
		return 0, err
	}

	// Exec query
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	result := 0
	for rows.Next() {
		rows.Scan(&result)
	}

	return result, nil
}
