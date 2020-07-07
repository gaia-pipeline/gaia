package rbac

import (
	"errors"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gopkg.in/yaml.v2"

	"github.com/gaia-pipeline/gaia/helper/assethelper"
)

type (
	apiMapping struct {
		Description string        `json:"description"`
		Endpoints   []apiEndpoint `json:"endpoints"`
	}
	apiEndpoint struct {
		Method   string `json:"method"`
		Path     string `json:"path"`
		Resource string `json:"resource"`
	}

	apiLookup         map[string]apiLookupEndpoint
	apiLookupEndpoint struct {
		Param   string            `yaml:"param"`
		Methods map[string]string `yaml:"methods"`
	}

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
		DeleteUser(username string) error
	}

	enforcerService struct {
		adapter       persist.BatchAdapter
		enforcer      casbin.IEnforcer
		rbacapiLookup apiLookup
	}
)

// NewEnforcerSvc creates a new EnforcerService.
func NewEnforcerSvc(debug bool, adapter persist.BatchAdapter) (Service, error) {
	model, err := loadModel()
	if err != nil {
		return nil, fmt.Errorf("error loading model: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(model, adapter)
	if err != nil {
		return nil, fmt.Errorf("error instantiating casbin enforcer: %w", err)
	}

	if debug {
		enforcer.EnableLog(true)
	}

	rbacapiMappings, err := loadAPIMappings()
	if err != nil {
		return nil, fmt.Errorf("error loading rbac api mappings: %w", err)
	}

	return &enforcerService{
		enforcer:      enforcer,
		rbacapiLookup: rbacapiMappings,
	}, nil
}

func loadModel() (model.Model, error) {
	modelStr, err := assethelper.LoadRBACModel()
	if err != nil {
		return nil, fmt.Errorf("error loading rbac model from assethelper: %w", err)
	}

	model, err := model.NewModelFromString(modelStr)
	if err != nil {
		return nil, fmt.Errorf("error creating model from string: %w", err)
	}

	return model, nil
}

func loadAPIMappings() (apiLookup, error) {
	mappings, err := assethelper.LoadRBACAPIMappings()
	if err != nil {
		return nil, fmt.Errorf("error loading loading api mapping from assethelper: %w", err)
	}

	var apiMappings map[string]apiMapping
	if err := yaml.UnmarshalStrict([]byte(mappings), &apiMappings); err != nil {
		return nil, fmt.Errorf("error unmarshalling api mappings yaml: %w", err)
	}

	endpoints := apiLookup{}
	for mappingPath, mapping := range apiMappings {
		for _, e := range mapping.Endpoints {
			path, hasPath := endpoints[e.Path]
			if !hasPath {
				endpoints[e.Path] = apiLookupEndpoint{
					Methods: map[string]string{e.Method: mappingPath},
					Param:   e.Resource,
				}
				continue
			}
			path.Methods[e.Method] = mappingPath
		}
	}

	return endpoints, nil
}

// DeleteRole deletes a role.
func (e *enforcerService) DeleteRole(role string) error {
	exist, err := e.enforcer.DeleteRole(role)
	if !exist {
		return errors.New("role does not exist")
	}
	if err != nil {
		return fmt.Errorf("error deleting role: %w", err)
	}
	return nil
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
		return fmt.Errorf("error adding policies: %w", err)
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
	roles, err := e.enforcer.GetRolesForUser(username)
	if err != nil {
		return nil, fmt.Errorf("error getting roles for user: %w", err)
	}
	return roles, nil
}

// GetRoleAttachedUsers get all users attached to a specific role.
func (e *enforcerService) GetRoleAttachedUsers(role string) ([]string, error) {
	users, err := e.enforcer.GetUsersForRole(role)
	if err != nil {
		return nil, fmt.Errorf("error getting users for role: %w", err)
	}
	return users, nil
}

// AttachRole attaches a role to a user.
func (e *enforcerService) AttachRole(username string, role string) error {
	if _, err := e.enforcer.AddRoleForUser(username, role); err != nil {
		return fmt.Errorf("error attatching role to user: %w", err)
	}
	return nil
}

// DetachRole detaches a role from a user.
func (e *enforcerService) DetachRole(username string, role string) error {
	exists, err := e.enforcer.DeleteRoleForUser(username, role)
	if err != nil {
		return fmt.Errorf("error detatching role from user: %w", err)
	}
	if !exists {
		return errors.New("role does not exists for user")
	}
	return nil
}

// DeleteUser removes the user from the rbac model.
func (e *enforcerService) DeleteUser(username string) error {
	if _, err := e.enforcer.DeleteUser(username); err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	return nil
}
