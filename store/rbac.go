package store

import (
	"encoding/json"

	bolt "github.com/coreos/bbolt"
	"github.com/gaia-pipeline/gaia"
)

// AuthPolicyAssignmentPut adds a new users policy assignments.
func (s *BoltStore) AuthPolicyAssignmentPut(assignment gaia.AuthPolicyAssignment) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(authPolicyAssignments)

		m, err := json.Marshal(assignment)
		if err != nil {
			return err
		}

		return b.Put([]byte(assignment.Username), m)
	})
}

// AuthPolicyAssignmentGet gets a users policy assignments.
func (s *BoltStore) AuthPolicyAssignmentGet(username string) (*gaia.AuthPolicyAssignment, error) {
	assignment := &gaia.AuthPolicyAssignment{}

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(authPolicyAssignments)

		g := b.Get([]byte(username))
		if g == nil {
			return nil
		}

		return json.Unmarshal(g, &assignment)
	})
	if err != nil {
		return nil, err
	}

	return assignment, nil
}

// AuthPolicyResourcePut is used to save gaia.AuthPolicyResourceV1 into the resource.authorization.policy store.
func (s *BoltStore) AuthPolicyResourcePut(spec gaia.AuthPolicyResourceV1) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(authPolicyResources)

		bts, err := s.rbacMarshaller.Marshal(spec)
		if err != nil {
			return err
		}

		return b.Put([]byte(spec.Name), bts)
	})
}

// AuthPolicyResourceGet is used get a gaia.AuthPolicyResourceV1 from the authorization.policy store.
func (s *BoltStore) AuthPolicyResourceGet(name string) (gaia.AuthPolicyResourceV1, error) {
	var spec gaia.AuthPolicyResourceV1

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(authPolicyResources)

		bts := b.Get([]byte(name))
		if bts == nil {
			return nil
		}

		return s.rbacMarshaller.Unmarshal(bts, &spec)
	})
	if err != nil {
		return gaia.AuthPolicyResourceV1{}, err
	}

	return spec, nil
}

// AuthPolicyResourceGetAll gets all auth policies resources.
func (s *BoltStore) AuthPolicyResourceGetAll() ([]gaia.AuthPolicyResourceV1, error) {
	var policies []gaia.AuthPolicyResourceV1

	return policies, s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(workerBucket)

		return b.ForEach(func(k, v []byte) error {
			p := gaia.AuthPolicyResourceV1{}

			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}

			policies = append(policies, p)

			return nil
		})
	})
}
