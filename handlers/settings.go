package handlers

import (
	"net/http"

	"github.com/gaia-pipeline/gaia/workers/pipeline"

	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
)

// SettingsPollOn turn on polling functionality.
func SettingsPollOn(c echo.Context) error {
	gaia.Cfg.Poll = true
	err := pipeline.StartPoller()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Polling is turned on.")
}

// SettingsPollOff turn off polling functionality.
func SettingsPollOff(c echo.Context) error {
	gaia.Cfg.Poll = false
	err := pipeline.StopPoller()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Polling is turned off.")
}

// SettingsPollGet get status of polling functionality.
func SettingsPollGet(c echo.Context) error {
	poll := struct {
		Status bool
	}{
		Status: gaia.Cfg.Poll,
	}
	return c.JSON(http.StatusOK, poll)
}
