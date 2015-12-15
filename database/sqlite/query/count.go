package query

import (
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/GitbookIO/micro-analytics/database"
)

// Wrapper for querying a Database struct
func Count(db *sql.DB, timeRange *database.TimeRange) (*database.Count, error) {
	// Query
	queryBuilder := sq.
		Select("COUNT(*) AS total", "COUNT(DISTINCT ip) AS uniqueCount").
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
	count := database.Count{}
	err = db.QueryRow(query).Scan(&count.Total, &count.Unique)
	if err != nil {
		return nil, err
	}

	return &count, nil
}
