package scheduler

import (
	"crypto/tls"
	"errors"
	"hash/fnv"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/plugin"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/store"
	hclog "github.com/hashicorp/go-hclog"
	uuid "github.com/satori/go.uuid"
)

type PluginFake struct{}

func (p *PluginFake) NewPlugin(ca security.CAAPI) plugin.Plugin { return &PluginFake{} }
func (p *PluginFake) Init(cmd *exec.Cmd, logPath *string) error { return nil }
func (p *PluginFake) Validate() error                           { return nil }
func (p *PluginFake) Execute(j *gaia.Job) error {
	j.Status = gaia.JobSuccess
	return nil
}
func (p *PluginFake) GetJobs() ([]*gaia.Job, error) { return prepareJobs(), nil }
func (p *PluginFake) FlushLogs() error              { return nil }
func (p *PluginFake) Close()                        {}

type CAFake struct{}

func (c *CAFake) CreateSignedCertWithValidOpts(hostname string, hoursBeforeValid, hoursAfterValid time.Duration) (string, string, error) {
	return "", "", nil
}
func (c *CAFake) CreateSignedCert() (string, string, error)                       { return "", "", nil }
func (c *CAFake) GenerateTLSConfig(certPath, keyPath string) (*tls.Config, error) { return nil, nil }
func (c *CAFake) CleanupCerts(crt, key string) error                              { return nil }
func (c *CAFake) GetCACertPath() (string, string)                                 { return "", "" }

type VaultFake struct{}

func (v *VaultFake) LoadSecrets() error             { return nil }
func (v *VaultFake) GetAll() []string               { return []string{} }
func (v *VaultFake) SaveSecrets() error             { return nil }
func (v *VaultFake) Add(key string, value []byte)   {}
func (v *VaultFake) Remove(key string)              {}
func (v *VaultFake) Get(key string) ([]byte, error) { return []byte{}, nil }

type MemDBFake struct{}

func (m *MemDBFake) SyncStore() error                                { return nil }
func (m *MemDBFake) GetAllWorker() []*gaia.Worker                    { return []*gaia.Worker{} }
func (m *MemDBFake) UpsertWorker(w *gaia.Worker, persist bool) error { return nil }
func (m *MemDBFake) GetWorker(id string) (*gaia.Worker, error)       { return &gaia.Worker{}, nil }
func (m *MemDBFake) DeleteWorker(id string, persist bool) error      { return nil }
func (m *MemDBFake) InsertPipelineRun(p *gaia.PipelineRun) error     { return nil }
func (m *MemDBFake) PopPipelineRun(tags []string) (*gaia.PipelineRun, error) {
	return &gaia.PipelineRun{}, nil
}
func (m *MemDBFake) UpsertSHAPair(pair gaia.SHAPair) error {
	return nil
}
func (m *MemDBFake) GetSHAPair(pipelineID string) (ok bool, pair gaia.SHAPair, err error) {
	return
}

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
	gaia.Cfg.Worker = 2
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	s.Init()
}

type PluginFakeFailed struct{}

func (p *PluginFakeFailed) NewPlugin(ca security.CAAPI) plugin.Plugin { return &PluginFakeFailed{} }
func (p *PluginFakeFailed) Init(cmd *exec.Cmd, logPath *string) error { return nil }
func (p *PluginFakeFailed) Validate() error                           { return nil }
func (p *PluginFakeFailed) Execute(j *gaia.Job) error {
	j.Status = gaia.JobFailed
	j.FailPipeline = true
	return errors.New("job failed")
}
func (p *PluginFakeFailed) GetJobs() ([]*gaia.Job, error) { return prepareJobs(), nil }
func (p *PluginFakeFailed) FlushLogs() error              { return nil }
func (p *PluginFakeFailed) Close()                        {}

func TestPrepareAndExecFail(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExecFail")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// jobs should be existent
	if len(run.Jobs) == 0 {
		t.Fatal("No jobs in pipeline run found.")
	}

	// Check run status
	if run.Status != gaia.RunFailed {
		t.Fatalf("Run should be of type %s but was %s\n", gaia.RunFailed, run.Status)
	}
}

func TestPrepareAndExecInvalidType(t *testing.T) {
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

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	p.Type = gaia.PTypeUnknown
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Check run status
	if run.Status != gaia.RunFailed {
		t.Fatalf("Run should be of type %s but was %s\n", gaia.RunFailed, run.Status)
	}
}

func TestPrepareAndExecJavaType(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExecJavaType")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	javaExecName = "go"
	p.Type = gaia.PTypeJava
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// jobs should be existent
	if len(run.Jobs) == 0 {
		t.Fatal("No jobs in pipeline run found.")
	}

	// Iterate jobs
	for _, job := range run.Jobs {
		if job.Status != gaia.JobSuccess {
			t.Fatalf("job status should be success but was %s", string(job.Status))
		} else {
			t.Logf("Job %s has been executed...", job.Title)
		}
	}
}

func TestPrepareAndExecPythonType(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExecPythonType")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	pythonExecName = "go"
	p.Type = gaia.PTypePython
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// jobs should be existent
	if len(run.Jobs) == 0 {
		t.Fatal("No jobs in pipeline run found.")
	}

	// Iterate jobs
	for _, job := range run.Jobs {
		if job.Status != gaia.JobSuccess {
			t.Fatalf("job status should be success but was %s", string(job.Status))
		} else {
			t.Logf("Job %s has been executed...", job.Title)
		}
	}
}

