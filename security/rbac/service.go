package rbac

import (
	"fmt"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/cachehelper"
	"github.com/gaia-pipeline/gaia/store"
)

type Service interface {
	GetPolicy(name string) (gaia.AuthPolicyResourceV1, error)
}

type service struct {
	Store store.GaiaStore
	Cache *cachehelper.Cache
}

func NewService(store store.GaiaStore, cache *cachehelper.Cache) *service {
	return &service{Store: store, Cache: cache}
}

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
