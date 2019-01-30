package store

import (
	"encoding/json"

	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
)

// Returns the permission data for a given username
func (s *BoltStore) UserPermissionsGet(username string) (*gaia.UserPermission, error) {
	var perms *gaia.UserPermission

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(userPermsBucket)

		g := b.Get([]byte(username))
		if g == nil {
			return nil
		}

		return json.Unmarshal(g, &perms)
	})
	if err != nil {
		return nil, err
	}

	return perms, nil
}

// Creates or updates user permissions
func (s *BoltStore) UserPermissionsPut(perms *gaia.UserPermission) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(userPermsBucket)

		m, err := json.Marshal(perms)
		if err != nil {
			return err
		}

		return b.Put([]byte(perms.Username), m)
	})
}

// Deletes permission data for a given username
func (s *BoltStore) UserPermissionsDelete(username string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(userPermsBucket)
		return b.Delete([]byte(username))
	})
}
