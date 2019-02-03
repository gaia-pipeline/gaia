package mocks

import (
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
)

type Store struct {
	store.GaiaStore
	UserAuthFunc            func(u *gaia.User, updateLastLogin bool) (*gaia.User, error)
	UserPermissionsGetFunc  func(username string) (*gaia.UserPermission, error)
	UserPermissionsPutFunc  func(perms *gaia.UserPermission) error
	UserPermissionsGroupGet func(name string) (*gaia.UserPermissionGroup, error)
}

func (s *Store) UserAuth(u *gaia.User, updateLastLogin bool) (*gaia.User, error) {
	return s.UserAuthFunc(u, updateLastLogin)
}

func (s *Store) UserPermissionsGet(username string) (*gaia.UserPermission, error) {
	return s.UserPermissionsGetFunc(username)
}

func (s *Store) UserPermissionsPut(perms *gaia.UserPermission) error {
	return s.UserPermissionsPutFunc(perms)
}

func (s *Store) UserPermissionGroupGet(name string) (*gaia.UserPermissionGroup, error) {
	return s.UserPermissionsGroupGet(name)
}
