package rbac

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gaia-pipeline/gaia"
)

type mockSvc struct {
}

func (s mockSvc) GetPolicy(name string) (gaia.AuthPolicyResourceV1, error) {
	switch name {
	case "test-policy-a":
		return gaia.AuthPolicyResourceV1{
			Statement: []gaia.AuthPolicyStatementV1{
				{
					ID:     "test-id",
					Effect: "allow",
					Action: []string{
						"namespace-a/action-name-a",
						"namespace-b/action-name-b",
					},
				},
			},
		}, nil
	case "test-policy-b":
		return gaia.AuthPolicyResourceV1{
			Statement: []gaia.AuthPolicyStatementV1{
				{
					ID:     "test-id",
					Effect: "allow",
					Action: []string{
						"namespace-b/action-name-a",
						"namespace-b/action-name-b",
						"namespace-a/*",
					},
				},
			},
		}, nil
	}
	return gaia.AuthPolicyResourceV1{}, errors.New("not found")
}

func TestPolicyEnforcer_Enforce_WithWildcardMultipleMergedPolicies_IsTrue(t *testing.T) {
	enforcer := NewPolicyEnforcer(mockSvc{})

	isValid := enforcer.Enforce([]string{"test-policy-a", "test-policy-b"}, "namespace-a", "action-name-a")

	if !isValid {
		t.Fatal("expected isValid to be true")
	}
}

func TestPolicyEnforcer_Enforce_WithMissingPolicyStatement_IsFalse(t *testing.T) {
	enforcer := NewPolicyEnforcer(mockSvc{})

	isValid := enforcer.Enforce([]string{"test-policy-a"}, "namespace-c", "action-name-a")

	if isValid {
		t.Fatal("expected isValid to be false")
	}
}

func Test_PolicyEnforcer_ResolvePolicies_MergedPolicies(t *testing.T) {
	enforcer := NewPolicyEnforcer(mockSvc{})
	rp := enforcer.ResolvePolicies([]string{"test-policy-a", "test-policy-b"})

	expectedRp := make(namespaceActionMap)
	expectedRp["namespace-a"] = map[gaia.RBACPolicyAction]interface{}{
		"*": "",
	}
	expectedRp["namespace-b"] = map[gaia.RBACPolicyAction]interface{}{
		"action-name-a": "",
		"action-name-b": "",
	}

	if !reflect.DeepEqual(rp, expectedRp) {
		t.Errorf("ResolvePolicies map: wanted %v, got %v", expectedRp, rp)
	}
}
