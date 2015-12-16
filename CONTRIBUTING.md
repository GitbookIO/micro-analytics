# Contributing to µAnalytics

## Building the binaries

The `script/` folder contains the logic to build **µAnalytics** binaries for both darwin_amd64 and linux_amd64 architectures on a Mac.

While the darwin_amd64 binary is built locally, you need to have [Docker](https://www.docker.com/) installed on your machine to build the linux_amd64 binary.

Once you got an instance of Docker running locally, go to the `script/` folder and run:

```Shell
$ ./all.sh
```

After the script finishes, the binaries are placed in the `script/build/` folder.

## GeoLite2 data file

The [Maxmind's GeoLite2 DB](http://dev.maxmind.com/geoip/geoip2/geolite2/) is pre-compiled in the source files using [go-bindata](https://github.com/jteeuwen/go-bindata).
The go file can be found in `/utils/geoip/data/geolite2db.go`.

To refresh `geolite2db.go` from a new `GeoLite2-Country.mmdb` file, go to the `/utils/geoip/data` folder and run:

```Shell
$ go generate
```

