package rbac

import (
	"strings"

	"github.com/gaia-pipeline/gaia"
)

type PolicyEnforcer interface {
	Enforce(policyNames []string, namespace gaia.AuthPolicyNamespace, action gaia.AuthPolicyAction) bool
}

type policyEnforcer struct {
	svc Service
}

type namespaceActionMap map[gaia.AuthPolicyNamespace]map[gaia.AuthPolicyAction]interface{}

func NewPolicyEnforcer(svc Service) *policyEnforcer {
	return &policyEnforcer{svc: svc}
}

func (s *policyEnforcer) Enforce(policyNames []string, namespace gaia.AuthPolicyNamespace, action gaia.AuthPolicyAction) bool {
	resolved := s.resolvePolicies(policyNames)

	if ns, nsOk := resolved[namespace]; nsOk {
		// First, check for a wildcard.
		if _, acOk := ns["*"]; acOk {
			return true
		}
		// Second, check for a specific action.
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
					if action == "*" {
						na[namespace] = make(map[gaia.AuthPolicyAction]interface{})
						na[namespace]["*"] = ""
						continue
					}
					nsActions[action] = ""
					continue
				}
				na[namespace] = make(map[gaia.AuthPolicyAction]interface{})
				na[namespace][action] = ""
			}
		}
	}

	return na
}

func (s *policyEnforcer) parseNamespaceAction(a string) (gaia.AuthPolicyNamespace, gaia.AuthPolicyAction) {
	splitAction := strings.Split(a, "/")
	namespace := gaia.AuthPolicyNamespace(splitAction[0])
	action := gaia.AuthPolicyAction(splitAction[1])
	return namespace, action
}
