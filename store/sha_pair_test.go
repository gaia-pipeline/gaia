package store

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gaia-pipeline/gaia"
)

func TestGetSHAPair(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestGetSHAPAir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	pair := gaia.SHAPair{}
	pair.UniqueID = 1
	pair.Original = []byte("original")
	pair.Worker = []byte("worker")
	err = store.UpsertSHAPair(pair)
	if err != nil {
		t.Fatal(err)
	}

	ok, p, err := store.GetSHAPair(1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("sha pair not found")
	}

	if p.UniqueID != pair.UniqueID {
		t.Fatalf("unique id match error. want %d got %d", pair.UniqueID, p.UniqueID)
	}
	if !bytes.Equal(p.Worker, pair.Worker) {
		t.Fatalf("worker sha match error. want %s got %s", pair.Worker, p.Worker)
	}
	if !bytes.Equal(p.Original, pair.Original) {
		t.Fatalf("original sha match error. want %s got %s", pair.Original, p.Original)
	}
}

func TestUpsertSHAPair(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestUpsertSHAPair")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	pair := gaia.SHAPair{}
	pair.UniqueID = 1
	pair.Original = []byte("original")
	pair.Worker = []byte("worker")
	err = store.UpsertSHAPair(pair)
	if err != nil {
		t.Fatal(err)
	}
	// Test is upsert overwrites existing records.
	pair.Original = []byte("original2")
	err = store.UpsertSHAPair(pair)
	if err != nil {
		t.Fatal(err)
	}

	ok, p, err := store.GetSHAPair(1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("sha pair not found")
	}

	if p.UniqueID != pair.UniqueID {
		t.Fatalf("unique id match error. want %d got %d", pair.UniqueID, p.UniqueID)
	}
	if !bytes.Equal(p.Worker, pair.Worker) {
		t.Fatalf("worker sha match error. want %s got %s", pair.Worker, p.Worker)
	}
	if !bytes.Equal(p.Original, pair.Original) {
		t.Fatalf("original sha match error. want %s got %s", pair.Original, p.Original)
	}
}
