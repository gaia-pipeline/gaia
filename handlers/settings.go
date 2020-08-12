package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

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

type rbacPutRequest struct {
	Enabled bool `json:"enabled"`
}

// @Summary Put RBAC settings
// @Description Save the given RBAC settings.
// @Tags settings
// @Accept json
// @Produce plain
// @Param RbacPutRequest body rbacPutRequest true "RBAC setting details."
// @Success 200 {string} string "Settings have been updated."
// @Failure 400 {string} string "Invalid body."
// @Failure 500 {string} string "Something went wrong while saving or retrieving rbac settings."
// @Router /settings/rbac [put]
func (h *settingsHandler) rbacPut(c echo.Context) error {
	var request rbacPutRequest
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

// @Summary Get RBAC settings
// @Description Get the given RBAC settings.
// @Tags settings
// @Produce json
// @Success 200 {object} rbacGetResponse
// @Failure 500 {string} string "Something went wrong while saving or retrieving rbac settings."
// @Router /settings/rbac [get]
func (h *settingsHandler) rbacGet(c echo.Context) error {
	settings, err := h.store.SettingsGet()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get store settings", "error", err.Error())
		return c.String(http.StatusInternalServerError, msgSomethingWentWrong)
	}

	response := rbacGetResponse{}
	// If RBAC is applied via config it takes priority.
	if gaia.Cfg.RBACEnabled {
		response.Enabled = true
	} else {
		response.Enabled = settings.RBACEnabled
	}

	return c.JSON(http.StatusOK, rbacGetResponse{Enabled: settings.RBACEnabled})
}
