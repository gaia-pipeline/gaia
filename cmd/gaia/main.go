package main

import (
	"os"

	"github.com/gaia-pipeline/gaia/server"
	_ "github.com/gaia-pipeline/gaia/docs"
)

// @title Gaia API
// @version 1.0
// @description This is the API that the Gaia Admin UI uses.
// @termsOfService https://github.com/gaia-pipeline/gaia/blob/master/LICENSE

// @contact.name API Support
// @contact.url https://github.com/gaia-pipeline/gaia

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host gaia-pipeline.io
// @BasePath /api/v1
func main() {
	// Start the server.
	if err := server.Start(); err != nil {
		os.Exit(1)
	}
}
