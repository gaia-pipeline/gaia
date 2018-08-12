package scheduler

import (
	"strconv"
	"sync"
	"testing"

	"github.com/gaia-pipeline/gaia"
)

func TestNewWorkload(t *testing.T) {
	mw := newManagedWorkloads()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			finished := make(chan bool)
			title := strconv.Itoa(j)
			wl := workload{
				done:        true,
				finishedSig: finished,
				job: gaia.Job{
					Description: "Test job",
					ID:          uint32(j),
					Title:       "Test " + title,
				},
				started: true,
			}
			mw.Append(wl)
		}(i)
	}
	wg.Wait()
	if len(mw.workloads) != 10 {
		t.Fatal("workload len want: 10, was:", len(mw.workloads))
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			wl := mw.GetByID(uint32(j))
			if wl == nil {
				t.Fatal("failed to find a job that was created previously. failed id: ", j)
			}
		}(i)
	}
	wg.Wait()
}

func TestReplaceWorkloadFlow(t *testing.T) {
	mw := newManagedWorkloads()
	finished := make(chan bool)
	wl := workload{
		done:        true,
		finishedSig: finished,
		job: gaia.Job{
			Description: "Test job",
			ID:          1,
			Title:       "Test",
		},
		started: true,
	}
	mw.Append(wl)
	t.Run("replace works", func(t *testing.T) {
		replaceWl := workload{
			done:        true,
			finishedSig: finished,
			job: gaia.Job{
				Description: "Test job replaced",
				ID:          1,
				Title:       "Test replaced",
			},
			started: true,
		}
		v := mw.Replace(replaceWl)
		if !v {
			t.Fatalf("return should be true. was false.")
		}
		l := mw.GetByID(1)
		if l.job.Title != "Test replaced" {
			t.Fatalf("got title: %s. wanted: 'Test replaced'", l.job.Title)
		}
	})

	t.Run("returns false if workload was not found", func(t *testing.T) {
		replaceWl := workload{
			done:        true,
			finishedSig: finished,
			job: gaia.Job{
				Description: "Test job replaced",
				ID:          2,
				Title:       "Test replaced",
			},
			started: true,
		}
		v := mw.Replace(replaceWl)
		if v {
			t.Fatalf("return should be false. was true.")
		}
		l := mw.GetByID(2)
		if l != nil {
			t.Fatal("should have not found id 2 which was replaced:", l)
		}
	})
}
