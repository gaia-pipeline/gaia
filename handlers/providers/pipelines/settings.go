package pipelines

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/services"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
)

type settingsHandler struct {
	store store.SettingsStore
}

func newSettingsHandler(store store.SettingsStore) *settingsHandler {
	return &settingsHandler{store: store}
}

// SettingsPollOn turn on polling functionality
func (pp *pipelineProvider) SettingsPollOn(c echo.Context) error {
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
	err = pp.deps.PipelineService.StartPoller()
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	configStore.Poll = true
	err = h.store.SettingsPut(configStore)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Polling is turned on.")
}

// SettingsPollOff turn off polling functionality.
func (pp *pipelineProvider) SettingsPollOff(c echo.Context) error {
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
	err = pp.deps.PipelineService.StopPoller()
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	configStore.Poll = true
	err = h.store.SettingsPut(configStore)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Polling is turned off.")
}

type pollStatus struct {
	Status bool
}

// SettingsPollGet get status of polling functionality.
func (pp *pipelineProvider) SettingsPollGet(c echo.Context) error {
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

func (h *settingsHandler) rbacPut(c echo.Context) error {
	var cfg gaia.RBACConfig
	if err := c.Bind(&cfg); err != nil {
		return c.String(http.StatusInternalServerError, "Unable to parse request body.")
	}

	if err := h.store.SettingsRBACPut(cfg); err != nil {
		return c.String(http.StatusInternalServerError, "An error occurred while saving the settings.")
	}

	return c.String(http.StatusOK, "Settings have been updated.")
}

func (h *settingsHandler) rbacGet(c echo.Context) error {
	config, err := h.store.SettingsRBACGet()
	if err != nil {
		return c.String(http.StatusInternalServerError, "An error has occurred when retrieving settings.")
	}

	return c.JSON(http.StatusOK, config)
}
