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
