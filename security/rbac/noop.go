package rbac

type noOpService struct{}

// NewNoOpService is used to instantiated a noOpService for when rbac enabled=false.
func NewNoOpService() Service {
	return &noOpService{}
}

// Enforce no-op enforcement. Allows everything.
func (n noOpService) Enforce(username, method, path string, params map[string]string) error {
	// Allow all
	return nil
}

// AddRole that errors since rbac is not enabled.
func (n noOpService) AddRole(role string, roleRules []RoleRule) error {
	return nil
}

// DeleteRole that errors since rbac is not enabled.
func (n noOpService) DeleteRole(role string) error {
	return nil
}

// GetAllRoles that returns nothing since rbac is not enabled.
func (n noOpService) GetAllRoles() []string {
	return []string{}
}

// GetUserAttachedRoles that errors since rbac is not enabled.
func (n noOpService) GetUserAttachedRoles(username string) ([]string, error) {
	return nil, nil
}

// GetRoleAttachedUsers that errors since rbac is not enabled.
func (n noOpService) GetRoleAttachedUsers(role string) ([]string, error) {
	return nil, nil
}

// AttachRole that errors since rbac is not enabled.
func (n noOpService) AttachRole(username string, role string) error {
	return nil
}

// DetachRole that errors since rbac is not enabled.
func (n noOpService) DetachRole(username string, role string) error {
	return nil
}

func (n noOpService) DeleteUser(username string) error {
	return nil
}
