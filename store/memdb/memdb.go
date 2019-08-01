package memdb

import (
	"errors"
	"github.com/gaia-pipeline/gaia/workers/docker"
	"time"

	"github.com/gaia-pipeline/gaia/helper/stringhelper"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
	memdb "github.com/hashicorp/go-memdb"
)

const (
	// Name of the worker table
	workerTableName = "worker"

	// Name of the pipeline run table
	pipelineRunTable = "pipelinerun"

	// Name of the docker worker table
	dockerWorkerTableName = "dockerworker"
)

// MemDB represents the implementation of the MemDB interface.
type MemDB struct {
	db *memdb.MemDB

	// Instance of store where changes to the memdb are stored.
	store store.GaiaStore
}

// GaiaMemDB is the interface used to talk to the MemDB implementation.
type GaiaMemDB interface {
	// SyncStore syncs the memdb with the store.
	SyncStore() error

	// GetAllWorker returns all worker.
	GetAllWorker() []*gaia.Worker

	// UpsertWorker inserts or updates the given worker in the memdb.
	// If persist is true, the given worker will be persisted in the store.
	UpsertWorker(w *gaia.Worker, persist bool) error

	// GetWorker returns a worker by the given identifier.
	GetWorker(id string) (*gaia.Worker, error)

	// DeleteWorker deletes a worker by the given identifier.
	// If persist is true, the worker will also be deleted from the store.
	DeleteWorker(id string, persist bool) error

	// InsertPipelineRun inserts a pipeline run in the memdb.
	InsertPipelineRun(p *gaia.PipelineRun) error

	// PopPipelineRun gets the oldest pipeline run by tags and removes it immediately
	// from the memdb.
	PopPipelineRun(tags []string) (*gaia.PipelineRun, error)

	// InsertDockerWorker inserts a docker worker into the memdb.
	InsertDockerWorker(w *docker.Worker) error

	// GetDockerWorker gets the docker worker by the given worker id.
	GetDockerWorker(workerID string) (*docker.Worker, error)

	// DeleteDockerWorker deletes the docker worker by then given worker id.
	DeleteDockerWorker(workerID string) error
}

// InitMemDB initiates a new memdb db.
func InitMemDB(s store.GaiaStore) (GaiaMemDB, error) {
	// Store must be existent
	if s == nil {
		return nil, errors.New("store is nil")
	}

	// Create new database
	db, err := memdb.NewMemDB(memDBSchema)
	if err != nil {
		return nil, err
	}

	return &MemDB{db: db, store: s}, nil
}

// SyncStore syncs the memdb with the store.
func (m *MemDB) SyncStore() error {
	// Load all worker objects from store
	worker, err := m.store.WorkerGetAll()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to load worker from store", "error", err.Error())
		return err
	}
	for _, w := range worker {
		if err = m.UpsertWorker(w, false); err != nil {
			return err
		}
	}
	return nil
}

// GetAllWorker returns all worker.
func (m *MemDB) GetAllWorker() []*gaia.Worker {
	var workers []*gaia.Worker

	// Create a read-only transaction
	txn := m.db.Txn(false)
	defer txn.Abort()

	// Get all objects from the worker table
	iter, err := txn.Get(workerTableName, "id_prefix")
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get worker objects from memdb via countworker", "error", err.Error())
		return workers
	}

	// Iterate through all items and add them
	for {
		item := iter.Next()

		if item == nil {
			break
		}

		// Convert into worker obj
		w, ok := item.(*gaia.Worker)
		if !ok {
			gaia.Cfg.Logger.Error("failed to convert worker into worker obj", "raw", item)
			continue
		}

		workers = append(workers, w)
	}

	return workers
}

// UpsertWorker inserts or updates the given worker in the memdb.
// If persist is true, the given worker will be persisted in the store.
func (m *MemDB) UpsertWorker(w *gaia.Worker, persist bool) error {
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

	// Store the worker object in the store first before we commit
	if persist {
		if err = m.store.WorkerPut(w); err != nil {
			gaia.Cfg.Logger.Error("failed to store worker in the store via upsert", "error", err.Error())
			txn.Abort()
			return err
		}
	}

	// Commit transaction
	txn.Commit()

	return nil
}

// GetWorker returns a worker by the given identifier.
func (m *MemDB) GetWorker(id string) (*gaia.Worker, error) {
	// Create read transaction
	txn := m.db.Txn(false)
	defer txn.Abort()

	// Get worker
	raw, err := txn.First(workerTableName, "id", id)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get worker from memdb", "error", err.Error(), "id", id)
		return nil, err
	}

	// If nil we couldn't find it
	if raw == nil {
		return nil, nil
	}

	// Convert into worker obj
	w, ok := raw.(*gaia.Worker)
	if !ok {
		gaia.Cfg.Logger.Error("failed to convert worker into worker obj", "raw", raw)
		return nil, errors.New("failed to convert worker into worker obj")
	}

	return w, nil
}

// DeleteWorker deletes a worker from the memdb.
// If persist is true, the worker will also be deleted from the store.
func (m *MemDB) DeleteWorker(id string, persist bool) error {
	// Create write transaction
	txn := m.db.Txn(true)

	// Find existing entry
	raw, err := txn.First(workerTableName, "id", id)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to load worker from memdb", "error", err.Error(), "id", id)
		return err
	}

	// Found existing entry
	if raw != nil {
		err = txn.Delete(workerTableName, raw)
		if err != nil {
			gaia.Cfg.Logger.Error("failed to delete worker from memdb", "error", err.Error(), "id", id)
			return err
		}
	}

	// Delete from store
	if persist {
		err = m.store.WorkerDelete(id)
		if err != nil {
			gaia.Cfg.Logger.Error("failed to delete worker from store via memdb delete", "error", err.Error(), "id", id)
			return err
		}
	}

	// Commit transaction
	txn.Commit()

	return nil
}

