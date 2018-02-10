package store

import (
	"encoding/json"

	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
)

// CreatePipelinePut adds a pipeline which
// is not yet compiled but is about to.
func (s *Store) CreatePipelinePut(p *gaia.CreatePipeline) error {
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
func (s *Store) CreatePipelineGet() ([]gaia.CreatePipeline, error) {
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
func (s *Store) PipelinePut(p *gaia.Pipeline) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get pipeline bucket
		b := tx.Bucket(pipelineBucket)

		// Generate ID for the pipeline.
		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		p.ID = int(id)

		// Marshal pipeline data into bytes.
		buf, err := json.Marshal(p)
		if err != nil {
			return err
		}

		// Persist bytes to pipelines bucket.
		return b.Put(itob(p.ID), buf)
	})
}

// PipelineGetByName looks up a pipeline by the given name.
// Returns nil if pipeline was not found.
func (s *Store) PipelineGetByName(n string) (*gaia.Pipeline, error) {
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

// PipelineGetScheduled returns the scheduled pipelines
//func (s *Store) PipelineGetScheduled() ([]gaia.Pipeline, error) {}
