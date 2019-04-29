package agent

import (
	"github.com/gaia-pipeline/gaia"
	"os/exec"
)

var (
	supportedBinaries = map[gaia.PipelineType]string{
		gaia.PTypePython: "python",
		gaia.PTypeJava:   "mvn",
		gaia.PTypeCpp:    "make",
		gaia.PTypeGolang: "go",
		gaia.PTypeRuby:   "gem",
	}
)

// findLocalBinaries finds all supported local binaries for local execution
// or build of pipelines in the local "path" variable.
func findLocalBinaries() []string {
	var foundSuppBinary []string

	// Iterate all supported binary names
	for key, binName := range supportedBinaries {
		// Check if the binary name is available
		if _, err := exec.LookPath(binName); err == nil {
			// It is available. Add the tag to the list.
			foundSuppBinary = append(foundSuppBinary, key.String())
		}
	}
	return foundSuppBinary
}
