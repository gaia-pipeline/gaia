package rbac

import (
	"fmt"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
)

// Service defines the interface for the RBAC service.
type Service interface {
	GetPolicy(policy string) (gaia.RBACPolicyResourceV1, error)
	PutPolicy(policy gaia.RBACPolicyResourceV1) error
	GetUserEvaluatedPolicies(username string) (gaia.RBACEvaluatedPermissions, bool)
	PutUserEvaluatedPolicies(username string, perms gaia.RBACEvaluatedPermissions) error
	PutUserBinding(username string, policy string) error
}

type service struct {
	store               store.RBACStore
	evaluatedPermsCache Cache
}

// NewService creates a new RBAC Service.
func NewService(store store.RBACStore, cache Cache) Service {
	return &service{store: store, evaluatedPermsCache: cache}
}

// GetPolicy gets a policy resource from the store.
func (s *service) GetPolicy(policy string) (gaia.RBACPolicyResourceV1, error) {
	p, err := s.store.RBACPolicyResourceGet(policy)
	if err != nil {
		return gaia.RBACPolicyResourceV1{}, fmt.Errorf("failed to get policy: %v", err.Error())
	}
	return p, nil
}

// PutPolicy creates or updates a policy resource in the store. If successful, we invalidate/clear the evaluated perms
// for all users since there may have been a change in one of their policies. Could be made more efficient in the
// future.
func (s *service) PutPolicy(policy gaia.RBACPolicyResourceV1) error {
	if err := s.store.RBACPolicyResourcePut(policy); err != nil {
		return fmt.Errorf("failed to put policy: %v", err.Error())
	}
	s.evaluatedPermsCache.Clear()
	return nil
}

// GetUserEvaluatedPolicies gets a users evaluated permissions from the cache.
func (s *service) GetUserEvaluatedPolicies(username string) (gaia.RBACEvaluatedPermissions, bool) {
	if policy, exists := s.evaluatedPermsCache.Get(username); exists {
		return policy, true
	}
	return nil, false
}

// PutUserEvaluatedPolicies creates or updates the users evaluated permissions within the cache.
func (s *service) PutUserEvaluatedPolicies(username string, perms gaia.RBACEvaluatedPermissions) error {
	s.evaluatedPermsCache.Put(username, perms)
	return nil
}

// PutUserBinding saves a user policy binding into the store.
func (s *service) PutUserBinding(username string, policy string) error {
	return s.store.RBACPolicyBindingPut(username, policy)
}
