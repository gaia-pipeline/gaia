package store

import (
	"encoding/json"
	"github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
)

//func (s *BoltStore) PermissionGroupHasUser(name, username string) (bool, error) {
//	hasUser := false
//	var pg *gaia.PermissionGroup
//
//	err := s.db.View(func(tx *bolt.Tx) error {
//		b := tx.Bucket(permGroupsBucket)
//
//		g := b.Get([]byte(name))
//		if g == nil {
//			return nil
//		}
//
//		return json.Unmarshal(g, pg)
//	})
//	if err != nil {
//		return false, err
//	}
//
//	for _, u := range pg.Users {
//		if username == u {
//			hasUser = true
//		}
//	}
//
//	return hasUser, err
//}

func (s *BoltStore) PermissionGroupGet(name string) (*gaia.PermissionGroup, error) {
	var pg *gaia.PermissionGroup

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(permGroupsBucket)

		g := b.Get([]byte(name))
		if g == nil {
			return nil
		}

		return json.Unmarshal(g, pg)
	})
	if err != nil {
		return nil, err
	}

	return pg, nil
}

func (s *BoltStore) PermissionGroupGetAll() ([]*gaia.PermissionGroup, error) {
	var pgs []*gaia.PermissionGroup

	return pgs, s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(permGroupsBucket)

		return b.ForEach(func(k, v []byte) error {
			var g *gaia.PermissionGroup

			err := json.Unmarshal(v, &g)
			if err != nil {
				return err
			}

			pgs = append(pgs, g)
			return nil
		})
	})
}

func (s *BoltStore) PermissionGroupPut(group *gaia.PermissionGroup) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(permGroupsBucket)

		m, err := json.Marshal(group)
		if err != nil {
			return err
		}

		return b.Put([]byte(group.Name), m)
	})
}

func (s *BoltStore) PermissionGroupCreate(group *gaia.PermissionGroup) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(permGroupsBucket)

		m, err := json.Marshal(group)
		if err != nil {
			return err
		}

		return b.Put([]byte(group.Name), m)
	})
}

func (s *BoltStore) PermissionGroupDelete(name string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(permGroupsBucket)
		return b.Delete([]byte(name))
	})
}
