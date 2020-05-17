package resourcehelper

import (
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"testing"

	"github.com/gaia-pipeline/gaia"
)

var specV1 = gaia.RBACPolicyResourceV1{
	ResourceType: gaia.ResourceType{
		Version: "v1",
		Type:    "authorization.policy",
	},
	ResourceMetadataV1: gaia.ResourceMetadataV1{
		Name:        "test",
		Description: "test policy.",
	},
	Statement: []gaia.RBACPolicyStatementV1{
		{
			ID:     "grant",
			Effect: "allow",
			Action: []string{
				"users/delete",
				"pipeline-runs/logs",
				"workers/deregister-worker",
			},
			Resource: "*",
		},
	},
}

var yamlSpecV1 = `version: v1
type: authorization.policy
metadata:
  name: test
  description: test policy.
statement:
- id: grant
  effect: allow
  action:
  - users/delete
  - pipeline-runs/logs
  - workers/deregister-worker
  resource: '*'
`

var yamlSpecV2 = `version: v2
type: authorization.policy
metadata:
  name: test
  description: test policy.
statement:
- id: grant-all
  effect: allow
  action:
  - users/delete
  - pipeline-runs/logs
  - workers/deregister-worker
resource: '*'
`

func Test_SpecMarshaller_MarshalV1(t *testing.T) {
	sm := NewMarshaller()

	got, err := sm.Marshal(specV1)

	assert.Check(t, cmp.Nil(err))
	assert.Check(t, cmp.Equal(string(got), yamlSpecV1))
}

func Test_SpecMarshaller_Unmarshal(t *testing.T) {
	sm := NewMarshaller()

	var unmarshalledSpec gaia.RBACPolicyResourceV1
	if err := sm.Unmarshal([]byte(yamlSpecV1), &unmarshalledSpec); err != nil {
		t.Errorf("Unmarshal() error = %v, wantErr no error", err)
		return
	}

	assert.Check(t, cmp.DeepEqual(unmarshalledSpec, specV1))
}

func Test_SpecMarshaller_Unmarshal_VersionMismatch_Errors(t *testing.T) {
	sm := NewMarshaller()

	var unmarshalledSpec gaia.RBACPolicyResourceV1
	err := sm.Unmarshal([]byte(yamlSpecV2), &unmarshalledSpec)

	assert.Check(t, cmp.Error(err, "version does not match struct RBACPolicyResourceV1"))
}

func Test_SpecMarshaller_Unmarshal_InvalidPolicyStruct_Errors(t *testing.T) {
	sm := NewMarshaller()

	var unmarshalledSpec struct{}
	err := sm.Unmarshal([]byte(yamlSpecV1), &unmarshalledSpec)

	assert.Check(t, cmp.Error(err, "policy specification struct not found"))
}
