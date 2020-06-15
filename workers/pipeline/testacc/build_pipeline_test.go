package pipeline

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/plugin"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/server"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	"github.com/gaia-pipeline/gaia/workers/scheduler/gaiascheduler"
)

func TestBuildPipelineAcceptanceTestTearUp(t *testing.T) {
	//if os.Getenv("GAIA_RUN_ACC") != "true" {
	//	t.Skip("skipping acceptance tests because GAIA_RUN_ACC is not 'true'")
	//}

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

	// Sleep a bit until all components are initialized and started.
	time.Sleep(2 * time.Second)

	// Define acceptance tests here.
	t.Run("BuildGoPluginTest", buildGoPluginTest)
	t.Run("BuildJavaPluginTest", buildJavaPluginTest)
	t.Run("BuildPythonPluginTest", buildPythonPluginTest)
	t.Run("BuildCppPluginTest", buildCppPluginTest)
	t.Run("BuildRubyPluginTest", buildRubyPluginTest)
	t.Run("BuildNodeJSPluginTest", buildNodeJSPluginTest)
}

func buildGoPluginTest(t *testing.T) {
	// Create test pipeline.
	v4, _ := uuid.NewV4()
	testPipeline := &gaia.CreatePipeline{
		ID: uuid.Must(v4, nil).String(),
		Pipeline: gaia.Pipeline{
			Name: "GoTestPipeline",
			Type: gaia.PTypeGolang,
			Repo: &gaia.GitRepo{URL: "https://github.com/gaia-pipeline/go-example"},
		},
	}

	store, _ := services.StorageService()
	db, _ := services.DefaultMemDBService()
	va, _ := services.DefaultVaultService()
	ca, _ := security.InitCA()
	scheduler, _ := gaiascheduler.NewScheduler(gaiascheduler.Dependencies{
		Store: store,
		DB:    db,
		PS:    &plugin.GoPlugin{},
		CA:    ca,
		Vault: va,
	})

	ps := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: scheduler,
	})
	// Build pipeline.
	ps.CreatePipeline(testPipeline)

	// Check if everything went smoothly.
	if testPipeline.StatusType != gaia.CreatePipelineSuccess {
		t.Errorf("create go pipeline failed: %s", testPipeline.Output)
	}
}

func buildJavaPluginTest(t *testing.T) {
	// Create test pipeline.
	v4, _ := uuid.NewV4()
	testPipeline := &gaia.CreatePipeline{
		ID: uuid.Must(v4, nil).String(),
		Pipeline: gaia.Pipeline{
			Name: "JavaTestPipeline",
			Type: gaia.PTypeJava,
			Repo: &gaia.GitRepo{URL: "https://github.com/gaia-pipeline/java-example"},
		},
	}

	store, _ := services.StorageService()
	db, _ := services.DefaultMemDBService()
	va, _ := services.DefaultVaultService()
	ca, _ := security.InitCA()
	scheduler, _ := gaiascheduler.NewScheduler(gaiascheduler.Dependencies{
		Store: store,
		DB:    db,
		PS:    &plugin.GoPlugin{},
		CA:    ca,
		Vault: va,
	})

	ps := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: scheduler,
	})
	// Build pipeline.
	ps.CreatePipeline(testPipeline)

	// Check if everything went smoothly.
	if testPipeline.StatusType != gaia.CreatePipelineSuccess {
		t.Errorf("create java pipeline failed: %s", testPipeline.Output)
	}
}

func buildPythonPluginTest(t *testing.T) {
	// Create test pipeline.
	v4, _ := uuid.NewV4()
	testPipeline := &gaia.CreatePipeline{
		ID: uuid.Must(v4, nil).String(),
		Pipeline: gaia.Pipeline{
			Name: "PythonTestPipeline",
			Type: gaia.PTypePython,
			Repo: &gaia.GitRepo{URL: "https://github.com/gaia-pipeline/python-example"},
		},
	}
	store, _ := services.StorageService()
	db, _ := services.DefaultMemDBService()
	va, _ := services.DefaultVaultService()
	ca, _ := security.InitCA()
	scheduler, _ := gaiascheduler.NewScheduler(gaiascheduler.Dependencies{
		Store: store,
		DB:    db,
		PS:    &plugin.GoPlugin{},
		CA:    ca,
		Vault: va,
	})

	ps := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: scheduler,
	})
	// Build pipeline.
	ps.CreatePipeline(testPipeline)

	// Check if everything went smoothly.
	if testPipeline.StatusType != gaia.CreatePipelineSuccess {
		t.Errorf("create python pipeline failed: %s", testPipeline.Output)
	}
}

