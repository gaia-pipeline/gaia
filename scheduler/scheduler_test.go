package scheduler

import (
	"crypto/tls"
	"hash/fnv"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/store"
	hclog "github.com/hashicorp/go-hclog"
	uuid "github.com/satori/go.uuid"
)

type PluginFake struct{}

func (p *PluginFake) NewPlugin(ca security.CAAPI) Plugin           { return &PluginFake{} }
func (p *PluginFake) Connect(cmd *exec.Cmd, logPath *string) error { return nil }
func (p *PluginFake) Execute(j *gaia.Job) error                    { return nil }
func (p *PluginFake) GetJobs() ([]gaia.Job, error)                 { return prepareJobs(), nil }
func (p *PluginFake) Close()                                       {}

type CAFake struct{}

func (c *CAFake) CreateSignedCert() (string, string, error)                       { return "", "", nil }
func (c *CAFake) GenerateTLSConfig(certPath, keyPath string) (*tls.Config, error) { return nil, nil }
func (c *CAFake) CleanupCerts(crt, key string) error                              { return nil }
func (c *CAFake) GetCACertPath() (string, string)                                 { return "", "" }

func TestInit(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestInit")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
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
	var ca security.CAAPI
	ca = &CAFake{}
	s := NewScheduler(storeInstance, &PluginFake{}, ca)
	err := s.Init()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrepareAndExec(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExec")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
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
	var ca security.CAAPI
	ca = &CAFake{}
	s := NewScheduler(storeInstance, &PluginFake{}, ca)
	s.prepareAndExec(r)

	// Iterate jobs
	for _, job := range r.Jobs {
		if job.Status != gaia.JobSuccess {
			t.Fatalf("job status should be success but was %s", string(job.Status))
		} else {
			t.Logf("Job %s has been executed...", job.Title)
		}
	}
}

func TestSchedulePipeline(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestSchedulePipeline")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
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
	var ca security.CAAPI
	ca = &CAFake{}
	s := NewScheduler(storeInstance, &PluginFake{}, ca)
	err := s.Init()
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SchedulePipeline(&p)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSchedule(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestSchedule")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
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
	var ca security.CAAPI
	ca = &CAFake{}
	s := NewScheduler(storeInstance, &PluginFake{}, ca)
	_, err := s.SchedulePipeline(&p)
	if err != nil {
		t.Fatal(err)
	}
	s.schedule()
	r, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != gaia.RunScheduled {
		t.Fatalf("run has status %s but should be %s\n", r.Status, string(gaia.RunScheduled))
	}
}

func TestSetPipelineJobs(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestSetPipelineJobs")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
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
	var ca security.CAAPI
	ca = &CAFake{}
	s := NewScheduler(storeInstance, &PluginFake{}, ca)
	err := s.SetPipelineJobs(&p)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Jobs) != 4 {
		t.Fatalf("Number of jobs should be 4 but was %d\n", len(p.Jobs))
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
