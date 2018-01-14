package main

import (
	"flag"

	"github.com/kataras/iris"
	"github.com/michelvocks/gaia"
	"github.com/michelvocks/gaia/handlers"
)

var (
	cfg          *gaia.Config
	irisInstance *iris.Application
)

func init() {
	cfg = &gaia.Config{}

	// command line arguments
	flag.StringVar(&cfg.ListenPort, "port", "8080", "Listen port for gaia")
	flag.StringVar(&cfg.Bolt.Path, "dbpath", "gaia.db", "Path to gaia bolt db file")

	// Default values
	cfg.Bolt.Mode = 0600
}

func main() {
	// Parse command line flgs
	flag.Parse()

	// Initialize IRIS
	irisInstance = iris.New()

	// Initialize handlers
	handlers.InitHandlers(irisInstance)

	// Start listen
	irisInstance.Run(iris.Addr(":" + cfg.ListenPort))
}
