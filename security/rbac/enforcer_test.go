package rbac

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gaia-pipeline/gaia"
)

type mockSvc struct {
}

func (s mockSvc) GetPolicy(policy string) (gaia.RBACPolicyResourceV1, error) {
	switch policy {
	case "test-policy-a":
		return gaia.RBACPolicyResourceV1{
			Statement: []gaia.RBACPolicyStatementV1{
				{
					Action: []string{
						"pipelines/get",
						"pipelines/create",
						"secrets/delete",
					},
					Resource: []string{"*"},
				},
				{
					Action: []string{
						"secrets/get",
					},
					Resource: []string{"secrets/key/my-secret"},
				},
			},
		}, nil
	case "test-policy-b":
		return gaia.RBACPolicyResourceV1{
			Statement: []gaia.RBACPolicyStatementV1{
				{
					Action: []string{
						"secrets/get",
						"secrets/create",
					},
					Resource: []string{"*"},
				},
			},
		}, nil
	}
	return gaia.RBACPolicyResourceV1{}, errors.New("not found")
}

func (s mockSvc) PutPolicy(policy gaia.RBACPolicyResourceV1) error {
	panic("implement me")
}

func (s mockSvc) PutUserBinding(username string, policy string) error {
	panic("implement me")
}

func (s mockSvc) GetUserEvaluatedPolicies(username string) (gaia.RBACEvaluatedPermissions, bool) {
	return make(gaia.RBACEvaluatedPermissions), false
}

func (s mockSvc) PutUserEvaluatedPolicies(username string, perms gaia.RBACEvaluatedPermissions) error {
	return nil
}

func TestPolicyEnforcer_Enforce_WithMissingPolicyStatement_ReturnsError(t *testing.T) {
	enforcer := policyEnforcer{
		svc: mockSvc{},
	}

	err := enforcer.Enforce(EnforcerConfig{
		User: User{
			Username: "test",
			Policies: map[string]interface{}{
				"test-policy-a": "",
			},
		},
		Namespace: "pipeline-runs",
		Action:    "get",
		Resource:  "pipeline:id",
	})

	if err == nil {
		t.Fatal("expected err to be nil")
	}
}

func Test_PolicyEnforcer_Evaluate_MergedPolicies(t *testing.T) {
	enforcer := policyEnforcer{
		svc: mockSvc{},
	}

	rp, _ := enforcer.Evaluate(User{
		Username: "test",
		Policies: map[string]interface{}{
			"test-policy-a": "",
			"test-policy-b": "",
		},
	})

	expected := gaia.RBACEvaluatedPermissions{
		"pipelines": {
			"create": {
				"*": "",
			},
			"get": {
				"*": "",
			},
		},
		"secrets": {
			"create": {
				"*": "",
			},
			"get": {
				"*":                     "",
				"secrets/key/my-secret": "",
			},
			"delete": {
				"*": "",
			},
		},
	}

	if !reflect.DeepEqual(rp, expected) {
		t.Errorf("Evaluate map: wanted %v, got %v", expected, rp)
	}
}
