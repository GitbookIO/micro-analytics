package database

import (
    "fmt"
    "log"
    "time"

    sq "github.com/Masterminds/squirrel"
    "github.com/GitbookIO/analytics/utils"
)

// Wrapper for querying a Database struct
func (db *Database) Query() ([]Analytic, error) {
    // Query
    // query := `SELECT time, type, path, ip, platform, refererDomain, countryCode FROM visits`
    query, _, err := sq.
        Select("time", "type", "path", "ip", "platform", "refererDomain", "countryCode").
        From("visits").
        ToSql()

    // Exec query
    rows, err := db.Conn.Query(query)
    if err != nil {
        log.Printf("[DBQuery] Error %v querying DB %s", err, db.Name)
        return nil, err
    }
    defer rows.Close()

    analytics := []Analytic{}
    for rows.Next() {
        analytic := Analytic{}
        rows.Scan(&analytic.Time,
            &analytic.Type,
            &analytic.Path,
            &analytic.Ip,
            &analytic.Platform,
            &analytic.RefererDomain,
            &analytic.CountryCode)
        analytics = append(analytics, analytic)
    }

    return analytics, nil
}

// Wrapper for querying a Database struct grouped by a property
func (db *Database) GroupBy(property string, timeRange *TimeRange) (*AggregateList, error) {
    // Query
    queryBuilder := sq.
        Select(property, "COUNT(*)").
        From("visits")

    // Add time constraints if timeRange provided
    if timeRange != nil {
        if !timeRange.Start.Equal(time.Time{}) {
            timeQuery := fmt.Sprintf("time > %d", timeRange.Start.Unix())
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
        log.Printf("[DBQueryByProp] Error %v building query for DB %s", err, db.Name)
        return nil, err
    }

    // Exec query
    rows, err := db.Conn.Query(query)
    if err != nil {
        log.Printf("[DBQueryByProp] Error %v querying DB %s", err, db.Name)
        return nil, err
    }
    defer rows.Close()

    list := AggregateList{}
    for rows.Next() {
        aggregate := Aggregate{}
        rows.Scan(&aggregate.Id, &aggregate.Count)

        // For countries, get fullname as Label
        if property == "countryCode" {
            aggregate.Label = utils.GetCountry(aggregate.Id)
        } else {
            aggregate.Label = aggregate.Id
        }

        // Get unique count for property
        unique, err := db.CountUniqueWhere(property, aggregate.Id, timeRange)
        if err != nil {
            log.Printf("[DBQueryByProp] Error %v getting unique COUNT for %s with DB %s", err, property, db.Name)
            unique = 0
        }
        aggregate.Unique = unique

        list.List = append(list.List, aggregate)
    }

    return &list, nil
}

// Wrapper for querying a Database struct grouped by a property
func (db *Database) GroupByUniq(property string, timeRange *TimeRange) (*AggregateList, error) {
    // Subquery for counting unique IPs
    subqueryBuilder := sq.
        Select(property, "COUNT(DISTINCT ip) AS uniqueCount").
        From("visits")

    // Add time constraints if timeRange provided
    if timeRange != nil {
        if !timeRange.Start.Equal(time.Time{}) {
            timeQuery := fmt.Sprintf("time > %d", timeRange.Start.Unix())
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
        log.Printf("[DBQueryByProp] Error %v building subquery for DB %s", err, db.Name)
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
            timeQuery := fmt.Sprintf("time > %d", timeRange.Start.Unix())
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
        log.Printf("[DBQueryByProp] Error %v building query for DB %s", err, db.Name)
        return nil, err
    }

    // Exec query
    rows, err := db.Conn.Query(query)
    if err != nil {
        log.Printf("[DBQueryByProp] Error %v querying DB %s", err, db.Name)
        return nil, err
    }
    defer rows.Close()

    // Format results
    list := AggregateList{}
    for rows.Next() {
        aggregate := Aggregate{}
        rows.Scan(&aggregate.Id, &aggregate.Count, &aggregate.Unique)

        // For countries, get fullname as Label
        if property == "countryCode" {
            aggregate.Label = utils.GetCountry(aggregate.Id)
        } else {
            aggregate.Label = aggregate.Id
        }

        list.List = append(list.List, aggregate)
    }

    return &list, nil
}

// Wrapper for querying a Database struct over a time interval
func (db *Database) OverTimeUniq(interval int, timeRange *TimeRange) (*Intervals, error) {
    // Subquery for counting unique IPs
    subqueryBuilder := sq.
        Select(fmt.Sprintf("(time / %d) * %d AS sqStartTime", interval, interval), "COUNT(DISTINCT ip) AS uniqueCount").
        From("visits")

    // Add time constraints if timeRange provided
    if timeRange != nil {
        if !timeRange.Start.Equal(time.Time{}) {
            timeQuery := fmt.Sprintf("time > %d", timeRange.Start.Unix())
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
        log.Printf("[DBQueryByProp] Error %v building subquery for DB %s", err, db.Name)
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
            timeQuery := fmt.Sprintf("time > %d", timeRange.Start.Unix())
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
        log.Printf("[DBQueryByProp] Error %v building query for DB %s", err, db.Name)
        return nil, err
    }

    // Exec query
    rows, err := db.Conn.Query(query)
    if err != nil {
        log.Printf("[DBQueryByProp] Error %v querying DB %s", err, db.Name)
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
        result.End = time.Unix(int64(startTime + interval), 0).Format(time.RFC3339)

        intervals.List = append(intervals.List, result)
    }

    return &intervals, nil
}

// Wrapper for querying a Database struct grouped by a property for which IP is unique
func (db *Database) CountUniqueWhere(property string, value string, timeRange *TimeRange) (int, error) {
    // Query
    queryBuilder := sq.
        Select("COUNT(DISTINCT ip)").
        From("visits")

    // Add time constraints if timeRange provided
    if timeRange != nil {
        if !timeRange.Start.Equal(time.Time{}) {
            timeQuery := fmt.Sprintf("time > %d", timeRange.Start.Unix())
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
        log.Printf("[DBQueryByProp] Error %v building query for DB %s", err, db.Name)
        return 0, err
    }

    // Exec query
    rows, err := db.Conn.Query(query)
    if err != nil {
        log.Printf("[DBQueryByProp] Error %v querying DB %s", err, db.Name)
        return 0, err
    }
    defer rows.Close()

    result := 0
    for rows.Next() {
        rows.Scan(&result)
    }

    return result, nil
}

// Wrapper for querying a Database struct over a time interval
func (db *Database) OverTime(interval int, timeRange *TimeRange) (*Intervals, error) {
    // Query
    queryBuilder := sq.
        Select(fmt.Sprintf("(time / %d) * %d AS startTime", interval, interval), "COUNT(*)").
        From("visits")

    // Add time constraints if timeRange provided
    if timeRange != nil {
        if !timeRange.Start.Equal(time.Time{}) {
            timeQuery := fmt.Sprintf("time > %d", timeRange.Start.Unix())
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
        log.Printf("[DBQueryByProp] Error %v building query for DB %s", err, db.Name)
        return nil, err
    }

    // Exec query
    rows, err := db.Conn.Query(query)
    if err != nil {
        log.Printf("[DBQueryByProp] Error %v querying DB %s", err, db.Name)
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
        result.End = time.Unix(int64(startTime + interval), 0).Format(time.RFC3339)

        intervals.List = append(intervals.List, result)
    }

    return &intervals, nil
}