package memdb

import (
	"testing"

	"github.com/gaia-pipeline/gaia/workers/docker"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
)

type mockStore struct {
	store.GaiaStore
}

func (m mockStore) WorkerGetAll() ([]*gaia.Worker, error) {
	return []*gaia.Worker{
		{UniqueID: "my-unique-id"},
	}, nil
}
func (m mockStore) WorkerPut(w *gaia.Worker) error { return nil }
func (m mockStore) WorkerDelete(id string) error   { return nil }

func TestInitMemDB(t *testing.T) {
	mockStore := mockStore{}
	if _, err := InitMemDB(mockStore); err != nil {
		t.Fatal(err)
	}
}

func TestSyncStore(t *testing.T) {
	mockStore := mockStore{}
	db, err := InitMemDB(mockStore)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.SyncStore(); err != nil {
		t.Fatal(err)
	}

	// Check if worker is in memdb
	if len(db.GetAllWorker()) != 1 {
		t.Fatalf("worker in db should be 1 but is %d", len(db.GetAllWorker()))
	}
	w, err := db.GetWorker("my-unique-id")
	if err != nil {
		t.Fatal(err)
	}
	if w == nil {
		t.Fatal("cannot find worker in memdb with id 'my-unique-id'")
	}
}

func TestGetAllWorker(t *testing.T) {
	mockStore := mockStore{}
	db, err := InitMemDB(mockStore)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.SyncStore(); err != nil {
		t.Fatal(err)
	}

	if err := db.UpsertWorker(&gaia.Worker{UniqueID: "other-worker"}, false); err != nil {
		t.Fatal(err)
	}

	// Check if worker is in memdb
	if len(db.GetAllWorker()) != 2 {
		t.Fatalf("worker in db should be 2 but is %d", len(db.GetAllWorker()))
	}
	w, err := db.GetWorker("my-unique-id")
	if err != nil {
		t.Fatal(err)
	}
	if w == nil {
		t.Fatal("cannot find worker in memdb with id 'my-unique-id'")
	}
}

func TestUpsertWorker(t *testing.T) {
	mockStore := mockStore{}
	db, err := InitMemDB(mockStore)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.SyncStore(); err != nil {
		t.Fatal(err)
	}

	if err := db.UpsertWorker(&gaia.Worker{UniqueID: "other-worker"}, false); err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertWorker(&gaia.Worker{UniqueID: "another-other-worker"}, true); err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertWorker(&gaia.Worker{UniqueID: "another-other-worker"}, true); err != nil {
		t.Fatal(err)
	}

	w, err := db.GetWorker("my-unique-id")
	if err != nil {
		t.Fatal(err)
	}
	if w == nil {
		t.Fatal("cannot find worker in memdb with id 'my-unique-id'")
	}
	w, err = db.GetWorker("other-worker")
	if err != nil {
		t.Fatal(err)
	}
	if w == nil {
		t.Fatal("cannot find worker in memdb with id 'other-worker'")
	}
	w, err = db.GetWorker("another-other-worker")
	if err != nil {
		t.Fatal(err)
	}
	if w == nil {
		t.Fatal("cannot find worker in memdb with id 'another-other-worker'")
	}
}

func TestDeleteWorker(t *testing.T) {
	mockStore := mockStore{}
	db, err := InitMemDB(mockStore)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.SyncStore(); err != nil {
		t.Fatal(err)
	}

	if err := db.UpsertWorker(&gaia.Worker{UniqueID: "other-worker"}, false); err != nil {
		t.Fatal(err)
	}

	if err := db.DeleteWorker("my-unique-id", false); err != nil {
		t.Fatal(err)
	}

	w, err := db.GetWorker("my-unique-id")
	if err != nil {
		t.Fatal(err)
	}
	if w != nil {
		t.Fatal("found worker in memdb with id 'my-unique-id' but should be deleted")
	}

	if err := db.DeleteWorker("other-worker", true); err != nil {
		t.Fatal(err)
	}

	w, err = db.GetWorker("other-worker")
	if err != nil {
		t.Fatal(err)
	}
	if w != nil {
		t.Fatal("found worker in memdb with id 'other-worker' but should be deleted")
	}
}

