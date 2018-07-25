package handlers

import (
	"log"
	"net/http"

	"github.com/gaia-pipeline/gaia/security"

	"github.com/gaia-pipeline/gaia"
	"github.com/labstack/echo"
)

type secret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AddSecret creates a secret using the vault.
func AddSecret(c echo.Context) error {
	s := new(secret)
	err := c.Bind(s)
	if err != nil {
		gaia.Cfg.Logger.Error("error reading secret", "error", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	v, err := security.NewVault()
	if err != nil {
		gaia.Cfg.Logger.Error("error initializing vault", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v.Add(s.Key, []byte(s.Value))
	err = v.CloseVault()
	if err != nil {
		gaia.Cfg.Logger.Error("error saving vault", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	gaia.Cfg.Logger.Info("secret successfully added")
	return c.String(http.StatusOK, "secret successfully added")
}

// ListSecrets retrieves all secrets from the vault.
func ListSecrets(c echo.Context) error {
	secrets := make([]secret, 0)
	v, err := security.NewVault()
	if err != nil {
		gaia.Cfg.Logger.Error("error initializing vault", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = v.OpenVault()
	if err != nil {
		gaia.Cfg.Logger.Error("error opening vault", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	kvs := v.GetAll()
	for k, v := range kvs {
		s := secret{Key: k, Value: string(v)}
		secrets = append(secrets, s)
	}
	log.Println(secrets)
	return c.JSON(http.StatusOK, secrets)
}

// RemoveSecret removes a secret from the vault.
func RemoveSecret(c echo.Context) error {
	gaia.Cfg.Logger.Info("received remove")
	return nil
}
