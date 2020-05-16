package resourcehelper

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gaia-pipeline/gaia"
)

var specV1 = gaia.RBACPolicyResourceV1{
	ResourceType: gaia.ResourceType{
		Version: "v1",
		Type:    "authorization.policy",
	},
	ResourceMetadataV1: gaia.ResourceMetadataV1{
		Name:        "test-name",
		Description: "test description",
	},
	Statement: []gaia.RBACPolicyStatementV1{
		{
			ID:     "test-id",
			Effect: "allow",
			Action: []string{
				"namespace/action-name-a",
				"namespace/action-name-b",
			},
		},
	},
}

var yamlSpecV1 = `version: v1
type: authorization.policy
metadata:
  name: test-name
  description: test description
statement:
- id: test-id
  effect: allow
  action:
  - namespace/action-name-a
  - namespace/action-name-b
`

var yamlSpecV2 = `version: v2
type: authorization.policy
metadata:
  name: test-name
  description: test description
statement:
- id: test-id
  effect: allow
  action:
  - namespace/action-name-a
  - namespace/action-name-b
`

func Test_SpecMarshaller_MarshalV1(t *testing.T) {
	sm := NewMarshaller()

	got, err := sm.Marshal(specV1)
	if err != nil {
		t.Errorf("Marshal() error = %v, wantErr no error", err)
		return
	}
	gotStr := string(got)
	if !reflect.DeepEqual(gotStr, yamlSpecV1) {
		t.Errorf("Marshal() got = %v, expected = %v", gotStr, yamlSpecV1)
	}
}

func Test_SpecMarshaller_Unmarshal(t *testing.T) {
	sm := NewMarshaller()

	var unmarshalledSpec gaia.RBACPolicyResourceV1
	if err := sm.Unmarshal([]byte(yamlSpecV1), &unmarshalledSpec); err != nil {
		t.Errorf("Unmarshal() error = %v, wantErr no error", err)
		return
	}
	if !reflect.DeepEqual(unmarshalledSpec, specV1) {
		t.Errorf("Unmarshal() got = %v, expected = %v", unmarshalledSpec, specV1)
	}
}

func Test_SpecMarshaller_Unmarshal_VersionMismatch_Errors(t *testing.T) {
	sm := NewMarshaller()

	expected := errors.New("version does not match struct RBACPolicyResourceV1")

	var unmarshalledSpec gaia.RBACPolicyResourceV1
	err := sm.Unmarshal([]byte(yamlSpecV2), &unmarshalledSpec)
	if err == nil {
		t.Errorf("Unmarshal error = %v, wantErr = %v", err, expected)
		return
	}
	if err.Error() != expected.Error() {
		t.Errorf("Unmarshal error = %v, wantErr = %v", err, expected)
	}
}

func Test_SpecMarshaller_Unmarshal_InvalidPolicyStruct_Errors(t *testing.T) {
	sm := NewMarshaller()

	expected := errors.New("policy specification struct not found")

	var unmarshalledSpec struct{}
	err := sm.Unmarshal([]byte(yamlSpecV1), &unmarshalledSpec)
	if err == nil {
		t.Errorf("Unmarshal error = %v, wantErr = %v", err, expected)
		return
	}
	if err.Error() != expected.Error() {
		t.Errorf("Unmarshal error = %v, wantErr = %v", err, expected)
	}
}
