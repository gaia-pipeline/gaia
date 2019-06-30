package handlers

import (
	"net/http"

	"github.com/gaia-pipeline/gaia/helper/stringhelper"

	"github.com/gaia-pipeline/gaia/services"
	"github.com/labstack/echo/v4"
)

type addSecret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type updateSecret struct {
	Key   string `json:"key"`
	Value string `json:"newvalue"`
}

// SetSecret creates or updates a given secret
func SetSecret(c echo.Context) error {
	var key, value string
	if c.Request().Method == "POST" {
		s := new(addSecret)
		err := c.Bind(s)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		key = s.Key
		value = s.Value
	} else if c.Request().Method == "PUT" {
		s := new(updateSecret)
		err := c.Bind(s)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		key = s.Key
		value = s.Value
	}

	// Handle ignored special keys
	if stringhelper.IsContainedInSlice(ignoredVaultKeys, key, true) {
		return c.String(http.StatusBadRequest, "key is reserved and cannot be set/changed")
	}

	v, err := services.VaultService(nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = v.LoadSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v.Add(key, []byte(value))
	err = v.SaveSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusCreated, "secret successfully set")
}

// ListSecrets retrieves all secrets from the vault.
func ListSecrets(c echo.Context) error {
	secrets := make([]addSecret, 0)
	v, err := services.VaultService(nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = v.LoadSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	kvs := v.GetAll()
	for _, k := range kvs {
		// Handle ignored special keys
		if stringhelper.IsContainedInSlice(ignoredVaultKeys, k, true) {
			continue
		}

		s := addSecret{Key: k, Value: "**********"}
		secrets = append(secrets, s)
	}
	return c.JSON(http.StatusOK, secrets)
}

// RemoveSecret removes a secret from the vault.
func RemoveSecret(c echo.Context) error {
	key := c.Param("key")
	if key == "" {
		return c.String(http.StatusBadRequest, "invalid key given")
	}

	// Handle ignored special keys
	if stringhelper.IsContainedInSlice(ignoredVaultKeys, key, true) {
		return c.String(http.StatusBadRequest, "key is reserved and cannot be deleted")
	}

	v, err := services.VaultService(nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = v.LoadSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v.Remove(key)
	err = v.SaveSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "secret successfully deleted")
}
