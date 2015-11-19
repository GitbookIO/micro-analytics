package web

import (
    "encoding/json"
    "log"
    "net/http"
    "net/url"
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
            renderError(w, &errors.InternalError)
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

        // // Open DB
        db, err := database.OpenOrCreate(dbName)
        if err != nil {
            log.Fatal("Error opening or creating DB", err)
        }
        defer db.Close()

        // Parse JSON POST data
        postData := PostData{}
        jsonDecoder := json.NewDecoder(req.Body)
        err = jsonDecoder.Decode(&postData)

        // Invalid JSON
        if err != nil {
            renderError(w, &errors.InvalidJSON)
            return
        }

        // Create Analytic to inject in DB
        analytic := database.Analytic{}

        // Set time if not in POST data
        if postData.Time != nil {
            analytic.Time = postData.Time
        } else {
            analytic.Time = time.Now()
        }
        analytic.Type = postData.Type
        analytic.Path = postData.Path
        analytic.Ip = postData.Ip

        // Get referer from headers
        referrerURL, err := url.ParseRequestURI(postData.Headers["referer"])
        if err == nil {
            analytic.RefererDomain = referrerURL.Host
        }

        // Get platform from headers
        analytic.Platform = utils.Platform(postData.Headers["user-agent"])

        // Get countryCode from GeoIp
        analytic.CountryCode = utils.GeoIpLookup(postData.Ip)

        log.Printf("%#v\n", analytic)

        // Insert data if everything's OK
        err = database.Insert(db, analytic)
        if err != nil {
            renderError(w, &errors.InsertFailed)
            return
        }

        // log.Printf("Successfully inserted analytic: %#v", analytic)
        render(w, nil, nil)
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