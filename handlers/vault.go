package handlers

import (
	"log"

	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
)

// AddSecret creates a secret using the vault.
func AddSecret(c echo.Context) error {
	log.Println(c.ParamValues())
	key := c.Param("key")
	value := c.Param("value")

	gaia.Cfg.Logger.Info("received add with: ", key, value)
	return nil
}

// ListSecrets retrieves all secrets from the vault.
func ListSecrets(c echo.Context) error {
	gaia.Cfg.Logger.Info("received list")
	return nil
}

// RemoveSecret removes a secret from the vault.
func RemoveSecret(c echo.Context) error {
	gaia.Cfg.Logger.Info("received remove")
	return nil
}
