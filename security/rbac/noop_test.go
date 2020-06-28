package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoOpService_Enforce_AlwaysReturnsNoError(t *testing.T) {
	svc := NewNoOpService()
	err := svc.Enforce("", "", "", map[string]string{})
	assert.NoError(t, err)
}

func TestNoOpService_AddRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.AddRole("", []RoleRule{})
	assert.EqualError(t, err, errNotEnabled.Error())
}

func TestNoOpService_DeleteRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.DeleteRole("")
	assert.EqualError(t, err, errNotEnabled.Error())
}

func TestNoOpService_GetAllRoles_ReturnsEmptySlice(t *testing.T) {
	svc := NewNoOpService()
	roles := svc.GetAllRoles()
	assert.Equal(t, roles, []string{})
}

func TestNoOpService_GetUserAttachedRoles_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	_, err := svc.GetUserAttachedRoles("")
	assert.EqualError(t, err, errNotEnabled.Error())
}

func TestNoOpService_GetRoleAttachedUsers_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	_, err := svc.GetRoleAttachedUsers("")
	assert.EqualError(t, err, errNotEnabled.Error())
}

func TestNoOpService_AttachRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.AttachRole("", "")
	assert.EqualError(t, err, errNotEnabled.Error())
}

func TestNoOpService_DetachRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.DetachRole("", "")
	assert.EqualError(t, err, errNotEnabled.Error())
}
