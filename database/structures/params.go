package structures

type Params struct {
	DBName    string
	Interval  int
	Property  string
	TimeRange *TimeRange
	Unique    bool
	URL       string
}
