package pipelinehelper

import (
	"testing"

	"github.com/gaia-pipeline/gaia"
)

func TestGetRealPipelineName(t *testing.T) {
	pipeGo := "my_pipeline_golang"
	if GetRealPipelineName(pipeGo, gaia.PTypeGolang) != "my_pipeline" {
		t.Fatalf("output should be my_pipeline but is %s", GetRealPipelineName(pipeGo, gaia.PTypeGolang))
	}
}
