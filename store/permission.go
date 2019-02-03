package store

import (
	"encoding/json"

	"github.com/coreos/bbolt"
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

// UserPermissionGroupPut adds or updates a permission group.
func (s *BoltStore) UserPermissionGroupPut(group *gaia.UserPermissionGroup) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(permissionGroupBucket)
		m, err := json.Marshal(group)
		if err != nil {
			return err
		}
		return b.Put([]byte(group.Name), m)
	})
}

// UserPermissionGroupGet gets a permission group with the specified name.
func (s *BoltStore) UserPermissionGroupGet(name string) (*gaia.UserPermissionGroup, error) {
	var perms *gaia.UserPermissionGroup
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(permissionGroupBucket)

		g := b.Get([]byte(name))
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

// UserPermissionGroupGetAll returns all permission groups.
func (s *BoltStore) UserPermissionGroupGetAll() ([]*gaia.UserPermissionGroup, error) {
	var groups []*gaia.UserPermissionGroup
	return groups, s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(permissionGroupBucket)
		return b.ForEach(func(k, v []byte) error {
			var group *gaia.UserPermissionGroup
			err := json.Unmarshal(v, &group)
			if err != nil {
				return err
			}
			groups = append(groups, group)
			return nil
		})
	})
}

// UserPermissionGroupDelete deletes a permission group with the specified name.
func (s *BoltStore) UserPermissionGroupDelete(name string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(permissionGroupBucket)
		return b.Delete([]byte(name))
	})
}
