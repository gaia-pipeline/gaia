package services

import (
	"os"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/plugin"
	"github.com/gaia-pipeline/gaia/scheduler"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/store"
)

// storeService is an instance of store.
// Use this to talk to the store.
var storeService store.GaiaStore

// schedulerService is an instance of scheduler.
var schedulerService scheduler.GaiaScheduler

// certificateService is the singleton holding the certificate manager.
var certificateService security.CAAPI

var vaultService security.VaultAPI

// StorageService initializes and keeps track of a storage service.
// If the internal storage service is a singleton.
func StorageService() store.GaiaStore {
	if storeService != nil {
		return storeService
	}
	storeService = store.NewBoltStore()
	err := storeService.Init()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize store", "error", err.Error())
		os.Exit(1)
	}
	return storeService
}

// MockStorageService sets the internal store singleton to the give
// mock implementation. A mock needs to be created in the test. The
// provider will make sure that everything that would use the store
// will use the mock instead.
func MockStorageService(store store.GaiaStore) {
	storeService = store
}

// SchedulerService initializes keeps track of the scheduler service.
// The internal service is a singleton.
func SchedulerService() scheduler.GaiaScheduler {
	if schedulerService != nil {
		return schedulerService
	}
	pS := &plugin.Plugin{}
	schedulerService = scheduler.NewScheduler(storeService, pS, certificateService)
	err := schedulerService.Init()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize scheduler:", "error", err.Error())
		os.Exit(1)
	}
	return schedulerService
}

// MockSchedulerService which replaces the scheduler service
// with a mocked one.
func MockSchedulerService(scheduler scheduler.GaiaScheduler) {
	schedulerService = scheduler
}

// CertificateService creates a certificate manager service.
func CertificateService() (security.CAAPI, error) {
	if certificateService != nil {
		return certificateService, nil
	}

	c, err := security.InitCA()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize certificate manager:", "error", err.Error())
		return nil, err
	}
	certificateService = c
	return certificateService, nil
}

// MockCertificateService provides a way to create and set a mock
// for the internal certificate service manager.
func MockCertificateService(service security.CAAPI) {
	certificateService = service
}

// VaultService creates a vault manager service.
func VaultService(vaultStore security.VaultStorer) (security.VaultAPI, error) {
	if vaultService != nil {
		return vaultService, nil
	}

	v, err := security.NewVault(certificateService, vaultStore)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize vault manager:", "error", err.Error())
		return nil, err
	}
	vaultService = v
	return vaultService, nil
}

// MockVaultService provides a way to create and set a mock
// for the internal vault service manager.
func MockVaultService(service security.VaultAPI) {
	vaultService = service
}
