package store

import (
	"encoding/json"

	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
)

// CreatePipelinePut adds a pipeline which
// is not yet compiled but is about to.
func (s *BoltStore) CreatePipelinePut(p *gaia.CreatePipeline) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(createPipelineBucket)

		// Marshal pipeline object
		m, err := json.Marshal(p)
		if err != nil {
			return err
		}

		// Put pipeline
		return b.Put([]byte(p.ID), m)
	})
}

// CreatePipelineGet returns all available create pipeline
// objects in the store.
func (s *BoltStore) CreatePipelineGet() ([]gaia.CreatePipeline, error) {
	// create list
	var pipelineList []gaia.CreatePipeline

	return pipelineList, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(createPipelineBucket)

		// Iterate all created pipelines.
		// TODO: We might get a huge list here. It might be better
		// to just search for the last 20 elements.
		return b.ForEach(func(k, v []byte) error {
			// create single pipeline object
			p := &gaia.CreatePipeline{}

			// Unmarshal
			err := json.Unmarshal(v, p)
			if err != nil {
				return err
			}

			pipelineList = append(pipelineList, *p)
			return nil
		})
	})
}

// PipelinePut puts a pipeline into the store.
// On persist, the pipeline will get a unique id.
func (s *BoltStore) PipelinePut(p *gaia.Pipeline) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get pipeline bucket
		b := tx.Bucket(pipelineBucket)

		// Generate ID for the pipeline if its new.
		if p.ID == 0 {
			id, err := b.NextSequence()
			if err != nil {
				return err
			}
			p.ID = int(id)
		}

		// Marshal pipeline data into bytes.
		buf, err := json.Marshal(p)
		if err != nil {
			return err
		}

		// Persist bytes to pipelines bucket.
		return b.Put(itob(p.ID), buf)
	})
}

// PipelineGet gets a pipeline by given id.
func (s *BoltStore) PipelineGet(id int) (*gaia.Pipeline, error) {
	var pipeline *gaia.Pipeline

	return pipeline, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineBucket)

		// Get pipeline
		v := b.Get(itob(id))

		// Check if we found the pipeline
		if v == nil {
			return nil
		}

		// Unmarshal pipeline object
		pipeline = &gaia.Pipeline{}
		err := json.Unmarshal(v, pipeline)
		if err != nil {
			return err
		}

		return nil
	})
}

// PipelineGetByName looks up a pipeline by the given name.
// Returns nil if pipeline was not found.
func (s *BoltStore) PipelineGetByName(n string) (*gaia.Pipeline, error) {
	var pipeline *gaia.Pipeline

	return pipeline, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineBucket)

		// Iterate all created pipelines.
		return b.ForEach(func(k, v []byte) error {
			// create single pipeline object
			p := &gaia.Pipeline{}

			// Unmarshal
			err := json.Unmarshal(v, p)
			if err != nil {
				return err
			}

			// Is this pipeline we are looking for?
			if p.Name == n {
				pipeline = p
			}

			return nil
		})
	})
}

// PipelineGetRunHighestID looks for the highest public id for the given pipeline.
func (s *BoltStore) PipelineGetRunHighestID(p *gaia.Pipeline) (int, error) {
	var highestID int

	return highestID, s.db.View(func(tx *bolt.Tx) error {
		// Get Bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipeline runs.
		return b.ForEach(func(k, v []byte) error {
			// create single run object
			r := &gaia.PipelineRun{}

			// Unmarshal
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Is this a run from our pipeline?
			if r.PipelineID == p.ID {
				// Check if the id is higher than what we found before?
				if r.ID > highestID {
					highestID = r.ID
				}
			}

			return nil
		})
	})
}

// PipelinePutRun takes the given pipeline run and puts it into the store.
// If a pipeline run already exists in the store it will be overwritten.
func (s *BoltStore) PipelinePutRun(r *gaia.PipelineRun) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineRunBucket)

		// Marshal data into bytes.
		buf, err := json.Marshal(r)
		if err != nil {
			return err
		}

		// Persist bytes into bucket.
		return b.Put([]byte(r.UniqueID), buf)
	})
}

