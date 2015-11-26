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

// Return ISOCode for an IP
func GeoIpLookup(ipStr string) (string, error) {
    data, err := geolite2db.Asset("GeoLite2-Country.mmdb")
    if err != nil {
        log.Printf("[GeoIP] Unable to open GeoLite2-Country asset file. Error %v\n", err)
        return "", err
    }

    db, err := maxminddb.FromBytes(data)
    if err != nil {
        log.Printf("[GeoIP] Unable to open GeoLite2-Country database. Error %v\n", err)
        return "", err
    }
    defer db.Close()

    ip := net.ParseIP(ipStr)

    result := lookupResult{}
    err = db.Lookup(ip, &result)
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