package structures

type Interval struct {
	Start  string `json:"start"`
	End    string `json:"end"`
	Total  int    `json:"total"`
	Unique int    `json:"unique"`
}

type Intervals struct {
	List []Interval `json:"list"`
}
