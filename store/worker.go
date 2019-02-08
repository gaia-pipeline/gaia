package store

import (
	"encoding/json"

	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
)

// WorkerPut stores the given worker in the bolt database.
// Worker object will be overwritten in case it already exist.
func (s *BoltStore) WorkerPut(w *gaia.Worker) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(workerBucket)

		// Marshal worker object
		m, err := json.Marshal(w)
		if err != nil {
			return err
		}

		// Put worker
		return b.Put([]byte(w.UniqueID), m)
	})
}
