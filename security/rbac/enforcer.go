package rbac

import (
	"strings"

	"github.com/gaia-pipeline/gaia"
)

// PolicyEnforcer is for enforcing RBAC policies.
type PolicyEnforcer interface {
	Enforce(policyNames []string, namespace gaia.RBACPolicyNamespace, action gaia.RBACPolicyAction) bool
}

type policyEnforcer struct {
	svc Service
}

type namespaceActionMap map[gaia.RBACPolicyNamespace]map[gaia.RBACPolicyAction]interface{}

// NewPolicyEnforcer creates a new policyEnforcer
func NewPolicyEnforcer(svc Service) PolicyEnforcer {
	return &policyEnforcer{svc: svc}
}

// Enforce takes a list of policy names from a Gaia users JWT claims, a required namespace and action. We get all the
// users policies using the names provided. We then merged all the policies together into the namespaceActionMap which
// acts as a quick lookup for the namespace and action that is being enforced.
func (s *policyEnforcer) Enforce(policyNames []string, namespace gaia.RBACPolicyNamespace, action gaia.RBACPolicyAction) bool {
	resolved := s.resolvePolicies(policyNames)

	if ns, nsOk := resolved[namespace]; nsOk {
		// first, check for a wildcard.
		if _, acOk := ns["*"]; acOk {
			return true
		}
		// second, check for the specific action.
		if _, acOk := ns[action]; acOk {
			return true
		}
	}

	return false
}

func (s *policyEnforcer) resolvePolicies(policyNames []string) namespaceActionMap {
	na := make(namespaceActionMap)

	// iterate through the user policy names provided
	for _, policy := range policyNames {
		// get the policy from the service/cache
		policyResource, _ := s.svc.GetPolicy(policy)
		// iterate through the policy statement contained in the retrieved policy
		for _, stmt := range policyResource.Statement {
			// iterate through the actions in the statement
			for _, stmtAction := range stmt.Action {
				// parse the namespace and action from the statement value
				namespace, action := s.parseNamespaceAction(stmtAction)
				if nsActions, exists := na[namespace]; exists {
					if s.checkWildcard(na, namespace, action) {
						continue
					}
					nsActions[action] = ""
					continue
				}
				s.newEntry(na, namespace, action)
			}
		}
	}

	return na
}

func (s *policyEnforcer) newEntry(na namespaceActionMap, namespace gaia.RBACPolicyNamespace, action gaia.RBACPolicyAction) {
	na[namespace] = make(map[gaia.RBACPolicyAction]interface{})
	na[namespace][action] = ""
}

func (s *policyEnforcer) checkWildcard(na namespaceActionMap, namespace gaia.RBACPolicyNamespace, action gaia.RBACPolicyAction) bool {
	if _, wildcardExists := na[namespace]["*"]; wildcardExists {
		// continue if we already have a wildcard
		return true
	}
	if action == "*" {
		// overwrite map and assign wildcard - takes priority always
		na[namespace] = make(map[gaia.RBACPolicyAction]interface{})
		na[namespace]["*"] = ""
		return true
	}
	return false
}

func (s *policyEnforcer) parseNamespaceAction(a string) (gaia.RBACPolicyNamespace, gaia.RBACPolicyAction) {
	splitAction := strings.Split(a, "/")
	namespace := gaia.RBACPolicyNamespace(splitAction[0])
	action := gaia.RBACPolicyAction(splitAction[1])
	return namespace, action
}
