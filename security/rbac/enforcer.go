package rbac

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"

	"github.com/gaia-pipeline/gaia"
)

var (
	errNamespaceNotFound = errors.New("namespace not found")
	errActionNotFound    = errors.New("action not found")
	errResourceNotFound  = errors.New("resource not found")
)

// EnforcerConfig represents the config required for RBAC.
type EnforcerConfig struct {
	User      User
	Namespace gaia.RBACPolicyNamespace
	Action    gaia.RBACPolicyAction
	Resource  gaia.RBACPolicyResource
}

// User represents the user to apply the enforcement to.
type User struct {
	Username string
	Policies map[string]interface{}
}

// PolicyEnforcer is for enforcing RBAC Policies.
type PolicyEnforcer interface {
	Enforce(cfg EnforcerConfig) error
	Evaluate(user User) (gaia.RBACEvaluatedPermissions, error)
	GetDefaultAPIGroup() gaia.RBACAPIGroup
}

type policyEnforcer struct {
	svc             Service
	defaultAPIGroup gaia.RBACAPIGroup
}

func (s *policyEnforcer) GetDefaultAPIGroup() gaia.RBACAPIGroup {
	return s.defaultAPIGroup
}

// NewPolicyEnforcer creates a new policyEnforcer
func NewPolicyEnforcer(svc Service) (PolicyEnforcer, error) {
	file, err := ioutil.ReadFile("apigroup-core.yml")
	if err != nil {
		return nil, err
	}

	var apiGroup gaia.RBACAPIGroup
	if err := yaml.Unmarshal(file, &apiGroup); err != nil {
		return nil, err
	}

	return &policyEnforcer{
		svc:             svc,
		defaultAPIGroup: apiGroup,
	}, nil
}

// Enforce takes an EnforcerConfig containing User information (name, policies) and the required namespace, action and
// resource to enforce. Evaluate all the users permissions into a more efficient single map based structure
// gaia.RBACEvaluatedPermissions. Using the gaia.RBACEvaluatedPermissions we check if a user has the required
// permissions.
func (s *policyEnforcer) Enforce(cfg EnforcerConfig) error {
	resolved, err := s.Evaluate(cfg.User)
	if err != nil {
		return fmt.Errorf("error evaluating policies: %v", err.Error())
	}

	ns, nsExists := resolved[cfg.Namespace]
	if !nsExists {
		return errNamespaceNotFound
	}

	act, actionExists := ns[cfg.Action]
	if !actionExists {
		return errActionNotFound
	}

	// Empty cfg.Resource means there is no resource requirement for this enforcement
	if cfg.Resource == "" {
		return nil
	}

	if _, wcExists := act["*"]; wcExists {
		return nil
	}

	if _, resExists := act[cfg.Resource]; !resExists {
		return errResourceNotFound
	}

	return nil
}

// Evaluate evaluates all the policies a user is part of into a single gaia.RBACEvaluatedPermissions map structure.
// We first see if the gaia.RBACEvaluatedPermissions is within the global evaluatedPerms cache. If its not we have
// to get each policy the user is bound to and build it.
func (s *policyEnforcer) Evaluate(user User) (gaia.RBACEvaluatedPermissions, error) {
	// Use the service to look into the cache for any existing evaluated policies
	if policies, ok := s.svc.GetUserEvaluatedPolicies(user.Username); ok {
		return policies, nil
	}

	// Nothing in the cache, so start getting the policies for this user
	var stmts []gaia.RBACPolicyStatementV1
	for policyName := range user.Policies {
		policyResource, _ := s.svc.GetPolicy(policyName)
		stmts = append(stmts, policyResource.Statement...)
	}

	eval := make(gaia.RBACEvaluatedPermissions)

	// Evaluate all the policies, creating a single map and point of reference for the user.
	// This is not particularly efficient O(n^2), but we have to parse all namespaces and actions.
	for _, stmt := range stmts {

		for _, stmtAction := range stmt.Action {

			namespace, action := ParseStatementAction(stmtAction)

			// check if the namespace already exists in the evaluated perms.
			ns, nsExists := eval[namespace]
			if !nsExists {
				eval[namespace] = map[gaia.RBACPolicyAction]map[gaia.RBACPolicyResource]interface{}{
					action: stmtToMap(stmt.Resource),
				}
				continue
			}

			// check if the action already exists for the namespace.
			act, actExists := ns[action]
			if !actExists {
				ns[action] = stmtToMap(stmt.Resource)
				continue
			}

			// merge all resources into the action.
			for _, rs := range stmt.Resource {
				r := gaia.RBACPolicyResource(rs)
				_, resExists := act[r]
				if !resExists {
					act[r] = ""
					continue
				}
			}
		}

	}

	if err := s.svc.PutUserEvaluatedPolicies(user.Username, eval); err != nil {
		return nil, fmt.Errorf("failed to put evalued policies: %v", err.Error())
	}

	return eval, nil
}

func stmtToMap(resources []string) map[gaia.RBACPolicyResource]interface{} {
	m := map[gaia.RBACPolicyResource]interface{}{}
	for _, r := range resources {
		m[gaia.RBACPolicyResource(r)] = ""
	}
	return m
}

// ParseStatementAction takes a standard string and attempts to parse it into a gaia.RBACPolicyNamespace and
// gaia.RBACPolicyAction
func ParseStatementAction(a string) (gaia.RBACPolicyNamespace, gaia.RBACPolicyAction) {
	splitAction := strings.Split(a, "/")
	namespace := gaia.RBACPolicyNamespace(splitAction[0])
	action := gaia.RBACPolicyAction(splitAction[1])
	return namespace, action
}
