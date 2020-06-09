package rbac

import (
	"errors"
	"log"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gopkg.in/yaml.v2"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/assethelper"
)

type (
	// RoleRule represents a Casbin role rule line in the format we expect.
	RoleRule struct {
		Namespace string `json:"namespace"`
		Action    string `json:"action"`
		Resource  string `json:"resource"`
		Effect    string `json:"effect"`
	}

	// Service wraps the Casbin enforcer and performs all actions we require to manage and use RBAC functions.
	Service interface {
		EndpointEnforcer
		AddRole(role string, roleRules []RoleRule) error
		DeleteRole(role string) error
		GetAllRoles() []string
		GetUserAttachedRoles(username string) ([]string, error)
		GetRoleAttachedUsers(role string) ([]string, error)
		AttachRole(username string, role string) error
		DetachRole(username string, role string) error
	}

	enforcerService struct {
		adapter         persist.BatchAdapter
		enforcer        casbin.IEnforcer
		rbacapiMappings gaia.RBACAPIMappings
	}
)

// NewEnforcerSvc creates a new EnforcerService.
func NewEnforcerSvc(adapter persist.BatchAdapter) (Service, error) {
	model, err := loadModel()
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(model, adapter)
	if err != nil {
		log.Fatal(err)
	}
	enforcer.EnableLog(true)

	rbacapiMappings, err := loadAPIMappings()
	if err != nil {
		return nil, err
	}

	return &enforcerService{
		enforcer:        enforcer,
		rbacapiMappings: rbacapiMappings,
	}, nil
}

func loadModel() (model.Model, error) {
	modelStr, err := assethelper.LoadRBACModel()
	if err != nil {
		return nil, err
	}

	model, err := model.NewModelFromString(modelStr)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func loadAPIMappings() (gaia.RBACAPIMappings, error) {
	mappings, err := assethelper.LoadRBACAPIMappings()
	if err != nil {
		return gaia.RBACAPIMappings{}, err
	}

	var rbacapiMappings gaia.RBACAPIMappings
	if err := yaml.Unmarshal([]byte(mappings), &rbacapiMappings); err != nil {
		return gaia.RBACAPIMappings{}, err
	}

	return rbacapiMappings, nil
}

// DeleteRole deletes a role.
func (e *enforcerService) DeleteRole(role string) error {
	exists, err := e.enforcer.DeleteRole(role)
	if !exists {
		return errors.New("role does not exist")
	}
	return err
}

// AddRole adds a role.
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

// GetAllRoles gets all roles.
func (e *enforcerService) GetAllRoles() []string {
	return e.enforcer.GetAllRoles()
}

// GetUserAttachedRoles gets all roles attached to a specific user.
func (e *enforcerService) GetUserAttachedRoles(username string) ([]string, error) {
	return e.enforcer.GetRolesForUser(username)
}

// GetRoleAttachedUsers get all users attached to a specific role.
func (e *enforcerService) GetRoleAttachedUsers(role string) ([]string, error) {
	return e.enforcer.GetUsersForRole(role)
}

// AttachRole attaches a role to a user.
func (e *enforcerService) AttachRole(username string, role string) error {
	if _, err := e.enforcer.AddRoleForUser(username, role); err != nil {
		return err
	}
	return nil
}

// DetachRole detaches a role from a user.
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
