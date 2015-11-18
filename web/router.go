package web

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "path"

    "github.com/gorilla/mux"

    "github.com/GitbookIO/analytics/database"
    "github.com/GitbookIO/analytics/utils"
    "github.com/GitbookIO/analytics/web/errors"
)

func NewRouter(mainDir string) http.Handler {
    r := mux.NewRouter()

    ////
    // Service methods
    ////

    // Welcome
    r.Path("/").
        Methods("GET").
        HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

        msg := map[string]string{
            "message":  "Welcome to analytics !",
        }
        render(w, msg, nil)
    })

    // Query a DB
    r.Path("/{dbName}").
        Methods("GET").
        HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

        // Parse form data
        if err := req.ParseForm(); err != nil {
            renderError(w, err)
            return
        }

        // Get dbName from URL
        vars := mux.Vars(req)
        dbName := vars["dbName"]

        // Check if DB exists
        dbPath := path.Join(mainDir, dbName, "analytics.db")
        dbExists, err := utils.PathExists(dbPath)
        // Error reading file -> Log
        if err != nil {
            log.Printf("Error [%#v] trying to reach file '%s'\n", err, dbPath)
            renderError(w, &errors.InvalidDatabaseName)
            return
        }

        // DB doesn't exist
        if !dbExists {
            renderError(w, &errors.InvalidDatabaseName)
            return
        }

        // Open DB
        db, err := database.OpenOrCreate(dbName)
        if err != nil {
            log.Fatal("Error opening or creating DB", err)
        }
        defer db.Close()

        // Return query result
        analytics := database.Query(db)
        render(w, analytics, nil)
    })

    // Push analytics to a DB
    r.Path("/{dbName}").
        Methods("POST").
        HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

        // Get dbName from URL
        vars := mux.Vars(req)
        dbName := vars["dbName"]

        // Open DB
        db, err := database.OpenOrCreate(dbName)
        if err != nil {
            log.Fatal("Error opening or creating DB", err)
        }
        defer db.Close()

        // Parse JSON POST data
        analytic := database.Analytic{}
        jsonDecoder := json.NewDecoder(req.Body)
        err = jsonDecoder.Decode(&analytic)

        // Invalid JSON
        if err != nil {
            renderError(w, &errors.InvalidJSON)
            return
        }

        // Insert data if everything's OK
        database.Insert(db, analytic)
        log.Printf("Successfully inserted analytic: %#v", analytic)
    })

    // Delete a DB
    r.Path("/{dbName}").
        Methods("DELETE").
        HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

        // Get dbName from URL
        vars := mux.Vars(req)
        dbName := vars["dbName"]

        // Delete full DB directory
        dbDir := path.Join(mainDir, dbName)
        os.RemoveAll(dbDir)
    })

    return r
}