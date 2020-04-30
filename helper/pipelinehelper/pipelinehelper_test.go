package pipelinehelper

import (
	"testing"

	"github.com/gaia-pipeline/gaia/helper/stringhelper"

	"github.com/gaia-pipeline/gaia"
)

func TestGetRealPipelineName(t *testing.T) {
	pipeGo := "my_pipeline_golang"
	if GetRealPipelineName(pipeGo, gaia.PTypeGolang) != "my_pipeline" {
		t.Fatalf("output should be my_pipeline but is %s", GetRealPipelineName(pipeGo, gaia.PTypeGolang))
	}
}

func TestAppendTypeToName(t *testing.T) {
	expected := []string{"my_pipeline_golang", "my_pipeline2_java", "my_pipeline_python"}
	input := []struct {
		name  string
		pType gaia.PipelineType
	}{
		{
			"my_pipeline",
			gaia.PTypeGolang,
		},
		{
			"my_pipeline2",
			gaia.PTypeJava,
		},
		{
			"my_pipeline",
			gaia.PTypePython,
		},
	}

	for _, inp := range input {
		got := AppendTypeToName(inp.name, inp.pType)
		if !stringhelper.IsContainedInSlice(expected, got, false) {
			t.Fatalf("expected name not contained: %s", got)
		}
	}
}
