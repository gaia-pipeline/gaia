package rbac

import (
	"errors"

	"github.com/casbin/casbin/v2"

	"github.com/gaia-pipeline/gaia"
)

type (
	RoleRule struct {
		Namespace string
		Action    string
		Resource  string
		Effect    string
	}

	EnforcerService interface {
		AddRole(role string, roleRules []RoleRule) error
		DeleteRole(role string) error

		GetAllRoles() ([]string)
		GetUserAttachedRoles(username string) ([]string, error)
		GetRoleAttachedUsers(role string) ([]string, error)
		AttachRole(username string, role string) error
		DetachRole(username string, role string) error
	}

	enforcerService struct {
		enforcer        casbin.IEnforcer
		rbacapiMappings gaia.RBACAPIMappings
	}
)

// NewEnforcerSvc creates a new enforcerService.
func NewEnforcerSvc(enforcer casbin.IEnforcer, apiMappingsFile string) (*enforcerService, error) {
	rbacapiMappings, err := loadAPIMappings(apiMappingsFile)
	if err != nil {
		return nil, err
	}

	return &enforcerService{
		enforcer:        enforcer,
		rbacapiMappings: rbacapiMappings,
	}, nil
}

func (e *enforcerService) DeleteRole(role string) error {
	exists, err := e.enforcer.DeleteRole(role)
	if !exists {
		return errors.New("role does not exist")
	}
	return err
}

func (e *enforcerService) AddRole(role string, roleRules []RoleRule) error {
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

func (e *enforcerService) GetAllRoles() ([]string) {
	return e.enforcer.GetAllRoles()
}

func (e *enforcerService) GetUserAttachedRoles(username string) ([]string, error) {
	return e.enforcer.GetRolesForUser(username)
}

func (e *enforcerService) GetRoleAttachedUsers(username string) ([]string, error) {
	return e.enforcer.GetUsersForRole(username)
}

func (e *enforcerService) AttachRole(username string, role string) error {
	if _, err := e.enforcer.AddRoleForUser(username, role); err != nil {
		return err
	}
	return nil
}

func (e *enforcerService) DetachRole(username string, role string) error {
	exists, err := e.enforcer.DeleteRoleForUser(username, role)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("role does not exists for user")
	}
	return nil
}
