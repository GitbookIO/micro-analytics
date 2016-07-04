package database

type Driver interface {
	// Count number of stats
	Count(params Params) (*Count, error)
	// Return aggregated stats by property
	GroupBy(params Params) (*Aggregates, error)
	// Return time serie sliced by a specific interval
	Series(params Params) (*Intervals, error)
	// Return all stats
	Query(params Params) (*Analytics, error)
	// Handle adding new stats
	Insert(params Params, analytic Analytic) error
	// Handle bulk insert
	BulkInsert(analytics map[string][]Analytic) error
	// Handle DB removal
	Delete(params Params) error
}

type DriverOpts struct {
	Directory      string
	MaxDBs         int
	IdleTimeout    int
	CacheDirectory string
	ClosingChannel chan bool
}
