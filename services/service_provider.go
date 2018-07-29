package services

import (
	"os"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/plugin"
	"github.com/gaia-pipeline/gaia/scheduler"
	"github.com/gaia-pipeline/gaia/store"
)

// storeService is an instance of store.
// Use this to talk to the store.
var storeService store.GaiaStore

// schedulerService is an instance of scheduler.
var schedulerService scheduler.GaiaScheduler

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
	schedulerService = scheduler.NewScheduler(StorageService(), pS)
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
