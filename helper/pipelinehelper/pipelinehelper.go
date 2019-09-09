package pipelinehelper

import (
	"fmt"
	"strings"

	"github.com/gaia-pipeline/gaia"
)

const (
	typeDelimiter = "_"
)

// GetRealPipelineName removes the suffix from the pipeline name.
func GetRealPipelineName(name string, pType gaia.PipelineType) string {
	return strings.TrimSuffix(name, typeDelimiter+pType.String())
}

// appendTypeToName appends the type to the output binary name.
// This allows later to define the pipeline type by the pipeline binary name.
func AppendTypeToName(n string, pType gaia.PipelineType) string {
	return fmt.Sprintf("%s%s%s", n, typeDelimiter, pType.String())
}
