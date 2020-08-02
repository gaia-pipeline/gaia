package rbac

import (
	"errors"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"gopkg.in/yaml.v2"

	"github.com/gaia-pipeline/gaia/helper/assethelper"
)

// rolePrefix is the prefix we give to the policy lines in the Casbin model for a role.
//  Roles are saved following this structure:
//  p, role:myrole, *, get-thing, *, allow
//  But individual user policies could look like:
//  p, myuser, *, get-thing, *, allow
const rolePrefix = "role:"

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

	// APILookup is a map that can be used for quick lookup of the API endpoints that a secured using RBAC.
	APILookup         map[string]apiLookupEndpoint
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
		enforcer      casbin.IEnforcer
		rbacAPILookup APILookup
	}
)

// NewEnforcerSvc creates a new EnforcerService.
func NewEnforcerSvc(enforcer casbin.IEnforcer, rbacAPILookup APILookup) Service {
	return &enforcerService{
		enforcer:      enforcer,
		rbacAPILookup: rbacAPILookup,
	}
}

// LoadModel loads the rbac model string from the assethelper and parses it into a Casbin model.Model.
func LoadModel() (model.Model, error) {
	modelStr, err := assethelper.LoadRBACModel()
	if err != nil {
		return nil, fmt.Errorf("error loading rbac model from assethelper: %w", err)
	}

	m, err := model.NewModelFromString(modelStr)
	if err != nil {
		return nil, fmt.Errorf("error creating model from string: %w", err)
	}

	return m, nil
}

// LoadAPILookup loads our yaml based RBACApiMappings and transforms them into a quicker lookup map.
func LoadAPILookup() (APILookup, error) {
	mappings, err := assethelper.LoadRBACAPIMappings()
	if err != nil {
		return nil, fmt.Errorf("error loading loading api mapping from assethelper: %w", err)
	}

	var apiMappings map[string]apiMapping
	if err := yaml.UnmarshalStrict([]byte(mappings), &apiMappings); err != nil {
		return nil, fmt.Errorf("error unmarshalling api mappings yaml: %w", err)
	}

	endpoints := APILookup{}
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

// AddRole adds a role into the RBAC model with the name role:myrole'.
func (e *enforcerService) AddRole(role string, roleRules []RoleRule) error {
	if !strings.HasPrefix(role, rolePrefix) {
		return fmt.Errorf("role must be prefixed with '%s'", rolePrefix)
	}

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

// GetAllRoles gets all roles. Here we actually call e.enforcer.GetAllSubjects() as roles are defined as subjects of
// the RBAC model. The e.enforcer.GetAllRoles() only gets roles that have actually been assigned to a user.
func (e *enforcerService) GetAllRoles() []string {
	roles := []string{}
	for _, sub := range e.enforcer.GetAllSubjects() {
		if strings.HasPrefix(sub, rolePrefix) {
			roles = append(roles, sub)
		}
	}
	return roles
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
	hasRole, err := e.enforcer.AddRoleForUser(username, role)
	if err != nil {
		return fmt.Errorf("error attatching role to user: %w", err)
	}
	if hasRole {
		return errors.New("user already has the role attached")
	}
	return nil
}

// DetachRole detaches a role from a user.
func (e *enforcerService) DetachRole(username string, role string) error {
	hasRole, err := e.enforcer.DeleteRoleForUser(username, role)
	if err != nil {
		return fmt.Errorf("error detatching role from user: %w", err)
	}
	if !hasRole {
		return errors.New("role not attached to user")
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
