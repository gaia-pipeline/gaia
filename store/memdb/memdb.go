package memdb

import (
	"github.com/gaia-pipeline/gaia"
	"github.com/hashicorp/go-memdb"
)

const (
	// Name of the worker table
	workerTableName = "worker"
)

// MemDB represents the implementation of the MemDB interface.
type MemDB struct {
	db *memdb.MemDB
}

// GaiaMemDB is the interface used to talk to the MemDB implementation.
type GaiaMemDB interface {
	// CountWorker counts the number of stored workers.
	CountWorker() (int, error)

	// UpsertWorker inserts or updates the given worker in the memdb.
	UpsertWorker(w *gaia.Worker) error
}

// InitMemDB initiates a new memdb db.
func InitMemDB() (GaiaMemDB, error) {
	// Create new database
	db, err := memdb.NewMemDB(memDBSchema)
	if err != nil {
		return nil, err
	}

	return &MemDB{db: db}, nil
}

// insert inserts a object at the given key path. Is has been
// designed to be used only internally.
// It returns an error in case something badly happened.
func (m *MemDB) insert(key string, value interface{}) error {
	// Create a write transaction
	txn := m.db.Txn(true)

	// Insert object
	if err := txn.Insert(key, value); err != nil {
		// Abort transaction if something went wrong
		txn.Abort()
		return err
	}

	// Commit transaction
	txn.Commit()

	return nil
}

// get returns an object by the given key and id. It has been
// designed to be used only internally.
// If none was found, the returning object will be nil.
// It returns an error in case something badly happened.
func (m *MemDB) get(key, id string) (interface{}, error) {
	// Create a read-only transaction
	txn := m.db.Txn(false)
	defer txn.Abort()

	// Lookup at given key for given id
	raw, err := txn.First(key, "id", id)
	if err != nil {
		return nil, err
	}

	return raw, err
}

// CountWorker counts all worker objects in the worker table.
// It returns an error in case something badly happens.
func (m *MemDB) CountWorker() (count int, err error) {
	// Create a read-only transaction
	txn := m.db.Txn(false)
	defer txn.Abort()

	// Get all objects from the worker table
	iter, err := txn.Get(workerTableName, "id_prefix")
	if err != nil {
		return
	}

	// Iterate through all items and count them. Exit when nil is returned.
	for {
		if item := iter.Next(); item == nil {
			break
		}
		count++
	}
	return
}

// UpsertWorker inserts or updates the given worker in the memdb.
func (m *MemDB) UpsertWorker(w *gaia.Worker) error {
	// Create a write transaction
	txn := m.db.Txn(true)

	// Find existing entry
	raw, err := txn.First(workerTableName, "id", w.UniqueID)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to lookup worker via upsert", "error", err.Error())
		return err
	}

	// Delete if it exists
	if raw != nil {
		err = txn.Delete(workerTableName, raw)
		if err != nil {
			gaia.Cfg.Logger.Error("failed to delete worker via upsert", "error", err.Error())
			return err
		}
	}

	// Insert it
	if err := txn.Insert(workerTableName, w); err != nil {
		gaia.Cfg.Logger.Error("failed to insert worker via upsert", "error", err.Error())
		return err
	}

	return nil
}
