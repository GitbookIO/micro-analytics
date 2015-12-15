package database

import (
	"net/url"
	"time"
)

type Count struct {
	Total  int `json:"total"`
	Unique int `json:"unique"`
}

type Analytic struct {
	Time          time.Time `json:"time"`
	Event         string    `json:"event"`
	Path          string    `json:"path"`
	Ip            string    `json:"ip"`
	Platform      string    `json:"platform"`
	RefererDomain string    `json:"refererDomain"`
	CountryCode   string    `json:"countryCode"`
}

type Analytics struct {
	List []Analytic `json:"list"`
}

type Aggregate struct {
	Id     string `json:"id"`
	Label  string `json:"label"`
	Total  int    `json:"total"`
	Unique int    `json:"unique"`
}

type Aggregates struct {
	List []Aggregate `json:"list"`
}

type Interval struct {
	Start  string `json:"start"`
	End    string `json:"end"`
	Total  int    `json:"total"`
	Unique int    `json:"unique"`
}

type Intervals struct {
	List []Interval `json:"list"`
}

type Params struct {
	DBName    string
	Interval  int
	Property  string
	TimeRange *TimeRange
	Unique    bool
	URL       *url.URL
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}
