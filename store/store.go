package store

import (
	"encoding/binary"
	"fmt"
	"path/filepath"

	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
)

var (
	// Name of the bucket where we store user objects
	userBucket = []byte("Users")

	// Name of the bucket where we store information about pipelines
	pipelineBucket = []byte("Pipelines")

	// Name of the bucket where we store information about pipelines
	// which are not yet compiled (create pipeline)
	createPipelineBucket = []byte("CreatePipelines")

	// Name of the bucket where we store all pipeline runs.
	pipelineRunBucket = []byte("PipelineRun")
)

const (
	// Username and password of the first admin user
	adminUsername = "admin"
	adminPassword = "admin"

	// Bolt database file name
	boltDBFileName = "gaia.db"
)

// Store represents the access type for store
type Store struct {
	db *bolt.DB
}

// NewStore creates a new instance of Store.
func NewStore() *Store {
	s := &Store{}

	return s
}

// Init creates the data folder if not exists,
// generates private key and bolt database.
// This should be called only once per database
// because bolt holds a lock on the database file.
func (s *Store) Init() error {
	// Open connection to bolt database
	path := filepath.Join(gaia.Cfg.DataPath, boltDBFileName)
	db, err := bolt.Open(path, gaia.Cfg.Bolt.Mode, nil)
	if err != nil {
		return err
	}
	s.db = db

	// Setup database
	return s.setupDatabase()
}

// setupDatabase create all buckets in the db.
// Additionally, it makes sure that the admin user exists.
func (s *Store) setupDatabase() error {
	// Create bucket if not exists function
	var bucketName []byte
	c := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	}

	// Make sure buckets exist
	bucketName = userBucket
	err := s.db.Update(c)
	if err != nil {
		return err
	}
	bucketName = pipelineBucket
	err = s.db.Update(c)
	if err != nil {
		return err
	}
	bucketName = createPipelineBucket
	err = s.db.Update(c)
	if err != nil {
		return err
	}
	bucketName = pipelineRunBucket
	err = s.db.Update(c)
	if err != nil {
		return err
	}

	// Make sure that the user "admin" does exist
	admin, err := s.UserGet(adminUsername)
	if err != nil {
		return err
	}

	// Create admin user if we cannot find it
	if admin == nil {
		err = s.UserPut(&gaia.User{
			DisplayName: adminUsername,
			Username:    adminUsername,
			Password:    adminPassword,
		}, true)

		if err != nil {
			return err
		}
	}

	return nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
