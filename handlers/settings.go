package handlers

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
)

const msgSomethingWentWrong = "Something went wrong while retrieving settings information."

type settingsHandler struct {
	store store.SettingsStore
}

func newSettingsHandler(store store.SettingsStore) *settingsHandler {
	return &settingsHandler{store: store}
}

// SettingsPollOn turn on polling functionality.
func (h *settingsHandler) pollOn(c echo.Context) error {
	configStore, err := h.store.SettingsGet()
	if err != nil {
		return c.String(http.StatusInternalServerError, msgSomethingWentWrong)
	}

	gaia.Cfg.Poll = true

	if err := pipeline.StartPoller(); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	configStore.Poll = true
	if err := h.store.SettingsPut(configStore); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Polling is turned on.")
}

// SettingsPollOff turn off polling functionality.
func (h *settingsHandler) pollOff(c echo.Context) error {
	configStore, err := h.store.SettingsGet()
	if err != nil {
		return c.String(http.StatusInternalServerError, msgSomethingWentWrong)
	}

	gaia.Cfg.Poll = false

	if err := pipeline.StopPoller(); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	configStore.Poll = true
	if err = h.store.SettingsPut(configStore); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Polling is turned off.")
}

type pollStatus struct {
	Status bool
}

// SettingsPollGet get status of polling functionality.
func (h *settingsHandler) pollGet(c echo.Context) error {
	configStore, err := h.store.SettingsGet()
	if err != nil {
		return c.String(http.StatusInternalServerError, msgSomethingWentWrong)
	}

	var ps pollStatus
	if gaia.Cfg.Poll {
		ps.Status = true
	} else {
		ps.Status = configStore.Poll
	}

	return c.JSON(http.StatusOK, ps)
}

type rbacToggleRequest struct {
	Enabled bool `json:"enabled"`
}

func (h *settingsHandler) rbacToggle(c echo.Context) error {
	var request rbacToggleRequest
	if err := c.Bind(&request); err != nil {
		gaia.Cfg.Logger.Error("failed to bind body", "error", err.Error())
		return c.String(http.StatusBadRequest, "Invalid body provided.")
	}

	settings, err := h.store.SettingsGet()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get store settings", "error", err.Error())
		return c.String(http.StatusInternalServerError, msgSomethingWentWrong)
	}

	settings.RBACEnabled = request.Enabled

	if err := h.store.SettingsPut(settings); err != nil {
		gaia.Cfg.Logger.Error("failed to put store settings", "error", err.Error())
		return c.String(http.StatusInternalServerError, "An error occurred while saving the settings.")
	}

	return c.String(http.StatusOK, "Settings have been updated.")
}

type rbacGetResponse struct {
	Enabled bool `json:"enabled"`
}

func (h *settingsHandler) rbacGet(c echo.Context) error {
	settings, err := h.store.SettingsGet()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get store settings", "error", err.Error())
		return c.String(http.StatusInternalServerError, msgSomethingWentWrong)
	}

	response := rbacGetResponse{}
	if gaia.Cfg.RBACEnabled {
		response.Enabled = true
	} else {
		response.Enabled = settings.RBACEnabled
	}

	return c.JSON(http.StatusOK, rbacGetResponse{Enabled: settings.RBACEnabled})
}
