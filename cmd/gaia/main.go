package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/handlers"
	"github.com/gaia-pipeline/gaia/pipeline"
	scheduler "github.com/gaia-pipeline/gaia/scheduler"
	"github.com/gaia-pipeline/gaia/store"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/kataras/iris"
)

var (
	irisInstance *iris.Application
)

const (
	dataFolder      = "data"
	pipelinesFolder = "pipelines"
)

func init() {
	gaia.Cfg = &gaia.Config{}

	// command line arguments
	flag.StringVar(&gaia.Cfg.ListenPort, "port", "8080", "Listen port for gaia")
	flag.StringVar(&gaia.Cfg.HomePath, "homepath", "", "Path to the gaia home folder")
	flag.StringVar(&gaia.Cfg.Bolt.Path, "dbpath", "gaia.db", "Path to gaia bolt db file")

	// Default values
	gaia.Cfg.Bolt.Mode = 0600
}

func main() {
	// Parse command line flgs
	flag.Parse()

	// Initialize shared logger
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})

	// Find path for gaia home folder if not given by parameter
	if gaia.Cfg.HomePath == "" {
		// Find executeable path
		execPath, err := findExecuteablePath()
		if err != nil {
			gaia.Cfg.Logger.Error("cannot find executeable path", "error", err.Error())
			os.Exit(1)
		}
		gaia.Cfg.HomePath = execPath
		gaia.Cfg.Logger.Debug("executeable path found", "path", execPath)
	}

	// Set data path and pipeline path relative to home folder and create it
	// if not exist.
	gaia.Cfg.DataPath = gaia.Cfg.HomePath + string(os.PathSeparator) + dataFolder
	gaia.Cfg.PipelinePath = gaia.Cfg.DataPath + string(os.PathSeparator) + pipelinesFolder
	err := os.MkdirAll(gaia.Cfg.PipelinePath, 0700)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create data folder", "error", err.Error(), "path", gaia.Cfg.DataPath)
		os.Exit(1)
	}

	// Initialize IRIS
	irisInstance = iris.New()

	// Initialize store
	store := store.NewStore()
	err = store.Init()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize store", "error", err.Error())
		os.Exit(1)
	}

	// Initialize handlers
	err = handlers.InitHandlers(irisInstance, store)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize handlers", "error", err.Error())
		os.Exit(1)
	}

	// Initialize scheduler
	scheduler := scheduler.NewScheduler(store)
	scheduler.Init()

	// Start ticker. Periodic job to check for new plugins.
	pipeline.InitTicker(store, scheduler)

	// Start listen
	irisInstance.Run(iris.Addr(":" + gaia.Cfg.ListenPort))
}

// findExecuteablePath returns the absolute path for the current
// process.
func findExecuteablePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}
