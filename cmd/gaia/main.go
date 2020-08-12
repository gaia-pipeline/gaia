package main

import (
	"os"

	_ "github.com/gaia-pipeline/gaia/docs"
	"github.com/gaia-pipeline/gaia/server"
)

// @title Gaia API
// @version 1.0
// @description This is the API that the Gaia Admin UI uses.
// @termsOfService https://github.com/gaia-pipeline/gaia/blob/master/LICENSE

// @contact.name API Support
// @contact.url https://github.com/gaia-pipeline/gaia

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Authorization

// @BasePath /api/v1
func main() {
	// Start the server.
	if err := server.Start(); err != nil {
		os.Exit(1)
	}
}
