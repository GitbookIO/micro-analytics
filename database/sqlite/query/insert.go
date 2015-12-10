package query

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	. "github.com/GitbookIO/micro-analytics/database/structures"
)

// Wrapper for inserting through a Database struct
func Insert(db *sql.DB, analytic Analytic) error {
	insertQuery := sq.
		Insert("visits").
		Columns("time", "event", "path", "ip", "platform", "refererDomain", "countryCode").
		Values(analytic.Time.Unix(),
		analytic.Event,
		analytic.Path,
		analytic.Ip,
		analytic.Platform,
		analytic.RefererDomain,
		analytic.CountryCode).
		RunWith(db)

	_, err := insertQuery.Exec()
	if err != nil {
		return err
	}

	return nil
}
