package structures

import (
	"time"
)

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
