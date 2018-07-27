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
var storeService *store.Store

// schedulerService is an instance of scheduler.
var schedulerService *scheduler.Scheduler

// StorageService initializes and keeps track of a storage service.
// If the internal storage service is a singleton.
func StorageService() *store.Store {
	if storeService != nil {
		return storeService
	}
	storeService = store.NewStore()
	err := storeService.Init()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize store", "error", err.Error())
		os.Exit(1)
	}
	return storeService
}

// SchedulerService initializes keeps track of the scheduler service.
// The internal service is a singleton.
func SchedulerService() *scheduler.Scheduler {
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
