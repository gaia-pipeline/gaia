package rbac

import (
	"fmt"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/cachehelper"
	"github.com/gaia-pipeline/gaia/store"
)

// Service is the contract for the RBAC service.
type Service interface {
	GetPolicy(name string) (gaia.AuthPolicyResourceV1, error)
}

type service struct {
	Store store.GaiaStore
	Cache cachehelper.Cache
}

// NewService creates a new RBAC Service.
func NewService(store store.GaiaStore, cache cachehelper.Cache) Service {
	return &service{Store: store, Cache: cache}
}

// GetPolicy attempts to get a policy from the cache. If not present it will goto the DB to retrieve the policy before
// adding it into the cache.
func (s *service) GetPolicy(name string) (gaia.AuthPolicyResourceV1, error) {
	if policy, exists := s.Cache.Get(name); exists {
		return policy.(gaia.AuthPolicyResourceV1), nil
	}

	policy, err := s.Store.AuthPolicyResourceGet(name)
	if err != nil {
		return gaia.AuthPolicyResourceV1{}, fmt.Errorf("failed to get policy resource from store: %w", err)
	}

	item := s.Cache.Put(name, policy).(gaia.AuthPolicyResourceV1)
	return item, nil
}
