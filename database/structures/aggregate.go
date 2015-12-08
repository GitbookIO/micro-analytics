package structures

type Aggregate struct {
	Id     string `json:"id"`
	Label  string `json:"label"`
	Total  int    `json:"total"`
	Unique int    `json:"unique"`
}

type Aggregates struct {
	List []Aggregate `json:"list"`
}
