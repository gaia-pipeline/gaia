package agent

import (
	"fmt"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/stringhelper"
	"os/exec"
	"strings"
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
// or build of pipelines in the local "path" variable. If the tags list contains
// a negative language value, the language check is skipped.
func findLocalBinaries(tags []string) []string {
	var foundSuppBinary []string

	// Iterate all supported binary names
	for key, binName := range supportedBinaries {
		// Check if negative tags value has been set
		if stringhelper.IsContainedInSlice(tags, fmt.Sprintf("-%s", key.String()), true) {
			continue
		}

		// Check if the binary name is available
		if _, err := exec.LookPath(binName); err == nil {
			// It is available. Add the tag to the list.
			foundSuppBinary = append(foundSuppBinary, key.String())
		}
	}

	// Add given tags
	for _, tag := range tags {
		if !strings.HasPrefix(tag, "-") {
			foundSuppBinary = append(foundSuppBinary, tag)
		}
	}

	return foundSuppBinary
}
