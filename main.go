package main

import (
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/azer/logger"
	"github.com/codegangsta/cli"
	"github.com/facebookgo/grace/gracehttp"

	"github.com/GitbookIO/micro-analytics/database"
	"github.com/GitbookIO/micro-analytics/utils"
	"github.com/GitbookIO/micro-analytics/utils/geoip"
)

func main() {
	// App meta-data
	app := cli.NewApp()
	app.Version = "0.0.1"
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
			EnvVar: "ROOT",
		},
		cli.IntFlag{
			Name:   "connections, c",
			Value:  10,
			Usage:  "Max number of alive DB connections",
			EnvVar: "POOL_SIZE",
		},
		cli.IntFlag{
			Name:   "cache-size, s",
			Value:  100000,
			Usage:  "Max number of cached requests",
			EnvVar: "CACHE_SIZE",
		},
	}

	var log = logger.New("[Main]")

	// Main app code
	app.Action = func(ctx *cli.Context) {
		// Extract options from CLI args
		driverOpts := database.DriverOpts{
			Directory:      path.Clean(ctx.String("directory")),
			MaxDBs:         ctx.Int("connections"),
			CacheSize:      ctx.Int("cache-size"),
			ClosingChannel: make(chan bool, 1),
		}

		// Create Analytics directory if inexistant
		dirExists, err := utils.PathExists(driverOpts.Directory)
		if err != nil {
			log.Error("Analytics directory path error [%v]", err)
			os.Exit(1)
		}
		if !dirExists {
			log.Info("Analytics directory doesn't exist: %s", driverOpts.Directory)
			log.Info("Creating Analytics directory...")
			os.MkdirAll(driverOpts.Directory, os.ModePerm)
		} else {
			log.Info("Working with existing Analytics directory: %s", driverOpts.Directory)
		}

		// Initiate Geolite2 DB Reader
		geolite2, err := geoip.GetGeoLite2Reader()
		if err != nil {
			log.Info("Error [%v] obtaining a geolite2Reader", err)
			log.Info("Running without Geolite2")
		}

		// Handle exit by softly closing DB connections
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		go func() {
			<-c
			log.Info("Closing database connections...")
			driverOpts.ClosingChannel <- true
			<-driverOpts.ClosingChannel
			log.Info("Connections closed successfully")
			log.Info("Closing Geolite2 connection...")
			geolite2.Close()
			log.Info("Geolite2 is now closed")
			log.Info("Goodbye!")
			os.Exit(1)
		}()

		// Setup server
		opts := ServerOpts{
			Port:           normalizePort(ctx.String("port")),
			Version:        app.Version,
			DriverOpts:     driverOpts,
			Geolite2Reader: geolite2,
		}

		log.Info("Launching server with: %#v", opts)

		server, err := NewServer(opts)
		if err != nil {
			log.Error("ServerSetup error [%v]", err)
			os.Exit(1)
		}

		// Run server
		if err := gracehttp.Serve(server); err != nil {
			log.Error("ListenAndServe error [%v]", err)
			os.Exit(1)
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
