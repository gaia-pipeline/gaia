package store

import (
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/auth"
)

// Iterates the users and creates default permissions if they don't exist after upgrading.
func (s *BoltStore) CreatePermissionsIfNotExisting() error {
	users, _ := s.UserGetAll()
	for _, user := range users {
		perms, err := s.UserPermissionsGet(user.Username)
		if err != nil {
			return err
		}
		if perms == nil {
			perms := &gaia.UserPermission{
				Username: user.Username,
				Roles:    auth.FlattenUserCategoryRoles(auth.DefaultUserRoles),
				Groups:   []string{},
			}
			err := s.UserPermissionsPut(perms)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
