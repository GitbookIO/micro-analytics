package query

import (
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/GitbookIO/micro-analytics/database"
)

// Wrapper for querying a Database struct
func Query(db *sql.DB, timeRange *database.TimeRange) (*database.Analytics, error) {
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

	analytics := database.Analytics{}
	for rows.Next() {
		analytic := database.Analytic{}
		var analyticTime int64
		rows.Scan(&analyticTime,
			&analytic.Event,
			&analytic.Path,
			&analytic.Ip,
			&analytic.Platform,
			&analytic.RefererDomain,
			&analytic.CountryCode)

		analytic.Time = time.Unix(analyticTime, 0).UTC()
		analytics.List = append(analytics.List, analytic)
	}

	return &analytics, nil
}
