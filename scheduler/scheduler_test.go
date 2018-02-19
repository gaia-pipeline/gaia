package scheduler

import (
	"hash/fnv"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
	uuid "github.com/satori/go.uuid"
)

func TestScheduleJobsByPriority(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.TestMode = true

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

	p := &gaia.Pipeline{
		ID:   1,
		Name: "Test Pipeline",
		Type: gaia.GOLANG,
	}
	r := &gaia.PipelineRun{
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

	s := NewScheduler(store.NewStore())
	s.scheduleJobsByPriority(r, p)

	// Iterate jobs
	for _, job := range r.Jobs {
		if job.Status != gaia.JobSuccess {
			t.Fatalf("job status should be success but was %s", string(job.Status))
		} else {
			t.Logf("Job %s has been executed...", job.Title)
		}
	}
}

// hash hashes the given string.
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
