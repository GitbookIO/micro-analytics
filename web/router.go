package web

import (
    "encoding/json"
    "log"
    "net/http"
    "net/url"
    "time"

    "github.com/gorilla/mux"

    "github.com/GitbookIO/analytics/database"
    "github.com/GitbookIO/analytics/utils"
    "github.com/GitbookIO/analytics/web/errors"
)

func NewRouter(mainDir string, maxDBs int) http.Handler {
    // Create the app DB manager
    dbManager := database.NewManager(mainDir, maxDBs)

    // Create the app router
    r := mux.NewRouter()

    /////
    // Welcome
    /////
    r.Path("/").
        Methods("GET").
        HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

        msg := map[string]string{
            "message":  "Welcome to analytics !",
        }
        render(w, msg, nil)
    })

    /////
    // Query a DB by property
    /////
    r.Path("/{dbName}/{property}").
        Methods("GET").
        HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        // Map of allowed requests
        allowedProperties := map[string]bool{
            "countries": true,
            "platforms": true,
            "domains": true,
            "types":true,
        }
        // Get params from URL
        vars := mux.Vars(req)
        dbName := vars["dbName"]
        property := vars["property"]

        // Check that property is allowed to be queried
        if _, ok := allowedProperties[property]; !ok {
            renderError(w, &errors.InvalidProperty)
            return
        }

        // Check if DB file exists
        dbExists, err := dbManager.DBExists(dbName)
        if err != nil {
            renderError(w, &errors.InternalError)
            return
        }

        // DB doesn't exist
        if !dbExists {
            renderError(w, &errors.InvalidDatabaseName)
            return
        }

        // Parse request query
        if err := req.ParseForm(); err != nil {
            renderError(w, err)
            return
        }

        // Return query result
        db, err := dbManager.GetDB(dbName)
        if err != nil {
            renderError(w, &errors.InternalError)
            return
        }
        analytics := db.Query()
        render(w, analytics, nil)
    })

    /////
    // Full query a DB
    /////
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

        // Check if DB file exists
        dbExists, err := dbManager.DBExists(dbName)
        if err != nil {
            renderError(w, &errors.InternalError)
            return
        }

        // DB doesn't exist
        if !dbExists {
            renderError(w, &errors.InvalidDatabaseName)
            return
        }

        // Return query result
        db, err := dbManager.GetDB(dbName)
        if err != nil {
            renderError(w, &errors.InternalError)
            return
        }
        analytics := db.Query()
        render(w, analytics, nil)
    })

    /////
    // Push analytics to a DB
    /////
    r.Path("/{dbName}").
        Methods("POST").
        HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

        // Get dbName from URL
        vars := mux.Vars(req)
        dbName := vars["dbName"]

        // Parse JSON POST data
        postData := PostData{}
        jsonDecoder := json.NewDecoder(req.Body)
        err := jsonDecoder.Decode(&postData)

        // Invalid JSON
        if err != nil {
            renderError(w, &errors.InvalidJSON)
            return
        }

        // Create Analytic to inject in DB
        analytic := database.Analytic{
            Time:   time.Now(),
            Type:   postData.Type,
            Path:   postData.Path,
            Ip:     postData.Ip,
        }

        // Set time from POST data if passed
        if len(postData.Time) > 0 {
            analytic.Time, err = time.Parse(time.RFC3339, postData.Time)
        }

        // Get referer from headers
        refererHeader := postData.Headers["referer"]
        if referrerURL, err := url.ParseRequestURI(refererHeader); err == nil {
            analytic.RefererDomain = referrerURL.Host
        }

        // Get platform from headers
        analytic.Platform = utils.Platform(postData.Headers["user-agent"])

        // Get countryCode from GeoIp
        analytic.CountryCode = utils.GeoIpLookup(postData.Ip)

        // Insert data if everything's OK
        db, err := dbManager.GetDB(dbName)
        if err != nil {
            renderError(w, &errors.InternalError)
            return
        }
        if err = db.Insert(analytic); err != nil {
            renderError(w, &errors.InsertFailed)
            return
        }

        log.Printf("[Router] Successfully inserted analytic: %#v", analytic)
        render(w, nil, nil)
    })

    /////
    // Delete a DB
    /////
    r.Path("/{dbName}").
        Methods("DELETE").
        HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

        // Get dbName from URL
        vars := mux.Vars(req)
        dbName := vars["dbName"]

        // Delete full DB directory
        err := dbManager.DeleteDB(dbName)
        render(w, nil, err)
    })

    return r
}