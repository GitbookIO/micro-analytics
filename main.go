package main

import (
    "log"
    "os"
    "path"
    "strings"

    "github.com/codegangsta/cli"
    "github.com/facebookgo/grace/gracehttp"

    "github.com/GitbookIO/analytics/utils"
)

func main() {
    // App meta-data
    app := cli.NewApp()
    app.Version = "0.0.1"
    app.Name = "analytics"
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
        opts := ServerOpts{
            Port:       normalizePort(ctx.String("port")),
            Directory:  path.Clean(ctx.String("directory")),
            MaxDBs:     ctx.Int("connections"),
            Version:    app.Version,
        }

        log.Printf("Launching server with: %#v\n\n", opts)

        // Create Analytics directory if inexistant
        dirExists, err := utils.PathExists(opts.Directory)
        if err != nil {
            log.Fatal("Analytics directory path error:", err)
        }
        if !dirExists {
            log.Printf("Analytics directory doesn't exist: %s", opts.Directory)
            log.Printf("Creating Analytics directory...")
            os.Mkdir(opts.Directory, os.ModePerm)
        } else {
            log.Printf("Working with existing Analytics directory: %s", opts.Directory)
        }

        // Setup server
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