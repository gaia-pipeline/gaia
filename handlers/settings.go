package handlers

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
)

type settingsHandler struct {
	store store.SettingsStore
}

func newSettingsHandler(store store.SettingsStore) *settingsHandler {
	return &settingsHandler{store: store}
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
