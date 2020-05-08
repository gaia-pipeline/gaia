package services

import (
	"reflect"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/store/memdb"
)

// storeService is an instance of store.
// Use this to talk to the store.
var storeService store.GaiaStore

// vaultService is an instance of the internal Vault.
var vaultService security.GaiaVault

// memDBService is an instance of the internal memdb.
var memDBService memdb.GaiaMemDB

// StorageService initializes and keeps track of a storage service.
// If the internal storage service is a singleton. This function retruns an error
// but most of the times we don't care about it, because it's only ever
// initialized once in the main.go. If it wouldn't work, main would
// os.Exit(1) and the rest of the application would just stop.
func StorageService() (store.GaiaStore, error) {
	if storeService != nil && !reflect.ValueOf(storeService).IsNil() {
		return storeService, nil
	}
	storeService = store.NewBoltStore()
	err := storeService.Init(gaia.Cfg.DataPath)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize store", "error", err.Error())
		return storeService, err
	}
	return storeService, nil
}

// MockStorageService sets the internal store singleton to the give
// mock implementation. A mock needs to be created in the test. The
// provider will make sure that everything that would use the store
// will use the mock instead.
func MockStorageService(store store.GaiaStore) {
	storeService = store
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
func DefaultMemDBService() (memdb.GaiaMemDB, error) {
	s, err := StorageService()
	if err != nil {
		return nil, err
	}
	return MemDBService(s)
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
