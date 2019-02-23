package handlers

import (
	"net/http"

	"github.com/gaia-pipeline/gaia/workers/pipeline"

	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
)

// SettingsPollingOn turn on polling functionality.
func SettingsPollingOn(c echo.Context) error {
	gaia.Cfg.Poll = true
	err := pipeline.StartPoller()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Polling is turned on.")
}

// SettingsPollingOff turn off polling functionality.
func SettingsPollingOff(c echo.Context) error {
	gaia.Cfg.Poll = false
	err := pipeline.StopPoller()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Polling is turned off.")
}

// SettingsPollingGet get status of polling functionality.
func SettingsPollingGet(c echo.Context) error {
	poll := struct {
		status bool
	}{
		status: gaia.Cfg.Poll,
	}
	return c.JSON(http.StatusOK, poll)
}
