package store

import (
	"encoding/json"
	"fmt"
	"os"

	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
	"golang.org/x/crypto/bcrypt"
)

var (
	// Name of the bucket where we store user objects
	userBucket = []byte("Users")

	// Name of the bucket where we store information about pipelines
	pipelineBucket = []byte("Pipelines")

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

// UserPut takes the given user and saves it
// to the bolt database. User will be overwritten
// if it already exists.
// It also clears the password field afterwards.
func (s *Store) UserPut(u *gaia.User) error {
	// Encrypt password before we save it
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)

	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(userBucket)

		// Marshal user object
		m, err := json.Marshal(u)
		if err != nil {
			return err
		}

		// Clear password from origin object
		u.Password = ""

		// Put user
		return b.Put([]byte(u.Username), m)
	})
}

// UserAuth looks up a user by given username.
// Then it compares passwords and returns user obj if
// given password is valid. Returns nil if password was
// wrong or user not found.
func (s *Store) UserAuth(u *gaia.User) (*gaia.User, error) {
	// Look up user
	user, err := s.UserGet(u.Username)

	// Error occured and/or user not found
	if err != nil || user == nil {
		return nil, err
	}

	// Check if password is valid
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)); err != nil {
		return nil, nil
	}

	// We will use the user object later.
	// But we don't need the password anymore.
	user.Password = ""

	// Return user
	return user, nil
}

// UserGet looks up a user by given username.
// Returns nil if user was not found.
func (s *Store) UserGet(username string) (*gaia.User, error) {
	user := &gaia.User{}
	err := s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(userBucket)

		// Lookup user
		userRaw := b.Get([]byte(username))

		// User found?
		if userRaw == nil {
			// Nope. That is not an error so just leave
			user = nil
			return nil
		}

		// Unmarshal
		return json.Unmarshal(userRaw, user)
	})

	return user, err
}

// Init creates the data folder if not exists,
// generates private key and bolt database.
// This should be called only once per database
// because bolt holds a lock on the database file.
func (s *Store) Init(cfg *gaia.Config) error {
	// Make sure data folder exists
	err := os.MkdirAll(cfg.DataPath, 0700)
	if err != nil {
		return err
	}

	// Open connection to bolt database
	path := cfg.DataPath + string(os.PathSeparator) + cfg.Bolt.Path
	db, err := bolt.Open(path, cfg.Bolt.Mode, nil)
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
