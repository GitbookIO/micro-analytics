package main

import (
    "log"
    "os"
    "os/signal"
    "path"
    "strings"
    "syscall"

    "github.com/codegangsta/cli"
    "github.com/facebookgo/grace/gracehttp"

    "github.com/GitbookIO/analytics/database"
    "github.com/GitbookIO/analytics/utils"
)

func main() {
    // App meta-data
    app := cli.NewApp()
    app.Version = "0.9.0"
    app.Name = "µAnalytics"
    app.Author = "Johan Preynat"
    app.Email = "johan.preynat@gmail.com"
    app.Usage = "Fast sharded analytics database"
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:   "port, p",
            Value:  "7070",
            Usage:  "Port to listen on",
            EnvVar: "PORT",
        },
        cli.StringFlag{
            Name:   "directory, d",
            Value:  "./dbs/",
            Usage:  "Database directory",
        },
        cli.IntFlag{
            Name:   "connections, c",
            Value:  10,
            Usage:  "Max number of alive DB connections",
        },
    }

    // Main app code
    app.Action = func(ctx *cli.Context) {
        // Extract options from CLI args
        mainDirectory := path.Clean(ctx.String("directory"))
        maxDBs := ctx.Int("connections")

        // Create Analytics directory if inexistant
        dirExists, err := utils.PathExists(mainDirectory)
        if err != nil {
            log.Fatal("[Main] Analytics directory path error:", err)
        }
        if !dirExists {
            log.Printf("[Main] Analytics directory doesn't exist: %s\n", mainDirectory)
            log.Println("[Main] Creating Analytics directory...")
            os.Mkdir(mainDirectory, os.ModePerm)
        } else {
            log.Printf("[Main] Working with existing Analytics directory: %s\n", mainDirectory)
        }

        // Initiate DBManager
        dbManager := database.NewManager(mainDirectory, maxDBs)

        // Handle exit by softly closing DB connections
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt)
        signal.Notify(c, syscall.SIGTERM)
        go func() {
            <-c
            log.Println("[Main] Purging DB manager...")
            dbManager.Purge()
            log.Println("[Main] Finished purging DB manager")
            log.Println("[Main] Goodbye!")
            os.Exit(1)
        }()

        // Setup server
        opts := ServerOpts{
            Port:       normalizePort(ctx.String("port")),
            Version:    app.Version,
            DBManager:  dbManager,
        }

        log.Printf("[Main] Launching server with: %#v\n", opts)

        server, err := NewServer(opts)
        if err != nil {
            log.Fatal("ServerSetup:", err)
        }

        // Run server
        if err := gracehttp.Serve(server); err != nil {
            log.Fatal("ListenAndServe:", err)
        }
    }

    // Parse CLI args and run
    app.Run(os.Args)
}

// Normalize port string to an "addr"
// as expected by ListenAndServe
func normalizePort(port string) string {
    if strings.Contains(port, ":") {
        return port
    }
    return ":" + port
}