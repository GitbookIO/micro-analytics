# Shardalytics

A micro multi-website analytics database service designed to be fast and robust, built with Go and SQLite.


## Principle

Analytics databases tend to grow fast and exponentially.
Requesting data for one specific website from a single database thus become very slow over time.
But analytics data are highly decoupled between two websites.

The idea behind **Shardalytics** is to shard your analytics data on a key, which is usually a website name.
Each shard thus only contains a specific website data, allowing faster response times and easy horizontal scaling.

To handle requests even faster, **Shardalytics** automatically manages a pool of connections to multiple shards at a time.
By default, the service keeps 10 connections alive.
But you can easily increase/decrease the max number of alive shards with the `--conections` flag when launching the app.


## Analytics schema

All shards of the **Shardalytics** database share the same TABLE schema :
```SQL
CREATE TABLE visits (
    time            INTEGER,
    event           TEXT,
    path            TEXT,
    ip              TEXT,
    platform        TEXT,
    refererDomain   TEXT,
    countryCode     TEXT
)
```


## Service requests

### GET requests

##### Common Parameters

Every query for a specific website can be executed using a time range.
Every following GET request thus takes the two following optional query string parameters :

Name | Type | Description | Default | Example
---- | ---- | ---- | ---- | ----
`start` | Date | Start date to query a range | none | `"2015-11-20T12:00:00.000Z"`
`end` | Date | End date to query a range | none | `"2015-11-21T12:00:00.000Z"`

The dates can be passed either as :
 - ISO (RFC3339) `"2015-11-20T12:00:00.000Z"`
 - UTC (RFC1123) `"Fri, 20 Nov 2015 12:00:00 GMT"`


##### Common Response Values

Every response to a GET request will contain the two following values :

Name | Type | Description
---- | ---- | ----
`total` | Integer | Total number of visits
`unique` | Integer | Total number of unique visitors based on `ip`


#### GET `/:website/countries`

Returns the number of visits per `countryCode`.

##### Response

`label` contains the country full name.

```JavaScript
{
    "list": [
        {
            "id": "fr",
            "label": "France",
            "total": 1000,
            "unique": 900
        },
        ...
    ]
}
```

#### GET `/:website/platforms`

Returns the number of visits per `platform`.

##### Response

```JavaScript
{
    "list": [
        {
            "id": "Linux",
            "label": "Linux",
            "total": 1000,
            "unique": 900
        },
        ...
    ]
}
```

#### GET `/:website/domains`

Returns the number of visits per `refererDomain`.

##### Response

```JavaScript
{
    "list": [
        {
            "id": "gitbook.com",
            "label": "gitbook.com",
            "total": 1000,
            "unique": 900
        },
        ...
    ]
}
```

#### GET `/:website/types`

Returns the number of visits per `type`.

##### Response

```JavaScript
{
    "list": [
        {
            "id": "download",
            "label": "download",
            "total": 1000,
            "unique": 900
        },
        ...
    ]
}
```

#### GET `/:website/time`

Returns the number of visits as a time serie. The interval in seconds can be specified as an optional query string parameter. Its default value is `86400`, equivalent to one day.

##### Parameters

Name | Type | Description | Default | Example
---- | ---- | ---- | ---- | ----
`interval` | Integer | Interval of the time serie | `86400` (1 day) | `3600`

##### Response

Example with interval set to `3600` :

```JavaScript
{
    "list": [
        {
            "start": "2015-11-24T12:00:00.000Z",
            "end": "2015-11-24T13:00:00.000Z",
            "total": 450,
            "unique": 390
        },
        {
            "start": "2015-11-24T13:00:00.000Z",
            "end": "2015-11-24T14:00:00.000Z",
            "total": 550,
            "unique": 510
        },
        ...
    ]
}
```

### POST requests

#### POST `/:website`

Insert new data for the specified website.

##### POST Body

```JavaScript
{
    "time": "2015-11-24T13:00:00.000Z", // optional
    "type": "download",
    "ip": "127.0.0.1",
    "path": "/README.md",
    "headers": {
        ...
        // HTTP headers received from your visitor
    }
}
```

The `time` parameter is optional and is set to the date of your POST request by default.

Passing the HTTP headers in the POST body allows the service to extract the `refererDomain` and `platform` values.
The `countryCode` will be deduced from the passed `ip` parameter using [Maxmind's GeoLite2 database](http://dev.maxmind.com/geoip/geoip2/geolite2/).


## Application's parameters

Running the application is as simple as running :
```
$ ./analytics
```

You can also provide the following parameters :

Parameter | Usage | Type | Default Value
---- | ---- | ---- | ----
`-port, --p` | Port to listen on | String | `"7070"`
`-directory, -d` | Database directory | String | `"./dbs"`
`-connections, -c` | Max number of alive shards connections | Number | `10`


## Using GeoIp

To be able to use [Maxmind's GeoLite2 DB](http://dev.maxmind.com/geoip/geoip2/geolite2/), your application should embed the `GeoLite2-Country.mmdb` file in the `data/` folder.

Your application's root folder should look like this :
```
bin/
    .
    ..
    analytics
    data/
        GeoLite2-Country.mmdb
```