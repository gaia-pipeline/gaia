package store

import (
	"encoding/json"

	bolt "go.etcd.io/bbolt"

	"github.com/gaia-pipeline/gaia"
)

// UserPermissionsGet gets the permission data for a given username.
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

// UserPermissionsPut adds or updates user permissions.
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

// UserPermissionsDelete deletes permission data for a given username.
func (s *BoltStore) UserPermissionsDelete(username string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(userPermsBucket)
		return b.Delete([]byte(username))
	})
}
