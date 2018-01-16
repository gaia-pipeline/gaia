package main

import (
	"flag"
	"os"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/kataras/iris"
	"github.com/michelvocks/gaia"
	"github.com/michelvocks/gaia/handlers"
	"github.com/michelvocks/gaia/store"
)

var (
	cfg          *gaia.Config
	irisInstance *iris.Application
)

func init() {
	cfg = &gaia.Config{}

	// command line arguments
	flag.StringVar(&cfg.ListenPort, "port", "8080", "Listen port for gaia")
	flag.StringVar(&cfg.DataPath, "datapath", "data", "Path to the data folder")
	flag.StringVar(&cfg.Bolt.Path, "dbpath", "gaia.db", "Path to gaia bolt db file")

	// Default values
	cfg.Bolt.Mode = 0600
}

func main() {
	// Parse command line flgs
	flag.Parse()

	// Initialize shared logger
	cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})

	// Initialize IRIS
	irisInstance = iris.New()

	// Initialize store
	s := store.NewStore()
	err := s.Init(cfg)
	if err != nil {
		cfg.Logger.Error("cannot initialize store", "error", err.Error())
		os.Exit(1)
	}

	// Initialize handlers
	err = handlers.InitHandlers(cfg, irisInstance, s)
	if err != nil {
		cfg.Logger.Error("cannot initialize handlers", "error", err.Error())
		os.Exit(1)
	}

	// Start listen
	irisInstance.Run(iris.Addr(":" + cfg.ListenPort))
}
