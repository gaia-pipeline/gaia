package store

import (
	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
)

// ResourceAuthRBACPut is used to save gaia.RBACPolicyV1 into the resource.authorization.rbac store.
func (s *BoltStore) ResourceAuthRBACPut(spec gaia.RBACPolicyV1) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(resourceAuthRBACBucket)

		bts, err := s.rbacMarshaller.Marshal(spec)
		if err != nil {
			return err
		}

		return b.Put([]byte(spec.Name), bts)
	})
}

// ResourceAuthRBACGet is used get a gaia.RBACPolicyV1 from the resource.authorization.rbac store.
func (s *BoltStore) ResourceAuthRBACGet(name string) (gaia.RBACPolicyV1, error) {
	var spec gaia.RBACPolicyV1

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(resourceAuthRBACBucket)

		bts := b.Get([]byte(name))
		if bts == nil {
			return nil
		}

		return s.rbacMarshaller.Unmarshal(bts, &spec)
	})
	if err != nil {
		return gaia.RBACPolicyV1{}, err
	}

	return spec, nil
}
