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
					Resource: "*",
					Effect:   "allow",
				},
				{
					Action: []string{
						"workers/get",
					},
					Resource: "worker:2",
					Effect:   "deny",
				},
				{
					Action: []string{
						"secrets/get",
					},
					Resource: "secret:1",
					Effect:   "allow",
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
					Resource: "*",
					Effect:   "deny",
				},
				{
					Action: []string{
						"workers/get",
					},
					Resource: "worker:1",
					Effect:   "allow",
				},
			},
		}, nil
	}
	return gaia.RBACPolicyResourceV1{}, errors.New("not found")
}

func (s mockSvc) PutPolicy(policy gaia.RBACPolicyResourceV1) error {
	panic("implement me")
}

func (s mockSvc) GetUserEvaluatedPolicies(username string) (gaia.RBACEvaluatedPermissions, bool) {
	panic("implement me")
}

func (s mockSvc) PutUserEvaluatedPolicies(username string, perms gaia.RBACEvaluatedPermissions) error {
	panic("implement me")
}

func TestPolicyEnforcer_Enforce_WithMissingPolicyStatement_ReturnsError(t *testing.T) {
	enforcer := NewPolicyEnforcer(mockSvc{})

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

func Test_PolicyEnforcer_ResolvePolicies_MergedPolicies(t *testing.T) {
	enforcer := NewPolicyEnforcer(mockSvc{})
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
				"*": "allow",
			},
			"get": {
				"*": "allow",
			},
		},
		"secrets": {
			"create": {
				"*": "deny",
			},
			"get": {
				"*":        "deny",
				"secret:1": "allow",
			},
			"delete": {
				"*": "allow",
			},
		},
		"workers": {
			"get": {
				"worker:1": "allow",
				"worker:2": "deny",
			},
		},
	}

	if !reflect.DeepEqual(rp, expected) {
		t.Errorf("Evaluate map: wanted %v, got %v", expected, rp)
	}
}