// PipelineGetScheduled returns the scheduled pipelines with a return limit.
func (s *BoltStore) PipelineGetScheduled(limit int) ([]*gaia.PipelineRun, error) {
	// create returning list
	var runList []*gaia.PipelineRun

	return runList, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipelines.
		return b.ForEach(func(k, v []byte) error {
			// Check if we already reached the limit
			if len(runList) >= limit {
				return nil
			}

			// Unmarshal
			r := &gaia.PipelineRun{}
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Check if the run is still scheduled not executed yet
			if r.Status == gaia.RunNotScheduled {
				// Append to our list
				runList = append(runList, r)
			}

			return nil
		})
	})
}

// PipelineGetRunByPipelineIDAndID looks for pipeline run by given pipeline id and run id.
func (s *BoltStore) PipelineGetRunByPipelineIDAndID(pipelineid int, runid int) (*gaia.PipelineRun, error) {
	var pipelineRun *gaia.PipelineRun

	return pipelineRun, s.db.View(func(tx *bolt.Tx) error {
		// Get Bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipeline runs.
		return b.ForEach(func(k, v []byte) error {
			// create single run object
			r := &gaia.PipelineRun{}

			// Unmarshal
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Is this a run from our pipeline?
			if r.PipelineID == pipelineid && r.ID == runid {
				pipelineRun = r
			}

			return nil
		})
	})
}

// PipelineGetRunByID returns the pipeline run by internal unique id.
func (s *BoltStore) PipelineGetRunByID(runID string) (*gaia.PipelineRun, error) {
	var pipelineRun *gaia.PipelineRun

	return pipelineRun, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineRunBucket)

		// Get pipeline run
		v := b.Get([]byte(runID))

		// Check if we found the pipeline
		if v == nil {
			return nil
		}

		// Unmarshal pipeline object
		err := json.Unmarshal(v, pipelineRun)
		if err != nil {
			return err
		}

		return nil
	})
}

// PipelineGetAllRunsByPipelineID looks for all pipeline runs by the given pipeline id.
func (s *BoltStore) PipelineGetAllRunsByPipelineID(pipelineID int) ([]gaia.PipelineRun, error) {
	var runs []gaia.PipelineRun

	return runs, s.db.View(func(tx *bolt.Tx) error {
		// Get Bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipeline runs.
		return b.ForEach(func(k, v []byte) error {
			// create single run object
			r := &gaia.PipelineRun{}

			// Unmarshal
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Is this a run from our pipeline?
			if r.PipelineID == pipelineID {
				// add this to our list
				runs = append(runs, *r)
			}

			return nil
		})
	})
}

// PipelineGetAllRuns loads all existing pipeline runs.
func (s *BoltStore) PipelineGetAllRuns() ([]gaia.PipelineRun, error) {
	var runs []gaia.PipelineRun

	return runs, s.db.View(func(tx *bolt.Tx) error {
		// Get Bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipeline runs.
		return b.ForEach(func(k, v []byte) error {
			// create single run object
			r := &gaia.PipelineRun{}

			// Unmarshal
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Append this run
			runs = append(runs, *r)

			return nil
		})
	})
}

// PipelineGetLatestRun returns the latest run by the given pipeline id.
func (s *BoltStore) PipelineGetLatestRun(pipelineID int) (*gaia.PipelineRun, error) {
	var run *gaia.PipelineRun

	return run, s.db.View(func(tx *bolt.Tx) error {
		// Get Bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipeline runs.
		return b.ForEach(func(k, v []byte) error {
			// create single run object
			r := &gaia.PipelineRun{}

			// Unmarshal
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Is this a run from our pipeline?
			if r.PipelineID == pipelineID {
				// Check if this is the latest run
				if run == nil || run.StartDate.Before(r.StartDate) {
					// set it
					run = r
				}
			}

			return nil
		})
	})
}

// PipelineDelete deletes the pipeline with the given id.
func (s *BoltStore) PipelineDelete(id int) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineBucket)

		// Delete pipeline
		return b.Delete(itob(id))
	})
}

// PipelineRunDelete deletes the pipeline run with the given id.
func (s *BoltStore) PipelineRunDelete(uniqueID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineRunBucket)

		// Delete pipeline
		return b.Delete([]byte(uniqueID))
	})
}
