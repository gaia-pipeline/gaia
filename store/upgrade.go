package store

import "github.com/gaia-pipeline/gaia"

func (s *BoltStore) CreateDefaultPermissions() error {
	users, _ := s.UserGetAll()
	for _, user := range users {
		perms, err := s.UserPermissionsGet(user.Username)
		if err != nil {
			return err
		}
		if perms == nil {
			gaia.Cfg.Logger.Info("creating default permissions", "user", user.Username)
			perms := &gaia.UserPermission{
				Username: user.Username,
				Roles:    gaia.GetFlattenedUserRoles(),
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
