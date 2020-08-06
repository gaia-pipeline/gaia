package rbac

import (
	"errors"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/stretchr/testify/require"
)

type mockCasbinEnforcer struct {
	casbin.IEnforcer
	addPolicyFn         func(rules [][]string) (bool, error)
	getAllSubjectsFn    func() []string
	deleteFn            func(role string) (bool, error)
	getRolesForUserFn   func(name string, domain ...string) ([]string, error)
	getUsersForRoleFn   func(name string, domain ...string) ([]string, error)
	addRoleForUserFn    func(user string, role string) (bool, error)
	deleteRoleForUserFn func(user string, role string) (bool, error)
	deleteUserFn        func(user string) (bool, error)
}

func (m *mockCasbinEnforcer) AddPolicies(rules [][]string) (bool, error) {
	return m.addPolicyFn(rules)
}

func (m *mockCasbinEnforcer) GetAllSubjects() []string {
	return m.getAllSubjectsFn()
}

func (m *mockCasbinEnforcer) DeleteRole(role string) (bool, error) {
	return m.deleteFn(role)
}

func (m *mockCasbinEnforcer) GetRolesForUser(name string, domain ...string) ([]string, error) {
	return m.getRolesForUserFn(name, domain...)
}

func (m *mockCasbinEnforcer) GetUsersForRole(name string, domain ...string) ([]string, error) {
	return m.getUsersForRoleFn(name, domain...)
}

func (m *mockCasbinEnforcer) AddRoleForUser(user string, role string) (bool, error) {
	return m.addRoleForUserFn(user, role)
}

func (m *mockCasbinEnforcer) DeleteRoleForUser(user string, role string) (bool, error) {
	return m.deleteRoleForUserFn(user, role)
}

func (m *mockCasbinEnforcer) DeleteUser(user string) (bool, error) {
	return m.deleteUserFn(user)
}

func TestEnforcerService_AddRole_WithMissingPrefix_ReturnsError(t *testing.T) {
	ce := &mockCasbinEnforcer{}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.AddRole("noprefix", []RoleRule{})
	require.EqualError(t, err, "role must be prefixed with 'role:'")
}

