package rbac

import (
	"fmt"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
)

// Service is the contract for the RBAC service.
type Service interface {
	GetPolicy(name string) (gaia.AuthPolicyResourceV1, error)
	PutPolicy(policy gaia.AuthPolicyResourceV1) error
}

type service struct {
	Store store.GaiaStore
	Cache Cache
}

// NewService creates a new RBAC Service.
func NewService(store store.GaiaStore, cache Cache) Service {
	return &service{Store: store, Cache: cache}
}

// GetPolicy gets a policy from the cache which, if not present, will goto the DB to retrieve the policy before
// adding it into the cache.
func (s *service) GetPolicy(name string) (gaia.AuthPolicyResourceV1, error) {
	if policy, exists := s.Cache.Get(name); exists {
		return policy, nil
	}

	policy, err := s.Store.AuthPolicyResourceGet(name)
	if err != nil {
		return gaia.AuthPolicyResourceV1{}, fmt.Errorf("failed to get policy resource from store: %w", err)
	}

	return policy, nil
}

// PutPolicy will put a policy into the db. If successful, it will look into the cache to see if an items exists that
// should be refreshed.
func (s *service) PutPolicy(policy gaia.AuthPolicyResourceV1) error {
	err := s.Store.AuthPolicyResourcePut(policy)
	if err != nil {
		return fmt.Errorf("failed to put policy resource: %v", err.Error())
	}

	// Refresh the cache item.
	_, exists := s.Cache.Get(policy.Name)
	if exists {
		s.Cache.Put(policy.Name, policy)
	}

	return nil
}