// InsertPipelineRun inserts a pipeline run into the memdb.
func (m *MemDB) InsertPipelineRun(p *gaia.PipelineRun) error {
	// Create a write transaction
	txn := m.db.Txn(true)

	// Insert the pipeline run
	if err := txn.Insert(pipelineRunTable, p); err != nil {
		gaia.Cfg.Logger.Error("failed to insert pipeline run via insert", "error", err.Error())
		return err
	}

	// Commit transaction
	txn.Commit()

	return nil
}

// PopPipelineRun gets the oldest pipeline run filtered by tags and removes it immediately
// from the memdb.
func (m *MemDB) PopPipelineRun(tags []string) (*gaia.PipelineRun, error) {
	// Create a read transaction
	txn := m.db.Txn(false)

	// Get all objects from the pipeline run table
	iter, err := txn.Get(pipelineRunTable, "id_prefix")
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get pipeline run objects from memdb via poppipelinerun", "error", err.Error())
		return nil, err
	}

	// Iterate through all items
	var oldestPipelineRunID string
	var oldestPipelineRunDate time.Time
RunLoop:
	for {
		item := iter.Next()
		if item == nil {
			break
		}

		// Convert into pipeline run object
		pipelineRun, ok := item.(*gaia.PipelineRun)
		if !ok {
			gaia.Cfg.Logger.Error("failed to convert pipeline run to data struct via poppipelinerun", "item", item)
			continue
		}

		// Filter by pipeline type
		if !stringhelper.IsContainedInSlice(tags, pipelineRun.PipelineType.String(), true) {
			continue
		}

		// Filter by tags
		for _, pipelineTag := range pipelineRun.PipelineTags {
			// Find a match
			if !stringhelper.IsContainedInSlice(tags, pipelineTag, true) {
				continue RunLoop
			}
		}

		// Check if the current pipeline run is older than the previous one
		if oldestPipelineRunID == "" || oldestPipelineRunDate.After(pipelineRun.ScheduleDate) {
			oldestPipelineRunID = pipelineRun.UniqueID
			oldestPipelineRunDate = pipelineRun.ScheduleDate
		}
	}

	// Finish read transaction
	txn.Abort()

	// Check if we found a valid pipeline run to pop
	if oldestPipelineRunID != "" {
		// Create a write transaction
		txn := m.db.Txn(true)

		// Get the pipeline run
		pipelineRunRaw, err := txn.First(pipelineRunTable, "id", oldestPipelineRunID)
		if err != nil {
			gaia.Cfg.Logger.Error("failed to get oldest pipeline run via poppipelinerun", "error", err.Error())
			return nil, err
		}

		// Delete pipeline run from memdb
		if err = txn.Delete(pipelineRunTable, pipelineRunRaw); err != nil {
			gaia.Cfg.Logger.Error("failed to delete oldest pipeline run via poppipelinerun", "error", err.Error())
			return nil, err
		}

		// Commit transaction
		txn.Commit()

		// Convert pipeline run
		pipelineRun, ok := pipelineRunRaw.(*gaia.PipelineRun)
		if !ok {
			gaia.Cfg.Logger.Error("failed to convert pipeline run via poppipelinerun", "item", pipelineRunRaw)
			return nil, err
		}

		// Return pipeline run
		return pipelineRun, nil
	}

	return nil, nil
}

// InsertDockerWorker inserts a docker worker into the memdb.
func (m *MemDB) InsertDockerWorker(w *docker.Worker) error {
	// Create a write transaction
	txn := m.db.Txn(true)

	// Insert the docker worker object
	if err := txn.Insert(dockerWorkerTableName, w); err != nil {
		gaia.Cfg.Logger.Error("failed to insert docker worker via insert", "error", err.Error())
		return err
	}

	// Commit transaction
	txn.Commit()

	return nil
}

// GetDockerWorker returns the docker worker by the given worker id.
func (m *MemDB) GetDockerWorker(workerID string) (*docker.Worker, error) {
	// Create read transaction
	txn := m.db.Txn(false)
	defer txn.Abort()

	// Get worker
	raw, err := txn.First(dockerWorkerTableName, "id", workerID)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get docker worker from memdb", "error", err.Error(), "id", workerID)
		return nil, err
	}

	// If nil we couldn't find it
	if raw == nil {
		return nil, nil
	}

	// Convert into docker worker obj
	w, ok := raw.(*docker.Worker)
	if !ok {
		gaia.Cfg.Logger.Error("failed to convert docker worker into docker worker obj", "raw", raw)
		return nil, errors.New("failed to convert docker worker into docker worker obj")
	}

	return w, nil
}

// DeleteDockerWorker deletes the docker worker by then given worker id.
func (m *MemDB) DeleteDockerWorker(workerID string) error {
	// Create write transaction
	txn := m.db.Txn(true)

	// Find existing entry
	raw, err := txn.First(dockerWorkerTableName, "id", workerID)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to load docker worker from memdb", "error", err.Error(), "id", workerID)
		return err
	}

	// Found existing entry
	if raw != nil {
		err = txn.Delete(dockerWorkerTableName, raw)
		if err != nil {
			gaia.Cfg.Logger.Error("failed to delete docker worker from memdb", "error", err.Error(), "id", workerID)
			return err
		}
	}

	// Commit transaction
	txn.Commit()

	return nil
}
