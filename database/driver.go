package database

import (
	. "github.com/GitbookIO/micro-analytics/database/structures"
)

type Driver interface {
	// Return aggregated stats by property
	GroupBy(params Params) (*Aggregates, error)
	// Return time serie sliced by a specific interval
	OverTime(params Params) (*Intervals, error)
	// Return all stats
	Query(params Params) (*Analytics, error)
	// Handle adding new stats
	Push(params Params, analytic Analytic) error
	// Handle DB removal
	Delete(params Params) error
}

type DriverOpts struct {
	Directory      string
	MaxDBs         int
	CacheSize      int
	ClosingChannel chan bool
}
