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
