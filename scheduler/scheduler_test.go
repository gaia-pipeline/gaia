package scheduler

import (
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
	hclog "github.com/hashicorp/go-hclog"
	uuid "github.com/satori/go.uuid"
)

type PluginFake struct{}

func (p *PluginFake) NewPlugin() Plugin                            { return &PluginFake{} }
func (p *PluginFake) Connect(cmd *exec.Cmd, logPath *string) error { return nil }
func (p *PluginFake) Execute(j *gaia.Job) error                    { return nil }
func (p *PluginFake) GetJobs() ([]gaia.Job, error)                 { return prepareJobs(), nil }
func (p *PluginFake) Close()                                       {}

func TestInit(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewStore()
	gaia.Cfg.DataPath = os.TempDir()
	gaia.Cfg.WorkspacePath = filepath.Join(os.TempDir(), "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	gaia.Cfg.Worker = "2"
	if err := storeInstance.Init(); err != nil {
		t.Fatal(err)
	}
	s := NewScheduler(storeInstance, &PluginFake{})
	err := s.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove(filepath.Join(os.TempDir(), "gaia.db"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrepareAndExec(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewStore()
	gaia.Cfg.DataPath = os.TempDir()
	gaia.Cfg.WorkspacePath = filepath.Join(os.TempDir(), "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})

	if err := storeInstance.Init(); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	storeInstance.PipelinePut(&p)
	s := NewScheduler(storeInstance, &PluginFake{})
	s.prepareAndExec(r)

	// Iterate jobs
	for _, job := range r.Jobs {
		if job.Status != gaia.JobSuccess {
			t.Fatalf("job status should be success but was %s", string(job.Status))
		} else {
			t.Logf("Job %s has been executed...", job.Title)
		}
	}
	err := os.Remove(filepath.Join(os.TempDir(), "gaia.db"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSchedulePipeline(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewStore()
	gaia.Cfg.DataPath = os.TempDir()
	gaia.Cfg.WorkspacePath = filepath.Join(os.TempDir(), "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	gaia.Cfg.Worker = "2"
	if err := storeInstance.Init(); err != nil {
		t.Fatal(err)
	}
	p, _ := prepareTestData()
	storeInstance.PipelinePut(&p)
	s := NewScheduler(storeInstance, &PluginFake{})
	err := s.Init()
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SchedulePipeline(&p)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Remove(filepath.Join(os.TempDir(), "gaia.db"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSchedule(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewStore()
	gaia.Cfg.DataPath = os.TempDir()
	gaia.Cfg.WorkspacePath = filepath.Join(os.TempDir(), "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	gaia.Cfg.Worker = "2"
	if err := storeInstance.Init(); err != nil {
		t.Fatal(err)
	}
	p, _ := prepareTestData()
	storeInstance.PipelinePut(&p)
	s := NewScheduler(storeInstance, &PluginFake{})
	err := s.Init()
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SchedulePipeline(&p)
	if err != nil {
		t.Fatal(err)
	}
	// Wait some time to pickup work and finish.
	// We have to wait at least 3 seconds for scheduler tick interval.
	time.Sleep(5 * time.Second)
	r, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, job := range r.Jobs {
		if job.Status != gaia.JobSuccess {
			t.Fatalf("Job %s has status %s but should be %s!\n", job.Title, string(job.Status), string(gaia.JobSuccess))
		}
	}
	err = os.Remove(filepath.Join(os.TempDir(), "gaia.db"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetPipelineJobs(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewStore()
	gaia.Cfg.DataPath = os.TempDir()
	gaia.Cfg.WorkspacePath = filepath.Join(os.TempDir(), "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	if err := storeInstance.Init(); err != nil {
		t.Fatal(err)
	}
	p, _ := prepareTestData()
	p.Jobs = nil
	s := NewScheduler(storeInstance, &PluginFake{})
	err := s.SetPipelineJobs(&p)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Jobs) != 4 {
		t.Fatalf("Number of jobs should be 4 but was %d\n", len(p.Jobs))
	}
	err = os.Remove(filepath.Join(os.TempDir(), "gaia.db"))
	if err != nil {
		t.Fatal(err)
	}
}

func prepareJobs() []gaia.Job {
	job1 := gaia.Job{
		ID:       hash("Job1"),
		Title:    "Job1",
		Priority: 0,
		Status:   gaia.JobWaitingExec,
	}
	job2 := gaia.Job{
		ID:       hash("Job2"),
		Title:    "Job2",
		Priority: 10,
		Status:   gaia.JobWaitingExec,
	}
	job3 := gaia.Job{
		ID:       hash("Job3"),
		Title:    "Job3",
		Priority: 20,
		Status:   gaia.JobWaitingExec,
	}
	job4 := gaia.Job{
		ID:       hash("Job4"),
		Title:    "Job4",
		Priority: 20,
		Status:   gaia.JobWaitingExec,
	}

	return []gaia.Job{
		job1,
		job2,
		job3,
		job4,
	}
}

func prepareTestData() (pipeline gaia.Pipeline, pipelineRun gaia.PipelineRun) {
	pipeline = gaia.Pipeline{
		ID:   1,
		Name: "Test Pipeline",
		Type: gaia.PTypeGolang,
		Jobs: prepareJobs(),
	}
	pipelineRun = gaia.PipelineRun{
		ID:         1,
		PipelineID: 1,
		Status:     gaia.RunNotScheduled,
		UniqueID:   uuid.Must(uuid.NewV4(), nil).String(),
	}
	return
}

// hash hashes the given string.
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
