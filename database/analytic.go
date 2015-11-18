package database

import (
    "time"
)

type Analytic struct {
    Time            time.Time   `json:"time"`
    Type            string      `json:"type"`
    Path            string      `json:"path"`
    Ip              string      `json:"ip"`
    Platform        string      `json:"platform"`
    RefererDomain   string      `json:"refererDomain"`
    CountryCode     string      `json:"countryCode"`
}