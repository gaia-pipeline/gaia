package security

import (
	"errors"
	"fmt"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
)

type PermissionManagerIface interface {
	CreateGroup(group *gaia.PermissionGroup, overwrite bool) error
}

type PermissionManager struct {
	store store.GaiaStore
}

func NewPermissionManager(store store.GaiaStore) *PermissionManager {
	return &PermissionManager{store: store}
}

func (pm *PermissionManager) CreateGroup(group *gaia.PermissionGroup, overwrite bool) error {
	gStore := pm.store

	if !overwrite {
		pg, err := gStore.PermissionGroupGet(group.Name)
		if err != nil {
			return err
		}

		if pg != nil {
			return errors.New(fmt.Sprintf("Group already exists with the name %s", group.Name))
		}
	}

	err := gStore.PermissionGroupPut(group)
	return err
}
