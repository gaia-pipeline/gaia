package handlers

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	rbacProvider "github.com/gaia-pipeline/gaia/providers/rbac"
	userProvider "github.com/gaia-pipeline/gaia/providers/user"
	"github.com/gaia-pipeline/gaia/security/rbac"

	"github.com/gaia-pipeline/gaia/providers/pipelines"
	"github.com/gaia-pipeline/gaia/providers/workers"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/store"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
)

type mockStorageService struct {
	store.GaiaStore
	mockPipeline *gaia.Pipeline
}

func (s *mockStorageService) PipelineGetRunByPipelineIDAndID(pipelineid int, runid int) (*gaia.PipelineRun, error) {
	return generateTestData(), nil
}
func (s *mockStorageService) PipelinePutRun(r *gaia.PipelineRun) error { return nil }
func (s *mockStorageService) PipelineGet(id int) (pipeline *gaia.Pipeline, err error) {
	return s.mockPipeline, nil
}
func (s *mockStorageService) PipelineGetRunByID(runID string) (*gaia.PipelineRun, error) {
	return &gaia.PipelineRun{}, nil
}

func TestInitHandler(t *testing.T) {
	dataDir, err := ioutil.TempDir("", "TestInitHandler")
	if err != nil {
		t.Fatalf("error creating data dir %v", err.Error())
	}

	defer func() {
		gaia.Cfg = nil
		_ = os.RemoveAll(dataDir)
	}()
	gaia.Cfg = &gaia.Config{
		Logger:    hclog.NewNullLogger(),
		DataPath:  dataDir,
		CAPath:    dataDir,
		VaultPath: dataDir,
		HomePath:  dataDir,
		Mode:      gaia.ModeServer,
		DevMode:   true,
	}
	e := echo.New()

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	p := gaia.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
		Repo: &gaia.GitRepo{
			URL: "https://github.com/Codertocat/Hello-World",
		},
	}

	ap.Append(p)
	ms := &mockScheduleService{}
	// Initialize handlers
	pipelineService := pipeline.NewGaiaPipelineService(pipeline.Dependencies{
		Scheduler: ms,
	})
	pp := pipelines.NewPipelineProvider(pipelines.Dependencies{
		Scheduler:       ms,
		PipelineService: pipelineService,
	})
	ca, err := security.InitCA()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create CA", "error", err.Error())
		return
	}
	wp := workers.NewWorkerProvider(workers.Dependencies{
		Certificate: ca,
		Scheduler:   ms,
	})
	mStore := &mockStorageService{mockPipeline: &p}
	rbacService := rbac.NewNoOpService()
	rbacPrv := rbacProvider.NewProvider(rbacService)
	userPrv := userProvider.NewProvider(mStore, rbacService)
	handlerService := NewGaiaHandler(Dependencies{
		Scheduler:        ms,
		PipelineService:  pipelineService,
		PipelineProvider: pp,
		Certificate:      ca,
		WorkerProvider:   wp,
		Store:            mStore,
		UserProvider:     userPrv,
		RBACProvider:     rbacPrv,
	})
	if err := handlerService.InitHandlers(e); err != nil {
		t.Fatal(err)
	}
}

func generateTestData() *gaia.PipelineRun {
	return &gaia.PipelineRun{
		UniqueID:   "first-pipeline-run",
		ID:         1,
		PipelineID: 1,
		Jobs: []*gaia.Job{
			{
				ID:     1,
				Title:  "first-job",
				Status: gaia.JobWaitingExec,
			},
		},
	}
}
