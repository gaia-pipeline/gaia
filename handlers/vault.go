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

// CreateSecret creates a secret
// @Summary Create a secret.
// @Description Creates a secret.
// @Tags secrets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param secret body addSecret true "The secret to create"
// @Success 201 {string} string "secret successfully set"
// @Failure 400 {string} string "Error binding or key is reserved."
// @Failure 500 {string} string "Cannot get or load secrets"
// @Router /secret [post]
func CreateSecret(c echo.Context) error {
	var key, value string
	s := new(addSecret)
	err := c.Bind(s)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	key = s.Key
	value = s.Value

	return upsertSecret(c, key, err, value)
}

// UpdateSecret updates a given secret
// @Summary Update a secret.
// @Description Update a secret.
// @Tags secrets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param secret body updateSecret true "The secret to update with the new value"
// @Success 201 {string} string "secret successfully set"
// @Failure 400 {string} string "Error binding or key is reserved."
// @Failure 500 {string} string "Cannot get or load secrets"
func UpdateSecret(c echo.Context) error {
	var key, value string
	s := new(updateSecret)
	err := c.Bind(s)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	key = s.Key
	value = s.Value
	return upsertSecret(c, key, err, value)
}

// updates or creates a secret
func upsertSecret(c echo.Context, key string, err error, value string) error {
	// Handle ignored special keys
	if stringhelper.IsContainedInSlice(ignoredVaultKeys, key, true) {
		return c.String(http.StatusBadRequest, "key is reserved and cannot be set/changed")
	}

	v, err := services.DefaultVaultService()
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
// @Summary List all secrets.
// @Description Retrieves all secrets from the vault.
// @Tags secrets
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} addSecret "Secrets"
// @Failure 500 {string} string "Cannot get or load secrets"
// @Router /secrets [get]
func ListSecrets(c echo.Context) error {
	secrets := make([]addSecret, 0)
	v, err := services.DefaultVaultService()
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
// @Summary Removes a secret from the vault..
// @Description Removes a secret from the vault.
// @Tags secrets
// @Produce plain
// @Security ApiKeyAuth
// @Param key body string true "Key"
// @Success 200 {string} string "secret successfully deleted"
// @Failure 400 {string} string "key is reserved and cannot be deleted"
// @Failure 500 {string} string "Cannot get or load secrets"
// @Router /secret/:key [delete]
func RemoveSecret(c echo.Context) error {
	key := c.Param("key")
	if key == "" {
		return c.String(http.StatusBadRequest, "invalid key given")
	}

	// Handle ignored special keys
	if stringhelper.IsContainedInSlice(ignoredVaultKeys, key, true) {
		return c.String(http.StatusBadRequest, "key is reserved and cannot be deleted")
	}

	v, err := services.DefaultVaultService()
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
