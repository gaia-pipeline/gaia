package scheduler

import (
	"sync"

	"github.com/gaia-pipeline/gaia"
)

// workload is a wrapper around a single job object.
type workload struct {
	finishedSig chan bool
	done        bool
	started     bool
	job         gaia.Job
}

// managedWorkloads holds workloads.
// managedWorkloads can be safely shared between goroutines.
type managedWorkloads struct {
	sync.RWMutex

	workloads []workload
}

// newManagedWorkloads creates a new instance of managedWorkloads.
func newManagedWorkloads() *managedWorkloads {
	mw := &managedWorkloads{
		workloads: make([]workload, 0),
	}

	return mw
}

// Append appends a new workload to managedWorkloads.
func (mw *managedWorkloads) Append(wl workload) {
	mw.Lock()
	defer mw.Unlock()

	mw.workloads = append(mw.workloads, wl)
}

// GetByID looks up the workload by the given id.
func (mw *managedWorkloads) GetByID(id uint32) *workload {
	var foundWorkload workload
	for wl := range mw.Iter() {
		if wl.job.ID == id {
			foundWorkload = wl
		}
	}

	if foundWorkload.job.Title == "" {
		return nil
	}

	return &foundWorkload
}

// Replace takes the given workload and replaces it in the managedWorkloads
// slice. Return true when success otherwise false.
func (mw *managedWorkloads) Replace(wl workload) bool {
	mw.Lock()
	defer mw.Unlock()

	// Search for the id
	i := -1
	for id, currWL := range mw.workloads {
		if currWL.job.ID == wl.job.ID {
			i = id
			break
		}
	}

	// We got it?
	if i == -1 {
		return false
	}

	// Yes
	mw.workloads[i] = wl
	return true
}

// Iter iterates over the workloads in the concurrent slice.
func (mw *managedWorkloads) Iter() <-chan workload {
	c := make(chan workload)

	go func() {
		mw.RLock()
		defer mw.RUnlock()
		for _, mw := range mw.workloads {
			c <- mw
		}
		close(c)
	}()

	return c
}
