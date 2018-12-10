package pipeline

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/server"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	uuid "github.com/satori/go.uuid"
)

func TestBuildPipelineAcceptanceTestTearUp(t *testing.T) {
	if os.Getenv("GAIA_RUN_ACC") != "true" {
		t.Skip("skipping acceptance tests because GAIA_RUN_ACC is not 'true'")
	}

	// Create temp folder for acceptance test.
	tmp, _ := ioutil.TempDir("", "TestBuildPipelineAcceptanceTestTearUp")
	gaia.Cfg.HomePath = tmp
	defer func() {
		os.RemoveAll(tmp)
	}()

	// Start the server as background process.
	go func() {
		err := server.Start()
		if err != nil {
			t.Errorf("cannot start test server: %+v", err)
		}
	}()

	// Sleep a bit until all components has been initialized and started.
	time.Sleep(2 * time.Second)

	// Define acceptance tests here.
	t.Run("BuildGoPluginTest", buildGoPluginTest)
	t.Run("BuildJavaPluginTest", buildJavaPluginTest)
	t.Run("BuildPythonPluginTest", buildPythonPluginTest)
	t.Run("BuildCppPluginTest", buildCppPluginTest)
}

func buildGoPluginTest(t *testing.T) {
	// Create test pipeline.
	testPipeline := &gaia.CreatePipeline{
		ID: uuid.Must(uuid.NewV4(), nil).String(),
		Pipeline: gaia.Pipeline{
			Name: "GoTestPipeline",
			Type: gaia.PTypeGolang,
			Repo: gaia.GitRepo{URL: "https://github.com/gaia-pipeline/go-example"},
		},
	}

	// Build pipeline.
	pipeline.CreatePipeline(testPipeline)

	// Check if everything went smoothly.
	if testPipeline.StatusType == gaia.CreatePipelineFailed {
		t.Errorf("create go pipeline failed: %s", testPipeline.Output)
	}
}

func buildJavaPluginTest(t *testing.T) {
	// Create test pipeline.
	testPipeline := &gaia.CreatePipeline{
		ID: uuid.Must(uuid.NewV4(), nil).String(),
		Pipeline: gaia.Pipeline{
			Name: "JavaTestPipeline",
			Type: gaia.PTypeJava,
			Repo: gaia.GitRepo{URL: "https://github.com/gaia-pipeline/java-example"},
		},
	}

	// Build pipeline.
	pipeline.CreatePipeline(testPipeline)

	// Check if everything went smoothly.
	if testPipeline.StatusType == gaia.CreatePipelineFailed {
		t.Errorf("create java pipeline failed: %s", testPipeline.Output)
	}
}

func buildPythonPluginTest(t *testing.T) {
	// Create test pipeline.
	testPipeline := &gaia.CreatePipeline{
		ID: uuid.Must(uuid.NewV4(), nil).String(),
		Pipeline: gaia.Pipeline{
			Name: "PythonTestPipeline",
			Type: gaia.PTypePython,
			Repo: gaia.GitRepo{URL: "https://github.com/gaia-pipeline/python-example"},
		},
	}

	// Build pipeline.
	pipeline.CreatePipeline(testPipeline)

	// Check if everything went smoothly.
	if testPipeline.StatusType == gaia.CreatePipelineFailed {
		t.Errorf("create python pipeline failed: %s", testPipeline.Output)
	}
}

func buildCppPluginTest(t *testing.T) {
	// Create test pipeline.
	testPipeline := &gaia.CreatePipeline{
		ID: uuid.Must(uuid.NewV4(), nil).String(),
		Pipeline: gaia.Pipeline{
			Name: "CppTestPipeline",
			Type: gaia.PTypeCpp,
			Repo: gaia.GitRepo{URL: "https://github.com/gaia-pipeline/cpp-example"},
		},
	}

	// Build pipeline.
	pipeline.CreatePipeline(testPipeline)

	// Check if everything went smoothly.
	if testPipeline.StatusType == gaia.CreatePipelineFailed {
		t.Errorf("create cpp pipeline failed: %s", testPipeline.Output)
	}
}
