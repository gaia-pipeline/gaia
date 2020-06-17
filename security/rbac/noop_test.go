package rbac

import (
	"gotest.tools/assert"
	"testing"
)

func TestNoOpService_Enforce_AlwaysReturnsTrue(t *testing.T) {
	svc := NewNoOpService()
	result, err := svc.Enforce("", "", "", map[string]string{})
	assert.NilError(t, err)
	assert.Equal(t, result, true)
}

func TestNoOpService_AddRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.AddRole("", []RoleRule{})
	assert.Error(t, err, errNotEnabled.Error())
}

func TestNoOpService_DeleteRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.DeleteRole("")
	assert.Error(t, err, errNotEnabled.Error())
}

func TestNoOpService_GetAllRoles_ReturnsEmptySlice(t *testing.T) {
	svc := NewNoOpService()
	roles := svc.GetAllRoles()
	assert.DeepEqual(t, roles, []string{})
}

func TestNoOpService_GetUserAttachedRoles_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	_, err := svc.GetUserAttachedRoles("")
	assert.Error(t, err, errNotEnabled.Error())
}

func TestNoOpService_GetRoleAttachedUsers_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	_, err := svc.GetRoleAttachedUsers("")
	assert.Error(t, err, errNotEnabled.Error())
}

func TestNoOpService_AttachRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.AttachRole("", "")
	assert.Error(t, err, errNotEnabled.Error())
}

func TestNoOpService_DetachRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.DetachRole("", "")
	assert.Error(t, err, errNotEnabled.Error())
}
