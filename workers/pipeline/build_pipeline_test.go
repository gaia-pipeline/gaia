package pipeline

import (
	"os"
	"testing"

	"github.com/gaia-pipeline/gaia/server"
)

func TestAcceptanceTestTearUp(t *testing.T) {
	if os.Getenv("GAIA_RUN_ACC") != "true" {
		t.Skip("skipping acceptance tests because GAIA_RUN_ACC is not 'true'")
	}

	// Start the server.
	err := server.Start()

	// Define acceptance tests here.
	t.Run("BuildGoPluginTest", buildGoPluginTest)
}

func buildGoPluginTest(t *testing.T) {

}
