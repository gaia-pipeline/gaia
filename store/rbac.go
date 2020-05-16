package store

import (
	"encoding/json"
	"fmt"

	bolt "github.com/coreos/bbolt"

	"github.com/gaia-pipeline/gaia"
)

// RBACPolicyBindingsPut adds a new users policy assignments.
func (s *BoltStore) RBACPolicyBindingsPut(username string, policy string) error {
	existing, err := s.RBACPolicyBindingsGet(username)
	if err != nil {
		return fmt.Errorf("failed to get bindings: %v", err.Error())
	}
	existing[policy] = ""

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(authPolicyBindings)

		m, err := json.Marshal(existing)
		if err != nil {
			return err
		}

		return b.Put([]byte(username), m)
	})
}

// RBACPolicyBindingsGet gets a users policy assignments.
func (s *BoltStore) RBACPolicyBindingsGet(username string) (map[string]interface{}, error) {
	assignment := make(map[string]interface{})

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(authPolicyBindings)

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

// RBACPolicyResourcePut is used to save gaia.RBACPolicyResourceV1 into the resource.authorization.policy store.
func (s *BoltStore) RBACPolicyResourcePut(spec gaia.RBACPolicyResourceV1) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(authPolicyResources)

		bts, err := s.rbacMarshaller.Marshal(spec)
		if err != nil {
			return err
		}

		return b.Put([]byte(spec.Name), bts)
	})
}

// RBACPolicyResourceGet is used get a gaia.RBACPolicyResourceV1 from the authorization.policy store.
func (s *BoltStore) RBACPolicyResourceGet(name string) (gaia.RBACPolicyResourceV1, error) {
	var spec gaia.RBACPolicyResourceV1

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(authPolicyResources)

		bts := b.Get([]byte(name))
		if bts == nil {
			return nil
		}

		return s.rbacMarshaller.Unmarshal(bts, &spec)
	})
	if err != nil {
		return gaia.RBACPolicyResourceV1{}, err
	}

	return spec, nil
}
