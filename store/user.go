package store

import (
	"encoding/json"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
	"golang.org/x/crypto/bcrypt"
)

// UserPut takes the given user and saves it
// to the bolt database. User will be overwritten
// if it already exists.
// It also clears the password field afterwards.
func (s *BoltStore) UserPut(u *gaia.User, encryptPassword bool) error {
	// Encrypt password before we save it
	if encryptPassword {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)
		if err != nil {
			return err
		}
		u.Password = string(hash)
	}

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
func (s *BoltStore) UserAuth(u *gaia.User, updateLastLogin bool) (*gaia.User, error) {
	// Look up user
	user, err := s.UserGet(u.Username)

	// Error occured and/or user not found
	if err != nil || user == nil {
		return nil, err
	}

	// Check if password is valid
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)); err != nil {
		return nil, err
	}

	// Update last login field
	if updateLastLogin {
		user.LastLogin = time.Now()
		err = s.UserPut(user, false)
		if err != nil {
			return nil, err
		}
	}

	// We will use the user object later.
	// But we don't need the password anymore.
	user.Password = ""

	// Return user
	return user, nil
}

// UserGet looks up a user by given username.
// Returns nil if user was not found.
func (s *BoltStore) UserGet(username string) (*gaia.User, error) {
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

// UserGetAll returns all stored users.
func (s *BoltStore) UserGetAll() ([]gaia.User, error) {
	var users []gaia.User

	return users, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(userBucket)

		// Iterate all users and add them to slice
		return b.ForEach(func(k, v []byte) error {
			// create single user object
			u := &gaia.User{}

			// Unmarshal
			err := json.Unmarshal(v, u)
			if err != nil {
				return err
			}

			// Remove password for security reasons
			u.Password = ""

			users = append(users, *u)
			return nil
		})
	})
}

// UserDelete deletes the given user.
func (s *BoltStore) UserDelete(u string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(userBucket)

		// Delete user
		return b.Delete([]byte(u))
	})
}
