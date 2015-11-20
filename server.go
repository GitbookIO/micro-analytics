package main

import (
    "net/http"
    "os"

    "github.com/gorilla/handlers"

    "github.com/GitbookIO/analytics/web"
)

type ServerOpts struct {
    Port        string  `json:"port"`
    Directory   string  `json:"directory"`
    MaxDBs      int     `json:"maxDBs"`
    Version     string  `json:"version"`
}

// Build a http.Server based on the options
func NewServer(opts ServerOpts) (*http.Server, error) {
    // Define handler
    handler := web.NewRouter(opts.Directory, opts.MaxDBs)

    // Use logging
    handler = handlers.LoggingHandler(os.Stderr, handler)

    return &http.Server{
        Addr:       opts.Port,
        Handler:    handler,
    }, nil
}