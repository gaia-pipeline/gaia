package handlers

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
)

const msgSomethingWentWrong = "Something went wrong while retrieving settings information."

type settingsHandler struct {
	store store.SettingsStore
}

func newSettingsHandler(store store.SettingsStore) *settingsHandler {
	return &settingsHandler{store: store}
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
