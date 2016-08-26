package query

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"github.com/GitbookIO/micro-analytics/database"
)

// Wrapper for inserting through a Database struct
func BulkInsert(db *sql.DB, analytics []database.Analytic) error {
	// Base query
	insertQuery := sq.
		Insert("visits").
		Columns("time", "event", "path", "ip", "platform", "refererDomain", "countryCode")

	// Add values for each analytic object
	for _, analytic := range analytics {
		insertQuery = insertQuery.Values(
			analytic.Time.Unix(),
			analytic.Event,
			analytic.Path,
			analytic.Ip,
			analytic.Platform,
			analytic.RefererDomain,
			analytic.CountryCode)
	}

	// Add database
	insertQuery = insertQuery.RunWith(db)

	_, err := insertQuery.Exec()
	if err != nil {
		return err
	}

	return nil
}
