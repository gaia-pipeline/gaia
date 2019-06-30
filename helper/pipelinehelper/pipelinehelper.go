package pipelinehelper

import (
	"github.com/gaia-pipeline/gaia"
	"strings"
)

const (
	typeDelimiter = "_"
)

// GetRealPipelineName removes the suffix from the pipeline name.
func GetRealPipelineName(name string, pType gaia.PipelineType) string {
	return strings.TrimSuffix(name, typeDelimiter+pType.String())
}