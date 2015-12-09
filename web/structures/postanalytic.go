package structures

type PostAnalytic struct {
	Time          int    `json:"time"`
	Event         string `json:"event"`
	Path          string `json:"path"`
	Ip            string `json:"ip"`
	Platform      string `json:"platform"`
	RefererDomain string `json:"refererDomain"`
	CountryCode   string `json:"countryCode"`
}

type PostAnalytics struct {
	List []PostAnalytic `json:"list"`
}
