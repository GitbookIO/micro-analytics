package database

type Aggregate struct {
    Id      string  `json:"id"`
    Label   string  `json:"label"`
    Count   int     `json:"count"`
    Unique  int     `json:"unique"`
}

type AggregateList struct {
    List    []Aggregate  `json:"list"`
}