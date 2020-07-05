package services

import (
	"reflect"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/store/memdb"
)

// vaultService is an instance of the internal Vault.
var vaultService security.GaiaVault

// memDBService is an instance of the internal memdb.
var memDBService memdb.GaiaMemDB

// NewStorageService a creates new storage service and initializes it.
func NewStorageService() (store.GaiaStore, error) {
	storeService := store.NewBoltStore()
	if err := storeService.Init(gaia.Cfg.DataPath); err != nil {
		gaia.Cfg.Logger.Error("cannot initialize store", "error", err.Error())
		return nil, err
	}
	return storeService, nil
}

// DefaultVaultService provides a vault with a FileStorer backend.
func DefaultVaultService() (security.GaiaVault, error) {
	return VaultService(&security.FileVaultStorer{})
}

// VaultService creates a vault manager service.
func VaultService(vaultStore security.VaultStorer) (security.GaiaVault, error) {
	if vaultService != nil && !reflect.ValueOf(vaultService).IsNil() {
		return vaultService, nil
	}

	// TODO: For now use this to keep the refactor of certificate out of the refactor of Vault Service.
	ca, err := security.InitCA()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize certificate:", "error", err.Error())
		return nil, err
	}
	v, err := security.NewVault(ca, vaultStore)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize vault manager:", "error", err.Error())
		return nil, err
	}
	vaultService = v
	return vaultService, nil
}

// MockVaultService provides a way to create and set a mock
// for the internal vault service manager.
func MockVaultService(service security.GaiaVault) {
	vaultService = service
}

// DefaultMemDBService provides a default memDBService with an underlying storer.
func DefaultMemDBService(store store.GaiaStore) (memdb.GaiaMemDB, error) {
	return MemDBService(store)
}

// MemDBService creates a memdb service instance.
func MemDBService(store store.GaiaStore) (memdb.GaiaMemDB, error) {
	if memDBService != nil && !reflect.ValueOf(memDBService).IsNil() {
		return memDBService, nil
	}

	db, err := memdb.InitMemDB(store)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize memdb service", "error", err.Error())
		return nil, err
	}
	memDBService = db
	return memDBService, nil
}

// MockMemDBService provides a way to create and set a mock
// for the internal memdb service manager.
func MockMemDBService(db memdb.GaiaMemDB) {
	memDBService = db
}
