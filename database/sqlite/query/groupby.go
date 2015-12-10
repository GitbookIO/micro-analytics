package query

import (
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/utils/geoip"
)

// Wrapper for querying a Database struct grouped by a property
func GroupBy(db *sql.DB, property string, timeRange *database.TimeRange) (*database.Aggregates, error) {
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

	list := database.Aggregates{}
	for rows.Next() {
		aggregate := database.Aggregate{}
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
func GroupByUniq(db *sql.DB, property string, timeRange *database.TimeRange) (*database.Aggregates, error) {
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
	list := database.Aggregates{}
	for rows.Next() {
		aggregate := database.Aggregate{}
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
