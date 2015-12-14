package main

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/oschwald/maxminddb-golang"

	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/web"
)

type ServerOpts struct {
	Port           string
	Version        string
	DriverOpts     database.DriverOpts
	Geolite2Reader *maxminddb.Reader
}

// Build a http.Server based on the options
func NewServer(opts ServerOpts) (*http.Server, error) {
	// Define handler
	routerOpts := web.RouterOpts{
		DriverOpts:     opts.DriverOpts,
		Geolite2Reader: opts.Geolite2Reader,
		Version:        opts.Version,
	}

	handler, err := web.NewRouter(routerOpts)
	if err != nil {
		return nil, err
	}

	// Use logging
	handler = handlers.LoggingHandler(os.Stderr, handler)

	return &http.Server{
		Addr:    opts.Port,
		Handler: handler,
	}, nil
}
