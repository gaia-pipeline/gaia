package store

import (
	"encoding/json"

	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
)

// UpsertSHAPair creates or updates a record for a SHA pair of the original SHA and the
// rebuilt Worker SHA for a pipeline.
func (s *BoltStore) UpsertSHAPair(pair gaia.SHAPair) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(shaPairBucket)

		// Marshal SHAPair struct
		m, err := json.Marshal(pair)
		if err != nil {
			return err
		}

		// Put SHAPair
		return b.Put(itob(pair.PipelineID), m)
	})
}

// GetSHAPair returns a pair of shas for this pipeline run.
func (s *BoltStore) GetSHAPair(pipelineID int) (ok bool, pair gaia.SHAPair, err error) {
	return ok, pair, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(shaPairBucket)

		// Get SHAPair
		v := b.Get(itob(pipelineID))

		// Check if we found the SHAPair
		if v == nil {
			ok = false
			return nil
		}

		// Unmarshal SHAPair struct
		err := json.Unmarshal(v, &pair)
		if err != nil {
			return err
		}
		ok = true
		return nil
	})
}
