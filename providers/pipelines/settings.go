package pipelines

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/gaia-pipeline/gaia"
)

// SettingsPollOn turn on polling functionality
// @Summary Turn on polling functionality.
// @Description Turns on the polling functionality for Gaia which periodically checks if there is new code to deploy for all pipelines.
// @Tags settings
// @Produce plain
// @Success 200 {string} string "Polling is turned on."
// @Failure 400 {string} string "Error while toggling poll setting."
// @Failure 500 {string} string "Internal server error while getting setting."
// @Router /settings/poll/on [post]
func (pp *PipelineProvider) SettingsPollOn(c echo.Context) error {
	settingsStore := pp.deps.SettingsStore

	configStore, err := settingsStore.SettingsGet()
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
	err = settingsStore.SettingsPut(configStore)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Polling is turned on.")
}

// SettingsPollOff turn off polling functionality.
// @Summary Turn off polling functionality.
// @Description Turns off the polling functionality for Gaia which periodically checks if there is new code to deploy for all pipelines.
// @Tags settings
// @Produce plain
// @Success 200 {string} string "Polling is turned off."
// @Failure 400 {string} string "Error while toggling poll setting."
// @Failure 500 {string} string "Internal server error while getting setting."
// @Router /settings/poll/off [post]
func (pp *PipelineProvider) SettingsPollOff(c echo.Context) error {
	settingsStore := pp.deps.SettingsStore

	configStore, err := settingsStore.SettingsGet()
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
	err = settingsStore.SettingsPut(configStore)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Polling is turned off.")
}

type pollStatus struct {
	Status bool
}

// SettingsPollGet get status of polling functionality.
// @Summary Get the status of the poll setting.
// @Description Gets the status of the poll setting.
// @Tags settings
// @Produce json
// @Success 200 {object} pollStatus "Poll status"
// @Failure 500 {string} string "Internal server error while getting setting."
// @Router /settings/poll [get]
func (pp *PipelineProvider) SettingsPollGet(c echo.Context) error {
	settingsStore := pp.deps.SettingsStore

	configStore, err := settingsStore.SettingsGet()
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
