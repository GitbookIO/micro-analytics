package utils

import (
    "log"
    "net"

    "github.com/oschwald/maxminddb-golang"
)

type lookupResult struct {
    Country struct {
        ISOCode string `maxminddb:"iso_code"`
    } `maxminddb:"country"`
}

// Return ISOCode for an IP
func GeoIpLookup(ipStr string) string {
    db, err := maxminddb.Open("data/GeoLite2-Country.mmdb")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    ip := net.ParseIP(ipStr)

    result := lookupResult{}
    err = db.Lookup(ip, &result)
    if err != nil {
        log.Fatal(err)
    }

    return result.Country.ISOCode
}