package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/handlers"
	"github.com/gaia-pipeline/gaia/pipeline"
	scheduler "github.com/gaia-pipeline/gaia/scheduler"
	"github.com/gaia-pipeline/gaia/store"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
)

var (
	echoInstance *echo.Echo
)

const (
	// Version is the current version of gaia.
	Version = "0.1.1"

	dataFolder      = "data"
	pipelinesFolder = "pipelines"
	workspaceFolder = "workspace"
)

func init() {
	gaia.Cfg = &gaia.Config{}

	// command line arguments
	flag.StringVar(&gaia.Cfg.ListenPort, "port", "8080", "Listen port for gaia")
	flag.StringVar(&gaia.Cfg.HomePath, "homepath", "", "Path to the gaia home folder")
	flag.StringVar(&gaia.Cfg.Worker, "worker", "2", "Number of worker gaia will use to execute pipelines in parallel")
	flag.BoolVar(&gaia.Cfg.DevMode, "dev", false, "If true, gaia will be started in development mode. Don't use this in production!")
	flag.BoolVar(&gaia.Cfg.VersionSwitch, "version", false, "If true, will print the version and immediately exit")

	// Default values
	gaia.Cfg.Bolt.Mode = 0600
}

func main() {
	// Parse command line flgs
	flag.Parse()

	// Check version switch
	if gaia.Cfg.VersionSwitch {
		fmt.Printf("Gaia Version: V%s\n", Version)
		os.Exit(0)
	}

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

	// Set data path, workspace path and pipeline path relative to home folder and create it
	// if not exist.
	gaia.Cfg.DataPath = filepath.Join(gaia.Cfg.HomePath, dataFolder)
	err := os.MkdirAll(gaia.Cfg.DataPath, 0700)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create folder", "error", err.Error(), "path", gaia.Cfg.DataPath)
		os.Exit(1)
	}
	gaia.Cfg.PipelinePath = filepath.Join(gaia.Cfg.HomePath, pipelinesFolder)
	err = os.MkdirAll(gaia.Cfg.PipelinePath, 0700)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create folder", "error", err.Error(), "path", gaia.Cfg.PipelinePath)
		os.Exit(1)
	}
	gaia.Cfg.WorkspacePath = filepath.Join(gaia.Cfg.HomePath, workspaceFolder)
	err = os.MkdirAll(gaia.Cfg.WorkspacePath, 0700)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create data folder", "error", err.Error(), "path", gaia.Cfg.WorkspacePath)
		os.Exit(1)
	}

	// Initialize echo instance
	echoInstance = echo.New()

	// Initialize store
	store := store.NewStore()
	err = store.Init()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize store", "error", err.Error())
		os.Exit(1)
	}

	// Initialize scheduler
	scheduler := scheduler.NewScheduler(store)
	err = scheduler.Init()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize scheduler:", "error", err.Error())
		os.Exit(1)
	}

	// Initialize handlers
	err = handlers.InitHandlers(echoInstance, store, scheduler)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize handlers", "error", err.Error())
		os.Exit(1)
	}

	// Start ticker. Periodic job to check for new plugins.
	pipeline.InitTicker(store, scheduler)

	// Start listen
	echoInstance.Logger.Fatal(echoInstance.Start(":" + gaia.Cfg.ListenPort))
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
