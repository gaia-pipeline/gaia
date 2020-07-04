package pipelinehelper

import (
	"errors"
	"fmt"
	"path/filepath"
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

// AppendTypeToName appends the type to the output binary name.
// This allows later to define the pipeline type by the pipeline binary name.
func AppendTypeToName(n string, pType gaia.PipelineType) string {
	return fmt.Sprintf("%s%s%s", n, typeDelimiter, pType.String())
}

// GetLocalDestinationForPipeline computes the local location of a pipeline on disk based on the pipeline's
// type and configuration of Gaia such as, temp folder and data folder.
func GetLocalDestinationForPipeline(p gaia.Pipeline) (string, error) {
	tmpFolder, err := tmpFolderFromPipelineType(p)
	if err != nil {
		gaia.Cfg.Logger.Error("Pipeline type invalid", "type", p.Type)
		return "", err
	}
	return filepath.Join(gaia.Cfg.HomePath, gaia.TmpFolder, tmpFolder, gaia.SrcFolder, p.UUID), nil
}

// tmpFolderFromPipelineType returns the gaia specific tmp folder for a pipeline
// based on the type of the pipeline.
func tmpFolderFromPipelineType(foundPipeline gaia.Pipeline) (string, error) {
	switch foundPipeline.Type {
	case gaia.PTypeCpp:
		return gaia.TmpCppFolder, nil
	case gaia.PTypeGolang:
		return gaia.TmpGoFolder, nil
	case gaia.PTypeNodeJS:
		return gaia.TmpNodeJSFolder, nil
	case gaia.PTypePython:
		return gaia.TmpPythonFolder, nil
	case gaia.PTypeRuby:
		return gaia.TmpRubyFolder, nil
	case gaia.PTypeJava:
		return gaia.TmpJavaFolder, nil
	}
	return "", errors.New("invalid pipeline type")
}
