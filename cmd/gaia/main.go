package main

import (
	"flag"

	"github.com/kataras/iris"
	"github.com/michelvocks/gaia/handlers"
)

var (
	cfg          *Config
	irisInstance *iris.Application
)

// Config holds all config options
type Config struct {
	ListenPort string
}

func init() {
	cfg = &Config{}

	// command line arguments
	flag.StringVar(&cfg.ListenPort, "port", "8080", "Listen port for gaia WebUI")
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