func TestPopPipelineRun(t *testing.T) {
	mockStore := mockStore{}
	db, err := InitMemDB(mockStore)
	if err != nil {
		t.Fatal(err)
	}

	// Create test data
	pR1 := &gaia.PipelineRun{
		UniqueID:     "first-pipelinerun",
		PipelineType: gaia.PTypeGolang,
		PipelineTags: []string{"first-tag", "second-tag", "third-tag"},
	}
	pR2 := &gaia.PipelineRun{
		UniqueID:     "second-pipelinerun",
		PipelineType: gaia.PTypeCpp,
		PipelineTags: []string{"other-tag", "first-tag", "blubb-tag"},
	}

	// Insert pipeline run
	if err := db.InsertPipelineRun(pR1); err != nil {
		t.Fatal(err)
	}
	if err := db.InsertPipelineRun(pR2); err != nil {
		t.Fatal(err)
	}

	// Create tags with one missing (third-tag)
	tags := []string{gaia.PTypeGolang.String(), "first-tag", "second-tag"}
	pRun, err := db.PopPipelineRun(tags)
	if err != nil {
		t.Fatal(err)
	}
	if pRun != nil {
		t.Fatalf("run should be nil but is %#v", pRun)
	}

	// Pop success
	tags = append(tags, "third-tag")
	pRun, err = db.PopPipelineRun(tags)
	if err != nil {
		t.Fatal(err)
	}
	if pRun == nil {
		t.Fatal("pipeline run is nil")
	}
	if pRun.UniqueID != "first-pipelinerun" {
		t.Fatalf("popped pipeline run should be 'first-pipelinerun' but is '%s'", pRun.UniqueID)
	}

	// Pop again
	pRun, err = db.PopPipelineRun(tags)
	if err != nil {
		t.Fatal(err)
	}
	if pRun != nil {
		t.Fatalf("run should be nil but is %#v", pRun)
	}

	// Pop success
	tags = []string{gaia.PTypeCpp.String(), "other-tag", "first-tag", "blubb-tag"}
	pRun, err = db.PopPipelineRun(tags)
	if err != nil {
		t.Fatal(err)
	}
	if pRun == nil {
		t.Fatal("pipeline run is nil")
	}
	if pRun.UniqueID != "second-pipelinerun" {
		t.Fatalf("popped pipeline run should be 'second-pipelinerun' but is '%s'", pRun.UniqueID)
	}

	// Pop again
	pRun, err = db.PopPipelineRun(tags)
	if err != nil {
		t.Fatal(err)
	}
	if pRun != nil {
		t.Fatalf("run should be nil but is %#v", pRun)
	}
}

func TestDeletePipelineRun(t *testing.T) {
	mockStore := mockStore{}
	db, err := InitMemDB(mockStore)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.SyncStore(); err != nil {
		t.Fatal(err)
	}

	if err := db.InsertPipelineRun(&gaia.PipelineRun{UniqueID: "pipelinerun"}); err != nil {
		t.Fatal(err)
	}

	if err := db.DeletePipelineRun("pipelinerun"); err != nil {
		t.Fatal(err)
	}
}

func TestDockerWorker(t *testing.T) {
	mockStore := mockStore{}
	db, err := InitMemDB(mockStore)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.SyncStore(); err != nil {
		t.Fatal(err)
	}

	if err := db.InsertDockerWorker(&docker.Worker{WorkerID: "testworker"}); err != nil {
		t.Fatal(err)
	}

	w, err := db.GetDockerWorker("testworker")
	if err != nil {
		t.Fatal(err)
	}
	if w == nil {
		t.Fatal("expected non-nil response")
	}
	workers, err := db.GetAllDockerWorker()
	if err != nil {
		t.Fatal(err)
	}
	if len(workers) != 1 {
		t.Fatalf("expected 1 but got '%d': %#v", len(workers), workers)
	}

	// Delete docker worker
	if err := db.DeleteDockerWorker("testworker"); err != nil {
		t.Fatal(err)
	}
	workers, err = db.GetAllDockerWorker()
	if err != nil {
		t.Fatal(err)
	}
	if len(workers) != 0 {
		t.Fatalf("expected zero workers returned but got '%d': %#v", len(workers), workers)
	}
}
