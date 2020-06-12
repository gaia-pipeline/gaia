package store

import (
	"encoding/json"

	bolt "go.etcd.io/bbolt"

	"github.com/gaia-pipeline/gaia"
)

// WorkerPut stores the given worker in the bolt database.
// Worker object will be overwritten in case it already exist.
func (s *BoltStore) WorkerPut(w *gaia.Worker) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(workerBucket)

		// Marshal worker object
		m, err := json.Marshal(*w)
		if err != nil {
			return err
		}

		// Put worker
		return b.Put([]byte(w.UniqueID), m)
	})
}

// WorkerGetAll returns all existing worker objects from the store.
// It returns an error when the action failed.
func (s *BoltStore) WorkerGetAll() ([]*gaia.Worker, error) {
	var worker []*gaia.Worker

	return worker, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(workerBucket)

		// Iterate all worker.
		return b.ForEach(func(k, v []byte) error {
			// Unmarshal
			w := &gaia.Worker{}
			err := json.Unmarshal(v, w)
			if err != nil {
				return err
			}

			// Append to our list
			worker = append(worker, w)

			return nil
		})
	})
}

// WorkerDeleteAll deletes all worker objects in the bucket.
func (s *BoltStore) WorkerDeleteAll() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Delete bucket
		if err := tx.DeleteBucket(workerBucket); err != nil {
			return err
		}

		_, err := tx.CreateBucket(workerBucket)
		return err
	})
}

// WorkerDelete deletes a worker by the given identifier.
func (s *BoltStore) WorkerDelete(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(workerBucket)

		// Delete entry
		return b.Delete([]byte(id))
	})
}

// WorkerGet gets a worker by the given identifier.
func (s *BoltStore) WorkerGet(id string) (*gaia.Worker, error) {
	var worker *gaia.Worker

	return worker, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(workerBucket)

		// Get worker
		v := b.Get([]byte(id))

		// Check if we found the worker
		if v == nil {
			return nil
		}

		// Unmarshal pipeline object
		worker = &gaia.Worker{}
		err := json.Unmarshal(v, worker)
		if err != nil {
			return err
		}

		return nil
	})
}
