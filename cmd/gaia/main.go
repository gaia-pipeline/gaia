package main

import (
	"os"

	"github.com/gaia-pipeline/gaia/server"
)

func main() {
	// Start the server.
	if err := server.Start(); err != nil {
		os.Exit(1)
	}
}
