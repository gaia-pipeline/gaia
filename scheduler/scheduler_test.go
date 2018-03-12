package scheduler

import (
	"fmt"
	"hash/fnv"
	"os"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
	uuid "github.com/satori/go.uuid"
)

func TestScheduleJobsByPriority(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	storeInstance := store.NewStore()
	gaia.Cfg.DataPath = "data"
	gaia.Cfg.Bolt.Mode = 0600

	// Create test folder
	err := os.MkdirAll(gaia.Cfg.DataPath, 0700)
	if err != nil {
		fmt.Printf("cannot create data folder: %s\n", err.Error())
		t.Fatal(err)
	}

	if err = storeInstance.Init(); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	s := NewScheduler(storeInstance)
	s.scheduleJobsByPriority(r, p)

	// Iterate jobs
	for _, job := range r.Jobs {
		if job.Status != gaia.JobSuccess {
			t.Fatalf("job status should be success but was %s", string(job.Status))
		} else {
			t.Logf("Job %s has been executed...", job.Title)
		}
	}

	// cleanup
	err = os.Remove("data/gaia.db")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove("data")
	if err != nil {
		t.Fatal(err)
	}
}

func prepareTestData() (pipeline *gaia.Pipeline, pipelineRun *gaia.PipelineRun) {
	job1 := gaia.Job{
		ID:       hash("Job1"),
		Title:    "Job1",
		Priority: 0,
		Status:   gaia.JobSuccess,
	}
	job2 := gaia.Job{
		ID:       hash("Job2"),
		Title:    "Job2",
		Priority: 10,
		Status:   gaia.JobSuccess,
	}
	job3 := gaia.Job{
		ID:       hash("Job3"),
		Title:    "Job3",
		Priority: 20,
		Status:   gaia.JobSuccess,
	}
	job4 := gaia.Job{
		ID:       hash("Job4"),
		Title:    "Job4",
		Priority: 20,
		Status:   gaia.JobSuccess,
	}

	pipeline = &gaia.Pipeline{
		ID:   1,
		Name: "Test Pipeline",
		Type: gaia.PTypeGolang,
	}
	pipelineRun = &gaia.PipelineRun{
		ID:         1,
		PipelineID: 1,
		Status:     gaia.RunNotScheduled,
		UniqueID:   uuid.Must(uuid.NewV4()).String(),
		Jobs: []gaia.Job{
			job1,
			job2,
			job3,
			job4,
		},
	}
	return
}

// hash hashes the given string.
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