func TestPrepareAndExecCppType(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExecCppType")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	p.Type = gaia.PTypeCpp
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// jobs should be existent
	if len(run.Jobs) == 0 {
		t.Fatal("No jobs in pipeline run found.")
	}

	// Iterate jobs
	for _, job := range run.Jobs {
		if job.Status != gaia.JobSuccess {
			t.Fatalf("job status should be success but was %s", string(job.Status))
		} else {
			t.Logf("Job %s has been executed...", job.Title)
		}
	}
}

func TestPrepareAndExecRubyType(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExecRubyType")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	p.Type = gaia.PTypeRuby
	rubyExecName = "go"
	rubyGemName = "echo"
	findRubyGemCommands = []string{"name: rubytest"}
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// jobs should be existent
	if len(run.Jobs) == 0 {
		t.Fatal("No jobs in pipeline run found.")
	}

	// Iterate jobs
	for _, job := range run.Jobs {
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
	gaia.Cfg.Worker = 2
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, _ := prepareTestData()
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	s.Init()
	_, err = s.SchedulePipeline(&p, prepareArgs())
	if err != nil {
		t.Fatal(err)
	}
}

func TestSchedulePipelineParallel(t *testing.T) {
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
	gaia.Cfg.Worker = 2
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p1 := gaia.Pipeline{
		ID:   0,
		Name: "Test Pipeline 1",
		Type: gaia.PTypeGolang,
		Jobs: prepareJobs(),
	}
	p2 := gaia.Pipeline{
		ID:   1,
		Name: "Test Pipeline 2",
		Type: gaia.PTypeGolang,
		Jobs: prepareJobs(),
	}
	_ = storeInstance.PipelinePut(&p1)
	_ = storeInstance.PipelinePut(&p2)
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	s.Init()
	var run1 *gaia.PipelineRun
	var run2 *gaia.PipelineRun
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		run1, _ = s.SchedulePipeline(&p1, prepareArgs())
		wg.Done()
	}()
	go func() {
		run2, _ = s.SchedulePipeline(&p2, prepareArgs())
		wg.Done()
	}()
	wg.Wait()
	if run1.ID == run2.ID {
		t.Fatal("the two run jobs id should not have equalled. was: ", run1.ID, run2.ID)
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
	gaia.Cfg.Worker = 2
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, _ := prepareTestData()
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SchedulePipeline(&p, prepareArgs())
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
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, _ := prepareTestData()
	p.Jobs = nil
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	err = s.SetPipelineJobs(&p)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Jobs) != 4 {
		t.Fatalf("Number of jobs should be 4 but was %d\n", len(p.Jobs))
	}
}

func TestStopPipelineRunFailIfPipelineNotInRunningState(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestStopPipelineRunFailIfPipelineNotInRunningState")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	gaia.Cfg.Worker = 2
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, _ := prepareTestData()
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SchedulePipeline(&p, prepareArgs())
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
	err = s.StopPipelineRun(&p, 1)
	if err == nil {
		t.Fatal("error was nil. should have failed")
	}
	if err.Error() != "pipeline is not in running state" {
		t.Fatal("error was not what was expected 'pipeline is not in running state'. got: ", err.Error())
	}
}

func TestStopPipelineRun(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestStopPipelineRun")
	gaia.Cfg.DataPath = tmp
	gaia.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	gaia.Cfg.Bolt.Mode = 0600
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	gaia.Cfg.Worker = 2
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{})
	if err != nil {
		t.Fatal(err)
	}

	r.Status = gaia.RunRunning
	_ = storeInstance.PipelinePutRun(&r)

	run, _ := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	err = s.StopPipelineRun(&p, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	run, _ = storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if run.Status != gaia.RunCancelled {
		t.Fatal("expected pipeline state to be cancelled. got: ", r.Status)
	}
}

func prepareArgs() []*gaia.Argument {
	arg1 := gaia.Argument{
		Description: "First Arg",
		Key:         "firstarg",
		Type:        "textfield",
	}
	arg2 := gaia.Argument{
		Description: "Second Arg",
		Key:         "secondarg",
		Type:        "textarea",
	}
	arg3 := gaia.Argument{
		Description: "Vault Arg",
		Key:         "vaultarg",
		Type:        "vault",
	}
	return []*gaia.Argument{&arg1, &arg2, &arg3}
}

func prepareJobs() []*gaia.Job {
	job1 := gaia.Job{
		ID:        hash("Job1"),
		Title:     "Job1",
		DependsOn: []*gaia.Job{},
		Status:    gaia.JobWaitingExec,
		Args:      prepareArgs(),
	}
	job2 := gaia.Job{
		ID:        hash("Job2"),
		Title:     "Job2",
		DependsOn: []*gaia.Job{&job1},
		Status:    gaia.JobWaitingExec,
	}
	job3 := gaia.Job{
		ID:        hash("Job3"),
		Title:     "Job3",
		DependsOn: []*gaia.Job{&job2},
		Status:    gaia.JobWaitingExec,
	}
	job4 := gaia.Job{
		ID:        hash("Job4"),
		Title:     "Job4",
		DependsOn: []*gaia.Job{&job3},
		Status:    gaia.JobWaitingExec,
	}

	return []*gaia.Job{
		&job1,
		&job2,
		&job3,
		&job4,
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
		Jobs:       pipeline.Jobs,
	}
	return
}

// hash hashes the given string.
func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
