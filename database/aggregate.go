package database

type Aggregate struct {
    Id      string  `json:"id"`
    Label   string  `json:"label"`
    Total   int     `json:"total"`
    Unique  int     `json:"unique"`
}

type AggregateList struct {
    List    []Aggregate  `json:"list"`
}