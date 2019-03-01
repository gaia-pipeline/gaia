package handlers

import (
	"net/http"

	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/workers/pipeline"

	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
)

// SettingsPollOn turn on polling functionality.
func SettingsPollOn(c echo.Context) error {
	storeService, err := services.StorageService()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Something went wrong while getting storage service.")
	}
	configStore, err := storeService.SettingsGet()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Something went wrong while retrieving settings information.")
	}
	if configStore == nil {
		configStore = &gaia.StoreConfig{}
	}

	gaia.Cfg.Poll = true
	err = pipeline.StartPoller()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}

	configStore.Poll = true
	err = storeService.SettingsPut(configStore)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Polling is turned on.")
}

// SettingsPollOff turn off polling functionality.
func SettingsPollOff(c echo.Context) error {
	storeService, err := services.StorageService()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Something went wrong while getting storage service.")
	}
	configStore, err := storeService.SettingsGet()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Something went wrong while retrieving settings information.")
	}
	if configStore == nil {
		configStore = &gaia.StoreConfig{}
	}
	gaia.Cfg.Poll = false
	err = pipeline.StopPoller()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	configStore.Poll = true
	err = storeService.SettingsPut(configStore)
	return c.String(http.StatusOK, "Polling is turned off.")
}

type pollStatus struct {
	Status bool
}

// SettingsPollGet get status of polling functionality.
func SettingsPollGet(c echo.Context) error {
	storeService, err := services.StorageService()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Something went wrong while getting storage service.")
	}
	configStore, err := storeService.SettingsGet()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Something went wrong while retrieving settings information.")
	}
	var ps pollStatus
	if configStore == nil {
		ps.Status = gaia.Cfg.Poll
	} else {
		ps.Status = configStore.Poll
	}
	return c.JSON(http.StatusOK, ps)
}
