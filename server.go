package main

import (
    "net/http"
    "os"

    "github.com/gorilla/handlers"

    "github.com/GitbookIO/analytics/database"
    "github.com/GitbookIO/analytics/web"
)

type ServerOpts struct {
    Port        string
    Version     string
    DBManager   *database.DBManager
}

// Build a http.Server based on the options
func NewServer(opts ServerOpts) (*http.Server, error) {
    // Define handler
    handler := web.NewRouter(opts.DBManager)

    // Use logging
    handler = handlers.LoggingHandler(os.Stderr, handler)

    return &http.Server{
        Addr:       opts.Port,
        Handler:    handler,
    }, nil
}