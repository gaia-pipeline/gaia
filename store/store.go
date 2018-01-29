package store

import (
	"fmt"
	"os"

	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
)

const (
	dataFolder = "data"
)

var (
	// Name of the bucket where we store user objects
	userBucket = []byte("Users")

	// Name of the bucket where we store information about pipelines
	pipelineBucket = []byte("Pipelines")

	// Name of the bucket where we store information about pipelines
	// which are not yet compiled (create pipeline)
	createPipelineBucket = []byte("CreatePipelines")

	// Username and password of the first admin user
	adminUsername = "admin"
	adminPassword = "admin"
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
	// Make sure data folder exists
	folder := gaia.Cfg.HomePath + string(os.PathSeparator) + dataFolder
	err := os.MkdirAll(folder, 0700)
	if err != nil {
		return err
	}

	// Open connection to bolt database
	path := folder + string(os.PathSeparator) + gaia.Cfg.Bolt.Path
	db, err := bolt.Open(path, gaia.Cfg.Bolt.Mode, nil)
	if err != nil {
		return err
	}
	s.db = db

	// Setup database
	return setupDatabase(s)
}

// setupDatabase create all buckets in the db.
// Additionally, it makes sure that the admin user exists.
func setupDatabase(s *Store) error {
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
		})

		if err != nil {
			return err
		}
	}

	return nil
}
