package main

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/oschwald/maxminddb-golang"

	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/web"
)

type ServerOpts struct {
	Port           string
	Version        string
	DriverOpts     database.DriverOpts
	Geolite2Reader *maxminddb.Reader
	Auth           *web.BasicAuth
}

// Build a http.Server based on the options
func NewServer(opts ServerOpts) (*http.Server, error) {
	// Create base router
	r := mux.NewRouter()

	// Hello world
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Welcome to micro-analytics!", "version":"` + opts.Version + `"}`))
	})

	// Define private routes handler
	routerOpts := web.RouterOpts{
		DriverOpts:     opts.DriverOpts,
		Geolite2Reader: opts.Geolite2Reader,
		Version:        opts.Version,
	}

	handler, err := web.NewRouter(routerOpts)
	if err != nil {
		return nil, err
	}

	// Use authentication if username provided
	if len(opts.Auth.Name) > 0 {
		handler = web.BasicAuthMiddleware(opts.Auth, handler)
	}

	// Attach to main router
	r.PathPrefix("/").Handler(handler)

	// Use logging
	handler = handlers.LoggingHandler(os.Stderr, r)


	return &http.Server{
		Addr:    opts.Port,
		Handler: handler,
	}, nil
}
