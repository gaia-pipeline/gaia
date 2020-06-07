package rbac

import (
	"errors"

	"github.com/casbin/casbin/v2"

	"github.com/gaia-pipeline/gaia"
)

type (
	// RoleRule represents a Casbins role rule line in the format we expect.
	RoleRule struct {
		Namespace string
		Action    string
		Resource  string
		Effect    string
	}

	// Service wraps the Casbin enforcer and performs all actions we require to manage and use RBAC functions.
	Service interface {
		AddRole(role string, roleRules []RoleRule) error
		DeleteRole(role string) error
		GetAllRoles() []string
		GetUserAttachedRoles(username string) ([]string, error)
		GetRoleAttachedUsers(role string) ([]string, error)
		AttachRole(username string, role string) error
		DetachRole(username string, role string) error
	}

	EnforcerService struct {
		enforcer        casbin.IEnforcer
		rbacapiMappings gaia.RBACAPIMappings
	}
)

// NewEnforcerSvc creates a new EnforcerService.
func NewEnforcerSvc(enforcer casbin.IEnforcer, apiMappingsFile string) (*EnforcerService, error) {
	rbacapiMappings, err := loadAPIMappings(apiMappingsFile)
	if err != nil {
		return nil, err
	}

	return &EnforcerService{
		enforcer:        enforcer,
		rbacapiMappings: rbacapiMappings,
	}, nil
}

// DeleteRole deletes a role.
func (e *EnforcerService) DeleteRole(role string) error {
	exists, err := e.enforcer.DeleteRole(role)
	if !exists {
		return errors.New("role does not exist")
	}
	return err
}

// AddRole adds a role.
func (e *EnforcerService) AddRole(role string, roleRules []RoleRule) error {
	rules := [][]string{}
	for _, p := range roleRules {
		r := []string{role, p.Namespace, p.Action, p.Resource, p.Effect}
		rules = append(rules, r)
	}

	ok, err := e.enforcer.AddPolicies(rules)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("rule already exists for role")
	}

	return nil
}

// GetAllRoles gets all roles.
func (e *EnforcerService) GetAllRoles() []string {
	return e.enforcer.GetAllRoles()
}

// GetUserAttachedRoles gets all roles attached to a specific user.
func (e *EnforcerService) GetUserAttachedRoles(username string) ([]string, error) {
	return e.enforcer.GetRolesForUser(username)
}

// GetRoleAttachedUsers get all users attached to a specific role.
func (e *EnforcerService) GetRoleAttachedUsers(role string) ([]string, error) {
	return e.enforcer.GetUsersForRole(role)
}

// AttachRole attaches a role to a user.
func (e *EnforcerService) AttachRole(username string, role string) error {
	if _, err := e.enforcer.AddRoleForUser(username, role); err != nil {
		return err
	}
	return nil
}

// DetachRole detaches a role from a user.
func (e *EnforcerService) DetachRole(username string, role string) error {
	exists, err := e.enforcer.DeleteRoleForUser(username, role)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("role does not exists for user")
	}
	return nil
}