func TestEnforcerService_DeleteRole_WithRoleNotExists_ReturnsError(t *testing.T) {
	ce := &mockCasbinEnforcer{
		deleteFn: func(role string) (bool, error) {
			require.Equal(t, "notexisting", role)
			return false, nil
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.DeleteRole("notexisting")
	require.EqualError(t, err, "role does not exist")
}

func TestEnforcerService_DeleteRole_WithError_ReturnsError(t *testing.T) {
	ce := &mockCasbinEnforcer{
		deleteFn: func(role string) (bool, error) {
			return true, errors.New("an error")
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.DeleteRole("role:valid")
	require.EqualError(t, err, "error deleting role: an error")
}

func TestEnforcerService_DeleteRole_WithValid_Success(t *testing.T) {
	ce := &mockCasbinEnforcer{
		deleteFn: func(role string) (bool, error) {
			require.Equal(t, "role:valid", role)
			return true, nil
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.DeleteRole("role:valid")
	require.NoError(t, err)
}

func TestEnforcerService_AddRole_WithNoRulesArgs_Success(t *testing.T) {
	ce := &mockCasbinEnforcer{
		addPolicyFn: func(rules [][]string) (bool, error) {
			return true, nil
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.AddRole("role:newrole", []RoleRule{})
	require.NoError(t, err)
}

func TestEnforcerService_AddRole_WithValidRules_Success(t *testing.T) {
	expectedRules := [][]string{{"role:newrole", "ns-a", "act", "*", "allow"}, {"role:newrole", "ns-b", "act", "*", "deny"}}

	ce := &mockCasbinEnforcer{
		addPolicyFn: func(rules [][]string) (bool, error) {
			require.EqualValues(t, expectedRules, rules)
			return true, nil
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.AddRole("role:newrole", []RoleRule{
		{
			Namespace: "ns-a",
			Action:    "act",
			Resource:  "*",
			Effect:    "allow",
		},
		{
			Namespace: "ns-b",
			Action:    "act",
			Resource:  "*",
			Effect:    "deny",
		},
	})
	require.NoError(t, err)
}

func TestEnforcerService_GetAllRoles(t *testing.T) {
	ce := &mockCasbinEnforcer{
		getAllSubjectsFn: func() []string {
			return []string{"user", "role:admin", "role:test", "user-b", "role:super"}
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	roles := svc.GetAllRoles()
	require.Equal(t, []string{"role:admin", "role:test", "role:super"}, roles)
}

func TestEnforcerService_GetUserAttachedRoles_WithError_ReturnsError(t *testing.T) {
	ce := &mockCasbinEnforcer{
		getRolesForUserFn: func(name string, domain ...string) ([]string, error) {
			return nil, errors.New("an error")
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	_, err := svc.GetUserAttachedRoles("admin")
	require.EqualError(t, err, "error getting roles for user: an error")
}

func TestEnforcerService_GetUserAttachedRoles_Success(t *testing.T) {
	ce := &mockCasbinEnforcer{
		getRolesForUserFn: func(name string, domain ...string) ([]string, error) {
			require.Equal(t, "admin", name)
			require.Equal(t, []string(nil), domain)
			return []string{"role:admin", "role:another"}, nil
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	roles, err := svc.GetUserAttachedRoles("admin")
	require.NoError(t, err)
	require.Equal(t, []string{"role:admin", "role:another"}, roles)
}

func TestEnforcerService_GetRoleAttachedUsers_WithError_ReturnsError(t *testing.T) {
	ce := &mockCasbinEnforcer{
		getUsersForRoleFn: func(name string, domain ...string) ([]string, error) {
			return nil, errors.New("an error")
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	_, err := svc.GetRoleAttachedUsers("role:admin")
	require.EqualError(t, err, "error getting users for role: an error")
}

func TestEnforcerService_GetRoleAttachedUsers_Success(t *testing.T) {
	ce := &mockCasbinEnforcer{
		getUsersForRoleFn: func(name string, domain ...string) ([]string, error) {
			require.Equal(t, "role:admin", name)
			require.Equal(t, []string(nil), domain)
			return []string{"admin", "sam"}, nil
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	roles, err := svc.GetRoleAttachedUsers("role:admin")
	require.NoError(t, err)
	require.Equal(t, []string{"admin", "sam"}, roles)
}

func TestEnforcerService_AttachRole_WhenAlreadyAttached_ReturnError(t *testing.T) {
	ce := &mockCasbinEnforcer{
		addRoleForUserFn: func(user string, role string) (bool, error) {
			return true, nil
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.AttachRole("admin", "role:admin")
	require.EqualError(t, err, "user already has the role attached")
}

func TestEnforcerService_AttachRole_WhenErrorOccurs_ReturnError(t *testing.T) {
	ce := &mockCasbinEnforcer{
		addRoleForUserFn: func(user string, role string) (bool, error) {
			return false, errors.New("an error")
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.AttachRole("admin", "role:admin")
	require.EqualError(t, err, "error attatching role to user: an error")
}

func TestEnforcerService_AttachRole_Success(t *testing.T) {
	ce := &mockCasbinEnforcer{
		addRoleForUserFn: func(user string, role string) (bool, error) {
			require.Equal(t, "admin", user)
			require.Equal(t, "role:admin", role)
			return false, nil
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.AttachRole("admin", "role:admin")
	require.NoError(t, err)
}

func TestEnforcerService_DetatchRole_WhenNotAttached_ReturnError(t *testing.T) {
	ce := &mockCasbinEnforcer{
		deleteRoleForUserFn: func(user string, role string) (bool, error) {
			return false, nil
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.DetachRole("admin", "role:admin")
	require.EqualError(t, err, "role not attached to user")
}

func TestEnforcerService_DetachRole_WhenErrorOccurs_ReturnError(t *testing.T) {
	ce := &mockCasbinEnforcer{
		deleteRoleForUserFn: func(user string, role string) (bool, error) {
			return false, errors.New("an error")
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.DetachRole("admin", "role:admin")
	require.EqualError(t, err, "error detatching role from user: an error")
}

func TestEnforcerService_DetachRole_Success(t *testing.T) {
	ce := &mockCasbinEnforcer{
		deleteRoleForUserFn: func(user string, role string) (bool, error) {
			require.Equal(t, "admin", user)
			require.Equal(t, "role:admin", role)
			return true, nil
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.DetachRole("admin", "role:admin")
	require.NoError(t, err)
}

func TestEnforcerService_DeleteUser_WhenErrorOccurs_ReturnError(t *testing.T) {
	ce := &mockCasbinEnforcer{
		deleteUserFn: func(user string) (bool, error) {
			return false, errors.New("an error")
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.DeleteUser("admin")
	require.EqualError(t, err, "error deleting user: an error")
}

func TestEnforcerService_DeleteUser_Success(t *testing.T) {
	ce := &mockCasbinEnforcer{
		deleteUserFn: func(user string) (bool, error) {
			require.Equal(t, "admin", user)
			return false, nil
		},
	}
	apiLookup := APILookup{}

	svc := NewEnforcerSvc(ce, apiLookup)

	err := svc.DeleteUser("admin")
	require.NoError(t, err)
}
