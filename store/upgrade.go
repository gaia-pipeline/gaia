package store

import (
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/auth"
)

func (s *BoltStore) CreateDefaultPermissions() error {
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