func buildCppPluginTest(t *testing.T) {
	// Create test pipeline.
	v4, _ := uuid.NewV4()
	testPipeline := &gaia.CreatePipeline{
		ID: uuid.Must(v4, nil).String(),
		Pipeline: gaia.Pipeline{
			Name: "CppTestPipeline",
			Type: gaia.PTypeCpp,
			Repo: &gaia.GitRepo{URL: "https://github.com/gaia-pipeline/cpp-example"},
		},
	}
	store, _ := services.StorageService()
	db, _ := services.DefaultMemDBService()
	va, _ := services.DefaultVaultService()
	ca, _ := security.InitCA()
	scheduler, _ := gaiascheduler.NewScheduler(gaiascheduler.Dependencies{
		Store: store,
		DB:    db,
		PS:    &plugin.GoPlugin{},
		CA:    ca,
		Vault: va,
	})

	ps := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: scheduler,
	})
	// Build pipeline.
	ps.CreatePipeline(testPipeline)

	// Check if everything went smoothly.
	if testPipeline.StatusType != gaia.CreatePipelineSuccess {
		t.Errorf("create cpp pipeline failed: %s", testPipeline.Output)
	}
}

func buildRubyPluginTest(t *testing.T) {
	// Create test pipeline.
	v4, _ := uuid.NewV4()
	testPipeline := &gaia.CreatePipeline{
		ID: uuid.Must(v4, nil).String(),
		Pipeline: gaia.Pipeline{
			Name: "RubyTestPipeline",
			Type: gaia.PTypeRuby,
			Repo: &gaia.GitRepo{URL: "https://github.com/gaia-pipeline/ruby-example"},
		},
	}
	store, _ := services.StorageService()
	db, _ := services.DefaultMemDBService()
	va, _ := services.DefaultVaultService()
	ca, _ := security.InitCA()
	scheduler, _ := gaiascheduler.NewScheduler(gaiascheduler.Dependencies{
		Store: store,
		DB:    db,
		PS:    &plugin.GoPlugin{},
		CA:    ca,
		Vault: va,
	})

	ps := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: scheduler,
	})
	// Build pipeline.
	ps.CreatePipeline(testPipeline)

	// Check if everything went smoothly.
	if testPipeline.StatusType != gaia.CreatePipelineSuccess {
		t.Errorf("create ruby pipeline failed: %s", testPipeline.Output)
	}
}

func buildNodeJSPluginTest(t *testing.T) {
	// Create test pipeline.
	v4, _ := uuid.NewV4()
	testPipeline := &gaia.CreatePipeline{
		ID: uuid.Must(v4, nil).String(),
		Pipeline: gaia.Pipeline{
			Name: "NodeJSTestPipeline",
			Type: gaia.PTypeNodeJS,
			Repo: &gaia.GitRepo{URL: "https://github.com/gaia-pipeline/nodejs-example"},
		},
	}
	store, _ := services.StorageService()
	db, _ := services.DefaultMemDBService()
	va, _ := services.DefaultVaultService()
	ca, _ := security.InitCA()
	scheduler, _ := gaiascheduler.NewScheduler(gaiascheduler.Dependencies{
		Store: store,
		DB:    db,
		PS:    &plugin.GoPlugin{},
		CA:    ca,
		Vault: va,
	})

	ps := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: scheduler,
	})
	// Build pipeline.
	ps.CreatePipeline(testPipeline)

	// Check if everything went smoothly.
	if testPipeline.StatusType != gaia.CreatePipelineSuccess {
		t.Errorf("create nodejs pipeline failed: %s", testPipeline.Output)
	}
}
