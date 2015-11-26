package utils

import (
    "log"
    "net"
    "strings"

    "github.com/oschwald/maxminddb-golang"
    "github.com/GitbookIO/geo-utils-go"

    "github.com/GitbookIO/analytics/data"
)

type lookupResult struct {
    Country struct {
        ISOCode string `maxminddb:"iso_code"`
    } `maxminddb:"country"`
}

func GetGeoLite2Reader() (*maxminddb.Reader, error) {
    data, err := geolite2db.Asset("GeoLite2-Country.mmdb")
    if err != nil {
        log.Printf("[GeoIP] Unable to open GeoLite2-Country asset file. Error %v\n", err)
        return nil, err
    }

    db, err := maxminddb.FromBytes(data)
    if err != nil {
        log.Printf("[GeoIP] Unable to open GeoLite2-Country database. Error %v\n", err)
        return nil, err
    }

    return db, nil
}

// Return ISOCode for an IP
func GeoIpLookup(geolite2 *maxminddb.Reader, ipStr string) (string, error) {
    ip := net.ParseIP(ipStr)

    result := lookupResult{}
    err := geolite2.Lookup(ip, &result)
    if err != nil {
        log.Printf("[GeoIP] Unable to lookup for IP %s\n", ipStr)
        return "", err
    }

    return strings.ToLower(result.Country.ISOCode), nil
}

// Return a country fullname from countryCode
func GetCountry(countryCode string) string {
    return geoutils.GetCountry(countryCode)
}