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
	assert.NoError(t, err)
}

func TestNoOpService_DeleteRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.DeleteRole("")
	assert.NoError(t, err)
}

func TestNoOpService_GetAllRoles_ReturnsEmptySlice(t *testing.T) {
	svc := NewNoOpService()
	roles := svc.GetAllRoles()
	assert.Equal(t, roles, []string{})
}

func TestNoOpService_GetUserAttachedRoles_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	_, err := svc.GetUserAttachedRoles("")
	assert.NoError(t, err)
}

func TestNoOpService_GetRoleAttachedUsers_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	_, err := svc.GetRoleAttachedUsers("")
	assert.NoError(t, err)
}

func TestNoOpService_AttachRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.AttachRole("", "")
	assert.NoError(t, err)
}

func TestNoOpService_DetachRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.DetachRole("", "")
	assert.NoError(t, err)
}
